package tools

import (
	"log/slog"
	"os"

	"github.com/ocelot-cloud/deepstack"
	u "github.com/ocelot-cloud/shared/utils"
)

const (
	MaxPayloadSize = 1024 * 1024 // = 1 MiB
	MaxStorageSize = 10 * MaxPayloadSize
	Port           = "8080"
	CookieName     = "auth"
)

type Config struct {
	UseMailMockClient       bool
	UseSampleDataForTesting bool
	OpenWipeEndpoint        bool
}

func NewConfig() *Config {
	config := &Config{}
	if os.Getenv("PROFILE") == "TEST" {
		config.UseMailMockClient = true
		config.UseSampleDataForTesting = true
		config.OpenWipeEndpoint = true
		u.Logger = deepstack.NewDeepStackLogger(slog.LevelDebug)
	} else {
		config.UseMailMockClient = false
		config.UseSampleDataForTesting = false
		config.OpenWipeEndpoint = false
		u.Logger = deepstack.NewDeepStackLogger(slog.LevelInfo, u.NewFileHandler(slog.LevelInfo))
	}

	return config
}
