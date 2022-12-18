package goca_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/neb42/goca"
)

func Example_minimal() {

	// Define the GOCAPTH (Default is current dir)
	os.Setenv("CAPATH", "/opt/GoCA/CA")

	// Root cert for creation
	rootCert := x509.Certificate{
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

	// Create the New Root CA or loads existent from disk ($CAPATH)
	RootCA, err := goca.New("mycompany.com", &rootCert)
	if err != nil {
		// Loads in case it exists
		fmt.Println("Loading CA")
		RootCA, err = goca.Load("gocaroot.nl")
		if err != nil {
			log.Fatal(err)
		}

		// Check the CA status and shows the CA Certificate
		fmt.Println(RootCA.Status())
		fmt.Println(RootCA.GetCertificate())

	} else {
		log.Fatal(err)
	}

	// Issue certificate for example intranet server
	certRequest := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization:       []string{"Intranet Company Inc."},
			OrganizationalUnit: []string{"Global Intranet"},
			Country:            []string{"NL"},
			Locality:           []string{"Noord-Brabant"},
			Province:           []string{"Veldhoven"},
		},
		DNSNames: []string{"w3.intranet.example.com", "www.intranet.example.com"},
	}

	intranetCert, err := RootCA.IssueCertificate("intranet.example.com", &certRequest, 100)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(intranetCert.GetCertificate())

	// Shows all CA Certificates
	fmt.Println(RootCA.ListCertificates())
}
