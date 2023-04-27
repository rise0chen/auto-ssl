package main

import (
	"crypto"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	cas "github.com/alibabacloud-go/cas-20200407/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	slb "github.com/alibabacloud-go/slb-20140515/v4/client"
	oss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/registration"
)

var (
	ALICLOUD_RAM_ROLE   = os.Getenv("RAM_ROLE")
	ALICLOUD_ACCESS_KEY = os.Getenv("ACCESS_KEY")
	ALICLOUD_SECRET_KEY = os.Getenv("ACCESS_SECRET")
)

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type AliDns struct {
	AccessID  string
	AccessKey string
	Client    *alidns.Client
	ID        []string
}

func NewAliDns(id, key, region string) AliDns {
	client, err := alidns.NewClient(&openapi.Config{AccessKeyId: &id, AccessKeySecret: &key, RegionId: &region})
	if err != nil {
		log.Fatal(err)
	}
	return AliDns{
		AccessID:  id,
		AccessKey: key,
		Client:    client,
		ID:        make([]string, 0),
	}
}

func (ali *AliDns) AddRecord(domain, rr, r_type, value string) {
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

func (ali *AliDns) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	rr, success := strings.CutSuffix(fqdn, "."+domain+".")
	if !success {
		return fmt.Errorf("no such suffix: %s", fqdn)
	}
	ali.AddRecord(domain, rr, "TXT", value)
	return nil
}

func (ali *AliDns) CleanUp(domain, token, keyAuth string) error {
	for _, id := range ali.ID {
		_, err := ali.Client.DeleteDomainRecord(&alidns.DeleteDomainRecordRequest{RecordId: &id})
		if err != nil {
			log.Fatal(err)
		}
	}
	ali.ID = make([]string, 0)
	return nil
}

func main() {
	region := "cn-shanghai"
	domain := "lebai.ltd"
	dns := NewAliDns(ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY, region)
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	myUser := MyUser{
		Email: "opensoft@lebai.ltd",
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	// config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// We specify an HTTP port of 5002 and an TLS port of 5001 on all interfaces
	// because we aren't running as root and can't bind a listener to port 80 and 443
	// (used later when we attempt to pass challenges). Keep in mind that you still
	// need to proxy challenge traffic to port 5002 and 5001.
	err = client.Challenge.SetDNS01Provider(&dns)
	if err != nil {
		log.Fatal(err)
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{domain, "*." + domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	// fmt.Printf("%#v\n", certificates)

	cert := string(certificates.Certificate)
	key := string(certificates.PrivateKey)

	aliConfig := openapi.Config{AccessKeyId: &ALICLOUD_ACCESS_KEY, AccessKeySecret: &ALICLOUD_SECRET_KEY, RegionId: &region}
	casClient, err := cas.NewClient(&aliConfig)
	if err != nil {
		log.Fatal(err)
	}
	slbClient, err := slb.NewClient(&aliConfig)
	if err != nil {
		log.Fatal(err)
	}
	endpoint := "oss-cn-shanghai.aliyuncs.com"
	ossClient, err := oss.New("https://"+endpoint, ALICLOUD_ACCESS_KEY, ALICLOUD_SECRET_KEY)
	if err != nil {
		log.Fatal(err)
	}
	ali := AliClient{
		dns:    &dns,
		oss:    ossClient,
		slb:    slbClient,
		cas:    casClient,
		cert:   cert,
		key:    key,
		domain: domain,
	}
	certId, err := ali.UploadUserCertificate()
	if err != nil {
		log.Fatal(err)
	}

	_, err = ali.slb.UploadServerCertificate(&slb.UploadServerCertificateRequest{RegionId: &region, ServerCertificate: &cert, PrivateKey: &key})
	if err != nil {
		log.Fatal(err)
	}

	buckets, err := ossClient.ListBuckets()
	if err != nil {
		log.Fatal(err)
	}
	bucketsName := make([]string, 0)
	for _, bucket := range buckets.Buckets {
		bucketsName = append(bucketsName, bucket.Name)
	}

	for _, bucket := range bucketsName {
		ali.AddBucketCname(bucket, certId)
	}
	dns.CleanUp("", "", "")
}

func (client *AliClient) AddBucketCname(bucket string, certId int64) {
	cname := bucket + "." + client.domain
	cnResult, err := client.oss.ListBucketCname(bucket)
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range cnResult.Cname {
		if n.Domain == cname {
			putCnameConfig := oss.PutBucketCname{
				Cname: cname,
				CertificateConfiguration: &oss.CertificateConfiguration{
					CertId:      fmt.Sprint(certId),
					Certificate: client.cert,
					PrivateKey:  client.key,
					Force:       true,
				},
			}
			err = client.oss.PutBucketCnameWithCertificate(bucket, putCnameConfig)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
	}
}

func (client *AliClient) UploadUserCertificate() (int64, error) {
	now := time.Now().String()
	resp, err := client.cas.UploadUserCertificate(&cas.UploadUserCertificateRequest{Name: &now, Cert: &client.cert, Key: &client.key})
	if err != nil {
		return 0, err
	}
	certId := resp.Body.CertId
	return *certId, nil
}

type AliClient struct {
	dns    *AliDns
	oss    *oss.Client
	slb    *slb.Client
	cas    *cas.Client
	cert   string
	key    string
	domain string
}
