# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import glob
import os
from pathlib import Path
from typing import NoReturn, Iterable

from bin.environ import Env
from bin.global_exclusive_control import GlobalExclusiveController
from common.constants.base_constants import CommonConstants, MefBusyConstants, RestoreConfigConstants
from common.constants.upgrade_constants import OMUpgradeConstants
from common.db.base_models import Base
from common.file_utils import FileConfusion
from common.file_utils import FileUtils, FileCopy
from common.log.logger import run_log
from common.utils import ability_policy
from common.utils.ability_policy import OmAbility, AbilityConfig
from common.utils.exception_utils import OperateBaseError
from common.utils.exec_cmd import ExecCmd
from lib.Linux.systems.disk.models import MountWhitelistPath
from lib.Linux.systems.nfs.cfg_mgr import query_nfs_configs
from lib.Linux.systems.nfs.models import NfsCfg
from lib.Linux.systems.nfs.nfs_manage import NfsManage
from lib.Linux.systems.security_service.cert_alarm_time import CertAlarmTime
from lib.Linux.systems.security_service.models import LoginRules
from common.common_methods import CommonMethods
from monitor_db.session import session_maker
from om_lib.Linux.om_systems.recover_mini_os import RecoverMiniOS
from om_lib.Linux.om_systems.recover_mini_os import RecoverOSError


class OMRestoreConfigError(OperateBaseError):
    pass


class OMRestoreConfig:
    RESTART_DELAY_TIME = 5
    MAX_NUMS_OF_FILE = 20
    NTP_CONFIG_PATH = "/home/data/ies/NTPEnable.ini"
    SHIELD_ALARM_INI = "/home/data/ies/shield_alarm.ini"
    CERT_DIR_PATH = "/home/data/config/default"
    OM_ALG_JSON_PATH = os.path.join(CERT_DIR_PATH, "om_alg.json")
    WEB_CERT_COLLECTIONS = (
        "om_cert_backup.keystore", "om_cert.keystore", "server_kmc.cert", "server_kmc.priv", "server_kmc.psd",
    )
    RESTORE_IP_SCRIPT_PATH = "/usr/local/scripts/save_netcfg.sh"
    MEF_RESTORE_SCRIPT_PATH = os.path.join(os.path.realpath(CommonConstants.A500_MEF_INSTALL_PATH),
                                           "edge_installer", "script", "restore_config.sh")
    CERT_ALARM_TIME = 10
    LOCK_TIMEOUT = 60
    restore_config_lock = GlobalExclusiveController()

    def restore_cert_alarm_time(self):
        ret = CertAlarmTime().patch_request({"CertAlarmTime": self.CERT_ALARM_TIME})
        if ret[0] != CommonMethods.OK:
            raise OMRestoreConfigError(f"restore cert alarm time failed: {ret[1]}")
        run_log.info("restore cert alarm time success")

    def restore_ntp_config(self):
        FileUtils.delete_file_or_link(self.NTP_CONFIG_PATH)
        run_log.info("delete NTPEnable.ini success")

    def restore_shield_alarm(self):
        with os.fdopen(os.open(self.SHIELD_ALARM_INI, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o640), "w") as f:
            pass
        run_log.info("delete shield_alarm.ini success")

    def restore_monitor_config(self):
        with session_maker() as session:
            # 重置证书告警时间
            self.restore_cert_alarm_time()

            # 关闭NTP使能
            self.restore_ntp_config()

            # 重置告警屏蔽配置
            self.restore_shield_alarm()

            # 清空NFS挂载配置
            NfsRestoreDefault.restore_default(session)

            # 清空挂载白名单
            MountWhitePathRestoreDefault.restore_default(session)

            # 清空登录规则
            LoginRuleRestoreDefault.restore_default(session)

            # 重置加密算法
            self.restore_om_alg()

            # 清空web通信证书
            self.restore_web_cert()

            # 清理备区文件
            self.clear_backup_file()

    def clear_backup_file(self):
        for index, file in enumerate(glob.iglob(os.path.join(CommonConstants.MONITOR_BACKUP_DIR, "*"))):
            if index > self.MAX_NUMS_OF_FILE:
                run_log.warning("The number of backup files is greater than %s", self.MAX_NUMS_OF_FILE)
                break
            FileUtils.delete_file_or_link(file)

    def restore_mef_config(self):
        # MEF能力项关闭，则跳过
        ability_policy.init(AbilityConfig.CONFIG_FILE)
        if not ability_policy.is_allow(OmAbility.MEF_CONFIG):
            run_log.warning("mef ability is disable")
            return

        ret = FileUtils.check_script_file_valid(self.MEF_RESTORE_SCRIPT_PATH, "root", "root")
        if not ret:
            raise OMRestoreConfigError("check mef scripts failed")

        ret_code = ExecCmd.exec_cmd((self.MEF_RESTORE_SCRIPT_PATH,))
        if ret_code == MefBusyConstants.MEF_BUSY_INNER_CODE:
            raise OMRestoreConfigError("mef is busy")
        elif ret_code != 0:
            raise OMRestoreConfigError("restore mef config failed")
        run_log.info("restore mef config success")

    def restore_web_ip(self, reserve_ip):
        if reserve_ip:
            return

        ret = FileUtils.check_script_file_valid(self.RESTORE_IP_SCRIPT_PATH, "root", "root")
        if not ret:
            raise OMRestoreConfigError(f"restore web ip failed, {ret.error}")

        if ExecCmd.exec_cmd((self.RESTORE_IP_SCRIPT_PATH, "default")) != 0:
            raise OMRestoreConfigError("restore web ip failed")
        run_log.info("restore web ip success")

    def restore_web_cert(self):
        if Env().start_from_emmc:
            # eMMC启动的场景，需要从golden分区的om_certs目录恢复证书
            try:
                RecoverMiniOS.mount_golden_dev()
                self._copy_web_certs_from_golden()
            except RecoverOSError as err:
                raise OMRestoreConfigError("copy web certs from p1 failed: {err}") from err
            finally:
                RecoverMiniOS.umount_golden_dev()
        else:
            # m.2启动的场景，直接删除相关文件，重启之后重新生成
            for cert in self.WEB_CERT_COLLECTIONS:
                target_file = os.path.join(self.CERT_DIR_PATH, cert)
                if not os.path.islink(target_file):
                    FileConfusion.confusion_path(target_file)

                FileUtils.delete_file_or_link(target_file)

        # web证书清空后，需要删除用户导入的标志文件
        FileUtils.delete_file_or_link(OMUpgradeConstants.IMPORT_CERT_FLAG)
        run_log.info("restore web cert success")

    def restore_om_alg(self):
        FileUtils.delete_file_or_link(self.OM_ALG_JSON_PATH)
        om_alg_src = os.path.join(CommonConstants.OM_WORK_DIR_PATH, "software/ibma/config/nginx_default/om_alg.json")
        ret = FileCopy.copy_file(om_alg_src, self.OM_ALG_JSON_PATH,
                                 mode=0o600, user=CommonConstants.NGINX_USER, group=CommonConstants.NGINX_USER)
        if not ret:
            raise OMRestoreConfigError(f"copy om_alg.json failed: {ret.error}")
        run_log.info("restore om_alg.json success")

    def post_request(self, request_data):
        if self.restore_config_lock.locked():
            return [CommonMethods.ERROR, "The operation is busy"]
        self.restore_config_lock.acquire(self.LOCK_TIMEOUT)

        try:
            # 校验参数
            if not isinstance(request_data, dict) or not isinstance(request_data.get("ReserveIP"), bool):
                raise OMRestoreConfigError("param is invalid")

            # 恢复默认配置
            if request_data.get("op_type") == RestoreConfigConstants.RESTORE_MEF:
                self.restore_mef_config()
            if request_data.get("op_type") == RestoreConfigConstants.RESTORE_MONITOR:
                self.restore_monitor_config()
            if request_data.get("op_type") == RestoreConfigConstants.RESTORE_NETWORK:
                self.restore_web_ip(request_data.get("ReserveIP"))

        except OMRestoreConfigError as err:
            run_log.error("restore configuration failed: %s", err.err_msg)
            self.restore_config_lock.release()
            return [CommonMethods.ERROR, f"restore config failed: {err.err_msg}"]

        except Exception as err:
            run_log.error("restore configuration failed: %s", err)
            self.restore_config_lock.release()
            return [CommonMethods.ERROR, f"restore config failed: {err}"]

        self.restore_config_lock.release()
        return [CommonMethods.OK, "restore config success"]

    def _copy_web_certs_from_golden(self):
        # 如果设备执行过白牌，不拷贝pki证书
        if os.path.isdir(OMUpgradeConstants.WHITEBOX_BACKUP_DIR):
            run_log.info("The device has been white-labeled, not copy the pki certificate.")
            return

        src_path = "/mnt/p1/om_certs"
        if not os.path.exists(src_path):
            raise RecoverOSError(f"path not exists in p1: {src_path}")

        for file in self.WEB_CERT_COLLECTIONS + (OMUpgradeConstants.OM_ALG_JSON,):
            src_file = os.path.join(src_path, file)
            dest_file = os.path.join(self.CERT_DIR_PATH, file)
            ret = FileCopy.copy_file(src_file, dest_file, mode=0o600, user=CommonConstants.NGINX_USER,
                                     group=CommonConstants.NGINX_USER)
            if not ret:
                raise RecoverOSError(f"copy file {file} from p1 failed: {ret.error}")


