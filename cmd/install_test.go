package cmd

import (
	"fmt"
	"testing"

	"github.com/sol-eng/wbi/internal/languages"
	"github.com/sol-eng/wbi/internal/operatingsystem"
	"github.com/sol-eng/wbi/internal/quarto"
	"github.com/stretchr/testify/assert"
)

// TestInstallParamsValidate tests the install command parameters
func TestInstallParamsValidate(t *testing.T) {
	// Determine OS
	osType, err := operatingsystem.DetectOS()
	if err != nil {
		t.Fatalf("issue detecting OS: %v", err)
	}

	validRVersions, err := languages.RetrieveValidRVersions()
	if err != nil {
		t.Fatalf("failed to retrieve valid R versions: %v", err)
	}
	validPythonVersions, err := languages.RetrieveValidPythonVersions(osType)
	if err != nil {
		t.Fatalf("failed to retrieve valid Python versions: %v", err)
	}
	var validQuartoVersions []string

	for pagenum := 1; pagenum <= 5; pagenum++ {
		pagedQuartionVerions, err := quarto.RetrieveValidQuartoVersions(pagenum)
		if err != nil {
			t.Fatalf("error retrieving valid Quarto versions: %v", err)
		}
		validQuartoVersions = append(validQuartoVersions, pagedQuartionVerions...)
		if len(validQuartoVersions) > 10 {
			pagenum = 5

		}
	}
	if err != nil {
		t.Fatalf("failed to retrieve valid Quarto versions: %v", err)
	}

	tests := map[string]struct {
		args        []string
		flags       installOpts
		expectError string
	}{
		// general arguement tests
		"no argument or key": {
			args:        []string{},
			flags:       installOpts{},
			expectError: "no arguments provided, please provide one argument",
		},
		"too many arguments": {
			args:        []string{"workbench", "r"},
			flags:       installOpts{},
			expectError: "too many arguments provided, please provide only one argument",
		},
		// r argument tests
		"r argument only succeeds": {
			args:        []string{"r"},
			flags:       installOpts{},
			expectError: "",
		},
		"r argument with a symlink flag succeeds": {
			args:        []string{"r"},
			flags:       installOpts{symlink: true},
			expectError: "",
		},
		"r argument with a valid r version and a symlink flag succeeds": {
			args:        []string{"r"},
			flags:       installOpts{symlink: true, versions: validRVersions[0:1]},
			expectError: "",
		},
		"r argument with an invalid r version and a symlink flag fails": {
			args:        []string{"r"},
			flags:       installOpts{symlink: true, versions: []string{"1.6.1"}},
			expectError: "version 1.6.1 is not a valid R version",
		},
		"r argument with single valid version flag succeeds": {
			args:        []string{"r"},
			flags:       installOpts{versions: validRVersions[0:1]},
			expectError: "",
		},
		"r argument with single invalid version flag fails": {
			args:        []string{"r"},
			flags:       installOpts{versions: []string{"1.6.1"}},
			expectError: "version 1.6.1 is not a valid R version",
		},
		"r argument with multiple valid version flags succeeds": {
			args:        []string{"r"},
			flags:       installOpts{versions: validRVersions[0:2]},
			expectError: "",
		},
		"r argument with multiple invalid version flags fails": {
			args:        []string{"r"},
			flags:       installOpts{versions: []string{"1.6.1", "2.2.2"}},
			expectError: "version 1.6.1 is not a valid R version",
		},
		"r argument with one valid and one invalid version flags fails": {
			args:        []string{"r"},
			flags:       installOpts{versions: []string{validRVersions[0], "2.2.2"}},
			expectError: "version 2.2.2 is not a valid R version",
		},
		"r argument with path flag fails": {
			args:        []string{"r"},
			flags:       installOpts{path: "/opt/python/3.6.1/bin/python"},
			expectError: "the path flag is only supported for jupyter",
		},
		"r argument with addToPATH flag fails": {
			args:        []string{"r"},
			flags:       installOpts{addToPATH: true},
			expectError: "the add-to-path flag is only supported for python",
		},
		// python argument tests
		"python argument only succeeds": {
			args:        []string{"python"},
			flags:       installOpts{},
			expectError: "",
		},
		"python argument with a symlink flag fails": {
			args:        []string{"python"},
			flags:       installOpts{symlink: true},
			expectError: "the symlink flag is only supported for r",
		},
		"python argument with a addToPATH flag succeeds": {
			args:        []string{"python"},
			flags:       installOpts{addToPATH: true},
			expectError: "",
		},
		"python argument with a valid python version and a addToPATH flag succeeds": {
			args:        []string{"python"},
			flags:       installOpts{addToPATH: true, versions: validPythonVersions[0:1]},
			expectError: "",
		},
		"python argument with an invalid python version and a addToPATH flag fails": {
			args:        []string{"python"},
			flags:       installOpts{addToPATH: true, versions: []string{"2.11.1"}},
			expectError: "version 2.11.1 is not a valid Python version",
		},
		"python argument with single valid version flag succeeds": {
			args:        []string{"python"},
			flags:       installOpts{versions: validPythonVersions[0:1]},
			expectError: "",
		},
		"python argument with single invalid version flag fails": {
			args:        []string{"python"},
			flags:       installOpts{versions: []string{"1.6.1"}},
			expectError: "version 1.6.1 is not a valid Python version",
		},
		"python argument with multiple valid version flags succeeds": {
			args:        []string{"python"},
			flags:       installOpts{versions: validPythonVersions[0:2]},
			expectError: "",
		},
		"python argument with multiple invalid version flags fails": {
			args:        []string{"python"},
			flags:       installOpts{versions: []string{"3.15.1", "2.2.2"}},
			expectError: "version 3.15.1 is not a valid Python version",
		},
		"python argument with one valid and one invalid version flags fails": {
			args:        []string{"python"},
			flags:       installOpts{versions: []string{validPythonVersions[0], "3.3.2"}},
			expectError: "version 3.3.2 is not a valid Python version",
		},
		"python argument with path flag fails": {
			args:        []string{"python"},
			flags:       installOpts{path: "/opt/python/3.6.1/bin/python"},
			expectError: "the path flag is only supported for jupyter",
		},
		// quarto argument tests
		"quarto argument only succeeds": {
			args:        []string{"quarto"},
			flags:       installOpts{},
			expectError: "",
		},
		"quarto argument with a symlink flag succeeds": {
			args:        []string{"quarto"},
			flags:       installOpts{symlink: true},
			expectError: "",
		},
		"quarto argument with a valid quarto version and a symlink flag succeeds": {
			args:        []string{"quarto"},
			flags:       installOpts{symlink: true, versions: validQuartoVersions[0:1]},
			expectError: "",
		},
		"quarto argument with an invalid quarto version and a symlink flag fails": {
			args:        []string{"quarto"},
			flags:       installOpts{symlink: true, versions: []string{"0.6.1"}},
			expectError: "version 0.6.1 is not a valid Quarto version",
		},
		"quarto argument with single valid version flag succeeds": {
			args:        []string{"quarto"},
			flags:       installOpts{versions: validQuartoVersions[0:1]},
			expectError: "",
		},
		"quarto argument with single invalid version flag fails": {
			args:        []string{"quarto"},
			flags:       installOpts{versions: []string{"0.6.1"}},
			expectError: "version 0.6.1 is not a valid Quarto version",
		},
		"quarto argument with multiple valid version flags succeeds": {
			args:        []string{"quarto"},
			flags:       installOpts{versions: validQuartoVersions[0:2]},
			expectError: "",
		},
		"quarto argument with multiple invalid version flags fails": {
			args:        []string{"quarto"},
			flags:       installOpts{versions: []string{"0.6.1", "0.2.2"}},
			expectError: "version 0.6.1 is not a valid Quarto version",
		},
		"quarto argument with one valid and one invalid version flags fails": {
			args:        []string{"quarto"},
			flags:       installOpts{versions: []string{validQuartoVersions[0], "0.2.2"}},
			expectError: "version 0.2.2 is not a valid Quarto version",
		},
		"quarto argument with path flag fails": {
			args:        []string{"quarto"},
			flags:       installOpts{path: "/opt/python/3.6.1/bin/python"},
			expectError: "the path flag is only supported for jupyter",
		},
		"quarto argument with addToPATH flag fails": {
			args:        []string{"quarto"},
			flags:       installOpts{addToPATH: true},
			expectError: "the add-to-path flag is only supported for python",
		},
		// Workbench argument tests
		"workbench argument only succeeds": {
			args:        []string{"workbench"},
			flags:       installOpts{},
			expectError: "",
		},
		"workbench argument with a symlink flag fails": {
			args:        []string{"workbench"},
			flags:       installOpts{symlink: true},
			expectError: "the symlink flag is only supported for r",
		},
		"workbench argument with a version flag fails": {
			args:        []string{"workbench"},
			flags:       installOpts{versions: []string{"2.11.1"}},
			expectError: "workbench does not support specifying versions",
		},
		"workbench argument with a path flag fails": {
			args:        []string{"workbench"},
			flags:       installOpts{path: "/opt/python/3.6.1/bin/python"},
			expectError: "the path flag is only supported for jupyter",
		},
		"workbench argument with a add-to-path flag fails": {
			args:        []string{"workbench"},
			flags:       installOpts{addToPATH: true},
			expectError: "the add-to-path flag is only supported for python",
		},
		// Pro Drivers argument tests
		"prodrivers argument only succeeds": {
			args:        []string{"prodrivers"},
			flags:       installOpts{},
			expectError: "",
		},
		"prodrivers argument with a symlink flag fails": {
			args:        []string{"prodrivers"},
			flags:       installOpts{symlink: true},
			expectError: "the symlink flag is only supported for r",
		},
		"prodrivers argument with a version flag fails": {
			args:        []string{"prodrivers"},
			flags:       installOpts{versions: []string{"2.11.1"}},
			expectError: "prodrivers does not support specifying versions",
		},
		"prodrivers argument with a path flag fails": {
			args:        []string{"prodrivers"},
			flags:       installOpts{path: "/opt/python/3.6.1/bin/python"},
			expectError: "the path flag is only supported for jupyter",
		},
		"prodrivers argument with a add-to-path flag fails": {
			args:        []string{"prodrivers"},
			flags:       installOpts{addToPATH: true},
			expectError: "the add-to-path flag is only supported for python",
		},
		// Jupyter argument tests
		"jupyter argument only succeeds": {
			args:        []string{"jupyter"},
			flags:       installOpts{},
			expectError: "",
		},
		"jupyter argument with an invalid path flag succeeds in validating but returns the path does not exit": {
			args:        []string{"jupyter"},
			flags:       installOpts{path: "/opt/python/3.6.1/bin/python"},
			expectError: "the path provided does not exist",
		},
		"jupyter argument with a symlink flag fails": {
			args:        []string{"jupyter"},
			flags:       installOpts{symlink: true},
			expectError: "the symlink flag is only supported for r",
		},
		"jupyter argument with a version flag fails": {
			args:        []string{"jupyter"},
			flags:       installOpts{versions: []string{"2.11.1"}},
			expectError: "jupyter does not support specifying versions",
		},
		"jupyter argument with a add-to-path flag fails": {
			args:        []string{"jupyter"},
			flags:       installOpts{addToPATH: true},
			expectError: "the add-to-path flag is only supported for python",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			installCmd := newInstallCmd()
			// set the flags
			installCmd.opts = tc.flags
			// run validation
			err := installCmd.opts.Validate(tc.args)

			if err != nil && tc.expectError != "" {
				// if we expect an error, check that it contains the expected error
				assert.Containsf(t, err.Error(), tc.expectError, "expected error containing %q, got %s", tc.expectError, err)
			} else if err != nil && tc.expectError == "" {
				// if we expect no error but get one then fail
				t.Fatalf("expected no error, but got %s", err)
			} else if err == nil && tc.expectError != "" {
				// if we expect an error but don't get one then fail
				t.Fatalf("expected error containing %q, but the command ran without error", tc.expectError)
			}
			// otherwise we expect the command to succeed so pass the test
		})
	}

}

