package main

import (
	"log"
	"os"
	"strings"

	"encoding/json"

	"github.com/go-acme/lego/v4/challenge"

	"lebai.ltd/auto_ssl/cert"
	"lebai.ltd/auto_ssl/deployment"
	"lebai.ltd/auto_ssl/deployment/aliyun"
	"lebai.ltd/auto_ssl/deployment/k8s"
	"lebai.ltd/auto_ssl/dns"
)

type Config struct {
	Dns        dns.DnsConfig               `json:"dns"`
	Cert       cert.CertConfig             `json:"cert"`
	Deployment deployment.DeploymentConfig `json:"deployment"`
}

func main() {
	configStr := os.Getenv("CONFIG")
	config := Config{}
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(config)
	if strings.HasPrefix(config.Dns.Aliyun.AccessKey, "***") {
		config.Dns.Aliyun.AccessKey = os.Getenv("ALIYUN_ACCESSKEY" + config.Dns.Aliyun.AccessKey[3:])
	}
	if strings.HasPrefix(config.Dns.Aliyun.SecretKey, "***") {
		config.Dns.Aliyun.SecretKey = os.Getenv("ALIYUN_SECRETKET" + config.Dns.Aliyun.SecretKey[3:])
	}
	if strings.HasPrefix(config.Deployment.Aliyun.AccessKey, "***") {
		config.Deployment.Aliyun.AccessKey = os.Getenv("ALIYUN_ACCESSKEY" + config.Deployment.Aliyun.AccessKey[3:])
	}
	if strings.HasPrefix(config.Deployment.Aliyun.SecretKey, "***") {
		config.Deployment.Aliyun.SecretKey = os.Getenv("ALIYUN_SECRETKET" + config.Deployment.Aliyun.SecretKey[3:])
	}
	if strings.HasPrefix(config.Deployment.K8s.Kube, "***") {
		config.Deployment.K8s.Kube = os.Getenv("KUBE_CONFIG" + config.Deployment.K8s.Kube[3:])
	}

	var dnsProvider challenge.Provider
	if config.Dns.Aliyun.AccessKey != "" && !strings.HasPrefix(config.Dns.Aliyun.AccessKey, "***") {
		dns, err := dns.NewAliyunDns(config.Dns.Aliyun)
		if err != nil {
			log.Fatal(err)
		}
		dnsProvider = &dns
	} else {
		log.Fatal("No DNS Provider")
		return
	}

	user, err := cert.NewUser(config.Cert.Email)
	if err != nil {
		log.Fatal(err)
	}
	certificate, err := cert.GetCertificate(user, dnsProvider, config.Cert.Domain, config.Cert.Test)
	if err != nil {
		log.Fatal(err)
	}

	if config.Deployment.Aliyun.AccessKey != "" {
		err := aliyun.DeploymentAliyun(config.Deployment.Aliyun, certificate)
		if err != nil {
			log.Fatal(err)
		}
	}
	if config.Deployment.K8s.Kube != "" {
		err := k8s.DeploymentK8s(config.Deployment.K8s, certificate)
		if err != nil {
			log.Fatal(err)
		}
	}
}
