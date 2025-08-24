package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ExecCmd(command string, args ...string) (bool, string, error) {
	cmd := exec.Command(command, args...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(stderr.String())
		return false, "", fmt.Errorf("rndc failed: %s", stderr.String())
	}
	return true, out.String(), nil
}
