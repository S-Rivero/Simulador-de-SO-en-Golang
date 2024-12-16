package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/apis"
	"github.com/sisoputnfrba/tp-golang/kernel/apisSend"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/planLargo"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
)

func main() {
	var wg sync.WaitGroup

	// Validar argumentos de línea de comandos
	if len(os.Args) < 4 {
		log.Fatalf("Uso: %s <archivo_pseudocodigo> <tamanio_proceso> <prioridad> [test_name]", os.Args[0])
	}

	archivo := os.Args[1]
	tamanio := commons.StrToInt(os.Args[2])
	prioridad := commons.StrToInt(os.Args[3])
	/*
			// Obtener el nombre de la prueba si se proporciona
			testName := ""
			variant := ""
			if len(os.Args) > 4 {
				testName = os.Args[4]
				if len(os.Args) > 5 {
					variant = os.Args[5]
				}
			}

		// Inicializa el log y el config
		globals.Config = commons.InstanciarPathsWithTest[globals.ModuleConfig](testName, variant)

	*/
	globals.Config = commons.InstanciarPaths[globals.ModuleConfig]()
	commons.InstanciarIPs(globals.Config.IpCpu, "", globals.Config.Ip, globals.Config.IpMemory)

	// Interfaz //
	mux := http.NewServeMux()
	mux.HandleFunc("POST /mensaje", commons.RecibirMensaje)
	mux.HandleFunc("/handshake", commons.HandshakeHandler)
	mux.HandleFunc("/NewHandshake", commons.Handler_HandshakeProlijo)

	// Handlers para recibir respuestas
	mux.HandleFunc("POST /DevolverHilo", apis.RecibirHiloDevuelto)

	// Syscalls
	mux.HandleFunc("POST /PROCESS_CREATE", apis.HandleProcessCreate)
	mux.HandleFunc("POST /PROCESS_EXIT", apis.HandlerProcessExit)
	mux.HandleFunc("POST /THREAD_CREATE", apis.HandlerThreadCreate)
	mux.HandleFunc("POST /THREAD_JOIN", apis.HandlerThreadJoin)
	mux.HandleFunc("POST /THREAD_CANCEL", apis.HandlerThreadCancel)
	mux.HandleFunc("POST /THREAD_EXIT", apis.HandlerThreadExit)
	mux.HandleFunc("POST /MUTEX_CREATE", apis.HandlerMutexCreate)
	mux.HandleFunc("POST /MUTEX_LOCK", apis.HandlerMutexLock)
	mux.HandleFunc("POST /MUTEX_UNLOCK", apis.HandlerMutexUnlock)
	mux.HandleFunc("POST /DUMP_MEMORY", apis.HandlerDumpMemory)
	mux.HandleFunc("POST /IO", apis.HandlerIO)
	mux.HandleFunc("POST /SegmentationFault", apis.HandlerSegmentationFault)

	// Testing
	mux.HandleFunc("GET /mostrarColas", apisSend.MostrarColas)

	port := fmt.Sprintf(":%d", globals.Config.Port)

	go commons.LevantarServidor(port, mux, &wg)
	log.Printf("El módulo kernel está a la escucha en el puerto %s", port)

	commons.EsperarConexion("CPU")
	commons.EsperarConexion("MEMORIA")

	time.Sleep(2 * time.Second)

	go planLargo.CrearProceso(archivo, tamanio, prioridad)

	wg.Wait()
}
