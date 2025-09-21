package conf_test

// import (
// 	"log"
// 	"testing"
// 	"vault-unlocker/conf"

// 	"github.com/stretchr/testify/assert"
// )

// func TestEmptyConfig(t *testing.T) {
// 	data := []byte(``)
// 	c, err := conf.NewConfig(data)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, c)
// 	assert.NotNil(t, c.Storage)
// 	assert.NotNil(t, c.Storage.BoltDB)
// 	assert.NotNil(t, c.Unlocker)
// 	assert.NotNil(t, c.Unlocker)
// }

// func TestEmptyVault(t *testing.T) {
// 	data := []byte(`
// unlocker: {}
//     `)
// 	c, err := conf.NewConfig(data)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, c)
// 	assert.NotNil(t, c.Storage)
// 	assert.NotNil(t, c.Storage.BoltDB)
// 	assert.NotNil(t, c.Unlocker)
// 	assert.NotNil(t, c.Unlocker)
// }

// func TestGoodConfig(t *testing.T) {
// 	data := []byte(`
// unlocker:
//   number_keys: 3
//   url: myurl
//   `)

// 	c, err := conf.NewConfig(data)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, c.Unlocker)
// 	assert.Equal(t, 3, c.Unlocker.NumberKeys)
// 	assert.Equal(t, "myurl", c.Unlocker.Url)

// }

// func TestBadNumberKeys(t *testing.T) {
// 	scenarios := []struct {
// 		data        []byte
// 		expectedErr string
// 	}{
// 		{
// 			data: []byte(`
// unlocker:
//   number_keys: -1
// `),
// 			expectedErr: "-1",
// 		},
// 		{
// 			data: []byte(`
// unlocker:
//   number_keys: 6
// `),
// 			expectedErr: "6",
// 		},
// 	}

// 	for _, scenario := range scenarios {
// 		_, err := conf.NewConfig(scenario.data)
// 		assert.ErrorContains(t, err, scenario.expectedErr)
// 	}

// }

// func TestDefaultUnlocker(t *testing.T) {

// 	scenarios := []struct {
// 		data        []byte
// 		expectedErr string
// 	}{
// 		{
// 			data: []byte(`
// unlocker:
// `),
// 			expectedErr: "-1",
// 		},
// 		{
// 			data:        []byte(``),
// 			expectedErr: "0",
// 		},
// 	}

// 	for _, scenario := range scenarios {
// 		c, err := conf.NewConfig(scenario.data)
// 		assert.NoError(t, err)
// 		assert.NotNil(t, c.Unlocker)
// 		assert.Equal(t, 3, c.Unlocker.NumberKeys)
// 	}
// }

// func TestDefaultStorage(t *testing.T) {

// 	scenarios := []struct {
// 		data        []byte
// 		expectedErr string
// 	}{
// 		{
// 			data: []byte(`
// storage:
// `),
// 			expectedErr: "-1",
// 		},
// 		{
// 			data:        []byte(``),
// 			expectedErr: "0",
// 		},
// 	}

// 	for _, scenario := range scenarios {
// 		c, err := conf.NewConfig(scenario.data)
// 		assert.NoError(t, err)
// 		assert.NotNil(t, c.Storage)
// 		assert.Equal(t, "boltdb", c.Storage.StorageType)
// 		assert.Equal(t, "/home/vaultmanager/data/bolt.db", c.Storage.BoltDB.Path)

// 	}
// }

// func TestKubernetesConfig(t *testing.T) {

// 	scenarios := []struct {
// 		data              []byte
// 		expectedAccess    string
// 		expectedNamespace string
// 	}{
// 		{
// 			data: []byte(`
// storage:
//   type: kubernetes
//   kubernetes:
//     access: in-cluster
//     namespace: my-namespace
// `),
// 			expectedAccess:    "in-cluster",
// 			expectedNamespace: "my-namespace",
// 		},

