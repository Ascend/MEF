include_directories(${CMAKE_SOURCE_DIR}/cpp/certmanage)

set(sources
  check_cert.c
  check_mount.c
  check_space.c
  fault_check.c
)

add_library(fault_check SHARED ${sources})

target_link_libraries(fault_check certmanage securec pthread common)
