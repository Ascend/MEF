/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:

   http://license.coscl.org.cn/MulanPSL2

   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */

#include <stdlib.h>
#include <string.h>
#include <pwd.h>
#include <unistd.h>
#include <utmp.h>
#include <arpa/inet.h>

#include "go_utmp.h"

#define TTY_NAME_LEN_ERR (-2)
#define IP_ADDR_ERR (-3)

int GetSSHIP(char *ut_name, char *ut_host)
{
    const int kLenDev = 5; // length of "/dev/"

    if (isatty(0) == 0) {
        return -1;
    }
    const char *tty = ttyname(0);
    if (tty == NULL) {
        return -1;
    }
    if (strlen(tty) <= kLenDev) {
        return TTY_NAME_LEN_ERR;
    }
    setutent();
    struct utmp *ut = NULL;
    const int maxLoop = 1024;
    int i = 0;
    for (i = 0; i < maxLoop; i++) {
        ut = getutent();
        if (ut == NULL) {
            break;
        }
        if (strcmp(ut->ut_line, tty + kLenDev) == 0) {
            break;
        }
    }
    endutent();
    if (ut != NULL) {
        for (i = 0; i < UT_NAMESIZE; i++) {
            ut_name[i] = ut->ut_user[i];
        }
        struct in_addr addr;
        if (inet_aton(ut->ut_host, &addr) == 0) {
            return IP_ADDR_ERR;
        }
        for (i = 0; i < UT_HOSTSIZE; i++) {
            ut_host[i] = ut->ut_host[i];
        }
        return 0;
    }
    return IP_ADDR_ERR;
}