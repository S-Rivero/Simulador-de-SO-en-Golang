package planLargo

import (
	"log"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/apisSend"
	"github.com/sisoputnfrba/tp-golang/kernel/funcs"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/kernel/planCorto"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

//----------------- PROCESOS ----------------\\

func CrearProceso(filename string, tamanio int, prioridad int) {
	nuevoProceso := &globals.Pcb{
		Pid:          globals.ContProcesos,
		ContadorTids: 0,
		Tids:         []*globals.Tcb{},
		Mutexs:       []globals.Mutex_type{},
		Estado:       "NEW",
	}
	globals.ContProcesos++

	globals.Mutex_ColaProcesos.Lock()
	globals.ColaProcesos = append(globals.ColaProcesos, nuevoProceso)
	globals.Mutex_ColaProcesos.Unlock()
	globals.LogMin("crearProc", nuevoProceso.Pid, 0, "")

	if len(globals.ColaNew) == 0 {
		if apisSend.InicializarProcesoMem(tamanio, nuevoProceso.Pid) {
			CrearHilo(nuevoProceso, prioridad, filename)
			return
		}
	}

	// Añadir el proceso a la cola NEW de forma segura
	x := &globals.NewProc{
		Pcb:            nuevoProceso,
		PrioridadHilo0: prioridad,
		Tamanio:        tamanio,
		Filename:       filename,
	}
	globals.Mutex_ColaNew.Lock()
	globals.ColaNew = append(globals.ColaNew, x)
	globals.Mutex_ColaNew.Unlock()

	// Notificar al planificador que hay un nuevo proceso en la cola NEW
	globals.SemaforoColaNew <- struct{}{}
}

func FinalizarProceso(pid int) {
	if apisSend.FinalizarProcesoMem(pid) {
		for i, cola := range globals.ColaProcesos {
			if cola.Pid == pid {
				for _, hilo := range globals.ColaProcesos[i].Tids {
					FinalizarHilo(hilo.Pcb.Pid, hilo.Tid, false)
				}
				globals.Mutex_ColaProcesos.Lock()
				globals.ColaProcesos = append(globals.ColaProcesos[:i], globals.ColaProcesos[i+1:]...)
				globals.Mutex_ColaProcesos.Unlock()
				break
			}
		}
		globals.LogMin("finProc", pid, 0, "")
	}
}

func ActualizarNew() {
	time.Sleep(1 * time.Second)
	if len(globals.ColaNew) > 0 {
		if apisSend.InicializarProcesoMem(globals.ColaNew[0].Tamanio, globals.ColaNew[0].Pcb.Pid) {
			var nuevoProceso *globals.NewProc
			globals.Mutex_ColaNew.Lock()
			nuevoProceso, globals.ColaNew = globals.ColaNew[0], globals.ColaNew[1:]
			globals.Mutex_ColaNew.Unlock()
			CrearHilo(nuevoProceso.Pcb, nuevoProceso.PrioridadHilo0, nuevoProceso.Filename)
		}
	}
	if funcs.ChequearSeguirFuncionando() {
		planCorto.ReadyToExec()
	} else {
		funcs.ColaReadyVacia()
	}
}

//----------------- HILOS ----------------\\

func CrearHilo(proceso *globals.Pcb, prioridad int, filename string) {
	nuevoHilo := &globals.Tcb{
		Tid:       proceso.ContadorTids,
		Prioridad: prioridad,
		Estado:    "READY",
		Pcb:       proceso,
		Filename:  filename,
	}
	if prioridad > globals.PrioridadMinima {
		globals.PrioridadMinima = prioridad
	}
	if apisSend.InicializarHiloMem(*nuevoHilo) {
		proceso.Tids = append(proceso.Tids, nuevoHilo)
		proceso.ContadorTids++
		globals.LogMin("crearHilo", nuevoHilo.Pcb.Pid, nuevoHilo.Tid, "")
		planCorto.InsertarHiloReady(nuevoHilo)
		return
	}
	log.Println("No se pudo inicializar el hilo")
}

func FinalizarHilo(pid int, tid int, informar bool) {
	var avisoMem bool
	if informar {
		avisoMem = apisSend.FinalizarHiloMem(pid, tid)
	} else {
		avisoMem = true
	}
	if avisoMem {
		globals.Mutex_ColaExec.Lock()
		if globals.ColaExec != nil {
			if globals.ColaExec.Tid == tid && globals.ColaExec.Pcb.Pid == pid {
				tcbActual := globals.ColaExec
				globals.Mutex_ColaExec.Unlock()
				if informar {
					apisSend.DesalojarHilo("FinHilo")
				} else {
					apisSend.DesalojarHilo("FinProceso")
				}
				DesbloquearPendientes(tcbActual)
				globals.LogMin("finHilo", tcbActual.Pcb.Pid, tcbActual.Tid, "")
				return
			}
		}
		globals.Mutex_ColaExec.Unlock()
		var tcbActual *globals.Tcb
		tcbActual = funcs.BuscarHiloPorPidYTid("READY", pid, tid)
		if tcbActual == nil {
			tcbActual = funcs.BuscarHiloPorPidYTid("BLOCK", pid, tid)
		}
		if tcbActual == nil {
			tcbActual = funcs.BuscarHiloPorPidYTid("EXIT", pid, tid)
			if tcbActual != nil {
				log.Printf("## (%d:%d) Hilo finalizado previamente ", pid, tid)
				return
			}
		}
		if tcbActual == nil {
			log.Printf("Hilo (%d:%d) no encontrado\n", pid, tid)
			return
		}
		DesbloquearPendientes(tcbActual)
		planCorto.MoverHilo(tcbActual, "EXIT")
		globals.LogMin("finHilo", tcbActual.Pcb.Pid, tcbActual.Tid, "")
	}
}

//----------------- Mutex ----------------\\

func CrearMutex(pcb *globals.Pcb, recurso string) {
	mutex := globals.Mutex_type{
		NombreMutex: recurso,
		Estado:      false,
		Tid:         -1,
		Bloqueados:  make([]int, 0),
	}
	pcb.Mutexs = append(pcb.Mutexs, mutex)
}

func MutexLock(pcb *globals.Pcb, recurso string, tid int) {
	i := BuscarMutex(pcb, recurso)
	if i == -1 {
		log.Println("El mutex "+recurso+" del proceso %d no existe", pcb.Pid)
		FinalizarHilo(pcb.Pid, tid, true)
		return
	}
	status := CheckMutex(pcb.Mutexs[i], tid)
	if status == 1 {
		log.Println("ERROR. El mutex "+recurso+" del proceso %d esta tomado por el hilo solicitante (%d)", pcb.Pid, tid)
		return
	}
	if status == 0 {
		pcb.Mutexs[i].Estado = true
		pcb.Mutexs[i].Tid = tid
	} else {
		pcb.Mutexs[i].Bloqueados = append(pcb.Mutexs[i].Bloqueados, tid)
		apisSend.DesalojarHilo("MUTEX")
	}

}

func MutexUnlock(pcb *globals.Pcb, recurso string, tid int) {
	i := BuscarMutex(pcb, recurso)
	if i == -1 {
		log.Println("El mutex "+recurso+" del proceso %d no existe", pcb.Pid)
		FinalizarHilo(pcb.Pid, tid, true)
		return
	}
	if CheckMutex(pcb.Mutexs[i], tid) != 1 {
		log.Println("El mutex "+recurso+" del proceso %d no esta tomado por el hilo %d", pcb.Pid, tid)
		return
	}

	if len(pcb.Mutexs[i].Bloqueados) == 0 {
		pcb.Mutexs[i].Estado = false
		pcb.Mutexs[i].Tid = -1
		return
	}

	var HiloADesbloquear int

	if len(pcb.Mutexs[i].Bloqueados) > 1 {
		HiloADesbloquear, pcb.Mutexs[i].Bloqueados = pcb.Mutexs[i].Bloqueados[0], pcb.Mutexs[i].Bloqueados[1:]
	} else {
		HiloADesbloquear, pcb.Mutexs[i].Bloqueados = pcb.Mutexs[i].Bloqueados[0], make([]int, 0)
	}

	hilo := funcs.BuscarHiloPorPidYTid("BLOCK", pcb.Pid, HiloADesbloquear)
	if hilo == nil {
		return
	}
	pcb.Mutexs[i].Tid = HiloADesbloquear
	planCorto.DesbloquearHilo(hilo)
}

// DEVUELVE: '0' No esta tomado ----- '1' Si esta tomado por este Hilo ----- '2' Si esta tomado por otro Hilo
func CheckMutex(mutex globals.Mutex_type, tid int) int {
	if !mutex.Estado {
		return 0
	}
	if mutex.Tid == tid {
		return 1
	} else {
		return 2
	}
}

// Devuelve la posicion del mutex. Devuelve -1 si no existe
func BuscarMutex(pcb *globals.Pcb, recurso string) int {
	for i, mutex := range pcb.Mutexs {
		if mutex.NombreMutex == recurso {
			return i
		}
	}
	return -1
}

func DesbloquearPendientes(tcb *globals.Tcb) {
	for i := 0; i < len(tcb.HilosBloqueados); i++ {
		planCorto.DesbloquearHilo(tcb.HilosBloqueados[i])
		tcb.HilosBloqueados = append(tcb.HilosBloqueados[:i], tcb.HilosBloqueados[i+1:]...)
		i--
	}
}

func DoDump(hilo *globals.Tcb) {
	time.Sleep(1 * time.Second)
	paquete := globalvar.Paquete_PidTid{
		Pid: hilo.Pcb.Pid,
		Tid: hilo.Tid,
	}
	resp := commons.EnviarPaqueteYEsperarOK[globalvar.Paquete_PidTid](paquete, "MEMORIA", "DUMP_MEMORY")

	if resp {
		planCorto.DesbloquearHilo(hilo)
	} else {
		FinalizarProceso(paquete.Pid)
		ActualizarNew()
	}
}