// 		{
// 			data: []byte(`
// storage:
//   type: kubernetes
//   kubernetes:
//     access: out-cluster
// `),
// 			expectedAccess:    "out-cluster",
// 			expectedNamespace: "",
// 		},
// 		{
// 			data: []byte(`
// storage:
//   kubernetes:
// `),
// 			expectedAccess: "in-cluster",
// 		},
// 	}

// 	for _, scenario := range scenarios {
// 		_, err := conf.NewConfig(scenario.data)
// 		assert.NoError(t, err)

// 	}
// }

// func TestBoltDBConfig(t *testing.T) {

// 	scenarios := []struct {
// 		data     []byte
// 		expected string
// 	}{
// 		{
// 			data: []byte(`
// storage:
//   type: boltdb
//   boltdb:
//     path: /etc/path
// `),
// 			expected: "/etc/path",
// 		},
// 		{
// 			data: []byte(`
// storage:
//   type: boltdb
// `),
// 			expected: "/home/vaultmanager/data/bolt.db",
// 		},
// 	}

// 	for _, scenario := range scenarios {
// 		c, err := conf.NewConfig(scenario.data)
// 		assert.NoError(t, err)
// 		assert.NotNil(t, c.Storage.BoltDB)
// 		assert.Equal(t, scenario.expected, c.Storage.BoltDB.Path)
// 		assert.Equal(t, 1, len(c.Storage.BoltDB.Buckets))
// 	}
// }

// // func TestInvalidKubernetes(t *testing.T) {

// // 	scenarios := []struct {
// // 		data        []byte
// // 		expectedErr string
// // 	}{
// // 		{
// // 			data: []byte(`
// // storage:
// //   type: kubernetes
// //   kubernetes:
// //     access: oua-cluster
// // `),
// // 			expectedErr: "invalid",
// // 		},
// // 	}

// // 	for _, scenario := range scenarios {
// // 		_, err := conf.NewConfig(scenario.data)
// // 		assert.ErrorContains(t, err, scenario.expectedErr)
// // 	}
// // }

// func TestGeneratorConfig(t *testing.T) {
// 	scenarios := []struct {
// 		name                   string
// 		data                   []byte
// 		expectedLenMounts      int
// 		expectedType           string
// 		expectedPath           string
// 		expectedLenSecrets     int
// 		expectedsecretName1    string
// 		expectedSecretPAth2    string
// 		expectedLenDataSecret1 int
// 		expectedLenDataSecret2 int
// 	}{
// 		{
// 			data: []byte(`
// manager:
//   repeat_interval: 60 # seconds
//   operation_timeout: 50 # seconds

// unlocker:
//   number_keys: 3
//   request_timeout: 5
//   # url: http://localhost:8200

// encryption:
//   path: "./tests/vault/data/"

// storage:
//   type: boltdb
//   kubernetes:
//     access: out-cluster
//     namespace: monitoring
//   boltdb:
//     path: "./tests/vault/data/integration.db"

// provisioner:
//   mounts:
//   - type: kv-v2
//     path: somepath
//     secrets:
//       - path: abc/def
//         name: secret-name
//         data:
//           k1: v1
//           k2: v2
//       - path: xxx/yyy
//         name: secret-name2
//         data:
//           k1: v1
// `),
// 			name:                   "Happy Provisioner",
// 			expectedLenMounts:      1,
// 			expectedType:           "kv-v2",
// 			expectedPath:           "somepath",
// 			expectedLenSecrets:     2,
// 			expectedsecretName1:    "secret-name",
// 			expectedSecretPAth2:    "xxx/yyy",
// 			expectedLenDataSecret1: 2,
// 			expectedLenDataSecret2: 1,
// 		},
// 	}

// 	for _, scenario := range scenarios {

// 		cfg, err := conf.NewConfig(scenario.data)
// 		assert.NoError(t, err)
// 		prov := cfg.Provisioner

// 		if len(prov.Mount) != scenario.expectedLenMounts {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, len(prov.Mount), scenario.expectedLenMounts)
// 		}

