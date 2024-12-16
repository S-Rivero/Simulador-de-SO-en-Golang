package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/commons"
	"github.com/sisoputnfrba/tp-golang/utils/globalvar"
)

type ModuleConfig struct {
	Ip         string `json:"ip"`
	IpMemory   string `json:"ip_memory"`
	PortMemory int    `json:"port_memory"`
	IpKernel   string `json:"ip_kernel"`
	PortKernel int    `json:"port_kernel"`
	Port       int    `json:"port"`
	LogLevel   string `json:"log_level"`
}

var Config *ModuleConfig
var Registros *globalvar.Registros
var HayInterrupcion bool = false
var MotivoInterrupcion string = ""

func RecibirHilo(w http.ResponseWriter, r *http.Request) {
	MotivoInterrupcion = ""
	var paquete globalvar.Paquete_PidTid

	// Decodifica el hilo enviado por el Kernel
	err := json.NewDecoder(r.Body).Decode(&paquete)
	if err != nil {
		//http.Error(w, "Error al decodificar PID/TID", http.StatusBadRequest)
		commons.ResponderError(w, "Error al decodificar PID/TID")
		return
	}

	// Mostrar los valores recibidos
	log.Printf("## Hilo recibido TID: %d - PID: %d\n", paquete.Tid, paquete.Pid)

	//confirmar la recepción
	responseMessage := fmt.Sprintf("TID %d y PID %d recibidos correctamente", paquete.Tid, paquete.Pid)

	// Llama a la función para solicitar el contexto de ejecución a la Memoria

	//log.Printf("## PID: %d - Recibiendo Contexto de ejecucion", paquete.Tid)

	contexto := SolicitarContextoMemoria(paquete.Pid, paquete.Tid)
	if contexto == nil {
		commons.ResponderError(w, "Error al obtener el contexto de memoria")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseMessage))
	//log.Printf("## TID: %d - Contexto de ejecucion recibido: %+v", paquete.Tid, contexto)

	CicloDeInstruccion(contexto)

	//PedirInstruccionMemoria(paquete.Pid, paquete.Tid, contexto.Registros.PC) //(PARA PROBAR)
	//EjecutarProceso(contexto)                                                //(PARA PROBAR)
}

func SolicitarContextoMemoria(pid int, tid int) *globalvar.ContextoEjecucion {
	paquete := globalvar.Paquete_PidTid{
		Pid: pid,
		Tid: tid,
	}
	resp, err := commons.EnviarPaqueteYEsperar[globalvar.Paquete_PidTid, globalvar.ContextoEjecucion](paquete, "MEMORIA", "SolicitudContexto")
	if err != nil {
		log.Printf("Error enviando la solicitud de contexto")
		return nil
	}

	log.Printf("## TID: %d - Solicito Contexto Ejecución", tid)

	return &resp
}

func PedirInstruccionMemoria(pidProceso int, tidProceso int, pcProceso uint32) ([]string, error) {
	paquete := globalvar.Pedir_instruccion_memoria{
		Pid:             pidProceso,
		Tid:             tidProceso,
		Program_counter: pcProceso,
	}
	resp, err := commons.EnviarPaqueteYEsperar[globalvar.Pedir_instruccion_memoria, globalvar.Paquete_Instruccion](paquete, "MEMORIA", "SolicitudInstruccion")
	if err != nil {
		log.Printf("Error enviando la solicitud de contexto")
		return make([]string, 0), err
	}
	return resp.Instruccion, nil
}

func CicloDeInstruccion(contexto *globalvar.ContextoEjecucion) {

	for {
		// Fetch desde la memoria
		instruccion, err := Fetch(contexto)
		if err != nil {
			log.Println("Error en Fetch:", err)
			break
		}

		// Decode
		var str string = "Instrucción obtenida: "
		for _, arg := range instruccion {
			str = str + " " + arg
		}
		log.Println(str)

		// Execute
		err2 := EjecutarProceso(contexto, instruccion)
		if !err2 {
			log.Println("Error en la ejecución")
			break
		}

		if CheckInterrupt(*contexto) {
			break
		}
	}
}

