# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os.path

from common.constants.base_constants import CommonConstants, MefBusyConstants


class UpgradeConstants:
    """校验升级包，调用升级脚本相关常量"""
    # 固件包升级脚本执行升级超时时间600秒
    UPGRADE_FIRMWARE_TIMEOUT = 600
    # 升级家目录
    UPGRADE_HOME_PATH = "/run/upgrade"
    # 升级包xml版本文件
    VERSION_XML_NAME = "version.xml"
    VERCFG_XML_NAME = "vercfg.xml"
    VERCFG_CMS_NAME = "vercfg.xml.cms"
    VERCFG_CRL_NAME = "vercfg.xml.crl"
    # 驱动升级进度文件
    PROGRESS_FILE = os.path.join(UPGRADE_HOME_PATH, "upgrade_progress")
    # 驱动升级脚本入口
    UPGRADE_DRV_SH = "/usr/local/scripts/upgrade_drv.sh"
    # OM升级脚本入口
    UPGRADE_OM_SH = "upgrade_om.sh"

    # SDK升级脚本通用入口
    SDK_UPGRADE_SH = "upgrade.sh"
    # SDK升级包解压目录
    SDK_UPGRADE_UNPACK_PATH = os.path.join(UPGRADE_HOME_PATH, "sdk_upgrade")
    # SDK固件版本信息
    SDK_PACKAGE_VERSION_JSON = "/home/data/config/sdk_version.json"

    # 签名校验动态库
    VERIFY_LIB = "libverify.so"

    # 需要通过驱动获取升级进度的类型
    FIRMWARE_TYPE = "Ascend-firmware"
    MCU_TYPE = "MCU"
    NPU_TYPE = "NPU Driver"
    OMSDK_TYPE = "MindXOM"
    UPGRADE_DRV_PROCESS_TYPE = (MCU_TYPE, NPU_TYPE)

    # G2版本crl
    ASCENDSIP_G2_CRL = "ascendsip_g2.crl"


class UpgradeEffectConstants:
    """
    记录是否允许重启生效：
        0 表示初始状态，允许重启；
        1 表示升级成功后，允许重启；
        2 表示升级失败，不允许重启
    """
    RESTART_DELAY_TIME = 10
    LOG_ROTATE_PATH = "/home/log/kbox_last_logs"
    LOG_COLLECTIONS = ("upgrade.log", "ntp_service.log", "ibma_edge", "manager", "redfish", "web_edge")


class UpgradeState:
    """
    记录升级状态的状态常量与响应描述信息：
        0表示空闲，1表示正在升级，2表示升级失败，3表示升级成功
    """
    UPGRADE_RUNNING_STATE = 1
    UPGRADE_NO_STATE = 0
    UPGRADE_FAILED = 2
    UPGRADE_SUCCESS = 3

    UPGRADE_STATE_MSG = {
        UPGRADE_RUNNING_STATE: "Running",
        UPGRADE_NO_STATE: "New",
        UPGRADE_FAILED: "Failed",
        UPGRADE_SUCCESS: "Success",
    }


