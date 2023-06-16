package main

import (
	"C"
	"github.com/sol-eng/wbi/internal/config"
	"github.com/sol-eng/wbi/internal/languages"
	"github.com/sol-eng/wbi/internal/prodrivers"
	"github.com/sol-eng/wbi/internal/quarto"
	"github.com/sol-eng/wbi/internal/workbench"
	"strings"
)

// https://goreleaser.com/cookbooks/using-main.version
var (
	version string = "dev"
	commit  string = "none"
	date    string = "unknown"
)

func main() {
}

func OSSwitch(osGo string) config.OperatingSystem {
	var osType config.OperatingSystem
	switch osGo {
	case "U18":
		osType = config.Ubuntu18
	case "U20":
		osType = config.Ubuntu20
	case "U22":
		osType = config.Ubuntu22
	case "RH7":
		osType = config.Redhat7
	case "RH8":
		osType = config.Redhat8
	case "RH9":
		osType = config.Redhat9
	default:
		osType = config.Redhat8
	}

	return osType

}

//export rversions
func rversions() *C.char {
	result, _ := languages.RetrieveValidRVersions()
	return C.CString(strings.Join(result, " "))
}

//export rurls
func rurls(os *C.char, rVersion *C.char) *C.char {
	osGo := C.GoString(os)
	rVersionGo := C.GoString(rVersion)
	osType := OSSwitch(osGo)
	installerInfo, _ := languages.PopulateInstallerInfo("r", rVersionGo, osType)
	//fmt.Println(installerInfo.URL)
	return C.CString(installerInfo.URL)
}

//export pythonversions
func pythonversions(os *C.char) *C.char {
	osGo := C.GoString(os)
	osType := OSSwitch(osGo)

	result, _ := languages.RetrieveValidPythonVersions(osType)
	return C.CString(strings.Join(result, " "))
}

//export pythonurls
func pythonurls(os *C.char, pythonVersion *C.char) *C.char {
	osGo := C.GoString(os)
	pythonVersionGo := C.GoString(pythonVersion)
	osType := OSSwitch(osGo)
	installerInfo, _ := languages.PopulateInstallerInfo("python", pythonVersionGo, osType)
	//fmt.Println(installerInfo.URL)
	return C.CString(installerInfo.URL)
}

//export quartoversions
func quartoversions() *C.char {
	result, _ := quarto.RetrieveValidQuartoVersions()
	return C.CString(strings.Join(result, " "))
}

//export quartourls
func quartourls(os *C.char, quartoVersion *C.char) *C.char {
	osGo := C.GoString(os)
	quartoVersionGo := C.GoString(quartoVersion)
	osType := OSSwitch(osGo)
	quartoURL := quarto.GenerateQuartoInstallURL(quartoVersionGo, osType)

	//fmt.Println(installerInfo.URL)
	return C.CString(quartoURL)
}

//export workbenchurl
func workbenchurl(os *C.char) (*C.char, *C.char) {
	osGo := C.GoString(os)
	osType := OSSwitch(osGo)
	// Retrieve JSON data
	rstudio, _ := workbench.RetrieveWorkbenchInstallerInfo()
	// Retrieve installer info
	installerInfo, _ := rstudio.GetInstallerInfo(osType)
	//fmt.Println(installerInfo.URL)
	return C.CString(installerInfo.Version), C.CString(installerInfo.URL)
}

//export driverurl
func driverurl(os *C.char) (*C.char, *C.char) {
	osGo := C.GoString(os)
	osType := OSSwitch(osGo)
	// Retrieve JSON data
	rstudio, _ := prodrivers.RetrieveProDriversInstallerInfo()
	// Retrieve installer info
	installerInfo, _ := rstudio.GetInstallerInfo(osType)
	//fmt.Println(installerInfo.URL)
	return C.CString(installerInfo.Version), C.CString(installerInfo.URL)
}
