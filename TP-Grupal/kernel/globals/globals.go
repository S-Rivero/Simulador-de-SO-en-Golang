package globals

import (
	"log"
	"strconv"
	"sync"
)

type ModuleConfig struct {
	IpMemory           string `json:"ip_memory"`
	PortMemory         int    `json:"port_memory"`
	IpCpu              string `json:"ip_cpu"`
	PortCpu            int    `json:"port_cpu"`
	SchedulerAlgorithm string `json:"scheduler_algorithm"`
	Quantum            int    `json:"quantum"`
	LogLevel           string `json:"log_level"`
	Port               int    `json:"port"`
	Ip                 string `json:"ip"`
}

var HayPcbEnColaNew = make(chan int)

var Config *ModuleConfig

// ------------- Structs --------------\\

type Pcb struct {
	Pid          int
	ContadorTids int
	Tids         []*Tcb
	Mutexs       []Mutex_type
	Estado       string
}
type Mutex_type struct {
	NombreMutex string
	Estado      bool
	Tid         int
	Bloqueados  []int
}
type NewProc struct {
	Pcb            *Pcb
	PrioridadHilo0 int
	Tamanio        int
	Filename       string
}
type Tcb struct {
	Tid             int    `json:"tid"`
	Prioridad       int    `json:"prioridad"`
	Estado          string `json:"estado"`
	Pcb             *Pcb   `json:"pid"`
	Filename        string `json:"Filename"`
	HilosBloqueados []*Tcb `json:"hilosbloqueados"`
}
type IO_type struct {
	Tcb    *Tcb `json:"tcb"`
	Tiempo int  `json:"tiempo"`
}

// ------------- Colas --------------\\

var ColaNew []*NewProc = []*NewProc{}
var ColaReady []*Tcb = []*Tcb{}
var ColaProcesos []*Pcb = []*Pcb{}      // Procesos activos (No Exit)
var ColaMultiPrio [][]*Tcb = [][]*Tcb{} // ColaMultiPrio [Prioridad][Hilos]
var ColaExec *Tcb = nil
var ColaBlock []*Tcb = []*Tcb{}
var ColaExit []*Tcb = []*Tcb{}
var ColaIO []*IO_type = []*IO_type{}

// ------------- Variables Globales --------------\\

var ContProcesos int = 0
var PrioridadMinima int = 0
var Ejecutando bool = false
var IO_Ejecutando bool = false
var Inicial bool = true
var ContEjecucion int = 0

// ------------- Mutex --------------\\

var Mutex_ColaNew sync.Mutex
var Mutex_ColaProcesos sync.Mutex
var Mutex_ColaReady sync.Mutex
var Mutex_ColaExec sync.Mutex
var Mutex_ColaBlock sync.Mutex
var Mutex_ColaExit sync.Mutex
var Mutex_ColaIO sync.Mutex
var Mutex_Multicola sync.Mutex

// ------------- Semaforos --------------\\

var SemaforoColaNew = make(chan struct{}, 1)   // Semáforo para NEW queue
var SemaforoColaReady = make(chan struct{}, 1) // Semáforo para READY queue

// ------------- Logs Minimos y Obligatorios --------------\\

// MOTIVOS: "syscall" -- "crearProc" -- "crearHilo" -- "bloq" -- "finIO" -- "quantum" -- "finProc" --  "finHilo"
// __________ syscall y bloq tiene param string
func LogMin(motivo string, pidInt int, tidInt int, param string) {
	pid := strconv.Itoa(pidInt)
	var str string
	if motivo == "finProc" {
		str = "## Finaliza el proceso " + pid
	} else {
		tid := strconv.Itoa(tidInt)
		str = "## (" + pid + ":" + tid + ") "
		switch motivo {
		case "syscall":
			str += "- Solicitó syscall: " + param
		case "crearProc":
			str += "Se crea el proceso - Estado: NEW"
		case "crearHilo":
			str += "Se crea el Hilo - Estado: Ready"
		case "bloq": // "Motivo de bloqueo"
			str += "- Bloqueado por: " + param // <PTHREAD_JOIN / MUTEX / IO / DUMP_MEMORY>
		case "finIO":
			str += "Finalizó IO y pasa a READY"
		case "quantum":
			str += "- Desalojado por fin de Quantum"
		case "finHilo":
			str += "Finaliza el hilo"
		}
	}
	log.Println(str)
}
