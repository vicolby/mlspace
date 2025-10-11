package services

import (
	"aispace/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type PVCStatusUpdate struct {
	Namespace string    `json:"namespace"`
	PVCName   string    `json:"pvcName"`
	Status    PVCStatus `json:"status"`
	Reason    string    `json:"reason,omitempty"`
}

type PVCStatus int

const (
	Pending PVCStatus = iota
	Succeeded
	Failed
	Unknown
)

func (s PVCStatus) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Succeeded:
		return "Succeeded"
	case Failed:
		return "Failed"
	case Unknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

type KuberService struct {
	cfg             *config.Config
	clientset       *kubernetes.Clientset
	informerFactory informers.SharedInformerFactory
	pvcInformer     cache.SharedIndexInformer
	pvcLister       cache.Indexer
	stopCh          chan struct{}
}

func NewKuberService(cfg *config.Config) *KuberService {
	config, _ := clientcmd.BuildConfigFromFlags("", cfg.Kuber.KubeConfigPath)
	clientset, _ := kubernetes.NewForConfig(config)

	factory := informers.NewSharedInformerFactory(clientset, time.Second*10)
	pvcInformer := factory.Core().V1().PersistentVolumeClaims().Informer()

	kService := &KuberService{
		cfg:             cfg,
		clientset:       clientset,
		informerFactory: factory,
		pvcInformer:     pvcInformer,
		pvcLister:       pvcInformer.GetIndexer(),
		stopCh:          make(chan struct{}),
	}

	kService.informerFactory.Start(kService.stopCh)
	kService.informerFactory.WaitForCacheSync(kService.stopCh)

	return kService
}

func (k *KuberService) CreateNamespace(ctx context.Context, name string, ownerEmail string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"mlspace.io/onwer-email": ownerEmail,
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err := k.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})

	return err
}

func (k *KuberService) CreatePVC(
	ctx context.Context,
	namespace string,
	pvcName string,
	size string,
	ownerEmail string,
) (*corev1.PersistentVolumeClaim, error) {
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, err
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
			Annotations: map[string]string{
				"mlspace.io/onwer-email": ownerEmail,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: quantity,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	createdPVC, err := k.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createdPVC, nil

}

func (k *KuberService) mapK8sPVCPhaseToServiceStatus(phase corev1.PersistentVolumeClaimPhase) PVCStatus {
	switch phase {
	case corev1.ClaimPending:
		return Pending
	case corev1.ClaimBound:
		return Succeeded
	case corev1.ClaimLost:
		return Failed
	default:
		return Unknown
	}
}

func (k *KuberService) GetPVCStatusFromCache(ctx context.Context, namespace, pvcName string) (PVCStatus, error) {
	key := fmt.Sprintf("%s/%s", namespace, pvcName)
	obj, exists, err := k.pvcLister.GetByKey(key)

	if err != nil {
		return Unknown, err
	}
	if !exists {
		return Unknown, fmt.Errorf("PVC %s not found in cache", key)
	}

	if pvc, ok := obj.(*corev1.PersistentVolumeClaim); ok {
		status := k.mapK8sPVCPhaseToServiceStatus(pvc.Status.Phase)
		return status, nil
	}

	return Unknown, fmt.Errorf("object is not a PersistentVolumeClaim")
}

func (k *KuberService) GetPVCStatus(ctx context.Context, namespace, pvcName string) (PVCStatus, error) {
	cachedStatus, err := k.GetPVCStatusFromCache(ctx, namespace, pvcName)

	if err != nil {
		log.Printf("Failed to get PVC status from cache: %v", err)
		return Unknown, nil
	} else if cachedStatus != Unknown {
		return cachedStatus, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 500*time.Microsecond)
	defer cancel()

	pvc, err := k.clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})

	if err != nil {
		return Failed, fmt.Errorf("failed to get PVC %s in namespace %s: %w", pvcName, namespace, err)
	}

	return k.mapK8sPVCPhaseToServiceStatus(pvc.Status.Phase), nil
}

func (k *KuberService) StopInformer() {
	close(k.stopCh)
	fmt.Println("KuberService: Informer stopped.")
}

func (k *KuberService) DeleteNamespace(ctx context.Context, name string) error {
	err := k.clientset.CoreV1().Namespaces().Delete(ctx, fmt.Sprintf("project-%s", name), metav1.DeleteOptions{})
	return err
}

func ProvideKuberService(cfg *config.Config) *KuberService {
	return NewKuberService(cfg)
}
