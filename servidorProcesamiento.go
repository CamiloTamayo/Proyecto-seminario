package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Variable y constantes globales
var mu sync.Mutex

var ipServer = ""
var ipApi = ""
var ipWEB = ""

// Estructura que se utilizará como plantilla para la decodificación de las requests
type MaquinaVirtual struct {
	Id               int    `json:"id"`
	Estado           string `json:"estado"`
	Hostname         string `json:"hostname"`
	IP               string `json:"ip"`
	Nombre           string `json:"nombre"`
	IdMF             int    `json:"idMF"`
	IdUser           int    `json:"idUser"`
	Contrasenia      string `json:"contrasenia"`
	TipoMV           string `json:"tipoMV"`
	Solicitud        string `json:"solicitud"`
	NumeroNombre     string `json:"numeroNombre"`
	SistemaOperativo int    `json:"sistemaOperativo"`
	NombreDisco      string `json:"nombreDisco"`
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

type Caracteristicas struct {
	Ram     int `json:"ramMB"`
	Cpu     int `json:"cpu"`
	Storage int `json:"storageGB"`
}

type StringReq struct {
	Nombre string `json:"nombre"`
}

type AppConfig struct {
	UrlAPI      string `json:"api"`
	UrlServidor string `json:"servidor"`
	UrlWEB      string `json:"web"`
}

// Función para atender las solicitudes entrantes
func handlervm(w http.ResponseWriter, r *http.Request) {
	var mf MaquinaFisica
	var isHere bool = false
	enableCors(&w, r)
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error: %v", err)
		fmt.Fprintf(w, "server: no se pudo leer el body: %s\n", err)
		return
	}

	request := MaquinaVirtual{}
	derr := json.Unmarshal(reqBody, &request)
	if derr != nil {
		log.Printf("Error: %v", err)
		return
	}

	if strings.Compare(request.Solicitud, "create") == 0 {
		trimmedOutput := strings.TrimSpace(r.RemoteAddr)
		pattern := `(\d+\.\d+\.\d+\.\d+)`
		re := regexp.MustCompile(pattern)
		match := re.FindString(trimmedOutput)
		fmt.Println(match)
		mf, isHere = asignar(match)
		request.IdMF = mf.Id
		newvm := guardarVM(request)
		time.Sleep(4 * time.Second)
		fmt.Println(newvm.Id)
		request.Id = newvm.Id
	} else {
		mf = obtenerMF(request.IdMF)
		request.IdMF = mf.Id
	}

	estado := clasificar(request, mf, isHere)
	response := `{"estado":"` + estado + `"}`
	fmt.Fprintf(w, response)
}

// Función para habilitar la política CORS
func enableCors(w *http.ResponseWriter, r *http.Request) {
	switch host := r.Header.Get("Origin"); host {
	case "http://localhost:4200":
		(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		break
	case ipWEB + ":4200":
		(*w).Header().Set("Access-Control-Allow-Origin", ipWEB+":4200")
	}
}

// Función de asignación de una máquina física para las solicitudes de creación de VMs
func asignar(ip string) (MaquinaFisica, bool) {
	var flag bool = true
	var isHere bool = false
	serverUser, _ := user.Current()
	addr := serverUser.HomeDir + `/.ssh`
	mf := MaquinaFisica{}
	requestURL := fmt.Sprintf(ipApi + "/api/getmfs")
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		log.Printf("Error: %v", err)
	}
	resBody, err := io.ReadAll(res.Body)
	lista := []MaquinaFisica{}
	derr := json.Unmarshal(resBody, &lista)
	if derr != nil {
		log.Printf("Error: %v", derr)
	}

	for _, mfisica := range lista {
		if ip == mfisica.Ip {
			isHere = true
			mf = mfisica
			fmt.Println("HEREEEEE")
		}
	}

	for flag && !isHere {
		fmt.Println("NOT HEREEEEE")
		var ale int = rand.Intn(len(lista))
		mf = lista[ale]
		break
		var respuesta string = sendSSH(mf, addr+`/known_hosts`, addr+`/id_rsa`, "")
		if !strings.Contains(respuesta, "Error") {
			flag = false
		}
	}
	return mf, isHere
}

