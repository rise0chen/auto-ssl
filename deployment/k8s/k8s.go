package k8s

import (
	"context"

	"lebai.ltd/auto_ssl/cert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sConfig struct {
	Kube string `json:"kube"`

	Secrets []SecretItem `json:"secrets"`
}
type SecretItem struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func DeploymentK8s(config K8sConfig, certificate cert.Certificate) error {
	kubeConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(config.Kube))
	if err != nil {
		return err
	}
	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	// 定义保密字典对象
	secret := corev1.Secret{
		StringData: map[string]string{
			"tls.key": certificate.Private,
			"tls.crt": certificate.Public,
		},
		Type: corev1.SecretTypeTLS,
	}
	for _, item := range config.Secrets {
		secret.ObjectMeta.Namespace = item.Namespace
		secret.ObjectMeta.Name = item.Name
		// 创建保密字典
		_, err := clientset.CoreV1().Secrets(item.Namespace).Create(context.Background(), &secret, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