func Fetch(unProceso *globalvar.ContextoEjecucion) ([]string, error) {

	pid_proceso := unProceso.Pid
	tid_proceso := unProceso.Tid
	pc_proceso := unProceso.Registros.PC

	log.Printf("## TID: %d PID: %d- FETCH - Program counter: %d", tid_proceso, pid_proceso, pc_proceso)

	fetch, err := PedirInstruccionMemoria(pid_proceso, tid_proceso, pc_proceso)
	if err != nil {
		log.Println("Error al pedir instruccion a memoria")
		return make([]string, 0), err
	}

	return fetch, nil

}

// EXECUTE
func EjecutarProceso(unProceso *globalvar.ContextoEjecucion, palabras []string) bool {

	//palabras, err := Fetch(unProceso)

	/*
		if err != nil {
			log.Println("Error al obtener instruccion")
			return false
		}
	*/

	//palabras es un slice con cada palabra de "instruccion" en una celda distinta
	//separadas por el espacio " "
	//palabras := strings.Split(instruccion, " ")
	switch palabras[0] {

	case "SET":

		unProceso.Registros.PC++

		registro := palabras[1]

		valor := strings.TrimRight(palabras[2], "\n") //le saca el espacio a la celda

		log.Printf("TID: %d - Ejecutando: SET - %s %s", unProceso.Tid, registro, valor)

		valorInt, err := strconv.Atoi(valor) //convierte el tercer valor de "palabras" de string a int

		if err != nil {
			fmt.Println("Error al convertir el string a entero:", err)
			return false
		}

		valorInt32 := uint32(valorInt)

		SetearValorRegistro(registro, unProceso, valorInt32)

		//log.Printf("Registros: %+v", unProceso.Registros) //BORRAR

		log.Printf("TID: %d - Ejecutado SET - %s = %d", unProceso.Tid, registro, valorInt32)

	case "SUM":

		unProceso.Registros.PC++

		registroDestino := palabras[1]
		registroOrigen := strings.TrimRight(palabras[2], "\n")

		log.Printf("TID: %d - Ejecutando: SUM - %s %s", unProceso.Tid, registroDestino, registroOrigen)

		valorDestino, err := ObtenerValorDeRegistro(registroDestino, unProceso)

		if err != nil {
			return false
		}

		valorOrigen, err := ObtenerValorDeRegistro(registroOrigen, unProceso)
		if err != nil {
			log.Println("Error al obtener valor de registro")
			return false
		}

		suma := valorDestino + valorOrigen

		SetearValorRegistro(registroDestino, unProceso, suma)

		log.Printf("TID: %d - Ejecutado SUM - %s = %d", unProceso.Tid, registroDestino, suma)

	case "SUB":

		unProceso.Registros.PC++

		registroDestino := palabras[1]
		registroOrigen := strings.TrimRight(palabras[2], "\n")

		log.Printf("TID: %d - Ejecutando: SUB - %s %s", unProceso.Tid, registroDestino, registroOrigen)

		var valorDestino, valorOrigen uint32
		var err error

		valorDestino, err = ObtenerValorDeRegistro(registroDestino, unProceso)

		if err != nil {
			log.Println("Error al obtener valor de registro")
			return false
		}

		valorOrigen, err = ObtenerValorDeRegistro(registroOrigen, unProceso)

		if err != nil {
			log.Println("Error al obtener valor de registro")
			return false
		}

		resta := valorDestino - valorOrigen

		SetearValorRegistro(registroDestino, unProceso, resta)

		log.Printf("TID: %d - Ejecutado SUB - %s = %d", unProceso.Tid, registroDestino, resta)

	case "JNZ":

		registro := palabras[1]
		instruccion := strings.TrimRight(palabras[2], "\n")

		valorInt, err := strconv.Atoi(instruccion)

		if err != nil {
			fmt.Println("Error al convertir el string a entero:", err)
			return false
		}

		valorInt32 := uint32(valorInt)

		valorRegistro, err := ObtenerValorDeRegistro(registro, unProceso)

		log.Printf("TID: %d - Ejecutando: JNZ - %s %s", unProceso.Tid, registro, instruccion)

		if err != nil {
			log.Println("Error al obtener valor de registro")
			return false
		}

		if valorRegistro != 0 {
			unProceso.Registros.PC = valorInt32
			log.Printf("TID: %d - Ejecutado JNZ, se ha actualizado el PC a %d", unProceso.Tid, valorInt32)
		} else {
			log.Printf("TID: %d - Ejecutado JNZ, el valor del registro es 0, no se ha actualizado el PC", unProceso.Tid)
			unProceso.Registros.PC++
		}

	case "LOG":

		unProceso.Registros.PC++

		registro := palabras[1]
		valorRegistro, err := ObtenerValorDeRegistro(registro, unProceso)

		if err != nil {
			log.Println("Error al obtener valor de registro")
			return false
		}

		log.Printf("TID: %d - Ejecutado LOG, el valor del registro %s es %d", unProceso.Tid, registro, valorRegistro)

	case "READ_MEM":

		unProceso.Registros.PC++

		registroDatos := palabras[1]
		registroDireccion := strings.TrimRight(palabras[2], "\n")

		log.Printf("TID: %d - Ejecutando: READ_MEM - %s %s", unProceso.Tid, registroDatos, registroDireccion)

		// Obtener la dirección lógica del registroDireccion
		direccionLogica, err := ObtenerValorDeRegistro(registroDireccion, unProceso)
		if err != nil {
			log.Println("Error: No se logro obtener el offset del registro")
			return false
		}

		// convierte la dir logica en una dir fisica de la RAM
		direccionFisica := MMU(direccionLogica, unProceso.Registros.Base, unProceso.Registros.Limite)
		if direccionFisica == -1 {
			SegmentationFault(unProceso)
			return false
		}

		// Busca con la dir fisica de la RAM el contenido que hay en memoria en esa
		// posicion fisica
		valorMemoria, err := LeerDeMemoria(direccionFisica)
		if err != nil {
			log.Println("Error al leer memoria")
			return false
		}

		// Almacenar el valor leído en el registro de destino
		err = SetearValorRegistro(registroDatos, unProceso, valorMemoria)
		if err != nil {
			log.Println("Error al setear el valor en el registro")
			return false
		}

		log.Printf("TID: %d - Ejecutado READ_MEM - %s = %d", unProceso.Tid, registroDatos, valorMemoria)

	case "WRITE_MEM":

		unProceso.Registros.PC++

		// Obtener los registros de datos y dirección
		registroDireccion := palabras[1]
		registroDatos := strings.TrimRight(palabras[2], "\n")

		log.Printf("TID: %d - Ejecutando: WRITE_MEM - %s %s", unProceso.Tid, registroDatos, registroDireccion)

		direccionLogica, err := ObtenerValorDeRegistro(registroDireccion, unProceso)
		if err != nil {
			log.Println("Error: No se logro obtener el offset del registro")
			return false
		}

		direccionFisica := MMU(direccionLogica, unProceso.Registros.Base, unProceso.Registros.Limite)
		if direccionFisica == -1 {
			SegmentationFault(unProceso)
			return false
		}

		// Obtener el valor del registro de datos
		valorDatos, err := ObtenerValorDeRegistro(registroDatos, unProceso)
		if err != nil {
			log.Println("Error al obtener valor de registro de datos")
			return false
		}

		// Escribir el valor en la memoria en la dirección física
		res := EscribirMemoria(direccionFisica, valorDatos)

		if !res {
			log.Println("Error al escribir en memoria")
			return false
		}

		log.Printf("TID: %d - Ejecutado WRITE_MEM - Valor %d escrito en la dirección física %d", unProceso.Tid, valorDatos, direccionFisica)

	case "PROCESS_CREATE":
		unProceso.Registros.PC++
		archivo := palabras[1]
		tamanio, err := strconv.Atoi(palabras[2])
		if err != nil {
			log.Println("Error al convertir el tamaño a entero:", err)
			return false
		}
		prioridad, err := strconv.Atoi(palabras[3])
		if err != nil {
			log.Println("Error al convertir la prioridad a entero:", err)
			return false
		}

		paquete := globalvar.Request_PROCESS_CREATE{
			NombreArchivo:  archivo,
			Tamanio:        tamanio,
			PrioridadHilo0: prioridad,
		}
		commons.EnviarPaquete[globalvar.Request_PROCESS_CREATE](paquete, "KERNEL", "PROCESS_CREATE")

		log.Printf("TID: %d - Ejecutado PROCESS_CREATE - El archivo es %s, el tamaño es %d y la prioridad es %d", unProceso.Tid, archivo, tamanio, prioridad)

	case "PROCESS_EXIT":
		unProceso.Registros.PC++

		paquete := globalvar.Request_PROCESS_EXIT{}
		res := commons.EnviarPaqueteYEsperarOK[globalvar.Request_PROCESS_EXIT](paquete, "KERNEL", "PROCESS_EXIT")

		if res {
			log.Printf("TID: %d - Ejecutado PROCESS_EXIT", unProceso.Tid)
		}
	case "THREAD_CREATE":
		unProceso.Registros.PC++
		archivo := palabras[1]
		prioridad, err := strconv.Atoi(palabras[2])
		if err != nil {
			log.Println("Error al convertir la prioridad a entero:", err)
			return false
		}

		paquete := globalvar.Request_THREAD_CREATE{
			NombreArchivo: archivo,
			Prioridad:     prioridad,
		}
		commons.EnviarPaquete[globalvar.Request_THREAD_CREATE](paquete, "KERNEL", "THREAD_CREATE")

		log.Printf("TID: %d - Ejecutado THREAD_CREATE - El archivo es %s y la prioridad es %d", unProceso.Tid, archivo, prioridad)

	case "THREAD_JOIN":
		unProceso.Registros.PC++
		tid, err := strconv.Atoi(palabras[1])
		if err != nil {
			log.Println("Error al convertir el TID a entero:", err)
			return false
		}

		paquete := globalvar.Request_THREAD_JOIN{
			Tid: tid,
		}
		commons.EnviarPaquete[globalvar.Request_THREAD_JOIN](paquete, "KERNEL", "THREAD_JOIN")

		log.Printf("TID: %d - Ejecutado THREAD_JOIN - El TID es %d", unProceso.Tid, tid)

	case "THREAD_CANCEL":
		unProceso.Registros.PC++
		tid, err := strconv.Atoi(palabras[1])
		if err != nil {
			log.Println("Error al convertir el TID a entero:", err)
			return false
		}

		paquete := globalvar.Request_THREAD_CANCEL{
			Tid: tid,
		}
		commons.EnviarPaquete[globalvar.Request_THREAD_CANCEL](paquete, "KERNEL", "THREAD_CANCEL")

		log.Printf("TID: %d - Ejecutado THREAD_CANCEL - El TID es %d", unProceso.Tid, tid)

	case "THREAD_EXIT":
		unProceso.Registros.PC++

		paquete := globalvar.Request_THREAD_EXIT{}
		res := commons.EnviarPaqueteYEsperarOK[globalvar.Request_THREAD_EXIT](paquete, "KERNEL", "THREAD_EXIT")

		if res {
			log.Printf("TID: %d - Ejecutado THREAD_EXIT", unProceso.Tid)
		}

	case "MUTEX_CREATE":
		unProceso.Registros.PC++
		recurso := palabras[1]

		paquete := globalvar.Request_MUTEX_CREATE{
			Recurso: recurso,
		}
		commons.EnviarPaquete[globalvar.Request_MUTEX_CREATE](paquete, "KERNEL", "MUTEX_CREATE")

		log.Printf("TID: %d - Ejecutado MUTEX_CREATE - El recurso es %s", unProceso.Tid, recurso)

	case "MUTEX_LOCK":
		unProceso.Registros.PC++
		recurso := palabras[1]

		paquete := globalvar.Request_MUTEX_LOCK{
			Recurso: recurso,
		}
		commons.EnviarPaquete[globalvar.Request_MUTEX_LOCK](paquete, "KERNEL", "MUTEX_LOCK")

		log.Printf("TID: %d - Ejecutado MUTEX_LOCK - El recurso es %s", unProceso.Tid, recurso)

	case "MUTEX_UNLOCK":
		unProceso.Registros.PC++
		recurso := palabras[1]

		paquete := globalvar.Request_MUTEX_UNLOCK{
			Recurso: recurso,
		}
		commons.EnviarPaquete[globalvar.Request_MUTEX_UNLOCK](paquete, "KERNEL", "MUTEX_UNLOCK")

		log.Printf("TID: %d - Ejecutado MUTEX_UNLOCK - El recurso es %s", unProceso.Tid, recurso)

	case "DUMP_MEMORY":
		unProceso.Registros.PC++

		paquete := globalvar.Request_DUMP_MEMORY{}
		commons.EnviarPaquete[globalvar.Request_DUMP_MEMORY](paquete, "KERNEL", "DUMP_MEMORY")

		log.Printf("TID: %d - Ejecutado DUMP_MEMORY", unProceso.Tid)

	case "IO":
		unProceso.Registros.PC++
		tiempo, err := strconv.Atoi(palabras[1])
		if err != nil {
			log.Println("Error al convertir el tiempo a entero:", err)
			return false
		}
		log.Printf("TID: %d - Ejecutando: IO - El valor del tiempo es %s", unProceso.Tid, palabras[1])

		paquete := globalvar.Request_IO{
			Tiempo: tiempo,
		}
		commons.EnviarPaquete[globalvar.Request_IO](paquete, "KERNEL", "IO")
	}
	//log.Printf("## TID: %d - Actualizo contexto de ejecucion", unProceso.Tid)
	return true
}

