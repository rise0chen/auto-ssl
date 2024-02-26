package dns

import (
	"fmt"
	"log"
	"strings"
	"time"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

type AliyunConfig struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type AliyunDns struct {
	Client *alidns.Client
	ID     []string
}

func NewAliyunDns(config AliyunConfig) (AliyunDns, error) {
	client, err := alidns.NewClient(&openapi.Config{AccessKeyId: &config.AccessKey, AccessKeySecret: &config.SecretKey})
	if err != nil {
		return AliyunDns{}, err
	}
	return AliyunDns{
		Client: client,
		ID:     make([]string, 0),
	}, nil
}

func (ali *AliyunDns) AddRecord(domain, rr, r_type, value string) {
	resp, err := ali.Client.AddDomainRecord(&alidns.AddDomainRecordRequest{
		DomainName: &domain,
		RR:         &rr,
		Type:       &r_type,
		Value:      &value,
	})
	if err != nil {
		log.Fatal(err)
	}
	ali.ID = append(ali.ID, *resp.Body.RecordId)
}

func (ali *AliyunDns) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	rr, success := strings.CutSuffix(info.FQDN, "."+domain+".")
	if !success {
		return fmt.Errorf("no such suffix: %s", info)
	}
	ali.AddRecord(domain, rr, "TXT", info.Value)
	return nil
}

func (ali *AliyunDns) CleanUp(domain, token, keyAuth string) error {
	for _, id := range ali.ID {
		_, err := ali.Client.DeleteDomainRecord(&alidns.DeleteDomainRecordRequest{RecordId: &id})
		if err != nil {
			log.Fatal(err)
		}
	}
	ali.ID = make([]string, 0)
	return nil
}

func (ali *AliyunDns) Timeout() (timeout, interval time.Duration) {
	return 12 * time.Minute, 10 * time.Second
}
