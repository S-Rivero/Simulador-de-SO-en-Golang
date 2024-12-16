package configs

import (
	"encoding/json"
	"log"
	"os"
)

func IniciarConfiguracion(filePath string, moduleConfig interface{}) interface{} {
	configFile, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&moduleConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	return moduleConfig
}
