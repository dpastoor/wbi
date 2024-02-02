package languages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/sol-eng/wbi/internal/config"
	"github.com/sol-eng/wbi/internal/install"
	cmdlog "github.com/sol-eng/wbi/internal/logging"
	"github.com/sol-eng/wbi/internal/system"
)

type availablePythonVersions struct {
	PythonVersions []string `json:"python_versions"`
}

type unavailablePythonVersions struct {
	NewestPythonVersions   []string
	OldestPythonVersions   []string
	SpecificPythonVersions []string
}

var globalPythonPaths = []string{
	"/usr/bin/python",
	"/usr/bin/Python",
	"/usr/local/bin/python",
	"/usr/local/bin/Python",
	"/usr/lib/python",
	"/usr/lib/Python",
}
var rootPythonDirs = []string{
	"/opt/python",
	"/opt/Python",
	"/usr/local/lib/python",
	"/usr/local/lib/Python",
}

// GetPythonRootDirs returns the root directories for Python
func GetPythonRootDirs() []string {
	return rootPythonDirs
}

// GetPythonPaths returns the paths workbench will look for Python
// underneath the root directories with the format
// /root/{pythonVersion}/bin/python
func GetPythonPaths() []string {
	return globalPythonPaths
}

func isPythonDir(path string) (string, bool) {
	pythonPath := filepath.Join(path, "bin", "python")
	if _, err := os.Stat(pythonPath); err == nil {
		return pythonPath, true
	}
	return pythonPath, false
}

// CheckPromptAndSetPythonPATH prompts user to set Python PATH
func CheckPromptAndSetPythonPATH(pythonPaths []string) error {
	// check if a wbi_python.sh file exists already and skip asking if it does
	pythonPathSet := CheckIfPythonProfileDExists()
	if !pythonPathSet {
		setPathPythonChoice, err := PythonPATHPrompt()
		if err != nil {
			return fmt.Errorf("issue selecting adding Python to PATH: %w", err)
		}
		if setPathPythonChoice {
			// Remove python/python3 from each path so other binaries will be available in path such as pip, jupyter, etc.
			pythonPathsBin, err := RemovePythonFromPathSlice(pythonPaths)
			if err != nil {
				return fmt.Errorf("issue removing python from slice of locations: %w", err)
			}
			pythonPathChoice, err := PythonLocationPATHPrompt(pythonPathsBin)
			if err != nil {
				return fmt.Errorf("issue selecting Python binary to add to PATH: %w", err)
			}
			err = system.AddToPATH(pythonPathChoice, "python")
			if err != nil {
				return fmt.Errorf("issue adding Python binary to PATH: %w", err)
			}

			system.PrintAndLogInfo("\nThe selected Python directory  " + pythonPathChoice + " has been successfully added to PATH in the /etc/profile.d/wbi_python.sh file.\n")
		}
	}
	return nil
}

// PythonLocationPATHPrompt asks users which Python binary they want to add to PATH
func PythonLocationPATHPrompt(pythonPaths []string) (string, error) {
	// Allow the user to select a version of Python to target
	target := ""
	messageText := `Please select a Python binary to add to PATH.`
	prompt := &survey.Select{
		Message: messageText,
		Options: pythonPaths,
	}
	err := survey.AskOne(prompt, &target)
	if err != nil {
		return "", errors.New("there was an issue with the Python selection prompt for adding to PATH")
	}
	if target == "" {
		return target, errors.New("no Python binary selected to add to PATH")
	}
	log.Info(messageText)
	log.Info(target)
	return target, nil
}

// PromptAndInstallPython Prompts user if they want to install Python and does the installation
func PromptAndInstallPython(osType config.OperatingSystem) ([]string, error) {
	installPythonChoice, err := PythonInstallPrompt()
	if err != nil {
		return []string{}, fmt.Errorf("issue selecting Python installation: %w", err)
	}
	if installPythonChoice {
		validPythonVersions, err := RetrieveValidPythonVersions(osType)
		if err != nil {
			return []string{}, fmt.Errorf("issue retrieving Python versions: %w", err)
		}

		var installPythonVersions []string
		for {
			installPythonVersions, err = PythonSelectVersionsPrompt(validPythonVersions)
			if err != nil {
				return []string{}, fmt.Errorf("issue selecting Python versions: %w", err)
			}
			if len(installPythonVersions) == 0 {
				system.PrintAndLogInfo(`No Python versions selected. Please select at least one version to install.`)
			} else {
				break
			}
		}

		for _, pythonVersion := range installPythonVersions {
			err = DownloadAndInstallPython(pythonVersion, osType)
			if err != nil {
				return []string{}, fmt.Errorf("issue installing Python version: %w", err)
			}
		}
		return installPythonVersions, nil
	}
	return []string{}, nil
}

