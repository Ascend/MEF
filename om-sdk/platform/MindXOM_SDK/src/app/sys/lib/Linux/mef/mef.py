# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import threading
from typing import List

from common.constants.base_constants import MefNetStatus
from common.constants.base_constants import MefOperate
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result
from lib.Linux.mef.mef_info import MefInfo
from common.common_methods import CommonMethods


class Mef:
    """MEF对接模块"""
    MEF_LOCK = threading.Lock()

    def __init__(self) -> None:
        self.mef_net_status = MefNetStatus.UNKNOWN.value
        self.mef_ca_data = None

    @staticmethod
    def _exec_exchange_root_ca(mef_info: MefInfo) -> Result:
        """交换MEF根证书"""
        ret = mef_info.check_script_file_valid(mef_info.exchange_root_ca_sh)
        if not ret:
            return Result(False, err_msg=f"{mef_info.exchange_root_ca_sh} path invalid: {ret.error}")

        exchange_crt_cmd = (
            mef_info.exchange_root_ca_sh, "--import_path", mef_info.import_ca_path,
            "--export_path", mef_info.export_ca_path
        )
        status, out = ExecCmd.exec_cmd_get_output(exchange_crt_cmd)
        if status != 0:
            return Result(False, err_msg=f"status is {status}")

        return Result(True)

    @staticmethod
    def _exec_restart_mef(mef_info: MefInfo):
        """重启MEF"""
        ret = mef_info.check_script_file_valid(mef_info.run_sh)
        if not ret:
            return Result(False, err_msg=f"{mef_info.run_sh} path invalid: {ret.error}")

        mef_restart_cmd = (mef_info.run_sh, "restart")
        status, out = ExecCmd.exec_cmd_get_output(mef_restart_cmd)
        if status != 0:
            return Result(False, err_msg=f"status is {status}")

        ret = AppCommonMethod.check_service_status_is_active(mef_info.service)
        if not ret:
            return ret

        return Result(True)

    @staticmethod
    def _exec_stop_mef(mef_info: MefInfo):
        """停止MEF"""
        ret = mef_info.check_script_file_valid(mef_info.run_sh)
        if not ret:
            return Result(False, err_msg=f"{mef_info.run_sh} path invalid: {ret.error}")

        mef_stop_cmd = (mef_info.run_sh, "stop")
        status, out = ExecCmd.exec_cmd_get_output(mef_stop_cmd)
        if status != 0:
            return Result(False, err_msg=f"status is {status}")

        return Result(True)

    def get_all_info(self, need_ca_data=False) -> List:
        mef_info = MefInfo()
        self.mef_net_status = mef_info.status
        if need_ca_data:
            self.mef_ca_data = mef_info.ca_data
        return [CommonMethods.OK, "Get all info success."]

    def post_request(self, request_dict):
        """Post请求处理"""
        if Mef.MEF_LOCK.locked():
            run_log.error("Config MEF is busy.")
            return [CommonMethods.ERROR, "Config MEF is busy."]

        with Mef.MEF_LOCK:
            try:
                operate_type = MefOperate(request_dict.get("operate_type"))
            except ValueError:
                run_log.error("invalid parameter: operate_type.")
                return [CommonMethods.ERROR, "operate_type is invalid."]

            mef_info = MefInfo()
            if not mef_info.install_path:
                run_log.error("Get the installation path of MEF failed.")
                return [CommonMethods.ERROR, "get the installation path of MEF failed"]

            if operate_type == MefOperate.EXCHANGE_CRT:
                return self.exchange_root_ca(mef_info)

            elif operate_type == MefOperate.RESTART:
                return self.restart_mef(mef_info)

            elif operate_type == MefOperate.STOP:
                return self.stop_mef(mef_info)

            run_log.error("operate_type is invalid.")
            return [CommonMethods.ERROR, "Operate MEF failed, operate_type is invalid."]

    def exchange_root_ca(self, mef_info: MefInfo) -> List:
        """执行与MEF交换证书的相关操作"""
        ret = self._exec_exchange_root_ca(mef_info)
        if not ret:
            run_log.error("Exchange root ca cert with MEF failed: %s.", ret.error)
            return [CommonMethods.ERROR, "exchange root ca cert with MEF failed"]

        run_log.info("Exchange root ca cert with MEF successfully.")

        ret = self._exec_restart_mef(mef_info)
        if not ret:
            run_log.error("Restart MEF failed: %s.", ret.error)
            return [CommonMethods.ERROR, "restart MEF failed"]

        run_log.info("Restart MEF successfully.")
        return [CommonMethods.OK, ""]

    def restart_mef(self, mef_info: MefInfo) -> List:
        """执行MEF重启操作"""
        ret = self._exec_restart_mef(mef_info)
        if not ret:
            run_log.error("Restart MEF failed: %s.", ret.error)
            return [CommonMethods.ERROR, "restart MEF failed"]

        run_log.info("Restart MEF successfully.")
        return [CommonMethods.OK, ""]

    def stop_mef(self, mef_info: MefInfo) -> List:
        """执行MEF停止操作"""
        ret = self._exec_stop_mef(mef_info)
        if not ret:
            run_log.error("Stop MEF failed: %s.", ret.error)
            return [CommonMethods.ERROR, "stop MEF failed"]

        run_log.info("Stop MEF successfully.")
        return [CommonMethods.OK, ""]
