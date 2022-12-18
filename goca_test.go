package goca

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const CaTestFolder string = "./DoNotUseThisCAPATHTestOnly"
const GoodKeyPerms os.FileMode = 0600

func tearDown() {
	os.Unsetenv("GOCATEST")
	os.RemoveAll(CaTestFolder)
}

// TestFunctionalRootCACreation creates a RootCA
func TestFunctionalRootCACreation(t *testing.T) {
	tearDown()
	os.Setenv("CAPATH", CaTestFolder)
	os.Setenv("GOCATEST", "true")

	cert := x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			Organization:       []string{"GO CA Root Company Inc."},
			OrganizationalUnit: []string{"Certificates Management"},
			Country:            []string{"NL"},
			Locality:           []string{"Noord-Brabant"},
			Province:           []string{"Veldhoven"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		IsCA:        true,
		DNSNames:    []string{"www.go-root.ca", "secure.go-root.ca"},
		ExtKeyUsage: []x509.ExtKeyUsage{},
		KeyUsage:    x509.KeyUsageCRLSign | x509.KeyUsageCertSign,
	}

	RootCompanyCA, err := New("go-root.ca", &cert)
	if err != nil {
		t.Errorf("Failing to create the CA")
	}
	if RootCompanyCA.IsIntermediate() != false {
		t.Errorf("Intermediate is true instead false")
	}

	if RootCompanyCA.Status() != "Certificate Authority is ready." {
		t.Errorf(RootCompanyCA.Status())
	}

	fi, err := os.Stat(filepath.Join(CaTestFolder, "go-root.ca", "ca", "key.pem"))
	if err != nil {
		t.Errorf("key.pem does not exist for the CA")
	}
	if fi.Mode() != GoodKeyPerms {
		t.Errorf("Expected key.pem permissions " + fmt.Sprint(GoodKeyPerms) + " but got: " + fmt.Sprint(fi.Mode()))
	}

	t.Log("Tested Creating a Root CA")

}

// Creates a Intermediate CA
func TestFunctionalIntermediateCACreation(t *testing.T) {
	os.Setenv("CAPATH", CaTestFolder)

	cert := x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			Organization:       []string{"GO CA Intermediate Company Inc."},
			OrganizationalUnit: []string{"Certificates Management"},
			Country:            []string{"NL"},
			Locality:           []string{"Noord-Brabant"},
			Province:           []string{"Veldhoven"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		IsCA:        true,
		DNSNames:    []string{"www.go-intermediate.ca", "secure.go-intermediate.ca"},
		ExtKeyUsage: []x509.ExtKeyUsage{},
		KeyUsage:    x509.KeyUsageCRLSign | x509.KeyUsageCertSign,
	}

	IntermediateCA, err := NewCA("go-intermediate.ca", "go-root.ca", &cert)
	if err != nil {
		t.Log(err)
		t.Errorf("Failing to create the CA")
	}

	if IntermediateCA.IsIntermediate() != true {
		t.Errorf("Intermediate is false instead true")
	}

	fi, err := os.Stat(filepath.Join(CaTestFolder, "go-intermediate.ca", "ca", "key.pem"))
	if err != nil {
		t.Errorf("key.pem does not exist for the CA")
	}
	if fi.Mode() != GoodKeyPerms {
		t.Errorf("Expected key.pem permissions " + fmt.Sprint(GoodKeyPerms) + " but got: " + fmt.Sprint(fi.Mode()))
	}

	t.Log("Tested Creating a Intermediate CA")

}

func TestFunctionalListCAs(t *testing.T) {
	if len(List()) == 0 {
		t.Error("Empty list of CAs")
	}
	t.Log(List())
}

func TestFunctionalRootCAIssueNewCertificate(t *testing.T) {
	certRequest := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization:       []string{"SFTP Server CA Company Inc."},
			OrganizationalUnit: []string{"Intermediate Certificates Management"},
			Country:            []string{"NL"},
			Locality:           []string{"Noord-Brabant"},
			Province:           []string{"Veldhoven"},
		},
		DNSNames: []string{"w3.intranet.go-root.ca"},
	}

	RootCA, err := Load("go-root.ca")
	if err != nil {
		t.Log(err)
		t.Errorf("Failed to load Root CA")
	}

	intranetCert, err := RootCA.IssueCertificate("intranet.go-root.ca", &certRequest, 100)
	if err != nil {
		t.Log(err)
		t.Errorf("Failed to Root CA issue new certificate (intranet.go-root.ca)")
	}

	fmt.Println(RootCA.ListCertificates())

	if RootCA.GetCertificate() != intranetCert.GetCACertificate() {
		t.Log(RootCA.GetCertificate())
		t.Log(intranetCert.GetCACertificate())
		t.Error("The CA Certificate is not the same as the Certificate CA Certificate")
	}

	fi, err := os.Stat(filepath.Join(CaTestFolder, "go-root.ca", "certs", "intranet.go-root.ca", "key.pem"))
	if err != nil {
		t.Errorf("key.pem does not exist for the identity")
	}
	if fi.Mode() != GoodKeyPerms {
		t.Errorf("Expected key.pem permissions " + fmt.Sprint(GoodKeyPerms) + " but got: " + fmt.Sprint(fi.Mode()))
	}
}

