cmake_minimum_required(VERSION 3.16)
project(MindXOM)

# build
add_custom_target(build ALL
     COMMAND fakeroot bash ${CMAKE_CURRENT_SOURCE_DIR}/build/build.sh
)