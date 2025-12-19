/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   OMSDK is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 * Description: 磁盘挂载故障检测功能实现
 */
#include <stdio.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdlib.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/wait.h>
#include "fault_check.h"
#include "securec.h"
#include "certproc.h"
#include "file_checker.h"

#define CHECK_SCRIPTS_PATH          "/usr/local/mindx/MindXOM/scripts/mount_check.sh"
#define CHECK_BUF_LEN               32
#define MOUNT_CFG_ERR_MASK          0x1
#define MOUNT_PART_LOSE_MASK        0x2
#define MOUNT_DEV_LOSE_MASK         0x4
#define MOUNT_NOT_DO_MASK           0x8
#define MOUNT_POINT_ERR_MASK        0x10
#define CHECK_ERR_BUF_LEN           1024

#define FAULT_EXE_RETRY_CNT         30000
#define FAULT_EXE_TASK_DELAY        10000

#define EXIT_WITH_CMD_NOT_EXECUTE   126
#define EXIT_WITH_CMD_NOT_FOUND     127

typedef struct {
    char *out_buf;
    unsigned int out_len;
    char *err_buf;
    unsigned int err_len;
} PROCESS_BUF_INFO;

static int giMountStatus = 0;

static void process_pipe_in_info(int pipe_in, char *out_buf, unsigned int out_len)
{
    ssize_t count = 0;
    ssize_t read_len;
    int try_cnt = 0;

    if ((out_len > 0) && (out_buf != NULL)) {
        /* 将读设置为非阻塞, 读取错误信息超时时间为100ms */
        (void)fcntl(pipe_in, F_SETFL, O_NONBLOCK);
        while (1) {
            read_len = read(pipe_in, out_buf + count, out_len - count);
            try_cnt++;
            if (try_cnt > FAULT_EXE_RETRY_CNT) {
                FAULT_LOG_ERR("execve timeout,count:%ld, n:%ld.", count, read_len);
                break;
            }

            if (read_len < 0) {
                (void)usleep(FAULT_EXE_TASK_DELAY);
                continue;
            }

            count += read_len;
            if ((count > 0) && (read_len == 0)) {
                break;
            }
            (void)usleep(FAULT_EXE_TASK_DELAY);
        }
    }
}

static int check_task_exit_status(pid_t pid)
{
    int status = 0;
    int ret;

    do {
        ret = waitpid(pid, &status, 0);
    } while ((ret < 0) && (errno == EINTR));

    if (WIFEXITED(status) != 0) {
        ret = WEXITSTATUS(status);
        if (ret != 0) {
            FAULT_LOG_ERR("waitpid exit status:%d,ret:%d.", status, ret);
        }
    } else {
        ret = errno;
        FAULT_LOG_ERR("pid exit illegal.errno:%d.", ret);
    }
    return ret;
}

static void child_task_process(int fd_out[2], int fd_err[2], const char *path, char * const cmd[], char num)
{
    char * const envs[] = {NULL};

    close(fd_out[0]);
    close(fd_err[0]);

    if (fd_out[1] != STDOUT_FILENO) {
        if (dup2(fd_out[1], STDOUT_FILENO) != STDOUT_FILENO) {
            close(fd_out[1]);
            close(fd_err[1]);
            _exit(EXIT_WITH_CMD_NOT_EXECUTE);
        }
    }

    if (fd_err[1] != STDERR_FILENO) {
        if (dup2(fd_err[1], STDERR_FILENO) != STDERR_FILENO) {
            close(fd_out[1]);
            close(fd_err[1]);
            _exit(EXIT_WITH_CMD_NOT_EXECUTE);
        }
    }

    close(fd_out[1]);
    close(fd_err[1]);

    if (execve(path, cmd, envs) == -1) {
        _exit(EXIT_WITH_CMD_NOT_FOUND);
    }

    exit(0);
}

static int parent_task_process(pid_t pid, int fd_out[2], int fd_err[2], PROCESS_BUF_INFO *pstBufInfo)
{
    if ((pstBufInfo == NULL) ||
        ((pstBufInfo->out_buf == NULL) && (pstBufInfo->out_len != 0)) ||
        ((pstBufInfo->err_buf == NULL) && (pstBufInfo->err_len != 0))) {
        FAULT_LOG_ERR("invalid inputs.");
        return EDGE_ERR;
    }

    close(fd_out[1]);
    close(fd_err[1]);
    close(fd_err[0]);

    process_pipe_in_info(fd_out[0], pstBufInfo->out_buf, pstBufInfo->out_len);
    close(fd_out[0]);

    return check_task_exit_status(pid);
}

