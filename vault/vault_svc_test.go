package vault_manager

import (
	"context"
	"log/slog"
	"reflect"
	"testing"
	"vault-unlocker/conf"
	"vault-unlocker/storage"

	"github.com/stretchr/testify/assert"
)

type VaultManagerTest struct {
	vaultManager
	token string
}

func setup(data []byte) (*vaultManager, error) {
	appCfg, err := conf.NewConfig(data)
	if err != nil {
		return nil, err
	}

	client, err := NewVaultClient(appCfg.Unlocker)
	if err != nil {
		return nil, err
	}
	store, err := storage.NewBoltDBStorage(appCfg.Storage.BoltDB)
	if err != nil {
		return nil, err
	}

	vm, err := NewVaultManager(appCfg.Unlocker, appCfg.Provisioner, client, store)
	if err != nil {
		return nil, err
	}
	return vm, nil

}

func setUpWithoutToken(data []byte) (*VaultManagerTest, error) {

	appCfg, err := conf.NewConfig(data)
	if err != nil {
		return nil, err
	}

	client, err := NewVaultClient(appCfg.Unlocker)
	if err != nil {
		return nil, err
	}
	store, err := storage.NewBoltDBStorage(appCfg.Storage.BoltDB)
	if err != nil {
		return nil, err
	}

	token, err := store.RetrieveKey("keys", "token")
	if err != nil {
		return nil, err
	}
	vm, err := NewVaultManager(appCfg.Unlocker, nil, client, store)
	if err != nil {
		return nil, err
	}
	return &VaultManagerTest{vaultManager: *vm, token: token}, nil

}

func TestAddPUserPolicy(t *testing.T) {

	var data = []byte(`
unlocker:
  number_keys: 2
  url: http://localhost:8200
storage:
  type: boltdb
  boltdb:
    path: ../tests/vault/data/integration.db
`)

	policy := `
path "unlocker/data/keys" {
  capabilities = [ "read", "list" ]
}
`

	vmt, err := setUpWithoutToken(data)
	assert.NoError(t, err)

	ctx := context.Background()
	err = vmt.createPolicy(ctx, "unlocker", policy, vmt.token)
	assert.NoError(t, err)

}

func TestSecretSecret(t *testing.T) {

	var data = []byte(`
manager:
  repeat_interval: 60 # seconds
  operation_timeout: 50 # seconds

unlocker:
  number_keys: 3
  request_timeout: 5
  # url: http://localhost:8200

encryption:
  path: "./tests/vault/data/"

storage:
  type: boltdb
  kubernetes:
    access: out-cluster
    namespace: monitoring
  boltdb:
    path: "../tests/vault/data/integration.db"

provisioner:
  mounts:
  - type: kv-v2
    path: something
    secrets:
      - path: abc/def
        name: secret-name
        data:
          k1: v1
          k2: "*random*"
      - path: xxx/yyy
        name: secret-name
        data:
          k1: v1
          k2: v2
`)

	vm, err := setup(data)
	assert.NoError(t, err)

	token, err := vm.storage.RetrieveKey("keys", "token")
	assert.NoError(t, err)

	slog.Info("making it easier for manual checking on localhost:8200", "token", token)
	ctx := context.Background()

	err = vm.provisioningSecrets(ctx)
	assert.NoError(t, err)

}

func TestSecretNoSecret(t *testing.T) {

	var data = []byte(`
manager:
  repeat_interval: 60 # seconds
  operation_timeout: 50 # seconds

unlocker:
  number_keys: 3
  request_timeout: 5
  # url: http://localhost:8200

encryption:
  path: "./tests/vault/data/"

storage:
  type: boltdb
  kubernetes:
    access: out-cluster
    namespace: monitoring
  boltdb:
    path: "../tests/vault/data/integration.db"

`)

	vm, err := setup(data)
	assert.NoError(t, err)

	token, err := vm.storage.RetrieveKey("keys", "token")
	assert.NoError(t, err)

	slog.Info("making it easier for manual checking on localhost:8200", "token", token)
	ctx := context.Background()

	err = vm.provisioningSecrets(ctx)
	assert.NoError(t, err)

}

func TestRandomize(t *testing.T) {
	// Test case 1: Empty map
	t.Run("Empty map", func(t *testing.T) {
		input := map[string]interface{}{}
		result := randomize(input, 32)
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})

	// Test case 2: Map with no "*random*" values
	t.Run("No random values", func(t *testing.T) {
		input := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}
		result := randomize(input, 32)
		if !reflect.DeepEqual(result, input) {
			t.Errorf("Expected unchanged map %v, got %v", input, result)
		}
	})

	// Test case 3: Map with "*random*" values
	t.Run("With random values", func(t *testing.T) {
		input := map[string]interface{}{
			"key1": "value1",
			"key2": "*random*",
			"key3": 42,
		}
		result := randomize(input, 32)

		// Check that the structure is preserved
		if len(result) != len(input) {
			t.Errorf("Expected map of length %d, got %d", len(input), len(result))
		}

		// Check that non-random values are unchanged
		if result["key1"] != "value1" || result["key3"] != 42 {
			t.Errorf("Non-random values were changed unexpectedly")
		}

		// Check that random value was replaced with a 16-char string
		randomVal, ok := result["key2"].(string)
		if !ok {
			t.Errorf("Expected string for key2, got %T", result["key2"])
		}
		if randomVal == "*random*" {
			t.Errorf("Value for key2 was not randomized")
		}
		if len(randomVal) != 16 {
			t.Errorf("Expected random string of length 16, got length %d", len(randomVal))
		}
	})

	// Test case 4: Nested maps
	t.Run("Nested maps", func(t *testing.T) {
		input := map[string]interface{}{
			"key1": "value1",
			"nested": map[string]interface{}{
				"nestedKey1": "*random*",
				"nestedKey2": "keep",
			},
		}
		result := randomize(input, 32)

		// Check top level
		if result["key1"] != "value1" {
			t.Errorf("Expected key1=value1, got key1=%v", result["key1"])
		}

		// Check nested map
		nestedResult, ok := result["nested"].(map[string]interface{})
		if !ok {
			t.Errorf("Expected nested map, got %T", result["nested"])
			return
		}

		// Check nested values
		if nestedResult["nestedKey2"] != "keep" {
			t.Errorf("Expected nestedKey2=keep, got nestedKey2=%v", nestedResult["nestedKey2"])
		}

		randomVal, ok := nestedResult["nestedKey1"].(string)
		if !ok {
			t.Errorf("Expected string for nestedKey1, got %T", nestedResult["nestedKey1"])
		}
		if randomVal == "*random*" {
			t.Errorf("Value for nestedKey1 was not randomized")
		}
		if len(randomVal) != 16 {
			t.Errorf("Expected random string of length 16, got length %d", len(randomVal))
		}
	})

	// Test case 5: Multiple calls should produce different random values
	t.Run("Random values differ between calls", func(t *testing.T) {
		input := map[string]interface{}{"key": "*random*"}
		result1 := randomize(input, 32)
		result2 := randomize(input, 32)

		val1, _ := result1["key"].(string)
		val2, _ := result2["key"].(string)

		if val1 == val2 {
			t.Errorf("Expected different random values, but got the same value: %s", val1)
		}
	})
}
