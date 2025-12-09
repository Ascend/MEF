set(sources
    lpeblock.c
)

include_directories(
    ${CMAKE_CURRENT_LIST_DIR}
)
add_library(lpeblock SHARED ${sources})
target_link_libraries(lpeblock dl)
