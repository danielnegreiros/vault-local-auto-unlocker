package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	// storage
	defaultStorageType = "boltdb"
	// boldtb
	defaultBotlDBPath = "/home/vaultmanager/data/bolt.db"
	// kubernetes
	defaultAccessKeysMode = "in-cluster"
	// vault
	defaultAccessKeysNumber = 3
	defaultVaultUrl         = "http://localhost:8200"
	// encryption
	defaultEncryptionPath = "/home/vaultmanager/data/encryption/"
)

type config struct {
	Unlocker   *Unlocker   `yaml:"unlocker"`
	Encryption *Encryption `yaml:"encryption"`
	Storage    *Storage    `yaml:"storage"`
}

type Unlocker struct {
	NumberKeys int    `yaml:"number_keys"`
	Url        string `yaml:"url"`
}

type Encryption struct {
	Path string `yaml:"path"`
}

type Storage struct {
	StorageType string      `yaml:"type"`
	Kubernetes  *Kubernetes `yaml:"kubernetes"`
	BoltDB      *BoltBD     `yaml:"boltdb"`
}

type Kubernetes struct {
	Access    string `yaml:"access"`
	Namespace string `yaml:"namespace"`
}

type BoltBD struct {
	Path    string `yaml:"path"`
	Buckets []string
}

func NewConfig(content []byte) (*config, error) {

	c := &config{}
	err := yaml.Unmarshal(content, c)
	if err != nil {
		return nil, err
	}

	if c.Unlocker == nil {
		c.Unlocker = getDefaultUnlocker()
	}

	if c.Storage == nil {
		c.Storage = getDefaultStorage()
	}

	if c.Encryption == nil {
		c.Encryption = getDefaultEncryption()
	}

	return c, nil

}

func (c *config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = config{}
	type plain config
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}

	return nil
}

func (u *Unlocker) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*u = Unlocker{}
	type plain Unlocker
	err := unmarshal((*plain)(u))
	if err != nil {
		return err
	}

	if u.NumberKeys < 0 || u.NumberKeys > 5 {
		return fmt.Errorf("invalid number of unlock keys: %d", u.NumberKeys)
	}

	if u.Url == "" {
		u.Url = defaultVaultUrl
	}

	if u.NumberKeys == 0 {
		u.NumberKeys = defaultAccessKeysNumber
	}

	return nil
}

func (e *Encryption) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*e = Encryption{}
	type plain Encryption
	err := unmarshal((*plain)(e))
	if err != nil {
		return err
	}

	if e.Path == "" {
		e.Path = defaultEncryptionPath
	}

	return nil
}

func (s *Storage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*s = Storage{}
	type plain Storage
	err := unmarshal((*plain)(s))
	if err != nil {
		return err
	}

	if s.Kubernetes == nil {
		s.Kubernetes = getDefaultKubernetes()
	}

	if s.BoltDB == nil {
		s.BoltDB = getDefaultBoltDB()
	}

	if s.StorageType == "" {
		s.StorageType = defaultStorageType
	}

	if s.StorageType != "kubernetes" && s.StorageType != "boltdb" {
		return fmt.Errorf("invalid storage type :%s", s.StorageType)
	}

	return nil
}

func (k *Kubernetes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*k = Kubernetes{}
	type plain Kubernetes
	err := unmarshal((*plain)(k))
	if err != nil {
		return err
	}

	if k.Access == "" {
		k.Access = defaultAccessKeysMode
		return nil
	}

	if k.Access != "in-cluster" && k.Access != "out-cluster" {
		return fmt.Errorf("kubernetes configuration invalid, choose one of [in-cluster, out-cluster]. option=%v", k.Access)
	}

	return nil
}

func (b *BoltBD) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*b = BoltBD{}
	type plain BoltBD
	err := unmarshal((*plain)(b))
	if err != nil {
		return err
	}

	if b.Path == "" {
		b.Path = defaultBotlDBPath
	}

	b.Buckets = []string{"keys"}

	return nil
}

func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func getDefaultUnlocker() *Unlocker {
	return &Unlocker{
		NumberKeys: defaultAccessKeysNumber,
		Url:        defaultVaultUrl,
	}
}

func getDefaultStorage() *Storage {
	return &Storage{
		StorageType: defaultStorageType,
		BoltDB:      getDefaultBoltDB(),
	}
}

func getDefaultKubernetes() *Kubernetes {
	return &Kubernetes{
		Access: defaultAccessKeysMode,
	}
}

func getDefaultBoltDB() *BoltBD {
	return &BoltBD{
		Path:    defaultBotlDBPath,
		Buckets: []string{"keys"},
	}
}

func getDefaultEncryption() *Encryption {
	return &Encryption{
		Path: defaultEncryptionPath,
	}
}
