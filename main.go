package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"encoding/json"

	"github.com/go-acme/lego/v4/challenge"

	"lebai.ltd/auto_ssl/cert"
	"lebai.ltd/auto_ssl/deployment"
	"lebai.ltd/auto_ssl/deployment/aliyun"
	"lebai.ltd/auto_ssl/deployment/k8s"
	"lebai.ltd/auto_ssl/deployment/qiniu"
	"lebai.ltd/auto_ssl/dns"
)

type Config struct {
	Dns        dns.DnsConfig               `json:"dns"`
	Cert       cert.CertConfig             `json:"cert"`
	Deployment deployment.DeploymentConfig `json:"deployment"`
}

func main() {
	configStr := os.Getenv("CONFIG")
	if len(configStr) == 0 {
		data, err := ioutil.ReadFile("config.json")
		if err == nil {
			configStr = string(data)
		}
	}
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
	if strings.HasPrefix(config.Dns.Vercel.AuthToken, "***") {
		config.Dns.Vercel.AuthToken = os.Getenv("VERCEL_AUTHTOKEN" + config.Dns.Vercel.AuthToken[3:])
	}
	if strings.HasPrefix(config.Dns.Vercel.TeamID, "***") {
		config.Dns.Vercel.TeamID = os.Getenv("VERCEL_TEAMID" + config.Dns.Vercel.TeamID[3:])
	}
	if strings.HasPrefix(config.Deployment.Aliyun.AccessKey, "***") {
		config.Deployment.Aliyun.AccessKey = os.Getenv("ALIYUN_ACCESSKEY" + config.Deployment.Aliyun.AccessKey[3:])
	}
	if strings.HasPrefix(config.Deployment.Aliyun.SecretKey, "***") {
		config.Deployment.Aliyun.SecretKey = os.Getenv("ALIYUN_SECRETKET" + config.Deployment.Aliyun.SecretKey[3:])
	}
	if strings.HasPrefix(config.Deployment.Qiniu.AccessKey, "***") {
		config.Deployment.Qiniu.AccessKey = os.Getenv("QINIU_ACCESSKEY" + config.Deployment.Qiniu.AccessKey[3:])
	}
	if strings.HasPrefix(config.Deployment.Qiniu.SecretKey, "***") {
		config.Deployment.Qiniu.SecretKey = os.Getenv("QINIU_SECRETKET" + config.Deployment.Qiniu.SecretKey[3:])
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
	} else if config.Dns.Vercel.AuthToken != "" && !strings.HasPrefix(config.Dns.Vercel.AuthToken, "***") {
		dns, err := dns.NewVercelDns(config.Dns.Vercel)
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
	certificate, err := cert.GetCertificate(user, dnsProvider, config.Cert)
	if err != nil {
		log.Fatal(err)
	}

	if config.Deployment.Aliyun.AccessKey != "" {
		err := aliyun.DeploymentAliyun(config.Deployment.Aliyun, certificate)
		if err != nil {
			log.Fatal(err)
		}
	}
	if config.Deployment.Qiniu.AccessKey != "" {
		err := qiniu.DeploymentQiniu(config.Deployment.Qiniu, certificate)
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
