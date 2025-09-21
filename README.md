# Vault Local Auto-Unlocker

**Docker Image**: https://hub.docker.com/r/danielnegreiros/vault-local-auto-unlocker

A comprehensive HashiCorp Vault management tool that provides automated initialization, unlocking, and provisioning capabilities. This application streamlines Vault operations by handling the complete lifecycle from initial setup to ongoing maintenance and configuration management.

![coverage](./coverage_badge.png)

## 🚀 Features

- **Auto-unlock**: Automatically unseal Vault using configurable key thresholds
- **Sync Provisioning**: Maintain consistent Vault configuration across environments
- **Policy Management**: Define and apply Vault policies declaratively
- **Auth Method Configuration**: Support for multiple authentication methods (userpass, Kubernetes, AppRole)
- **Mount Management**: Automated creation and configuration of secret engines
- **Secret Provisioning**: Bulk secret creation with support for random value generation
- **Kubernetes Integration**: Export configurations for Kubernetes environments

## 📋 Prerequisites

- HashiCorp Vault server (accessible via API)
- Go 1.19+ (for building from source)
- Appropriate network access to Vault instance
- Valid Vault permissions for management operations

## ⚙️ Configuration

The application uses a YAML configuration file to define all operational parameters:

```yaml
manager:
  repeat_interval: 60 # seconds - How often to run management cycles
  operation_timeout: 50 # seconds - Timeout for individual operations

unlocker:
  number_keys: 3 # Number of unseal keys required
  request_timeout: 5 # seconds - API request timeout
  url: http://localhost:8200 # Vault server URL

encryption:
  path: "./tests/vault/data/" # Path for storing encrypted data

storage:
  type: boltdb # Storage backend type
  boltdb:
    path: "./tests/vault/data/integration.db" # Database file path

provisioner:
  policies:
    - name: unlocker
      rules: |
        path "unlocker/data/*" { capabilities = [ "read", "list" ]}
        path "unlocker/metadata/*" { capabilities = [ "read", "list" ]}
    - name: external-secret-operator
      rules: |
        path "cluster/metadata/*" { capabilities = ["read","list"] }
        path "cluster/data/*" { capabilities = ["read","list"] }
  
  auth:
    - type: userpass
      path: userpass
      users:
        - name: unlocker
          pass: unlocker
          policies:
            - unlocker
    
    - type: kubernetes
      path: kubernetes
    
    - type: approle
      path: approle
      approles:
      - name: external-secret-operator
        policies:
          - external-secret-operator-policy
        secret_id_ttl: 3600
        token_ttl: 3600
        token_max_ttl: 7200
        export:
          namespace: security
  
  mounts:
  - type: kv-v2
    path: unlocker
  - type: kv-v2
    path: cluster
    secrets:
      - path: smoke/test
        name: secret-name
        data:
          k1: v1
          k2: "*random*" # Generates random value

exporter:
  kubernetes:
    access: out-cluster # Access mode for Kubernetes integration
```

## 🏃 Usage

### Deployment as Vault Sidecar

This application is designed to be deployed as a sidecar container alongside HashiCorp Vault in Kubernetes environments. The sidecar pattern ensures the manager runs in the same pod as Vault, providing seamless access and management.

#### Kubernetes Deployment Example

```yaml
vault:
  server:
    extraContainers:
      - name: vault-local-auto-unlocker
        image: danielnegreiros/vault-local-auto-unlocker:latest
        volumeMounts:
          - mountPath: /home/vaultmanager
            name: home-volume
    volumes:
      - name: home-volume
        persistentVolumeClaim:
          claimName: home-pvc
```

#### Setup Steps

1. **Prepare Configuration**: Create your configuration file and store it in a ConfigMap or persistent volume

2. **Deploy as Sidecar**: Add the sidecar container to your Vault deployment configuration

3. **Volume Mounts**: Ensure proper volume mounts for:
   - Configuration files
   - Encryption data storage
   - BoltDB database persistence

4. **Monitor Operations**: The sidecar will automatically start managing Vault operations based on the configured `repeat_interval`

### Key Operations

#### Vault Initialization
The application automatically detects uninitialized Vault instances and performs initial setup with the configured parameters.

#### Auto-unlocking
Continuously monitors Vault seal status and automatically unseals using the configured number of keys when needed.

#### Provisioning Sync
Ensures Vault configuration matches the desired state defined in the configuration file:
- Creates and updates policies
- Configures authentication methods
- Manages secret engine mounts
- Provisions initial secrets

## 📁 Configuration Sections

### Manager Settings
- `repeat_interval`: Frequency of management cycles
- `operation_timeout`: Maximum time for individual operations

### Unlocker Configuration
- `number_keys`: Unseal key threshold
- `request_timeout`: API request timeout
- `url`: Vault server endpoint

### Encryption Settings
- `path`: Directory for storing encrypted operational data

### Storage Backend
Currently supports BoltDB for local storage of application state.

### Provisioner Configuration
Defines the desired Vault state:
- **Policies**: Vault policy definitions with HCL rules
- **Auth Methods**: Authentication backend configuration
- **Mounts**: Secret engine mounts and initial secrets
- **Special Values**: Use `*random*` for auto-generated values

### Exporter Settings
- **Kubernetes**: Configure integration with Kubernetes clusters

## 🔐 Security Considerations

- Store configuration files securely with appropriate file permissions
- Use strong passwords for userpass authentication
- Regularly rotate AppRole credentials
- Monitor application logs for security events
- Ensure network security between the application and Vault


**Note**: Always test configuration changes in a non-production environment before applying to production Vault instances.