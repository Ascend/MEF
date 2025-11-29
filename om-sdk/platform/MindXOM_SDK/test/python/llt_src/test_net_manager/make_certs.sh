#!/bin/bash

CERT_DIR="$1"
rm -r "${CERT_DIR}"
mkdir -p "${CERT_DIR}"

touch "${CERT_DIR}"/openssl.cnf
# 准备配置文件
echo """
[ v3_ca ]
subjectKeyIdentifier        = hash
authorityKeyIdentifier      = keyid:always,issuer
basicConstraints            = critical,CA:true
keyUsage                    = keyCertSign, cRLSign
extendedKeyUsage            = serverAuth, clientAuth

[ req ]
prompt             = no
default_md         = sha256
distinguished_name = dn

[ dn ]
C  = CN
ST = SiChuan
L  = ChengDu
O  = Test Company
OU = Test Department
CN = kubernetes

[ server ]
basicConstraints            = CA:false
extendedKeyUsage            = serverAuth, clientAuth
subjectKeyIdentifier        = hash
authorityKeyIdentifier      = keyid:always,issuer

[ ca ]
default_ca      = CA_default            # The default ca section

[ CA_default ]
certificate     = "${CERT_DIR}"/rootCa/cacert.pem
database        = "${CERT_DIR}"/rootCa/index.txt
crlnumber       = "${CERT_DIR}"/rootCa/crlnumber
private_key     = "${CERT_DIR}"/rootCa/cakey.pem
x509_extensions = usr_cert
default_days    = 365
default_crl_days= 30
default_md      = default
preserve        = no
""" > "${CERT_DIR}"/openssl.cnf

openssl_config_gen()
{
	mkdir -p "${CERT_DIR}"/rootCa/interCa_$1
	touch "${CERT_DIR}"/rootCa/interCa_$1/index.txt
	echo 1000 > "${CERT_DIR}"/rootCa/interCa_$1/crlnumber
	echo """
	[ v3_ca ]
	subjectKeyIdentifier        = hash
	authorityKeyIdentifier      = keyid:always,issuer
	basicConstraints            = critical,CA:true
	keyUsage                    = keyCertSign, cRLSign
	extendedKeyUsage            = serverAuth, clientAuth

	[ req ]
	prompt = no
	default_md = sha256
	distinguished_name = dn

	[ dn ]
	C = CN
	ST = SiChuan
	L = ChengDu
	O = Test Intermediate Compoany
	OU = Test Intermediate Department SUBCA $1
	CN = kubernetes

	[ server ]
	basicConstraints = CA:false
	extendedKeyUsage = serverAuth, clientAuth

	[ ca ]
	default_ca      = CA_default            # The default ca section

	[ CA_default ]
	certificate     = "${CERT_DIR}"/rootCa/interCa_$1/subcacrt.pem
	database        = "${CERT_DIR}"/rootCa/interCa_$1/index.txt
	crlnumber       = "${CERT_DIR}"/rootCa/interCa_$1/crlnumber
	private_key     = "${CERT_DIR}"/rootCa/interCa_$1/subcakey.pem
	x509_extensions = usr_cert
	default_days    = 365
	default_crl_days= 30
	default_md      = default
	preserve        = no
	""" > "${CERT_DIR}"/subca_$1_openssl.cnf
}

create_rootca_and_subca1()
{
	echo "1. generate rootCA"
	rm -rf "${CERT_DIR}"/rootCa
	mkdir "${CERT_DIR}"/rootCa
	touch "${CERT_DIR}"/rootCa/index.txt
    echo 1000 > "${CERT_DIR}"/rootCa/crlnumber
	openssl req -x509 -days 7300 -newkey rsa:3072 -keyout "${CERT_DIR}"/rootCa/cakey.pem -out "${CERT_DIR}"/rootCa/cacert.pem -nodes -config "${CERT_DIR}"/openssl.cnf -extensions v3_ca
	openssl req -x509 -days 7300 -newkey ec:<(openssl ecparam -name secp256r1) -keyout "${CERT_DIR}"/rootCa/cakey_ec_256.pem -out "${CERT_DIR}"/rootCa/cacert_ec_256.pem -nodes -config "${CERT_DIR}"/openssl.cnf -extensions v3_ca
	openssl req -x509 -days 7300 -newkey ec:<(openssl ecparam -name secp128r1) -keyout "${CERT_DIR}"/rootCa/cakey_ec_128.pem -out "${CERT_DIR}"/rootCa/cacert_ec_128.pem -nodes -config "${CERT_DIR}"/openssl.cnf -extensions v3_ca


	echo "2. generate subCA 01"
	openssl_config_gen 01
	openssl req -new -newkey rsa:3072 -keyout "${CERT_DIR}"/rootCa/interCa_01/subcakey.pem -out "${CERT_DIR}"/rootCa/interCa_01/subcacrt.csr -config "${CERT_DIR}"/subca_01_openssl.cnf -extensions v3_ca -nodes
	openssl x509 -req -in "${CERT_DIR}"/rootCa/interCa_01/subcacrt.csr -CA "${CERT_DIR}"/rootCa/cacert.pem -CAkey "${CERT_DIR}"/rootCa/cakey.pem -CAserial "${CERT_DIR}"/rootCa/serial.txt -CAcreateserial -out "${CERT_DIR}"/rootCa/interCa_01/subcacrt.pem -days 7300 -extensions v3_ca -extfile "${CERT_DIR}"/subca_01_openssl.cnf
}

