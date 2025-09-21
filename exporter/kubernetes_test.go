package exporter

// import (
// 	"context"
// 	"testing"
// 	"vault-unlocker/conf"

// 	"github.com/stretchr/testify/assert"
// )

// func TestAbc(t *testing.T) {
// 	data := `
// exporters:
//   type: kubernetes
//   kubernetes:
//     access: out-cluster
//     namespace: observability
// `
// 	appCfg, err := conf.NewConfig([]byte(data))
// 	assert.NoError(t, err)

// 	client, err := NewKubernetesClient(appCfg.Exporter)
// 	assert.NoError(t, err)

// 	ctx := context.Background()
// 	_, err = client.ListSecrets(ctx, "security")
// 	assert.NoError(t, err)
// 	err = client.ReadSecret(ctx, "security", "grafana")
// 	assert.NoError(t, err)
// }

// func TestCreateSecret(t *testing.T) {
// 	data := `
// exporters:
//   type: kubernetes
//   kubernetes:
//     access: out-cluster
//     namespace: observability
// `
// 	appCfg, err := conf.NewConfig([]byte(data))
// 	assert.NoError(t, err)

// 	client, err := NewKubernetesClient(appCfg.Exporter)
// 	assert.NoError(t, err)

// 	secretData := map[string][]byte{
// 		"username": []byte("admin"),
// 		"password": []byte("s3cr3t"),
// 		"enabled":  []byte("true"),
// 		"count":    []byte("42"),
// 	}

// 	sec, err := client.CreateOrUpdateSecret(context.Background(), "security", "my-secret", secretData)
// 	assert.NoError(t, err)

// 	t.Logf("Created secret: %v", sec)

// }
