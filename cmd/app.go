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
	validateConfiguration(applicationConfig)
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

// validateConfiguration ensures critical security settings are properly configured
func validateConfiguration(config deployment.ApplicationConfiguration) {
	// CORS validation - Wildcard origin with credentials is a critical security vulnerability
	if config.CorsConfiguration.Hostname == "" || config.CorsConfiguration.Hostname == "*" {
		log.Fatal("SECURITY ERROR: CORS_CONFIGURATION_HOST_NAME must be set to specific origin(s).\n" +
			"Wildcard '*' is not allowed for security reasons.\n" +
			"Examples:\n" +
			"  Production: CORS_CONFIGURATION_HOST_NAME=https://app.dailychexly.com\n" +
			"  Development: CORS_CONFIGURATION_HOST_NAME=http://localhost:3000\n" +
			"  Multiple: CORS_CONFIGURATION_HOST_NAME=https://app1.com,https://app2.com")
	}

	log.Printf("CORS configuration validated: %s", config.CorsConfiguration.Hostname)
}
