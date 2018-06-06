package util

import (
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func LoadConfig(configFileName string, cfg interface{}) bool {
	defer func() {
		if rev := recover(); rev != nil {
			log.Println("LoadConfig", "rev", rev)
		}
	}()

	var err error

	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, err = filepath.Abs(configFileName)
	if err != nil {
		log.Println("LoadConfig", "filepath.Abs err", err)
		return false
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Println("LoadConfig", "os.Open File error: ", err.Error())
		return false
	}
	defer configFile.Close()

	if _, err = toml.DecodeReader(configFile, cfg); err != nil {
		log.Println("LoadConfig", "toml.DecodeReader error: ", err.Error())
		return false
	}

	return true
}
