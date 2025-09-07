package exporter

import (
	"context"
	"testing"
	"vault-unlocker/conf"

	"github.com/stretchr/testify/assert"
)

func TestAbc(t *testing.T) {
	data := `
exporters:
  type: kubernetes
  kubernetes:
    access: out-cluster
    namespace: observability
`
	appCfg, err := conf.NewConfig([]byte(data))
	assert.NoError(t, err)

	client, err := NewkubernetesClient(appCfg.Exporter)
	assert.NoError(t, err)

	ctx := context.Background()
	err = client.ListSecrets(ctx, "security")
	assert.NoError(t, err)
	err = client.ReadkSecret(ctx, "security", "grafana")
	assert.NoError(t, err)
}

func TestCreateSecret(t *testing.T) {
	data := `
exporters:
  type: kubernetes
  kubernetes:
    access: out-cluster
    namespace: observability
`
	appCfg, err := conf.NewConfig([]byte(data))
	assert.NoError(t, err)

	client, err := NewkubernetesClient(appCfg.Exporter)
	assert.NoError(t, err)

	secretData := map[string]interface{}{
		"username": "admin",
		"password": "s3cr3t",
		"enabled":  true,
		"count":    42,
	}

	converted, err := convertToByteMap(secretData)
	assert.NoError(t, err)

	sec, err := client.CreateSecret(context.Background(), "security", "my-secret", converted)
	assert.NoError(t, err)

	t.Logf("Created secret: %v", sec)

}
