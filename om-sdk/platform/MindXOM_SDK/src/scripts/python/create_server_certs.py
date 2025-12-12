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

"""
生成证书
"""
import ctypes
import os
import sys

from common.constants.base_constants import CertInfo
from common.file_utils import FileCheck
from common.file_utils import FileConfusion
from logger import upgrade_log


class CertMgr:

    def __init__(self, uid, gid, cert_dir, is_client):
        self.uid = uid
        self.gid = gid
        self.key = None
        self.cert = None
        self.is_client = is_client
        self.serial_number = 0
        self.cert_dir = cert_dir
        self.primary_keystore_file = os.path.join(cert_dir, CertInfo.PRIMARY_KEYSTORE_NAME)
        self.backup_keystore_file = os.path.join(cert_dir, CertInfo.BACKUP_KEYSTORE_NAME)
        self.alg_json_file = os.path.join(cert_dir, CertInfo.ALG_JSON_NAME)
        self.root_cert_file = os.path.join(cert_dir, f"{CertInfo.ROOT_CERT_NAME}.cert")
        self.root_key_file = os.path.join(cert_dir, f"{CertInfo.ROOT_CERT_NAME}.priv")
        self.root_pwd_file = os.path.join(cert_dir, f"{CertInfo.ROOT_CERT_NAME}.psd")
        self.server_cert_file = os.path.join(cert_dir, f"{CertInfo.CERT_NAME}.cert")
        self.server_key_file = os.path.join(cert_dir, f"{CertInfo.CERT_NAME}.priv")
        self.server_pwd_file = os.path.join(cert_dir, f"{CertInfo.CERT_NAME}.psd")

    def is_cert_ok(self, cert_dir):
        cert_list = []
        cert_file_list = []
        if cert_dir == CertInfo.CERTS_DIR:
            cert_list = [
                self.server_cert_file, self.server_key_file, self.server_pwd_file,
                self.primary_keystore_file, self.backup_keystore_file
            ]
            cert_file_list = [self.server_cert_file]
        if cert_dir == CertInfo.UDS_CERT_DIR or cert_dir == CertInfo.UDS_CERT_UPGRADE_DIR:
            cert_list = [
                self.root_cert_file, self.server_cert_file, self.server_key_file, self.server_pwd_file,
                self.primary_keystore_file, self.backup_keystore_file
            ]
            cert_file_list = [self.root_cert_file, self.server_cert_file]
        for file in cert_list:
            if not FileCheck.check_path_is_exist_and_valid(file):
                return False

        # 判断证书是否过期
        from lib.Linux.systems.security_service.security_service_clib import check_cert_expired
        for file in cert_file_list:
            if not check_cert_expired(file):
                return False
        return True

    def remove_old_cert_files(self):
        for file in (self.root_cert_file, self.root_key_file, self.root_pwd_file,
                     self.server_cert_file, self.server_key_file, self.server_pwd_file):
            if os.path.exists(file):
                FileConfusion.confusion_path(file, True)
                os.remove(file)

    def create_cert(self):
        file_path = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(os.path.realpath(__file__)))),
                                 "lib/libcertmanage.so")
        try:
            libcreate_cert = ctypes.CDLL(file_path)
        except Exception as err:
            upgrade_log.error("load libcertmanage so failed: %s", err)
            raise ValueError("load libcertmanage so failed.") from err

        cert_manage_init = libcreate_cert.cert_manage_init
        cert_manage_init.argtypes = (ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p)
        cert_manage_init.restype = ctypes.c_int
        ret_init = cert_manage_init(self.primary_keystore_file.encode(encoding='utf-8'),
                                    self.backup_keystore_file.encode(encoding='utf-8'),
                                    self.alg_json_file.encode(encoding='utf-8'))
        if ret_init != 0:
            upgrade_log.error("init kmc failed %s" % ret_init)
            raise ValueError("init kmc failed.")

        cert_manage_finalize = libcreate_cert.cert_manage_finalize
        cert_manage_finalize.restype = ctypes.c_int

        create_server_cert_func = libcreate_cert.create_server_cert
        create_server_cert_func.argtypes = (
            ctypes.c_char_p, ctypes.c_ulong, ctypes.c_char_p, ctypes.c_char_p, ctypes.c_char_p
        )
        create_server_cert_func.restype = ctypes.c_int
        ca_domain_name = CertInfo.gen_domain_name_with_uuid()
        server_domain_name = CertInfo.gen_domain_name_with_uuid()
        client_domain_name = CertInfo.gen_domain_name_with_uuid()
        cert_path_p = ctypes.create_string_buffer(self.cert_dir.encode(encoding="utf-8"))
        ca_domain_p = ctypes.create_string_buffer(ca_domain_name.encode(encoding="utf-8"))
        server_domain_p = ctypes.create_string_buffer(server_domain_name.encode(encoding="utf-8"))
        client_domain_p = ctypes.create_string_buffer(client_domain_name.encode(encoding="utf-8"))
        is_client = ctypes.c_ulong(self.is_client)
        ret_set = create_server_cert_func(cert_path_p, is_client, ca_domain_p, server_domain_p, client_domain_p)
        if ret_set != 0:
            upgrade_log.error("create server cert failed %s", self.cert_dir)
            cert_manage_finalize()
            raise ValueError("create server cert failed")

        cert_manage_finalize()
        upgrade_log.info("create server cert success.")

    def change_server_cert_fils(self):
        for file in [self.root_cert_file, self.root_key_file, self.root_pwd_file,
                     self.server_cert_file, self.server_key_file, self.server_pwd_file,
                     self.primary_keystore_file, self.backup_keystore_file]:
            os.lchown(file, self.uid, self.gid)
            os.chmod(file, 0o600)

    def remove_root_ca(self):
        for file in [self.root_cert_file, self.root_key_file, self.root_pwd_file]:
            if os.path.exists(file):
                FileConfusion.confusion_path(file, True)
                os.remove(file)