class OMUpgradeConstants:
    """OM升级相关常量"""
    # 执行cmd超时时间
    EXEC_CMD_TIMEOUT = 120
    # 升级完成的标记文件
    UPGRADE_FINISHED_FLAG = "/home/data/upgrade_finished_flag"
    # 生效时，将文件内容同步到数据库标记
    CONFIG_TO_DB_FLAG = "/home/data/file_to_db_flag"
    # 切区同步标记
    SYNC_FLAG = "/home/data/om_sync_flag"
    # OM恢复出厂标记
    OM_RESET_FLAG = "/home/data/om_reset_flag"
    # 配置文件路径
    CONFIG_HOME_PATH = "/home/data/config"
    IES_HOME_PATH = "/home/data/ies"
    NGINX_CONFIG_DIR = os.path.join(CONFIG_HOME_PATH, "default")
    OM_ALG_JSON = "om_alg.json"
    OM_CERT_BACKUP_KEYSTORE = "om_cert_backup.keystore"
    OM_CERT_KEYSTORE = "om_cert.keystore"
    SERVER_KMC_CERT = "server_kmc.cert"
    SERVER_KMC_PRIV = "server_kmc.priv"
    SERVER_KMC_PSD = "server_kmc.psd"
    ROOT_CA_CERT = "root_ca.cert"
    ROOT_CA_PRIV = "root_ca.priv"
    ROOT_CA_PSD = "root_ca.psd"
    CLIENT_KMC_CERT = "client_kmc.cert"
    CLIENT_KMC_PRIV = "client_kmc.priv"
    CLIENT_KMC_PSD = "client_kmc.psd"
    IMPORT_CERT_FLAG = "/home/data/cert_backup/cert_flag"
    NGINX_CONF = "nginx.conf"
    ACCESS_CONTROL_INI = "access_control.ini"

    # sys文件路径
    MINDX_OM_ENV_CONF = "mindx_om_env.conf"
    OS_CMD_CONF = "os_cmd.conf"
    OM_INIT_SERVICE = "om-init.service"
    IBMA_EDGE_START_SERVICE = "ibma-edge-start.service"
    PLATFORM_APP_SERVICE = "platform-app.service"
    START_NGINX_SERVICE = "start-nginx.service"
    SYNC_SYS_FILES_SERVICE = "sync-sys-files.service"

    # OM软件需要的磁盘空间大小
    OM_SOFTWARE_SPACE_BYTES = 150 * 1024 * 1024
    # OM配置文件需要的磁盘空间大小
    OM_CONFIG_SPACE_BYTES = 50 * 1024 * 1024

    PROCESSOR_ARCHITECTURE = "ARM"
    MODULE_NAME = "MindXOM"

    # 升级白牌的标记
    WHITEBOX_BACKUP_DIR = "/home/data/WhiteboxConfig"
    OM_WHITEBOX_CONFIG_PATH = "software/nginx/html/manager/WhiteboxConfig"

    OM_EFFECT_SHELL = os.path.join(CommonConstants.OM_UPGRADE_DIR_PATH, "scripts", "copy_om_sys_file.sh")

    # / 为系统主分区挂载目录，/mnt/p2、/mnt/p3为系统备区挂载目录
    SYS_MAIN_PART = "/"
    SYS_BACKUP_P2_PART = "/mnt/p2"
    SYS_BACKUP_P3_PART = "/mnt/p3"
    OM_SYS_PATH = (SYS_MAIN_PART, SYS_BACKUP_P2_PART, SYS_BACKUP_P3_PART)

    COPY_INSTALL = "install"
    COPY_TO_BACK_AREA = "backup_area"
    COPY_UPGRADE = "omsdk_upgrade"
    OM_COPY_TYPE = (COPY_UPGRADE, COPY_TO_BACK_AREA, COPY_INSTALL)

    OM_COPY_PATH = (CommonConstants.OM_UPGRADE_DIR_PATH, CommonConstants.OM_WORK_DIR_PATH)


class FDResultMsg:
    """返回给FD的结果信息"""
    PROCESSING: str = "processing"
    SUCCESS: str = "success"
    FAILED: str = "failed"


