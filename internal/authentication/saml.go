package authentication

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
)

// Prompt asking users to provide a username attribute for SAML
func PromptSAMLAttribute() (string, error) {
	name := "Username"
	prompt := &survey.Input{
		Message: "SAML IdP username attribute:",
		Default: "Username",
	}
	err := survey.AskOne(prompt, &name)
	if err != nil {
		return "", errors.New("there was an issue with the SAML username attribute prompt")
	}
	return name, nil
}

// Prompt asking users to provide a metadata URL for SAML
func PromptSAMLMetadataURL() (string, error) {
	name := ""
	prompt := &survey.Input{
		Message: "SAML IdP metadata URL:",
	}
	err := survey.AskOne(prompt, &name)
	if err != nil {
		return "", errors.New("there was an issue with the SAML IdP URL prompt")
	}
	return name, nil
}
