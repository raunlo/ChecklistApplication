package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"com.raunlo.checklist/internal/deployment"
	"github.com/spf13/viper"
)

func main() {
	applicationConfig := getApplicationConfig("./application.yaml")
	application := deployment.Init(applicationConfig)
	if err := application.StartApplication(); err != nil {
		panic(err)
	}
	log.Println("Application started")
}

func getApplicationConfig(configPath string) deployment.ApplicationConfiguration {
	var applicationConfiguration deployment.ApplicationConfiguration
	filePath, err := filepath.Abs(configPath)
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	expanded := os.Expand(string(data), func(variable string) string {
		parts := strings.SplitN(variable, ":", 2)
		if val, ok := os.LookupEnv(parts[0]); ok {
			return val
		}
		if len(parts) == 2 {
			return parts[1]
		}
		return ""
	})

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(expanded)); err != nil {
		panic(err)
	}
	if err := v.UnmarshalKey("applicationConfiguration", &applicationConfiguration); err != nil {
		panic(err)
	}
	return applicationConfiguration
}