// ScanAndHandlePythonVersions scans for Python versions, handles result/errors and creates PythonConfig
func ScanAndHandlePythonVersions(osType config.OperatingSystem) error {
	pythonVersionsOrig, err := ScanForPythonVersions()
	if err != nil {
		return fmt.Errorf("issue occured in scanning for Python versions: %w", err)
	}

	system.PrintAndLogInfo("\nFound Python versions:")
	system.PrintAndLogInfo(strings.Join(pythonVersionsOrig, "\n"))

	if len(pythonVersionsOrig) == 0 {
		scanMessage := "no Python versions found at locations: \n" + strings.Join(GetPythonRootDirs(), "\n")
		system.PrintAndLogInfo(scanMessage)

		installedPythonVersion, err := PromptAndInstallPython(osType)
		if err != nil {
			return fmt.Errorf("issue installing Python: %w", err)
		}
		if len(installedPythonVersion) == 0 {
			return errors.New("no Python versions have been installed")
		}
	} else {
		anyOptLocations := []string{}
		for _, value := range pythonVersionsOrig {
			matched, err := regexp.MatchString(".*/opt.*", value)
			if err == nil && matched {
				anyOptLocations = append(anyOptLocations, value)
			}
		}
		if len(anyOptLocations) == 0 {
			system.PrintAndLogInfo("Posit recommends installing version of Python into the /opt directory to not conflict/rely on the system installed version of Python.")
		}
		_, err := PromptAndInstallPython(osType)
		if err != nil {
			return fmt.Errorf("issue installing Python: %w", err)
		}
	}

	pythonVersions, err := ScanForPythonVersions()
	if err != nil {
		return fmt.Errorf("issue occured in scanning for Python versions: %w", err)
	}

	err = CheckPromptAndSetPythonPATH(pythonVersions)
	if err != nil {
		return fmt.Errorf("issue setting Python PATH: %w", err)
	}

	system.PrintAndLogInfo("\nFound Python versions:")
	system.PrintAndLogInfo(strings.Join(pythonVersions, "\n"))
	return nil
}

// ScanForPythonVersions scans for Python versions in locations workbench will also look
func ScanForPythonVersions() ([]string, error) {
	foundVersions := []string{}
	foundOptVersions := []string{}
	// This is somewhat naive with respect to actually checking whether
	// this is _really_ a working version of Python by launching it,
	// vs just matches the path to Python
	for _, pyPath := range GetPythonPaths() {
		if _, err := os.Stat(pyPath); err == nil {
			foundVersions = append(foundVersions, pyPath)
		}

	}
	for _, rootPath := range GetPythonRootDirs() {
		entries, err := os.ReadDir(rootPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return foundVersions, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				pythonPath, isPython := isPythonDir(filepath.Join(rootPath, entry.Name()))
				if isPython {
					if rootPath == "/opt/python" {
						foundOptVersions = append(foundOptVersions, pythonPath)
					} else {
						foundVersions = append(foundVersions, pythonPath)
					}
				}
			}
		}
	}

	maybePython, err := exec.LookPath("python3")
	if err == nil {
		foundVersions = AppendIfMissing(foundVersions, maybePython)
	}

	// sort /opt/R versions
	foundOptVersionsSortedPaths, err := sortOptPythonVersionPaths(foundOptVersions)
	if err != nil {
		return []string{}, fmt.Errorf("issue sorting /opt/python versions: %w", err)
	}

	finalVersionPathSorted := append(foundOptVersionsSortedPaths, foundVersions...)

	return finalVersionPathSorted, nil
}

func sortOptPythonVersionPaths(versionPaths []string) ([]string, error) {
	foundOptVersionsOnly := []string{}
	for _, optVersion := range versionPaths {
		i := strings.Index(optVersion, "python")
		j := strings.Index(optVersion, "bin")
		foundOptVersionsOnly = append(foundOptVersionsOnly, optVersion[i+7:j-1])
	}
	versions, err := ConvertStringSliceToVersionSlice(foundOptVersionsOnly)
	if err != nil {
		return []string{}, fmt.Errorf("issue converting string slice to version slice: %w", err)
	}

	sortedVersions := SortVersionsDesc(versions)
	foundOptVersionsSorted := ConvertVersionSliceToStringSlice(sortedVersions)
	foundOptVersionsSortedPaths := []string{}
	for _, optVersion := range foundOptVersionsSorted {
		foundOptVersionsSortedPaths = append(foundOptVersionsSortedPaths, "/opt/python/"+optVersion+"/bin/python")
	}
	return foundOptVersionsSortedPaths, nil
}

