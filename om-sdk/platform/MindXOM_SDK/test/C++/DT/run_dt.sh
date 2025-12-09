#!/bin/bash
CURRENT_PATH="$(dirname $(readlink -f $0))"
CUR_PATH=$(cd "$(dirname "$0")";pwd)
TOP_DIR=$(readlink -f ${CUR_PATH}/../../..)

function main(){
    echo "Running DT for MindXOM now..."
    cd ${TOP_DIR}/test/C++/build
    chmod +x build_ut.sh
    sudo sh build_ut.sh
    if [[ $? -ne 0 ]]; then
        exit 1
    fi
    sudo mkdir -p ${TOP_DIR}/test/C++/DT/result/xmls/
    sudo cp -f ${TOP_DIR}/test/C++/outputDTCenter.xml ${TOP_DIR}/test/C++/DT/result/xmls/test_detail.xml
    echo "All DT for security-C over, now working in dir:"
    pwd
}

start=`date +%s`
main
end=`date +%s`
echo "DT running took :`expr $end - $start`" seconds