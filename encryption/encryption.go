package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
)

type Crypto struct {
	path       string
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func NewCrypto(path string) (*Crypto, error) {
	c := &Crypto{path: path}

	err := c.loadOrCreateKeys()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Crypto) Encrypt(text string) (string, error) {
	if c.publicKey == nil {
		return "", errors.New("public key is not loaded")
	}
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, c.publicKey, []byte(text))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (c *Crypto) Decrypt(encoded string) (string, error) {
	if c.privateKey == nil {
		return "", errors.New("private key is not loaded")
	}
	cipherText, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, cipherText)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}

func (c *Crypto) loadOrCreateKeys() error {
	privateKeypath := filepath.Join(c.path, "private.pem")
	publicKeypath := filepath.Join(c.path, "public.pem")

	// Load existing keys
	if fileExists(privateKeypath) && fileExists(publicKeypath) {
		return c.loadKeys(privateKeypath, publicKeypath)
	}

	// Generate and save keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	publicKey := &privateKey.PublicKey

	err = savePEMKey(privateKeypath, privateKey)
	if err != nil {
		return err
	}
	err = savePublicPEMKey(publicKeypath, publicKey)
	if err != nil {
		return err
	}

	c.privateKey = privateKey
	c.publicKey = publicKey

	return nil
}

func (c *Crypto) loadKeys(privatepath, publicpath string) error {
	privData, err := os.ReadFile(privatepath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(privData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return errors.New("failed to decode private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	pubData, err := os.ReadFile(publicpath)
	if err != nil {
		return err
	}
	block, _ = pem.Decode(pubData)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return errors.New("failed to decode public key")
	}
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return err
	}

	c.privateKey = privateKey
	c.publicKey = publicKey
	return nil
}

func savePEMKey(fileName string, key *rsa.PrivateKey) error {
	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.Encode(outFile, privateKey)
}

func savePublicPEMKey(fileName string, pubkey *rsa.PublicKey) error {
	asn1Bytes := x509.MarshalPKCS1PublicKey(pubkey)
	var pemkey = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}
	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return pem.Encode(outFile, pemkey)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}
