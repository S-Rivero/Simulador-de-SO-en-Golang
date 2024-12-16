package apisSend

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

//---------------------- MEMORIA ----------------------\\

func FinalizarProcesoMem(pid int) bool {
	paquete := globalvar.Paquete_Pid{
		Pid: pid,
	}
	res := commons.EnviarPaqueteYEsperarOK[globalvar.Paquete_Pid](paquete, "MEMORIA", "FinalizarProceso")
	return res
}

func FinalizarHiloMem(pid int, tid int) bool {
	paquete := globalvar.Paquete_PidTid{
		Pid: pid,
		Tid: tid,
	}
	res := commons.EnviarPaqueteYEsperarOK[globalvar.Paquete_PidTid](paquete, "MEMORIA", "FinalizarHilo")
	return res
}

func InicializarProcesoMem(tamanio int, pid int) bool {
	paquete := globalvar.Paquete_Tamanio{
		Tamanio: tamanio,
		Pid:     pid,
	}
	res := commons.EnviarPaqueteYEsperarOK[globalvar.Paquete_Tamanio](paquete, "MEMORIA", "InicializarProceso")
	return res
}
func InicializarHiloMem(hilo globals.Tcb) bool {
	paquete := globalvar.Paquete_CrearHilo{
		Pid:       hilo.Pcb.Pid,
		Tid:       hilo.Tid,
		Prioridad: hilo.Prioridad,
		Filename:  hilo.Filename,
	}
	res := commons.EnviarPaqueteYEsperarOK[globalvar.Paquete_CrearHilo](paquete, "MEMORIA", "InicializarHilo")
	return res
}

//---------------------- CPU --------------------------\\

func MandarHiloCPU(pid int, tid int) {
	paquete := globalvar.Paquete_PidTid{
		Tid: tid,
		Pid: pid,
	}
	globals.Ejecutando = true
	log.Printf("## (%d:%d) Siendo enviado a CPU", pid, tid)
	commons.EnviarPaquete[globalvar.Paquete_PidTid](paquete, "CPU", "RecibirHilo")
}

func DesalojarHilo(motivo string) {
	paquete := globalvar.Paquete_Motivo{
		Motivo: motivo,
	}
	globals.Ejecutando = false
	//if globals.Config.Quantum > 500 || motivo != "Quantum" {
	//	log.Printf("## (%d:%d) - Desalojado por %s", globals.ColaExec.Pcb.Pid, globals.ColaExec.Tid, motivo)
	//}
	commons.EnviarPaquete[globalvar.Paquete_Motivo](paquete, "CPU", "Interrupcion")
}

// TESTING
type MostrarColas_ struct {
	New      int `json:"new"`
	Ready    int `json:"ready"`
	Blocked  int `json:"blocked"`
	Finished int `json:"finished"`
}

func MostrarColas(w http.ResponseWriter, r *http.Request) {
	colas := MostrarColas_{
		New:      len(globals.ColaNew),
		Ready:    len(globals.ColaReady),
		Blocked:  len(globals.ColaBlock),
		Finished: len(globals.ColaExit),
	}

	w.Header().Set("Content-Type", "application/json")

	respuesta, err := json.Marshal(colas)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(respuesta); err != nil {
		log.Printf("Error al escribir la respuesta: %v", err)
		return
	}

	log.Printf("ColaNew: %d", len(globals.ColaNew))
	log.Printf("ColaReady: %d", len(globals.ColaReady))
	log.Printf("ColaBlock: %d", len(globals.ColaBlock))
	log.Printf("ColaExit: %d", len(globals.ColaExit))
	log.Printf("ColaExec: %v", globals.ColaExec)
}

func MostrarColasLocal() {

	new := make([]globals.NewProc, 0)
	for _, hilo := range globals.ColaNew {
		new = append(new, *hilo)
	}
	ready := make([]globals.Tcb, 0)
	for _, hilo := range globals.ColaReady {
		ready = append(ready, *hilo)
	}
	block := make([]globals.Tcb, 0)
	for _, hilo := range globals.ColaBlock {
		block = append(block, *hilo)
	}
	exit := make([]globals.Tcb, 0)
	for _, hilo := range globals.ColaExit {
		exit = append(exit, *hilo)
	}
	multi := make([][]globals.Tcb, 0)
	log.Printf("Multi: %v", multi)
	for i, cola := range globals.ColaMultiPrio {
		multi = append(multi, make([]globals.Tcb, 0))
		for _, hilo := range cola {
			multi[i] = append(multi[i], *hilo)
		}
	}
	log.Println("############")
	log.Printf("ColaExec: %v", globals.ColaExec)
	log.Printf("ColaReady[%d]: %v", len(globals.ColaReady), ready)
	for i, cola := range multi {
		if len(cola) > 0 {
			log.Printf("## Prioridad [%d]: (%v)", i, cola)
		}
	}
	log.Printf("ColaBlock[%d]: %v", len(globals.ColaBlock), block)
	log.Printf("ColaExit[%d]: %v", len(globals.ColaExit), exit)
	log.Printf("ColaNew[%d]: %v", len(globals.ColaNew), new)
	log.Println("############")
}
