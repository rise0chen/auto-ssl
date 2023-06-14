package aliyun

import (
	"fmt"

	oss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OssConfig struct {
	List []OssItem `json:"list"`
}
type OssItem struct {
	Bucket string `json:"bucket"`
	Domain string `json:"domain"`
}

func DeploymentOss(aliyun AliyunDeployment, config OssConfig) error {
	endpoint := "https://oss-" + aliyun.Region + ".aliyuncs.com"
	ossClient, err := oss.New(endpoint, aliyun.AccessKey, aliyun.SecretKey)
	if err != nil {
		return err
	}

	for _, item := range config.List {
		putCnameConfig := oss.PutBucketCname{
			Cname: item.Domain,
			CertificateConfiguration: &oss.CertificateConfiguration{
				CertId:      fmt.Sprint(aliyun.CertId),
				Certificate: aliyun.Certificate.Public,
				PrivateKey:  aliyun.Certificate.Private,
				Force:       true,
			},
		}
		err = ossClient.PutBucketCnameWithCertificate(item.Bucket, putCnameConfig)
		if err != nil {
			return err
		}
	}
	return nil
}
