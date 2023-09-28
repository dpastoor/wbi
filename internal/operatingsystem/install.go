package operatingsystem

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sol-eng/wbi/internal/config"
	"github.com/sol-eng/wbi/internal/install"
	"github.com/sol-eng/wbi/internal/system"
)

func InstallPrereqs(osType config.OperatingSystem) error {
	system.PrintAndLogInfo("Installing prerequisites...")
	// Update apt and install gdebi-core if an Ubuntu system
	if osType == config.Ubuntu22 || osType == config.Ubuntu20 || osType == config.Ubuntu18 {
		AptErr := UpgradeApt()
		if AptErr != nil {
			return fmt.Errorf("UpgradeApt: %w", AptErr)
		}

		GdebiCoreErr := InstallGdebiCore()
		if GdebiCoreErr != nil {
			return fmt.Errorf("InstallGdebiCore: %w", GdebiCoreErr)
		}
	} else if osType == config.Redhat9 || osType == config.Redhat8 {
		// Enable the Extra Packages for Enterprise Linux (EPEL) repository
		EnableEPELErr := EnableEPELRepo(osType)
		if EnableEPELErr != nil {
			return fmt.Errorf("EnableEPELRepo: %w", EnableEPELErr)
		}
		// Enable the CodeReady Linux Builder repository
		OnCloud, err := PromptCloud()
		if err != nil {
			return fmt.Errorf("PrompOnPremCloud: %w", err)
		}
		EnableCodeReadyErr := EnableCodeReadyRepo(osType, OnCloud)
		if EnableCodeReadyErr != nil {
			return fmt.Errorf("EnableCodeReadyRepo: %w", EnableCodeReadyErr)
		}
	} else if osType == config.Redhat7 {
		// Enable the Extra Packages for Enterprise Linux (EPEL) repository
		EnableEPELErr := EnableEPELRepo(osType)
		if EnableEPELErr != nil {
			return fmt.Errorf("EnableEPELRepo: %w", EnableEPELErr)
		}
		// Enable the Extras Repo
		err := EnableExtraRepo()
		if err != nil {
			return fmt.Errorf("EnableExtraRepo: %w", err)
		}
		// Enable the CodeReady Linux Builder repository
		OnCloud, err := PromptCloud()
		if err != nil {
			return fmt.Errorf("PrompOnPremCloud: %w", err)
		}
		EnableCodeReadyErr := EnableCodeReadyRepo(osType, OnCloud)
		if EnableCodeReadyErr != nil {
			return fmt.Errorf("EnableCodeReadyRepo: %w", EnableCodeReadyErr)
		}
	} else {
		return errors.New("unsupported operating system")
	}
	system.PrintAndLogInfo("\nPrerequisites successfully installed!")
	return nil
}

// Installs Gdebi Core
func InstallGdebiCore() error {
	gdebiCoreCommand := "apt-get install -y gdebi-core"
	err := system.RunCommand(gdebiCoreCommand, true, 1, true)
	if err != nil {
		return fmt.Errorf("issue installing gdebi-core with the command '%s': %w", gdebiCoreCommand, err)
	}

	system.PrintAndLogInfo("\ngdebi-core has been successfully installed!")
	return nil
}

// Upgrades Apt
func UpgradeApt() error {
	aptUpgradeCommand := "apt-get update"
	err := system.RunCommand(aptUpgradeCommand, true, 1, true)
	if err != nil {
		return fmt.Errorf("issue upgrading apt with the command '%s': %w", aptUpgradeCommand, err)
	}

	system.PrintAndLogInfo("\napt has been successfully upgraded!")
	return nil
}

