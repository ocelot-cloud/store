package tools

var (
	SampleUser     = "samplemaintainer"
	SampleApp      = "gitea"
	SampleVersion  = "0.0.1"
	SampleEmail    = "testuser@example.com"
	SamplePassword = "mypassword"
	SampleForm     = &RegistrationForm{
		SampleUser,
		SamplePassword,
		SampleEmail,
	}
)

func GetValidVersionBytes() []byte {
	versionBytes, err := ZipDirectory(SamplesDir + "/test-compose-files/allow-app-yml")
	if err != nil {
		Logger.Fatal("Failed to read sample version file: %v", err)
	}
	return versionBytes
}
