package deployment

import "github.com/raunlo/pgx-with-automapper/pool"

type ApplicationConfiguration struct {
	ServerConfiguration        `yaml:"serverConfiguration"`
	pool.DatabaseConfiguration `yaml:"databaseConfiguration"`
	CorsConfiguration          `yaml:"corsConfiguration"`
}

type (
	EndpointOverrideConfiguration struct {
		EndpointURL *string `yaml:"endpointUrl"`
		Region      *string `yaml:"region"`
	}
	ServerConfiguration struct {
		Port string `yaml:"port"`
	}
	CorsConfiguration struct {
		Hostname string `yaml:"hostname"`
	}
)
