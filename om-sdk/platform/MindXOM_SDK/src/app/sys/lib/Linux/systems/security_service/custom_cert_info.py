# coding: utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
class CustomCertInfo:
    # 保存生成csr过程中的私钥和加密口令文件
    EXPORT_TMP_DIR = "/run/exportcert"
    # 用户需要操作这个目录：下载保存生成的csr文件、用户签名之后的cert证书文件
    IMPORT_TMP_DIR = "/home/data/exportcert"
    GOLDEN_MOUNT_DIR = "/mnt/p1"
    PACKAGE_MOUNT_DIR = "/home/package"
    CERT_WORK_DIR = "/home/data/config/default"
    SERVER_KMC_CSR = "server_kmc.csr"
    SERVER_KMC_CERT = "server_kmc.cert"
    SERVER_KMC_PRI = "server_kmc.priv"
    SERVER_KMC_PSD = "server_kmc.psd"
    OM_CERT_KEYSTORE = "om_cert.keystore"
    OM_CERT_BACKUP_KEYSTORE = "om_cert_backup.keystore"
    OM_ALG_JSON = "om_alg.json"
    DEFAULT_DIR_MODE = 0o700
    DEFAULT_FILE_MODE = 0o400
    WORK_DIR_MODE = 0o700
    WORK_DIR_MODE_STR = "700"
    WORK_FILE_MODE = 0o600
    WORK_DIR_OWNER = "nobody"
    WORK_DIR_GROUP = "nobody"
    BACKUP_FILE_MODE = 0o600
    BACKUP_DIR_OWNER = "root"
    BACKUP_DIR_GROUP = "root"
    FILE_MAX_SIZE = 10 * 1024
