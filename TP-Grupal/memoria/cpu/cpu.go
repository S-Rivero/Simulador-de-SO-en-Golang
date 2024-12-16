package cpu

import (
	"encoding/binary"
	"strings"

	"github.com/sisoputnfrba/tp-golang/memoria/funcs"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

func DevolverContexto(pid int, tid int) (globalvar.ContextoEjecucion, string) {
	i := funcs.BuscarEstructuraProceso(pid)
	if i == -1 {
		return globalvar.ContextoEjecucion{Pid: -1}, "Error: No se a encontrado la estructura del Proceso en Memoria de Sistema"
	}
	j := funcs.BuscarEstructuraHilo(i, tid)
	if j == -1 {
		return globalvar.ContextoEjecucion{Pid: -1}, "Error: No se a encontrado la estructura del Hilo en Memoria de Sistema"
	}
	return funcs.ContextoMemToCpu(i, j), ""
}

func BuscarInstruccion(pid int, tid int, PC uint32) ([]string, string) {
	i := funcs.BuscarEstructuraProceso(pid)
	if i == -1 {
		return make([]string, 0), "Error: No se a encontrado la estructura del Proceso en Memoria de Sistema"
	}
	j := funcs.BuscarEstructuraHilo(i, tid)
	if j == -1 {
		return make([]string, 0), "Error: No se a encontrado la estructura del Hilo en Memoria de Sistema"
	}
	return strings.Split(globals.MemSis[i].Hilos[j].Pseudocodigo[PC], " "), ""
}

/*
	func EscribirEnRAM_Instruccion(pid int, offset uint32, valor uint32) string {
		particion := funcs.BuscarParticionPorPID(pid)
		if particion.Pid == -1 {
			return "Error: No se a encontrado la particion del Proceso"
		}

		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, valor)
		copy(globals.MemUs.RAM[int(particion.Base+offset):int(particion.Base+offset+4)], bytes)

		return ""
	}

	func LeerRAM_Instruccion(pid int, offset uint32) (uint32, string) {
		bytes := funcs.ContenidoEnRAM(pid)
		if len(bytes) == 0 {
			return 0, "Error: No se a encontrado la particion del Proceso"
		}
		return binary.LittleEndian.Uint32(bytes[int(offset) : int(offset)+4]), ""
	}
*/
func EscribirEnRAM_Instruccion(dir int, valor uint32) {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, valor)
	copy(globals.MemUs.RAM[dir:dir+4], bytes)
}
func LeerRAM_Instruccion(dir int) uint32 {
	return binary.LittleEndian.Uint32(globals.MemUs.RAM[dir : dir+4])
}
