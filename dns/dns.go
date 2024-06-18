package dns

type DnsConfig struct {
	Aliyun AliyunConfig `json:"aliyun"`
	Vercel VercelConfig `json:"vercel"`
}
