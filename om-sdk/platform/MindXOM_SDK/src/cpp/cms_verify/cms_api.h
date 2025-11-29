/*
 * Copyright: Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */
#ifndef _CMS_API_H_
#define _CMS_API_H_

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <sys/types.h>
#include <sys/errno.h>
#include <sys/syscall.h>
#include <unistd.h>

int prepareUpgradeImageCms(const char *pathname_cms, const char *pathname_crl, const char *pathname_tar);

#endif
