cmake_minimum_required(VERSION 3.10)
project(MindXOM)

# for security
set(CMAKE_CXX_FLAGS "-fprofile-arcs  -ftest-coverage -std=c++14")
set(TEST_ON 1)

add_definitions(-DDT_COMPILE_GCC -D_GLIBCXX_USE_CXX11_ABI=0 -DLLT_TEST)

set(DTCENTER /opt/buildtools/dtcenter)
set(top_path ${PROJECT_SOURCE_DIR}/../../..)

set(LIBRARY_INC_PATH
    ${DTCENTER}/include
    "${top_path}/test/C++/include"
    "${top_path}/test/C++/include/include"
    "${top_path}/platform/HuaweiSecureC/include"
    "${top_path}/platform/cmscbb_csec/include"
    "${top_path}/src/de/include"
    "${top_path}/src/cpp/devm"
    "${top_path}/src/cpp/ens/base"
    "${top_path}/src/cpp/alarm_process"
    "${top_path}/src/cpp/fault_check"
    "${top_path}/src/cpp/certmanage"
    "${top_path}/src/cpp/cms_lib/src/cmscbb_common"
    "${top_path}/src/cpp/cms_lib/src/src"
    "${top_path}/src/cpp/cms_lib/src/include"
    "${top_path}/src/cpp/cms_lib/src/asn1"
    "${top_path}/src/cpp/cms_lib/src/cms"
    "${top_path}/src/cpp/cms_lib/src/pki"
    "${top_path}/src/cpp/cms_lib/src/x509"
    "${top_path}/src/cpp/cms_verify"
    "${top_path}/src/cpp/common"
    "${top_path}/output/include"
    "${top_path}/output/include/mbedTLS"
    "${top_path}/output/include/cmscbb_csec"
    "${top_path}/src/cpp/lpeblock"
)

set(LIBRARY_LIB_PATH
    ${DTCENTER}/lib
    "${top_path}/test/C++/lib"
)

message(${LIBRARY_INC_PATH})
message(${LIBRARY_LIB_PATH})

include_directories(${LIBRARY_INC_PATH})
link_directories(${LIBRARY_LIB_PATH})

AUX_SOURCE_DIRECTORY("${top_path}/test/C++/src" sources)

ADD_EXECUTABLE(MindXOM_TEST ${sources})

TARGET_LINK_LIBRARIES(MindXOM_TEST dtcenter dmtserver dl pthread ssl crypto certmanage verify kmc securec sdp c_nonshared base common alarm_process lpeblock)

SET(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/../output/)