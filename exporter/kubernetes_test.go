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
	err = client.RetrieveKeys(ctx)
	assert.NoError(t, err)
}
