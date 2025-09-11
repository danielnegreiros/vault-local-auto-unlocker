package vault_manager

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"
	"vault-unlocker/conf"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	vapi "github.com/hashicorp/vault/api"
)

type vaultClient struct {
	ep      string
	timeout int
	client  *vault.Client
}

func NewVaultClient(cfg *conf.Unlocker) (*vaultClient, error) {

	vm := &vaultClient{
		ep:      cfg.Url,
		timeout: 5,
	}

	var err error
	vm.client, err = vault.New(
		vault.WithAddress(vm.ep),
		vault.WithRequestTimeout(time.Duration(vm.timeout)*time.Second),
	)

	return vm, err
}

func (v *vaultClient) isSealed(ctx context.Context) (bool, error) {

	resp, err := v.client.System.SealStatus(ctx)
	if err != nil {
		return true, err
	}

	slog.Info("is sealed", "operation", "completed")
	return resp.Data.Sealed, nil
}

func (v *vaultClient) isInitialized(ctx context.Context) (bool, error) {

	resp, err := v.client.System.ReadInitializationStatus(ctx)
	if err != nil {
		return true, err
	}

	slog.Info("is initialized", "operation", "completed")
	return resp.Data["initialized"].(bool), nil

}

func (v *vaultClient) init(ctx context.Context, accessKeysNum int32) (map[string]interface{}, error) {
	resp, err := v.client.System.Initialize(ctx, schema.InitializeRequest{
		SecretShares:    accessKeysNum,
		SecretThreshold: accessKeysNum,
	})

	if err != nil {
		return nil, fmt.Errorf("vault init: [%w]", err)
	}

	slog.Info("initialization successfully completed")
	return resp.Data, nil

}

func (v *vaultClient) unseal(ctx context.Context, keys []interface{}) error {
	for _, k := range keys {
		_, err := v.client.System.Unseal(ctx, schema.UnsealRequest{
			Key: k.(string),
		})

		if err != nil {
			return fmt.Errorf("unseal: [%w]", err)
		}
	}

	slog.Info("unseal", "operation", "completed")
	return nil
}

func (v *vaultClient) enableAuth(ctx context.Context, engType string, mountPath string, token string) error {
	_, err := v.client.System.AuthEnableMethod(ctx, engType, schema.AuthEnableMethodRequest{Type: engType},
		vault.WithToken(token), vault.WithMountPath(mountPath))
	if err != nil {
		return fmt.Errorf("enable %s [%w]", engType, err)
	}
	slog.Info("enable auth operation completed", "type", engType, "path", mountPath)
	return nil
}

func (v *vaultClient) createUserPassAuthUser(ctx context.Context, mountPath string, user string, pass string, policies []string, token string) error {
	_, err := v.client.Auth.UserpassWriteUser(ctx, user, schema.UserpassWriteUserRequest{
		Password: pass, TokenPolicies: policies,
	}, vault.WithToken(token), vault.WithMountPath(mountPath))

	if err != nil {
		return fmt.Errorf("create userpass [%w]", err)
	}
	slog.Info("create userpass operation completed", "user", user, "path", mountPath)
	return nil
}

func (v *vaultClient) mountKvEnginePath(ctx context.Context, path string, kvType string, token string) (*vault.Response[map[string]interface{}], error) {
	resp, err := v.client.System.MountsEnableSecretsEngine(ctx, path, schema.MountsEnableSecretsEngineRequest{
		Type: kvType,
	}, vault.WithToken(token))
	if err != nil {
		return nil, fmt.Errorf("enable kv [%w]", err)
	}
	slog.Info("enable kv operation completed", "type", kvType, "mountPath", path)
	return resp, nil
}

func (v *vaultClient) creteOrUpdateKvV2Secret(ctx context.Context, secretPath string, mountPath string, data map[string]interface{}, token string) error {
	_, err := v.client.Secrets.KvV2Write(ctx, secretPath, schema.KvV2WriteRequest{
		Data: data,
	}, vault.WithToken(token),
		vault.WithMountPath(mountPath))
	if err != nil {
		return fmt.Errorf("enable kv [%w]", err)
	}
	slog.Info("add kv keys operation completed", "path", secretPath, "mountPath", mountPath)
	return nil
}

func (v *vaultClient) isKVSecretExistent(ctx context.Context, mountPath string, path string, token string) error {
	slog.Info("checking if secret is existent", "mount", mountPath, "path", path)
	_, err := v.client.Secrets.KvV2Read(ctx, path, vault.WithMountPath(mountPath), vault.WithToken(token))
	return err
}

func (v *vaultClient) ensurePolicy(ctx context.Context, policyName string, policy string, token string) error {

	_, err := v.client.System.PoliciesWriteAclPolicy(ctx, policyName, schema.PoliciesWriteAclPolicyRequest{
		Policy: policy,
	}, vault.WithToken(token))
	if err != nil {
		return fmt.Errorf("create policy: [%w]", err)
	}
	slog.Info("create policy completed", "name", policyName)

	return nil
}

func (v *vaultClient) ensureAppRoleCreate(ctx context.Context, roleName string, mountPath string, policies []string, secretIDTTl int, token string) (*vault.Response[map[string]interface{}], error) {
	res, err := v.client.Auth.AppRoleWriteRole(ctx, roleName, schema.AppRoleWriteRoleRequest{
		SecretIdTtl: strconv.Itoa(secretIDTTl),
		Policies: policies,
	}, vault.WithMountPath(mountPath), vault.WithToken(token))

	slog.Info("approle create with success", "role", roleName, "path", mountPath, "policies", policies, "secretTTL", secretIDTTl)
	return res, err
}

func (v *vaultClient) generateAppRoleSecretID(ctx context.Context, roleName string, path string, token string) (string, error) {
	// Using another vault API client sdk because vault-client-go does not support this operation yet
	// or at least it is trowing an odd erros
	slog.Info("generating secret id", "role", roleName, "path", path)

	config := vapi.DefaultConfig()
	config.Address = v.ep

	client, err := vapi.NewClient(config)
	if err != nil {
		return "", err
	}

	client.SetToken(token)

	request := fmt.Sprintf("auth/%s/role/%s/secret-id", path, roleName)
	secret, err := client.Logical().Write(request, nil)
	if err != nil {
		return "", err
	}

	// Extract SecretID and Accessor
	if secret != nil && secret.Data != nil {
		return secret.Data["secret_id"].(string), nil
	} else {
		return "", fmt.Errorf("no secret data found")
	}

}

func (v *vaultClient) getAppRoleRoleID(ctx context.Context, roleName string, path string, token string) (string, error) {
	resp, err := v.client.Auth.AppRoleReadRoleId(ctx, roleName, vault.WithMountPath(path), vault.WithToken(token))
	if err != nil {
		return "", fmt.Errorf("get role id: [%w]", err)
	}
	slog.Info("get role id with success", "role", roleName, "path", path)
	return resp.Data.RoleId, nil
}
