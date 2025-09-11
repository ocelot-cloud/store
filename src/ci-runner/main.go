package main

import (
	"os"
	"path/filepath"

	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
)

var (
	tr = taskrunner.GetTaskRunner()

	srcDir      = getAbsoluteParentDir()
	backendDir  = srcDir + "/backend"
	ciRunnerDir = srcDir + "/ci-runner"

	backendDockerDir = backendDir + "/docker"
	backendCheckDir  = backendDir + "/check"
	backendSetupDir  = backendDir + "/setup"

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

	// TODO !! get rid of those env variables; only use "PROFILE=TEST" when doing component test
	// TODO !! also re-add production test
	// TODO !! when deploying to hetzner server, I should have a
	tr.Config.DefaultEnvironmentVariables = []string{"USE_MOCK_EMAIL_CLIENT=true", "RUN_NATIVELY=true", "LOG_LEVEL=DEBUG"}

	rootCmd := &cobra.Command{
		Use:   "ci-runner",
		Short: "ci-runner is a service that runs CI jobs",
	}

	testCmd.AddCommand(testUnitsCmd, testComponentCmd, testAllCmd)
	buildCmd.AddCommand(buildBackendCmd)
	rootCmd.AddCommand(testCmd, updateCmd, deployCmd, analyzeCmd, buildCmd)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	err := rootCmd.Execute()
	if err != nil {
		tr.ExitWithError()
	}
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates project dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		update()
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy current version to server",
	Run: func(cmd *cobra.Command, args []string) {
		deploy()
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "run tests",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var testUnitsCmd = &cobra.Command{
	Use:   "units",
	Short: "run unit tests",
	Run: func(cmd *cobra.Command, args []string) {
		TestUnits()
	},
}

var testComponentCmd = &cobra.Command{
	Use:   "component",
	Short: "run component tests",
	Run: func(cmd *cobra.Command, args []string) {
		TestComponent()
	},
}

var testAllCmd = &cobra.Command{
	Use:   "all",
	Short: "run all tests",
	Run: func(cmd *cobra.Command, args []string) {
		TestAll()
	},
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "runs code analysis tools",
	Run: func(cmd *cobra.Command, args []string) {
		u.AnalyzeCode(tr, backendDir)
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
	Short: "builds the backend including all production code and test code to find compilation errors",
	Run: func(cmd *cobra.Command, args []string) {
		u.BuildWholeGoProject(tr, backendDir)
	},
}
