set(sources
  alarm_process.c
  alarm_shield.c
)


add_library(alarm_process SHARED ${sources})

target_link_libraries(alarm_process pthread securec common)
