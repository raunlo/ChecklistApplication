package deployment

import "github.com/raunlo/pgx-with-automapper/pool"

type ApplicationConfiguration struct {
	ServerConfiguration        `yaml:"serverConfiguration"`
	pool.DatabaseConfiguration `yaml:"databaseConfiguration"`
	CorsConfiguration          `yaml:"corsConfiguration"`
	GoogleSSOConfiguration     `yaml:"googleSSOConfiguration"`
	SessionAuthConfiguration   `yaml:"sessionAuthConfiguration"`
}

type (
	EndpointOverrideConfiguration struct {
		EndpointURL *string `yaml:"endpointUrl"`
		Region      *string `yaml:"region"`
	}
	ServerConfiguration struct {
		Port        string `yaml:"port"`
		BaseUrl     string `yaml:"baseUrl"`
		FrontendUrl string `yaml:"frontendUrl"`
	}
	CorsConfiguration struct {
		Hostname string `yaml:"hostname"`
	}
	GoogleSSOConfiguration struct {
		ClientID     string `yaml:"clientID"`
		ClientSecret string `yaml:"clientSecret"`
	}
	SessionAuthConfiguration struct {
		EncryptionKey string `yaml:"encryptionKey"`
	}
)
