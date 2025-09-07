package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
)

var (
	tr = taskrunner.GetTaskRunner()

	srcDir           = getAbsoluteParentDir()
	frontendDir      = srcDir + "/frontend"
	backendDir       = srcDir + "/backend"
	backendDockerDir = backendDir + "/docker"
	backendDataDir   = backendDir + "/data"
	backendDistDir   = backendDir + "/dist"
	backendCheckDir  = backendDir + "/check"

	acceptanceTestsDir = srcDir + "/cypress"
	ciRunnerDir        = srcDir + "/ci-runner"

	sshHost = "store"
)

func getAbsoluteParentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Dir(wd)
}

func main() {
	tr.EnableAbortForKeystrokeControlPlusC()
	tr.Config.CleanupFunc = func() {
		var potentiallyExistingProcesses = []string{
			"vue-tr-service",
			"vue-service",
			"vite",
		}
		tr.KillProcesses(potentiallyExistingProcesses)
	}
	defer tr.Cleanup()

	tr.Config.DefaultEnvironmentVariables = []string{"USE_MOCK_EMAIL_CLIENT=true", "RUN_NATIVELY=true", "LOG_LEVEL=DEBUG"}

	rootCmd := &cobra.Command{
		Use:   "ci-runner",
		Short: "ci-runner is a service that runs CI jobs",
	}

	buildCmd.AddCommand(buildBackendCmd)
	rootCmd.AddCommand(runCmd, testCmd, updateCmd, deployCmd, downloadDependenciesCmd, analyzeCmd, buildCmd)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	err := rootCmd.Execute()
	if err != nil {
		tr.ExitWithError()
	}
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the application locally",
	Run: func(cmd *cobra.Command, args []string) {
		build()
		tr.ExecuteInDir(backendDir, "./store")
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates project dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		tr.ExecuteInDir(ciRunnerDir, "go get -u ./...")
		tr.ExecuteInDir(ciRunnerDir, "go mod tidy")
		tr.ExecuteInDir(ciRunnerDir, "go build")

		tr.ExecuteInDir(backendDir, "go get -u ./...")
		tr.ExecuteInDir(backendDir, "go mod tidy")
		tr.ExecuteInDir(backendDir, "go build")

		tr.ExecuteInDir(frontendDir, "yarn upgrade --latest --pattern \"*\"")
		tr.ExecuteInDir(frontendDir, "yarn add vue@latest vite@latest")
		tr.ExecuteInDir(frontendDir, "yarn install")
		tr.ExecuteInDir(frontendDir, "yarn build")

		tr.ExecuteInDir(acceptanceTestsDir, "npx npm-check-updates -u")
		tr.ExecuteInDir(acceptanceTestsDir, "npm install cypress@latest")
		tr.ExecuteInDir(acceptanceTestsDir, "npm install")
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy current version to server",
	Run: func(cmd *cobra.Command, args []string) {
		var prompt = "Are you sure you want to replace the current production version of the App Store?"
		tr.PromptForContinuation(prompt)
		buildForDocker()
		executeOnServer("docker rm -f store")
		rsyncCmd := fmt.Sprintf("rsync -avz --delete docker/Dockerfile docker/docker-compose.yml assets store dist %s:", sshHost)
		tr.ExecuteInDir(backendDir, rsyncCmd)
		executeOnServer("docker compose up -d")
		executeOnServer("docker compose up -d --build --force-recreate --remove-orphans store")
	},
}

func executeOnServer(command string) {
	sshCommand := fmt.Sprintf("ssh %s %s", sshHost, command)
	tr.Execute(sshCommand)
}

var testCmd = &cobra.Command{
	Use:   "test [" + strings.Join(getKeys(hubTestTypes), "/") + "]",
	Short: "Run tests",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputTestType := args[0]
		if _, exists := hubTestTypes[inputTestType]; !exists {
			tr.Log.Error("\nerror: unknown hub test type: %s\n", inputTestType)
			tr.Log.Error("valid args: %s\n", strings.Join(getKeys(hubTestTypes), ", "))
			os.Exit(1)
		} else {
			hubTestTypes[inputTestType]()
		}
		tr.Log.Info(("\nSuccess! Hub tests passed.\n"))
	},
}

var hubTestTypes = map[string]func(){
	"units":      func() { TestUnits() },
	"backend":    func() { TestBackend() },
	"acceptance": func() { TestAcceptance() },
	"all":        func() { TestHubAll() },
}

func getKeys(m map[string]func()) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

var downloadDependenciesCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads application dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		tr.Log.Info("downloading dependencies")
		tr.ExecuteInDir(backendDir, "go mod tidy")
		tr.ExecuteInDir(frontendDir, "npm install")
		tr.ExecuteInDir(acceptanceTestsDir, "npm install")
	},
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "runs code analysis tools",
	Run: func(cmd *cobra.Command, args []string) {
		utils.AnalyzeCode(tr, backendDir)
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var buildBackendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Builds the backend",
	Run: func(cmd *cobra.Command, args []string) {
		utils.BuildWholeGoProject(tr, backendDir)
	},
}