// Enable the CodeReady Linux Builder repository:
func EnableCodeReadyRepo(osType config.OperatingSystem, CloudInstall bool) error {
	if CloudInstall {
		switch osType {
		case config.Redhat9:
			dnfPluginsCoreCommand := "dnf install -y dnf-plugins-core"
			err := system.RunCommand(dnfPluginsCoreCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue installing dnf-plugins-core with the command '%s': %w", dnfPluginsCoreCommand, err)
			}
			enableCodeReadyCommand := `dnf config-manager --set-enabled "*codeready-builder-for-rhel-9-*-rpms"`
			err = system.RunCommand(enableCodeReadyCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue enabling the CodeReady Linux Builder repo with the command '%s': %w", enableCodeReadyCommand, err)
			}
		case config.Redhat8:
			dnfPluginsCoreCommand := "dnf install -y dnf-plugins-core"
			err := system.RunCommand(dnfPluginsCoreCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue installing dnf-plugins-core with the command '%s': %w", dnfPluginsCoreCommand, err)
			}

			enableCodeReadyCommand := `dnf config-manager --set-enabled "*codeready-builder-for-rhel-8-*-rpms"`
			err = system.RunCommand(enableCodeReadyCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue enabling the CodeReady Linux Builder repo with the command '%s': %w", enableCodeReadyCommand, err)
			}
		case config.Redhat7:
			yumUtilsCoreCommand := "sudo yum install -y yum-utils"
			err := system.RunCommand(yumUtilsCoreCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue installing yum-utils with the command '%s': %w", yumUtilsCoreCommand, err)
			}

			enableCodeReadyCommand := `sudo yum-config-manager --enable "rhel-*-optional-rpms"`
			err = system.RunCommand(enableCodeReadyCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue enabling the CodeReady Linux Builder repo with the command '%s': %w", enableCodeReadyCommand, err)
			}
		}
	} else if !CloudInstall {
		switch osType {
		case config.Redhat9:
			OnPremCodeReadyEnableCommand := "sudo subscription-manager repos --enable codeready-builder-for-rhel-9-$(arch)-rpms\n"
			err := system.RunCommand(OnPremCodeReadyEnableCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue enabling codeready repo with the command '%s': %w", OnPremCodeReadyEnableCommand, err)
			}
		case config.Redhat8:
			OnPremCodeReadyEnableCommand := "sudo subscription-manager repos --enable codeready-builder-for-rhel-8-x86_64-rpms\n"
			err := system.RunCommand(OnPremCodeReadyEnableCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue enabling codeready repo with the command '%s': %w", OnPremCodeReadyEnableCommand, err)
			}
		case config.Redhat7:
			OnPremCodeReadyEnableCommand := "sudo subscription-manager repos --enable \"rhel-*-optional-rpms\""
			err := system.RunCommand(OnPremCodeReadyEnableCommand, true, 1, true)
			if err != nil {
				return fmt.Errorf("issue enabling codeready repo with the command '%s': %w", OnPremCodeReadyEnableCommand, err)
			}
		}
	} else {
		return fmt.Errorf("issue enabling codeready repo: CloudInstall boolean undefined")
	}
	system.PrintAndLogInfo("\nThe CodeReady Linux Builder repository has been successfully enabled!")
	return nil
}

// Enable the Extra Repo
func EnableExtraRepo() error {
	extraCommand := "yum-config-manager --enable rhel-7-server-rhui-extras-rpms"
	commandOutput, err := system.RunCommandAndCaptureOutput(extraCommand, true, 1, true)
	if err != nil {
		return fmt.Errorf("issue enabling extra repo with the command '%s': %w", commandOutput, err)
	}

	system.PrintAndLogInfo("\nThe Extra Repository has been successfully enabled!")
	return nil
}

// Enable the Extra Packages for Enterprise Linux (EPEL) repository
func EnableEPELRepo(osType config.OperatingSystem) error {
	var EPELURL string
	if osType == config.Redhat9 {
		EPELURL = "https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm"
	} else if osType == config.Redhat8 {
		EPELURL = "https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm"
	} else if osType == config.Redhat7 {
		EPELURL = "https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm"
	} else {
		return errors.New("operating system not supported")
	}

	EPELCommand, err := install.RetrieveInstallCommand(EPELURL, osType)
	if err != nil {
		return fmt.Errorf("issue retrieving EPEL install command: %w", err)
	}
	commandOutput, err := system.RunCommandAndCaptureOutput(EPELCommand, true, 1, true)
	if err != nil {
		if strings.Contains(commandOutput, "does not update installed package") && osType == config.Redhat7 {
			system.PrintAndLogInfo("\nThe Extra Packages for Enterprise Linux (EPEL) repository was already enabled.")
			return nil
		}
		return fmt.Errorf("issue enabling EPEL repo with the command '%s': %w", commandOutput, err)
	}
	if err != nil {
		return fmt.Errorf("issue enabling EPEL repo with the command '%s': %w", commandOutput, err)
	}

	system.PrintAndLogInfo("\nThe Extra Packages for Enterprise Linux (EPEL) repository has been successfully enabled!")
	return nil
}

// Disable local firewall on server
func DisableFirewall(osType config.OperatingSystem) error {
	var FWCommand string

	switch osType {
	case config.Ubuntu18, config.Ubuntu20, config.Ubuntu22:
		FWCommand = "ufw disable"
	case config.Redhat7, config.Redhat8, config.Redhat9:
		FWCommand = "systemctl stop firewalld && systemctl disable firewalld"
	default:
		return errors.New("Unsupported OS, setting FWCommand failed") //nolint:all
	}
	err := system.RunCommand(FWCommand, true, 1, true)
	if err != nil {
		return fmt.Errorf("issue disabling system firewall with the command '%s': %w", FWCommand, err)
	}

	system.PrintAndLogInfo("\nThe system firewall has been successfully disabled!")
	return nil
}

func DisableLinuxSecurity() error {

	setenforceCommand := "setenforce 0"
	err := system.RunCommand(setenforceCommand, true, 1, true)
	if err != nil {
		return fmt.Errorf("issue stopping selinux enforcement with the command '%s': %w", setenforceCommand, err)
	}

	disableSELinuxCommand := "sed -i s/^SELINUX=.*$/SELINUX=disabled/ /etc/selinux/config"
	err = system.RunCommand(disableSELinuxCommand, true, 1, true)
	if err != nil {
		return fmt.Errorf("issue disabling selinux with the command '%s': %w", disableSELinuxCommand, err)
	}

	system.PrintAndLogInfo("\nThe SELinux has been successfully changed to permissive mode, and will be disabled on next reboot!")
	return nil
}
