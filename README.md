# Vault Local Auto-Unlocker

https://hub.docker.com/r/danielnegreiros/vault-local-auto-unlocker


## 🚀 Overview

Environment: **Development only**

This tool helps streamline development workflows by automatically initializing and unsealing a local Vault instance running in "normal mode" (non-dev mode).

## ❓ Why Use It

Vault's dev mode is convenient but doesn't persist data across restarts. Reinitializing and unsealing Vault manually multiple times a day can be tedious.

This auto-unlocker solves that problem:
- Use Vault in normal mode.
- Avoid the hassle of manually storing unseal keys.
- Automatically initialize and unseal Vault on pod startup.
- Retain data across restarts.

## ⚙️ How It Works

Deploy this tool as a sidecar container alongside your Vault instance (compatible with the Vault Helm Chart). See example section [deploy](./examples/k8s/vault/README.md) [values.yaml](./examples/k8s/vault/values.yaml).
Upon pod startup, the sidecar:
- Checks if Vault is already initialized.
- If not, it initializes Vault and stores the keys securely (encrypted and local).
- Automatically unseals Vault using stored keys.
- Creates a user unlocker with password unlocker and gives it access to the root token.



## 🔐 How to Access Vault
Option 1: Vault UI
- Navigate to the Vault UI.
- Log in using method username:
  - Username: unlocker
  - Password: unlocker
- Browse to secrets engine and open unlocker

Option 2: Command Line

```bash
# Install jq if not already installed
sudo apt install jq
``` 

- Get root token and unseal keys

```bash
export VAULT_ADDR='http://<VAULT-IP:PORT>'
export VAULT_TOKEN=$(vault login -method=userpass username="unlocker" password="unlocker" -format=json | jq -r .auth.client_token)
vault kv get unlocker/keys
``` 

## 🛠 Customization Options

You're free to manage the unlocker user as you wish:
- Change the password.
- Remove the user entirely.
- Leave it as-is for convenience.

The Vault unseal keys are stored encrypted locally by the sidecar, so no need to persist or manage them manually during development.