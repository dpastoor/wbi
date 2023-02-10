package workbench

import (
	"errors"
	"fmt"
	"os/exec"
)

func VerifyWorkbench() (bool, error) {

	cmd := exec.Command("/bin/sh", "-c", "sudo rstudio-server version")
	stdout, err := cmd.Output()

	if err != nil {
		return false, errors.New("workbench installation not detected. Please install Workbench first by following the instructions at: https://docs.posit.co/rsw/installation/")
	}

	fmt.Println("Workbench installation detected: ", string(stdout))
	return true, nil

}
