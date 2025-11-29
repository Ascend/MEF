# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os


class NetManagerConstants:
    """网管配置模块常量定义."""
    CONFIG_ROOT_PATH = "/home/data/config"
    CERT_DOWNLOAD_PATH = "/run/web/cert"
    REDFISH_PATH = os.path.join(CONFIG_ROOT_PATH, "redfish")
    REDFISH_KSF = os.path.join(REDFISH_PATH, "redfish_encrypt.keystore")
    REDFISH_BAK_KSF = os.path.join(REDFISH_PATH, "redfish_encrypt_backup.keystore")
    REDFISH_ALG_CFG = os.path.join(REDFISH_PATH, "om_alg.json")

    # 文件限制1MB
    LIMIT_1M_SIZE = 1 * 1024 * 1024
    WEB = "Web"
    FUSION_DIRECTOR = "FusionDirector"
    FD_CERT_NAME = "FD.crt"
    SERVER_NAME = "fd.fusiondirector.huawei.com"
    PORT = "443"
    # FD固定的初始对接账号
    FD_INIT_ACCOUNT = "EdgeAccount"

    # FD来源证书数量限制
    CERT_FROM_FD_LIMIT_NUM = 3
    CERT_FORM_FD_AND_WEB_LIMIT_NUM = 4

    # FD 下发的source白名单
    SOURCE_WHITELIST = ("EdgeManager", "controller")


class ApiOperationMapper:
    OPERATION_MAPPER = {
        "GET NetManager": "Querying net manage configuration",
        "GET NetManager/NodeID": "Querying node id",
        "POST NetManager": "Configuring net manage",
        "POST NetManager/ImportFdCert": "Importing fd cert",
        "GET NetManager/QueryFdCert": "Querying fd cert",
        "POST NetManager/ImportFdCrl": "Importing fd crl",
    }
