set(sources certproc.c create_cert.c)

add_library(certmanage SHARED ${sources})
target_link_libraries(certmanage securec pthread dl common ssl crypto)