/*
Esta función tiene como propósito principal almacenar una máquina virtual en la base de datos de la plataforma Desktop Cloud.
Recibe como parámetro la máquina virtual a almacenar.
*/
func guardarVM(vm MaquinaVirtual) MaquinaVirtual{
	jsonBody := []byte(`{"nombre":` + `"MaquinaVirtual` + vm.NumeroNombre + `","ip":"` + vm.IP + `","hostname":""` + `,"idUser":"` + strconv.Itoa(vm.IdUser) + `","contrasenia":"` + vm.Contrasenia + `","estado":"` + vm.Estado + `","tipoMV":"` + vm.TipoMV + `","idMF":"` + strconv.Itoa(vm.IdMF) + `","os":"` + strconv.Itoa(vm.SistemaOperativo) + `"}`)
	req, err := http.NewRequest(http.MethodPost, ipApi+"/api/savevm", bytes.NewBuffer(jsonBody))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		log.Printf("Error: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Printf("client: response body GUARDARVM: %s\n", resBody)
	vmreturn := MaquinaVirtual{}
	derr := json.Unmarshal(resBody, &vmreturn)
	if derr != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Println(vmreturn)
	return vmreturn
}

// Función para obtener una máquina física de la base de datos dado su ID.
func obtenerMF(idMF int) MaquinaFisica {
	requestURL := fmt.Sprintf(ipApi+"/api/getmf/%d", idMF)
	res, err := http.Get(requestURL)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	resBody, err := io.ReadAll(res.Body)
	mf := MaquinaFisica{}
	derr := json.Unmarshal(resBody, &mf)
	if derr != nil {
		log.Printf("Error: %v", derr)
	}
	return mf
}

func clasificar(maquinaVirtual MaquinaVirtual, mf MaquinaFisica, isHere bool) string {

	comando := ""
	estado := ""
	nombre := "MaquinaVirtual" + maquinaVirtual.NumeroNombre
	serverUser, _ := user.Current()
	addr := serverUser.HomeDir + "/.ssh"
	fmt.Println(maquinaVirtual.Solicitud)
	switch maquinaVirtual.Solicitud {

	case "start":
		var ip string = ""
		fmt.Println("ENTRA START "+ strconv.FormatBool(isHere) )
		fmt.Println("ENTRA START "+ strconv.Itoa(maquinaVirtual.Id) )
		comando = "VBoxManage startvm " + nombre 
		if !isHere{
			comando += " --type headless"
		}else{
			comando += " --type gui"
		}
		fmt.Println("ENTRA START "+ comando )
		actualizarEstado(strconv.Itoa(maquinaVirtual.Id), "Procesando")
		mu.Lock()
		sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando)
		mu.Unlock()
		comando2 := `VBoxManage guestproperty get "` + nombre + `" "VMIP"`
		for true {
			ip = sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando2)
			fmt.Println(ip)
			if strings.Contains(ip, "No") {
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
		actualizarIP(strconv.Itoa(maquinaVirtual.Id), ip)
		estado = actualizarEstado(strconv.Itoa(maquinaVirtual.Id), "Iniciada")
		break

	case "create":
		var caracteristicas Caracteristicas = obtenerTipoMaquina(maquinaVirtual.TipoMV)
		var disco = maquinaVirtual.NombreDisco
		comando = `VBoxManage createvm --name ` + nombre + ` --register & VBoxManage modifyvm  ` + nombre + ` --cpus ` + strconv.Itoa(caracteristicas.Cpu) + ` --memory ` + strconv.Itoa(caracteristicas.Ram) + ` --vram 128 --nic1 bridged & VBoxManage modifyvm  ` + nombre + ` --ioapic on --graphicscontroller vmsvga --boot1 disk & VBoxManage modifyvm  ` + nombre + ` --bridgeadapter1 "` + mf.BridgeAdapter + `" & VBoxManage storagectl  ` + nombre + ` --name "SATA Controller" --add sata --bootable on & VBoxManage storageattach  ` + nombre + ` --storagectl "SATA Controller" --port 0 --device 0 --type hdd --medium "C:\Discos\` + disco + `.vdi"`
		sendSSH(mf, addr+"/known_hosts", addr+"/id_rsa", comando)
		maquinaVirtual.Solicitud = "start"
		clasificar(maquinaVirtual, mf, isHere)
		break

	case "finish":
		comando = "VBoxManage controlvm " + maquinaVirtual.Nombre + " poweroff"
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

func sendSSH(mf MaquinaFisica, addr string, addrKey string, comando string) string {
	file := addrKey
	key, errFile := ioutil.ReadFile(file)

	if errFile != nil {
		log.Printf("Error: No se pudo leer la llave privada: %v", errFile)
	}

	signer, errSecond := ssh.ParsePrivateKey(key)
	if errSecond != nil {
		log.Printf("Error: No se pudo convertir la llave privada: %v", errSecond)
	}

	config := &ssh.ClientConfig{
		User: mf.Hostname,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         0,
	}
	client, err := ssh.Dial("tcp", mf.Ip+":22", config)
	if err != nil {
		log.Printf("Error: Fallo al dial: %v", err)
		return "Error: Fallo al dial " + err.Error()
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		log.Printf("Error: No se pudo crear la sesion: %v", err)
		return "Error: No se pudo crear la sesión"
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b
	errRun := session.Run(comando)
	if errRun != nil {
		log.Printf("Error: Fallo al ejecutar instruccion ssh: %v", errRun)
		return "Error: Falló al ejecutar: "
	}
	return b.String()
}

func actualizarIP(idVM string, ip string) {

	var req *http.Request
	var error error

	trimmedOutput := strings.TrimSpace(ip)
	pattern := `(\d+\.\d+\.\d+\.\d+)`
	re := regexp.MustCompile(pattern)
	match := re.FindString(trimmedOutput)

	jsonBody := []byte(`{"id":"` + idVM + `","cambio":"` + match + `"}`)
	req, error = http.NewRequest(http.MethodPost, ipApi+"/api/updatevmi", bytes.NewBuffer(jsonBody))
	req.Header.Add("Content-Type", "application/json")

	if error != nil {
		log.Printf("Error: %v", error)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Printf("Error: %v", err)
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Println("client: response body: %s\n", resBody)
}

func actualizarEstado(id string, estado string) string {

	var req *http.Request
	var error error

	jsonBody := []byte(`{"id":"` + id + `","cambio":"` + estado + `"}`)
	req, error = http.NewRequest(http.MethodPost, ipApi+"/api/updatevms", bytes.NewBuffer(jsonBody))
	req.Header.Add("Content-Type", "application/json")

	if error != nil {
		log.Printf("Error: %v", error)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Printf("Error: %v", err)
	}
	estadoRqst := StringReq{}

	derr := json.Unmarshal(resBody, &estadoRqst)
	if derr != nil {
		log.Printf("Error: %v", derr)
	}
	return estadoRqst.Nombre
}

func obtenerTipoMaquina(nombre string) Caracteristicas {
	requestURL := fmt.Sprintf(ipApi + "/api/getTipoMaquina/" + nombre)
	res, err := http.Get(requestURL)
	if err != nil {
		log.Printf("Error: error making http request: %v", err)
	}
	resBody, err := io.ReadAll(res.Body)
	//Se crea una variable tipo request en la cual se guardarán los datos del Json
	tipoMaquina := Caracteristicas{}

	//Se decodifica el objeto Json y se guarda en la variable request
	derr := json.Unmarshal(resBody, &tipoMaquina)
	if derr != nil {
		log.Printf("Error: %v", err)
	}
	return tipoMaquina
}

func leerIPs() {
	fileContent, err := ioutil.ReadFile("ips.json")
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Decodificar el JSON en una estructura Go
	var config AppConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Acceder a las IPs
	ipApi = config.UrlAPI
	ipServer = config.UrlServidor
	ipWEB = config.UrlWEB
}

func main() {
	logFile, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("No se pudo abrir el archivo de registro:", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	leerIPs()
	fmt.Println("Escuchando en: " + ipServer)
	http.HandleFunc("/procSolic", handlervm)
	log.Fatal(http.ListenAndServe(ipServer, nil))
}
