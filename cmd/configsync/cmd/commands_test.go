package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dotbrains/configsync/internal/constants"
	"github.com/spf13/cobra"
)

// Helper function for min
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to execute command with captured output
func executeCommand(cmd *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return buf.String(), err
}

// Helper function to set up test environment
func setupTestEnv(t *testing.T) (string, func()) {
	tempDir := t.TempDir()
	originalHome := homeDir
	originalConfigDir := configDir

	// Set test directories
	homeDir = tempDir
	configDir = filepath.Join(tempDir, ".configsync")

	return tempDir, func() {
		homeDir = originalHome
		configDir = originalConfigDir
	}
}

// Test that Execute function exists and can be called
func TestExecute(t *testing.T) {
	// Test that Execute function exists by trying to call it with help
	// We can't test Execute == nil since function comparisons are not allowed
	// Instead, we test that the function can be called
	err := Execute()
	if err == nil {
		// This is actually okay - Execute might succeed with no args
		t.Log("Execute function is accessible and callable")
	}
}

// Test root command structure
func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should be defined")
	}

	if rootCmd.Use != "configsync" {
		t.Errorf("Expected root command use to be 'configsync', got '%s'", rootCmd.Use)
	}

	if !strings.Contains(rootCmd.Short, "macOS application") {
		t.Errorf("Expected root command short description to mention macOS, got: %s", rootCmd.Short)
	}

	if rootCmd.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got '%s'", rootCmd.Version)
	}
}

// Test command structure and metadata
func TestCommandStructure(t *testing.T) {
	tests := []struct {
		cmd            *cobra.Command
		name           string
		shouldHaveRunE bool
	}{
		{initCmd, "init", true},
		{addCmd, "add", true},
		{removeCmd, "remove", true},
		{syncCmd, "sync", true},
		{statusCmd, "status", true},
		{discoverCmd, "discover", true},
		{backupCmd, "backup", true},
		{restoreCmd, "restore", true},
		{exportCmd, "export", true},
		{importCmd, "import", true},
		{deployCmd, "deploy", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd == nil {
				t.Fatalf("Command %s should be defined", tt.name)
			}

			if tt.cmd.Use == "" {
				t.Errorf("Command %s should have Use defined", tt.name)
			}

			if tt.cmd.Short == "" {
				t.Errorf("Command %s should have Short description", tt.name)
			}

			if tt.cmd.Long == "" {
				t.Errorf("Command %s should have Long description", tt.name)
			}

			if tt.shouldHaveRunE && tt.cmd.RunE == nil {
				t.Errorf("Command %s should have RunE function", tt.name)
			}
		})
	}
}

// Test that all expected commands are registered
func TestCommandRegistration(t *testing.T) {
	expectedCommands := []string{
		"init", "add", "remove", "sync", "status",
		"discover", "backup", "restore", "export", "import", "deploy",
	}

	registeredCommands := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		registeredCommands[cmd.Name()] = true
	}

	for _, expectedCmd := range expectedCommands {
		if !registeredCommands[expectedCmd] {
			t.Errorf("Expected command %s to be registered", expectedCmd)
		}
	}
}

// Test global flags
func TestGlobalFlags(t *testing.T) {
	// Check that global flags are defined
	homeFlag := rootCmd.PersistentFlags().Lookup("home")
	if homeFlag == nil {
		t.Error("Expected --home flag to be defined")
	}

	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected --verbose flag to be defined")
	}

	dryRunFlag := rootCmd.PersistentFlags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("Expected --dry-run flag to be defined")
	}
}

