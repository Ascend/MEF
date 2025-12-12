#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
##############################################################
#  The entry script of building seimanager
#
## @Filename:build_seimanager.sh
#
## @Options:
#
## @History:
#
## @Created:201909241636
##############################################################
SCRIPT_NAME=$(basename $0)
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$CUR_DIR/..

BUILD_C_SCRIPT=$CUR_DIR/build_c_package.sh
OUTPUT_INC=${TOP_DIR}/output/include
OUTPUT_PACKAGE_DIR=$TOP_DIR/output/package

LATEST_TS_VERSION="4.9"

export CFLAGS_ENV="-Wall -fstack-protector-all -fPIC -D_FORTIFY_SOURCE=2 -O2 -Wl,-z,relro -Wl,-z,now -Wl,-z,noexecstack -s"
export CXXFLAGS_ENV="-Wall -fstack-protector-all -fPIC -D_FORTIFY_SOURCE=2 -O2 -Wl,-z,relro -Wl,-z,now -Wl,-z,noexecstack -s"

uname -a | grep aarch64
ret=$?
if [ "$ret" == "0" ];then
    ARCH="aarch64"
else
    ARCH="aarch64"
    uname -a | grep x86_64
    ret=$?
    if [ "$ret" == "0" ];then
        ARCH="x86_64"
    fi
fi

ATLASEDGE_VERSION_XMLFILE=${TOP_DIR}/config/version.xml
version=$(sed "/^  MindXOM_SDK:/!d;s/.*: //" "${CUR_DIR}/../../mindxedge/build/conf/config.yaml")
sed -i "s#{Version}#${version}#g" ${ATLASEDGE_VERSION_XMLFILE}
ATLASEDGE_RELEASE="Ascend-mindxedge-omsdk_${version}_linux-${ARCH}"
ATLASEDGE_RELEASE_FILE=${TOP_DIR}/output/${ATLASEDGE_RELEASE}.tar.gz

# replace fileName
sed -i "/<Package>/,/<\/Package>/ s|<FileName>.*|<FileName>${ATLASEDGE_RELEASE}.zip</FileName>|g" ${ATLASEDGE_VERSION_XMLFILE}
sed -i "/<Package>/,/<\/Package>/ s|<Version>.*|<Version>${version}</Version>|g" ${ATLASEDGE_VERSION_XMLFILE}

printf -v indent_space %8s
MINDX_OM_SDK_COMMIT_ID=$(git rev-parse HEAD)

architecture="x86"
if [ "$ARCH" == "aarch64" ]; then
    architecture="ARM"
fi
sed -i "/<Package>/,/<\/Package>/ s|<ProcessorArchitecture>.*|<ProcessorArchitecture>$architecture</ProcessorArchitecture>|g" ${ATLASEDGE_VERSION_XMLFILE}

function check_result()
{
    local ret=$1
    local message=$2

    if [ "$ret" -eq 0 ]; then
       echo "$message success."
       return 0
    else
       echo "$message failed, code $ret"
       exit 1
    fi
}

