package main

import (
	"github.com/ocelot-cloud/task-runner"
)

func TestHubAll() {
	tr.ExecuteInDir(backendDir, "rm -rf data")
	TestBackend()
	TestAcceptance()
}

func TestUnits() {
	tr.PrintTaskDescription("Testing units")
	defer tr.Cleanup()
	startCockroachDb()

	tr.ExecuteInDir(backendToolsDir, "go test -count=1 .")
	tr.ExecuteInDir(backendCheckDir, "go test -count=1 -tags=unit .")
}

var isPostgresDbStarted = false

func startCockroachDb() {
	if !isPostgresDbStarted {
		tr.ExecuteInDir(ciRunnerDir, "docker compose up -d")
		isPostgresDbStarted = true
	}
}

func TestBackend() {
	TestUnits()

	tr.PrintTaskDescription("Testing backend")
	defer tr.Cleanup()
	tr.ExecuteInDir(backendDir, "go build .")
	startCockroachDb()
	tr.ExecuteInDir(backendDir, "rm -f data/.env")
	tr.ExecuteInDir(backendDir, "bash -c './store || true'") // creates default data/.env file and exit
	tr.StartDaemon(backendDir, "./store", "PROFILE=TEST")
	tr.WaitUntilPortIsReady("8082")
	tr.ExecuteInDir(backendCheckDir, "go test -count=1 -tags=component ./...")
}

func TestAcceptance() {
	tr.PrintTaskDescription("Testing acceptance")
	defer tr.Cleanup()
	build()
	startCockroachDb()
	tr.ExecuteInDir(backendDir, "bash -c './store || true'") // creates data/.env file and exit
	tr.StartDaemon(backendDir, "./store")
	tr.WaitUntilPortIsReady("8082")
	tr.ExecuteInDir(backendCheckDir, "go test -count=1 -tags=acceptance ./...")
	tr.ExecuteInDir(acceptanceTestsDir, "npx cypress run --spec cypress/e2e/hub.cy.ts --headless")
}

func build() {
	tr.ExecuteInDir(backendDir, "rm -rf data dist")
	tr.ExecuteInDir(frontendDir, "npm run build")
	tr.ExecuteInDir(frontendDir, "cp -r ./dist "+backendDir)
	tr.ExecuteInDir(backendDir, "go build")
}
