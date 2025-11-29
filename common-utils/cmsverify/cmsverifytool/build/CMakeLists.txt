# CMake 最低版本号要求
cmake_minimum_required (VERSION 3.16)

set(CMAKE_C_FLAGS "-Wall -fstack-protector-strong -fPIC -D_FORTIFY_SOURCE=2 -O2 -Wl,-z,relro -Wl,-z,now -Wl,-z,noexecstack -s")
add_definitions(-DCMSCBB_SUPPORT_RSAPSS)
# 项目信息
project (mef)

set(CMAKE_INCLUDE_CURRENT_DIR ON)
set(MODULE_SRC_ROOT ${CMAKE_CURRENT_LIST_DIR}/..)
set(PROJECT_SRC_ROOT ${MODULE_SRC_ROOT}/../..)
set(CMAKE_SKIP_BUILD_RPATH TRUE)

set(CMAKE_SKIP_RPATH TRUE)

set(TOP_DIR_INCLUDE ${PROJECT_SRC_ROOT}/output/include)

include_directories(${TOP_DIR_INCLUDE})
MESSAGE(STATUS "TOP_DIR result: " ${TOP_DIR})

# 添加 CMSCBB 平台子目录
set(Cmscbb_PATH ${PROJECT_SRC_ROOT}/platform/cmscbb_csec)
include_directories(${Cmscbb_PATH}/include)
MESSAGE(STATUS "Cmscbb_PATH result: " ${Cmscbb_PATH})


include_directories(${MODULE_SRC_ROOT}/include)

aux_source_directory(${MODULE_SRC_ROOT}/src DIR_LIB_SRCS)
aux_source_directory(${Cmscbb_PATH}/src DIR_LIB_SRCS)

add_library(cms_verify SHARED ${DIR_LIB_SRCS})