function prepare()
{
    cd $TOP_DIR
    if [ ! -d "output/lib" ];then
        mkdir -p "output/lib"
    else
        rm -rf "output/lib/*"
    fi

    if [ ! -d "output/include" ]; then
        mkdir -p "output/include"
    else
        rm -rf output/include/*
    fi
    if [ ! -d "output/bin" ];then
        mkdir -p "output/bin"
    else
        rm -rf output/bin/*
    fi

    mkdir -p $TOP_DIR/output/opensource/python
}

function build_opensource_python()
{
    declare -a opensource_python=(
        "urllib3/src/urllib3"
        "click/src/click"
        "jinja2/src/jinja2"
        "werkzeug/src/werkzeug"
        "flask/src/flask"
        "sqlalchemy/lib/sqlalchemy"
        "websockets/src/websockets"
        "itsdangerous"
        "markupsafe"
    )

    mkdir -p "${TOP_DIR}"/opensource/python
    cd "${TOP_DIR}"/opensource || { echo "${TOP_DIR}/opensource does not exist."; return 3;}
    # flask的依赖使用pip3 install
    pip3 install -r "${CUR_DIR}"/requirements.txt --no-compile --target="${TOP_DIR}"/opensource

    # 检查当前opensource python是否都已经存在
    for ((j = 0; j < "${#opensource_python[@]}"; j++)); do
        if [[ ! -e "${opensource_python[j]}" ]]; then
            echo "check ${TOP_DIR}/opensource/${opensource_python[j]} does not exist."
            return 1
        fi
    done

    cp -arf urllib3/src/urllib3 "${TOP_DIR}"/output/opensource/python

    rm -f click/src/click/shell_completion.py
    rm -f click/src/click/testing.py
    rm -f click/src/click/_textwrap.py
    rm -f click/src/click/_winconsole.py
    cp -arf click/src/click "${TOP_DIR}"/output/opensource/python

    rm -f jinja2/src/jinja2/debug.py
    cp -arf jinja2/src/jinja2 "${TOP_DIR}"/output/opensource/python

    rm -f werkzeug/src/werkzeug/testapp.py
    rm -rf werkzeug/src/werkzeug/debug
    rm -rf werkzeug/src/werkzeug/middleware
    cp -arf werkzeug/src/werkzeug "${TOP_DIR}"/output/opensource/python

    rm -f flask/src/flask/debughelpers.py
    rm -f flask/src/flask/__main__.py
    rm -f flask/src/flask/testing.py
    rm -f flask/src/flask/views.py
    cp -arf flask/src/flask/ "${TOP_DIR}"/output/opensource/python

    rm -f sqlalchemy/lib/sqlalchemy/dialects/sqlite/provision.py
    rm -rf sqlalchemy/lib/sqlalchemy/dialects/firebird
    rm -rf sqlalchemy/lib/sqlalchemy/dialects/mssql
    rm -rf sqlalchemy/lib/sqlalchemy/dialects/mysql
    rm -rf sqlalchemy/lib/sqlalchemy/dialects/oracle
    rm -rf sqlalchemy/lib/sqlalchemy/dialects/postgresql
    rm -rf sqlalchemy/lib/sqlalchemy/dialects/sybase
    rm -rf sqlalchemy/lib/sqlalchemy/testing
    rm -rf sqlalchemy/lib/sqlalchemy/connectors
    rm -rf sqlalchemy/lib/sqlalchemy/ext/asyncio
    cp -arf sqlalchemy/lib/sqlalchemy "${TOP_DIR}"/output/opensource/python

    rm -f websockets/src/websockets/__main__.py
    cp -arf websockets/src/websockets "${TOP_DIR}"/output/opensource/python

    cp -arf itsdangerous "${TOP_DIR}"/output/opensource/python

    rm -f markupsafe/_speedups.*.so
    cp -arf markupsafe "${TOP_DIR}"/output/opensource/python
    echo "build_opensource_python done."

}

function build_opensource_c_package()
{
    OPENSOURCE_DIR=$TOP_DIR/opensource

    declare -a opensources=(openssl nginx)

    for taskname in ${opensources[@]}
    do
        mkdir -p $OPENSOURCE_DIR/${taskname}/ascend-ci/build
        cp -f ${CUR_DIR}/build_opensource/build_${taskname}.sh $OPENSOURCE_DIR/${taskname}/ascend-ci/build
        cd $OPENSOURCE_DIR/${taskname}/ascend-ci/build
        bash $OPENSOURCE_DIR/${taskname}/ascend-ci/build/build_${taskname}.sh
        if [[ $? != 0 ]];then
            echo build ${taskname} failed
            exit 1
        fi
    done

    cp $OPENSOURCE_DIR/openssl/ascend-ci/output/libcrypto.so* $TOP_DIR/output/lib -d
    cp $OPENSOURCE_DIR/openssl/ascend-ci/output/libssl.so* $TOP_DIR/output/lib -d
    cp -rf $OPENSOURCE_DIR/openssl/ascend-ci/output/include/* $OUTPUT_INC

}

function build_c_package()
{
    if [ ! -e "${BUILD_C_SCRIPT}" ]; then
        echo "build edge c failed, ${BUILD_C_SCRIPT} not found."
        exit 1
    fi

    chmod +x $BUILD_C_SCRIPT
    bash $BUILD_C_SCRIPT
    if [ $? -ne 0 ];then
        echo "build_c_package failed!"
        exit 1
    fi
    return 0
}

function package_sdk()
{
    cd "${CUR_DIR}"
    dos2unix package_sdk.sh
    chmod a+x package_sdk.sh
    ./package_sdk.sh
    return 0
}

function tar_package()
{
    cd "${OUTPUT_PACKAGE_DIR}"
    tar -czf "${ATLASEDGE_RELEASE_FILE}" *
    echo "packet MindXOM file successfully!"
    return 0
}


function web_build()
{
  local om_web_path="${TOP_DIR}"/platform/MindXOM_Web
  local package_name=mindxom-web.zip
  local nginx_path="${TOP_DIR}"/output/package/software/nginx/html/manager
  local temp_web="${TOP_DIR}"/output/web

  # 将sdk中的错误页面移动到临时目录中
  if [ ! -d "${temp_web}" ];then
      mkdir -p "${temp_web}"
  fi

  mv "${nginx_path}"/4xx.html "${temp_web}"
  mv "${nginx_path}"/50x.html "${temp_web}"

  if [ -d "${nginx_path}" ];then
      rm -r "${nginx_path}"/*
  fi

  if [ ! -d "${nginx_path}" ];then
      mkdir -p "${nginx_path}"
  fi

  if [ ! -f "${om_web_path}/${package_name}" ]; then
      echo "web build zip file not found."
      return 1
  fi

  unzip -q "${om_web_path}/${package_name}" -d "${nginx_path}/"
  mv "${nginx_path}"/dist/* "${nginx_path}"/
  mv "${temp_web}"/*.html "${nginx_path}"/
  rm -r "${nginx_path}"/dist
  rm -r "${temp_web}"
  return 0
}

function main()
{
    if [ -d "${TOP_DIR}"/output ];then
        rm -rf "${TOP_DIR}"/output/*
    fi

    declare -a build_steps=(
        prepare
        build_opensource_python
        build_opensource_c_package
        build_c_package
        package_sdk
        web_build
        tar_package)

    step_num=${#build_steps[@]}

    for ((i = 0; i < step_num; i++)); do
        local func=${build_steps[$i]}
        echo "build steps $((i + 1))/$step_num: $func"
        $func
        ret=$?
        if [ "$ret" != "0" ]; then
            echo "$func failed, ret is $ret, exit build"
            return $ret
        fi
    done

    echo "build finished"
    return 0

}

echo "begin to execute $SCRIPT_NAME"
main;ret=$?
echo "finish execute $SCRIPT_NAME, result is ${ret}!"
exit $ret
