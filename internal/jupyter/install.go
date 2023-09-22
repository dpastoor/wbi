package jupyter

import (
	"fmt"
	"strings"

	"github.com/sol-eng/wbi/internal/languages"
	"github.com/sol-eng/wbi/internal/system"
	"github.com/sol-eng/wbi/internal/workbench"
)

// InstallJupyter installs jupypter pip stuff
func InstallJupyter(pythonPath string) error {
	JupyterComponentsErr := InstallJupyterAndComponents(pythonPath)
	if JupyterComponentsErr != nil {
		return fmt.Errorf("InstallJupyterAndComponents: %w", JupyterComponentsErr)
	}
	JupyterNotebookExtensionsErr := InstallAndEnableJupyterNotebookExtensions(pythonPath)
	if JupyterNotebookExtensionsErr != nil {
		return fmt.Errorf("InstallAndEnableJupyterNotebookExtensions: %w", JupyterNotebookExtensionsErr)
	}
	return nil
}

// Install various Jupyter related packages from PyPI
func InstallJupyterAndComponents(pythonPath string) error {
	licenseCommand := "PIP_ROOT_USER_ACTION=ignore " + pythonPath + " -m pip install --no-warn-script-location --disable-pip-version-check jupyter notebook==6.5.6 jupyterlab==3.6.5 rsp_jupyter rsconnect_jupyter workbench_jupyterlab==1.1.1"
	err := system.RunCommand(licenseCommand, true, 2, true)
	if err != nil {
		return fmt.Errorf("issue installing Jupyter with the command '%s': %w", licenseCommand, err)
	}
	// TODO add some proper tests to ensure Jupyter is working
	system.PrintAndLogInfo("\nJupyter has been successfully installed!")
	return nil
}

// Install and enable various Jupyter notebook extensions
func InstallAndEnableJupyterNotebookExtensions(pythonPath string) error {

	pythonPathShort, err := languages.RemovePythonFromPath(pythonPath)
	if err != nil {
		return fmt.Errorf("issue shortening Python path: %w", err)
	}

	commands := []string{
		"jupyter-nbextension install --sys-prefix --py rsp_jupyter",
		"jupyter-nbextension enable --sys-prefix --py rsp_jupyter",
		"jupyter-nbextension install --sys-prefix --py rsconnect_jupyter",
		"jupyter-nbextension enable --sys-prefix --py rsconnect_jupyter",
		"jupyter-serverextension enable --sys-prefix --py rsconnect_jupyter",
	}

	for _, command := range commands {
		installCommand := pythonPathShort + "/" + command
		err := system.RunCommand(installCommand, true, 0, true)
		if err != nil {
			return fmt.Errorf("issue installing Jupyter notebook extension with the command '%s': %w", installCommand, err)
		}
	}

	// TODO add some proper tests to ensure Jupyter notebook extensions are working
	system.PrintAndLogInfo("\nJupyter notebook extensions have been successfully installed and enabled!")
	return nil
}

func InstallAndConfigJupyter(pythonPath string) error {
	err := InstallJupyter(pythonPath)
	if err != nil {
		return fmt.Errorf("issue installing Jupyter: %w", err)
	}
	// the path to jupyter must be set in the config, not python
	pythonSubPath, err := languages.RemovePythonFromPath(pythonPath)
	if err != nil {
		return fmt.Errorf("issue removing Python from path: %w", err)
	}
	jupyterPath := pythonSubPath + "/jupyter"
	err = workbench.WriteJupyterConfig(jupyterPath)
	if err != nil {
		return fmt.Errorf("issue writing Jupyter config: %w", err)
	}
	return nil
}

func RegisterJupyterKernels(additionalPythonPaths []string) error {

	// register the kernel for each additional python version
	for _, pythonPath := range additionalPythonPaths {
		// find the version
		versionCommand := pythonPath + " --version"
		pythonVersion, err := system.RunCommandAndCaptureOutput(versionCommand, false, 0, false)
		if err != nil {
			return fmt.Errorf("issue finding python version: %w", err)
		}
		// install ipykernel
		err = installIpykernel(pythonPath)
		if err != nil {
			return fmt.Errorf("issue installing ipykernel: %w", err)
		}
		// run the install command
		err = registerKernel(pythonPath, pythonVersion)
		if err != nil {
			return fmt.Errorf("issue registering kernel: %w", err)
		}
	}
	return nil
}

func installIpykernel(pythonPath string) error {
	basePath, err := languages.RemovePythonFromPath(pythonPath)
	if err != nil {
		return fmt.Errorf("issue removing python from the path: %w", err)
	}

	installCommand := "PIP_ROOT_USER_ACTION=ignore " + basePath + "/pip install --no-warn-script-location --disable-pip-version-check ipykernel"
	err = system.RunCommand(installCommand, true, 1, true)
	if err != nil {
		return fmt.Errorf("issue installing ipykernel with the command '%s': %w", installCommand, err)
	}
	return nil
}

func registerKernel(pythonPath string, pythonVersion string) error {
	pythonVersionClean := strings.Replace(pythonVersion, "Python ", "", 1)
	pythonVersionNoBreak := strings.Replace(pythonVersion, "\n", "", -1)

	installCommand := pythonPath + " -m ipykernel install --name py" + strings.TrimSpace(pythonVersionClean) + " --display-name" + " \"" + pythonVersionNoBreak + "\""

	err := system.RunCommand(installCommand, true, 0, true)
	if err != nil {
		return fmt.Errorf("issue registering the Python kernel with the command '%s': %w", installCommand, err)
	}
	return nil
}
