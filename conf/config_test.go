package conf_test

import (
	"testing"
	"vault-unlocker/conf"

	"github.com/stretchr/testify/assert"
)

func TestEmptyConfig(t *testing.T) {
	data := []byte(``)
	c, err := conf.NewConfig(data)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.Storage)
	assert.NotNil(t, c.Storage.BoltDB)
	assert.NotNil(t, c.Unlocker)
	assert.NotNil(t, c.Unlocker)
}

func TestEmptyVault(t *testing.T) {
	data := []byte(`
unlocker: {}
    `)
	c, err := conf.NewConfig(data)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.Storage)
	assert.NotNil(t, c.Storage.BoltDB)
	assert.NotNil(t, c.Unlocker)
	assert.NotNil(t, c.Unlocker)
}

func TestGoodConfig(t *testing.T) {
	data := []byte(`
unlocker:
  number_keys: 3
  url: myurl
  `)

	c, err := conf.NewConfig(data)
	assert.NoError(t, err)
	assert.NotNil(t, c.Unlocker)
	assert.Equal(t, 3, c.Unlocker.NumberKeys)
	assert.Equal(t, "myurl", c.Unlocker.Url)

}

func TestBadNumberKeys(t *testing.T) {
	scenarios := []struct {
		data        []byte
		expectedErr string
	}{
		{
			data: []byte(`
unlocker:
  number_keys: -1
`),
			expectedErr: "-1",
		},
		{
			data: []byte(`
unlocker:
  number_keys: 6
`),
			expectedErr: "6",
		},
	}

	for _, scenario := range scenarios {
		_, err := conf.NewConfig(scenario.data)
		assert.ErrorContains(t, err, scenario.expectedErr)
	}

}

func TestDefaultUnlocker(t *testing.T) {

	scenarios := []struct {
		data        []byte
		expectedErr string
	}{
		{
			data: []byte(`
unlocker:
`),
			expectedErr: "-1",
		},
		{
			data:        []byte(``),
			expectedErr: "0",
		},
	}

	for _, scenario := range scenarios {
		c, err := conf.NewConfig(scenario.data)
		assert.NoError(t, err)
		assert.NotNil(t, c.Unlocker)
		assert.Equal(t, 3, c.Unlocker.NumberKeys)
	}
}

func TestDefaultStorage(t *testing.T) {

	scenarios := []struct {
		data        []byte
		expectedErr string
	}{
		{
			data: []byte(`
storage:
`),
			expectedErr: "-1",
		},
		{
			data:        []byte(``),
			expectedErr: "0",
		},
	}

	for _, scenario := range scenarios {
		c, err := conf.NewConfig(scenario.data)
		assert.NoError(t, err)
		assert.NotNil(t, c.Storage)
		assert.Equal(t, "boltdb", c.Storage.StorageType)
		assert.Equal(t, "/vault/data/bolt.db", c.Storage.BoltDB.Path)

	}
}

func TestKubernetesConfig(t *testing.T) {

	scenarios := []struct {
		data              []byte
		expectedAccess    string
		expectedNamespace string
	}{
		{
			data: []byte(`
storage:
  type: kubernetes
  kubernetes:
    access: in-cluster
    namespace: my-namespace
`),
			expectedAccess:    "in-cluster",
			expectedNamespace: "my-namespace",
		},

		{
			data: []byte(`
storage:
  type: kubernetes
  kubernetes:
    access: out-cluster
`),
			expectedAccess:    "out-cluster",
			expectedNamespace: "",
		},
		{
			data: []byte(`
storage:
  kubernetes:
`),
			expectedAccess: "in-cluster",
		},
	}

	for _, scenario := range scenarios {
		c, err := conf.NewConfig(scenario.data)
		assert.NoError(t, err)
		assert.NotNil(t, c.Storage.Kubernetes)
		assert.Equal(t, scenario.expectedAccess, c.Storage.Kubernetes.Access)
		assert.Equal(t, scenario.expectedNamespace, c.Storage.Kubernetes.Namespace)

	}
}

func TestBoltDBConfig(t *testing.T) {

	scenarios := []struct {
		data     []byte
		expected string
	}{
		{
			data: []byte(`
storage:
  type: boltdb
  boltdb:
    path: /etc/path
`),
			expected: "/etc/path",
		},
		{
			data: []byte(`
storage:
  type: boltdb
`),
			expected: "/vault/data/bolt.db",
		},
	}

	for _, scenario := range scenarios {
		c, err := conf.NewConfig(scenario.data)
		assert.NoError(t, err)
		assert.NotNil(t, c.Storage.BoltDB)
		assert.Equal(t, scenario.expected, c.Storage.BoltDB.Path)
		assert.Equal(t, 1, len(c.Storage.BoltDB.Buckets))
	}
}

func TestInvalidKubernetes(t *testing.T) {

	scenarios := []struct {
		data        []byte
		expectedErr string
	}{
		{
			data: []byte(`
storage:
  type: kubernetes
  kubernetes:
    access: oua-cluster
`),
			expectedErr: "invalid",
		},
	}

	for _, scenario := range scenarios {
		_, err := conf.NewConfig(scenario.data)
		assert.ErrorContains(t, err, scenario.expectedErr)
	}
}
