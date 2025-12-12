set(sources
  devm.c
)

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fstack-protector-all")
add_library(devm SHARED ${sources})
target_link_libraries(devm securec common)
