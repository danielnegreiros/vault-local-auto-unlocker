package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"vault-unlocker/conf"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type kubernetesClient struct {
	Client     *kubernetes.Clientset
	AccessMode string
}

func NewkubernetesClient(appCfg *conf.Exporter) (*kubernetesClient, error) {

	storage := &kubernetesClient{
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
			return nil, err
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
func (h *kubernetesClient) ListSecrets(ctx context.Context, namespace string) error {
	secrets, err := h.Client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, secret := range secrets.Items {
		slog.Info("iterating", "secret", secret.Name)
	}

	return nil
}

// ReadkSecret implements storage.
func (h *kubernetesClient) ReadkSecret(ctx context.Context, namespace string, name string) error {
	secret, err := h.Client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	slog.Info("iterating", "secret", secret.Name, "data", secret.Data)

	return nil
}

func (h *kubernetesClient) CreateSecret(ctx context.Context, namespace string, name string, data map[string][]byte) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	return h.Client.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
}

func convertToByteMap(input map[string]interface{}) (map[string][]byte, error) {
	output := make(map[string][]byte)

	for k, v := range input {
		switch val := v.(type) {
		case string:
			output[k] = []byte(val)
		case []byte:
			output[k] = val
		case fmt.Stringer: // things like time.Time or custom types
			output[k] = []byte(val.String())
		default:
			// try JSON marshal as a fallback
			b, err := json.Marshal(val)
			if err != nil {
				return nil, fmt.Errorf("cannot convert key %s: %w", k, err)
			}
			output[k] = b
		}
	}

	return output, nil
}
