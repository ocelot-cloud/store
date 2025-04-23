package tools

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var (
	SamplesDir       = utils.FindDir("samples")
	sampleComposeDir = SamplesDir + "/test-compose-files"

	notAllowedTopLevelKeyword                 = "not allowed root keyword in docker-compose.yml: %s"
	notAllowedKeyInService                    = "not allowed key in service '%s': %s"
	notAllowedMountingHostDirectories         = "host directories are mounted in service '%s' which is forbidden"
	emptyDockerComposeIsNotAllowed            = "empty docker-compose.yml is not allowed"
	mainServiceMustBeDefined                  = "there must be a service with the name: %s"
	mainServiceNeedsContainerNameKeyword      = "service '%s' must have 'container_name' keyword"
	mainServiceNeedsCorrectContainerNameValue = "service '%s' must have the container_name '%s'"
	notAllowedMissingDockerImageTag           = "the image tag must consist of an image name and a tag separated by a colon, like 'gitea/gitea:10.5', but got: %s"
	notAllowedLatestDockerImageTag            = "the 'latest' tag is forbidden, to get reproducible apps, only fixed tags with specific software version should be used"
	notAllowedExposingDefaultHttpPorts        = "exposing port %s is forbidden, as it is reserved for Ocelot-Cloud"
	wrongVolumeNamePrefix                     = "volume names must start with '%s'"
	ocelotCloudAppAlreadyReserved             = "app name 'ocelotcloud' is not allowed"
	devicesKeywordIsForbidden                 = "'devices' keyword is not allowed"
	deployKeywordMustOnlyContainResources     = "'deploy' keyword must only contain 'resources' keyword"
	globalVolumeShouldNotHaveSubKeywords      = "volume has sub-keywords, which are not allowed"
	wrongContainerNamePrefix                  = "the container names must have the prefix: %s"
	containerNameMissing                      = "every service needs to have a 'container_name' keyword"
)

func ValidateVersion(zipBytes []byte, maintainerName, appName string) error {
	if appName == "ocelotcloud" {
		return errors.New(ocelotCloudAppAlreadyReserved)
	}

	tempDir, err := utils.UnzipToTempDir(zipBytes)
	if err != nil {
		return err
	}
	defer utils.RemoveDir(tempDir)

	if err := validateFilesInDir(tempDir); err != nil {
		return err
	}
	composePath := filepath.Join(tempDir, "docker-compose.yml")
	if err := parseAndValidateComposeFile(composePath, maintainerName, appName); err != nil {
		return err
	}
	if err := checkDockerComposeSyntax(composePath); err != nil {
		return err
	}
	return nil
}

func validateFilesInDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read temp dir: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("zip file is empty")
	}

	hasDockerCompose := false

	for _, file := range files {
		if file.IsDir() {
			return fmt.Errorf("directories are not allowed in the zip file: %s", file.Name())
		}
		fname := file.Name()
		if fname == "docker-compose.yml" {
			hasDockerCompose = true
		} else if fname == "app.yml" {
			err = checkAppYamlCorrectness(dir + "/" + fname)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unexpected file in zip: %s", fname)
		}
	}

	if !hasDockerCompose {
		return fmt.Errorf("docker-compose.yml file is missing in zip")
	}

	return nil
}

func checkAppYamlCorrectness(filePath string) error {
	data, err := os.ReadFile(filePath) // #nosec G304 (CWE-22): Potential file inclusion via variable; no problem since it is called internally
	if err != nil {
		return fmt.Errorf("failed to read app.yml: %v", err)
	}

	var appConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		return fmt.Errorf("failed to parse app.yml: %v", err)
	}

	allowedKeys := map[string]bool{
		"url_path": true,
		"port":     true,
	}

	for key := range appConfig {
		if !allowedKeys[key] {
			return fmt.Errorf("not allowed key in app.yml: %s", key)
		}
	}

	if urlPath, ok := appConfig["url_path"].(string); ok {
		if !IsValidURLPath(urlPath) {
			return fmt.Errorf("invalid url_path in app.yml: %s", urlPath)
		}
	}

	if port, ok := appConfig["port"].(int); ok {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid port in app.yml: %d", port)
		}
	}

	return nil
}

