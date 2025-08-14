package main

func TestHubAll() {
	tr.ExecuteInDir(backendDir, "rm -rf data")
	TestBackend()
	TestAcceptance()
}

func TestUnits() {
	tr.Log.TaskDescription("Testing units")
	defer tr.Cleanup()
	startCockroachDb()

	// TODO it should not be necessary to set a profile for unit tests; the TEST profile should become the default; PROD should become the app store packaged in a container I guess?
	tr.ExecuteInDir(backendDir, "go test -count=1 -tags=unit ./...", "PROFILE=TEST")
}

var isPostgresDbStarted = false

func startCockroachDb() {
	if !isPostgresDbStarted {
		tr.ExecuteInDir(backendDockerDir, "docker compose -f docker-compose-dev.yml up -d")
		isPostgresDbStarted = true
	}
}

func TestBackend() {
	TestUnits()

	tr.Log.TaskDescription("Testing backend")
	defer tr.Cleanup()
	tr.ExecuteInDir(backendDir, "go build .")
	startCockroachDb()
	tr.ExecuteInDir(backendDir, "rm -f data/.env")
	tr.StartDaemon(backendDir, "./store", "PROFILE=TEST")
	tr.WaitUntilPortIsReady("8082")
	tr.ExecuteInDir(backendCheckDir, "go test -count=1 -tags=component ./...")
}

func TestAcceptance() {
	tr.Log.TaskDescription("Testing acceptance")
	defer tr.Cleanup()
	build()
	startCockroachDb()
	tr.StartDaemon(backendDir, "./store", "PROFILE=TEST")
	tr.WaitUntilPortIsReady("8082")
	tr.ExecuteInDir(backendCheckDir, "go test -count=1 -tags=acceptance ./...")
	tr.ExecuteInDir(acceptanceTestsDir, "npx cypress run --spec cypress/e2e/hub.cy.ts --headless")
}

func build() {
	subBuild()
	tr.ExecuteInDir(backendDir, "go build")
}

func buildForDocker() {
	subBuild()
	tr.ExecuteInDir(backendDir, "go build -a -installsuffix cgo", "CGO_ENABLED=0", "GOOS=linux", "GOARCH=amd64")
}

func subBuild() {
	tr.Remove(backendDataDir, backendDistDir)
	tr.ExecuteInDir(frontendDir, "npm run build")
	tr.Copy(frontendDir, "dist", backendDir)
}
