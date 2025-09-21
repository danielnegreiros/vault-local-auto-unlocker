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
	"vault-unlocker/exporter"
	"vault-unlocker/storage"
)

const (
	kvKey  = "keys"
	kvPath = "unlocker"
	kvType = "kv-v2"
)

type vaultManager struct {
	*vaultClient
	accessKeysNum int
	storage       storage.Storage
	provisioner   *conf.Provisioner
	k8sClient     *exporter.KubernetesClient
}

func NewVaultManager(cfg *conf.Unlocker, prov *conf.Provisioner, vClient *vaultClient, store storage.Storage, k8sClient *exporter.KubernetesClient) (*vaultManager, error) {
	return &vaultManager{
		vaultClient:   vClient,
		accessKeysNum: cfg.NumberKeys,
		storage:       store,
		provisioner:   prov,
		k8sClient:     k8sClient,
	}, nil
}

func (v *vaultManager) Run(ctx context.Context) error {

	dataKeys, err := v.unlock(ctx)
	if err != nil {
		return err
	}

	token, err := v.storage.RetrieveKey("keys", "token")
	if err != nil {
		return fmt.Errorf("get secrets error: [%w]", err)
	}

	if err := v.ensurePoliciesProvisioned(ctx, token); err != nil {
		return err
	}

	if err := v.ensureAuthEnabled(ctx, token); err != nil {
		return err
	}

	if err := v.ensureSecretEngineMounts(ctx, token); err != nil {
		return err
	}

	if dataKeys != nil {
		err = v.creteOrUpdateKvV2Secret(ctx, kvKey, kvPath, dataKeys, token)
		if err != nil {
			return fmt.Errorf("add kv to secret: (%s, %s) [%w]", kvKey, kvPath, err)
		}
	}

	if v.k8sClient != nil {
		for _, authMount := range v.provisioner.Auth {
			if authMount.AppRoles == nil {
				continue
			}

			switch authMount.AuthType {
			case "approle":
				err = v.exportSecretstoK8s(ctx, authMount.Path, authMount.AppRoles, token)
				if err != nil {
					slog.Warn("not possible to export secret to kubernetes", "error", err)
				}
			default:
				slog.Info("auth type not supported for export, continuing...", "type", authMount.AuthType)
				continue
			}
		}
	}

	return nil
}

func (v *vaultManager) ensureSecretEngineMounts(ctx context.Context, token string) error {
	if v.provisioner == nil || v.provisioner.Mount == nil {
		slog.Warn("no auth are going to be enabled")
		return nil
	}

	for _, mount := range v.provisioner.Mount {

		switch mount.Type {
		case "kv-v2":
			_, err := v.mountKvEnginePath(ctx, mount.Path, mount.Type, token)
			if err != nil && !strings.Contains(err.Error(), "400 Bad") {
				return fmt.Errorf("enable kv: (%s, %s) [%w]", mount.Path, mount.Type, err)
			}

			if err := v.ensureSecretsProvisioned(ctx, mount.Path, mount.Secrets, token); err != nil {
				slog.Error("Not possible to provision all secrets, continuing ...", "err", err)
			}

		default:
			slog.Warn("Secret Engine type not implemeneted or found", "type", mount.Type)
		}

	}

	return nil
}

func (v *vaultManager) ensureAuthEnabled(ctx context.Context, token string) error {
	if v.provisioner == nil || v.provisioner.Auth == nil {
		slog.Warn("no auth are going to be enabled")
		return nil
	}

	for _, auth := range v.provisioner.Auth {
		err := v.enableAuth(ctx, auth.AuthType, auth.Path, token)
		if err != nil && !strings.Contains(err.Error(), "400 Bad") {
			return fmt.Errorf("error enabling auth: [%w]", err)
		}

		switch auth.AuthType {
		case "userpass":
			if auth.Users == nil {
				slog.Info("not available user for provisioning", "type", auth.AuthType, "path", auth.Path)
				continue
			}

			for _, user := range auth.Users {
				err := v.createUserPassAuthUser(ctx, auth.Path, user.Name, user.Pass, user.Policies, token)
				if err != nil {
					slog.Warn("not possible to create user, continuing...", "user", user.Name, "type", auth.AuthType, "path", auth.Path)
				}
			}
		case "approle":
			if auth.AppRoles == nil {
				slog.Info("not available approle for provisioning", "type", auth.AuthType, "path", auth.Path)
				continue
			}
			for _, role := range auth.AppRoles {
				_, err := v.ensureAppRoleCreate(ctx, role.Name, auth.Path, role.PolicyNames, role.SecretIdTTL, token)
				if err != nil {
					slog.Warn("not possible to create approle, continuing...", "role", role.Name, "type", auth.AuthType, "path", auth.Path)
				}
			}
		}
	}

	return nil
}

