cmake_minimum_required(VERSION 3.16)
project(om_src)

# no rpath
set(CMAKE_SKIP_RPATH TRUE)
set(CMAKE_SKIP_BUILD_RPATH TRUE)
set(CMAKE_SKIP_INSTALL_RPATH TRUE)

# for security
set(CMAKE_CXX_FLAGS "-Wall -fstack-protector-all -fPIE -Wl,-z,relro,-z,now -s")
set(CMAKE_C_FLAGS "-Wall -fstack-protector-all -fPIE -Wl,-z,relro,-z,now -g -s -D_GNU_SOURCE")
set(CMAKE_EXE_LINKER_FLAGS "-Wl,-z,relro,-z,now,-z,noexecstack,-E -pie -s")

include_directories(${CMAKE_CURRENT_SOURCE_DIR}/../output/include)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../output/lib)

add_subdirectory(cpp)
