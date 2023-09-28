package cmd

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/sol-eng/wbi/internal/config"
	"github.com/sol-eng/wbi/internal/connect"
	"github.com/sol-eng/wbi/internal/jupyter"
	"github.com/sol-eng/wbi/internal/languages"
	"github.com/sol-eng/wbi/internal/license"
	"github.com/sol-eng/wbi/internal/operatingsystem"
	"github.com/sol-eng/wbi/internal/packagemanager"
	"github.com/sol-eng/wbi/internal/prodrivers"
	"github.com/sol-eng/wbi/internal/quarto"
	"github.com/sol-eng/wbi/internal/ssl"
	"github.com/sol-eng/wbi/internal/system"
	"github.com/sol-eng/wbi/internal/workbench"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type setupCmd struct {
	cmd  *cobra.Command
	opts setupOpts
}

type setupOpts struct {
	step string
}

func newSetup(setupOpts setupOpts) error {

	// define step either "" if no flag or a step if flag is set
	step := setupOpts.step
	if step == "" {
		step = "start"
	}

	if step == "start" {
		system.PrintAndLogInfo("Welcome to the Workbench Installer!")
		step = "prereqs"
	}

	// Check if running as root
	err := operatingsystem.CheckIfRunningAsRoot()
	if err != nil {
		return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step start\"", err)
	}

	// Determine OS and install pre-requisites
	osType, err := operatingsystem.DetectOS()
	if err != nil {
		return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step start\"", err)
	}

	if step == "prereqs" {
		ConfirmInstall, err := operatingsystem.PromptInstallPrereqs()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step prereqs\"", err)
		}

		if ConfirmInstall {
			err = operatingsystem.InstallPrereqs(osType)
		} else if !ConfirmInstall {
			log.Fatal("Exited Workbench Installer")
		}
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step prereqs\"", err)
		}
		step = "firewall"
	}

	if step == "firewall" {
		// Determine if we should disable the local firewall, then disable it
		// TODO: Add support for Ubuntu ufw
		firewalldEnabled, err := operatingsystem.CheckFirewallStatus(osType)
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step firewall\"", err)
		}

		if firewalldEnabled {
			disableFirewall, err := operatingsystem.FirewallPrompt()
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step firewall\"", err)
			}

			if disableFirewall {
				err = operatingsystem.DisableFirewall(osType)
				if err != nil {
					return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step firewall\"", err)
				}
			}
		}
		step = "security"
	}

	if step == "security" {
		// Determine Linux security status for the OS, then disable it
		// TODO: Add support for Ubuntu AppArmor
		selinuxEnabled, err := operatingsystem.CheckLinuxSecurityStatus(osType)
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step security\"", err)
		}

		if selinuxEnabled {
			disableSELinux, err := operatingsystem.LinuxSecurityPrompt(osType)
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step security\"", err)
			}

			if disableSELinux {
				err = operatingsystem.DisableLinuxSecurity()
				if err != nil {
					return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step security\"", err)
				}
			}
		}
		step = "languages"
	}

	selectedLanguages := []string{"r", "python"}
	if step == "languages" {
		// Languages
		selectedLanguages, err = languages.PromptAndRespond()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step languages\"", err)
		}
		step = "r"
	}

	if step == "r" {
		// R
		err = languages.ScanAndHandleRVersions(osType)
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step r\"", err)
		}
		step = "python"
	}

	if step == "python" {
		// Python
		if lo.Contains(selectedLanguages, "python") {
			err := languages.ScanAndHandlePythonVersions(osType)
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step python\"", err)
			}
		}
		step = "workbench"
	}

	if step == "workbench" {
		// Workbench
		err = workbench.CheckPromptDownloadAndInstallWorkbench(osType)
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step workbench\"", err)
		}
		step = "license"
	}

	if step == "license" {
		// Licensing
		err = license.CheckPromptAndActivateLicense()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step license\"", err)
		}
		step = "quarto"
	}

	if step == "quarto" {
		// Quarto
		err := quarto.ScanAndHandleQuartoVersions(osType)
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step quarto\"", err)
		}
		step = "jupyter"
	}

	if step == "jupyter" {
		// Jupyter
		err = jupyter.ScanPromptInstallAndConfigJupyter()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step jupyter\"", err)
		}
		step = "prodrivers"
	}

	if step == "prodrivers" {
		// Pro Drivers
		err = prodrivers.CheckPromptDownloadAndInstallProDrivers(osType)
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step prodrivers\"", err)
		}
		step = "ssl"
	}

	if step == "ssl" {
		// SSL
		sslChoice, err := ssl.PromptSSL()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step ssl\"", err)
		}
		if sslChoice {
			serverURL, err := ssl.PromptServerURL()
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step ssl\"", err)
			}
			certPath, keyPath, err := ssl.PromptAndVerifySSL(osType)
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step ssl\"", err)
			}
			workbench.WriteSSLConfig(certPath, keyPath, serverURL)
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step ssl\"", err)
			}
		}
		step = "packagemanager"
	}

	if step == "packagemanager" {
		// Package Manager URL
		packageManagerChoice, err := packagemanager.PromptPackageManagerChoice()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step packagemanager\"", err)
		}
		if packageManagerChoice == "Posit Package Manager" {
			err = packagemanager.InteractivePackageManagerPrompts(osType)
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step packagemanager\"", err)
			}
		} else if packageManagerChoice == "Posit Public Package Manager" {
			err = packagemanager.VerifyAndBuildPublicPackageManager(osType)
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step packagemanager\"", err)
			}
		}
		step = "connect"
	}

	if step == "connect" {
		// Connect URL
		connectChoice, err := connect.PromptConnectChoice()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step connect\"", err)
		}
		if connectChoice {
			err = connect.PromptVerifyAndConfigConnect()
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step connect\"", err)
			}
		}
		step = "restart"
	}

	if step == "restart" {
		system.PrintAndLogInfo("\nRestarting RStudio Server and Launcher...")

		err = workbench.RestartRStudioServerAndLauncher()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step restart\"", err)
		}
		step = "status"
	}

	if step == "status" {
		system.PrintAndLogInfo("\nPrinting the status of RStudio Server and Launcher...")

		err = workbench.StatusRStudioServerAndLauncher()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step status\"", err)
		}
		step = "verify"
	}

	if step == "verify" {
		verifyChoice, err := workbench.PromptInstallVerify()
		if err != nil {
			return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step verify\"", err)
		}
		if verifyChoice {
			username, skip, err := operatingsystem.PromptAndVerifyUser()
			if err != nil {
				return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step verify\"", err)
			}
			if !skip {
				err = workbench.VerifyInstallation(username)
				if err != nil {
					return fmt.Errorf("%w.\nTo return to this step in the setup process use \"wbi setup --step verify\"", err)
				}
			}
		}
		step = "done"
	}

	var adDocURL string
	switch osType {
	case config.Ubuntu20, config.Ubuntu22:
		adDocURL = "https://support.posit.co/hc/en-us/articles/360024137174-Integrating-Ubuntu-with-Active-Directory-for-RStudio-Workbench-RStudio-Server-Pro"
	case config.Redhat7, config.Redhat8, config.Redhat9:
		adDocURL = "https://support.posit.co/hc/en-us/articles/360016587973-Integrating-RStudio-Workbench-RStudio-Server-Pro-with-Active-Directory-using-CentOS-RHEL"
	}

	var sslEnabled bool
	matched, err := system.CheckStringExists("ssl-enabled=1", "/etc/rstudio/rserver.conf")
	if err == nil && matched {
		sslEnabled = true
	}
	var serverAccessMessage string
	if sslEnabled {
		serverAccessMessage = "To access Workbench in a web browser navigate to https://YOUR_SERVER_URL.com, replacing YOUR_SERVER_URL.com with the actual URL of this server. \n\n"
	} else {
		serverAccessMessage = "To access Workbench in a web browser navigate to http://YOUR_SERVER_URL.com:8787 replacing YOUR_SERVER_URL.com with the actual URL of this server. By default Workbench runs on port 8787 when using HTTP, visit the Admin Guide for more information on how to change this: https://docs.posit.co/ide/server-pro/access_and_security/network_port_and_address.html \n\n"
	}

	finalMessage := "\nThank you for using wbi! \n\n" +
		"Workbench is now configured using the default PAM authentication method. Users with local Linux accounts and home directories should be able to log in to Workbench. \n\n" +
		serverAccessMessage +
		"Workbench integrates with a variety of Authentication types. To learn more about specific integrations, visit the documentation links below:\n" +
		"For more information on PAM authentication https://docs.posit.co/ide/server-pro/authenticating_users/pam_authentication.html. \n" + "For more information on Active Directory authentication " + adDocURL + ". \n" +
		"For more information on SAML Single Sign-On authentication https://docs.posit.co/ide/server-pro/authenticating_users/saml_sso.html. \n" +
		"For more information on OpenID Connect Single Sign-On authentication https://docs.posit.co/ide/server-pro/authenticating_users/openid_connect_authentication.html. \n" +
		"For more information on Proxied Authentication https://docs.posit.co/ide/server-pro/authenticating_users/proxied_authentication.html."

	system.PrintAndLogInfo(finalMessage)
	return nil
}

