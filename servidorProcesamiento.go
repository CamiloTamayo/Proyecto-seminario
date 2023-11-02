package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Variable que nos indica que el servidor se encuentra libre
var flagAvailable bool

// Estructura que se utilizará como plantilla para la decodificación de las requests
type MaquinaVirtual struct {
	Id           string `json:"id"`
	Estado       string `json:"estado"`
	Hostname     string `json:"hostname"`
	IP           string `json:"ip"`
	Nombre       string `json:"nombre"`
	IdMF         int    `json:"idMF"`
	IdUser       int    `json:"idUser"`
	TipoMV       int    `json:"tipoMV"`
	Solicitud    string `json:"solicitud"`
	NumeroNombre string `json:"numeroNombre"`
}

type MaquinaFisica struct {
	Id            int      `json:"idMF"`
	Ip            string   `json:"ip"`
	Ram           int      `json:"ramMB"`
	Cpu           int      `json:"cpu"`
	Storage       int      `json:"storageGB"`
	Hostname      string   `json:"hostname"`
	Os            string   `json:"os"`
	BridgeAdapter string   `json:"bridgeAdapter"`
	Maquinas      []string `json:"maquinas"`
}

// Función para atender las solicitudes de creación de máquinas virtuales
func handlervm(w http.ResponseWriter, r *http.Request) {
	var mf MaquinaFisica
	serverUser, _ := user.Current()
	addr := serverUser.HomeDir + "/.ssh"
	//Se envía la respuesta al cliente

	//Se lee el cuerpo de la solicitud y en caso de no poder leerlo, se imprime el error
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: no se pudo leer el body: %s\n", err)
	}

	//Se crea una variable tipo request en la cual se guardarán los datos del Json
	request := MaquinaVirtual{}

	//Se decodifica el objeto Json y se guarda en la variable request
	derr := json.Unmarshal(reqBody, &request)
	if derr != nil {
		panic(derr)
	}
	fmt.Print("REQUEST: ")
	fmt.Println(request)
	if strings.Compare(request.Solicitud, "create") == 0 {
		mf = asignar()
		request.IdMF = mf.Id
		guardarVM(request)
	} else {
		mf = obtenerMF(request.IdMF)
	}
	comando := clasificar(request, mf)
	request.IdMF = mf.Id
	request.TipoMV = 1
	sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando)
	response := `{"idMF":"` + strconv.Itoa(request.IdMF) + `", "tipoMV":"` + strconv.Itoa(request.TipoMV) + `"}`

	fmt.Fprintf(w, response)
}

func clasificar(maquinaVirtual MaquinaVirtual, mf MaquinaFisica) string {

	comando := ""
	nombre := "Debian" + maquinaVirtual.NumeroNombre

	switch maquinaVirtual.Solicitud {

	case "start":
		comando = "VBoxManage startvm " + maquinaVirtual.Nombre //+ request.Nombre
		fmt.Println(comando)
		break

	case "create":
		fmt.Println(mf.BridgeAdapter)
		comando = `VBoxManage createvm --name ` + nombre + ` --ostype Debian11_64 --register & VBoxManage modifyvm  ` + nombre + ` --cpus 2 --memory 1024 --vram 128 --nic1 bridged & VBoxManage modifyvm  ` + nombre + ` --ioapic on --graphicscontroller vmsvga --boot1 disk & VBoxManage modifyvm  ` + nombre + ` --bridgeadapter1 "` + mf.BridgeAdapter + `" & VBoxManage storagectl  ` + nombre + ` --name "SATA Controller" --add sata --bootable on & VBoxManage storageattach  ` + nombre + ` --storagectl "SATA Controller" --port 0 --device 0 --type hdd --medium "C:\Discos\Debian-Base2.vdi"`
		break

	case "finish":
		comando = "VBoxManage controlvm " + maquinaVirtual.Nombre + " poweroff" //+ request.Nombre + " poweroff"
		fmt.Println(comando)
		break

	case "delete":
		comando = "VBoxManage unregistervm " + maquinaVirtual.Nombre + " --delete"
		break
	}

	return comando

}

// Función para dar respuesta a las solicitudes de disponibilidad del servidor de procesamiento
func handler(w http.ResponseWriter, r *http.Request) {

	if flagAvailable {
		//Se envía la respuesta al cliente
		fmt.Fprintf(w, "true")
	} else {
		fmt.Fprintf(w, "false")
	}
}

