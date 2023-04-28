package ssl

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/sol-eng/wbi/internal/config"
	"github.com/sol-eng/wbi/internal/system"
)

func TrustRootCertificate(rootCA *x509.Certificate, osType config.OperatingSystem) error {
	fmt.Println(rootCA.Raw)
	block := pem.Block{Type: "CERTIFICATE", Bytes: rootCA.Raw}
	var pemCert []string
	pemCert = append(pemCert, string(pem.EncodeToMemory(&block)))
	switch osType {
	case config.Ubuntu18, config.Ubuntu20, config.Ubuntu22:
		err := system.WriteStrings(pemCert, "/usr/local/share/ca-certificates/workbenchCA.crt", 0755)
		if err != nil {
			return fmt.Errorf("writing certificate to disk failed: ", err)
		}
		err = system.RunCommand("update-ca-certificates", true, 1)
		if err != nil {
			return fmt.Errorf("running command to trust root certificate: ", err)
		}
	case config.Redhat7, config.Redhat8, config.Redhat9:
		err := system.WriteStrings(pemCert, "/etc/pki/ca-trust/source/anchors/workbenchCA.crt", 0755)
		if err != nil {
			return fmt.Errorf("writing CA certificate to disk failed: ", err)
		}
		err = system.RunCommand("update-ca-trust", true, 1)
		if err != nil {
			return fmt.Errorf("running command to trust root certificate: ", err)
		}
	}
	return nil
}
