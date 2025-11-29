set(KEY_TOP_DIR ${CMAKE_SOURCE_DIR}/..)

set(LIBRARY_INC_PATH
    "${KEY_TOP_DIR}/platform/cmscbb_csec/include"
)

file(GLOB inner_sources ${KEY_TOP_DIR}/platform/cmscbb_csec/src/**.c)
set(sources
  cms_api.c
  upgrade_cmscbb_callback.c
  ${inner_sources}
)

add_definitions(-DCMSCBB_SUPPORT_RSAPSS)

include_directories(${LIBRARY_INC_PATH})

add_library(verify SHARED ${sources})
target_link_libraries(verify securec pthread common ssl crypto)