func UpdateInterrupt(w http.ResponseWriter, r *http.Request) {
	var paquete globalvar.Paquete_Motivo
	err := commons.DecodificarJSON(w, r, &paquete)
	if err != nil {
		log.Println("Error: No se pudo decodificar la Interrupcion")
		return
	}

	MotivoInterrupcion = paquete.Motivo
	HayInterrupcion = true

	w.WriteHeader(http.StatusOK)
}
func CheckInterrupt(contexto globalvar.ContextoEjecucion) bool {
	if !HayInterrupcion {
		return false
	}
	HayInterrupcion = false
	if MotivoInterrupcion != "FinHilo" && MotivoInterrupcion != "FinProceso" {
		GuardarContexto(contexto)
	}
	DevolverTID()
	return true

}
func DevolverTID() {
	paquete := globalvar.Paquete_Motivo{
		Motivo: MotivoInterrupcion,
	}

	commons.EnviarPaquete[globalvar.Paquete_Motivo](paquete, "KERNEL", "DevolverHilo")
}
func GuardarContexto(contexto globalvar.ContextoEjecucion) {
	log.Printf("## TID: %d - Actualizo contexto de ejecucion", contexto.Tid)
	commons.EnviarPaquete[globalvar.ContextoEjecucion](contexto, "MEMORIA", "ActualizarContexto")
}

