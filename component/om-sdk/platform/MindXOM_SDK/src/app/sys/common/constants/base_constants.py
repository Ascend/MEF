# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
import uuid
from enum import Enum
from typing import Set


class CertPathConstant:
    BACKUP_CERT_PATH = "/home/data/config/certs"
    EULER_CERT_PATH = "/etc/pki/ca-trust/source/anchors"
    UBUNTU_CERT_PATH = "/usr/local/share/ca-certificates"


class FlagPathConstant:
    """标识文件路径常量"""
    # FD那边NTP使能标志，用于标识是否同步FD的时间
    NTP_SYNC_NET_MANAGER = "/home/data/config/redfish/sync_net_manager"


class MefBusyConstants:
    MEF_BUSY_INNER_CODE = 5
    MEF_BUSY_RESPONSE_CODE = 555
    MEF_BUSY_RESPONSE_MSG = "MEF service is busy"


class RestoreConfigConstants:
    RESTORE_MEF = "mef"
    RESTORE_MONITOR = "monitor"
    RESTORE_NETWORK = "network"


class CommonConstants(object):
    ERR_CODE_200 = 200
    ERR_CODE_201 = 201
    ERR_CODE_206 = 206
    ERR_CODE_400 = 400
    # 请求访问的资源不存在
    ERR_CODE_404 = 404
    # 请求资源的状态之间存在冲突
    ERR_CODE_500 = 500
    STATUS_OK = (200, 201, 202, 206)

    # A500环境 MEF安装路径
    A500_MEF_INSTALL_PATH = "/usr/local/mindx/MEFEdge/software"

    OM_HOME_PATH = "/usr/local/mindx"
    OM_LOG_HOME_PATH = "/var/plog"
    OM_LOG_IBMA_PATH = os.path.join(OM_LOG_HOME_PATH, "ibma_edge")
    LOCAL_CRACKLIB_DIR = "/usr/share/cracklib"
    LOCAL_SCRIPT_DIR = "/usr/local/scripts"
    HOME_DATA_PATH = "/home/data"
    OM_WORK_DIR_PATH = os.path.join(OM_HOME_PATH, "MindXOM")
    OM_UPGRADE_DIR_PATH = os.path.join(OM_HOME_PATH, "MindXOMUpgrade")
    OM_SYS_DIR_PATH = os.path.join(OM_HOME_PATH, "MindXOMSystem")
    REDFISH_LOG_PATH = os.path.join(OM_LOG_HOME_PATH, "redfish")
    MINDXOM_USER = "MindXOM"
    NGINX_USER = "nobody"
    MONITOR_USER = "root"
    ERR_GENERAL_INFO = "GeneralError"
    ERR_PARTICULAR_INFO = "ParticularError"
    CONFIG_HOME_PATH = os.path.join(HOME_DATA_PATH, "config")
    IBMA_EDGE_SERVICE_PATH = "/home/data/ies/ibma_edge_service.ini"
    REDFISH_EDGE_DB_FILE_PATH = os.path.join(CONFIG_HOME_PATH, "redfish/redfish_edge.db")
    MONITOR_EDGE_DB_FILE_PATH = os.path.join(CONFIG_HOME_PATH, "monitor/monitor_edge.db")
    DEFAULT_IBMA_EDGE_SERVICE_PATH = os.path.join(OM_WORK_DIR_PATH,
                                                  "software/ibma/lib/Linux/config/ibma_edge_service.ini")
    STR_DATE_FORMAT = "%Y-%m-%d %H:%M:%S"
    OM_VERSION_XML_FILE = os.path.join(OM_WORK_DIR_PATH, "version.xml")
    REDFISH_BACKUP_DIR = os.path.join(CONFIG_HOME_PATH, "redfish_backup")
    MONITOR_BACKUP_DIR = os.path.join(CONFIG_HOME_PATH, "monitor_backup")
    # 数据库迁移备库路径
    MIGRATE_SH = os.path.join(OM_UPGRADE_DIR_PATH, "scripts/migrate.sh")
    # 读取配置文件最大1MB
    OM_READ_FILE_MAX_SIZE_BYTES = 1 * 1024 * 1024
    USER_LOCK_STATE = "The user is locked."
    # 证书最大限制1M
    MAX_CERT_LIMIT = 1 * 1024 * 1024
    # 吊销列表最大限制10K
    MAX_CRL_LIMIT = 10 * 1024
    # 弱口令字典配置目录
    OS_CRACKLIB_DIR = "/usr/share/cracklib"
    OS_PW_DICT_DIR = "/usr/share/cracklib/pw_dict"

    # 循环次数限制
    MAX_ITER_LIMIT = 32000

    # 前端配置路径
    WEB_CONF = os.path.join(OM_WORK_DIR_PATH, "software/nginx/html/manager/WhiteboxConfig/json/config.json")

    # 200外设配置路径
    DEVICE_CFG_PATH = os.path.join(OM_WORK_DIR_PATH, "software", "ibma", "lib", "Linux", "config", "device_config.json")

    # Redfish进程证书和吊销列表临时文件
    REDFISH_CERT_TMP_FILE = "/run/web/tmp_crt.crt"
    REDFISH_CRL_TMP_FILE = "/run/web/tmp_crl.crl"

    # 硬件形态
    ATLAS_A200_A2 = "Atlas A200 A2"
    ATLAS_200I_A2 = "Atlas 200I A2"
    ATLAS_200I_DK_A2 = "Atlas 200I DK A2"
    A200_MODELS = (ATLAS_200I_DK_A2, ATLAS_200I_A2, ATLAS_A200_A2)

    ATLAS_A500_A2 = "Atlas A500 A2"
    ATLAS_500_A2 = "Atlas 500 A2"
    A500_MODELS = (ATLAS_A500_A2, ATLAS_500_A2)

    # kmc相关路径
    UDS_KS_DIR = os.path.join(OM_WORK_DIR_PATH, "software", "ibma", "cert")
    NGINX_KS_DIR = "/home/data/config/default"
    FLASK_KS_DIR = "/home/data/config/redfish"
    UDS_CLIENT_KS_DIR = os.path.join(OM_WORK_DIR_PATH, "software", "RedfishServer", "cert")
    REDFISH_KS_DIR = FLASK_KS_DIR

    # uds路径
    IBMA_SOCK_PATH = "/run/iBMA.sock"

    # 操作类型
    TYPE_IMPORT = "import"
    TYPE_EXPORT = "export"
    TYPE_DELETE = "delete"


