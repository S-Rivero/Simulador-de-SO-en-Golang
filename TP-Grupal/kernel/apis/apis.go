package apis

import (
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/apisSend"
	"github.com/sisoputnfrba/tp-golang/kernel/funcs"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/planCorto"
	"github.com/sisoputnfrba/tp-golang/kernel/planLargo"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

// ------------------ Recibir Post-Desalojo ---------------\\

func RecibirHiloDevuelto(w http.ResponseWriter, r *http.Request) {
	globals.Ejecutando = false
	var paquete globalvar.Paquete_Motivo
	err := commons.DecodificarJSON(w, r, &paquete)
	if err != nil {
		return
	}
	log.Printf("## (%d:%d) - Desalojado por %s", globals.ColaExec.Pcb.Pid, globals.ColaExec.Tid, paquete.Motivo)
	w.WriteHeader(http.StatusOK)

	globals.Mutex_ColaExec.Lock()
	hilo := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()

	switch paquete.Motivo {
	case "FinProceso":
		planCorto.MoverHilo(hilo, "EXIT")
		planLargo.ActualizarNew()
		return
	case "FinHilo":
		planCorto.MoverHilo(hilo, "EXIT")
		if !funcs.ChequearSeguirFuncionando() {
			return
		}
	case "MUTEX":
		planCorto.BloquearHilo(hilo, paquete.Motivo)
	case "THREAD_JOIN":
		planCorto.BloquearHilo(hilo, paquete.Motivo)
	case "DUMP_MEMORY":
		planCorto.BloquearHilo(hilo, paquete.Motivo)
	case "IO":
		planCorto.BloquearHilo(hilo, paquete.Motivo)
		if !globals.IO_Ejecutando {
			go planCorto.IO()
		}
	case "Quantum":
		planCorto.MoverHilo(hilo, "READY")
	case "Prioridad":
		planCorto.MoverHilo(hilo, "READY")
	case "":
		log.Printf("## (%d:%d) - Desalojado por VACIO", globals.ColaExec.Pcb.Pid, globals.ColaExec.Tid)
	}

	planCorto.ReadyToExec()
}

// ---------------------- SYSCALLS ---------------------\\
func HandlerSegmentationFault(w http.ResponseWriter, r *http.Request) {
	log.Printf("## Ha ocurrido un Segmentation Fault. Finalizando con el Proceso.")
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	planLargo.FinalizarProceso(tcbActual.Pcb.Pid)
	w.WriteHeader(http.StatusOK)
}
func HandleProcessCreate(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	if tcbActual == nil {
		globals.LogMin("syscall", globals.ContProcesos, 0, "PROCESS_CREATE")
	} else {
		globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "PROCESS_CREATE")
	}

	var req globalvar.Request_PROCESS_CREATE
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	go planLargo.CrearProceso(req.NombreArchivo, req.Tamanio, req.PrioridadHilo0)
	w.WriteHeader(http.StatusOK)
}

func HandlerProcessExit(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	if tcbActual == nil {
		log.Printf("No hay hilo en ejecución")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "PROCESS_EXIT")

	if tcbActual.Tid == 0 {
		globals.ContEjecucion++
		planLargo.FinalizarProceso(tcbActual.Pcb.Pid)
		w.WriteHeader(http.StatusOK)
	} else {
		log.Printf("El Hilo solicitante (%d:%d) no cumple con la condicion de ser Hilo Main (0)", tcbActual.Pcb.Pid, tcbActual.Tid)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("El Hilo solicitante no cumple con la condicion de ser Hilo Main (0)"))
	}
}

func HandlerThreadCreate(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "THREAD_CREATE")

	var req globalvar.Request_THREAD_CREATE

	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	planLargo.CrearHilo(tcbActual.Pcb, req.Prioridad, req.NombreArchivo)

	w.WriteHeader(http.StatusOK)
}

func HandlerThreadJoin(w http.ResponseWriter, r *http.Request) {
	var req globalvar.Request_THREAD_JOIN
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "THREAD_JOIN")

	var hiloJoin *globals.Tcb
	for _, hilo := range tcbActual.Pcb.Tids {
		if hilo.Tid == req.Tid {
			hiloJoin = hilo
			break
		}
	}
	if hiloJoin == nil {
		log.Printf("## (%d:%d) - THREAD_JOIN: Hilo %d no encontrado, Hilo %d continúa su ejecución\n", tcbActual.Pcb.Pid, tcbActual.Tid, req.Tid, tcbActual.Tid)
		return
	}
	if hiloJoin.Estado == "EXIT" {
		log.Printf("## (%d:%d) - THREAD_JOIN: Hilo %d ya ha finalizado, Hilo %d continúa su ejecución\n", tcbActual.Pcb.Pid, tcbActual.Tid, req.Tid, tcbActual.Tid)
		return
	}
	apisSend.DesalojarHilo("THREAD_JOIN")
	hiloJoin.HilosBloqueados = append(hiloJoin.HilosBloqueados, tcbActual)

	w.WriteHeader(http.StatusOK)
}

func HandlerThreadCancel(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "THREAD_CANCEL")

	var req globalvar.Request_THREAD_CANCEL
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	planLargo.FinalizarHilo(tcbActual.Pcb.Pid, req.Tid, true)

	w.WriteHeader(http.StatusOK)
}

func HandlerThreadExit(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "THREAD_EXIT")

	planLargo.FinalizarHilo(tcbActual.Pcb.Pid, tcbActual.Tid, true)

	w.WriteHeader(http.StatusOK)
}

func HandlerMutexCreate(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "MUTEX_CREATE")

	var req globalvar.Request_MUTEX_CREATE
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	planLargo.CrearMutex(tcbActual.Pcb, req.Recurso)
	w.WriteHeader(http.StatusOK)
}

func HandlerMutexLock(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "MUTEX_LOCK")

	var req globalvar.Request_MUTEX_LOCK
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	planLargo.MutexLock(tcbActual.Pcb, req.Recurso, tcbActual.Tid)
	w.WriteHeader(http.StatusOK)
}

func HandlerMutexUnlock(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "MUTEX_UNLOCK")

	var req globalvar.Request_MUTEX_LOCK
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	planLargo.MutexUnlock(tcbActual.Pcb, req.Recurso, tcbActual.Tid)
	w.WriteHeader(http.StatusOK)
}

func HandlerDumpMemory(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()

	if tcbActual == nil {
		http.Error(w, "No hay hilo en ejecución", http.StatusBadRequest)
		return
	}

	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "DUMP_MEMORY")

	apisSend.DesalojarHilo("DUMP_MEMORY")

	go planLargo.DoDump(tcbActual)
}

func HandlerIO(w http.ResponseWriter, r *http.Request) {
	globals.Mutex_ColaExec.Lock()
	tcbActual := globals.ColaExec
	globals.Mutex_ColaExec.Unlock()
	globals.LogMin("syscall", tcbActual.Pcb.Pid, tcbActual.Tid, "IO")

	var req globalvar.Request_IO
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	planCorto.InsertarIO(tcbActual, req.Tiempo)

	w.WriteHeader(http.StatusOK)
}