var re = regexp.MustCompile(`^/[a-zA-Z0-9_-]{0,100}$`)

func IsValidURLPath(path string) bool {
	return re.MatchString(path)
}

func checkDockerComposeSyntax(composePath string) error {
	cmd := exec.Command("docker", "compose", "-f", composePath, "config")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errorMsg := stderr.String()
		if idx := strings.Index(errorMsg, ": "); idx != -1 {
			errorMsg = errorMsg[idx+2:]
		}
		return fmt.Errorf("docker-compose.yml consistency check failed: %s", errorMsg)
	}
	return nil
}
func parseAndValidateComposeFile(composePath, maintainerName, appName string) error {
	compose, err := readComposeFile(composePath)
	if err != nil {
		return err
	}
	if err := validateTopLevelKeys(compose); err != nil {
		return err
	}
	if err := validateServices(compose, maintainerName, appName); err != nil {
		return err
	}
	if err := validateGlobalVolumes(compose); err != nil {
		return err
	}
	return nil
}

func readComposeFile(composePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(composePath) // #nosec G304 (CWE-22): Potential file inclusion via variable; no problem since it is called internally
	if err != nil {
		return nil, fmt.Errorf("failed to read docker-compose.yml: %v", err)
	}
	var c map[string]interface{}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse docker-compose.yml: %v", err)
	}
	if len(c) == 0 {
		return nil, errors.New(emptyDockerComposeIsNotAllowed)
	}
	return c, nil
}

func validateTopLevelKeys(compose map[string]interface{}) error {
	allowed := map[string]bool{"volumes": true, "services": true}
	for k := range compose {
		if !allowed[k] {
			return fmt.Errorf(notAllowedTopLevelKeyword, k)
		}
	}
	return nil
}

func validateServices(compose map[string]interface{}, maintainerName, appName string) error {
	services, ok := compose["services"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid 'services' section in docker-compose.yml")
	}
	isMainServicePresent := false
	for serviceName, serviceValue := range services {
		serviceMap, ok := serviceValue.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid service definition for service %v", serviceName)
		}
		if err := validateServiceKeys(serviceName, serviceMap); err != nil {
			return err
		}
		if err := validateImage(serviceName, serviceMap); err != nil {
			return err
		}
		if err := validateContainerName(serviceMap, maintainerName, appName); err != nil {
			return err
		}
		if serviceName == appName {
			isMainServicePresent = true
			if err := validateMainServiceContainerName(serviceName, serviceMap, maintainerName, appName); err != nil {
				return err
			}
		}
		if err := validatePorts(serviceName, serviceMap); err != nil {
			return err
		}
		if err := validateServiceVolumes(maintainerName, appName, serviceName, serviceMap); err != nil {
			return err
		}
		if err := validateServiceNetworks(serviceName, serviceMap); err != nil {
			return err
		}
		if err := validateDeploySection(serviceName, serviceMap); err != nil {
			return err
		}
	}
	if !isMainServicePresent {
		return fmt.Errorf(mainServiceMustBeDefined, appName)
	}
	return nil
}