func TestFunctionalRootCALoadCertificates(t *testing.T) {

	RootCA, err := Load("go-root.ca")
	if err != nil {
		t.Log(err)
		t.Errorf("Failed to load Root CA")
	}

	intranetCert, err := RootCA.LoadCertificate("intranet.go-root.ca")
	if err != nil {
		fmt.Println(err)
		t.Log(err)
	}

	if intranetCert.GetCACertificate() != "" {
		t.Log("Failed to load intranet")
	}
	intermediateCert, _ := RootCA.LoadCertificate("go-intermediate.ca")

	if RootCA.GetCertificate() != intermediateCert.GetCACertificate() {
		t.Log(RootCA.GetCertificate())
		t.Log(intermediateCert.GetCACertificate())
		t.Error("The CA Certificate is not the same as the Certificate CA Certificate")
	}

}

func TestFunctionalIntermediateCAIssueNewCertificate(t *testing.T) {
	certRequest := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization:       []string{"An Organization"},
			OrganizationalUnit: []string{"An Organizational Unit"},
			Country:            []string{"NL"},
			Locality:           []string{"Noord-Brabant"},
			Province:           []string{"Veldhoven"},
		},
		DNSNames: []string{"anorg.go-intermediate.ca"},
	}

	interCA, err := Load("go-intermediate.ca")
	if err != nil {
		t.Errorf("Failed to load intermediate CA")
	}

	idCert, err := interCA.IssueCertificate("anorg.go-intermediate.ca", &certRequest, 100)
	if err != nil {
		t.Error("Failed to issue certificate anorg.go-intermediate.ca")
	}

	fmt.Println(interCA.ListCertificates())

	if interCA.GetCertificate() != idCert.GetCACertificate() {
		t.Error("CA certificate mismatch between intermediate CA and issued certificate.")
	}
}

func TestFunctionalRevokeCertificate(t *testing.T) {
	RootCA, _ := Load("go-root.ca")
	intermediateCert, _ := RootCA.LoadCertificate("go-intermediate.ca")

	if RootCA.Data.crl == nil {
		t.Error("CRL is nil")
	}

	err := RootCA.RevokeCertificate("go-intermediate.ca")
	if err != nil {
		t.Error("Failed to revoke certificate")
	}
	t.Log(intermediateCert.certificate.SerialNumber)
	t.Log(RootCA.Data.crl.TBSCertList.RevokedCertificates[0].SerialNumber)
	result := intermediateCert.certificate.SerialNumber.Cmp(RootCA.Data.crl.TBSCertList.RevokedCertificates[0].SerialNumber)
	if result != 0 {
		t.Error("Certificate Serial Number is not in the CRL")
	}

	t.Log("Negative check")
	intranetCert, _ := RootCA.LoadCertificate("intranet.go-root.ca")
	t.Log(intranetCert.certificate.SerialNumber)
	t.Log(RootCA.Data.crl.TBSCertList.RevokedCertificates[0].SerialNumber)
	result = intranetCert.certificate.SerialNumber.Cmp(RootCA.Data.crl.TBSCertList.RevokedCertificates[0].SerialNumber)
	if result == 0 {
		t.Error("Non revoked certificate in list")
	}
	err = RootCA.RevokeCertificate("intranet.go-root.ca")
	if err != nil {
		t.Error("Failed to revoke.")
	}
	t.Log(RootCA.Data.crl.TBSCertList.RevokedCertificates)
	if len(RootCA.Data.crl.TBSCertList.RevokedCertificates) != 2 {
		t.Error("Not appending certificates to revoke list")
	}
	t.Logf("Test appending revoked certificates")

	if RootCA.GetCRL() == "" {
		t.Error("CRL X509 file is empty!")
	}
}
