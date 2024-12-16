package main

import (
	"fmt"
	"log"
	"net/http"

	//"os"
	"sync"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
)

func main() {
	var wg sync.WaitGroup

	// Obtener el nombre de la prueba de los argumentos
	/*
		testName := ""
		if len(os.Args) > 1 {
			testName = os.Args[1]
		}
	*/

	// Inicializa el log y el config con el nombre de la prueba
	//globals.Config = commons.InstanciarPathsWithTest[globals.ModuleConfig](testName, "")

	globals.Config = commons.InstanciarPaths[globals.ModuleConfig]()
	commons.InstanciarIPs(globals.Config.Ip, "", globals.Config.IpKernel, globals.Config.IpMemory)

	// Interfaz //
	mux := http.NewServeMux()

	mux.HandleFunc("/mensaje", commons.RecibirMensaje)
	mux.HandleFunc("/handshake", commons.HandshakeHandler)
	mux.HandleFunc("/NewHandshake", commons.Handler_HandshakeProlijo)
	mux.HandleFunc("/RecibirHilo", globals.RecibirHilo)
	mux.HandleFunc("POST /Interrupcion", globals.UpdateInterrupt)

	port := fmt.Sprintf(":%d", globals.Config.Port)

	go commons.LevantarServidor(port, mux, &wg)
	log.Printf("El módulo CPU está a la escucha en el puerto %s", port)

	commons.EsperarConexion("MEMORIA")

	wg.Wait()
}