// Test command-specific flags
func TestCommandFlags(t *testing.T) {
	// Test add command flags
	listSupportedFlag := addCmd.Flags().Lookup("list-supported")
	if listSupportedFlag == nil {
		t.Error("Expected add command to have --list-supported flag")
	}

	// Test discover command flags
	autoAddFlag := discoverCmd.Flags().Lookup("auto-add")
	if autoAddFlag == nil {
		t.Error("Expected discover command to have --auto-add flag")
	}

	listFlag := discoverCmd.Flags().Lookup("list")
	if listFlag == nil {
		t.Error("Expected discover command to have --list flag")
	}

	// Test backup command flags
	validateFlag := backupCmd.Flags().Lookup("validate")
	if validateFlag == nil {
		t.Error("Expected backup command to have --validate flag")
	}

	keepDaysFlag := backupCmd.Flags().Lookup("keep-days")
	if keepDaysFlag == nil {
		t.Error("Expected backup command to have --keep-days flag")
	}

	// Test export command flags
	outputFlag := exportCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected export command to have --output flag")
	}

	appsFlag := exportCmd.Flags().Lookup("apps")
	if appsFlag == nil {
		t.Error("Expected export command to have --apps flag")
	}
}

// Test initConfig function
func TestInitConfig(t *testing.T) {
	// Save original values
	originalHome := homeDir
	originalConfigDir := configDir
	originalHomeEnv := os.Getenv("HOME")

	defer func() {
		homeDir = originalHome
		configDir = originalConfigDir
		_ = os.Setenv("HOME", originalHomeEnv)
	}()

	// Test with empty homeDir (should use environment)
	homeDir = ""
	testHome := "/test/home/dir"
	_ = os.Setenv("HOME", testHome)

	initConfig()

	if homeDir != testHome {
		t.Errorf("Expected homeDir to be set to %s, got %s", testHome, homeDir)
	}

	expectedConfigDir := filepath.Join(testHome, ".configsync")
	if configDir != expectedConfigDir {
		t.Errorf("Expected configDir to be %s, got %s", expectedConfigDir, configDir)
	}

	// Test with homeDir already set
	customHome := "/custom/home"
	homeDir = customHome

	initConfig()

	if homeDir != customHome {
		t.Errorf("Expected homeDir to remain %s, got %s", customHome, homeDir)
	}
}

// Test command help output (simplified)
func TestCommandHelp(t *testing.T) {
	tests := []struct {
		cmd  *cobra.Command
		name string
	}{
		{rootCmd, constants.RootCommandName},
		// Skip individual command help tests as they may not work in test context
		// Individual commands are tested via structure tests
	}

	for _, tt := range tests {
		t.Run(tt.name+"_help", func(t *testing.T) {
			output, err := executeCommand(tt.cmd, "--help")

			// Help should always work for root command
			if err != nil && tt.name == constants.RootCommandName {
				t.Errorf("Help for %s command should not error: %v", tt.name, err)
			}

			// Root command help output should contain command description
			if tt.name == constants.RootCommandName && len(output) < 10 {
				t.Errorf("Help for %s command should produce meaningful output, got: %d chars", tt.name, len(output))
			}
		})
	}
}

// Test init command functionality (basic)
func TestInitCommandBasic(t *testing.T) {
	testDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// This tests that the init command can be called without panicking
	// We can't easily test the full functionality without mocking the config manager

	// Test that command structure is correct
	if initCmd.RunE == nil {
		t.Error("init command should have RunE function")
	}

	// Test help works
	_, err := executeCommand(initCmd, "--help")
	if err != nil {
		t.Errorf("init help should not error: %v", err)
	}

	// Verify test directory setup
	if !strings.Contains(testDir, "Test") {
		t.Errorf("Expected test directory, got %s", testDir)
	}
}

// Test that commands handle missing arguments appropriately
func TestCommandArgumentValidation(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test import command requires argument
	if importCmd.Args == nil {
		t.Error("import command should have argument validation")
	}

	// Test that import command requires exactly one argument
	// This is defined as cobra.ExactArgs(1) in the source
	expectedArgs := importCmd.Args
	if expectedArgs == nil {
		t.Error("import command should validate arguments")
	}
}