func setSetupOpts(setupOpts *setupOpts) {
	setupOpts.step = viper.GetString("step")
}

func (opts *setupOpts) Validate(args []string) error {
	// ensure no args are passed
	if len(args) > 0 {
		return fmt.Errorf("no arguments are supported for this command")
	}

	// ensure step is valid
	validSteps := []string{"start", "prereqs", "firewall", "security", "languages", "r", "python", "workbench", "license", "quarto", "jupyter", "prodrivers", "ssl", "packagemanager", "connect", "restart", "status", "verify"}
	if opts.step != "" && !lo.Contains(validSteps, opts.step) {
		return fmt.Errorf("invalid step: %s", opts.step)
	}

	return nil
}

func newSetupCmd() *setupCmd {
	root := &setupCmd{opts: setupOpts{}}

	// adding two spaces to have consistent formatting
	exampleText := []string{
		"To start an interactive setup process for Workbench:",
		"  wbi setup",
		"",
		"To start an interactive setup process for Workbench at a certain step:",
		"  wbi setup --step [STEP]",
	}

	cmd := &cobra.Command{
		Use:     "setup",
		Short:   "Launch an interactive setup process for Workbench",
		Example: strings.Join(exampleText, "\n"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			setSetupOpts(&root.opts)
			if err := root.opts.Validate(args); err != nil {
				return err
			}
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			//TODO: Add your logic to gather config to pass code here
			log.WithField("opts", fmt.Sprintf("%+v", root.opts)).Trace("setup-opts")
			if err := newSetup(root.opts); err != nil {
				return err
			}
			return nil
		},
		SilenceUsage: true,
	}

	stepHelp := `The step to start at. Valid steps are: start, prereqs, firewall, security, languages, r, python, workbench, license, quarto, jupyter, prodrivers, ssl, packagemanager, connect, restart, status, verify.`

	cmd.Flags().StringP("step", "s", "", stepHelp)
	viper.BindPFlag("step", cmd.Flags().Lookup("step"))

	root.cmd = cmd
	return root
}