class LoggerConstants:
    """
    日志收集
    """
    RUN_DIR_AVAILABLE_SIZE = 512 * 1024 * 1024
    RUN_DIR = "/run"
    # 日志收集：临时存放路径
    MONITOR_TMP_COLLECT_LOG = "/run/log_collect.tar.gz"
    REDFISH_LOG_COLLECT_DIR = "/run/web/log_collect"
    WEB_LOG_COLLECT_PATH = "/run/web/log_collect/web_log_collect.tar.gz"
    FD_LOG_COLLECT_PATH = "/run/web/log_collect/fd_log_collect.tar.gz"
    # 日志收集：大小限制，单位:字节
    COLLECT_LOG_UPPER_LIMIT = 512 * 1024 * 1024


class UserManagerConstants:
    USER_INFO_BY_TOKEN = 'get_user_info_by_token'
    # 用户登录创建token信息
    OPER_TYPE_GET_USER_TOKEN = 'get_token'
    OPER_TYPE_DELETE_USER_TOKEN = 'delete_token'
    OPER_TYPE_GET_SESSION_TIMEOUT = 'get_session_timeout'
    OPER_TYPE_MODIFY_SESSION_TIMEOUT = 'modify_session_timeout'
    # 获取密码有效期天数
    OPER_TYPE_GET_ACCOUNT_EXPIRATION_DAY = 'get_account_expiration_day'
    # 修改密码有效期天数
    OPER_TYPE_MODIFY_ACCOUNT_EXPIRATION_DAY = 'modify_account_expiration_day'
    # FD修改密码有效期天数
    OPER_TYPE_FD_MODIFY_ACCOUNT_EXPIRATION_DAY = 'fd_modify_account_expiration_day'
    # FD修改会话超时时间
    OPER_TYPE_FD_MODIFY_SESSION_TIMEOUT = 'fd_modify_session_timeout'
    # 获取用户id集合ibma_socket.py
    OPER_TYPE_GET_USER_ID_LIST = 'get_user_id_list'
    # 获取用户信息集合
    OPER_TYPE_GET_USER_INFO = 'get_user_info'
    OPER_TYPE_GET_USER_LIST = 'get_user_list'
    OPER_TYPE_MODIFY_USER_INFO = 'modify_user_info'
    OPER_TYPE_MODIFY_PASSWORD = 'modify_password'

    # 校验密码是否有效
    OPER_TYPE_CHECK_PASSWORD = 'check_password'
    # 校验弱字典
    OPER_TYPE_CHECK_WEAK_DICT = "check_weak_dict"
    # 校验登录规则
    OPER_TYPE_CHECK_SECURITY_CONFIG = "check_security_config"
    # 重置os密码
    OPER_TYPE_RESET_OS_PASSWORD = "reset_password"

    USERNAME_PATTERN = r'[^0-9a-zA-Z]'
    # 默认密码过期天数
    DEFAULT_EXPIRATION_DAYS = 90
    DEFAULT_LOCKING_DURATION = 300
    DEFAULT_LOCKING_TIMES = 5
    DEFAULT_TOKEN_TIMEOUT = 900
    # 证书校验
    CHECK_CERTIFICATE_PATH = "/home/data/config/default/server_kmc.cert"
    # session时长期范围限制
    MAX_SESSION_TIMEOUT = 120
    MIN_SESSION_TIMEOUT = 5
    # 密码有效期范围限制
    MAX_PASSWORD_EXPIRATION_DAY = 365
    MIN_PASSWORD_EXPIRATION_DAY = 0
    # 密码范围限制
    MIN_PASSWORD_LENGTH = 8
    MAX_PASSWORD_LENGTH = 20
    # 密码范围限制
    MIN_USERNAME_PASSWORD_LENGTH = 1
    MAX_USERNAME_PASSWORD_LENGTH = 16
    # session的长度范围限制
    MIN_SESSION_LENGTH = 1
    MAX_SESSION_LENGTH = 48


