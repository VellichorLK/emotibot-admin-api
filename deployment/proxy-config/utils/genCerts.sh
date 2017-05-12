#!/bin/bash

CERT_HOST=$1
if [ -z $CERT_HOST ]; then
    echo "# The script will generate a self signed cert for <hostname>.emotibot.com"
    echo "# Usage: $0 <hostname>"
    exit 1
fi

echo "Let's generate Self Signed Certificates for $CERT_HOST"

echo "
Country Name (2 letter code) [AU]:CN
State or Province Name (full name) [Some-State]:SH
Locality Name (eg, city) []:Shanghai
Organization Name (eg, company) [Internet Widgits Pty Ltd]:Emotibot
Organizational Unit Name (eg, section) []:Devops
Common Name (e.g. server FQDN or YOUR name) []:$CERT_HOST.emotibot.com
Email Address []:deployer@.emotibot.com"

rm -rf certs
mkdir -p certs
cd certs
openssl genrsa -out $CERT_HOST.key 2048
openssl req -new -key $CERT_HOST.key -out $CERT_HOST.csr
openssl x509 -req -days 3650 -in $CERT_HOST.csr -signkey $CERT_HOST.key -out $CERT_HOST.crt
openssl dhparam -out dhparam.pem 2048
cat $CERT_HOST.crt >> $CERT_HOST.emotibot.com.pem
cat $CERT_HOST.key >> $CERT_HOST.emotibot.com.pem
chmod 400 *

