package main

import (
	"os"
	"path/filepath"

	"com.raunlo.checklist/internal/deployment"
	"go.uber.org/config"
)

func main() {
	applicationConfig := getApplicationConfig("./application.yaml")
	application := deployment.Init(applicationConfig)
	if err := application.StartApplication(); err != nil {
		panic(err)
	}
}

func getApplicationConfig(configPath string) deployment.ApplicationConfiguration {
	var applicationConfiguration deployment.ApplicationConfiguration
	if filePath, err := filepath.Abs(configPath); err != nil {
		panic(err)
	} else if file, err := os.Open(filePath); err != nil {
		panic(err)
	} else if provider, err := config.NewYAML(config.Expand(os.LookupEnv), config.Source(file)); err != nil {
		panic(err)
	} else if err := provider.Get("applicationConfiguration").Populate(&applicationConfiguration); err != nil {
		panic(err)
	} else {
		return applicationConfiguration
	}
}
