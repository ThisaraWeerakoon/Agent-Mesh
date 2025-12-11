#!/bin/bash

# Create certs directory
mkdir -p certs
cd certs

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=US/ST=State/L=City/O=A2A/OU=Mesh/CN=*.a2a.mesh/emailAddress=admin@a2a.mesh"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# 2. Generate Web Server's Private Key and CSR (Certificate Signing Request)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=US/ST=State/L=City/O=A2A/OU=Mesh/CN=*.sidecar.mesh/emailAddress=sidecar@a2a.mesh"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile <(printf "subjectAltName=DNS:*.sidecar.mesh,DNS:localhost,IP:0.0.0.0")

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text

# 4. Generate Client's Private Key and CSR
openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=US/ST=State/L=City/O=A2A/OU=Mesh/CN=client.sidecar.mesh/emailAddress=client@a2a.mesh"

# 5. Use CA's private key to sign client's CSR and get back the signed certificate
openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile <(printf "subjectAltName=DNS:client.sidecar.mesh,DNS:localhost,IP:0.0.0.0")

echo "Client's signed certificate"
openssl x509 -in client-cert.pem -noout -text

echo "Certificates generated in certs/"
