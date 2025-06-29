package vault_manager

import (
	"context"
	"errors"
)

type vaultClientMock struct{}

var _ ivault = (*vaultClientMock)(nil)

// addKVtoSecret implements ivault.
func (v *vaultClientMock) addKVtoSecret(ctx context.Context, key string, path string, data map[string]interface{}, token string) error {
	return errors.New("unimplemented")
}

// createUserPass implements ivault.
func (v *vaultClientMock) createUserPass(ctx context.Context, user string, pass string, policy string, token string) error {
	return errors.New("unimplemented")
}

// enableKV implements ivault.
func (v *vaultClientMock) enableKV(ctx context.Context, path string, type_ string, token string) error {
	return errors.New("unimplemented")
}

// enableUserPass implements ivault.
func (v *vaultClientMock) enableUserPass(ctx context.Context, type_ string, token string) error {
	return errors.New("unimplemented")
}

// init implements ivault.
func (v *vaultClientMock) init(ctx context.Context, accessKeysNum int32) (map[string]interface{}, error) {
	return nil, errors.New("unimplemented")
}

// isInitialized implements ivault.
func (v *vaultClientMock) isInitialized(ctx context.Context) (bool, error) {
	return true, errors.New("unimplemented")
}

// isSealed implements ivault.
func (v *vaultClientMock) isSealed(ctx context.Context) (bool, error) {
	return true, errors.New("unimplemented")
}

// unseal implements ivault.
func (v *vaultClientMock) unseal(ctx context.Context, keys []interface{}) error {
	return errors.New("unimplemented")
}

func (v *vaultClientMock) createPolicy(ctx context.Context, user string, policy string, token string) error {
	return errors.New("unimplemented")
}
