package memUs

import (
	"fmt"
	"log"

	"github.com/sisoputnfrba/tp-golang/memoria/funcs"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
)

func InitRAM() {
	globals.MemUs.RAM = make([]byte, globals.Config.MemorySize)
	if globals.Config.Scheme == "FIJAS" {
		globals.MemUs.TablaParticiones = make([]globals.Type_TablaParticiones, len(globals.Config.Partitions))
		var limiteAnterior uint32 = 0
		for i, particion := range globals.Config.Partitions {
			globals.MemUs.TablaParticiones[i].Pid = -1
			globals.MemUs.TablaParticiones[i].Base = limiteAnterior
			if limiteAnterior == 0 {
				globals.MemUs.TablaParticiones[i].Limite = globals.MemUs.TablaParticiones[i].Base + uint32(particion) - 1
			} else {
				globals.MemUs.TablaParticiones[i].Limite = globals.MemUs.TablaParticiones[i].Base + uint32(particion)
			}
			globals.MemUs.TablaParticiones[i].Ocupado = false
			limiteAnterior = limiteAnterior + uint32(particion+1)
		}
	} else { // DINAMICAS
		globals.MemUs.TablaParticiones = make([]globals.Type_TablaParticiones, 0)
		globals.MemUs.TablaParticiones = append(globals.MemUs.TablaParticiones, globals.Type_TablaParticiones{
			Pid:     -1,
			Base:    0,
			Limite:  uint32(globals.Config.MemorySize) - 1,
			Ocupado: false,
		})
	}
	log.Println("## Memoria Instanciada")
}

// ------------- Asignar Particiones --------------\\

func AsignarParticion(limite uint32, pid int) int {
	espaciosDisponibles := funcs.EspaciosDisponibles(limite)
	if len(espaciosDisponibles) == 0 {
		compactado := SinEspacioLibre(limite)
		if !compactado {
			return -1
		}
		espaciosDisponibles = funcs.EspaciosDisponibles(limite)
	}
	var huecoSeleccionado globals.Type_EspacioLibre
	switch globals.Config.SearchAlgorithm {
	case "FIRST":
		huecoSeleccionado = AsignarFirst(espaciosDisponibles)
	case "BEST":
		huecoSeleccionado = AsignarBest(espaciosDisponibles)
	case "WORST":
		huecoSeleccionado = AsignarWorst(espaciosDisponibles)
	}

	i := huecoSeleccionado.Indice
	if globals.Config.Scheme == "FIJAS" {
		globals.MemUs.TablaParticiones[i].Pid = pid
		globals.MemUs.TablaParticiones[i].Ocupado = true
	} else { // DINAMICAS
		nuevaParticion := globals.Type_TablaParticiones{
			Pid:     pid,
			Base:    huecoSeleccionado.Base,
			Ocupado: true,
			Limite:  huecoSeleccionado.Base + limite,
		}
		if nuevaParticion.Base == 0 && nuevaParticion.Limite > nuevaParticion.Base {
			nuevaParticion.Limite = nuevaParticion.Limite - 1
		}
		if huecoSeleccionado.Espacio > nuevaParticion.Limite-nuevaParticion.Base {
			globals.MemUs.TablaParticiones[i].Base = nuevaParticion.Limite + 1
		}
		globals.MemUs.TablaParticiones = append(globals.MemUs.TablaParticiones, nuevaParticion)
		i = len(globals.MemUs.TablaParticiones) - 1
	}
	return i
}

func AsignarFirst(espaciosDisponibles []globals.Type_EspacioLibre) globals.Type_EspacioLibre {
	return espaciosDisponibles[0]
}
func AsignarBest(espaciosDisponibles []globals.Type_EspacioLibre) globals.Type_EspacioLibre {
	var huecoSeleccionado globals.Type_EspacioLibre = globals.Type_EspacioLibre{Espacio: uint32(globals.Config.MemorySize)}
	for _, hueco := range espaciosDisponibles {
		if hueco.Espacio < huecoSeleccionado.Espacio {
			huecoSeleccionado = hueco
		}
	}
	return huecoSeleccionado
}
func AsignarWorst(espaciosDisponibles []globals.Type_EspacioLibre) globals.Type_EspacioLibre {
	var huecoSeleccionado globals.Type_EspacioLibre = globals.Type_EspacioLibre{Espacio: 0}
	for _, hueco := range espaciosDisponibles {
		if hueco.Espacio > huecoSeleccionado.Espacio {
			huecoSeleccionado = hueco
		}
	}
	return huecoSeleccionado
}

