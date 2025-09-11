package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func TestAll() {
	TestUnits()
	TestComponent()
}

func TestUnits() {
	tr.Log.TaskDescription("Testing units")
	defer tr.Cleanup()
	// TODO it should not be necessary to set a profile for unit tests; the TEST profile should become the default; PROD should become the app store packaged in a container I guess?
	// TODO !! put the "delete mocks and wire_gen" logic from cloud in shared and use it here
	// TODO !! first wire and then mockery or vise versa? use cloud approach, maybe put to "shared"
	tr.ExecuteInDir(backendDir, "wire")
	tr.ExecuteInDir(backendDir, "go test -count=1 -tags=unit ./...", "PROFILE=TEST")
}

func TestComponent() {
	tr.Log.TaskDescription("Testing backend")
	defer tr.Cleanup()
	tr.ExecuteInDir(backendDir, "go build -installsuffix cgo", "CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64")
	// TODO !! abstract paths and image names etc
	tr.ExecuteInDir(backendDir, "docker build -t ocelotcloud/store:local -f docker/Dockerfile .")
	tr.ExecuteInDir(backendDockerDir, "docker compose -f docker-compose-dev.yml up -d", "PROFILE=TEST")
	waitForHealthEndpoint()
	tr.ExecuteInDir(backendCheckDir, "go test -count=1 -tags=component ./...")
}

func waitForHealthEndpoint() {
	deadline := time.Now().Add(10 * time.Second)
	client := &http.Client{Timeout: 1 * time.Second}
	tr.Log.Info("waiting for /api/healthcheck ")
	for time.Now().Before(deadline) {
		// TODO !! abstract "http://localhost"
		resp, err := client.Get("http://localhost/api/healthcheck")
		if err == nil {
			var v struct {
				Status string `json:"status"`
			}
			if resp.StatusCode == 200 && json.NewDecoder(resp.Body).Decode(&v) == nil && v.Status == "ok" {
				resp.Body.Close()
				tr.Log.Info(" -> healthcheck ok")
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		fmt.Print(".")
		time.Sleep(500 * time.Millisecond)
	}
	tr.Log.Error("healthcheck failed")
	os.Exit(1)
}

// TODO !! re-check, which assets can be deleted; also make sure that sample assets are not loaded in production mode (maybe write a test that database is empty or so in prod mode?)

func update() {
	tr.ExecuteInDir(ciRunnerDir, "go get -u ./...")
	tr.ExecuteInDir(ciRunnerDir, "go mod tidy")
	tr.ExecuteInDir(ciRunnerDir, "go build")

	tr.ExecuteInDir(backendDir, "go get -u ./...")
	tr.ExecuteInDir(backendDir, "go mod tidy")
	tr.ExecuteInDir(backendDir, "go build")
}

func deploy() {
	/* TODO !! re-implement: upload to dockerhub and pull from there to server
	var prompt = "Are you sure you want to replace the current production version of the App Store?"
	tr.PromptForContinuation(prompt)
	tr.ExecuteInDir(backendDir, "go build -a -installsuffix cgo", "CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64")
	executeOnServer("docker rm -f store")
	rsyncCmd := fmt.Sprintf("rsync -avz --delete docker/Dockerfile docker/docker-compose.yml assets store dist %s:", sshHost)
	tr.ExecuteInDir(backendDir, rsyncCmd)
	// TODO !! why two times docker compose up? Can be removed?
	executeOnServer("docker compose up -d")
	executeOnServer("docker compose up -d --build --force-recreate --remove-orphans store")
	*/
}

func executeOnServer(command string) {
	sshCommand := fmt.Sprintf("ssh %s %s", sshHost, command)
	tr.Execute(sshCommand)
}

// TODO !! implement "deploy local prod" command, delete "data" folder after test was run
