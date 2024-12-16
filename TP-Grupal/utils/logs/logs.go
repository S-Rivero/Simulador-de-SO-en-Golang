package logs

import (
	"io"
	"log"
	"os"
)

func ConfigurarLogger(path string) {
	//os.O_CREATE crea el file si no existe
	//os.O_APPEND agrega el log al final del file
	//os.O_RDWR permite leer y escribir el file
	//0666 establece los permisos

	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}

	//MultiWriter permite enviar logs a varios destinos al mismo tiempo
	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)
}
