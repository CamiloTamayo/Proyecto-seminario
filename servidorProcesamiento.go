package main

import (
	//"encoding/json"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Estructura que se utilizará como plantilla para la decodificación de las requests
type Request struct {
	UserId  int    `json:"userId"`
	MVType  string `json:"mvtype"`
	Request string `json:"request"`
}

// Función para atender las solicitudes de creación de máquinas virtuales
func handlervm(w http.ResponseWriter, r *http.Request) {

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
	fmt.Printf(request.MVType + " " + request.Request)
}

// Variable que nos indica que el servidor se encuentra libre
var flagAvailable bool

// Función para dar respuesta a las solicitudes de disponibilidad del servidor de procesamiento
func handler(w http.ResponseWriter, r *http.Request) {

	if flagAvailable {
		//Se envía la respuesta al cliente
		fmt.Fprintf(w, "true")
	} else {
		fmt.Fprintf(w, "false")
	}
}

func main() {

	flagAvailable = true

	http.HandleFunc("/", handler)
	http.HandleFunc("/procSolic", handlervm)
	fmt.Println("Servidor escuchando en el puerto :3333")
	http.ListenAndServe(":3333", nil)
}
