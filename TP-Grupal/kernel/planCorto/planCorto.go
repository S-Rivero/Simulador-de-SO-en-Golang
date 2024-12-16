package planCorto

import (
	"log"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/apisSend"
	"github.com/sisoputnfrba/tp-golang/kernel/funcs"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
)

//----------------- Planificador Corto ----------------\\

func InsertarHiloReady(hilo *globals.Tcb) {
	switch globals.Config.SchedulerAlgorithm {
	case "FIFO":
		InsertarFifo(hilo)
	case "PRIORIDADES":
		InsertarPrioridades(hilo)
	case "CMN":
		InsertarMulticolas(hilo)
	}
	if globals.Inicial {
		globals.Inicial = false
		ReadyToExec()
	}
}

func InsertarFifo(nuevoHilo *globals.Tcb) {
	globals.Mutex_ColaReady.Lock()
	globals.ColaReady = append(globals.ColaReady, nuevoHilo)
	globals.Mutex_ColaReady.Unlock()
}

func InsertarPrioridades(nuevoHilo *globals.Tcb) {
	globals.Mutex_ColaReady.Lock()
	globals.ColaReady = append(globals.ColaReady, nuevoHilo)
	globals.Mutex_ColaReady.Unlock()
	globals.Mutex_ColaExec.Lock()
	if globals.ColaExec != nil && globals.Ejecutando {
		if nuevoHilo.Prioridad < globals.ColaExec.Prioridad {
			globals.Mutex_ColaExec.Unlock()
			apisSend.DesalojarHilo("Prioridad")
			return
		}
	}
	globals.Mutex_ColaExec.Unlock()
}

func SiguientePrioridades() *globals.Tcb {
	globals.Mutex_ColaReady.Lock()
	if len(globals.ColaReady) > 0 {
		hilo := globals.ColaReady[0]

		for _, tcb := range globals.ColaReady {
			if hilo.Prioridad > tcb.Prioridad {
				hilo = tcb
			}
		}
		globals.Mutex_ColaReady.Unlock()
		return hilo
	}
	globals.Mutex_ColaReady.Unlock()
	return nil
}

// ---------------------- MULTICOLAS --------------------\\

func InsertarMulticolas(nuevoHilo *globals.Tcb) {
	// Instancia las colas de prioridad en caso de no existir todavia
	// len() devuelve la cantidad de ranuras existentes. Cuenta el 0 como uno mas, si tenemos 7 prioridades, necesitamos 8 ranuras
	// Estando hecho deberia ser: len() == k+1 && globals.PrioridadMinima == k
	if len(globals.ColaMultiPrio) <= globals.PrioridadMinima {
		for i := len(globals.ColaMultiPrio) - 1; i <= globals.PrioridadMinima; i++ {
			if len(globals.ColaMultiPrio) <= i { // Si la cola de prioridad no existe, crea una vacia
				sliceVacio := []*globals.Tcb{}
				globals.ColaMultiPrio = append(globals.ColaMultiPrio, sliceVacio)
			}
		}
	}
	// Inserta el hilo en la cola de prioridad correspondiente
	globals.ColaMultiPrio[nuevoHilo.Prioridad] = append(globals.ColaMultiPrio[nuevoHilo.Prioridad], nuevoHilo)
	ChequearDesalojo()
}
func ChequearDesalojo() {
	globals.Mutex_ColaExec.Lock()
	if globals.ColaExec != nil && globals.Ejecutando {
		for _, cola := range globals.ColaMultiPrio {
			if len(cola) > 0 {
				if cola[0].Prioridad < globals.ColaExec.Prioridad {
					globals.Mutex_ColaExec.Unlock()
					apisSend.DesalojarHilo("Prioridad")
					return
				}
			}
		}
	}
	globals.Mutex_ColaExec.Unlock()
}
func SiguienteMultiColas() *globals.Tcb {
	cola := SeleccionarColaPrioritaria()
	if cola == nil {
		return nil
	}
	return cola[0]
}
func SeleccionarColaPrioritaria() []*globals.Tcb {
	globals.Mutex_Multicola.Lock()
	if len(globals.ColaMultiPrio) > 0 {
		for _, cola := range globals.ColaMultiPrio {
			if len(cola) > 0 {
				globals.Mutex_Multicola.Unlock()
				return cola
			}
		}
	}
	globals.Mutex_Multicola.Unlock()
	return nil
}
func WaitQuantum(Instancia int) {
	time.Sleep(time.Duration(globals.Config.Quantum) * time.Millisecond)
	globals.Mutex_ColaExec.Lock()
	if Instancia == globals.ContEjecucion && globals.Ejecutando {
		globals.Mutex_ColaExec.Unlock()
		apisSend.DesalojarHilo("Quantum")
		return
	}
	globals.Mutex_ColaExec.Unlock()
}

// ---------------------- Ejecucion --------------------\\

func ReadyToExec() {
	var hilo *globals.Tcb
	if globals.Config.SchedulerAlgorithm == "CMN" {
		hilo = SiguienteMultiColas()
		if hilo == nil {
			if !funcs.ChequearSeguirFuncionando() {
				return
			}
			EsperarDesbloqueo()
			return
		}
	} else if globals.Config.SchedulerAlgorithm == "FIFO" {
		globals.Mutex_ColaReady.Lock()
		if len(globals.ColaReady) == 0 {
			globals.Mutex_ColaReady.Unlock()
			if !funcs.ChequearSeguirFuncionando() {
				return
			}
			EsperarDesbloqueo()
			return
		}
		hilo = globals.ColaReady[0]
		globals.Mutex_ColaReady.Unlock()
	} else { // PRIORIDADES
		hilo = SiguientePrioridades()
		if hilo == nil {
			if !funcs.ChequearSeguirFuncionando() {
				return
			}
			EsperarDesbloqueo()
			return
		}
	}
	MoverHilo(hilo, "EXEC")
	if globals.Config.SchedulerAlgorithm == "CMN" {
		globals.ContEjecucion++
		go WaitQuantum(globals.ContEjecucion)
	}
	apisSend.MandarHiloCPU(hilo.Pcb.Pid, hilo.Tid)
}

