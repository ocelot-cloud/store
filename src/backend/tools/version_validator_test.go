package tools

import (
	"fmt"
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
	tr "github.com/ocelot-cloud/task-runner"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var appName = "gitea"
var maintainerName = "samplemaintainer"
var expectedPrefix = "samplemaintainer_gitea_"

func TestMissingDockerComposeYaml(t *testing.T) {
	zipBytes, err := ZipDirectory(sampleComposeDir + "/missing-compose-yaml")
	assert.Nil(t, err)
	err = ValidateVersion(zipBytes, maintainerName, appName)
	assert.NotNil(t, err)
	assert.Equal(t, "docker-compose.yml file is missing in zip", err.Error())
}

func TestAllowAppYaml(t *testing.T) {
	zipBytes, err := ZipDirectory(sampleComposeDir + "/allow-app-yml")
	assert.Nil(t, err)
	err = ValidateVersion(zipBytes, maintainerName, appName)
	assert.Nil(t, err)
}

func TestInvalidAppYaml(t *testing.T) {
	zipBytes, err := ZipDirectory(sampleComposeDir + "/invalid-app-yml")
	assert.Nil(t, err)
	err = ValidateVersion(zipBytes, maintainerName, appName)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid port in app.yml: 123456", err.Error())
}

func TestExtraFile(t *testing.T) {
	zipBytes, err := ZipDirectory(sampleComposeDir + "/extra-file")
	assert.Nil(t, err)
	err = ValidateVersion(zipBytes, maintainerName, appName)
	assert.NotNil(t, err)
	assert.Equal(t, "unexpected file in zip: hello.md", err.Error())
}

func TestExtraDir(t *testing.T) {
	zipBytes, err := ZipDirectory(sampleComposeDir + "/extra-dir")
	assert.Nil(t, err)
	err = ValidateVersion(zipBytes, maintainerName, appName)
	assert.NotNil(t, err)
	assert.Equal(t, "directories are not allowed in the zip file: hello", err.Error())
}

func TestInvalidZipBytes(t *testing.T) {
	zipBytes := []byte("hello")
	err := ValidateVersion(zipBytes, maintainerName, appName)
	assert.NotNil(t, err)
	assert.Equal(t, "failed to read zip file: zip: not a valid zip file", err.Error())
}

func TestOcelotCloudAppIsDenied(t *testing.T) {
	zipBytes := []byte("hello")
	err := ValidateVersion(zipBytes, maintainerName, "ocelotcloud")
	assert.NotNil(t, err)
	assert.Equal(t, ocelotCloudAppAlreadyReserved, err.Error())
}

func TestValidation(t *testing.T) {
	testCases := []struct {
		file          string
		expectedError string
	}{
		{"sample-gitea.yml", ""},
		{"not-existing-keyword.yml", fmt.Sprintf(notAllowedTopLevelKeyword, "version2")},
		{"not-allowed-root-keyword.yml", fmt.Sprintf(notAllowedTopLevelKeyword, "configs")},
		{"using-external-network.yml", fmt.Sprintf(notAllowedTopLevelKeyword, "networks")},

		{"not-allowed-service-keyword.yml", fmt.Sprintf(notAllowedKeyInService, "gitea", "privileged")},
		{"host-network.yml", fmt.Sprintf(notAllowedKeyInService, "gitea", "network_mode")},
		{"using-healthcheck.yml", fmt.Sprintf(notAllowedKeyInService, "gitea", "healthcheck")},
		{"using-cap-drop.yml", fmt.Sprintf(notAllowedKeyInService, "gitea", "cap_drop")},
		{"using-restart.yml", fmt.Sprintf(notAllowedKeyInService, "gitea", "restart")},

		{"mounted-volume-1.yml", fmt.Sprintf(notAllowedMountingHostDirectories, "gitea")},
		{"mounted-volume-2.yml", fmt.Sprintf(notAllowedMountingHostDirectories, "gitea")},

		{"empty.yml", emptyDockerComposeIsNotAllowed},

		{"missing-app-service.yml", fmt.Sprintf(mainServiceMustBeDefined, "gitea")},

		{"empty-image-tag.yml", fmt.Sprintf(notAllowedMissingDockerImageTag, "gitea/gitea")},

		{"latest-image-tag.yml", notAllowedLatestDockerImageTag},

		{"exposing-port-53.yml", fmt.Sprintf(notAllowedExposingDefaultHttpPorts, "53")},
		{"exposing-port-80.yml", fmt.Sprintf(notAllowedExposingDefaultHttpPorts, "80")},
		{"exposing-port-443.yml", fmt.Sprintf(notAllowedExposingDefaultHttpPorts, "443")},

		{"docker-compose-consistency-check.yml", "docker-compose.yml consistency check failed: invalid compose project"},

		{"wrong-volume-prefix.yml", fmt.Sprintf(wrongVolumeNamePrefix, expectedPrefix)},

		{"not-resources-keyword-in-deploy.yml", deployKeywordMustOnlyContainResources},
		{"devices-keyword-in-resources.yml", devicesKeywordIsForbidden},

		{"volume-has-name-field.yml", globalVolumeShouldNotHaveSubKeywords},

		{"database-has-wrong-app-name-in-volume.yml", fmt.Sprintf(wrongVolumeNamePrefix, expectedPrefix)},

		{"non-main-container-name-requires-prefix.yml", fmt.Sprintf(wrongContainerNamePrefix, "samplemaintainer_gitea_")},

		{"side-service-missing-container-name.yml", containerNameMissing},
		{"main-service-missing-container-name.yml", containerNameMissing},
		{"main-service-wrong-container-name.yml", fmt.Sprintf(mainServiceNeedsCorrectContainerNameValue, "gitea", "samplemaintainer_gitea_gitea")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.file, func(t *testing.T) {
			zipBytes, err := createZipWithComposeFile(sampleComposeDir, tc.file)
			if err != nil {
				t.Fatalf("Failed to create ZIP archive: %v", err)
			}

			err = ValidateVersion(zipBytes, maintainerName, appName)
			if tc.expectedError == "" {
				if err != nil {
					t.Errorf("Validation failed for %s: %v", tc.file, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but got none for %s", tc.file)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s' but got '%v' for %s", tc.expectedError, err.Error(), tc.file)
				}
			}
		})
	}
}

func createZipWithComposeFile(dir, composeFileName string) ([]byte, error) {
	tempDir, err := os.MkdirTemp("", "testCompose")
	if err != nil {
		return nil, err
	}
	defer utils.RemoveDir(tempDir)

	composePath := filepath.Join(dir, composeFileName)
	if err := copyFile(composePath, filepath.Join(tempDir, "docker-compose.yml")); err != nil {
		return nil, err
	}

	return ZipDirectory(tempDir)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func TestCompleteDockerComposeYaml(t *testing.T) {
	defer tr.Remove("input.yml")
	tr.Copy(SamplesDir+"/yaml-keyword-completion", "input.yml", ".")
	err := CompleteDockerComposeYaml("samplemaintainer", "gitea", "input.yml")
	assert.Nil(t, err)
	expectedBytes, err := os.ReadFile(SamplesDir + "/yaml-keyword-completion/expected-output.yml")
	assert.Nil(t, err)
	actualBytes, err := os.ReadFile("input.yml")
	assert.Nil(t, err)
	AssertYamlEquality(t, expectedBytes, actualBytes)
}

func TestCheckAppYamlCorrectness(t *testing.T) {
	testCases := []struct {
		file          string
		expectedError string
	}{
		{"empty-is-valid", ""},
		{"full-valid", ""},
		{"port-out-of-range", "invalid port in app.yml: 123456"},
		{"not-allowed-field", "not allowed key in app.yml: not_allowed_field"},
		{"not-a-path", "invalid url_path in app.yml: <script>alert('XSS')</script>"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.file, func(t *testing.T) {
			err := checkAppYamlCorrectness(SamplesDir + "/app-yamls/" + tc.file + ".yml")
			if tc.expectedError == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}

func TestIsValidURLPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/", true},
		{"/valid-path", true},
		{"/valid_123", true},
		{"/Another-Valid", true},

		{"/this-is-a-really-long-path-that-should-not-exceed-one-hundred-characters-0123456789-0123456789-0123456789", false},
		{"does-not-start-with-slash/path", false},
		{"/invalid-symbols@path", false},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, IsValidURLPath(test.path))
	}
}
