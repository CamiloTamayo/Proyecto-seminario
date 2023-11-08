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

const ipSolic = "10.0.48.216"
const ipProc = "10.0.48.216"
const ipWEB = "10.0.48.216"
const serverPort = 3333

// Handler que atenderá las solicitudes de creación de máquinas virtuales
func handlercvm(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	flag := true
	//Se envía la respuesta al cliente
	fmt.Println(w.Header())

	//Se lee el cuerpo de la solicitud y en caso de no poder leerlo, se imprime el error
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("server: no se pudo leer el body: %s\n", err)
	}
	fmt.Println(bytes.NewReader(reqBody))

	//Agrega un elemento a la cola
	myQueue.Enqueue(reqBody)

	//Ciclo que nos permite escuchar el servidor para enviarle una Request
	for flag {
		//Se realiza la solicitud para saber si el servidor de procesamiento está disponible
		res, err := http.Get("http://" + ipProc + ":3333")
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		//Se obtiene la respuesta
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		//Se convierte la respuesta a boolean
		valorBool, err := strconv.ParseBool(string(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//Si el servidor de procesamiento está disponible, se envía la solicitud y se acaba el ciclo
		if valorBool {
			solicitud, _ := myQueue.Dequeue()
			fmt.Fprintf(w, string(solicitarMV(solicitud)))
			flag = false
		}
	}
}

func solicitarMV(request []byte) []byte {

	bodyReader := bytes.NewReader(request)
	requestURL := fmt.Sprintf("http://"+ipProc+":%d/procSolic", serverPort)
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
	return resBody
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

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://"+ipWEB+":4200")
}

func main() {
	http.HandleFunc("/crearmv", handlercvm)
	http.HandleFunc("/solicitud", handlercvm)
	fmt.Println("Servidor escuchando en el puerto :8000")
	http.ListenAndServe(ipSolic+":8000", nil)
}