func EsperarDesbloqueo() {
	log.Printf("\n## No se han encontrado hilos para ejecutar.\nEsperando un desbloqueo ...")
	for {
		if globals.Config.SchedulerAlgorithm == "CMN" {
			for _, cola := range globals.ColaMultiPrio {
				if len(cola) > 0 {
					go ReadyToExec()
					return
				}
			}
		} else {
			if len(globals.ColaReady) > 0 {
				go ReadyToExec()
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// ------------------ IN/OUT --------------------\\

func InsertarIO(hilo *globals.Tcb, tiempo int) {
	apisSend.DesalojarHilo("IO")
	globals.Mutex_ColaIO.Lock()
	globals.ColaIO = append(globals.ColaIO, &globals.IO_type{Tcb: hilo, Tiempo: tiempo})
	globals.Mutex_ColaIO.Unlock()
}
func IO() {
	globals.IO_Ejecutando = true

	for {
		var hilo *globals.IO_type
		globals.Mutex_ColaIO.Lock()
		if len(globals.ColaIO) > 1 {
			hilo, globals.ColaIO = globals.ColaIO[0], globals.ColaIO[1:]
		} else {
			hilo, globals.ColaIO = globals.ColaIO[0], make([]*globals.IO_type, 0)
		}
		globals.Mutex_ColaIO.Unlock()

		time.Sleep(time.Duration(hilo.Tiempo) * time.Millisecond)
		DesbloquearHilo(hilo.Tcb)

		globals.Mutex_ColaIO.Lock()
		if len(globals.ColaIO) == 0 {
			globals.Mutex_ColaIO.Unlock()
			globals.IO_Ejecutando = false
			break
		}
		globals.Mutex_ColaIO.Unlock()
	}
}

// ------------------ MOVIMIENTO ENTRE COLAS --------------------\\

// DESTINOS: "READY" -- "BLOCK" -- "EXIT" -- "EXEC"
func MoverHilo(hilo *globals.Tcb, destino string) {
	SacarHiloDeCola(hilo)
	hilo.Estado = destino
	switch destino {
	case "READY":
		InsertarHiloReady(hilo)
	case "BLOCK":
		globals.Mutex_ColaBlock.Lock()
		globals.ColaBlock = append(globals.ColaBlock, hilo)
		globals.Mutex_ColaBlock.Unlock()
	case "EXIT":
		globals.Mutex_ColaExit.Lock()
		globals.ColaExit = append(globals.ColaExit, hilo)
		globals.Mutex_ColaExit.Unlock()
	case "EXEC":
		globals.Mutex_ColaExec.Lock()
		globals.ColaExec = hilo
		globals.Mutex_ColaExec.Unlock()
	}
}
func SacarHiloDeCola(hilo *globals.Tcb) {
	switch hilo.Estado {
	case "READY":
		if globals.Config.SchedulerAlgorithm == "CMN" {
			globals.Mutex_Multicola.Lock()
			for i, encolado := range globals.ColaMultiPrio[hilo.Prioridad] {
				if encolado.Tid == hilo.Tid && encolado.Pcb.Pid == hilo.Pcb.Pid {
					globals.ColaMultiPrio[hilo.Prioridad] = append(globals.ColaMultiPrio[hilo.Prioridad][:i], globals.ColaMultiPrio[hilo.Prioridad][i+1:]...)
					break
				}
			}
			globals.Mutex_Multicola.Unlock()
		} else {
			globals.Mutex_ColaReady.Lock()
			for i, encolado := range globals.ColaReady {
				if encolado.Tid == hilo.Tid && encolado.Pcb.Pid == hilo.Pcb.Pid {
					globals.ColaReady = append(globals.ColaReady[:i], globals.ColaReady[i+1:]...)
					break
				}
			}
			globals.Mutex_ColaReady.Unlock()
		}
	case "BLOCK":
		globals.Mutex_ColaBlock.Lock()
		if len(globals.ColaBlock) == 1 {
			globals.ColaBlock = make([]*globals.Tcb, 0)
		} else {
			for i, encolado := range globals.ColaBlock {
				if encolado.Tid == hilo.Tid && encolado.Pcb.Pid == hilo.Pcb.Pid {
					globals.ColaBlock = append(globals.ColaBlock[:i], globals.ColaBlock[i+1:]...)
					break
				}
			}
		}
		globals.Mutex_ColaBlock.Unlock()
	case "EXEC":
		globals.Mutex_ColaExec.Lock()
		if globals.ColaExec.Tid == hilo.Tid && globals.ColaExec.Pcb.Pid == hilo.Pcb.Pid {
			globals.ColaExec = nil
		}
		globals.Mutex_ColaExec.Unlock()
	}
}

func BloquearHilo(hilo *globals.Tcb, motivo string) {
	globals.LogMin("bloq", hilo.Pcb.Pid, hilo.Tid, motivo)
	MoverHilo(hilo, "BLOCK")
}

func DesbloquearHilo(tcb *globals.Tcb) {
	log.Printf("## (%d:%d) Desbloqueado", tcb.Pcb.Pid, tcb.Tid)
	MoverHilo(tcb, "READY")
}
