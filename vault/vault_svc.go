package vault_manager

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	randv2 "math/rand/v2"
	"net/url"
	"strconv"
	"strings"
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
	provisioner   *conf.Provisioner
}

func NewVaultManager(cfg *conf.Unlocker, prov *conf.Provisioner, vClient ivault, store storage.Storage) (*vaultManager, error) {
	return &vaultManager{
		ivault:        vClient,
		accessKeysNum: cfg.NumberKeys,
		storage:       store,
		provisioner:   prov,
	}, nil
}

func (v *vaultManager) Process(ctx context.Context) error {

	if err := v.unlock(ctx); err != nil {
		return err
	}

	if err := v.provisioningSecrets(ctx); err != nil {
		return err
	}

	return nil
}

func (v *vaultManager) unlock(ctx context.Context) error {

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
		if err := v.storage.InsertKeyValue("keys", "token", token); err != nil {
			return fmt.Errorf("not possible to insert key in boldtd: [%w]", err)
		}

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

		_, err = v.enableKV(ctx, kvPath, kvType, token)
		if err != nil {
			return fmt.Errorf("enable kv: (%s, %s) [%w]", kvPath, kvType, err)
		}

		err = v.addKVtoSecret(ctx, kvKey, kvPath, dataKeys, token)
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

func (v *vaultManager) provisioningSecrets(ctx context.Context) error {

	if v.provisioner == nil {
		slog.Warn("no secrets are going to be provisioned")
		return nil
	}

	token, err := v.storage.RetrieveKey("keys", "token")
	if err != nil {
		return fmt.Errorf("set secrets error: [%w]", err)
	}

	for _, mountPath := range v.provisioner.Mount {
		_, err := v.enableKV(ctx, mountPath.Path, mountPath.Type, token)

		if err != nil && !strings.Contains(err.Error(), "400 Bad") {
			slog.Error("error mount secret engine", "type", mountPath.Type, "path", mountPath.Path, "error", err)
			continue
		}

		for _, secret := range mountPath.Secrets {
			fullPath, err := url.JoinPath(secret.Path, secret.Name)
			if err != nil {
				slog.Error("error manipulating secret path", "mount", mountPath.Path, "path", secret.Path, "secret", secret.Name, "err", err)
				continue
			}

			err = v.IsKVSecretExistent(ctx, mountPath.Path, fullPath, token)
			if err == nil {
				slog.Info("secret already exists, continuing....", "mount", mountPath.Path, "secret", fullPath)
				continue
			}

			if strings.Contains(err.Error(), "404") {
				err = v.addKVtoSecret(ctx, fullPath, mountPath.Path, randomize(secret.Data, 32), token)
				if err != nil {
					slog.Error("error when adding secret", "mount", mountPath.Path, "path", secret.Path, "secret", secret.Name, "error", err)
				}
			} else {
				slog.Info("not possible to check if secret exists", "mount", mountPath, "secret", fullPath)
				return err
			}
		}
	}

	return nil
}

func randomize(m map[string]interface{}, size int) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		if str, ok := v.(string); ok && str == "*random*" {
			result[k] = generateRandomString(size)
		} else if nestedMap, ok := v.(map[string]interface{}); ok {
			result[k] = randomize(nestedMap, size)
		} else {
			result[k] = v
		}
	}
	return result
}

// Helper function to generate a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		for i := range result {
			result[i] = charset[randv2.IntN(len(charset))]
		}
		return string(result)
	}

	for i, b := range randomBytes {
		result[i] = charset[int(b)%len(charset)]
	}
	return string(result)
}
