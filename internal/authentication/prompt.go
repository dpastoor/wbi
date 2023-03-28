package authentication

import (
	"errors"
	"fmt"
	"log"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sol-eng/wbi/internal/config"
	"github.com/sol-eng/wbi/internal/workbench"
)

// Prompt asking users if they wish to setup Authentication
func PromptAuth() (bool, error) {
	name := false
	prompt := &survey.Confirm{
		Message: "Would you like to setup Authentication?",
	}
	err := survey.AskOne(prompt, &name)
	if err != nil {
		return false, errors.New("there was an issue with the Authentication prompt")
	}
	return name, nil
}

func PromptAndConfigAuth(osType config.OperatingSystem) error {
	authType, err := PromptAndConvertAuthType()
	if err != nil {
		return fmt.Errorf("issue entering and converting AuthType: %w", err)
	}
	err = HandleAuthChoice(authType, osType)
	if err != nil {
		return fmt.Errorf("issue handling authentication: %w", err)
	}
	return nil
}

func PromptAndConvertAuthType() (config.AuthType, error) {

	authChoiceRaw, err := PromptAuthentication()
	if err != nil {
		return config.Other, fmt.Errorf("PromptAuthentication: %w", err)
	}
	authChoice, err := ConvertAuthType(authChoiceRaw)
	if err != nil {
		return config.Other, fmt.Errorf("PromptAuthentication: %w", err)
	}
	return authChoice, nil
}

// Prompt asking user for their desired authentication method
func PromptAuthentication() (string, error) {
	name := ""
	prompt := &survey.Select{
		Message: "Choose an authentication method:",
		Options: []string{"SAML", "OIDC", "Active Directory/LDAP", "PAM", "Other"},
	}
	err := survey.AskOne(prompt, &name)
	if err != nil {
		return "", errors.New("there was an issue with the authentication prompt")
	}
	return name, nil
}

// Convert authChoice from a string to a proper AuthType type
func ConvertAuthType(authChoice string) (config.AuthType, error) {
	switch authChoice {
	case "SAML":
		return config.SAML, nil
	case "OIDC":
		return config.OIDC, nil
	case "Active Directory/LDAP":
		return config.LDAP, nil
	case "PAM":
		return config.PAM, nil
	case "Other":
		return config.Other, nil
	}
	return config.Other, errors.New("unknown AuthType was selected")
}

// Route an authentication choice to the proper prompts/information
func HandleAuthChoice(authType config.AuthType, targetOS config.OperatingSystem) error {
	switch authType {
	case config.SAML:
		idpURL, err := PromptSAMLMetadataURL()
		if err != nil {
			return fmt.Errorf("issue entering SAML metadata URL: %w", err)
		}
		err = workbench.WriteSAMLAuthConfig(idpURL)
		if err != nil {
			return fmt.Errorf("failed to write SAML auth config: %w", err)
		}

		fmt.Println("Setting up SAML based authentication is a 2 step process. Step 1 was just completed by wbi writing the configuration file, however your SAML setup may require further configuration. \n\nTo complete step 2, you must configure your identify provider with Workbench following steps outlined here: https://docs.posit.co/ide/server-pro/authenticating_users/saml_sso.html#step-2.-configure-your-identity-provider-with-workbench")
	case config.OIDC:
		fmt.Println("Setting up OpenID Connect based authentication is a 2 step process. First configure your OpenID provider with the steps outlined here to complete step 1: https://docs.posit.co/ide/server-pro/authenticating_users/openid_connect_authentication.html#configuring-your-openid-provider \n\n As you register Workbench in the IdP, save the client-id and client-secret. Follow the next prompts to complete step 2.")

		clientID, err := PromptOIDCClientID()
		if err != nil {
			return fmt.Errorf("issue entering OIDC client ID: %w", err)
		}
		clientSecret, err := PromptOIDCClientSecret()
		if err != nil {
			return fmt.Errorf("issue entering OIDC client secret: %w", err)
		}
		idpURL, err := PromptOIDCIssuerURL()
		if err != nil {
			return fmt.Errorf("issue entering OIDC issuer URL: %w", err)
		}
		err = workbench.WriteOIDCAuthConfig(idpURL, "", clientID, clientSecret)
		if err != nil {
			return fmt.Errorf("failed to write OIDC auth config: %w", err)
		}
	case config.LDAP:
		switch targetOS {
		case config.Ubuntu18, config.Ubuntu20, config.Ubuntu22:
			fmt.Println("Posit Workbench connects to LDAP via PAM. Please follow this article for more details on configuration: \nhttps://support.posit.co/hc/en-us/articles/360024137174-Integrating-Ubuntu-with-Active-Directory-for-RStudio-Workbench-RStudio-Server-Pro")
		case config.Redhat7, config.Redhat8, config.Redhat9:
			fmt.Println("Posit Workbench connects to LDAP via PAM. Please follow this article for more details on configuration: \nhttps://support.posit.co/hc/en-us/articles/360016587973-Integrating-RStudio-Workbench-RStudio-Server-Pro-with-Active-Directory-using-CentOS-RHEL")
		default:
			log.Fatal("Unsupported operating system")
		}
	case config.PAM:
		fmt.Println("PAM requires no additional configuration, however there are some username considerations and home directory provisioning steps that can be taken. To learn more please visit: https://docs.posit.co/ide/server-pro/authenticating_users/pam_authentication.html")
	case config.Other:
		fmt.Println("To learn about configuring your desired method of authentication please visit: https://docs.posit.co/ide/server-pro/authenticating_users/authenticating_users.html")
	}
	return nil
}
