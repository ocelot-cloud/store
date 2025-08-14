package tools

import (
	"os"

	"github.com/ocelot-cloud/deepstack"
)

var (
	Logger     = deepstack.NewDeepStackLogger(os.Getenv("LOG_LEVEL"))
	Port       = "8082"
	RootUrl    = "http://localhost:" + Port
	CookieName = "auth"
	Profile    = getProfile()

	UseMailMockClient = false
)

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