class StorageOperation(Enum):
    """
    存储模块支持的操作类型
    SITE_TO_RAW: 将设备的小站名称还原为原始设备名称
    RAW_TO_SITE：将原始设备名称转换为设备的小站名称
    """
    SITE_TO_RAW = "site_to_raw"
    RAW_TO_SITE = "raw_to_site"


class MigrateOperate(Enum):
    MIGRATE = "migrate"
    INSTALL = "install"

    @classmethod
    def values(cls) -> Set[str]:
        return set(elem.value for elem in cls)


class PubkeyType(Enum):
    EVP_PKEY_RSA = 6
    EVP_PKEY_DSA = 116
    EVP_PKEY_DH = 28
    EVP_PKEY_EC = 408


class MefOperate(Enum):
    """MEF操作指令"""
    EXCHANGE_CRT = "exchange_crt"
    RESTART = "restart"
    STOP = "stop"


class MefNetStatus(Enum):
    """MEF网管配置状态查询"""
    FD_OM = 0  # FD并且需要OM联合管理
    UNKNOWN = 1  # 未知状态
    MEF = 2  # MEF独立管理


class ResetActionConst:
    ERR_RESET_SUCCEED = 0  # 系统复位成功
    ERR_RESET_FAILED = 1  # 系统复位失败
    ERR_RESET_UPGRADE = 2  # 系统正在升级

    EFFECT_CODE_MAP = {
        ERR_RESET_SUCCEED: f"ERR.0{ERR_RESET_SUCCEED}-Restart system successfully.",
        ERR_RESET_FAILED: f"ERR.0{ERR_RESET_FAILED}-Restart system failed",
        ERR_RESET_UPGRADE: f"ERR.0{ERR_RESET_UPGRADE}-System is upgrading.",

    }


class RecoverMiniOSConstants:
    FIRMWARE_NAME = "firmware.zip"
    PACKAGE_PATH = "/home/package"
    FIRMWARE_PATH = f"{PACKAGE_PATH}/firmware"
    CONFIG_PATH = "/home/package/config"
    RECOVER_FLAG = f"{CONFIG_PATH}/recover_mini_os_flag"
    NET_MANAGER_CONFIG = "net_manager_config.json"


class CertInfo:
    DOMAIN_NAME = "MindXOM"

    # 证书路径
    CERTS_DIR = "/home/data/config/default"
    UDS_CERT_DIR = "/usr/local/mindx/MindXOM/software/ibma/cert"
    UDS_CERT_UPGRADE_DIR = "/usr/local/mindx/MindXOMUpgrade/software/ibma/cert"
    CERT_CREATE_TYPE = ("normal", "force")
    PRIMARY_KEYSTORE_NAME = "om_cert.keystore"
    BACKUP_KEYSTORE_NAME = "om_cert_backup.keystore"
    ALG_JSON_NAME = "om_alg.json"
    ROOT_CERT_NAME = "root_ca"
    CERT_NAME = "server_kmc"

    @classmethod
    def gen_domain_name_with_uuid(cls) -> str:
        return f"{cls.DOMAIN_NAME}-{uuid.uuid4()}"
