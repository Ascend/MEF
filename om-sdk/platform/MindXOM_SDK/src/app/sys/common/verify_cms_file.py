#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import ctypes
import logging
import os
import sys
from common.log.logger import run_log


def verify_cms_file(libverify_so_path, cms_path, crl_path, tar_path):
    libverif = ctypes.CDLL(libverify_so_path)
    _prepare_upgrade_image_cms = libverif.prepareUpgradeImageCms
    _prepare_upgrade_image_cms.argtypes = [
        ctypes.POINTER(ctypes.c_char * 300),
        ctypes.POINTER(ctypes.c_char * 300),
        ctypes.POINTER(ctypes.c_char * 300)
    ]
    _prepare_upgrade_image_cms.restype = ctypes.c_int

    cms_path = cms_path.encode(encoding='utf-8')
    crl_path = crl_path.encode(encoding='utf-8')
    tar_path = tar_path.encode(encoding='utf-8')
    cms_path_ = ctypes.cast(cms_path, ctypes.POINTER(ctypes.c_char * 300))
    crl_path_ = ctypes.cast(crl_path, ctypes.POINTER(ctypes.c_char * 300))
    tar_path_ = ctypes.cast(tar_path, ctypes.POINTER(ctypes.c_char * 300))
    ret = _prepare_upgrade_image_cms(cms_path_, crl_path_, tar_path_)

    if ret == 0:
        run_log.info(f"check package success, file name={tar_path}")
        return True
    else:
        run_log.error(f"check package failed, file name={tar_path}, C lib return code is {ret}")
        return False


def check_parameter(argv):
    if len(argv) != 5:
        run_log.error(f"usage: python3 verify_cms_file.py <libverify_so_path> <cms_path> <crl_path> <tar_path>")
        return False

    if not os.path.isfile(argv[1]):
        run_log.error("libverify_so_path: is not file")
        return False

    if not os.path.isfile(argv[2]):
        run_log.error("cms_path: is not file")
        return False

    if not os.path.isfile(argv[3]):
        run_log.error("crl_path: is not file")
        return False

    if not os.path.isfile(argv[4]):
        run_log.error("tar_path: is not file")
        return False

    return True


def init_logger():
    logging.basicConfig(filename="/var/plog/upgrade.log",
                        level=logging.DEBUG,
                        format="%(asctime)s %(message)s",
                        datefmt="%Y-%m-%d %H:%M:%S")


if __name__ == "__main__":
    init_logger()
    # 由于此文件存在单独执行和被导入执行两种情况，其中单独执行时需要单独初始化日志模块，被导入执行时直接导入common.log中的run_log
    run_log = logging
    if not check_parameter(sys.argv):
        sys.exit(1)

    if not verify_cms_file(sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]):
        sys.exit(1)

    sys.exit(0)
