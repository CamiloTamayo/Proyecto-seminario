package main

import (
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

// Se declara la estructura tipo cola para guardar las solicitudes
type Queue []Request

var myQueue Queue
var aux int

// Handler que atenderá las solicitudes de creación de máquinas virtuales
func handlercvm(w http.ResponseWriter, r *http.Request) {

	//Se envía la respuesta al cliente
	fmt.Fprintf(w, "received")

	//Se lee el cuerpo de la solicitud y en caso de no poder leerolo, se imprime el error
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
	aux += 1
	//Agrega un elemento a la cola
	myQueue.Enqueue(request)

	//Elimnar elementos de la cola y mostrarlos
	if aux == 3 {
		for len(myQueue) > 0 {
			item, err := myQueue.Dequeue()
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println("Request extraido:, ", item)
		}
	}

}

// Método para agregar un Request a la cola
func (q *Queue) Enqueue(item Request) {
	*q = append(*q, item)
}

// Elimina un Request de la cola y lo devuelve
func (q *Queue) Dequeue() (Request, error) {
	if len(*q) == 0 {
		return Request{}, fmt.Errorf("La cola está vacía")
	}
	item := (*q)[0]
	*q = (*q)[1:]

	return item, nil
}

func main() {
	aux = 0
	http.HandleFunc("/crearmv", handlercvm)
	fmt.Println("Servidor escuchando en el puerto :8080")
	http.ListenAndServe(":8080", nil)
}
