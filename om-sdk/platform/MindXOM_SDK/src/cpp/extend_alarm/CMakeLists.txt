set(sources
        check_extend_alarm.c
        extend_alarm.c
    )

add_library(extend_alarm SHARED ${sources})

target_link_libraries(extend_alarm  pthread securec common)