package run

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/yourusername/quick-ci/internal/common"
)

// LoadPRCommands loads a PR with commands from a JSON file
func LoadPRCommands(filepath string) (*common.PRWithCommands, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var pr common.PRWithCommands
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &pr, nil
}

// CommandExecutor is a function that executes a shell command
type CommandExecutor func(cmd string) error

// ExecuteCommands runs a list of commands using the provided executor
func ExecuteCommands(commands []string, executor CommandExecutor) error {
	for _, cmd := range commands {
		if err := executor(cmd); err != nil {
			return fmt.Errorf("command failed: %s: %w", cmd, err)
		}
	}
	return nil
}

// ShellExecutor executes a command via sh -c
func ShellExecutor(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// ExecuteCommandWithOutput runs a command and captures its output
func ExecuteCommandWithOutput(cmd string) common.CommandResult {
	c := exec.Command("sh", "-c", cmd)

	var output bytes.Buffer
	c.Stdout = &output
	c.Stderr = &output

	err := c.Run()

	// Also print to console for real-time feedback
	fmt.Print(output.String())

	return common.CommandResult{
		Command: cmd,
		Success: err == nil,
		Output:  output.String(),
	}
}

// ExecutePhase runs all commands in a phase and returns the result
func ExecutePhase(name string, commands []string) common.PhaseResult {
	result := common.PhaseResult{
		Name:     name,
		Success:  true,
		Commands: make([]common.CommandResult, 0, len(commands)),
	}

	if len(commands) == 0 {
		return result
	}

	fmt.Printf("Running %s commands...\n", name)

	for _, cmd := range commands {
		cmdResult := ExecuteCommandWithOutput(cmd)
		result.Commands = append(result.Commands, cmdResult)

		if !cmdResult.Success {
			result.Success = false
			return result // Stop at first failure
		}
	}

	return result
}