// PythonInstallPrompt Prompt users if they would like to install Python versions
func PythonInstallPrompt() (bool, error) {
	name := true
	messageText := "Would you like to install version(s) of Python?"
	prompt := &survey.Confirm{
		Message: messageText,
	}
	err := survey.AskOne(prompt, &name)
	if err != nil {
		return false, errors.New("there was an issue with the Python install prompt")
	}
	log.Info(messageText)
	log.Info(fmt.Sprintf("%v", name))
	return name, nil
}

// PythonPATHPrompt asks users if they would like to set Python PATH
func PythonPATHPrompt() (bool, error) {
	name := true
	messageText := `Would you like to add a Python version to PATH? This is recommended so users can type "python" and "pip" in the terminal to access this specified version of python and associated tools.`
	prompt := &survey.Confirm{
		Message: messageText,
	}
	err := survey.AskOne(prompt, &name)
	if err != nil {
		return false, errors.New("there was an issue with the Python set PATH prompt")
	}
	log.Info(messageText)
	log.Info(fmt.Sprintf("%v", name))
	return name, nil
}

func RetrieveValidPythonVersions(osType config.OperatingSystem) ([]string, error) {
	pythonVersionURL := "https://cdn.posit.co/python/versions.json"

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, pythonVersionURL, nil)
	if err != nil {
		return []string{}, errors.New("error creating request")
	}
	res, err := client.Do(req)
	if err != nil {
		return []string{}, errors.New("error retrieving JSON data")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return []string{}, errors.New("error in HTTP status code")
	}

	var availVersions availablePythonVersions
	err = json.NewDecoder(res.Body).Decode(&availVersions)
	if err != nil {
		return []string{}, errors.New("error unmarshalling JSON data")
	}

	versions, err := ConvertStringSliceToVersionSlice(availVersions.PythonVersions)
	if err != nil {
		return []string{}, fmt.Errorf("issue converting string slice to version slice: %w", err)
	}

	unavailPythonVersions := unavailablePythonVersionsByOS(osType)

	if len(unavailPythonVersions.NewestPythonVersions) != 0 {
		for _, v := range unavailPythonVersions.NewestPythonVersions {
			versions, err = removeNewerVersions(versions, v)
			if err != nil {
				return []string{}, errors.New("failed removing newer unsupported versions of Python")
			}
		}
	}
	if len(unavailPythonVersions.OldestPythonVersions) != 0 {
		for _, v := range unavailPythonVersions.OldestPythonVersions {
			versions, err = removeOlderVersions(versions, v)
			if err != nil {
				return []string{}, errors.New("failed removing older unsupported versions of Python")
			}
		}
	}
	if len(unavailPythonVersions.SpecificPythonVersions) != 0 {
		for _, v := range unavailPythonVersions.SpecificPythonVersions {
			versions, err = removeSpecificVersions(versions, v)
			if err != nil {
				return []string{}, errors.New("failed removing specific unsupported versions of Python")
			}
		}
	}

	sortedVersions := SortVersionsDesc(versions)
	if err != nil {
		return []string{}, errors.New("failed to sort versions")
	}
	stringVersions := ConvertVersionSliceToStringSlice(sortedVersions)

	return stringVersions, nil
}

// PythonSelectVersionsPrompt Prompt asking users which Python version(s) they would like to install
func PythonSelectVersionsPrompt(availablePythonVersions []string) ([]string, error) {
	messageText := "Which version(s) of Python would you like to install?"
	var qs = []*survey.Question{
		{
			Name: "pythonVersions",
			Prompt: &survey.MultiSelect{
				Message: messageText,
				Options: availablePythonVersions,
				Default: availablePythonVersions[0],
			},
		},
	}
	pythonVersionsAnswers := struct {
		PythonVersions []string `survey:"pythonVersions"`
	}{}
	err := survey.Ask(qs, &pythonVersionsAnswers, survey.WithRemoveSelectAll(), survey.WithRemoveSelectNone())
	if err != nil {
		return []string{}, errors.New("there was an issue with the Python versions selection prompt")
	}
	log.Info(messageText)
	log.Info(strings.Join(pythonVersionsAnswers.PythonVersions, ", "))
	return pythonVersionsAnswers.PythonVersions, nil
}