/* 通过脚本进行检测 */
static int fault_exe(char *path, char * const cmd[], PROCESS_BUF_INFO *pstBufInfo)
{
#define FAULT_TWO 2
    int ret;
    pid_t pid;
    int fd_out[FAULT_TWO] = {-1, -1};
    int fd_err[FAULT_TWO] = {-1, -1};

    if ((path == NULL) || (cmd == NULL) || (pstBufInfo == NULL) ||
        ((pstBufInfo->out_buf == NULL) && (pstBufInfo->out_len != 0)) ||
        ((pstBufInfo->err_buf == NULL) && (pstBufInfo->err_len != 0))) {
        FAULT_LOG_ERR("invalid inputs.");
        return EDGE_ERR;
    }

    ret = pipe(fd_out);
    if (ret != 0) {
        ret = errno;
        FAULT_LOG_ERR("pipe create fd_out fail,error code:%d.", ret);
        return EDGE_ERR;
    }

    ret = pipe(fd_err);
    if (ret != 0) {
        ret = errno;
        FAULT_LOG_ERR("pipe create fd_err fail,error code:%d.", ret);
        goto close_fd_out;
    }

    pid = fork();
    if (pid == -1) {
        ret = errno;
        FAULT_LOG_ERR("fork fail,error code:%d.", ret);
        goto close_fd_out_and_err;
    } else if (pid == 0) {
        child_task_process(fd_out, fd_err, path, cmd, FAULT_TWO);
    } else {
        return parent_task_process(pid, fd_out, fd_err, pstBufInfo);
    }

close_fd_out_and_err:
    close(fd_err[0]);
    close(fd_err[1]);

close_fd_out:
    close(fd_out[0]);
    close(fd_out[1]);
    return EDGE_ERR;
}

int fault_get_mount_status(void)
{
    char *check_cmd[] = {CHECK_SCRIPTS_PATH, NULL};
    char p_buf[CHECK_BUF_LEN + 1] = {0};
    char errInfoBuf[CHECK_ERR_BUF_LEN + 1] = {0};
    static int last_status = 0;
    if (memset_s(p_buf, sizeof(p_buf), 0, sizeof(p_buf)) != 0 ||
        memset_s(errInfoBuf, sizeof(errInfoBuf), 0, sizeof(errInfoBuf)) != 0) {
        FAULT_LOG_ERR("memset_s error!");
        return EDGE_ERR;
    }

    if (check_file_path_valid(CHECK_SCRIPTS_PATH) != COMMON_SUCCESS) {
        FAULT_LOG_ERR("check_file_path_valid fail.");
        return EDGE_ERR;
    }

    PROCESS_BUF_INFO stTaskBufInfo = {0};
    stTaskBufInfo.out_buf = p_buf;
    stTaskBufInfo.out_len = CHECK_BUF_LEN;
    stTaskBufInfo.err_buf = errInfoBuf;
    stTaskBufInfo.err_len = CHECK_ERR_BUF_LEN;

    if (fault_exe(CHECK_SCRIPTS_PATH, (char * const *)check_cmd, &stTaskBufInfo) != 0) {
        FAULT_LOG_ERR("fault_exe fail.");
        return EDGE_ERR;
    }

    giMountStatus = (int)StrToInt(p_buf, CHECK_BUF_LEN + 1);
    if (last_status != giMountStatus) {
        FAULT_LOG_ERR("mount status(%d)-->(%d).err:%s.", last_status, giMountStatus, errInfoBuf);
    }
    last_status = giMountStatus;
    return EDGE_OK;
}

int fault_check_mount_cfg_err(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    (void)fault_id;
    (void)sub_id;
    if (value == NULL) {
        FAULT_LOG_ERR("input null.");
        return EDGE_ERR;
    }

    if ((giMountStatus & MOUNT_CFG_ERR_MASK) != 0) {
        *value = FAULT_STATUS_ERR;
    } else {
        *value = FAULT_STATUS_OK;
    }

    return EDGE_OK;
}