class MonitorRestoreDefaults:
    model: Base

    @classmethod
    def default_instance_generator(cls) -> Iterable[Base]:
        yield cls.model()

    @classmethod
    def save_defaults(cls, session) -> NoReturn:
        session.bulk_save_objects(cls.default_instance_generator())

    @classmethod
    def clear_config(cls, session):
        """清空配置"""
        session.query(cls.model).delete()

    @classmethod
    def restore_default(cls, session):
        cls.clear_config(session)
        cls.save_defaults(session)
        run_log.info("restore %s config success", cls.__name__)


class LoginRuleRestoreDefault(MonitorRestoreDefaults):
    model = LoginRules

    @classmethod
    def save_defaults(cls, session) -> NoReturn:
        pass


class NfsRestoreDefault(MonitorRestoreDefaults):
    model = NfsCfg

    @classmethod
    def save_defaults(cls, session) -> NoReturn:
        pass

    @classmethod
    def clear_config(cls, session):
        """清空配置"""
        nfs_manager = NfsManage()
        for cfg in query_nfs_configs():
            nfs_manager.umount_nfs(cfg.local_dir)
            session.query(MountWhitelistPath).filter_by(path=cfg.local_dir).delete()
        super().clear_config(session)


class MountWhitePathRestoreDefault(MonitorRestoreDefaults):
    model = MountWhitelistPath

    @classmethod
    def clear_config(cls, session):
        for white in session.query(cls.model).all():
            if Path(white.path).is_mount():
                continue
            session.delete(white)

    @classmethod
    def save_defaults(cls, session) -> NoReturn:
        pass
