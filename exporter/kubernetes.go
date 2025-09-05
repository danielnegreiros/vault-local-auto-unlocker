package exporter

import (
	"context"
	"log/slog"
	"path/filepath"
	"vault-unlocker/conf"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type kubernetesClient struct {
	Client     *kubernetes.Clientset
	Namespace  string
	AccessMode string
}

func NewkubernetesClient(appCfg *conf.Storage) (*kubernetesClient, error) {

	storage := &kubernetesClient{
		Namespace:  appCfg.Kubernetes.Namespace,
		AccessMode: appCfg.Kubernetes.Access,
	}

	var cfg *rest.Config
	var err error

	if storage.AccessMode == "out-cluster" {
		path := filepath.Join(homedir.HomeDir(), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			return nil, err
		}
	}

	if storage.AccessMode == "in-cluster" {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	storage.Client = client
	return storage, nil
}


// RetrieveKeys implements storage.
func (h *kubernetesClient) RetrieveKeys(ctx context.Context) error {
	secrets, err := h.Client.CoreV1().Secrets(h.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, secret := range secrets.Items {
		slog.Info("iterating", "secret", secret.Name)
	}

	return nil
}

// UpdateKeys implements storage.
func (h *kubernetesClient) UpdateKeys() {
	panic("unimplemented")
}
