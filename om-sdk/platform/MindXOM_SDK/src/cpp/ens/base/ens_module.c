/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Module operations.
 * Create 2020-11-07
 */

#include "ens_module.h"
#include "ens_base.h"
#include "file_checker.h"
#include "ens_log.h"
#include "ens_context.h"
#include "ens_rbtree.h"
#include "ens_dlist.h"
#include "ens_stack.h"
#include "ens_intf.h"


static int ens_module_start(ens_module_t *module)
{
    if (module == NULL) {
        ENS_LOG_ERR("input null");
        return ENS_ERR;
    }

    ENS_LOG_INFO("Module[%s] is starting.", module->name);
    int ret = module->start();
    if (ret == 0) {
        ENS_LOG_INFO("Module[%s] start success.", module->name);
        module->is_started = ENS_TRUE;
        ret = ENS_OK;
    } else {
        ENS_LOG_ERR("Module[%s] start fail.", module->name);
        ret = ENS_ERR;
    }

    return ret;
}

static int ens_module_load_work(const char *module_file, char *name, int len)
{
    unsigned int ret = 0;
    char intf_name[ENS_MODULE_INTF_NAME_MAXLEN];
    char *symbol = intf_name;
    if ((module_file == NULL) || (name == NULL) || (len > ENS_MODULE_FULLPATH_MAXLEN)) {
        return ENS_ERR;
    }

    void *handle = safety_dlopen(module_file, RTLD_NOW | RTLD_GLOBAL, COMMON_TRUE, 0);
    if (handle) {
        ens_module_t *module = malloc(sizeof(ens_module_t));
        if (module) {
            if (memset_s(module, sizeof(ens_module_t), 0, sizeof(ens_module_t)) != 0) {
                ENS_LOG_ERR("memset_s error!");
                free(module);
                module = NULL;
                dlclose(handle);
                return ENS_ERR;
            }
            ret |= (unsigned int)sprintf_s(module->name, ENS_MODULE_NAME_MAXLEN, "%s", name);
            ens_dlist_init_head(&module->dep_intfs);
            ret |= (unsigned int)sprintf_s(intf_name, ENS_MODULE_INTF_NAME_MAXLEN, "%s_load", module->name);
            module->load = (module_load_t)dlsym(handle, symbol);
            ret |= (unsigned int)sprintf_s(intf_name, ENS_MODULE_INTF_NAME_MAXLEN, "%s_unload", module->name);
            module->unload = (module_unload_t)dlsym(handle, symbol);
            ret |= (unsigned int)sprintf_s(intf_name, ENS_MODULE_INTF_NAME_MAXLEN, "%s_start", module->name);
            module->start = (module_start_t)dlsym(handle, symbol);
            ret |= (unsigned int)sprintf_s(intf_name, ENS_MODULE_INTF_NAME_MAXLEN, "%s_stop", module->name);
            module->stop = (module_stop_t)dlsym(handle, symbol);

            if ((module->load == 0) || (module->unload == 0) || (module->start == 0) || (module->stop == 0)) {
                ENS_LOG_ERR("Invalid module:%s, ret:%u", name, ret);
                free(module);
                module = NULL;
                dlclose(handle);
                return ENS_ERR;
            }

            ens_rbtree_init_node(&ens_ctx.modules, &module->rb_node, &module->name[0]);
            ens_rbtree_insert(&ens_ctx.modules, &module->rb_node);
            ens_ctx.module_num++;
            ens_ctx.curr_module = module;
            module->load();
            ens_ctx.curr_module = NULL;
            return ENS_OK;
        }

        ENS_LOG_FATAL("malloc fail for module[%s].", name);
        dlclose(handle);
        return ENS_ERR;
    }

    ENS_LOG_FATAL("load module [%s] fail", dlerror());
    return ENS_ERR;
}

int ens_module_load(char *name)
{
    char module_file[ENS_MODULE_FULLPATH_MAXLEN];

    if (name == NULL) {
        ENS_LOG_ERR("input null");
        return ENS_ERR;
    }

    ENS_LOG_INFO("load module:%s start", name);

    if (ens_rbtree_search(&ens_ctx.modules, name)) {
        ENS_LOG_WARN("Duplicated module:%s.", name);
        return ENS_OK;
    }

    int ret = sprintf_s(module_file, ENS_MODULE_FULLPATH_MAXLEN,
        "%s/modules/lib%s.so", ens_ctx.home_path, name);
    if (ret < 0) {
        ENS_LOG_ERR("sprintf fail");
        return ENS_ERR;
    }

    if (ens_module_load_work(module_file, name, ENS_MODULE_FULLPATH_MAXLEN) != ENS_OK) {
        ENS_LOG_ERR("load module:%s fail", name);
        return ENS_ERR;
    }

    ENS_LOG_INFO("load module:%s success", name);
    return ENS_OK;
}

