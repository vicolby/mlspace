package clients

import (
	"aispace/internal/config"
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KuberService struct {
	cfg       *config.Config
	clientset *kubernetes.Clientset
}

func NewKuberService(cfg *config.Config) *KuberService {
	config, _ := clientcmd.BuildConfigFromFlags("", cfg.Kuber.KubeConfigPath)
	clientset, _ := kubernetes.NewForConfig(config)

	return &KuberService{
		cfg:       cfg,
		clientset: clientset,
	}
}

func (k *KuberService) CreateNamespace(ctx context.Context, name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("project-%s", name),
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err := k.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})

	return err
}

func (k *KuberService) DeleteNamespace(ctx context.Context, name string) error {
	err := k.clientset.CoreV1().Namespaces().Delete(ctx, fmt.Sprintf("project-%s", name), metav1.DeleteOptions{})
	return err
}

func ProvideKuberService(cfg *config.Config) *KuberService {
	return NewKuberService(cfg)
}
