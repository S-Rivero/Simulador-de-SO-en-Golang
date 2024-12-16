package globals

import (
	"log"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

type ModuleConfig struct {
	Ip              string `json:"ip"`
	Port            int    `json:"port"`
	MemorySize      int    `json:"memory_size"`
	InstructionPath string `json:"instruction_path"`
	ResponseDelay   int    `json:"response_delay"`
	IpKernel        string `json:"ip_kernel"`
	PortKernel      int    `json:"port_kernel"`
	IpCpu           string `json:"ip_cpu"`
	PortCpu         int    `json:"port_cpu"`
	IpFilesystem    string `json:"ip_filesystem"`
	PortFilesystem  int    `json:"port_filesystem"`
	Scheme          string `json:"scheme"`           // FIJAS / DINAMICAS
	SearchAlgorithm string `json:"search_algorithm"` // FIRST / BEST / WORST
	Partitions      []int  `json:"partitions"`
	LogLevel        string `json:"log_level"`
}

var Config *ModuleConfig

// ------------- MEMORIA SISTEMA --------------\\

type ContextoProceso struct {
	Base   uint32                       `json:"base"`
	Limite uint32                       `json:"limite"`
	Hilos  map[int]*globalvar.Registros // La clave del submapa es el TID
}

type Type_MemSis_Contexto struct {
	Pid    int
	Base   uint32
	Limite uint32
	Hilos  []Type_MemSis_Thread
}

type Type_MemSis_Thread struct {
	Tid          int
	Prioridad    int
	Pseudocodigo []string
	Registros    Registros
}

type Registros struct {
	PC uint32 `json:"pc"`
	AX uint32 `json:"ax"`
	BX uint32 `json:"bx"`
	CX uint32 `json:"cx"`
	DX uint32 `json:"dx"`
	EX uint32 `json:"ex"`
	FX uint32 `json:"fx"`
	GX uint32 `json:"gx"`
	HX uint32 `json:"hx"`
}

// ------------- MEMORIA USUARIO --------------\\

type Type_TablaParticiones struct {
	Pid     int
	Base    uint32 `json:"base"`
	Limite  uint32 `json:"limite"`
	Ocupado bool
}
type Type_MemoriaUsuario struct {
	RAM              []byte
	TablaParticiones []Type_TablaParticiones
}

type Type_EspacioLibre struct {
	Base    uint32 `json:"base"`
	Limite  uint32 `json:"limite"`
	Espacio uint32
	Indice  int
}

// ------------- Variables Globales --------------\\

var MemoriaDeSistema = make(map[int]*ContextoProceso)

var MemUs Type_MemoriaUsuario
var MemSis []Type_MemSis_Contexto = make([]Type_MemSis_Contexto, 0)

// MOTIVOS: "proceso" -- "hilo" -- "contexto" -- "instruccion" -- "RAM" -- "dump" _____________
// ESCRITURA = TRUE --> Creación Proc/Hilo, Actualizacion Contexto, Escritura RAM _____________
// ESCRITURA = FALSE --> Destruccion Proc/Hilo, Solicitud Contexto, Lectura RAM _____________
// PARAMS: "proceso" --> param[0]= tamanio _____________
// PARAMS: "instruccion" --> []param= []instruccion _____________
// PARAMS: "RAM" --> param[0]= DirFisica +++ param[1]= tamanio
func LogMin(motivo string, escritura bool, pidInt int, tidInt int, params []string) {
	pid := strconv.Itoa(pidInt)
	tid := strconv.Itoa(tidInt)
	var str string
	switch motivo {
	case "proceso":
		str = "## Proceso "
		if escritura {
			str = str + "Creado "
		} else {
			str = str + "Destruido "
		}
		str = str + "-  PID: " + pid + " - Tamaño: " + params[0]
	case "hilo":
		str = "## Hilo "
		if escritura {
			str = str + "Creado "
		} else {
			str = str + "Destruido "
		}
		str = str + "- (PID:TID) - (" + pid + ":" + tid + ")"
	case "contexto":
		str = "## Contexto "
		if escritura {
			str = str + "Actualizado "
		} else {
			str = str + "Solicitado "
		}
		str = str + "- (PID:TID) - (" + pid + ":" + tid + ")"
	case "instruccion":
		str = "## Obtener instrucción - (PID:TID) - (" + pid + ":" + tid + ") - Instrucción:"
		for _, arg := range params {
			str = str + " " + arg
		}
	case "RAM":
		str = "## Contexto "
		if escritura {
			str = "## Escritura "
		} else {
			str = "## Lectura "
		}
		str = str + "- (PID:TID) - (" + pid + ":" + tid + ") - Dir. Física: " + params[0] + " - Tamaño: " + params[1]
	case "dump":
		str = "## Memory Dump solicitado - (PID:TID) - (" + pid + ":" + tid + ")"
	}

	log.Println(str)
}