func validateServiceKeys(serviceName string, serviceMap map[string]interface{}) error {
	allowed := []string{"image", "container_name", "ports", "volumes", "depends_on", "environment", "deploy", "tmpfs", "tty", "user", "command", "entrypoint"}
	for k := range serviceMap {
		found := false
		for _, a := range allowed {
			if k == a {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf(notAllowedKeyInService, serviceName, k)
		}
	}
	return nil
}

func validateImage(serviceName string, serviceMap map[string]interface{}) error {
	img, ok := serviceMap["image"]
	if !ok {
		return fmt.Errorf("service '%s' must have 'image' keyword", serviceName)
	}
	parts := strings.Split(img.(string), ":")
	if len(parts) < 2 {
		return fmt.Errorf(notAllowedMissingDockerImageTag, img)
	}
	if parts[1] == "latest" {
		return errors.New(notAllowedLatestDockerImageTag)
	}
	return nil
}

func validateContainerName(serviceMap map[string]interface{}, maintainerName, appName string) error {
	cn, ok := serviceMap["container_name"]
	if !ok {
		return errors.New(containerNameMissing)
	}
	prefix := fmt.Sprintf("%s_%s_", maintainerName, appName)
	if !strings.HasPrefix(cn.(string), prefix) {
		return fmt.Errorf(wrongContainerNamePrefix, prefix)
	}
	return nil
}

func validateMainServiceContainerName(serviceName string, serviceMap map[string]interface{}, maintainerName, appName string) error {
	acn, ok := serviceMap["container_name"]
	if !ok {
		return fmt.Errorf(mainServiceNeedsContainerNameKeyword, serviceName)
	}
	expected := fmt.Sprintf("%s_%s_%s", maintainerName, appName, appName)
	if acn != expected {
		return fmt.Errorf(mainServiceNeedsCorrectContainerNameValue, serviceName, expected)
	}
	return nil
}

func validatePorts(serviceName string, serviceMap map[string]interface{}) error {
	p, ok := serviceMap["ports"]
	if !ok {
		return nil
	}
	for _, port := range p.([]interface{}) {
		ps, ok := port.(string)
		if !ok {
			return fmt.Errorf("invalid port definition in service %v", serviceName)
		}
		fields := strings.Split(ps, ":")
		if fields[0] == "53" || fields[0] == "80" || fields[0] == "443" {
			return fmt.Errorf(notAllowedExposingDefaultHttpPorts, fields[0])
		}
	}
	return nil
}

func validateDeploySection(serviceName string, serviceMap map[string]interface{}) error {
	d, has := serviceMap["deploy"]
	if !has {
		return nil
	}
	dm, ok := d.(map[string]interface{})
	if !ok {
		return fmt.Errorf("'deploy' keyword in service '%s' must be a map", serviceName)
	}
	for dk := range dm {
		if dk != "resources" {
			return fmt.Errorf(deployKeywordMustOnlyContainResources, serviceName, dk)
		}
	}
	rm, rok := dm["resources"].(map[string]interface{})
	if !rok {
		return nil
	}
	res, hasRes := rm["reservations"]
	if !hasRes {
		return nil
	}
	rsm, rok := res.(map[string]interface{})
	if !rok {
		return fmt.Errorf("'reservations' in 'resources' of service '%s' must be a map", serviceName)
	}
	if _, hasDev := rsm["devices"]; hasDev {
		return errors.New(devicesKeywordIsForbidden)
	}
	return nil
}

func validateGlobalVolumes(compose map[string]interface{}) error {
	vm, ok := compose["volumes"].(map[string]interface{})
	if !ok {
		return nil
	}
	for _, v := range vm {
		vmap, ok := v.(map[string]interface{})
		if ok && len(vmap) > 0 {
			return errors.New(globalVolumeShouldNotHaveSubKeywords)
		}
	}
	return nil
}

func validateServiceVolumes(maintainerName, appName, serviceName string, serviceMap map[string]interface{}) error {
	vList, err := extractVolumesList(serviceName, serviceMap)
	if err != nil {
		return err
	}
	for _, vol := range vList {
		if err := validateVolumeEntry(maintainerName, appName, serviceName, vol); err != nil {
			return err
		}
	}
	return nil
}

func extractVolumesList(serviceName string, serviceMap map[string]interface{}) ([]string, error) {
	if serviceMap["volumes"] == nil {
		return nil, nil
	}
	v, ok := serviceMap["volumes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid 'volumes' in service %v", serviceName)
	}
	var result []string
	for _, item := range v {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("invalid volume entry in service %v", serviceName)
		}
		result = append(result, s)
	}
	return result, nil
}

func validateVolumeEntry(maintainerName, appName, serviceName, vol string) error {
	parts := strings.Split(vol, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid volume format in service %v: %s", serviceName, vol)
	}
	if filepath.IsAbs(parts[0]) || strings.HasPrefix(parts[0], ".") {
		return fmt.Errorf(notAllowedMountingHostDirectories, serviceName)
	}
	prefix := fmt.Sprintf("%s_%s_", maintainerName, appName)
	if !strings.HasPrefix(vol, prefix) {
		return fmt.Errorf(wrongVolumeNamePrefix, prefix)
	}
	return nil
}

