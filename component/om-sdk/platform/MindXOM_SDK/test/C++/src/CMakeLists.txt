cmake_minimum_required(VERSION 3.10)
project(MindXOM)

# GoogleTest requires at least C++17
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

set(TOP_DIR ${PROJECT_SOURCE_DIR}/../../..)
set(OUTPUT_DIR ${PROJECT_SOURCE_DIR}/../output/)
set(DEP_LIB_DIR ${PROJECT_SOURCE_DIR}/../lib/)

set(CMAKE_CXX_FLAGS "--coverage")
SET(CMAKE_C_FLAGS "--coverage")

SET(EXECUTABLE_OUTPUT_PATH ${OUTPUT_DIR})
SET(LIBRARY_OUTPUT_PATH ${DEP_LIB_DIR})

include_directories(${OUTPUT_DIR})
link_directories(${DEP_LIB_DIR})

add_subdirectory(${TOP_DIR}/src/cpp ${OUTPUT_DIR}/src)
add_subdirectory(${TOP_DIR}/opensource/googletest ${OUTPUT_DIR}/googletest)

AUX_SOURCE_DIRECTORY(${TOP_DIR}/test/C++/src sources)
ADD_EXECUTABLE(MindXOM_TEST ${sources})

target_include_directories(MindXOM_TEST
    PRIVATE
    ${TOP_DIR}/src/cpp/ens/base
    ${TOP_DIR}/src/cpp/common
    ${TOP_DIR}/src/cpp/fault_check
    ${TOP_DIR}/src/cpp/devm
    ${TOP_DIR}/src/cpp/certmanage
    ${TOP_DIR}/src/cpp/alarm_process
    ${TOP_DIR}/src/cpp/lpeblock
    ${TOP_DIR}/src/cpp/extend_alarm
    ${TOP_DIR}/test/C++/include
)

TARGET_LINK_LIBRARIES(MindXOM_TEST dl pthread ssl crypto certmanage securec base common alarm_process lpeblock gtest gtest_main)
