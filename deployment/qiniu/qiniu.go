package qiniu

import (
	"encoding/json"

	"lebai.ltd/auto_ssl/cert"

	"github.com/qiniu/go-sdk/v7/auth"
)

type QiniuConfig struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region"`

	Oss OssConfig `json:"oss"`
}

type QiniuDeployment struct {
	Client *auth.Credentials

	Certificate cert.Certificate
	CertId      string
}

func DeploymentQiniu(config QiniuConfig, certificate cert.Certificate) error {
	client := auth.New(config.AccessKey, config.SecretKey)
	req := map[string]string{
		"name":        certificate.Config.Domain[0],
		"common_name": certificate.Config.Email,
		"pri":         certificate.Private,
		"ca":          certificate.Public,
	}
	var response struct {
		Code   int    `json:"code"`
		Error  string `json:"error"`
		CertID string `json:"certID"`
	}
	resData, reqErr := request(client, "POST", "/sslcert", req)
	if reqErr != nil {
		return reqErr
	}
	if decodeErr := json.Unmarshal(resData, &response); decodeErr != nil {
		return decodeErr
	}

	qiniu := QiniuDeployment{
		Client: client,

		Certificate: certificate,
		CertId:      response.CertID,
	}

	if len(config.Oss.List) != 0 {
		err := DeploymentOss(qiniu, config.Oss)
		if err != nil {
			return err
		}
	}
	return nil
}
