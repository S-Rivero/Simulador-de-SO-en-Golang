package funcs

import (
	"fmt"
	"log"
	"strings"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

// ------------------ BUSQUEDA --------------------\\

func BuscarProcesoPorPid(pid int) *globals.Pcb {
	globals.Mutex_ColaProcesos.Lock()
	for _, proceso := range globals.ColaProcesos {
		if proceso.Pid == pid {
			globals.Mutex_ColaProcesos.Unlock()
			return proceso
		}
	}
	globals.Mutex_ColaProcesos.Unlock()
	return nil
}

func BuscarHiloPorPidYTid(cola string, pid int, Tid int) *globals.Tcb {
	switch cola {
	case "READY":
		if globals.Config.SchedulerAlgorithm == "CMN" {
			globals.Mutex_Multicola.Lock()
			for _, cola := range globals.ColaMultiPrio {
				for _, hilo := range cola {
					if hilo.Pcb.Pid == pid && hilo.Tid == Tid {
						globals.Mutex_Multicola.Unlock()
						return hilo
					}
				}
			}
			globals.Mutex_Multicola.Unlock()
		} else {
			globals.Mutex_ColaReady.Lock()
			for _, tcb := range globals.ColaReady {
				if tcb.Pcb.Pid == pid && tcb.Tid == Tid {
					globals.Mutex_ColaReady.Unlock()
					return tcb
				}
			}
			globals.Mutex_ColaReady.Unlock()
		}
	case "BLOCK":
		globals.Mutex_ColaBlock.Lock()
		for _, tcb := range globals.ColaBlock {
			if tcb.Pcb.Pid == pid && tcb.Tid == Tid {
				globals.Mutex_ColaBlock.Unlock()
				return tcb
			}
		}
		globals.Mutex_ColaBlock.Unlock()
	case "EXIT":
		globals.Mutex_ColaExit.Lock()
		for _, tcb := range globals.ColaExit {
			if tcb.Pcb.Pid == pid && tcb.Tid == Tid {
				globals.Mutex_ColaExit.Unlock()
				return tcb
			}
		}
		globals.Mutex_ColaExit.Unlock()
	}
	return nil
}

func ChequearSeguirFuncionando() bool {
	if globals.Config.SchedulerAlgorithm == "CMN" {
		for _, cola := range globals.ColaMultiPrio {
			if len(cola) > 0 {
				return true
			}
		}
	} else {
		if len(globals.ColaReady) > 0 {
			return true
		}
	}
	if len(globals.ColaBlock) > 0 {
		return true
	}
	if len(globals.ColaNew) > 0 {
		return true
	}
	return false
}

func ColaReadyVacia() {
	log.Println("No se encuentra ningun hilo disponible para seguir ejecutando")
	globals.Inicial = true
	var filename string
	var tamanio string
	var prio string
	log.Println("Ingrese parametros para iniciar un nuevo Proceso (<NombreArchivo> _ <Tamaño> _ <PrioridadHilo0>): ")
	for {
		_, err1 := fmt.Scan(&filename)
		_, err2 := fmt.Scan(&tamanio)
		_, err3 := fmt.Scan(&prio)
		if err1 == err2 && err1 == err3 && err3 == nil {
			filename = strings.Trim(filename, " ")
			tamanio = strings.Trim(tamanio, " ")
			prio = strings.Trim(prio, " ")
			if filename != "" && tamanio != "" && prio != "" {
				break
			}
		}
		if err1 != nil {
			log.Println("Error:", err1)
		}
		if err2 != nil {
			log.Println("Error:", err2)
		}
		if err3 != nil {
			log.Println("Error:", err3)
		}
		log.Println("Parametros ingresados de manera incorrecta. Por favor intente nuevamente.")
		log.Println("Ingrese parametros para iniciar un nuevo Proceso (<NombreArchivo> _ <Tamaño> _ <PrioridadHilo0>): ")
	}
	paquete := globalvar.Request_PROCESS_CREATE{
		NombreArchivo:  filename,
		Tamanio:        commons.StrToInt(tamanio),
		PrioridadHilo0: commons.StrToInt(prio),
	}
	commons.EnviarPaquete[globalvar.Request_PROCESS_CREATE](paquete, "KERNEL", "PROCESS_CREATE")
}
