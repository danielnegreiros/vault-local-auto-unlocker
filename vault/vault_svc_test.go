package vault_manager

import (
	"context"
	"testing"
	"vault-unlocker/conf"
	"vault-unlocker/storage"

	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	var data = []byte(`
unlocker:
  number_keys: 2
  url: http://localhost:8200
storage:
  type: boltdb
  boltdb:
    path: ../temp/bolt.db
`)

	appCfg, err := conf.NewConfig(data)
	assert.NoError(t, err)

	store, err := storage.NewBoltDBStorage(appCfg.Storage.BoltDB)
	assert.NoError(t, err)

	vmock := &vaultClientMock{}

	ctx := context.Background()
	vm, err := NewVaultManager(appCfg.Unlocker, vmock, store)
	assert.NoError(t, err)

	_, err = vm.isInitialized(ctx)
	assert.ErrorContains(t, err, "unimplemented")
}

// func TestIsInitialized(t *testing.T) {

// 	var data = []byte(`
// unlocker:
//   number_keys: 2
//   url: http://localhost:8200
// `)

// 	appCfg, err := conf.NewConfig(data)
// 	assert.NoError(t, err)

// 	ctx := context.Background()
// 	vm, err := NewVault(appCfg.Unlocker, nil)
// 	assert.NoError(t, err)

// 	res, err := vm.isInitialized(ctx)
// 	assert.NoError(t, err)
// 	assert.Equal(t, 2, vm.accessKeysNum)

// 	log.Println(res)

// }

// func TestInit(t *testing.T) {

// 	var data = []byte(`
// unlocker:
//   number_keys: 3
//   url: http://localhost:8200
// `)

// 	appCfg, err := conf.NewConfig(data)
// 	assert.NoError(t, err)

// 	ctx := context.Background()
// 	vm, err := NewVault(appCfg.Unlocker, nil)
// 	assert.NoError(t, err)

// 	resp, err := vm.init(ctx)
// 	assert.NotNil(t, resp)
// 	assert.NoError(t, err)

// }

// func TestIsSeal(t *testing.T) {

// 	var data = []byte(`
// unlocker:
//   number_keys: 3
//   url: http://localhost:8200
// `)

// 	appCfg, err := conf.NewConfig(data)
// 	assert.NoError(t, err)

// 	ctx := context.Background()
// 	vm, err := NewVault(appCfg.Unlocker, nil)
// 	assert.NoError(t, err)

// 	res, err := vm.isSealed(ctx)
// 	assert.NoError(t, err)

// 	log.Println(res)

// }

// func TestCreateUser(t *testing.T) {

// 	var data = []byte(`
// unlocker:
//   number_keys: 3
//   url: http://localhost:8200
// `)

// 	appCfg, err := conf.NewConfig(data)
// 	assert.NoError(t, err)

// 	ctx := context.Background()
// 	vm, err := NewVault(appCfg.Unlocker, nil)
// 	assert.NoError(t, err)

// 	err = vm.createUserPass(ctx, "token")
// 	assert.NoError(t, err)

// }

// func TestEnableKV(t *testing.T) {

// 	var data = []byte(`
// unlocker:
//   number_keys: 3
//   url: http://localhost:8200
// `)

// 	appCfg, err := conf.NewConfig(data)
// 	assert.NoError(t, err)

// 	ctx := context.Background()
// 	vm, err := NewVault(appCfg.Unlocker, nil)
// 	assert.NoError(t, err)

// 	err = vm.enableKV(ctx, "token")
// 	assert.NoError(t, err)

// }

// func TestAddToKV(t *testing.T) {

// 	var data = []byte(`
// unlocker:
//   number_keys: 3
//   url: http://localhost:8200
// `)

// 	appCfg, err := conf.NewConfig(data)
// 	assert.NoError(t, err)

// 	ctx := context.Background()
// 	vm, err := NewVault(appCfg.Unlocker, nil)
// 	assert.NoError(t, err)

// 	err = vm.addKVtoSecret(ctx, "token", "das", "def")
// 	assert.NoError(t, err)

// }

func TestAddPUserPolicy(t *testing.T) {

	var data = []byte(`
unlocker:
  number_keys: 2
  url: http://localhost:8200
storage:
  type: boltdb
  boltdb:
    path: ../temp/bolt.db
`)

	policy := `
path "unlocker/data/keys" {
  capabilities = [ "read", "list" ]
}
`

	appCfg, err := conf.NewConfig(data)
	assert.NoError(t, err)

	client, err := NewVaultClient(appCfg.Unlocker)
	assert.NoError(t, err)

	vm, err := NewVaultManager(appCfg.Unlocker, client, nil)
	assert.NoError(t, err)

	ctx := context.Background()
	vm.createPolicy(ctx, "unlocker", policy, "<token>")

}
