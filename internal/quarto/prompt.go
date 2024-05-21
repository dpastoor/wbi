package quarto

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/sol-eng/wbi/internal/config"
)

func ScanAndHandleQuartoVersions(osType config.OperatingSystem) error {
	// check if a Workbench bundled version of Quarto exists
	quartoBundled, err := checkForBundledQuartoVersion()
	if err != nil {
		return fmt.Errorf("issue checking for bundled Quarto version: %w", err)
	}
	var quartoBundledVersion string
	if quartoBundled {
		// detect the bundled Quarto version
		quartoBundledVersion, err = ScanForBundledQuartoVersion()
		if err != nil {
			return fmt.Errorf("issue scanning for bundled Quarto version: %w", err)
		}
	}

	// prompt the user to present the bundled version and ask if they want to install any other versions. If nothing is bundled then just ask if they want to install any versions
	quartoInstall, err := PromptQuartoInstall(quartoBundledVersion)
	if err != nil {
		return fmt.Errorf("there was an issue prompting for Quarto install: %w", err)
	}

	if quartoInstall {
		// retrieve other versions and present them to the user
		fmt.Printf("Retrieving available Quarto versions...")
		validQuartoVersions, err := RetrieveValidQuartoVersions()
		if err != nil {
			return fmt.Errorf("error retrieving valid Quarto versions: %w", err)
		}

		if err != nil {
			return fmt.Errorf("there was an issue retrieving valid Quarto versions: %w", err)
		}
		installQuartoVersions, err := QuartoSelectVersionsPrompt(validQuartoVersions)
		if err != nil {
			return fmt.Errorf("issue selecting Quarto versions: %w", err)
		}
		if len(installQuartoVersions) > 0 {
			// install the version(s)
			err = DownloadAndInstallQuartoVersions(installQuartoVersions, osType)
			if err != nil {
				return fmt.Errorf("there was an issue installing Quarto versions: %w", err)
			}
			// ask which version they want to use for default and symlink it to /usr/local/bin/quarto so Jupyter and VS Code will pick it up
			var quartoPaths []string
			if quartoBundled {
				quartoPaths = append(quartoVersionsToPaths(installQuartoVersions), "/usr/lib/rstudio-server/bin/quarto/bin/quarto")
			} else {
				quartoPaths = quartoVersionsToPaths(installQuartoVersions)
			}

			err = checkPromtAndSetQuartoSymlinks(quartoPaths)
			if err != nil {
				return fmt.Errorf("there was an issue setting Quarto symlinks: %w", err)
			}
		} else {
			// continue with the bundled version and symlink it to /usr/local/bin/quarto so Jupyter and VS Code will pick it up
			err = setQuartoSymlinks("/usr/lib/rstudio-server/bin/quarto/bin/quarto", true)
			if err != nil {
				return fmt.Errorf("issue setting Quarto symlinks: %w", err)
			}
		}
	} else {
		// continue with the bundled version and symlink it to /usr/local/bin/quarto so Jupyter and VS Code will pick it up
		err = setQuartoSymlinks("/usr/lib/rstudio-server/bin/quarto/bin/quarto", false)
		if err != nil {
			return fmt.Errorf("issue setting Quarto symlinks: %w", err)
		}
	}

	return nil
}

// quartoVersionsToPaths converts Quarto versions to full paths in /opt
func quartoVersionsToPaths(quartoVersions []string) []string {
	quartoPaths := []string{}
	for _, version := range quartoVersions {
		quartoPaths = append(quartoPaths, "/opt/quarto/"+version+"/bin/quarto")
	}
	return quartoPaths
}

func PromptQuartoInstall(bundledVersion string) (bool, error) {
	var name bool
	var messageText string
	if bundledVersion == "" {
		messageText = "Would you like to install Quarto?"
	} else {
		messageText = "Workbench bundles Quarto version " + bundledVersion + " Would you like to install any different version(s)?"
	}

	prompt := &survey.Confirm{
		Message: messageText,
	}
	err := survey.AskOne(prompt, &name)
	if err != nil {
		return false, errors.New("there was an issue with Quarto install prompt question")
	}
	log.Info(messageText)
	log.Info(fmt.Sprintf("%v", name))
	return name, nil
}

// QuartoSelectVersionsPrompt Prompt asking users which Quarto version(s) they would like to install
func QuartoSelectVersionsPrompt(availableQuartoVersions []string) ([]string, error) {
	messageText := "Which version(s) of Quarto would you like to install?"
	var qs = []*survey.Question{
		{
			Name: "quartoversions",
			Prompt: &survey.MultiSelect{
				Message: messageText,
				Options: availableQuartoVersions,
				Default: availableQuartoVersions[0],
			},
		},
	}
	quartoVersionsAnswers := struct {
		QuartoVersions []string `survey:"quartoversions"`
	}{}
	err := survey.Ask(qs, &quartoVersionsAnswers, survey.WithRemoveSelectAll(), survey.WithRemoveSelectNone())
	if err != nil {
		return []string{}, errors.New("there was an issue with the Quarto versions selection prompt")
	}
	log.Info(messageText)
	log.Info(strings.Join(quartoVersionsAnswers.QuartoVersions, ", "))
	return quartoVersionsAnswers.QuartoVersions, nil
}
