package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/sisoputnfrba/tp-golang/memoria/apis"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/memoria/memUs"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
)

func main() {
	var wg sync.WaitGroup

	// Obtener el nombre de la prueba y variante de los argumentos
	/*
		testName := ""
		variant := ""
		if len(os.Args) > 1 {
			testName = os.Args[1]
			if len(os.Args) > 2 {
				variant = os.Args[2]
			}
		}

		// Inicializa el log y el config
		globals.Config = commons.InstanciarPathsWithTest[globals.ModuleConfig](testName, variant)
	*/

	globals.Config = commons.InstanciarPaths[globals.ModuleConfig]()
	commons.InstanciarIPs(globals.Config.IpCpu, globals.Config.IpFilesystem, globals.Config.IpKernel, globals.Config.Ip)

	// Interfaz //
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", commons.RecibirMensaje)
	mux.HandleFunc("/handshake", commons.HandshakeHandler)
	mux.HandleFunc("/NewHandshake", commons.Handler_HandshakeProlijo)

	// CPU
	mux.HandleFunc("POST /SolicitudContexto", apis.Handler_SolicitudContexto)
	mux.HandleFunc("POST /ActualizarContexto", apis.Handler_ActualizarContexto)
	mux.HandleFunc("POST /SegmentationFault", apis.Handler_ActualizarContexto)
	mux.HandleFunc("POST /SolicitudInstruccion", apis.Handler_SolicitudInstruccion)
	mux.HandleFunc("POST /LeerMemoria", apis.Handler_LeerMemoria)
	mux.HandleFunc("POST /EscribirMemoria", apis.Handler_EscribirMemoria)

	// Kernel
	mux.HandleFunc("POST /InicializarProceso", apis.Handler_InicializarProceso)
	mux.HandleFunc("POST /FinalizarProceso", apis.Handler_FinalizarProceso)
	mux.HandleFunc("POST /InicializarHilo", apis.Handler_InicializarHilo)
	mux.HandleFunc("POST /FinalizarHilo", apis.Handler_FinalizarHilo)
	mux.HandleFunc("POST /DUMP_MEMORY", apis.Handler_DumpMemory)

	port := fmt.Sprintf(":%d", globals.Config.Port)

	go commons.LevantarServidor(port, mux, &wg)
	log.Printf("El módulo memoria está a la escucha en el puerto %s", port)

	memUs.InitRAM()
	commons.EsperarConexion("FS")

	wg.Wait()
}
