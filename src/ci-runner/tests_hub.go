package main

func TestAll() {
	tr.ExecuteInDir(backendDir, "rm -rf data")
	TestUnits()
	TestComponent()
}

func TestUnits() {
	tr.Log.TaskDescription("Testing units")
	defer tr.Cleanup()
	// TODO it should not be necessary to set a profile for unit tests; the TEST profile should become the default; PROD should become the app store packaged in a container I guess?
	tr.ExecuteInDir(backendDir, "go test -count=1 -tags=unit ./...", "PROFILE=TEST")
}

// TODO !! global var
var isPostgresDbStarted = false

func startPostgresDb() {
	if !isPostgresDbStarted {
		// TODO !! the "dev" compose can be deleted after refactoring
		tr.ExecuteInDir(backendDockerDir, "docker compose -f docker-compose-dev.yml up -d")
		isPostgresDbStarted = true
	}
}

func TestComponent() {
	tr.Log.TaskDescription("Testing backend")
	defer tr.Cleanup()
	tr.ExecuteInDir(backendDir, "go build .")
	startPostgresDb()
	tr.ExecuteInDir(backendDir, "rm -f data/.env")
	tr.StartDaemon(backendDir, "./store", "PROFILE=TEST")
	tr.WaitUntilPortIsReady("8082")
	tr.ExecuteInDir(backendCheckDir, "go test -count=1 -tags=component ./...")
}