// 		mnt := prov.Mount[0]
// 		if mnt.Path != scenario.expectedPath {
// 			t.Errorf("\n%s: Found: %s, Expected: %s", scenario.name, mnt.Path, scenario.expectedPath)
// 		}

// 		if mnt.Type != scenario.expectedType {
// 			t.Errorf("\n%s: Found: %s, Expected: %s", scenario.name, mnt.Type, scenario.expectedType)
// 		}

// 		if len(mnt.Secrets) != scenario.expectedLenSecrets {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, len(mnt.Secrets), scenario.expectedLenSecrets)
// 		}
// 		sec1 := mnt.Secrets[0]
// 		sec2 := mnt.Secrets[1]

// 		if sec1.Name != scenario.expectedsecretName1 {
// 			t.Errorf("\n%s: Found: %s, Expected: %s", scenario.name, sec1.Name, scenario.expectedsecretName1)
// 		}

// 		if sec2.Path != scenario.expectedSecretPAth2 {
// 			t.Errorf("\n%s: Found: %s, Expected: %s", scenario.name, sec2.Path, scenario.expectedSecretPAth2)
// 		}

// 		if len(sec1.Data) != scenario.expectedLenDataSecret1 {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, len(sec1.Data), scenario.expectedLenDataSecret1)
// 		}
// 		if len(sec2.Data) != scenario.expectedLenDataSecret2 {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, len(sec2.Data), scenario.expectedLenDataSecret2)
// 		}

// 	}
// }

// func TestAppRoleConfig(t *testing.T) {
// 	scenarios := []struct {
// 		name                                string
// 		data                                []byte
// 		expectedLenApproles                 int
// 		expectedLenPolicies                 int
// 		expectedFirstAppRoleName            string
// 		expectedFirstPolicyName             string
// 		expectedFirstAppRoleSecretTTL       int
// 		expectedFirstAppRoleTokenTTL        int
// 		expectedFirstAppRoleTokenMaxTTL     int
// 		expectedFirstAppRolenLenPolicies    int
// 		expectedFirstAppRoleExportNamespace string
// 		// expectedType           string
// 		// expectedPath           string
// 		// expectedLenSecrets     int
// 		// expectedsecretName1    string
// 		// expectedSecretPAth2    string
// 		// expectedLenDataSecret1 int
// 		// expectedLenDataSecret2 int
// 	}{
// 		{
// 			data: []byte(`
// manager:
//   repeat_interval: 60 # seconds
//   operation_timeout: 50 # seconds

// unlocker:
//   number_keys: 3
//   request_timeout: 5
//   # url: http://localhost:8200

// encryption:
//   path: "./tests/vault/data/"

// storage:
//   type: boltdb
//   kubernetes:
//     access: out-cluster
//     namespace: monitoring
//   boltdb:
//     path: "./tests/vault/data/integration.db"

// provisioner:
//   policies:
//     - name: external-secret-operator
//       rules: |
//         path "cluster/metadata/*" {
//           capabilities = ["read"," list"]
//         }
//         path "cluster/data/*" {
//           capabilities = ["read"," list"]
//         }
//     - name: external-secret-operator-2
//       rules: |
//         path "cluster/metadata/*" { capabilities = ["read"," list"] }
//         path "cluster/data/*" { capabilities = ["read"," list"] }
//   auth:
//   - type: approle
//     path: approle
//     approles:
//     - name: external-secret-operator
//       policies:
//         - external-secret-operator-policy
//         - my-sec-policy
//       secret_id_ttl: 0
//       token_ttl: 3600
//       token_max_ttl: 7200
//       export:
//         namespace: security
//     - name: external-secret-operator-2
//       policies:
//         - external-secret-operator-policy
//       secret_id_ttl: 0
//       token_ttl: 3600
//       token_max_ttl: 7200
//     - name: external-secret-operator-3
//       policies:
//         - external-secret-operator-policy
//       secret_id_ttl: 0
//       token_ttl: 3600
//       token_max_ttl: 7200
// `),
// 			name:                                "Happy Provisioner",
// 			expectedLenApproles:                 3,
// 			expectedLenPolicies:                 2,
// 			expectedFirstAppRoleName:            "external-secret-operator",
// 			expectedFirstPolicyName:             "external-secret-operator",
// 			expectedFirstAppRoleSecretTTL:       0,
// 			expectedFirstAppRoleTokenTTL:        3600,
// 			expectedFirstAppRoleTokenMaxTTL:     7200,
// 			expectedFirstAppRolenLenPolicies:    2,
// 			expectedFirstAppRoleExportNamespace: "security",
// 		},
// 	}