static int get_dep_intf_func(ens_intf_t *intf, ens_module_t **curr_module,
                             ens_stack_t *module_stack, ens_dep_intf_t *dep_intf)
{
    int ret = 0;

    if ((module_stack == NULL) || (curr_module == NULL) ||
        (*curr_module == NULL) || (dep_intf == NULL) ||
        (dep_intf->ppfn == NULL) || ((intf != NULL) && (intf->provider == NULL))) {
        ENS_LOG_FATAL("invalid inputs.");
        return -1;
    }

    if (intf != NULL) {
        if (intf->provider && intf->provider->is_started == ENS_TRUE) {
            *(dep_intf->ppfn) = intf->pfn;
            ret = 1;
        } else {
            ENS_LOG_TRACE("[%s] was not started, push [%s] to stack for interface [%s].",
                intf->provider->name, (*curr_module)->name, intf->name);

            ens_stack_push(module_stack, *curr_module);
            *curr_module = intf->provider;
            ret = -1;
        }
    } else {
        ENS_LOG_FATAL("Cannot find symbol '%s' depended by '%s'.", dep_intf->name, (*curr_module)->name);
        *(dep_intf->ppfn) = NULL;
    }

    return ret;
}

static int ens_module_start_module_by_dependency(ens_module_t *module)
{
    ens_dep_intf_t *dep_intf = NULL;
    ens_intf_t *intf = NULL;
    ens_module_t *curr_module = NULL;
    ens_stack_t *module_stack = NULL;
    ens_dlist_head_t *head = NULL;
    ens_dlist_node_t *dl_node = NULL;
    ens_dlist_node_t *temp_dl_node = NULL;

    module_stack = ens_stack_create();
    if (module_stack == NULL) {
        ENS_LOG_FATAL("Failed for creating stack.");
        return ENS_ERR;
    }

    curr_module = module;

    while (curr_module != NULL) {
        ens_module_t *tmp_module = curr_module;
        head = &curr_module->dep_intfs;
        for (dl_node = head->next, temp_dl_node = dl_node->next; dl_node != head;
            dl_node = temp_dl_node, temp_dl_node = dl_node->next) {
            dep_intf = ((ens_dep_intf_t *)(((void *)(dl_node)) - ((void *)(&((ens_dep_intf_t *)0)->dlist_node))));
            intf = ens_intf_get_by_name(dep_intf->name);
            if (get_dep_intf_func(intf, &curr_module, module_stack, dep_intf) < 0) {
                ENS_LOG_ERR("get_dep_intf_func fail.");
                break;
            }

            ens_dlist_remove(dl_node);
            free(dep_intf);
            dep_intf = NULL;
            if (curr_module->dep_intf_num > 0) {
                curr_module->dep_intf_num--;
            }
        }

        if (curr_module->dep_intf_num == 0) {
            if (curr_module->is_started != ENS_TRUE) {
                ens_module_start(curr_module);
            } else {
                ENS_LOG_FATAL("[%s] has been started.", curr_module->name);
            }
            curr_module = ens_stack_pop(module_stack);
        }

        if ((tmp_module == curr_module) && (curr_module->dep_intf_num != 0)) {
            ENS_LOG_WARN("[%s]init ok,num:%u will break.", curr_module->name, curr_module->dep_intf_num);
            break;
        }
    }

    ens_stack_destroy(module_stack);
    free(module_stack);
    module_stack = NULL;
    return ENS_OK;
}

int ens_module_assembly_all(void)
{
    ens_module_t *module = NULL;
    ens_rbtree_node_t *node = NULL;

    ENS_LOG_INFO("call ens_rbtree_first start.");
    node = ens_rbtree_first(&ens_ctx.modules);
    ENS_LOG_INFO("call ens_rbtree_first end.");

    while (node) {
        module = ((ens_module_t *)((void *)(node) - (size_t)(&((ens_module_t *)0)->rb_node)));
        if (module->is_started != ENS_TRUE) {
            ENS_LOG_INFO("call ens_module_start_module_by_dependency start.");
            ens_module_start_module_by_dependency(module);
            ENS_LOG_INFO("call ens_module_start_module_by_dependency end.");
        }

        ENS_LOG_INFO("call ens_rbtree_next start.");
        node = ens_rbtree_next(&ens_ctx.modules, node);
        ENS_LOG_INFO("call ens_rbtree_next end.");
    }

    return ENS_OK;
}

static int ens_module_compare_key(const char *key1, const char *key2)
{
    if ((key1 == NULL) || (key2 == NULL)) {
        ENS_LOG_ERR("inputs null");
        return -1;
    }

    return strcmp(key1, key2);
}

void ens_module_initialize(void)
{
    ens_rbtree_init(&ens_ctx.modules, ens_module_compare_key);
    return;
}