// TestInstallRCommandIntegration tests the install command with the r arg in a Docker container.
func TestInstallRCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	validRVersions, err := languages.RetrieveValidRVersions()
	if err != nil {
		t.Fatalf("failed to retrieve valid R versions: %v", err)
	}

	installCommand := []string{"./wbi", "install", "r", fmt.Sprintf("--version=%s", validRVersions[0])}
	successMessage := []string{fmt.Sprintf("R version %s successfully installed!", validRVersions[0])}

	IntegrationContainerRunner(t, "Dockerfile.Ubuntu", installCommand, successMessage, false)
}

// TestInstallPythonCommandIntegration tests the install command with the python arg in a Docker container.
func TestInstallPythonCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	// Determine OS
	osType, err := operatingsystem.DetectOS()
	if err != nil {
		t.Fatalf("issue detecting OS: %v", err)
	}

	validPythonVersions, err := languages.RetrieveValidPythonVersions(osType)
	if err != nil {
		t.Fatalf("failed to retrieve valid Python versions: %v", err)
	}

	installCommand := []string{"./wbi", "install", "python", fmt.Sprintf("--version=%s", validPythonVersions[0])}
	successMessage := []string{fmt.Sprintf("Python version %s successfully installed!", validPythonVersions[0])}

	IntegrationContainerRunner(t, "Dockerfile.Ubuntu", installCommand, successMessage, false)
}

// TODO TestInstallWorkbenchCommandIntegration tests the install command with the workbench arg in a Docker container.

// TestInstallProDriversCommandIntegration tests the install command with the prodrivers arg in a Docker container.
func TestInstallProDriversCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	installCommand := []string{"./wbi", "install", "prodrivers"}
	successMessage := []string{"The sample preconfigured odbcinst.ini has been appended to /etc/odbcinst.ini"}

	IntegrationContainerRunner(t, "Dockerfile.Ubuntu", installCommand, successMessage, false)
}

// TestInstallJupyterCommandIntegration tests the install command with the jupyter arg in a Docker container.
func TestInstallJupyterCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	installCommand := []string{"./wbi", "install", "jupyter", "--path=/opt/python/3.11.2/bin/python"}
	successMessage := []string{"Jupyter notebook extensions have been successfully installed and enabled!"}

	IntegrationContainerRunner(t, "Dockerfile.Ubuntu-R-Python", installCommand, successMessage, false)
}
