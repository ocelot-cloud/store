package main

import (
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	srcDir      = getAbsoluteParentDir()
	frontendDir = srcDir + "/frontend"
	backendDir  = srcDir + "/backend"

	backendToolsDir = backendDir + "/tools"
	backendCheckDir = backendDir + "/check"

	acceptanceTestsDir = srcDir + "/cypress"
	ciRunnerDir        = srcDir + "/ci-runner"
)

func getAbsoluteParentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Dir(wd)
}

func main() {
	tr.HandleSignals()
	tr.CustomCleanupFunc = func() {
		var potentiallyExistingProcesses = []string{
			"vue-tr-service",
			"vue-service",
			"vite",
		}
		tr.KillProcesses(potentiallyExistingProcesses)
	}
	defer tr.Cleanup()

	tr.DefaultEnvs = []string{"USE_MOCK_EMAIL_CLIENT=true", "LOG_LEVEL=DEBUG"}

	rootCmd := &cobra.Command{
		Use:   "ci-runner",
		Short: "ci-runner is a service that runs CI jobs",
	}

	rootCmd.AddCommand(runCmd, testCmd, updateCmd, deployCmd, downloadDependenciesCmd, analyzeCmd)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	err := rootCmd.Execute()
	if err != nil {
		tr.CleanupAndExitWithError()
	}
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the application locally",
	Run: func(cmd *cobra.Command, args []string) {
		build()
		tr.ExecuteInDir(backendDir, "bash -c './store || true'")
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

		build()
		var remoteHomeDir = "/home/user"
		executeOnServer("systemctl stop store")
		executeOnServer("mkdir -p %s/store", remoteHomeDir)
		rsyncCmd := fmt.Sprintf("rsync -avz --delete assets store dist ocelot:%s/store/", remoteHomeDir)
		tr.ExecuteInDir(backendDir, rsyncCmd)
		executeOnServer("chown -R user:user %s/store", remoteHomeDir)

		executeOnServer("chmod -R 700 %s/store", remoteHomeDir)

		executeOnServer("systemctl start store")
		executeOnServer("nmap -p 8082 localhost")
	},
}

func executeOnServer(command string, args ...string) {
	var cmd string
	if len(args) == 0 {
		cmd = command
	} else {
		cmd = fmt.Sprintf(command, args[0])
	}
	println("executing: " + cmd)
	tr.Execute("ssh ocelot \"" + cmd + "\"")
}

var testCmd = &cobra.Command{
	Use:   "test [" + strings.Join(getKeys(hubTestTypes), "/") + "]",
	Short: "Run tests",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputTestType := args[0]
		if _, exists := hubTestTypes[inputTestType]; !exists {
			tr.ColoredPrintln("\nerror: unknown hub test type: %s\n", inputTestType)
			tr.ColoredPrintln("valid args: %s\n", strings.Join(getKeys(hubTestTypes), ", "))
			os.Exit(1)
		} else {
			hubTestTypes[inputTestType]()
		}
		tr.ColoredPrintln("\nSuccess! Hub tests passed.\n")
	},
}

var hubTestTypes = map[string]func(){
	"unit":       func() { TestUnits() },
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
		tr.PrintTaskDescription("downloading dependencies")
		tr.ExecuteInDir(backendDir, "go mod tidy")
		tr.ExecuteInDir(frontendDir, "npm install")
		tr.ExecuteInDir(acceptanceTestsDir, "npm install")
	},
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "runs code analysis tools",
	Run: func(cmd *cobra.Command, args []string) {
		signal.Ignore(syscall.SIGPIPE)
		utils.AnalyzeCode(backendDir)
	},
}
