package tools

import "os"

const (
	MaxPayloadSize = 1024 * 1024 // = 1 MiB
	MaxStorageSize = 10 * MaxPayloadSize

	// TODO !! initializing config object should set this like the "cloud" project does
	// TODO !! use u.u.Logger instead
	Port       = "8082"
	RootUrl    = "http://localhost:" + Port
	CookieName = "auth"
)

type Config struct {
	UseMailMockClient    bool
	UseSpecialExpiration bool
	CreateSampleData     bool
	OpenWipeEndpoint     bool
}

func NewConfig() *Config {
	config := &Config{}
	if os.Getenv("PROFILE") == "TEST" {
		config.UseMailMockClient = true
		config.UseSpecialExpiration = true
		config.CreateSampleData = true
		config.OpenWipeEndpoint = true
	} else {
		config.UseMailMockClient = false
		config.UseSpecialExpiration = false
		config.CreateSampleData = false
		config.OpenWipeEndpoint = false
	}

	return config
}
