package vault_manager

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vault-unlocker/conf"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

type ivault interface {
	isSealed(ctx context.Context) (bool, error)
	isInitialized(ctx context.Context) (bool, error)
	init(ctx context.Context, accessKeysNum int32) (map[string]interface{}, error)
	unseal(ctx context.Context, keys []interface{}) error
	enableUserPass(ctx context.Context, type_ string, token string) error
	createUserPass(ctx context.Context, user string, pass string, policy string, token string) error
	enableKV(ctx context.Context, path string, type_ string, token string) error
	addKVtoSecret(ctx context.Context, key string, path string, data map[string]interface{}, token string) error
	createPolicy(ctx context.Context, user string, policy string, token string) error
}

type vaultClient struct {
	ep      string
	timeout int
	client  *vault.Client
}

var _ ivault = (*vaultClient)(nil)

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

func (v *vaultClient) enableUserPass(ctx context.Context, type_ string, token string) error {
	_, err := v.client.System.AuthEnableMethod(ctx, type_, schema.AuthEnableMethodRequest{Type: type_}, vault.WithToken(token))
	if err != nil {
		return fmt.Errorf("enable userpass [%w]", err)
	}
	slog.Info("enable userpass", "operation", "completed")
	return nil
}

func (v *vaultClient) createUserPass(ctx context.Context, user string, pass string, policy string, token string) error {
	_, err := v.client.Auth.UserpassWriteUser(ctx, user, schema.UserpassWriteUserRequest{
		Password: pass, TokenPolicies: []string{policy},
	}, vault.WithToken(token))

	if err != nil {
		return fmt.Errorf("create userpass [%w]", err)
	}
	slog.Info("create userpass", "operation", "completed")
	return nil
}

func (v *vaultClient) enableKV(ctx context.Context, path string, type_ string, token string) error {
	_, err := v.client.System.MountsEnableSecretsEngine(ctx, path, schema.MountsEnableSecretsEngineRequest{
		Type: type_,
	}, vault.WithToken(token))
	if err != nil {
		return fmt.Errorf("enable kv [%w]", err)
	}
	slog.Info("enable kv", "operation", "completed")
	return nil
}

func (v *vaultClient) addKVtoSecret(ctx context.Context, key string, path string, data map[string]interface{}, token string) error {
	_, err := v.client.Secrets.KvV2Write(ctx, key, schema.KvV2WriteRequest{
		Data: data,
	}, vault.WithToken(token),
		vault.WithMountPath(path))
	if err != nil {
		return fmt.Errorf("enable kv [%w]", err)
	}
	slog.Info("add kv keys", "operation", "completed")
	return nil
}

func (v *vaultClient) createPolicy(ctx context.Context, name string, policy string, token string) error {

	_, err := v.client.System.PoliciesWriteAclPolicy(ctx, name, schema.PoliciesWriteAclPolicyRequest{
		Policy: policy,
	}, vault.WithToken(token))
	if err != nil {
		return fmt.Errorf("create policy: [%w]", err)
	}
	slog.Info("create policy", "operation", "completed")

	return nil
}
