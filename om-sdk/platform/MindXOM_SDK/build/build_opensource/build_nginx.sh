#!/bin/bash
# Copyright © Huawei Technologies Co., Ltd. 2020-2025. All rights reserved.
CUR_DIR=$(dirname $(readlink -f "$0"))
CI_DIR=$(readlink -f $CUR_DIR/../)
TOP_DIR=${CI_DIR}/../../../
OM_WORK_DIR=/usr/local/mindx/MindXOM

# 编译nginx
sudo mkdir -p "${OM_WORK_DIR}/software/nginx/sbin"
sudo chmod -R 755 "${OM_WORK_DIR}"
sudo chown -R slave1:slave1 "${OM_WORK_DIR}/software/nginx"

if [[ -e "${TOP_DIR}/platform/nginx" ]]; then
    mv "${TOP_DIR}/platform/nginx" "${OM_WORK_DIR}/software/nginx/sbin/"

    if [[ -d "${TOP_DIR}/opensource/nginx/conf" ]]; then
        mv "${TOP_DIR}/opensource/nginx/conf" "${OM_WORK_DIR}/software/nginx/conf"
    fi

    if [[ ! -d "${OM_WORK_DIR}/software/nginx/logs" ]]; then
        mkdir -p "${OM_WORK_DIR}/software/nginx/logs"
    fi
fi

if [ -e "${OM_WORK_DIR}"/software/nginx/sbin/nginx ]; then
    echo "nginx is already compiled."
    exit 0
fi

if [ ! -d "${TOP_DIR}"/opensource/nginx ];then
    echo "nginx package is not exist!"
    exit 1
fi

if [ -d "${CI_DIR}"/output ];then
    echo "nginx is already compiled."
    exit 0
fi

cd ${TOP_DIR}/opensource/pcre && make clean
autoreconf -ivf

cd ${TOP_DIR}/opensource/nginx && make clean
chmod +x ./auto/configure
CFLAG="-Wall -O2 -fstack-protector-strong -fPIE"
LDFLAG="-Wl,-z,relro,-z,now,-z,noexecstack -pie -s"

./auto/configure --prefix=${OM_WORK_DIR}/software/nginx --with-pcre=$TOP_DIR/opensource/pcre/ \
        --with-openssl="${TOP_DIR}"/opensource/openssl/ \
        --with-http_ssl_module \
        --with-cc-opt="$CFLAG" --with-ld-opt="$LDFLAG" --without-http_auth_basic_module \
        --without-http_autoindex_module --without-http_map_module --without-http_fastcgi_module \
        --without-http_memcached_module --without-http_empty_gif_module --without-http_upstream_ip_hash_module \
        --with-openssl-opt='-Wall -fPIC -fstack-protector-all -O2 -fomit-frame-pointer' \
        --with-pcre-opt='-Wall -fPIC -fstack-protector-all -O2 -fomit-frame-pointer' \
        --http-client-body-temp-path=${OM_WORK_DIR}/software/nginx/client_body_temp

make && make install || exit 1

if [ ! -e "${OM_WORK_DIR}"/software/nginx/sbin/nginx ];then
    echo  "Build nginx fail. nginx is not exist!"
    exit 1
fi

echo "build nginx successfully!"

exit 0