# 1. Generate CA Private Key
openssl genrsa -out ~/.ssh/my_ca.key 4096

# 2. Generate Self-Signed CA Certificate
openssl req -x509 -new -key ~/.ssh/my_ca.key -sha256 -days 1024 \
  -out ~/.ssh/my_ca.crt -subj "/C=UA/ST=Ukraine/L=Dnepr/O=Produktor/CN=localhost"

# 3. Install CA Certificate
sudo cp ~/.ssh/my_ca.crt /usr/local/share/ca-certificates/my_ca.crt
sudo update-ca-certificates

# Generate private key for HTTPS server
openssl genrsa -out server.key 2048

# Generate CSR (Certificate Signing Request)
openssl req -new -key server.key -out server.csr -config san.conf

# Sign CSR with your CA cert
openssl x509 -req -in server.csr -CA ~/.ssh/my_ca.crt -CAkey ~/.ssh/my_ca.key -CAcreateserial \
  -out server.crt -days 365 -sha256 -extfile san.conf -extensions req_ext
