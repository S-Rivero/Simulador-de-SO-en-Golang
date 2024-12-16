package apis

import (
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/memoria/apisSend"
	"github.com/sisoputnfrba/tp-golang/memoria/cpu"
	"github.com/sisoputnfrba/tp-golang/memoria/funcs"
	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	"github.com/sisoputnfrba/tp-golang/memoria/memSis"
	"github.com/sisoputnfrba/tp-golang/memoria/memUs"
	"github.com/sisoputnfrba/tp-golang/utils/commons"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

// ------------- Kernel --------------\\

func Handler_InicializarProceso(w http.ResponseWriter, r *http.Request) {
	var req globalvar.Paquete_Tamanio
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	i := memUs.AsignarParticion(uint32(req.Tamanio), req.Pid)
	if i == -1 {
		log.Printf("No se pudo asignar partición para proceso %d: memoria insuficiente", req.Pid)
		http.Error(w, "Error: No hay memoria disponible", http.StatusInsufficientStorage)
		return
	}

	memSis.CrearEstructuraProceso(globals.MemUs.TablaParticiones[i])

	tamLog := make([]string, 0)
	tamLog = append(tamLog, strconv.Itoa(req.Tamanio))
	globals.LogMin("proceso", true, req.Pid, 0, tamLog)
	w.WriteHeader(http.StatusOK)
}

func Handler_FinalizarProceso(w http.ResponseWriter, r *http.Request) {
	var req globalvar.Paquete_Pid
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	particion := funcs.BuscarParticionPorPID(req.Pid)
	if particion.Pid == -1 {
		commons.ResponderError(w, "Error: No se a encontrado la Particion del Proceso "+strconv.Itoa(req.Pid))
		return
	}
	tamLog := make([]string, 0)
	if particion.Base == 0 {
		tamLog = append(tamLog, strconv.Itoa(int(particion.Limite-particion.Base)+1))
	} else {
		tamLog = append(tamLog, strconv.Itoa(int(particion.Limite-particion.Base)))
	}

	limite, resp := memUs.DesocuparParticion(req.Pid)
	if !resp {
		commons.ResponderError(w, "Error: No se logro desocupar la Particion del Proceso "+strconv.Itoa(req.Pid))
		return
	}

	resp = memSis.EliminarEstructuraProceso(req.Pid)
	if !resp {
		commons.ResponderError(w, "Error: No se a encontrado la estructura del Proceso ("+strconv.Itoa(req.Pid)+") en Memoria de Sistema")
		memUs.AsignarParticion(limite, req.Pid)
		return
	}

	globals.LogMin("proceso", false, req.Pid, 0, tamLog)
	w.WriteHeader(http.StatusOK)
}

func Handler_InicializarHilo(w http.ResponseWriter, r *http.Request) {
	var req globalvar.Paquete_CrearHilo
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	res := memSis.CrearEstructuraHilo(req.Pid, req.Tid, req.Prioridad, req.Filename)

	if res != "" {
		commons.ResponderError(w, res)
		return
	}
	globals.LogMin("hilo", true, req.Pid, req.Tid, make([]string, 0))
	w.WriteHeader(http.StatusOK)
}

func Handler_FinalizarHilo(w http.ResponseWriter, r *http.Request) {
	var req globalvar.Paquete_PidTid
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	res := memSis.EliminarEstructuraHilo(req.Pid, req.Tid)
	if res != "" {
		commons.ResponderError(w, res)
		return
	}
	globals.LogMin("hilo", false, req.Pid, req.Tid, make([]string, 0))
	w.WriteHeader(http.StatusOK)
}
func Handler_DumpMemory(w http.ResponseWriter, r *http.Request) {
	var req globalvar.Paquete_PidTid
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	res := apisSend.EnviarDumpAFileSystem(req.Pid, req.Tid)

	if res != "" {
		commons.ResponderError(w, res)
		return
	}
	globals.LogMin("dump", false, req.Pid, req.Tid, make([]string, 0))
	w.WriteHeader(http.StatusOK)
}

// ------------- CPU --------------\\

func Handler_SolicitudContexto(w http.ResponseWriter, r *http.Request) {
	funcs.AplicarRetardo()
	var req globalvar.Paquete_PidTid
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	contexto, err := cpu.DevolverContexto(req.Pid, req.Tid)
	if err != "" {
		commons.ResponderError(w, err)
		return
	}

	globals.LogMin("contexto", false, req.Pid, req.Tid, make([]string, 0))
	w.WriteHeader(http.StatusOK)
	commons.ResponderPost[globalvar.ContextoEjecucion](w, contexto)
}

func Handler_SolicitudInstruccion(w http.ResponseWriter, r *http.Request) {
	funcs.AplicarRetardo()
	var req globalvar.Pedir_instruccion_memoria
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	instruccion, err2 := cpu.BuscarInstruccion(req.Pid, req.Tid, req.Program_counter)

	if len(instruccion) == 0 {
		commons.ResponderError(w, err2)
		return
	}
	paquete := globalvar.Paquete_Instruccion{
		Instruccion: instruccion,
	}
	globals.LogMin("instruccion", false, req.Pid, req.Tid, instruccion)
	w.WriteHeader(http.StatusOK)
	commons.ResponderPost[globalvar.Paquete_Instruccion](w, paquete)
}

func Handler_LeerMemoria(w http.ResponseWriter, r *http.Request) {
	funcs.AplicarRetardo()
	var req globalvar.LeerMemoriaRequest
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}

	valor := cpu.LeerRAM_Instruccion(req.Direccion)
	paquete := globalvar.LeerMemoriaResponse{Valor: valor}

	// falta tid + Ver como pasamos dir fisica y tamanio
	//globals.LogMin("instruccion", false, req.Pid, , instruccion)
	w.WriteHeader(http.StatusOK)
	commons.ResponderPost[globalvar.LeerMemoriaResponse](w, paquete)
}

func Handler_EscribirMemoria(w http.ResponseWriter, r *http.Request) {
	funcs.AplicarRetardo()
	var req globalvar.EscribirMemoriaRequest
	if commons.DecodificarJSON(w, r, &req) != nil {
		return
	}
	cpu.EscribirEnRAM_Instruccion(req.Direccion, req.Valor)

	// falta tid + Ver como pasamos dir fisica y tamanio
	//globals.LogMin("instruccion", false, req.Pid, , instruccion)
	w.WriteHeader(http.StatusOK)
}

func Handler_ActualizarContexto(w http.ResponseWriter, r *http.Request) {
	funcs.AplicarRetardo()
	var contexto globalvar.ContextoEjecucion
	if commons.DecodificarJSON(w, r, &contexto) != nil {
		return
	}

	err := memSis.ActualizarContextoHilo(contexto)
	if err != "" {
		commons.ResponderError(w, err)
		return
	}
	globals.LogMin("contexto", true, contexto.Pid, contexto.Tid, make([]string, 0))
	w.WriteHeader(http.StatusOK)
}