func validateServiceNetworks(serviceName string, serviceMap map[string]interface{}) error {
	if serviceMap["networks"] == nil {
		return nil
	}
	return validateNetworksDefinition(serviceName, serviceMap["networks"])
}

func validateNetworksDefinition(serviceName string, networks interface{}) error {
	switch n := networks.(type) {
	case []interface{}:
		for _, net := range n {
			netStr, ok := net.(string)
			if !ok {
				return fmt.Errorf("invalid network entry in service %v", serviceName)
			}
			if netStr == "host" {
				return fmt.Errorf("using host network is forbidden in service %v", serviceName)
			}
		}
	case map[string]interface{}:
		for netName := range n {
			if netName == "host" {
				return fmt.Errorf("using host network is forbidden in service %v", serviceName)
			}
		}
	default:
		return fmt.Errorf("invalid 'networks' definition in service %v", serviceName)
	}
	return nil
}

func CompleteDockerComposeYaml(maintainer, appName, filePath string) error {
	m, err := readCompose(filePath)
	if err != nil {
		return err
	}
	addExternalNetwork(m, maintainer, appName)
	updateServices(m, maintainer, appName)
	updateVolumes(m)
	return writeCompose(filePath, m)
}

func readCompose(filePath string) (map[string]interface{}, error) {
	b, err := os.ReadFile(filePath) // #nosec G304 (CWE-22): Potential file inclusion via variable; no problem since it is called internally
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func addExternalNetwork(m map[string]interface{}, maintainer, appName string) {
	net := fmt.Sprintf("%s_%s", maintainer, appName)
	m["networks"] = map[string]interface{}{
		net: map[string]interface{}{"external": true},
	}
}

func updateServices(m map[string]interface{}, maintainer, appName string) {
	svcs, _ := m["services"].(map[string]interface{})
	net := fmt.Sprintf("%s_%s", maintainer, appName)
	for k, v := range svcs {
		vm, _ := v.(map[string]interface{})
		vm["networks"] = []interface{}{net}
		vm["restart"] = "unless-stopped"
		vm["cap_drop"] = []interface{}{"ALL"}
		vm["cap_add"] = []interface{}{
			"CAP_NET_BIND_SERVICE", "CAP_CHOWN", "CAP_FOWNER",
			"CAP_SETGID", "CAP_SETUID", "CAP_DAC_OVERRIDE",
		}
		svcs[k] = vm
	}
	m["services"] = svcs
}

func updateVolumes(m map[string]interface{}) {
	volumes, _ := m["volumes"].(map[string]interface{})
	if volumes == nil {
		return
	}
	for volName, volVal := range volumes {
		vm, ok := volVal.(map[string]interface{})
		if !ok || vm == nil {
			vm = make(map[string]interface{})
		}
		vm["name"] = volName
		volumes[volName] = vm
	}
	m["volumes"] = volumes
}

func writeCompose(filePath string, m map[string]interface{}) error {
	o, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, o, 0600)
}

func ZipDirectory(dirPath string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		fileInZip, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		fsFile, err := os.Open(path) // #nosec G304 (CWE-22): Potential file inclusion via variable; no problem since it is called internally
		if err != nil {
			return err
		}
		defer utils.Close(fsFile)

		_, err = io.Copy(fileInZip, fsFile)
		return err
	})

	if err != nil {
		return nil, err
	}

	if err = zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func AssertYamlEquality(t *testing.T, a, b []byte) {
	var m1, m2 map[string]interface{}
	if err := yaml.Unmarshal(a, &m1); err != nil {
		t.Fail()
	}
	if err := yaml.Unmarshal(b, &m2); err != nil {
		t.Fail()
	}
	if !reflect.DeepEqual(m1, m2) {
		fmt.Printf("FIRST YAML: \n\n%v\n", string(a))
		fmt.Printf("\n\nSECOND YAML: \n\n%v\n", string(b))
		t.Fail()
	}
}