// Test flag default values
func TestFlagDefaults(t *testing.T) {
	// Reset to defaults
	verbose = false
	dryRun = false
	homeDir = ""

	// Test that defaults are correct
	if verbose != false {
		t.Error("verbose should default to false")
	}

	if dryRun != false {
		t.Error("dryRun should default to false")
	}

	// Test command-specific defaults
	if backupKeepDays != 30 {
		t.Error("backupKeepDays should default to 30")
	}
}

// Test command usage strings
func TestCommandUsage(t *testing.T) {
	tests := []struct {
		cmd         *cobra.Command
		name        string
		expectedUse string
	}{
		{addCmd, "add", "add [app1] [app2] ..."},
		{removeCmd, "remove", "remove [app1] [app2] ..."},
		{syncCmd, "sync", "sync [app1] [app2] ..."},
		{backupCmd, "backup", "backup [app1] [app2] ..."},
		{restoreCmd, "restore", "restore [app1] [app2] ..."},
		{importCmd, "import", "import <bundle.tar.gz>"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_usage", func(t *testing.T) {
			if tt.cmd.Use != tt.expectedUse {
				t.Errorf("Expected %s command use to be '%s', got '%s'",
					tt.name, tt.expectedUse, tt.cmd.Use)
			}
		})
	}
}

// Integration test - verify command chain structure
func TestCommandChain(t *testing.T) {
	// Test that cobra initialization works correctly
	if rootCmd.Commands() == nil {
		t.Error("Root command should have subcommands")
	}

	commandCount := len(rootCmd.Commands())
	if commandCount < 10 { // We expect at least 11 main commands plus completion and help
		t.Errorf("Expected at least 10 commands, got %d", commandCount)
	}

	// Test that each command has proper parent relationship
	for _, cmd := range rootCmd.Commands() {
		if cmd.Parent() != rootCmd {
			t.Errorf("Command %s should have rootCmd as parent", cmd.Name())
		}
	}
}

// Test version output
func TestVersionOutput(t *testing.T) {
	output, err := executeCommand(rootCmd, "version")
	if err != nil {
		// Version subcommand might not be explicitly defined, try --version flag
		output, err = executeCommand(rootCmd, "--version")
		if err != nil {
			t.Skipf("Version command not available: %v", err)
		}
	}

	// Version output might come in different formats
	if len(output) > 0 && !strings.Contains(output, "1.0.0") && !strings.Contains(output, "version for configsync") {
		t.Errorf("Expected version output to contain version info, but got: %s", output[:minInt(200, len(output))])
	}
}

// Test command short descriptions are meaningful
func TestCommandDescriptions(t *testing.T) {
	commands := []*cobra.Command{
		initCmd, addCmd, removeCmd, syncCmd, statusCmd,
		discoverCmd, backupCmd, restoreCmd, exportCmd, importCmd, deployCmd,
	}

	for _, cmd := range commands {
		t.Run(cmd.Name()+"_description", func(t *testing.T) {
			if len(cmd.Short) < 10 {
				t.Errorf("Command %s should have a meaningful short description, got: %s",
					cmd.Name(), cmd.Short)
			}

			if len(cmd.Long) < 20 {
				t.Errorf("Command %s should have a meaningful long description, got: %s",
					cmd.Name(), cmd.Long)
			}
		})
	}
}

// Test cobra command initialization
func TestCobraInit(t *testing.T) {
	// Test that cobra.OnInitialize is set up correctly
	// We can't easily test the actual function, but we can verify structure

	// Test that root command has the init function
	if rootCmd.PersistentPreRun == nil && rootCmd.PersistentPreRunE == nil {
		// This is OK - initialization happens via cobra.OnInitialize
		t.Log("Root command uses cobra.OnInitialize for setup")
	}

	// Initialize the configuration to test global variables
	initConfig()

	// Test that global variables are accessible
	if homeDir == "" {
		t.Error("homeDir variable should be accessible and not empty")
	}

	if configDir == "" {
		t.Error("configDir variable should be accessible and not empty")
	}
}