// ------------- Finalizar Proceso --------------\\

func DesocuparParticion(pid int) (uint32, bool) {
	espacio := globals.Type_EspacioLibre{Indice: -1}
	for i, particion := range globals.MemUs.TablaParticiones {
		if particion.Pid == pid {
			globals.MemUs.TablaParticiones[i].Pid = -1
			globals.MemUs.TablaParticiones[i].Ocupado = false
			espacio.Base = particion.Base
			espacio.Limite = particion.Limite
			espacio.Indice = i
		}
	}
	if espacio.Indice == -1 {
		return 0, false
	}
	limite := espacio.Limite
	if globals.Config.Scheme == "DINAMICAS" {
		JuntarParticiones(espacio)
	}
	return limite, true
}

// ------------- Adicional --------------\\

func SinEspacioLibre(limite uint32) bool {
	if globals.Config.Scheme == "FIJAS" {
		return false
	}
	espaciosLibres := funcs.EspaciosLibres()
	if len(espaciosLibres) < 2 {
		return false
	}
	var huecoTotal uint32 = 0
	for _, hueco := range espaciosLibres {
		huecoTotal = huecoTotal + hueco.Espacio
	}
	if huecoTotal < limite {
		return false
	}
	CompactarMemoria()
	return true
}
func CompactarMemoria() error {
	ocupados := funcs.EspaciosOcupados()
	if len(ocupados) == 0 {
		return nil
	}

	totalSize := uint32(0)
	for _, espacio := range ocupados {
		size := espacio.Limite - espacio.Base
		if totalSize+size > uint32(globals.Config.MemorySize) {
			return fmt.Errorf("no hay suficiente espacio para compactar: necesario %d, disponible %d",
				totalSize+size, globals.Config.MemorySize)
		}
		totalSize += size
	}

	ocupados_bytes := funcs.EspaciosOcupadosEnRAM(ocupados)
	globals.MemUs.TablaParticiones = make([]globals.Type_TablaParticiones, 0)

	var offsetActual uint32 = 0
	for i, espacio := range ocupados {
		tamanio := espacio.Limite - espacio.Base

		reorganizado := globals.Type_TablaParticiones{
			Pid:     espacio.Pid,
			Ocupado: true,
			Base:    offsetActual,
			Limite:  offsetActual + tamanio,
		}

		if reorganizado.Limite > uint32(globals.Config.MemorySize) {
			return fmt.Errorf("error al compactar: límite %d excede tamaño de memoria %d",
				reorganizado.Limite, globals.Config.MemorySize)
		}

		globals.MemUs.TablaParticiones = append(globals.MemUs.TablaParticiones, reorganizado)
		funcs.MoverRAM(reorganizado, espacio, ocupados_bytes[i])

		offsetActual = reorganizado.Limite + 1
	}

	if offsetActual < uint32(globals.Config.MemorySize) {
		globals.MemUs.TablaParticiones = append(globals.MemUs.TablaParticiones, globals.Type_TablaParticiones{
			Pid:     -1,
			Base:    offsetActual,
			Limite:  uint32(globals.Config.MemorySize) - 1,
			Ocupado: false,
		})
	}

	return nil
}

func JuntarParticiones(espacio globals.Type_EspacioLibre) {
	nuevaParticion := globals.Type_TablaParticiones{
		Pid:     -1,
		Base:    espacio.Base,
		Limite:  espacio.Limite,
		Ocupado: false,
	}
	globals.MemUs.TablaParticiones = append(globals.MemUs.TablaParticiones[:espacio.Indice], globals.MemUs.TablaParticiones[espacio.Indice+1:]...)
	for i := 0; i < len(globals.MemUs.TablaParticiones); i++ {
		if !globals.MemUs.TablaParticiones[i].Ocupado && (globals.MemUs.TablaParticiones[i].Base == nuevaParticion.Limite+1 || globals.MemUs.TablaParticiones[i].Limite+1 == nuevaParticion.Base) {
			if nuevaParticion.Base < globals.MemUs.TablaParticiones[i].Base {
				nuevaParticion.Limite = globals.MemUs.TablaParticiones[i].Limite
			} else {
				nuevaParticion.Base = globals.MemUs.TablaParticiones[i].Base
			}
			globals.MemUs.TablaParticiones = append(globals.MemUs.TablaParticiones[:i], globals.MemUs.TablaParticiones[i+1:]...)
			i--
		}
	}
	globals.MemUs.TablaParticiones = append(globals.MemUs.TablaParticiones, nuevaParticion)
}
