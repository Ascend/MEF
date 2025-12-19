# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import glob
import os
import threading
from typing import NoReturn, Iterable

from common.constants.base_constants import MefBusyConstants, RestoreConfigConstants, CommonConstants
from common.db.base_models import Base
from common.file_utils import FileUtils
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.date_utils import DateUtils
from common.utils.exception_utils import OperateBaseError
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.manager.net_switch_manager import WebNetSwitchManager
from net_manager.models import CertManager
from redfish_db.session import session_maker
from user_manager.models import User, default_pw, HisPwd, EdgeConfig, Session


class RestoreConfigError(OperateBaseError):
    pass


class DefaultConfig:
    MAX_NUMS_OF_FILE = 2
    default_config_lock = threading.Lock()

    @staticmethod
    def check_param(request_dict):
        # 参数是否合法
        if not isinstance(request_dict, dict):
            raise RestoreConfigError(f"ERR.0{RestoreStatusCode.INVALID_PARAM}-param is invalid")

        reserve_ip = request_dict.get("ReserveIP")
        admin_pwd = request_dict.get("Password")
        if not isinstance(reserve_ip, bool) or not isinstance(admin_pwd, str):
            raise RestoreConfigError(f"ERR.0{RestoreStatusCode.INVALID_PARAM}-param is invalid")

    @staticmethod
    def restore_monitor_config(request_data):
        request_data["op_type"] = RestoreConfigConstants.RESTORE_MONITOR
        ret_dict = LibRESTfulAdapter.lib_restful_interface("OMRestoreConfig", "POST", request_data)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            raise RestoreConfigError(f"ERR.0{RestoreStatusCode.OM_RESTORE_FAILED}-{ret_dict.get('message')}")

    @staticmethod
    def restore_mef_config(request_data):
        request_data["op_type"] = RestoreConfigConstants.RESTORE_MEF
        ret_dict = LibRESTfulAdapter.lib_restful_interface("OMRestoreConfig", "POST", request_data)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            if "mef is busy" in ret_dict.get("message"):
                raise RestoreConfigError(f"ERR.0{RestoreStatusCode.MEF_IS_BUSY}-{ret_dict.get('message')}")
            raise RestoreConfigError(f"ERR.0{RestoreStatusCode.MEF_RESTORE_FAILED}-{ret_dict.get('message')}")

    @staticmethod
    def restore_network_config(request_data):
        request_data["op_type"] = RestoreConfigConstants.RESTORE_NETWORK
        ret_dict = LibRESTfulAdapter.lib_restful_interface("OMRestoreConfig", "POST", request_data)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            raise RestoreConfigError(f"ERR.0{RestoreStatusCode.NETWORK_RESTORE_FAILED}-{ret_dict.get('message')}")

    @staticmethod
    def reboot_system():
        ret_dict = LibRESTfulAdapter.lib_restful_interface("Actions", "POST", {"ResetType": "GracefulRestart"})
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            raise RestoreConfigError(f"ERR.0{RestoreStatusCode.REBOOT_FAILED}-{ret_dict.get('message')}")

    @staticmethod
    def restore_redfish_config(session):
        # 删除session
        TokenRestoreDefault.restore_default(session)

        # 恢复默认密码
        PasswordRestoreDefault.restore_default(session)
        HistoryPasswordRestoreDefault.restore_default(session)

        # 重置网管模式，清空FD证书数据库
        FDCertRestoreDefault.restore_default(session)

        # 重置web超时时间、密码有效期
        SessionRestoreDefault.restore_default(session)

    def restore_config(self, request_data):
        with session_maker() as session:
            # 恢复mef默认配置
            self.restore_mef_config(request_data)

            # 恢复redfish层默认配置
            self.restore_redfish_config(session)

            # 恢复monitor层默认配置
            self.restore_monitor_config(request_data)

            # 恢复网卡配置
            self.restore_network_config(request_data)

            # 恢复redfish层默认配置之后清理备份目录
            self.clear_backup_file()

            # 重启系统
            self.reboot_system()

    def clear_backup_file(self):
        for index, file in enumerate(glob.iglob(os.path.join(CommonConstants.REDFISH_BACKUP_DIR, "*"))):
            if index > self.MAX_NUMS_OF_FILE:
                run_log.warning("The number of backup files is greater than %s", self.MAX_NUMS_OF_FILE)
                break
            FileUtils.delete_file_or_link(file)

    def deal_request(self, request_data):
        try:
            # 校验参数
            self.check_param(request_data)
            # 恢复默认配置
            self.restore_config(request_data)
            run_log.info("Restore defaults configuration successfully.")

        except RestoreConfigError as err:
            run_log.error("restore configuration failed: %s", err.err_msg)
            return {"status": AppCommonMethod.ERROR, "message": err.err_msg}

        except Exception as err:
            run_log.error("restore configuration failed: %s", err)
            return {"status": AppCommonMethod.ERROR, "message": f"ERR.0{RestoreStatusCode.UNKNOWN_ERR}-unknown err"}

        return {"status": AppCommonMethod.OK, "message": f"ERR.0{RestoreStatusCode.RESTORE_SUCCESS}-restore success"}


class RedfishRestoreDefaults:
    model: Base

    @classmethod
    def default_instance_generator(cls) -> Iterable[Base]:
        yield cls.model()

    @classmethod
    def save_defaults(cls, session) -> NoReturn:
        session.bulk_save_objects(cls.default_instance_generator())

    @classmethod
    def clear_config(cls, session):
        session.query(cls.model).delete()

    @classmethod
    def restore_default(cls, session):
        cls.clear_config(session)
        cls.save_defaults(session)
        run_log.info("execute %s success", cls.__name__)


class TokenRestoreDefault(RedfishRestoreDefaults):
    model = Session

    @classmethod
    def save_defaults(cls, session):
        pass


class FDCertRestoreDefault(RedfishRestoreDefaults):
    model = CertManager

    @classmethod
    def clear_config(cls, session):
        WebNetSwitchManager({}).switch_deal()
        super().clear_config(session)


class SessionRestoreDefault(RedfishRestoreDefaults):
    model = EdgeConfig


class PasswordRestoreDefault(RedfishRestoreDefaults):
    model = User

    @classmethod
    def default_instance_generator(cls) -> Iterable:
        default_user_id = 1
        modify_time = DateUtils.default_time()
        pw_hash = default_pw()
        yield cls.model(id=default_user_id, pword_hash=pw_hash, pword_modify_time=modify_time)


class HistoryPasswordRestoreDefault(RedfishRestoreDefaults):
    model = HisPwd

    @classmethod
    def default_instance_generator(cls) -> Iterable:
        default_user_id = 1
        modify_time = DateUtils.default_time()
        pw_hash = default_pw()
        yield cls.model(user_id=default_user_id, history_pword_hash=pw_hash, pword_modify_time=modify_time)


class RestoreStatusCode:
    RESTORE_SUCCESS = 10
    INVALID_PARAM = 11
    MEF_RESTORE_FAILED = 12
    OM_RESTORE_FAILED = 13
    NETWORK_RESTORE_FAILED = 14
    REBOOT_FAILED = 15
    UNKNOWN_ERR = 16
    MEF_IS_BUSY = MefBusyConstants.MEF_BUSY_RESPONSE_CODE
