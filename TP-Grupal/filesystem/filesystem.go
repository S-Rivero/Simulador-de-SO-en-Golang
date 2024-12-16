package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/sisoputnfrba/tp-golang/filesystem/globals"
	"github.com/sisoputnfrba/tp-golang/filesystem/utils"
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

		// Inicializa el log y el config
		globals.Config = commons.InstanciarPathsWithTest[globals.ModuleConfig](testName, "")
	*/

	globals.Config = commons.InstanciarPaths[globals.ModuleConfig]()
	commons.InstanciarIPs("", globals.Config.Ip, "", globals.Config.IpMemory)

	mux := http.NewServeMux()
	mux.HandleFunc("/NewHandshake", commons.Handler_HandshakeProlijo)
	mux.HandleFunc("POST /DUMP_MEMORY", utils.HandleCreateDump)

	port := fmt.Sprintf(":%d", globals.Config.Port)

	go commons.LevantarServidor(port, mux, &wg)
	log.Printf("El módulo fileSystem está a la escucha en el puerto %s", port)

	// Inicializar filesystem
	if err := utils.FSInit(); err != nil {
		log.Fatalf("Error inicializando el sistema de archivos: %v", err)
	}

	wg.Wait()
}
