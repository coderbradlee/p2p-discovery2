package utils

import (
	"fmt"
	"github.com/BurntSushi/toml"
	// "log"
	"os"
	"path/filepath"
)
type log struct{
	Dir string `json:"dir"`
	Name string `json:"name"`
	Console bool `json:"console"`
	Num int32 `json:"num"`
	Size int64 `json:"size"`
	Level string `json:"level"`
}
type Config struct {
	Coin        string         `json:"coin"`
	BlockNumber int64 `json:"blockNumber"`
	Log log          `json:"log"`
}

func LoadConfig(configFileName string, cfg interface{}) bool {

	var err error

	//configFileName := "api.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, err = filepath.Abs(configFileName)
	if err != nil {
		fmt.Printf("LoadConfig", "filepath.Abs err", err)
		return false
	}
	configFile, err := os.Open(configFileName)
	if err != nil {
		fmt.Printf("LoadConfig", "os.Open File error: ", err.Error())
		return false
	}
	defer configFile.Close()

	if _, err = toml.DecodeReader(configFile, cfg); err != nil {
		fmt.Printf("LoadConfig", "toml.DecodeReader error: ", err.Error())
		return false
	}

	return true
}
