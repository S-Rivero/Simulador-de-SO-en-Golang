package memSis

import (
	"strconv"

	"github.com/sisoputnfrba/tp-golang/memoria/funcs"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

func CrearEstructuraProceso(particion globals.Type_TablaParticiones) {
	globals.MemSis = append(globals.MemSis, globals.Type_MemSis_Contexto{
		Pid:    particion.Pid,
		Base:   particion.Base,
		Limite: particion.Limite,
		Hilos:  make([]globals.Type_MemSis_Thread, 0),
	})
}

func EliminarEstructuraProceso(pid int) bool {
	//log.Printf("#### Pid: %d --- Memsis: %v", pid, globals.MemSis)
	i := funcs.BuscarEstructuraProceso(pid)
	if i == -1 {
		return false
	}
	if len(globals.MemSis) == 1 {
		globals.MemSis = make([]globals.Type_MemSis_Contexto, 0)
	} else if len(globals.MemSis) == i+1 {
		globals.MemSis = globals.MemSis[:i]
	} else {
		globals.MemSis = append(globals.MemSis[:i], globals.MemSis[i+1:]...)
	}
	return true
}

func CrearEstructuraHilo(pid int, tid int, prio int, filename string) string {
	i := funcs.BuscarEstructuraProceso(pid)
	if i == -1 {
		return "Error: No se a encontrado la estructura del Proceso (" + strconv.Itoa(pid) + ") en Memoria de Sistema"
	}

	tcb := globals.Type_MemSis_Thread{
		Tid:       tid,
		Prioridad: prio,
		Registros: globals.Registros{
			PC: 0,
			AX: 0,
			BX: 0,
			CX: 0,
			DX: 0,
			EX: 0,
			FX: 0,
			GX: 0,
			HX: 0,
		},
	}

	tcb.Pseudocodigo = funcs.LeerArchivoPseudocodigo(filename)
	if len(tcb.Pseudocodigo) == 0 {
		return "Error: No se ha podido leer el archivo de pseudocodigo " + filename
	}

	globals.MemSis[i].Hilos = append(globals.MemSis[i].Hilos, tcb)
	return ""
}

func EliminarEstructuraHilo(pid int, tid int) string {
	i := funcs.BuscarEstructuraProceso(pid)
	if i == -1 {
		return "Error: No se a encontrado la estructura del Proceso (" + strconv.Itoa(pid) + ") en Memoria de Sistema"
	}
	j := funcs.BuscarEstructuraHilo(i, tid)
	if j == -1 {
		return "Error: No se a encontrado la estructura del Hilo (" + strconv.Itoa(pid) + ":" + strconv.Itoa(tid) + ") en Memoria de Sistema"
	}
	if len(globals.MemSis[i].Hilos) == 1 {
		globals.MemSis[i].Hilos = make([]globals.Type_MemSis_Thread, 0)
	} else if len(globals.MemSis[i].Hilos) == j+1 {
		globals.MemSis[i].Hilos = globals.MemSis[i].Hilos[:j]
	} else {
		globals.MemSis[i].Hilos = append(globals.MemSis[i].Hilos[:j], globals.MemSis[i].Hilos[j+1:]...)
	}
	return ""
}

func ActualizarContextoHilo(contexto globalvar.ContextoEjecucion) string {

	i := funcs.BuscarEstructuraProceso(contexto.Pid)
	if i == -1 {
		return "Error: No se a encontrado la estructura del Proceso (" + strconv.Itoa(contexto.Pid) + ") en Memoria de Sistema"
	}
	j := funcs.BuscarEstructuraHilo(i, contexto.Tid)
	if j == -1 {
		return "Error: No se a encontrado la estructura del Hilo (" + strconv.Itoa(contexto.Pid) + ":" + strconv.Itoa(contexto.Tid) + ") en Memoria de Sistema"
	}

	globals.MemSis[i].Hilos[j].Registros = funcs.ContextoCpuToMem(contexto)

	return ""
}
