package main

import (
	"C"
	"github.com/sol-eng/wbi/internal/config"
	"github.com/sol-eng/wbi/internal/languages"
	"github.com/sol-eng/wbi/internal/quarto"
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

//export rversions
func rversions() *C.char {
	result, _ := languages.RetrieveValidRVersions()
	return C.CString(strings.Join(result, " "))
}

//export rurls
func rurls(os *C.char, rVersion *C.char) *C.char {
	var osType config.OperatingSystem
	osGo := C.GoString(os)
	rVersionGo := C.GoString(rVersion)
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
	}
	installerInfo, _ := languages.PopulateInstallerInfo("r", rVersionGo, osType)
	//fmt.Println(installerInfo.URL)
	return C.CString(installerInfo.URL)
}

//export pythonversions
func pythonversions(os *C.char) *C.char {
	var osType config.OperatingSystem
	osGo := C.GoString(os)
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
	}

	result, _ := languages.RetrieveValidPythonVersions(osType)
	return C.CString(strings.Join(result, " "))
}

//export pythonurls
func pythonurls(os *C.char, pythonVersion *C.char) *C.char {
	var osType config.OperatingSystem
	osGo := C.GoString(os)
	pythonVersionGo := C.GoString(pythonVersion)
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
	}
	installerInfo, _ := languages.PopulateInstallerInfo("python", pythonVersionGo, osType)
	//fmt.Println(installerInfo.URL)
	return C.CString(installerInfo.URL)
}

//export quartoversions
func quartoversions() *C.char {
	result, _ := quarto.RetrieveValidQuartoVersions()
	return C.CString(strings.Join(result, " "))
}
