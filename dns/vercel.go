package dns

import (
	"time"

	"github.com/go-acme/lego/v4/providers/dns/vercel"
)

type VercelConfig struct {
	AuthToken string `json:"authToken"`
	TeamID    string `json:"teamID"`
}

type VercelDns struct {
	Client *vercel.DNSProvider
}

func NewVercelDns(config VercelConfig) (VercelDns, error) {
	cfg := vercel.NewDefaultConfig()
	cfg.AuthToken = config.AuthToken
	cfg.TeamID = config.TeamID

	client, err := vercel.NewDNSProviderConfig(cfg)
	if err != nil {
		return VercelDns{}, err
	}
	return VercelDns{
		Client: client,
	}, nil
}

func (dns *VercelDns) Present(domain, token, keyAuth string) error {
	return dns.Client.Present(domain, token, keyAuth)
}

func (dns *VercelDns) CleanUp(domain, token, keyAuth string) error {
	return dns.Client.CleanUp(domain, token, keyAuth)
}

func (dns *VercelDns) Timeout() (timeout, interval time.Duration) {
	return dns.Client.Timeout()
}
