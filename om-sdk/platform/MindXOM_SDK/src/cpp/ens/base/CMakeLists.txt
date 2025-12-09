set(sources
  ens_stack.c
  ens_rbtree.c
  ens.c
  ens_conf_file.c
  ens_context.c
  ens_intf.c
  ens_module.c
)

add_library(base SHARED ${sources})
target_link_libraries(base securec dl common)
