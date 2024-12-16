package globalvar

//------------ Structs Generales -----------\\

type Paquete_PidTid struct {
	Pid int `json:"pid"`
	Tid int `json:"tid"`
}
type Paquete_Pid struct {
	Pid int `json:"pid"`
}

type Paquete_Motivo struct {
	Motivo string `json:"motivo"`
}

type PaqueteError struct {
	Mensaje string `json:"mensaje"`
}

//------------ Structs CPU-MEMORIA -----------\\

type Instruccion struct {
	Instruccion string `json:"instruccion"`
}
type Paquete_Instruccion struct {
	Instruccion []string `json:"instruccion"`
}

type Pedir_instruccion_memoria struct {
	Pid             int    `json:"pid"`
	Tid             int    `json:"tid"`
	Program_counter uint32 `json:"program_counter"`
}

type Path_proceso struct {
	Tid  int    `json:"tid"`
	Path string `json:"path"`
}
type ContextoEjecucion struct {
	Tid       int
	Pid       int
	Registros Registros
}
type Registros struct {
	PC     uint32 `json:"pc"`
	AX     uint32 `json:"ax"`
	BX     uint32 `json:"bx"`
	CX     uint32 `json:"cx"`
	DX     uint32 `json:"dx"`
	EX     uint32 `json:"ex"`
	FX     uint32 `json:"fx"`
	GX     uint32 `json:"gx"`
	HX     uint32 `json:"hx"`
	Base   uint32 `json:"base"`
	Limite uint32 `json:"limite"`
}

// ------------ Structs KERNEL-MEMORIA -----------\\
type Paquete_Tamanio struct {
	Pid     int `json:"pid"`
	Tamanio int `json:"tamanio"`
}
type Paquete_CrearHilo struct {
	Pid       int
	Tid       int
	Prioridad int
	Filename  string
}

// ------------ Structs MEMORIA-FILESYSTEM -----------\\

type DumpToFS_Type struct {
	Pid   int    `json:"pid"`
	Tid   int    `json:"tid"`
	Bytes []byte `json:"content"`
	Size  int    `json:"size"`
}

//----------------- Structs para Syscalls ----------------\\

type Request_PROCESS_CREATE struct {
	NombreArchivo  string `json:"archivo"`
	Tamanio        int    `json:"tamanio"`
	PrioridadHilo0 int    `json:"prioridad"`
}

type Request_PROCESS_EXIT struct{}

type Request_THREAD_CREATE struct {
	NombreArchivo string `json:"archivo"`
	Prioridad     int    `json:"prioridad"`
}

type Request_THREAD_JOIN struct {
	Tid int `json:"Tid"`
}
type Request_THREAD_CANCEL struct {
	Tid int `json:"tid"`
}

type Request_THREAD_EXIT struct{}

type Request_MUTEX_CREATE struct {
	Recurso string `json:"recurso"`
}

type Request_MUTEX_LOCK struct {
	Recurso string `json:"recurso"`
}

type Request_MUTEX_UNLOCK struct {
	Recurso string `json:"recurso"`
}

type Request_DUMP_MEMORY struct{}

type Request_IO struct {
	Tiempo int `json:"tiempo"`
}

//----------------- Structs para Comunicacion Memoria - CPU ----------------\\

type LeerMemoriaRequest struct {
	Direccion int `json:"direccion"`
}

type LeerMemoriaResponse struct {
	Valor uint32 `json:"valor"`
}

type EscribirMemoriaRequest struct {
	Direccion int    `json:"direccion"`
	Valor     uint32 `json:"valor"`
}
