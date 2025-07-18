package tools

import (
	"github.com/ocelot-cloud/shared/utils"
	"os"
)

var (
	Logger     = utils.ProvideLogger("DEBUG", true)
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

func getProfile() PROFILE {
	envProfile := os.Getenv("PROFILE")
	if envProfile == "TEST" {
		UseMailMockClient = true
		return TEST
	} else {
		return PROD
	}
}

const MaxPayloadSize = 1024 * 1024 // = 1 MiB
const MaxStorageSize = 10 * MaxPayloadSize
