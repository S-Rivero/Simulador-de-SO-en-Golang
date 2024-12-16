package commons

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/configs"
	"github.com/sisoputnfrba/tp-golang/utils/logs"
)

type Mensaje struct {
	Mensaje string `json:"mensaje"`
}

// w envia una respuesta HTTP al cliente
// r contiene la solicitud enviada por el cliente (body, headers)
func RecibirMensaje(w http.ResponseWriter, r *http.Request) {
	var mensaje Mensaje

	// Decodifica r y la guarda en mensaje
	err := DecodificarJSON(w, r, &mensaje)
	if err != nil {
		return
	}

	log.Printf("Mensaje recibido %+v\n", mensaje.Mensaje)

	w.WriteHeader(http.StatusOK)

	// w.Write devuelve el número de bytes escritos y un error,
	// acá las ignoramos (_, _) porque no las necesitamos.
	_, _ = w.Write([]byte("Mensaje recibido"))
}

// Decodifica r y la guarda en requestStruct. Si hay error devuelve != nil
func DecodificarJSON(w http.ResponseWriter, r *http.Request, requestStruct interface{}) error {
	err := json.NewDecoder(r.Body).Decode(requestStruct)

	if err != nil {
		log.Printf("Error al decodificar JSON: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Error al decodificar JSON"))
	}
	return err
}

func EnviarMensaje(ip string, puerto int, mensajeTxt string) {

	for {
		if Handshake(ip, puerto) {
			mensaje := Mensaje{Mensaje: mensajeTxt}
			body, err := json.Marshal(mensaje)
			if err != nil {
				log.Printf("error codificando mensaje: %s", err.Error())
			}

			url := fmt.Sprintf("http://%s:%d/mensaje", ip, puerto)
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
			if err != nil {
				log.Printf("error enviando mensaje a ip:%s puerto:%d", ip, puerto)
			}

			log.Printf("respuesta del servidor: %s", resp.Status)

			break //funcionó el handshake
		}

		log.Println("Handshake fallido, esperamos 3 segundos")
		time.Sleep(3 * time.Second)
	}

}

func LevantarServidor(port string, mux http.Handler, wg *sync.WaitGroup) {
	wg.Add(1)
	var err = http.ListenAndServe(port, mux)
	if err != nil {
		wg.Done()
		panic(err)
	}
}

// HANDSHAKE: Se envia un PING Antes de enviar cualquier mensaje al modulo
// si este responde ("PONG"), se procede a enviar el mensaje
func Handshake(ip string, puerto int) bool {
	url := fmt.Sprintf("http://%s:%d/handshake", ip, puerto)
	resp, err := http.Get(url) // Intentamos conectarnos al servidor.
	if err != nil {
		log.Printf("No se pudo conectar al servidor para handshake: %s", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		//log.Println("Handshake exitoso con el servidor")
		return true
	}

	log.Printf("El servidor respondió con un estado inesperado: %d", resp.StatusCode)
	return false
}

func HandshakeHandler(w http.ResponseWriter, r *http.Request) {
	// Si el método no es GET, devolvemos un error.
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Respondemos con OK para indicar que estamos listos.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Servidor listo"))
}

func InstanciarPaths[ConfigStruct any]() *ConfigStruct {

	// Recibe Path relativo
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Recorta la ultima seccion del path, obteniendo el modulo que esta instanciando
	slice := strings.Split(path, "/")
	module := slice[len(slice)-1]

	// Configura la dir donde se guardan los logs
	logs.ConfigurarLogger(filepath.Join(path, module+".log"))

	// IniciarConfiguracion: carga config del .json
	globals := configs.IniciarConfiguracion(filepath.Join(path, "config.json"), new(ConfigStruct)).(*ConfigStruct)
	if globals == nil {
		log.Fatalln("Error al cargar la configuración de " + module)
	}
	return globals
}

// ValidateKernelArgs valida los argumentos para el módulo Kernel
func ValidateKernelArgs() (string, int, int, error) {
	if len(os.Args) < 4 {
		return "", 0, 0, fmt.Errorf("uso: %s <archivo_pseudocodigo> <tamanio_proceso> <prioridad>", os.Args[0])
	}

	archivo := os.Args[1]

	var tamanio int
	var prioridad int

	defer func() {
		if r := recover(); r != nil {
			tamanio = 0
			prioridad = 0
		}
	}()

	tamanio = StrToInt(os.Args[2])
	prioridad = StrToInt(os.Args[3])

	if tamanio == 0 {
		return "", 0, 0, fmt.Errorf("el tamaño del proceso debe ser un número entero mayor a 0")
	}

	return archivo, tamanio, prioridad, nil
}

func StrToInt(str string) int {
	x, err := strconv.Atoi(str)
	if err != nil {
		log.Fatal(err)
	}
	return x
}

func EsperarConexion(modulo string) {
	ip, port := ElegirIpPuerto(modulo)
	log.Printf("Conectando con %s", modulo)
	for {
		if HandshakeProlijo(ip, port) {
			break
		}
	}
}
func HandshakeProlijo(ip string, port int) bool {
	url := fmt.Sprintf("http://%s:%d/NewHandshake", ip, port)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
func Handler_HandshakeProlijo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func EnviarPaquete[tipoPaquete any](paquete tipoPaquete, modulo string, url string) {
	ip, port := ElegirIpPuerto(modulo)
	for {
		if HandshakeProlijo(ip, port) {
			body, err := json.Marshal(paquete)
			if err != nil {
				log.Printf("Error al formatear el paquete: %s", err.Error())
			}

			url := fmt.Sprintf("http://%s:%d/%s", ip, port, url)
			http.Post(url, "application/json", bytes.NewBuffer(body))
			return
		}
	}
}

func EnviarPaqueteYEsperar[tipoPaquete any, tipoRespuesta any](paquete tipoPaquete, modulo string, url string) (resp tipoRespuesta, err error) {
	ip, port := ElegirIpPuerto(modulo)
	for {
		if HandshakeProlijo(ip, port) {
			body, err := json.Marshal(paquete)
			if err != nil {
				log.Printf("Error al formatear el paquete: %s", err.Error())
			}

			url := fmt.Sprintf("http://%s:%d/%s", ip, port, url)
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
			if err != nil {
				log.Printf("Error enviando paquete a ip:%s puerto:%d -- Error: %s", ip, port, err.Error())
			}
			//log.Printf("Respuesta sobre el paquete enviado: %s %d", resp.Status, port)

			var respuesta tipoRespuesta
			err2 := json.NewDecoder(resp.Body).Decode(&respuesta)
			if err2 != nil {
				log.Println("Error: No se pudo decodificar la respuesta JSON")
				return respuesta, err2
			}
			return respuesta, err
		}
	}
}

func EnviarPaqueteYEsperarOK[tipoPaquete any](paquete tipoPaquete, modulo string, url string) bool {
	ip, port := ElegirIpPuerto(modulo)
	for {
		if HandshakeProlijo(ip, port) {
			body, err := json.Marshal(paquete)
			if err != nil {
				log.Printf("Error al formatear el paquete: %s", err.Error())
			}
			//log.Printf("Enviando paquete a ip:%s puerto:%d url:%s", ip, port, url)
			url := fmt.Sprintf("http://%s:%d/%s", ip, port, url)

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
			if err != nil {
				log.Printf("Error enviando paquete a ip:%s puerto:%d -- Error: %s", ip, port, err.Error())
			}
			//log.Printf(" Respuesta sobre el paquete enviado: %s %d", resp.Status, port)

			return resp.StatusCode == http.StatusOK
		}
	}
}

func ResponderPost[tipoPaquete any](w http.ResponseWriter, paquete tipoPaquete) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paquete)
}

