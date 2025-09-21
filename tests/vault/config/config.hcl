ui = true
disable_mlock = "true"

storage "file" {
  path = "/vault/data/"
}

listener "tcp" {
  address = "[::]:8212"
  tls_disable = "true"
  # tls_cert_file = "/certs/server.crt"
  # tls_key_file  = "/certs/server.key"
}

api_addr = "http://localhost:8212"
cluster_addr = "http://localhost:8213"