def create_server_certs(uid, gid, cert_dir, cert_create_type, is_client=0):
    upgrade_log.info("start to create server certs")
    cert_mgr = CertMgr(uid, gid, cert_dir, is_client)
    if cert_create_type == "normal":
        if cert_mgr.is_cert_ok(cert_dir):
            upgrade_log.info("server certs is already ok, no need to create again")
            return True
        else:
            upgrade_log.info("server certs not ok, need create it")
    else:
        upgrade_log.info("cert create type is force, will remove old cert and create new cert")

    try:
        cert_mgr.remove_old_cert_files()
        cert_mgr.create_cert()
        cert_mgr.change_server_cert_fils()

        # 生成https证书时，生成后移除根证书
        if cert_dir == CertInfo.CERTS_DIR:
            cert_mgr.remove_root_ca()

        upgrade_log.info("server certs create success")
        return True
    except Exception as error_info:
        upgrade_log.error(f"server certs create failed, reason: {error_info}")
        return False


def check_parameter(argv):
    if len(argv) != 5 and len(argv) != 6:
        upgrade_log.error("usage: python3 create_server_certs.py <uid> <gid> <cert_dir> <cert_create_type> <client>")
        return False

    # 校验id
    max_id_str_len = 10
    uid_str, gid_str = argv[1], argv[2]
    if len(uid_str) > max_id_str_len or len(gid_str) > max_id_str_len:
        upgrade_log.error("uid or gid invalid")
        return False
    if not uid_str.isdigit() or not gid_str.isdigit():
        upgrade_log.error("uid or gid invalid")
        return False

    argv[1] = int(uid_str)
    argv[2] = int(gid_str)

    # 校验证书目录
    cert_dir = argv[3]
    if cert_dir not in [CertInfo.CERTS_DIR, CertInfo.UDS_CERT_DIR, CertInfo.UDS_CERT_UPGRADE_DIR]:
        upgrade_log.error("cert dir invalid.")
        return False
    ret_info = FileCheck.check_path_is_exist_and_valid(cert_dir)
    if not ret_info:
        upgrade_log.error(f"check {cert_dir} failed, reason is {ret_info.error}")
        return False
    alg_json_file = os.path.join(cert_dir, CertInfo.ALG_JSON_NAME)
    ret_info = FileCheck.check_path_is_exist_and_valid(alg_json_file)
    if not ret_info:
        upgrade_log.error(f"check {alg_json_file} failed, reason is {ret_info.error}")
        return False

    # 校验类型
    cert_create_type = argv[4]
    if cert_create_type not in CertInfo.CERT_CREATE_TYPE:
        upgrade_log.error("cert create type invalid.")
        return False

    if len(argv) == 6:
        if argv[5] != "client":
            upgrade_log.error("cert create type invalid.")
            return False
        argv[5] = 1

    return True


def main(argv):
    if not check_parameter(argv):
        return False

    if not create_server_certs(*argv[1:]):
        return False

    return True


if __name__ == "__main__":
    RET = main(sys.argv)
    if not RET:
        sys.exit(1)
    sys.exit(0)
