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
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Variable que nos indica que el servidor se encuentra libre
var flagAvailable bool

// Estructura que se utilizará como plantilla para la decodificación de las requests
type MaquinaVirtual struct {
	Id           int    `json:"id"`
	Estado       string `json:"estado"`
	Hostname     string `json:"hostname"`
	IP           string `json:"ip"`
	Nombre       string `json:"nombre"`
	IdMF         int    `json:"idMF"`
	IdUser       int    `json:"idUser"`
	Contrasenia  string `json:"contrasenia"`
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

type HostName struct {
	Nombre string `json:"nombre"`
}

// Función para atender las solicitudes de creación de máquinas virtuales
func handlervm(w http.ResponseWriter, r *http.Request) {
	var mf MaquinaFisica
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
	request.IdMF = mf.Id
	request.TipoMV = 1
	estado := clasificar(request, mf)
	response := `{"estado":"` + estado + `"}`

	fmt.Fprintf(w, response)
}

func clasificar(maquinaVirtual MaquinaVirtual, mf MaquinaFisica) string {

	comando := ""
	estado := ""
	nombre := "Debian" + maquinaVirtual.NumeroNombre
	serverUser, _ := user.Current()
	addr := serverUser.HomeDir + "/.ssh"

	switch maquinaVirtual.Solicitud {

	case "start":
		comando = "VBoxManage startvm " + maquinaVirtual.Nombre //+ request.Nombre
		fmt.Println(comando)
		actualizarEstado(strconv.Itoa(maquinaVirtual.Id), "Procesando")
		sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando)
		time.Sleep(120 * time.Second)
		comandoIP := `VBoxManage guestproperty get "` + maquinaVirtual.Nombre + `" "/VirtualBox/GuestInfo/Net/0/V4/IP"`
		ip := sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comandoIP)
		actualizar(strconv.Itoa(maquinaVirtual.Id), ip, "ip")
		estado = actualizarEstado(strconv.Itoa(maquinaVirtual.Id), "Iniciada")
		break

	case "create":
		fmt.Println(mf.BridgeAdapter)
		comando = `VBoxManage createvm --name ` + nombre + ` --ostype Debian11_64 --register & VBoxManage modifyvm  ` + nombre + ` --cpus 2 --memory 1024 --vram 128 --nic1 bridged & VBoxManage modifyvm  ` + nombre + ` --ioapic on --graphicscontroller vmsvga --boot1 disk & VBoxManage modifyvm  ` + nombre + ` --bridgeadapter1 "` + mf.BridgeAdapter + `" & VBoxManage storagectl  ` + nombre + ` --name "SATA Controller" --add sata --bootable on & VBoxManage storageattach  ` + nombre + ` --storagectl "SATA Controller" --port 0 --device 0 --type hdd --medium "C:\Discos\VMTipo1.vdi"`
		sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando)
		break

	case "finish":
		comando = "VBoxManage controlvm " + maquinaVirtual.Nombre + " poweroff"
		fmt.Println(comando)
		sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando)
		estado = actualizarEstado(strconv.Itoa(maquinaVirtual.Id), "Apagada")
		break

	case "delete":
		comando = "VBoxManage unregistervm " + maquinaVirtual.Nombre + " --delete"
		sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando)
		break
	}

	return estado

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
	fmt.Println(comando)
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
	fmt.Println("COMANDO SSH: " + comando)
	if errRun != nil {
		fmt.Println("Error 3 " + err.Error())
		return "Error: Falló al ejecutar: " + errRun.Error()
	}

	return b.String()
}

func guardarVM(vm MaquinaVirtual) {
	jsonBody := []byte(`{"nombre":` + `"Debian` + vm.NumeroNombre + `","ip":"` + vm.IP + `","hostname":"` + obtenerHostName(vm.TipoMV) + `","idUser":"` + strconv.Itoa(vm.IdUser) + `","contrasenia":"` + vm.Contrasenia + `","estado":"` + vm.Estado + `","tipoMV":"` + strconv.Itoa(vm.TipoMV) + `","idMF":"` + strconv.Itoa(vm.IdMF) + `"}`)
	fmt.Println(`{"nombre":` + `"Debian` + vm.NumeroNombre + `","ip":"` + vm.IP + `","hostname":"` + obtenerHostName(vm.TipoMV) + `","idUser":"` + strconv.Itoa(vm.IdUser) + `","contrasenia":"` + vm.Contrasenia + `,"estado":"` + vm.Estado + `","tipoMV":"` + strconv.Itoa(vm.TipoMV) + `","idMF":"` + strconv.Itoa(vm.IdMF) + `"}`)
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

func actualizar(id string, cambio string, tipoCambio string) {

	var req *http.Request
	var error error

	trimmedOutput := strings.TrimSpace(cambio)
	fmt.Println(trimmedOutput)
	pattern := `(\d+\.\d+\.\d+\.\d+)`
	re := regexp.MustCompile(pattern)
	match := re.FindString(trimmedOutput)
	fmt.Println(match)
	fmt.Println(`{"id":"` + id + `","cambio":"` + match + `"}`)

	jsonBody := []byte(`{"id":"` + id + `","cambio":"` + match + `"}`)
	req, error = http.NewRequest(http.MethodPost, "http://localhost:8080/api/updatevmi", bytes.NewBuffer(jsonBody))
	req.Header.Add("Content-Type", "application/json")

	if error != nil {
		fmt.Printf("client: could not create request: %s\n", error)
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
	fmt.Println("client: response body: %s\n", resBody)
}

func actualizarEstado(id string, estado string) string {

	var req *http.Request
	var error error

	jsonBody := []byte(`{"id":"` + id + `","cambio":"` + estado + `"}`)
	req, error = http.NewRequest(http.MethodPost, "http://localhost:8080/api/updatevms", bytes.NewBuffer(jsonBody))
	req.Header.Add("Content-Type", "application/json")

	if error != nil {
		fmt.Printf("client: could not create request: %s\n", error)
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
	estadoRqst := HostName{}

	derr := json.Unmarshal(resBody, &estadoRqst)
	if derr != nil {
		panic(derr)
	}
	fmt.Println(estadoRqst.Nombre)
	return estadoRqst.Nombre
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

func obtenerHostName(id int) string {
	requestURL := fmt.Sprintf("http://localhost:8080/api/getHostname/%d", id)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	//Se crea una variable tipo request en la cual se guardarán los datos del Json
	hostname := HostName{}

	//Se decodifica el objeto Json y se guarda en la variable request
	derr := json.Unmarshal(resBody, &hostname)
	if derr != nil {
		panic(derr)
	}
	return hostname.Nombre
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