func ResponderError(w http.ResponseWriter, error string) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Printf("%s", error)
}

// GetConfigPath devuelve la ruta al archivo de configuración correspondiente
func GetConfigPath(basePath string, testName string, variant string) string {
	// Si no se especifica prueba, usar config.json por defecto
	if testName == "" {
		return filepath.Join(basePath, "config.json")
	}

	// Si hay una variante específica, intentar usar esa configuración
	if variant != "" {
		variantPath := filepath.Join(basePath, "configs", fmt.Sprintf("prueba_%s_%s.config.json", testName, variant))
		if _, err := os.Stat(variantPath); err == nil {
			return variantPath
		}
	}

	// Intentar usar la configuración base de la prueba
	configPath := filepath.Join(basePath, "configs", fmt.Sprintf("prueba_%s.config.json", testName))
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Si no se encuentra configuración específica, usar la default
	return filepath.Join(basePath, "config.json")
}

// InstanciarPathsWithTest es una versión modificada de InstanciarPaths que acepta el nombre de la prueba
func InstanciarPathsWithTest[ConfigStruct any](testName string, variant string) *ConfigStruct {
	// Obtener el directorio actual
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Obtener el nombre del módulo
	slice := strings.Split(path, "/")
	module := slice[len(slice)-1]

	// Configurar logger
	logs.ConfigurarLogger(filepath.Join(path, module+".log"))

	// Obtener la ruta de configuración apropiada
	configPath := GetConfigPath(path, testName, variant)

	// Cargar la configuración
	globals := configs.IniciarConfiguracion(configPath, new(ConfigStruct)).(*ConfigStruct)
	if globals == nil {
		log.Fatalln("Error al cargar la configuración de " + module)
	}
	return globals
}

var IpCPU *string
var IpKernel *string
var IpMemoria *string
var IpFS *string

func InstanciarIPs(cpu string, fs string, kernel string, memoria string) {
	IpCPU = &cpu
	IpFS = &fs
	IpKernel = &kernel
	IpMemoria = &memoria
}

func ElegirIpPuerto(modulo string) (string, int) {
	switch modulo {
	case "CPU":
		return *IpCPU, 8001
	case "FS":
		return *IpFS, 8002
	case "KERNEL":
		return *IpKernel, 8003
	case "MEMORIA":
		return *IpMemoria, 8004
	}
	return "", 0
}
