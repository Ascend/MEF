set(sources
  ensd.c
)

add_subdirectory(base)

add_executable(ensd ${sources})
target_link_libraries(ensd dl securec base common)
