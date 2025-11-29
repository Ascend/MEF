/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: Configuration file operations.
 * Create: 2020-11-07
 */

#include "ens_conf_file.h"
#include "file_checker.h"
#include "ens_log.h"
#include "ens_context.h"
#include "ens_dlist.h"
#include "ens_module.h"

static ens_uint32_t ens_conf_get_token(char *buf, ens_uint32_t size, char **result)
{
    char *pc = buf;
    char *token_start = NULL;
    char *token = NULL;
    size_t token_size;

    if ((buf == NULL) || (result == NULL)) {
        ENS_LOG_ERR("inputs null");
        if (result) {
           *result = NULL;
        }
        return 0;
    }

    while ((is_space(*pc) == ENS_OK) && (pc - buf < size)) {
        pc++;
    }
    if (pc - buf < size) {
        token_start = pc;
        while ((is_valide_char(*pc) == ENS_OK) && (pc - buf < size)) {
            pc++;
        }
        token_size = pc - token_start;
        if (token_size > 0) {
            token = malloc(token_size + 1);
            if (token == NULL) {
                ENS_LOG_FATAL("malloc fail for token.");
                return 0;
            }
            if (memset_s(token, token_size + 1, 0, token_size + 1) != 0) {
                ENS_LOG_FATAL("memset_s error!");
                free(token);
                token = NULL;
                return 0;
            }
            if (strncpy_s(token, token_size + 1, token_start, token_size) != 0) {
                ENS_LOG_FATAL("strncpy_s fail for token.");
                free(token);
                token = NULL;
                return 0;
            }
            *result = token;
        }
    }

    return pc - buf;
}

static ens_uint32_t ens_conf_parse_item(char *buf, int size)
{
    char *value = NULL;
    char *key = NULL;

    if (buf == NULL) {
        ENS_LOG_ERR("buf null");
        return 0;
    }

    char *pc = buf;
    ens_uint32_t token_size = ens_conf_get_token(pc, size - (pc - buf), &key);
    pc += token_size;
    if (key) {
        token_size = ens_conf_get_token(pc, size - (pc - buf), &value);
        pc += token_size;
        if (value == NULL) {
            ENS_LOG_ERR("Failed getting value for [%s]", key);
            free(key);
            key = NULL;
            return pc - buf;
        }

        ens_conf_item_t *item = malloc(sizeof(ens_conf_item_t));
        if (item) {
            if (memset_s(item, sizeof(ens_conf_item_t), 0, sizeof(ens_conf_item_t)) != 0) {
                ENS_LOG_ERR("memset_s error!");
                free(key);
                key = NULL;
                free(value);
                value = NULL;
                free(item);
                item = NULL;
                return pc - buf;
            }
            ens_dlist_init_node(&item->dl_node);
            item->key = key;
            item->value = value;
            ens_dlist_add_before(&item->dl_node, &ens_ctx.config.items);
            ens_ctx.config.item_n++;
        } else {
            ENS_LOG_FATAL("malloc fail for item {key: %s, value: %s}.", key, value);
            free(key);
            key = NULL;
            free(value);
            value = NULL;
        }
    }

    return pc - buf;
}

static int ens_conf_parse_buffer(char *buf, int size)
{
    char *pc = buf;
    ens_uint32_t item_size;

    if (buf == NULL) {
        ENS_LOG_ERR("buf null");
        return ENS_ERR;
    }

    while (pc - buf < size) {
        item_size = ens_conf_parse_item (pc, size - (pc - buf));
        if (item_size == 0) {
            break;
        }
        pc += item_size;
    }

    return ENS_OK;
}

static int ens_conf_parse_file(int fd, int fsize)
{
    size_t fpos = 0;
    size_t bpos = 0;
    size_t fremain, bremain;
    ssize_t act_rn;
    size_t try_rn;

    if ((fsize <= 0) || (fsize > ENS_CONF_FILE_MAX_SIZE)) {
        ENS_LOG_ERR("fsize:%d too large.", fsize);
        return ENS_ERR;
    }

    size_t bsize = ENS_CONF_FILE_MAX_SIZE;
    char *buf = malloc(bsize);
    if (buf != NULL) {
        if (memset_s(buf, bsize, 0, bsize) != 0) {
            ENS_LOG_ERR("memset_s error!");
            free(buf);
            buf = NULL;
            return ENS_ERR;
        }
        fremain = fsize - fpos;
        bremain = bsize - bpos;
        try_rn = (fremain < bremain) ? fremain : bremain;
        act_rn = read(fd, buf, try_rn);
        if ((size_t)act_rn != try_rn) {
            ENS_LOG_ERR("Failed read config file.");
            free(buf);
            buf = NULL;
            return ENS_ERR;
        }
        if (act_rn == ENS_CONF_FILE_MAX_SIZE) {
            ENS_LOG_ERR("act_rn too large.");
            free(buf);
            buf = NULL;
            return ENS_ERR;
        }

        buf[bsize - 1] = '\0';
        ens_conf_parse_buffer(buf, act_rn);
        free(buf);
        buf = NULL;
        return ENS_OK;
    }

    ENS_LOG_ERR("malloc fail.");
    return ENS_ERR;
}

int ens_conf_load(const char *filename)
{
    int ret = ENS_ERR;
    int fd;
    struct stat finfo = {0};
    int fsize;

    if (check_file_path_valid(filename) != COMMON_SUCCESS) {
        ENS_LOG_ERR("conf load fail for filename null");
        return ENS_ERR;
    }

    fd = open(filename, O_RDONLY, ENS_FILE_DEFAULT_ACCESS);
    if (fd >= 0) {
        if (fstat(fd, &finfo) >= 0) {
            fsize = finfo.st_size;
            if ((fsize > 0) && (fsize <= ENS_CONF_FILE_MAX_SIZE)) {
                ens_conf_parse_file(fd, fsize);
                ret = ENS_OK;
            } else {
                ENS_LOG_ERR("fsize:%d should le max:%d", fsize, ENS_CONF_FILE_MAX_SIZE);
            }
        } else {
            ENS_LOG_ERR("fstat fail");
        }
        close(fd);
    } else {
        ENS_LOG_ERR("open file fail");
    }

    return ret;
}

int ens_conf_apply(void)
{
    ens_dlist_head_t *head = NULL;
    ens_dlist_node_t *node = NULL;
    ens_dlist_node_t *temp_node = NULL;
    ens_conf_item_t *item = NULL;

    head = &ens_ctx.config.items;
    for (node = head->next, temp_node = node->next; node != head; node = temp_node, temp_node = node->next) {
        item = ((ens_conf_item_t *)(((void *)(node)) - ((void *)(&((ens_conf_item_t *)0)->dl_node))));
        if (item->key == NULL) {
            ENS_LOG_ERR("item key is null.");
        }

        if ((item->key != NULL) && (strcmp("LoadModule", item->key) == 0)) {
            if (item->value != NULL) {
                ens_module_load(item->value);
            } else {
                ENS_LOG_FATAL("Unexpected configuration item.");
            }
        }

        ens_dlist_remove(node);
        free(item->value);
        item->value = NULL;
        free(item->key);
        item->key = NULL;
        free(item);
        item = NULL;
    }

    if (ens_ctx.module_num > 0) {
        ENS_LOG_INFO("call ens_module_assembly_all start.");
        ens_module_assembly_all();
        ENS_LOG_INFO("call ens_module_assembly_all end.");
    }

    return ENS_OK;
}

void ens_conf_initialize(void)
{
    ens_dlist_init_head(&ens_ctx.config.items);
    return;
}