int fault_check_mount_part_lose(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    (void)fault_id;
    (void)sub_id;
    if (value == NULL) {
        FAULT_LOG_ERR("input null.");
        return EDGE_ERR;
    }

    if ((giMountStatus & MOUNT_PART_LOSE_MASK) != 0) {
        *value = FAULT_STATUS_ERR;
    } else {
        *value = FAULT_STATUS_OK;
    }

    return EDGE_OK;
}

int fault_check_mount_dev_lose(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    (void)fault_id;
    (void)sub_id;
    if (value == NULL) {
        FAULT_LOG_ERR("input null.");
        return EDGE_ERR;
    }

    if ((giMountStatus & MOUNT_DEV_LOSE_MASK) != 0) {
        *value = FAULT_STATUS_ERR;
    } else {
        *value = FAULT_STATUS_OK;
    }

    return EDGE_OK;
}

int fault_check_mount_not_do(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    (void)fault_id;
    (void)sub_id;
    if (value == NULL) {
        FAULT_LOG_ERR("input null.");
        return EDGE_ERR;
    }

    if ((giMountStatus & MOUNT_NOT_DO_MASK) != 0) {
        *value = FAULT_STATUS_ERR;
    } else {
        *value = FAULT_STATUS_OK;
    }

    return EDGE_OK;
}

int fault_check_mount_point_err(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    (void)fault_id;
    (void)sub_id;
    if (value == NULL) {
        FAULT_LOG_ERR("input null.");
        return EDGE_ERR;
    }

    if ((giMountStatus & MOUNT_POINT_ERR_MASK) != 0) {
        *value = FAULT_STATUS_ERR;
    } else {
        *value = FAULT_STATUS_OK;
    }

    return EDGE_OK;
}

#define NFS_MAX_LINES 32
#define NFS_STATUS_ERROR_STR "error"
#define NFS_STATUS_INFO_FILE "/run/nfs/nfs_status_info"
#define NFS_INFO_SPLIT_CHAR " ="
#define NFS_MAX_LINE_SIZE 1024
int fault_check_nfs_mount_err(unsigned int fault_id, unsigned int sub_id, unsigned short *value)
{
    char *ptr = NULL;
    char *tmp_ptr = NULL;
    char line_buff[NFS_MAX_LINE_SIZE] = {0};

    if (value == NULL) {
        FAULT_LOG_ERR("nfs check parameter invalid!");
        return EDGE_ERR;
    }

    if ((access(NFS_STATUS_INFO_FILE, F_OK)) == -1) {
        // 文件不存在，无告警直接返回
        return EDGE_OK;
    }

    FILE *nfs_fd = safety_fopen(NFS_STATUS_INFO_FILE, "r");
    if (nfs_fd == NULL) {
        FAULT_LOG_ERR("nfs open status file failed!");
        return EDGE_ERR;
    }

    *value = FAULT_STATUS_OK;

    for (int i = 0; i < NFS_MAX_LINES; i++) {
        if (memset_s(line_buff, NFS_MAX_LINE_SIZE, 0, NFS_MAX_LINE_SIZE) != 0) {
            FAULT_LOG_ERR("memset_s error!");
            *value = FAULT_STATUS_ERR;
            break;
        }
        ptr = fgets(line_buff, NFS_MAX_LINE_SIZE, nfs_fd);
        if (ptr == NULL) {
            break;
        }

        tmp_ptr = line_buff;
        while (tmp_ptr != NULL) {
            tmp_ptr = strstr(tmp_ptr, NFS_INFO_SPLIT_CHAR);
            if (tmp_ptr == NULL) {
                break;
            }

            tmp_ptr = tmp_ptr + strlen(NFS_INFO_SPLIT_CHAR);
            ptr = tmp_ptr;
        }

        if (ptr == NULL) {
            FAULT_LOG_ERR("nfs info line (%d) have no = char!", i);
            continue;
        }

        if (strncmp(ptr, NFS_STATUS_ERROR_STR, strlen(NFS_STATUS_ERROR_STR)) == 0) {
            *value = FAULT_STATUS_ERR;
            break;
        }
    }

    if (fclose(nfs_fd) != 0) {
        FAULT_LOG_ERR("nfs fault check close config file failed!");
    }

    return EDGE_OK;
}
