package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// LoadPRCommands loads a PR with commands from a JSON file
func LoadPRCommands(filepath string) (*PRWithCommands, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var pr PRWithCommands
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
