set(sources
  file_checker.c
  log_common.c
)

add_library(common SHARED ${sources})
target_link_libraries(common securec pthread)
