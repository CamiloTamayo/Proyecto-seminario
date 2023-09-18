package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Se declara la estructura tipo cola para guardar las solicitudes
type Queue []([]byte)

var myQueue Queue

const serverPort = 3333

// Handler que atenderá las solicitudes de creación de máquinas virtuales
func handlercvm(w http.ResponseWriter, r *http.Request) {

	flag := true
	//Se envía la respuesta al cliente
	fmt.Fprintf(w, "sReqst: received")

	//Se lee el cuerpo de la solicitud y en caso de no poder leerlo, se imprime el error
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: no se pudo leer el body: %s\n", err)
	}

	//Agrega un elemento a la cola
	myQueue.Enqueue(reqBody)

	for flag {
		res, err := http.Get("http://localhost:3333")
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		valorBool, err := strconv.ParseBool(string(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if valorBool {
			solicitud, _ := myQueue.Dequeue()
			solicitarMV(solicitud)
			flag = false
		}
	}
}

func solicitarMV(request []byte) {
	fmt.Printf("AAAAAAAAAAAAAAA")
	bodyReader := bytes.NewReader(request)
	requestURL := fmt.Sprintf("http://localhost:%d/procSolic", serverPort)
	req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("client: response body: %s\n", resBody)
}

// Método para agregar un Request a la cola
func (q *Queue) Enqueue(item []byte) {
	*q = append(*q, item)
}

// Elimina un Request de la cola y lo devuelve
func (q *Queue) Dequeue() ([]byte, error) {

	if len(*q) == 0 {
		return []byte{}, fmt.Errorf("La cola está vacía")
	}
	item := (*q)[0]
	*q = (*q)[1:]

	return item, nil
}

func main() {
	http.HandleFunc("/crearmv", handlercvm)
	fmt.Println("Servidor escuchando en el puerto :8080")
	http.ListenAndServe(":8080", nil)
}
