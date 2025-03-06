# TLS Certificate Generation Guide 💻🔐

This guide provides step-by-step instructions on generating TLS certificates for both a server and a client using OpenSSL. We'll ensure the certificates are production-ready and cover how to handle Subject Alternative Names (SANs) using a configuration file (`san.cnf`).

## Prerequisites 📋

Before you begin, make sure you have the following installed on your system:
- OpenSSL

## Steps 🚀

1. **Generate CA Certificate**

   ```bash
   openssl req -new -x509 -days 365 -keyout ca.key -out ca.crt -subj "/C=US/ST=California/L=Los Angeles/O=YourOrg/CN=CA"
   ```

   This command generates a self-signed CA (Certificate Authority) certificate.

2. **Generate Server Certificate**

   - Create a `san.cnf` file with the following content:

     ```conf
     [req]
     default_bits       = 2048
     distinguished_name = req_distinguished_name
     req_extensions     = req_ext
     x509_extensions    = v3_req
     prompt             = no

     [req_distinguished_name]
     countryName                = US
     stateOrProvinceName        = California
     localityName               = Los Angeles
     organizationName           = YourOrg
     typesName                 = localhost

     [req_ext]
     subjectAltName             = @alt_names

     [v3_req]
     subjectKeyIdentifier       = hash
     authorityKeyIdentifier     = keyid:always,issuer
     basicConstraints           = CA:true
     keyUsage                   = digitalSignature, keyEncipherment
     extendedKeyUsage           = serverAuth, clientAuth
     subjectAltName             = @alt_names

     [alt_names]
     DNS.1                      = localhost
     IP.1                       = 127.0.0.1
     ```

   - Generate a server key and a Certificate Signing Request (CSR):

     ```bash
     openssl req -new -newkey rsa:2048 -nodes -keyout server.key -out server.csr -subj "/C=US/ST=California/L=Los Angeles/O=YourOrg/CN=localhost" -config san.cnf
     ```

   - Generate a server certificate:

     ```bash
     openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256 -extfile san.cnf -extensions v3_req
     ```

3. **Generate Client Certificate**

   - Generate a client key and a CSR:

     ```bash
     openssl req -new -newkey rsa:2048 -nodes -keyout client.key -out client.csr -subj "/C=US/ST=California/L=Los Angeles/O=YourOrg/CN=localhost" -config san.cnf
     ```

   - Generate a client certificate:

     ```bash
     openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256 -extfile san.cnf -extensions v3_req
     ```

## Notes 📝

- The `san.cnf` file specifies the certificate details and SANs. Modify it to fit your specific requirements.
- Ensure to securely manage and store your private keys and certificates.
- These certificates are valid for 365 days (`-days 365`). Adjust the expiration period as needed.

## Conclusion 🎉

Congratulations! You've successfully generated TLS certificates using OpenSSL. These certificates are ready to be used in your production environment.

Happy coding! 😊🚀
