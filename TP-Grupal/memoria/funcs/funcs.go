package funcs

import (
	"encoding/binary"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

// ------------- MANEJO MEMORIA USUARIO --------------\\

func EspaciosDisponibles(limite uint32) []globals.Type_EspacioLibre {
	lista := EspaciosLibres()
	huecosDisponibles := make([]globals.Type_EspacioLibre, 0)
	for _, hueco := range lista {
		if hueco.Espacio >= limite {
			huecosDisponibles = append(huecosDisponibles, hueco)
		}
	}
	return huecosDisponibles
}
func EspaciosLibres() []globals.Type_EspacioLibre {
	lista := make([]globals.Type_EspacioLibre, 0)
	for i, particion := range globals.MemUs.TablaParticiones {
		if !particion.Ocupado {
			hueco := globals.Type_EspacioLibre{
				Base:   particion.Base,
				Limite: particion.Limite,
				Indice: i,
			}
			if particion.Base == 0 {
				hueco.Espacio = particion.Limite + 1
			} else {
				hueco.Espacio = particion.Limite - particion.Base
			}

			lista = append(lista, hueco)
		}
	}
	return lista
}
func EspaciosOcupados() []globals.Type_TablaParticiones {
	lista := make([]globals.Type_TablaParticiones, 0)
	for _, particion := range globals.MemUs.TablaParticiones {
		if particion.Ocupado {
			lista = append(lista, particion)
		}
	}
	return lista
}
func EspaciosOcupadosEnRAM(ocupados []globals.Type_TablaParticiones) [][]byte {
	lista := make([][]byte, 0)
	for _, espacio := range ocupados {
		bytes := globals.MemUs.RAM[int(espacio.Base):int(espacio.Limite)]
		lista = append(lista, bytes)
	}
	return lista
}
func MoverRAM(particionNueva globals.Type_TablaParticiones, particionVieja globals.Type_TablaParticiones, bytes_ocupados []byte) {
	copy(globals.MemUs.RAM[int(particionVieja.Base):int(particionVieja.Limite)], make([]byte, int(particionVieja.Limite)-int(particionVieja.Base)))
	copy(globals.MemUs.RAM[int(particionNueva.Base):int(particionNueva.Limite)], bytes_ocupados)
}
func InsertarRAM(particion globals.Type_TablaParticiones, bytes []byte) {
	copy(globals.MemUs.RAM[int(particion.Base):int(particion.Limite)], bytes)
}

func ContenidoEnRAM(pid int) []byte {
	particion := BuscarParticionPorPID(pid)
	if particion.Pid == -1 {
		return make([]byte, 0)
	}
	return globals.MemUs.RAM[int(particion.Base):int(particion.Limite)]
}
func BuscarParticionPorPID(pid int) globals.Type_TablaParticiones {
	for _, particion := range globals.MemUs.TablaParticiones {
		if particion.Pid == pid {
			return particion
		}
	}
	return globals.Type_TablaParticiones{
		Pid: -1,
	}
}

// ------------- MANEJO MEMORIA SISTEMA --------------\\

func BuscarEstructuraProceso(pid int) int {
	for i, pcb := range globals.MemSis {
		if pcb.Pid == pid {
			return i
		}
	}
	return -1
}
func BuscarEstructuraHilo(i int, tid int) int {
	for j, tcb := range globals.MemSis[i].Hilos {
		if tcb.Tid == tid {
			return j
		}
	}
	return -1
}

func LeerArchivoPseudocodigo(filename string) []string {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("Error obteniendo directorio actual: %v", err)
		return make([]string, 0)
	}

	// Subir un nivel para llegar a la raíz del proyecto
	projectRoot := filepath.Dir(currentDir)
	testPath := filepath.Join(projectRoot, globals.Config.InstructionPath, filename)

	log.Printf("Buscando archivo en: %s", testPath)

	bytes, err := os.ReadFile(testPath)
	if err != nil {
		log.Printf("Error: No se a podido leer el archivo de pseudocodigo %s", testPath)
		return make([]string, 0)
	}

	content := string(bytes)
	return strings.Split(content, "\n")
}

// ------------- Funciones para comunicación con CPU --------------\\

// Se escriben los primeros 4 bytes a partir de la dirección física dada en la memoria de usuario
func EscribirMemoria(direccion uint32, valor uint32) error {
	// Verifica si la dirección está dentro de los límites de la memoria de usuario
	if direccion+4 > uint32(len(globals.MemUs.RAM)) {
		return errors.New("dirección fuera de los límites")
	}

	// Escribir los primeros 4 bytes a partir de la dirección dada
	binary.LittleEndian.PutUint32(globals.MemUs.RAM[direccion:direccion+4], valor)
	return nil
}

func AplicarRetardo() {
	time.Sleep(time.Duration(globals.Config.ResponseDelay) * time.Millisecond)
}

func ContextoMemToCpu(i int, j int) globalvar.ContextoEjecucion {
	return globalvar.ContextoEjecucion{
		Pid: globals.MemSis[i].Pid,
		Tid: globals.MemSis[i].Hilos[j].Tid,
		Registros: globalvar.Registros{
			PC:     globals.MemSis[i].Hilos[j].Registros.PC,
			AX:     globals.MemSis[i].Hilos[j].Registros.AX,
			BX:     globals.MemSis[i].Hilos[j].Registros.BX,
			CX:     globals.MemSis[i].Hilos[j].Registros.CX,
			DX:     globals.MemSis[i].Hilos[j].Registros.DX,
			EX:     globals.MemSis[i].Hilos[j].Registros.EX,
			FX:     globals.MemSis[i].Hilos[j].Registros.FX,
			GX:     globals.MemSis[i].Hilos[j].Registros.GX,
			HX:     globals.MemSis[i].Hilos[j].Registros.HX,
			Base:   globals.MemSis[i].Base,
			Limite: globals.MemSis[i].Limite,
		},
	}
}

func ContextoCpuToMem(contexto globalvar.ContextoEjecucion) globals.Registros {
	return globals.Registros{
		PC: contexto.Registros.PC,
		AX: contexto.Registros.AX,
		BX: contexto.Registros.BX,
		CX: contexto.Registros.CX,
		DX: contexto.Registros.DX,
		EX: contexto.Registros.EX,
		FX: contexto.Registros.FX,
		GX: contexto.Registros.GX,
		HX: contexto.Registros.HX,
	}
}