// 	for _, scenario := range scenarios {

// 		cfg, err := conf.NewConfig(scenario.data)
// 		assert.NoError(t, err)
// 		prov := cfg.Provisioner

// 		if len(prov.Policies) != scenario.expectedLenPolicies {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, len(prov.Policies), scenario.expectedLenPolicies)
// 		}

// 		approles := prov.Auth[0].AppRoles
// 		if len(approles) != scenario.expectedLenApproles {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, len(approles), scenario.expectedLenApproles)
// 		}

// 		appRole1 := approles[0]
// 		if appRole1.Name != scenario.expectedFirstAppRoleName {
// 			t.Errorf("\n%s: Found: %s, Expected: %s", scenario.name, appRole1.Name, scenario.expectedFirstAppRoleName)
// 		}

// 		policyName := prov.Policies[0].Name
// 		if policyName != scenario.expectedFirstPolicyName {
// 			t.Errorf("\n%s: Found: %s, Expected: %s", scenario.name, policyName, scenario.expectedFirstPolicyName)
// 		}

// 		if appRole1.SecretIdTTL != scenario.expectedFirstAppRoleSecretTTL {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, appRole1.SecretIdTTL, scenario.expectedFirstAppRoleSecretTTL)
// 		}

// 		if appRole1.TokenTTL != scenario.expectedFirstAppRoleTokenTTL {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, appRole1.TokenTTL, scenario.expectedFirstAppRoleTokenTTL)
// 		}

// 		if appRole1.TokenMaxTTL != scenario.expectedFirstAppRoleTokenMaxTTL {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, appRole1.TokenMaxTTL, scenario.expectedFirstAppRoleTokenMaxTTL)
// 		}

// 		if len(appRole1.PolicyNames) != scenario.expectedFirstAppRolenLenPolicies {
// 			t.Errorf("\n%s: Found: %d, Expected: %d", scenario.name, len(appRole1.PolicyNames), scenario.expectedFirstAppRolenLenPolicies)
// 		}

// 		if appRole1.Export.Namespace != scenario.expectedFirstAppRoleExportNamespace {
// 			t.Errorf("\n%s: Found: %s, Expected: %s", scenario.name, appRole1.Export.Namespace, scenario.expectedFirstAppRoleExportNamespace)
// 		}

// 		for _, p := range prov.Policies {
// 			assert.NotEmpty(t, p.Rules)
// 		}
// 	}
// }

// func TestExporterConfig(t *testing.T) {
// 	scenarios := []struct {
// 		data           []byte
// 		name           string
// 		expectedAccess string
// 	}{
// 		{
// 			data: []byte(`
// exporters:
//   type: kubernetes
//   kubernetes:
//     access: out-cluster
// `),
// 			name:           "Happy Exporter",
// 			expectedAccess: "out-cluster",
// 		},
// 	}

// 	for _, scenario := range scenarios {
// 		cfg, err := conf.NewConfig(scenario.data)
// 		assert.NoError(t, err)
// 		if cfg.Exporter.Kubernetes == nil {
// 			log.Fatalf("No kubernetes config found")
// 		}
// 		if cfg.Exporter.Kubernetes.Access != scenario.expectedAccess {
// 			t.Errorf("\n%s: Found: %s, Expected: %s", scenario.name, cfg.Exporter.Kubernetes.Access, scenario.expectedAccess)
// 		}
// 	}

// }