sign_subCA_cert()
{
	echo "[*] signing subCA $1 certificate using $2"
	openssl_config_gen $1
	openssl req -new -newkey rsa:$3 -keyout "${CERT_DIR}"/rootCa/interCa_$1/subcakey.pem -out "${CERT_DIR}"/rootCa/interCa_$1/subcacrt.csr -config "${CERT_DIR}"/subca_$1_openssl.cnf -extensions v3_ca -nodes
	openssl x509 -req -in "${CERT_DIR}"/rootCa/interCa_$1/subcacrt.csr -CA "${CERT_DIR}"/rootCa/interCa_$2/subcacrt.pem -CAkey "${CERT_DIR}"/rootCa/interCa_$2/subcakey.pem -CAserial "${CERT_DIR}"//rootCa/serial.txt -CAcreateserial -out "${CERT_DIR}"/rootCa/interCa_$1/subcacrt.pem -days 7300 -extensions v3_ca -extfile "${CERT_DIR}"/subca_$1_openssl.cnf
}

sign_server_cert()
{
	mkdir -p "${CERT_DIR}"/rootCa/server_$1
	echo '[*] generating server certificate...'
	openssl req -new -newkey rsa:3072 -keyout "${CERT_DIR}"/rootCa/server_$1/server.key -out "${CERT_DIR}"/rootCa/server_$1/server.csr -config "${CERT_DIR}"/openssl.cnf -extensions server -nodes
	openssl x509 -req -in "${CERT_DIR}"/rootCa/server_$1/server.csr -CA "${CERT_DIR}"/rootCa/interCa_$1/subcacrt.pem -CAkey "${CERT_DIR}"/rootCa/interCa_$1/subcakey.pem -CAserial "${CERT_DIR}"/rootCa/serial.txt -CAcreateserial -out "${CERT_DIR}"/rootCa/server_$1.crt -days 365 -extensions server -extfile "${CERT_DIR}"/openssl.cnf
}

create_subCA_crl()
{
	echo "[*] Begin generating interCa crl file..."
	if [ $# -eq 2 ];
	then
		openssl ca -revoke $2 -config "${CERT_DIR}"/subca_$1_openssl.cnf
	fi
	openssl ca -gencrl -out "${CERT_DIR}"/rootCa/interCa_$1.crl -config "${CERT_DIR}"/subca_$1_openssl.cnf
}

create_rootCA_crl()
{
	echo "[*] Begin generating rootCa crl file..."
	if [ $# -eq 1 ];
	then
		openssl ca -revoke $1 -config "${CERT_DIR}"/openssl.cnf
	fi
	openssl ca -gencrl -out "${CERT_DIR}"/rootCa/rootca.crl -config "${CERT_DIR}"/openssl.cnf
}

find "${CERT_DIR}" -name index.txt | xargs rm

create_rootca_and_subca1
sign_subCA_cert 02 01 3072
sign_subCA_cert 03 02 3072
sign_subCA_cert 04 03 3072
sign_subCA_cert 05 04 2048

sign_server_cert 01
sign_server_cert 04

create_rootCA_crl
create_subCA_crl 01
create_subCA_crl 02
create_subCA_crl 03 "${CERT_DIR}"//rootCa/interCa_04/subcacrt.pem
create_subCA_crl 04
create_subCA_crl 05