func LoadConfig(path string) error {

	// Abre el configFile (path),
	// Si lo abre, configFile contendrá un puntero al archivo abierto y err = nil.
	configFile, err := os.Open(path)

	if err != nil {
		return err
	}

	// Garantizamos que se cierre el configFile, aunque ocurra un error después.
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	Config = &ModuleConfig{}

	// Usa el decoder para leer el configFile
	err = decoder.Decode(Config)

	if err != nil {
		return err
	}

	return nil
}

func SetearValorRegistro(registro string, unProceso *globalvar.ContextoEjecucion, unValor uint32) error {

	//cambia el valor de un registro por el pasado por parametro

	switch registro {
	case "PC":
		unProceso.Registros.PC = unValor
	case "AX":
		unProceso.Registros.AX = unValor
	case "BX":
		unProceso.Registros.BX = unValor
	case "CX":
		unProceso.Registros.CX = unValor
	case "DX":
		unProceso.Registros.DX = unValor
	case "EX":
		unProceso.Registros.EX = unValor
	case "FX":
		unProceso.Registros.FX = unValor
	case "GX":
		unProceso.Registros.GX = unValor
	case "HX":
		unProceso.Registros.HX = unValor
	default:
		return fmt.Errorf("registro no válido: %s", registro)
	}

	return nil
}

