package languages

import (
	"errors"
	"strings"

	"github.com/sol-eng/wbi/internal/config"
)

// InstallerInfo contains the information needed to download and install R and Python
type InstallerInfo struct {
	Name    string
	URL     string
	Version string
}

func PopulateInstallerInfo(language string, version string, osType config.OperatingSystem) (InstallerInfo, error) {
	switch osType {
	case config.Ubuntu20:
		return InstallerInfo{
			Name:    language + "-" + version + "_1_amd64.deb",
			URL:     "https://cdn.rstudio.com/" + language + "/ubuntu-2004/pkgs/" + language + "-" + version + "_1_amd64.deb",
			Version: version,
		}, nil
	case config.Ubuntu22:
		return InstallerInfo{
			Name:    language + "-" + version + "_1_amd64.deb",
			URL:     "https://cdn.rstudio.com/" + language + "/ubuntu-2204/pkgs/" + language + "-" + version + "_1_amd64.deb",
			Version: version,
		}, nil
	case config.Redhat7:
		// Redhat 7 R URL uses lowercase "r" in the beginning but then "R" in the 2nd occurance
		if language == "r" {
			return InstallerInfo{
				Name:    strings.ToUpper(language) + "-" + version + "-1-1.x86_64.rpm",
				URL:     "https://cdn.rstudio.com/" + language + "/centos-7/pkgs/" + strings.ToUpper(language) + "-" + version + "-1-1.x86_64.rpm",
				Version: version,
			}, nil
		} else {
			return InstallerInfo{
				Name:    language + "-" + version + "-1-1.x86_64.rpm",
				URL:     "https://cdn.rstudio.com/" + language + "/centos-7/pkgs/" + language + "-" + version + "-1-1.x86_64.rpm",
				Version: version,
			}, nil
		}
	case config.Redhat8:
		// Redhat 8 R URL uses lowercase "r" in the beginning but then "R" in the 2nd occurance
		if language == "r" {
			return InstallerInfo{
				Name:    strings.ToUpper(language) + "-" + version + "-1-1.x86_64.rpm",
				URL:     "https://cdn.rstudio.com/" + language + "/centos-8/pkgs/" + strings.ToUpper(language) + "-" + version + "-1-1.x86_64.rpm",
				Version: version,
			}, nil
		} else {
			return InstallerInfo{
				Name:    language + "-" + version + "-1-1.x86_64.rpm",
				URL:     "https://cdn.rstudio.com/" + language + "/centos-8/pkgs/" + language + "-" + version + "-1-1.x86_64.rpm",
				Version: version,
			}, nil
		}
	case config.Redhat9:
		// Redhat 9 R URL uses lowercase "r" in the beginning but then "R" in the 2nd occurance
		if language == "r" {
			return InstallerInfo{
				Name:    strings.ToUpper(language) + "-" + version + "-1-1.x86_64.rpm",
				URL:     "https://cdn.rstudio.com/" + language + "/rhel-9/pkgs/" + strings.ToUpper(language) + "-" + version + "-1-1.x86_64.rpm",
				Version: version,
			}, nil
		} else {
			return InstallerInfo{
				Name:    language + "-" + version + "-1-1.x86_64.rpm",
				URL:     "https://cdn.rstudio.com/" + language + "/rhel-9/pkgs/" + language + "-" + version + "-1-1.x86_64.rpm",
				Version: version,
			}, nil
		}
	default:
		return InstallerInfo{}, errors.New("operating system not supported")
	}
}
