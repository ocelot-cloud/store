package tools

import (
	"log/slog"
	"os"

	"github.com/ocelot-cloud/deepstack"
)

var (
	// TODO !! initializing config object should set this like the "cloud" project does
	// TODO !! use utils.Logger instead
	Logger     = deepstack.NewDeepStackLogger(slog.LevelDebug)
	Port       = "8082"
	RootUrl    = "http://localhost:" + Port
	CookieName = "auth"

	// TODO !! remove profile if not needed,
	Profile           = getProfile()
	UseMailMockClient = false
)

// TODO !! replace .env file so that instead the email config is stored in the database and can be changed via REST API
// TODO !! get rid of profile, just pass a config object
type PROFILE int

const (
	PROD PROFILE = iota
	TEST
)

func (p PROFILE) String() string {
	switch p {
	case PROD:
		return "prod"
	case TEST:
		return "test"
	default:
		return "unknown"
	}
}

func getProfile() PROFILE {
	envProfile := os.Getenv("PROFILE")
	var profile PROFILE
	if envProfile == "TEST" {
		UseMailMockClient = true
		profile = TEST
	} else {
		UseMailMockClient = false
		profile = PROD
	}
	Logger.Info("profile set", ProfileField, profile.String())
	return profile
}

const MaxPayloadSize = 1024 * 1024 // = 1 MiB
const MaxStorageSize = 10 * MaxPayloadSize
