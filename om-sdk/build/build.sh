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

OUTPUT_DIR=$TOP_DIR/output/
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
version=$(sed "/^  MindXOM:/!d;s/.*: //" "${CUR_DIR}/../../mindxedge/build/conf/config.yaml")
sed -i "s#{Version}#${version}#g" ${ATLASEDGE_VERSION_XMLFILE}
ATLASEDGE_RELEASE="Ascend-mindxedge-om_${version}_linux-${ARCH}"
ATLASEDGE_RELEASE_FILE=${TOP_DIR}/output/${ATLASEDGE_RELEASE}.tar.gz

# replace fileName
sed -i "/<Package>/,/<\/Package>/ s|<FileName>.*|<FileName>${ATLASEDGE_RELEASE}.tar.gz</FileName>|g" ${ATLASEDGE_VERSION_XMLFILE}
sed -i "/<Package>/,/<\/Package>/ s|<Version>.*|<Version>${version}</Version>|g" ${ATLASEDGE_VERSION_XMLFILE}

printf -v indent_space %8s
MINDX_OM_COMMIT_ID=$(git rev-parse HEAD)

architecture="x86"
if [ "$ARCH" == "aarch64" ]; then
    architecture="ARM"
fi
sed -i "/<Package>/,/<\/Package>/ s|<ProcessorArchitecture>.*|<ProcessorArchitecture>$architecture</ProcessorArchitecture>|g" ${ATLASEDGE_VERSION_XMLFILE}


function prepare()
{
    local omsdk_zip_path="${TOP_DIR}"/platform/MindXOM_SDK
    local omsdk_tmp_dir="${TOP_DIR}"/output/tmp_omsdk

    mkdir -p "${omsdk_tmp_dir}"

    unzip -q "${omsdk_zip_path}"/*.zip -d "${omsdk_tmp_dir}"

    local omsdk_tag_gz=$(find "${omsdk_tmp_dir}/" -maxdepth 1 -type f | grep "tar.gz$")
    if [ ! -f "${omsdk_tag_gz}" ]; then
        echo "invalid omsdk package, tar.gz not found."
        [ -n "${omsdk_tmp_dir}" ] && rm -rf "${omsdk_tmp_dir}"
        exit 1
    else
        if [ ! -d "${OUTPUT_DIR}" ];then
            mkdir -p "${OUTPUT_DIR}"
        fi
        mv "${omsdk_tag_gz}" "${OUTPUT_DIR}"/om-sdk.tar.gz

        if [ ! -d "${OUTPUT_PACKAGE_DIR}" ];then
            mkdir -p "${OUTPUT_PACKAGE_DIR}"
        fi
    fi

    rm -rf "${omsdk_tmp_dir}"
    cp -rf "${TOP_DIR}"/config/version.xml "${OUTPUT_DIR}"
    return 0
}

function package_om()
{
    cd $CUR_DIR
    dos2unix package_om.sh
    chmod a+x package_om.sh
    echo "package_om start!"
    ./package_om.sh
    if [ $? -ne 0 ];then
        echo "package_om failed!"
        exit 1
    fi
    echo "package_om end!"
    return 0
}

function tar_package()
{
    # 将om代码打包成A500-A2-om.tar.gz
    local chmod_prepare_script="${CUR_DIR}/chmod_prepare.sh"
    dos2unix "${chmod_prepare_script}"
    chmod u+x "${chmod_prepare_script}"

    cd "${OUTPUT_PACKAGE_DIR}"
    sh "${chmod_prepare_script}" "${OUTPUT_PACKAGE_DIR}"
    if [ $? -ne 0 ];then
        echo "chmod_prepare failed!"
        exit 1
    fi

    tar -czf A500-A2-om.tar.gz *

    mv A500-A2-om.tar.gz "${OUTPUT_DIR}"
    [ -n "${OUTPUT_PACKAGE_DIR}" ] && rm -rf "${OUTPUT_PACKAGE_DIR}"

    # 将om代码和omsdk代码打包
    cd "${OUTPUT_DIR}"
    sh "${chmod_prepare_script}" "${OUTPUT_DIR}"
    if [ $? -ne 0 ];then
        echo "chmod_prepare failed!"
        exit 1
    fi
    tar -czf ${ATLASEDGE_RELEASE_FILE} *

    # 删除已经被打包的内容
    rm A500-A2-om.tar.gz
    rm om-sdk.tar.gz

    echo "packet om file successfully!"
    return 0
}

function clear_build()
{
    rm -rf ${TOP_DIR}/src/build
    return 0
}

function main()
{
    if [ -d "${TOP_DIR}"/output ];then
        rm -rf ${TOP_DIR}/output/*
    fi

    declare -a build_steps=(
        prepare
        package_om
        tar_package
        clear_build)

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