func ObtenerValorDeRegistro(registro string, unProceso *globalvar.ContextoEjecucion) (uint32, error) {

	//establece en una variable de tipo UINT32 el valor de un registro y lo retorna

	var valor uint32

	switch registro {
	case "PC":
		valor = unProceso.Registros.PC
	case "AX":
		valor = unProceso.Registros.AX
	case "BX":
		valor = unProceso.Registros.BX
	case "CX":
		valor = unProceso.Registros.CX
	case "DX":
		valor = unProceso.Registros.DX
	case "EX":
		valor = unProceso.Registros.EX
	case "FX":
		valor = unProceso.Registros.FX
	case "GX":
		valor = unProceso.Registros.GX
	case "HX":
		valor = unProceso.Registros.HX
	default:
		return 0, fmt.Errorf("registro no válido: %s", registro)
	}
	return valor, nil
}

// Devuelve el offset del registro. Devuelve -1 si hay error
func ObtenerDireccionLogica(registro string) int {
	var dir int = 4
	switch registro {
	case "AX":
		dir = dir * 0
	case "BX":
		dir = dir * 1
	case "CX":
		dir = dir * 2
	case "DX":
		dir = dir * 3
	case "EX":
		dir = dir * 4
	case "FX":
		dir = dir * 5
	case "GX":
		dir = dir * 6
	case "HX":
		dir = dir * 7
	default:
		dir = -1
	}
	return dir
}

