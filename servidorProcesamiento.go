package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/user"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Variable que nos indica que el servidor se encuentra libre
var flagAvailable bool

// Estructura que se utilizará como plantilla para la decodificación de las requests
type Request struct {
	UserId  int    `json:"userId"`
	MVType  string `json:"mvtype"`
	Request string `json:"request"`
}

// Función para atender las solicitudes de creación de máquinas virtuales
func handlervm(w http.ResponseWriter, r *http.Request) {

	serverUser, _ := user.Current()
	addr := serverUser.HomeDir + ".ssh/"
	//Se envía la respuesta al cliente
	fmt.Fprintf(w, "sReqst: received")

	//Se lee el cuerpo de la solicitud y en caso de no poder leerlo, se imprime el error
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: no se pudo leer el body: %s\n", err)
	}

	//Se crea una variable tipo request en la cual se guardarán los datos del Json
	request := Request{}

	//Se decodifica el objeto Json y se guarda en la variable request
	derr := json.Unmarshal(reqBody, &request)
	if derr != nil {
		panic(derr)
	}

	sendSSH("192.168.1.40", addr+"known_hosts", addr+"id_rsa")
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

func sendSSH(ip string, addr string, addrKey string) {

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
		User: "jucat",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         0,
	}
	client, err := ssh.Dial("tcp", ip+":22", config)
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
	errRun := session.Run(`VBoxManage createvm --name Debian --ostype Debian11_64 --register & VBoxManage modifyvm Debian --cpus 2 --memory 1024 --vram 128 --nic1 bridged & VBoxManage modifyvm Debian --ioapic on --graphicscontroller vmsvga --boot1 disk & VBoxManage modifyvm Debian --bridgeadapter1 "Realtek PCIe GbE Family Controller" & VBoxManage storagectl Debian --name "SATA Controller" --add sata --bootable on & VBoxManage storageattach Debian --storagectl "SATA Controller" --port 0 --device 0 --type hdd --medium "C:\Discos\DiscoMulti.vdi"`)
	if errRun != nil {
		fmt.Println("Falló al ejecutar: " + errRun.Error())
	}
	fmt.Println(b.String())

}

func main() {

	flagAvailable = true
	http.HandleFunc("/", handler)
	http.HandleFunc("/procSolic", handlervm)
	fmt.Println("Servidor escuchando en el puerto :3333")
	http.ListenAndServe(":3333", nil)
}
