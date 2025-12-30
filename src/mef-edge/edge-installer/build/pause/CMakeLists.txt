cmake_minimum_required (VERSION 3.16.5)
project (MEFEdge)

set(CMAKE_INCLUDE_CURRENT_DIR ON)
set(PROJECT_SRC_ROOT ${CMAKE_CURRENT_LIST_DIR}/)
set(CMAKE_SKIP_BUILD_RPATH TRUE)

aux_source_directory(${PROJECT_SRC_ROOT}/ DIR_LIB_SRCS)
add_executable (pause ${DIR_LIB_SRCS})

target_compile_options(pause  PUBLIC -fPIE -fstack-protector-all -D_FORTIFY_SOURCE=2 -O2)
target_link_options(pause  PUBLIC -Wl,-z,relro,-z,now,-z,noexecstack -pie -s)