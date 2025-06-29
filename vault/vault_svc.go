package vault_manager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"vault-unlocker/conf"
	"vault-unlocker/storage"
)

const (
	kvKey  = "keys"
	kvPath = "unlocker"
	kvType = "kv-v2"

	userPassUser   = "unlocker"
	userPassPass   = "unlocker"
	userPolicyName = "unlocker"

	authType = "userpass"
)

type vaultManager struct {
	ivault
	accessKeysNum int
	storage       storage.Storage
}

func NewVaultManager(cfg *conf.Unlocker, vClient ivault, store storage.Storage) (*vaultManager, error) {

	return &vaultManager{
		ivault:        vClient,
		accessKeysNum: cfg.NumberKeys,
		storage:       store,
	}, nil

}

func (v *vaultManager) Unlock(ctx context.Context) error {

	isInit, err := v.isInitialized(ctx)
	if err != nil {
		return fmt.Errorf("is vault unlock: [%w]", err)
	}

	var dataKeys map[string]interface{}
	var token string
	var unsealKeys []interface{}

	if !isInit {
		dataKeys, err = v.init(ctx, int32(v.accessKeysNum))
		if err != nil {
			return err
		}

		tmp, ok := dataKeys["root_token"]
		if !ok {
			return errors.New("root_token not received")
		}
		token = tmp.(string)

		tmp, ok = dataKeys["keys"]
		if !ok {
			return errors.New("keys not received")
		}
		unsealKeys = tmp.([]interface{})

		for i, key := range unsealKeys {
			err := v.storage.InsertKeyValue(kvKey, strconv.Itoa(i), key.(string))
			if err != nil {
				return fmt.Errorf("unlock store keys: [%w]", err)
			}
		}

		err = v.unseal(ctx, unsealKeys)
		if err != nil {
			return fmt.Errorf("unseal: [%w]", err)
		}

		err = v.enableUserPass(ctx, authType, token)
		if err != nil {
			return fmt.Errorf("unlock: [%w]", err)
		}

		v.enableKV(ctx, kvPath, kvType, token)
		if err != nil {
			return fmt.Errorf("enable kv: (%s, %s) [%w]", kvPath, kvType, err)
		}

		v.addKVtoSecret(ctx, kvKey, kvPath, dataKeys, token)
		if err != nil {
			return fmt.Errorf("add kv to secret: (%s, %s) [%w]", kvKey, kvPath, err)
		}

		policy := `
path "unlocker/data/*" { capabilities = [ "read", "list" ]}
path "unlocker/metadata/*" { capabilities = [ "read", "list" ]}
`
		err = v.createPolicy(ctx, userPolicyName, policy, token)
		if err != nil {
			return fmt.Errorf("create policy: (%s) [%w]", policy, err)
		}

		err = v.createUserPass(ctx, userPassUser, userPassPass, userPolicyName, token)
		if err != nil {
			return fmt.Errorf("create user pass: [%w]", err)
		}

		return nil
	}

	sealed, err := v.isSealed(ctx)
	if err != nil {
		return fmt.Errorf("unlock is unseal: [%w]", err)
	}

	if !sealed {
		slog.Info("vault unsealed")
		return nil
	}

	if len(unsealKeys) == 0 {
		for i := range v.accessKeysNum {
			res, err := v.storage.RetrieveKey(kvKey, strconv.Itoa(i))
			if err != nil {
				return fmt.Errorf("retrieve key: [%w]", err)
			}
			unsealKeys = append(unsealKeys, res)
		}
		slog.Info("keys retrieval", "operation", "completed")
	}

	err = v.unseal(ctx, unsealKeys)
	if err != nil {
		return fmt.Errorf("unseal: [%w]", err)
	}

	return nil
}
