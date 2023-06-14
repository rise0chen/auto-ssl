package deployment

import (
	"lebai.ltd/auto_ssl/deployment/aliyun"
	"lebai.ltd/auto_ssl/deployment/k8s"
)

type DeploymentConfig struct {
	Aliyun aliyun.AliyunConfig `json:"aliyun"`
	K8s    k8s.K8sConfig       `json:"k8s"`
}