class UpgradeResult:
    """记录升级结果的返回码与响应描述信息"""
    ERR_NO_UPGRADE = -1  # 未进行升级
    ERR_UPGRADE_SUCCEED = 0  # 升级成功
    ERR_UPGRADE_CONFLICT = 1  # 正在进行升级，升级冲突
    ERR_UPGRADE_VERIFY_FAILED = 2  # 固件包校验失败
    ERR_CREATE_FILE_FAILED = 3  # 创建升级标记失败
    ERR_UPGRADE_OM = 4  # 升级OM失败
    ERR_UPGRADE_ROOTFS = 6  # OS升级失败
    ERR_UPGRADE_MEF = 7  # 升级mef失败
    ERR_UPGRADE_NPU = 8  # NPU升级失败
    ERR_UPGRADE_MCU = 9  # mcu升级失败
    ERR_DELETE_FILE_FAILED = 10  # 文件删除失败
    ERR_UPGRADE_FILE_NOT_EXIST = 11  # 升级文件不存在
    ERR_UPGRADE_FILE_DECOMPRESS_FAIL = 12  # 文件解压失败
    ERR_PARAM_INVALID = 13  # 参数校验错误
    ERR_EFFECT_CONFLICT = 14  # 正在进行生效
    ERR_SYNC_FILES_FAILED = 15  # 主备同步失败
    ERR_UPGRADE_TOOLBOX_FAILED = 16  # 升级toolbox失败
    ERR_SDK_UPGRADE_FAILED = 17  # 执行SDK通用升级失败
    ERR_M2_UPGRADE_TO_RC1_FAILED = 18  # M.2 FW不能升级到RC1
    ERR_MEF_IS_BUSY = MefBusyConstants.MEF_BUSY_RESPONSE_CODE  # 升级MEF时，MEF服务正忙，提示稍后重试

    UPGRADE_ERROR_CODE_MAP = {
        ERR_NO_UPGRADE: f"ERR.0{ERR_NO_UPGRADE}-Not upgraded",
        ERR_UPGRADE_MCU: f"ERR.0{ERR_UPGRADE_MCU}-Upgrade Mcu failed to upgrade firmware package failed.",
        ERR_UPGRADE_ROOTFS: f"ERR.0{ERR_UPGRADE_ROOTFS}-Upgrade RootFs failed. ",
        ERR_UPGRADE_NPU: f"ERR.0{ERR_UPGRADE_NPU}-Upgrade Minid failed.",
        ERR_UPGRADE_FILE_NOT_EXIST: f"ERR.0{ERR_UPGRADE_FILE_NOT_EXIST}-Upgrade file do not exist.",
        ERR_UPGRADE_FILE_DECOMPRESS_FAIL: f"ERR.0{ERR_UPGRADE_FILE_DECOMPRESS_FAIL}-Upgrade file decompress failed.",
        ERR_UPGRADE_VERIFY_FAILED: f"ERR.0{ERR_UPGRADE_VERIFY_FAILED}-Firmware package verification failed.",
        ERR_UPGRADE_CONFLICT: f"ERR.0{ERR_UPGRADE_CONFLICT}-The upgrade script is being executed.",
        ERR_UPGRADE_SUCCEED: f"ERR.0{ERR_UPGRADE_SUCCEED}-Upgrade all firmware successfully",
        ERR_UPGRADE_OM: f"ERR.0{ERR_UPGRADE_OM}-Upgrade MindXOM failed",
        ERR_DELETE_FILE_FAILED: f"ERR.0{ERR_DELETE_FILE_FAILED}-Delete file failed.",
        ERR_CREATE_FILE_FAILED: f"ERR.0{ERR_CREATE_FILE_FAILED}-Create flag failed.",
        ERR_PARAM_INVALID: f"ERR.0{ERR_PARAM_INVALID}-Param invalid.",
        ERR_EFFECT_CONFLICT: f"ERR.0{ERR_EFFECT_CONFLICT}-Upgrade effect is executing.",
        ERR_SYNC_FILES_FAILED: f"ERR.0{ERR_SYNC_FILES_FAILED}-Sync files failed.",
        ERR_UPGRADE_TOOLBOX_FAILED: f"ERR.0{ERR_UPGRADE_TOOLBOX_FAILED}-Upgrade toolbox failed.",
        ERR_M2_UPGRADE_TO_RC1_FAILED: f"ERR.0{ERR_M2_UPGRADE_TO_RC1_FAILED}-M.2 not support upgrade to RC1",
        ERR_SDK_UPGRADE_FAILED: f"ERR.0{ERR_SDK_UPGRADE_FAILED}-Exec upgrade.sh failed.",
        ERR_MEF_IS_BUSY: f"ERR.0{MefBusyConstants.MEF_BUSY_RESPONSE_CODE}-MEF is busy, please try again later.",
    }


class EffectResult:
    """记录生效结果的返回码与响应描述信息"""
    ERR_EFFECT_SUCCEED = 0  # 生效成功
    ERR_EFFECT_BUSY = 1  # 正在进行生效
    ERR_INVALID_PARAM = 2  # 参数非法
    ERR_UPGRADE_BUSY = 3  # 正在进行升级
    ERR_UPGRADE_FAILED = 4  # 升级失败，不允许生效
    ERR_MIGRATE_DB_FAILED = 5  # 数据库平滑失败
    ERR_BACKLOG_FAILED = 6  # 备份日志失败
    ERR_REGISTER_OM_FILES_FAILED = 7  # 注册OM系统文件失败
    ERR_EFFECT_MEF_FAILED = 8  # 生效mef失败

    EFFECT_CODE_MAP = {
        ERR_EFFECT_SUCCEED: f"ERR.0{ERR_EFFECT_SUCCEED}-Effect firmware success.",
        ERR_EFFECT_BUSY: f"ERR.0{ERR_EFFECT_BUSY}-Upgrade effect is busy.",
        ERR_INVALID_PARAM: f"ERR.0{ERR_INVALID_PARAM}-Request data is invalid.",
        ERR_UPGRADE_BUSY: f"ERR.0{ERR_UPGRADE_BUSY}-Currently upgrading, restart is not allowed.",
        ERR_UPGRADE_FAILED: f"ERR.0{ERR_UPGRADE_FAILED}-Upgrade failed, restart is not allowed.",
        ERR_MIGRATE_DB_FAILED: f"ERR.0{ERR_MIGRATE_DB_FAILED}-Migrate database failed, restart is not allowed.",
        ERR_BACKLOG_FAILED: f"ERR.0{ERR_BACKLOG_FAILED}-Log backup failed.",
        ERR_REGISTER_OM_FILES_FAILED: f"ERR.0{ERR_REGISTER_OM_FILES_FAILED}-Register MindXOM system service failed.",
        ERR_EFFECT_MEF_FAILED: f"ERR.0{ERR_EFFECT_MEF_FAILED}-Effect mef failed.",
    }
