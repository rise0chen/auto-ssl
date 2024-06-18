package cert

import (
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type CertConfig struct {
	Test   bool     `json:"test"`
	Email  string   `json:"email"`
	Domain []string `json:"domain"`
}

type Certificate struct {
	Config  CertConfig
	Public  string
	Private string
}

func GetCertificate(user User, dns challenge.Provider, cfg CertConfig) (Certificate, error) {
	config := lego.NewConfig(&user)
	if cfg.Test {
		config.CADirURL = lego.LEDirectoryStaging
	} else {
		config.CADirURL = lego.LEDirectoryProduction
	}

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	// config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return Certificate{}, err
	}

	// We specify an HTTP port of 5002 and an TLS port of 5001 on all interfaces
	// because we aren't running as root and can't bind a listener to port 80 and 443
	// (used later when we attempt to pass challenges). Keep in mind that you still
	// need to proxy challenge traffic to port 5002 and 5001.
	err = client.Challenge.SetDNS01Provider(dns)
	if err != nil {
		return Certificate{}, err
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return Certificate{}, err
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: cfg.Domain,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return Certificate{}, err
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	// fmt.Printf("%#v\n", certificates)
	return Certificate{
		Config:  cfg,
		Public:  string(certificates.Certificate),
		Private: string(certificates.PrivateKey),
	}, nil
}