// DownloadAndInstallPython Downloads the Python installer, and installs Python
func DownloadAndInstallPython(pythonVersion string, osType config.OperatingSystem) error {
	// Create InstallerInfoPython with the proper information
	installerInfo, err := PopulateInstallerInfo("python", pythonVersion, osType)
	if err != nil {
		return fmt.Errorf("PopulateInstallerInfoPython: %w", err)
	}
	// Download installer
	installerPath, err := install.DownloadFile("Python", installerInfo.URL, installerInfo.Name)
	if err != nil {
		return fmt.Errorf("DownloadPython: %w", err)
	}
	// Install Python
	err = install.InstallLanguage("python", installerPath, osType, pythonVersion)
	if err != nil {
		return fmt.Errorf("InstallLanguage: %w", err)
	}
	// save to command log
	installCommand, err := install.RetrieveInstallCommand(installerInfo.Name, osType)
	if err != nil {
		return fmt.Errorf("RetrieveInstallCommand: %w", err)
	}
	cmdlog.Info("curl -O " + installerInfo.URL)
	cmdlog.Info(installCommand)
	// Upgrade pip, setuptools, and wheel
	err = UpgradePythonTools(pythonVersion)
	if err != nil {
		return fmt.Errorf("UpgradePythonTools: %w", err)
	}

	return nil
}

func UpgradePythonTools(pythonVersion string) error {
	upgradeCommand := "PIP_ROOT_USER_ACTION=ignore /opt/python/" + pythonVersion + "/bin/pip install --upgrade --no-warn-script-location --disable-pip-version-check pip setuptools wheel"
	err := system.RunCommand(upgradeCommand, true, 2, true)
	if err != nil {
		return fmt.Errorf("issue upgrading pip, setuptools and wheel for Python with the command '%s': %w", upgradeCommand, err)
	}

	successMessage := "\npip, setuptools and wheel have been upgraded for Python version " + pythonVersion + "\n"
	system.PrintAndLogInfo(successMessage)

	return nil
}

// RemovePythonFromPath removes python or python3 from the end of a path so the directory can be used
func RemovePythonFromPath(pythonPath string) (string, error) {
	if _, err := regexp.MatchString(".*/python.*", pythonPath); err == nil {
		i := strings.LastIndex(pythonPath, "/python")
		excludingLast := pythonPath[:i] + strings.Replace(pythonPath[i:], "/python", "", 1)
		return excludingLast, nil
	} else if _, err := regexp.MatchString(".*/python3.*", pythonPath); err == nil {
		i := strings.LastIndex(pythonPath, "/python3")
		excludingLast := pythonPath[:i] + strings.Replace(pythonPath[i:], "/python3", "", 1)
		return excludingLast, nil
	} else {
		return pythonPath, nil
	}
}

// RemovePythonFromPathSlice removes python or python3 from the end of a set of path strings in a slice so the directories can be used
func RemovePythonFromPathSlice(pythonPaths []string) ([]string, error) {
	var newPythonPaths []string
	for _, pythonPath := range pythonPaths {
		newPythonPath, err := RemovePythonFromPath(pythonPath)
		if err != nil {
			return []string{}, err
		}
		newPythonPaths = append(newPythonPaths, newPythonPath)
	}
	return newPythonPaths, nil
}

func ValidatePythonVersions(pythonVersions []string, osType config.OperatingSystem) error {
	availablePythonVersions, err := RetrieveValidPythonVersions(osType)
	if err != nil {
		return fmt.Errorf("error retrieving valid R versions: %w", err)
	}
	for _, pythonVersion := range pythonVersions {
		if !lo.Contains(availablePythonVersions, pythonVersion) {
			return errors.New("version " + pythonVersion + " is not a valid Python version")
		}
	}
	return nil
}

func CheckIfPythonProfileDExists() bool {
	_, err := os.Stat("/etc/profile.d/wbi_python.sh")
	if err != nil {
		return false
	}

	system.PrintAndLogInfo("\nAn existing /etc/profile.d/wbi_python.sh file was found, skipping setting Python path.")
	return true
}
func unavailablePythonVersionsByOS(osType config.OperatingSystem) unavailablePythonVersions {

	var pythonVersions unavailablePythonVersions
	switch osType {
	case config.Redhat7:
		pythonVersions.NewestPythonVersions = []string{"3.10.0", "3.11.0", "3.8.16", "3.9.15"}
	case config.Redhat9:
		pythonVersions.SpecificPythonVersions = []string{"3.7.3", "3.7.4", "3.7.5"}
	}
	// global
	pythonVersions.NewestPythonVersions = []string{"3.12.0"}

	return pythonVersions
}
