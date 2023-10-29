package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Variable que nos indica que el servidor se encuentra libre
var flagAvailable bool

// Estructura que se utilizará como plantilla para la decodificación de las requests
type MaquinaVirtual struct {
	Id        string `json:"id"`
	Estado    string `json:"estado"`
	Hostname  string `json:"hostname"`
	IP        string `json:"ip"`
	Nombre    string `json:"nombre"`
	IdMF      int    `json:"idMF"`
	IdUser    int    `json:"idUser"`
	TipoMV    int    `json:"tipoMV"`
	Solicitud string `json:"solicitud"`
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

	//serverUser, _ := user.Current()
	//addr := serverUser.HomeDir + "/.ssh"
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
	fmt.Println(request)
	mf := obtenerMF(2)
	//comando := clasificar(request, mf)
	request.IdMF = mf.Id
	request.TipoMV = 1
	//sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando)
	response := `{"idMF":"` + strconv.Itoa(request.IdMF) + `", "tipoMV":"` + strconv.Itoa(request.TipoMV) + `"}`
	guardarVM(request)
	fmt.Fprintf(w, response)
}

func clasificar(maquinaVirtual MaquinaVirtual, mf MaquinaFisica) string {

	comando := ""

	switch maquinaVirtual.Solicitud {

	case "start":
		comando = "VBoxManage startvm Debian" //+ request.Nombre
		break

	case "create":
		fmt.Println(mf.BridgeAdapter)
		comando = `VBoxManage createvm --name Debian --ostype Debian11_64 --register & VBoxManage modifyvm Debian --cpus 2 --memory 1024 --vram 128 --nic1 bridged & VBoxManage modifyvm Debian --ioapic on --graphicscontroller vmsvga --boot1 disk & VBoxManage modifyvm Debian --bridgeadapter1 "` + mf.BridgeAdapter + `" & VBoxManage storagectl Debian --name "SATA Controller" --add sata --bootable on & VBoxManage storageattach Debian --storagectl "SATA Controller" --port 0 --device 0 --type hdd --medium "C:\DiscoMulti.vdi"`
		fmt.Println(comando)

	case "finish":
		comando = "VBoxManage controlvm Debian poweroff" //+ request.Nombre + " poweroff"
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

func sendSSH(mf MaquinaFisica, addr string, addrKey string, comando string) {

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
	client, err := ssh.Dial("tcp", mf.Ip+":22", config)
	if err != nil {
		panic("Fallo al dial: " + err.Error())
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		panic("Falló al crear la sesión: " + err.Error())
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	errRun := session.Run(comando)
	fmt.Println("CORRE COMANDO")
	if errRun != nil {
		fmt.Println("Falló al ejecutar: " + errRun.Error())
	}
	fmt.Println(b.String())

}

func guardarVM(vm MaquinaVirtual) {
	vm.Id = ""
	//reqBody, err := json.Marshal(vm)

	jsonBody := []byte(`{"nombre":"` + vm.Nombre + `","ip":"` + vm.IP + `","hostname":"` + vm.Hostname + `","idUser": ` + strconv.Itoa(vm.IdUser) + `,"estado":"` + vm.Estado + `","tipoMV":"` + strconv.Itoa(vm.TipoMV) + `","idMF": ` + strconv.Itoa(vm.IdMF) + `}`)
	//bodyReader := bytes.NewReader(jsonBody)
	fmt.Println(string(jsonBody))

	//requestURL := fmt.Sprintf("http://localhost:8080/api/savevm")

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

func main() {

	flagAvailable = true
	http.HandleFunc("/", handler)
	http.HandleFunc("/procSolic", handlervm)
	fmt.Println("Servidor escuchando en el puerto :3333")
	http.ListenAndServe(":3333", nil)
}