func LeerDeMemoria(direccionFisica int) (uint32, error) {
	paquete := globalvar.LeerMemoriaRequest{
		Direccion: direccionFisica,
	}
	resp, err := commons.EnviarPaqueteYEsperar[globalvar.LeerMemoriaRequest, globalvar.LeerMemoriaResponse](paquete, "MEMORIA", "LeerMemoria")
	if err != nil {
		return 0, err
	}

	return resp.Valor, nil
}

func EscribirMemoria(direccionFisica int, valor uint32) bool {
	paquete := globalvar.EscribirMemoriaRequest{
		Direccion: direccionFisica,
		Valor:     valor,
	}
	return commons.EnviarPaqueteYEsperarOK[globalvar.EscribirMemoriaRequest](paquete, "MEMORIA", "EscribirMemoria")
}

func MMU(desplazamiento uint32, base uint32, limite uint32) int {

	dirFisica := base + desplazamiento

	if dirFisica > limite {
		fmt.Println("Segmentation fault")
		log.Println("Error al traducir dirección lógica a física")
		return -1
	}

	return int(dirFisica)
}

func SegmentationFault(unProceso *globalvar.ContextoEjecucion) {
	GuardarContexto(*unProceso)
	paquete := globalvar.Paquete_Motivo{Motivo: "SegmentationFault"}
	res := commons.EnviarPaqueteYEsperarOK[globalvar.Paquete_Motivo](paquete, "KERNEL", "SegmentationFault")
	if res {
		return
	}
	return
}

/*
// ------------- Logs Minimos y Obligatorios --------------\\

// MOTIVOS: "syscall" -- "crearProc" -- "crearHilo" -- "bloq" -- "finIO" -- "finQuantum" -- "finProc" --  "finHilo"
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
		case "finQuantum":
			str += "- Desalojado por fin de Quantum"
		case "finHilo":
			str += "Finaliza el hilo"
		}
	}
	log.Println(str)
}
*/
