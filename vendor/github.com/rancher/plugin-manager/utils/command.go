package utils

import (
	"os"
	"os/exec"
	"strings"
)

func execCmdNoStdoutNoStderr(cmd *exec.Cmd) error {
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func execCmdNoStderr(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func buildCommand(cmdStr string) *exec.Cmd {
	cmd := strings.Split(strings.TrimSpace(cmdStr), " ")
	return exec.Command(cmd[0], cmd[1:]...)
}

// RunNoStderr runs the given command with no output to stderr
func RunNoStderr(cmdStr string) error {
	cmd := buildCommand(cmdStr)
	return execCmdNoStderr(cmd)
}

// RunNoStdoutNoStderr runs the given command with no output to stdout and stderr
func RunNoStdoutNoStderr(cmdStr string) error {
	cmd := buildCommand(cmdStr)
	return execCmdNoStdoutNoStderr(cmd)
}

// RunCombined is used to run the given command with combined stdout and stderr outputs
func RunCombined(cmdStr string) ([]byte, error) {
	cmd := buildCommand(cmdStr)
	return cmd.CombinedOutput()
}

// RunOutput is used to run the given command and return output from stdout
func RunOutput(cmdStr string) ([]byte, error) {
	cmd := buildCommand(cmdStr)
	return cmd.Output()
}

// Run with given command with stdout and stderr connected
func Run(cmdStr string) error {
	cmd := buildCommand(cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
