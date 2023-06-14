package aliyun

import (
	"time"

	"lebai.ltd/auto_ssl/cert"

	cas "github.com/alibabacloud-go/cas-20200407/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
)

type AliyunConfig struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region"`

	Oss OssConfig `json:"oss"`
}

type AliyunDeployment struct {
	AccessKey   string
	SecretKey   string
	Region      string
	Certificate cert.Certificate
	CertId      int64
}

func DeploymentAliyun(config AliyunConfig, certificate cert.Certificate) error {
	aliConfig := openapi.Config{AccessKeyId: &config.AccessKey, AccessKeySecret: &config.SecretKey, RegionId: &config.Region}
	casClient, err := cas.NewClient(&aliConfig)
	if err != nil {
		return err
	}
	now := time.Now().String()
	resp, err := casClient.UploadUserCertificate(&cas.UploadUserCertificateRequest{Name: &now, Cert: &certificate.Public, Key: &certificate.Private})
	if err != nil {
		return err
	}
	aliyun := AliyunDeployment{
		AccessKey:   config.AccessKey,
		SecretKey:   config.SecretKey,
		Region:      config.Region,
		Certificate: certificate,
		CertId:      *resp.Body.CertId,
	}

	if len(config.Oss.List) != 0 {
		err := DeploymentOss(aliyun, config.Oss)
		if err != nil {
			return err
		}
	}
	return nil
}