func (v *vaultManager) ensurePoliciesProvisioned(ctx context.Context, token string) error {
	if v.provisioner == nil || v.provisioner.Policies == nil {
		slog.Warn("no policies are going to be provisioned")
		return nil
	}

	for _, policy := range v.provisioner.Policies {
		err := v.ensurePolicy(ctx, policy.Name, policy.Rules, token)
		if err != nil {
			return fmt.Errorf("create policy: (%s) [%w]", policy, err)
		}
	}

	return nil
}

func (v *vaultManager) unlock(ctx context.Context) (map[string]interface{}, error) {

	isInit, err := v.isInitialized(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking if vault is initialized: [%w]", err)
	}

	var dataKeys map[string]interface{}
	var token string
	var unsealKeys []interface{}

	if !isInit {
		dataKeys, err = v.init(ctx, int32(v.accessKeysNum))
		if err != nil {
			return nil, err
		}

		tmp, ok := dataKeys["root_token"]
		if !ok {
			return nil, errors.New("root_token not received")
		}

		token = tmp.(string)
		if err := v.storage.InsertKeyValue("keys", "token", token); err != nil {
			return nil, fmt.Errorf("not possible to insert key in boldtd: [%w]", err)
		}

		tmp, ok = dataKeys["keys"]
		if !ok {
			return nil, errors.New("keys not received")
		}
		unsealKeys = tmp.([]interface{})

		for i, key := range unsealKeys {
			err := v.storage.InsertKeyValue(kvKey, strconv.Itoa(i), key.(string))
			if err != nil {
				return nil, fmt.Errorf("unlock store keys: [%w]", err)
			}
		}

		err = v.unseal(ctx, unsealKeys)
		if err != nil {
			return nil, fmt.Errorf("unseal: [%w]", err)
		}

		return dataKeys, nil
	}

	sealed, err := v.isSealed(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking if vailt is unseald: [%w]", err)
	}

	if !sealed {
		slog.Info("vault is already unsealed")
		return nil, nil
	}

	if len(unsealKeys) == 0 {
		for i := range v.accessKeysNum {
			res, err := v.storage.RetrieveKey(kvKey, strconv.Itoa(i))
			if err != nil {
				return nil, fmt.Errorf("retrieve key: [%w]", err)
			}
			unsealKeys = append(unsealKeys, res)
		}
		slog.Info("keys retrieval", "operation", "completed")
	}

	err = v.unseal(ctx, unsealKeys)
	if err != nil {
		return nil, fmt.Errorf("unseal: [%w]", err)
	}

	return nil, nil
}

func (v *vaultManager) ensureSecretsProvisioned(ctx context.Context, mountPath string, secrets []conf.Secrets, token string) error {

	for _, secret := range secrets {
		secretPathName, err := url.JoinPath(secret.Path, secret.Name)
		if err != nil {
			slog.Error("error manipulating secret path", "mount", mountPath, "path", secret.Path, "secret", secret.Name, "err", err)
			continue
		}

		err = v.isKVSecretExistent(ctx, mountPath, secretPathName, token)
		if err == nil {
			slog.Info("secret already exists, continuing....", "mount", mountPath, "secret", secretPathName)
			continue
		}

		if strings.Contains(err.Error(), "404") {
			err = v.creteOrUpdateKvV2Secret(ctx, secretPathName, mountPath, randomize(secret.Data, 32), token)
			if err != nil {
				slog.Error("error when adding secret", "mount", mountPath, "path", secret.Path, "secret", secret.Name, "error", err)
			}
		} else {
			slog.Info("not possible to check if secret exists", "mount", mountPath, "secret", secretPathName)
			return err
		}
	}

	return nil
}

func (v *vaultManager) exportSecretstoK8s(ctx context.Context, path string, roles []conf.AppRole, token string) error {
	for _, role := range roles {
		if role.Export == nil || role.Export.Namespace == "" {
			continue
		}

		roleID, err := v.getAppRoleRoleID(ctx, role.Name, path, token)
		if err != nil {
			slog.Warn("not possible to get roleID for approle, continuing...", "role", role.Name, "path", path, "err", err)
			continue
		}

		slog.Info("retrieved roleID for approle", "role", role.Name, "roleID", roleID)

		slog.Info("exporting approle to kubernetes", "role", role.Name)
		secretID, err := v.generateAppRoleSecretID(ctx, role.Name, path, token)
		if err != nil {
			slog.Warn("not possible to generate secretID for approle, continuing...", "role", role.Name, "path", path, "err", err)
			continue
		}

		_, err = v.k8sClient.CreateOrUpdateSecret(ctx, role.Export.Namespace, role.Name, map[string][]byte{
			"role_id":   []byte(roleID),
			"secret_id": []byte(secretID),
		})
		if err != nil {
			slog.Warn("not possible to create or update secret in kubernetes, continuing...", "role", role.Name, "namespace", role.Export.Namespace, "err", err)
			continue
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
