package qiniu

import (
	"fmt"
	"log"
)

type OssConfig struct {
	List []OssItem `json:"list"`
}
type OssItem struct {
	Domain string `json:"domain"`
}

func DeploymentOss(qiniu QiniuDeployment, config OssConfig) error {
	for _, item := range config.List {
		req := map[string]string{
			"certId":      qiniu.CertId,
			"forceHttps":  "false",
			"http2Enable": "true",
		}
		uri := fmt.Sprintf("/domain/%s/httpsconf", item.Domain)
		resData, _ := request(qiniu.Client, "PUT", uri, req)
		log.Println(string(resData))
	}
	return nil
}