func sendSSH(mf MaquinaFisica, addr string, addrKey string, comando string) string {
	hostKeyCallback, err := knownhosts.New(addr)
	if err != nil {
		log.Fatal(err)
	}
	file := addrKey
	key, errFile := ioutil.ReadFile(file)

	if errFile != nil {
		log.Fatalf("No se pudo leer la llave privada: %v", errFile)
	}

	signer, errSecond := ssh.ParsePrivateKey(key)
	if errSecond != nil {
		log.Fatalf("No se pudo convertir la llave privada: %v", errSecond)
	}

	config := &ssh.ClientConfig{
		User: mf.Hostname,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         0,
	}
	fmt.Println(mf.Ip)
	client, err := ssh.Dial("tcp", mf.Ip+":22", config)
	if err != nil {
		fmt.Println("Error 1" + err.Error())
		return "Error: Fallo al dial " + err.Error()
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		fmt.Println("Error 2" + err.Error())
		return "Error: No se pudo crear la sesión"
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	errRun := session.Run(comando)

	if errRun != nil {
		fmt.Println("Error 3 " + err.Error())
		return "Error: Falló al ejecutar: " + errRun.Error()
	}

	//fmt.Println("TERMINA SSH")

	return b.String()
}

func guardarVM(vm MaquinaVirtual) {
	vm.Id = ""
	fmt.Print("TIPOMV: ")
	fmt.Println(vm.TipoMV)
	jsonBody := []byte(`{"nombre":` + `"Debian` + vm.NumeroNombre + `","ip":"` + vm.IP + `","hostname":"` + vm.Hostname + `","idUser": ` + strconv.Itoa(vm.IdUser) + `,"estado":"` + vm.Estado + `","tipoMV":"` + strconv.Itoa(vm.TipoMV) + `","idMF": ` + strconv.Itoa(vm.IdMF) + `}`)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/savevm", bytes.NewBuffer(jsonBody))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	resBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: %s\n", resBody)

}

func eliminarVM(vm MaquinaVirtual) {

}

func obtenerMF(idMF int) MaquinaFisica {
	requestURL := fmt.Sprintf("http://localhost:8080/api/getmf/%d", idMF)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	//Se crea una variable tipo request en la cual se guardarán los datos del Json
	mf := MaquinaFisica{}

	//Se decodifica el objeto Json y se guarda en la variable request
	derr := json.Unmarshal(resBody, &mf)
	if derr != nil {
		panic(derr)
	}
	return mf
}

func obtenerIdMV() int {
	requestURL := fmt.Sprintf("http://localhost:8080/api/obtenerMayor")
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	resBody, err := ioutil.ReadAll(res.Body)

	var idMaquinaVirtual int

	derr := json.Unmarshal(resBody, &idMaquinaVirtual)
	if derr != nil {
		panic(derr)
	}
	fmt.Print("OBTENERIDMV: ")
	fmt.Println(idMaquinaVirtual)
	return idMaquinaVirtual + 1
}

func asignar() MaquinaFisica {
	fmt.Println("SE LLAMA ASIGNAR")
	serverUser, _ := user.Current()
	addr := serverUser.HomeDir + `\.ssh`
	mf := MaquinaFisica{}
	var flag bool = true
	requestURL := fmt.Sprintf("http://localhost:8080/api/getmfs")
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	lista := []MaquinaFisica{}

	derr := json.Unmarshal(resBody, &lista)
	if derr != nil {
		panic(derr)
	}
	//fmt.Println(lista)
	for flag {
		var ale int = rand.Intn(len(lista))
		mf = lista[ale]
		fmt.Print(flag)
		var respuesta string = sendSSH(mf, addr+`\known_hosts`, addr+`\id_rsa`, "calc")
		//fmt.Println(addr)
		if !strings.Contains(respuesta, "Error") {
			flag = false
			fmt.Println(respuesta)
		}
	}

	return mf
}

func main() {

	//fmt.Println(obtenerIdMV())
	flagAvailable = true
	http.HandleFunc("/", handler)
	http.HandleFunc("/procSolic", handlervm)
	fmt.Println("Servidor escuchando en el puerto :3333")
	http.ListenAndServe(":3333", nil)
}
