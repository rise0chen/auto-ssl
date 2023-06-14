package deployment

import (
	"lebai.ltd/auto_ssl/deployment/aliyun"
)

type DeploymentConfig struct {
	Aliyun aliyun.AliyunConfig `json:"aliyun"`
}
