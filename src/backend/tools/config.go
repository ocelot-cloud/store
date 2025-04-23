package tools

import (
	"github.com/ocelot-cloud/shared/utils"
	"os"
)

var (
	Logger     = utils.ProvideLogger("DEBUG")
	Port       = "8082"
	RootUrl    = "http://localhost:" + Port
	CookieName = "auth"
	Profile    = getProfile()

	apiPrefix    = "/api"
	WipeDataPath = apiPrefix + "/wipe-data"

	userPath            = apiPrefix + "/account"
	RegistrationPath    = userPath + "/registration"
	EmailValidationPath = userPath + "/validate"
	LoginPath           = userPath + "/login"
	LogoutPath          = userPath + "/logout"
	AuthCheckPath       = userPath + "/auth-check"
	DeleteUserPath      = userPath + "/delete"
	ChangePasswordPath  = userPath + "/change-password"

	versionPath       = apiPrefix + "/versions"
	VersionUploadPath = versionPath + "/upload"
	VersionDeletePath = versionPath + "/delete"
	GetVersionsPath   = versionPath + "/list"
	DownloadPath      = versionPath + "/download"

	appPath         = apiPrefix + "/apps"
	AppCreationPath = appPath + "/create"
	AppGetListPath  = appPath + "/get-list"
	AppDeletePath   = appPath + "/delete"
	SearchAppsPath  = appPath + "/search"

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
