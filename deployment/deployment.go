package deployment

import (
	"lebai.ltd/auto_ssl/deployment/aliyun"
	"lebai.ltd/auto_ssl/deployment/k8s"
	"lebai.ltd/auto_ssl/deployment/qiniu"
)

type DeploymentConfig struct {
	Aliyun aliyun.AliyunConfig `json:"aliyun"`
	Qiniu  qiniu.QiniuConfig   `json:"qiniu"`
	K8s    k8s.K8sConfig       `json:"k8s"`
}
