package quarto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/samber/lo"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sol-eng/wbi/internal/config"
	cmdlog "github.com/sol-eng/wbi/internal/logging"
	"github.com/sol-eng/wbi/internal/system"
)

type Assets []struct {
	BrowserDownloadURL string `json:"browser_download_url"`
}
type Quarto []struct {
	Assets     Assets `json:"assets"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
}

type result struct {
	index int
	res   http.Response
	err   error
}

func RetrieveValidQuartoVersions() ([]string, error) {
	var availQuartoVersions []string
	var results []result
	var quarto Quarto
	var urls []string

	for pagenum := 1; pagenum < 5; pagenum++ {
		urls = append(urls, "https://api.github.com/repos/quarto-dev/quarto-cli/releases?per_page=100&page="+strconv.Itoa(pagenum))
	}
	wg := sync.WaitGroup{}

	for i, url := range urls {
		wg.Add(1)

		go func(i int, url string) {

			res, err := http.Get(url)
			if err != nil {
				return
			}
			result := &result{i, *res, err}
			results = append(results, *result)
			wg.Done()

		}(i, url)

	}

	wg.Wait()

	// let's sort these results real quick
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})

	for _, result := range results {

		err := json.NewDecoder(result.res.Body).Decode(&quarto)
		if err != nil {
			return nil, err
		}
		for _, release := range quarto {
			if release.Prerelease == false {
				availQuartoVersions = append(availQuartoVersions, release.Name)
			}
		}
	}
	return availQuartoVersions, nil
}

func ValidateQuartoVersions(quartoVersions []string) error {

	availQuartoVersions, err := RetrieveValidQuartoVersions()
	if err != nil {
		return fmt.Errorf("error retrieving valid Quarto versions: %w", err)
	}

	for _, quartoVersion := range quartoVersions {
		if !lo.Contains(availQuartoVersions, quartoVersion) {
			return errors.New("version " + quartoVersion + " is not a valid Quarto version")
		}
	}
	return nil
}

func DownloadAndInstallQuartoVersions(quartoVersions []string, osType config.OperatingSystem) error {
	for _, quartoVersion := range quartoVersions {
		err := DownloadAndInstallQuarto(quartoVersion, osType)
		if err != nil {
			return fmt.Errorf("issue installing Quarto version: %w", err)
		}
	}
	return nil
}

func DownloadAndInstallQuarto(quartoVersion string, osType config.OperatingSystem) error {
	// Find URL
	quartoURL := generateQuartoInstallURL(quartoVersion, osType)
	// Download installer
	installerPath, err := downloadFileQuarto(quartoURL, quartoVersion)
	if err != nil {
		return fmt.Errorf("DownloadFileQuarto: %w", err)
	}
	// Install Quarto
	err = installQuarto(installerPath, osType, quartoVersion, true)

	if err != nil {
		return fmt.Errorf("InstallQuarto: %w", err)
	}
	// save to command log
	quartoPath := fmt.Sprintf("/opt/quarto/%s", quartoVersion)
	cmdlog.Info("curl -o quarto.tar.gz -L " + quartoURL)
	cmdlog.Info("mkdir -p " + quartoPath)
	cmdlog.Info(fmt.Sprintf(`tar -zxvf quarto.tar.gz -C "%s" --strip-components=1`, quartoPath))
	cmdlog.Info("rm quarto.tar.gz")
	return nil
}

func generateQuartoInstallURL(quartoVersion string, osType config.OperatingSystem) string {
	// treat RHEL 7 differently as specified here: https://docs.posit.co/resources/install-quarto/#specify-quarto-version-tar
	var url string
	if osType == config.Redhat7 {
		url = fmt.Sprintf("https://github.com/quarto-dev/quarto-cli/releases/download/%s/quarto-%s-linux-rhel7-amd64.tar.gz", quartoVersion, strings.Replace(quartoVersion, "v", "", -1))
	} else {
		url = fmt.Sprintf("https://github.com/quarto-dev/quarto-cli/releases/download/%s/quarto-%s-linux-amd64.tar.gz", quartoVersion, strings.Replace(quartoVersion, "v", "", -1))
	}
	return url
}

func downloadFileQuarto(url string, version string) (string, error) {
	system.PrintAndLogInfo("Downloading Quarto Version: " + version + " installer from: " + url)

	// Create the file
	filename := "*_" + fmt.Sprintf("quarto-%s-linux-amd64.tar.gz", version)
	tmpFile, err := os.CreateTemp("", filename)
	if err != nil {
		return tmpFile.Name(), err
	}

	defer func() {
		if tempErr := tmpFile.Close(); tempErr != nil {
			err = tempErr
		}
	}()

	if err != nil {
		return tmpFile.Name(), err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, url, nil)
	if err != nil {
		return "", errors.New("error creating request")
	}
	res, err := client.Do(req)
	if err != nil {
		return "", errors.New("error downloading " + filename + " installer")
	}
	defer func() {
		if tempErr := res.Body.Close(); tempErr != nil {
			err = tempErr
		}
	}()

	if res.StatusCode != http.StatusOK {
		return "", errors.New("error retrieving " + filename + " installer")
	}

	// Writer the body to file
	_, err = io.Copy(tmpFile, res.Body)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// Installs Quarto
func installQuarto(filepath string, osType config.OperatingSystem, version string, save bool) error {

	// create the /opt/quarto directory if it doesn't exist
	path := fmt.Sprintf("/opt/quarto/%s", version)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	installCommand := fmt.Sprintf(`tar -zxvf "%s" -C "%s" --strip-components=1`, filepath, path)

	err := system.RunCommand(installCommand, false, 0, false)
	if err != nil {
		return fmt.Errorf("the command '%s' failed to run: %w", installCommand, err)
	}

	successMessage := "\nQuarto version " + version + " successfully installed!\n"
	system.PrintAndLogInfo(successMessage)
	return nil
}

// promptAndSetQuartoSymlinks prompts user to set the Quarto symlink
func promptAndSetQuartoSymlink(quartoPaths []string) error {
	setQuartoSymlinkChoice, err := quartoSymlinkPrompt()
	if err != nil {
		return fmt.Errorf("an issue occured during the selection of Quarto symlink choice: %w", err)
	}
	if setQuartoSymlinkChoice {
		quartoPathChoice, err := quartoLocationSymlinksPrompt(quartoPaths)
		if err != nil {
			return fmt.Errorf("issue selecting Quarto binary to add symlinks: %w", err)
		}
		err = setQuartoSymlinks(quartoPathChoice, true)
		if err != nil {
			return fmt.Errorf("issue setting Quarto symlinks: %w", err)
		}

		system.PrintAndLogInfo("\n " + quartoPathChoice + " has been successfully symlinked and will be available on the default system PATH.\n")
	}
	return nil
}

// setQuartoSymlinks sets the Quarto symlink
func setQuartoSymlinks(quartoPath string, display bool) error {
	quartoCommand := "ln -s " + quartoPath + " /usr/local/bin/quarto"
	err := system.RunCommand(quartoCommand, display, 0, true)
	if err != nil {
		return fmt.Errorf("error setting Quarto symlink with the command '%s': %w", quartoCommand, err)
	}
	return nil
}

// quartoLocationSymlinksPrompt asks users which Quarto binary they want to symlink
func quartoLocationSymlinksPrompt(quartoPaths []string) (string, error) {
	// Allow the user to select a version of Quarto to target
	target := ""
	messageText := "Select a Quarto binary to symlink:"
	prompt := &survey.Select{
		Message: messageText,
		Options: quartoPaths,
	}
	err := survey.AskOne(prompt, &target)
	if err != nil {
		return "", errors.New("there was an issue with the Quarto selection prompt for symlink")
	}
	if target == "" {
		return target, errors.New("no Quarto binary selected to be symlinked")
	}
	log.Info(messageText)
	log.Info(target)
	return target, nil
}

// quartoSymlinkPrompt asks users if they would like to set the quarto symlink
func quartoSymlinkPrompt() (bool, error) {
	name := true
	messageText := `Would you like to symlink a Quarto version to make it available on PATH? This is recommended so Workbench can default to this version of Quarto in each of the IDEs and users can type "quarto" in the terminal.`
	prompt := &survey.Confirm{
		Message: messageText,
	}
	err := survey.AskOne(prompt, &name)
	if err != nil {
		return false, errors.New("there was an issue with the symlink Quarto prompt")
	}
	log.Info(messageText)
	log.Info(fmt.Sprintf("%v", name))
	return name, nil
}
