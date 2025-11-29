# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
import time
from functools import partial
from threading import Thread
from typing import NoReturn

from common.constants.error_code_constants import ErrorCode
from common.file_utils import FileCheck
from common.file_utils import FileCreate
from common.file_utils import FileUtils
from common.utils.result_base import Result
from common.constants.upgrade_constants import UpgradeConstants
from common.log.logger import run_log
from common.schema import AdapterResult
from common.utils.app_common_method import AppCommonMethod
from common.utils.url_downloader import https_download_file
from fd_msg_process.fd_common_methods import FdCommonMethod
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.manager.fd_cfg_manager import FdCfgManager
from upgrade_service.errors import UpgradeError, TimeOutError, BaseError, DownloadError, ExternalParmaError
from upgrade_service.models import UpgradePayload, UpgradeInfo
from upgrade_service.upgrade_report_manager import UpgradeReportManager


class UpgradeService:
    # 上报间隔时间
    REPORT_INTERVAL = 5
    # 最大上报次数, 升级脚本执行超时时间(含失败重试1次+间隔时间15秒)，此处取2倍时间换算的次数，保证等待时间足够
    REPORT_MAX_TIMES = UpgradeConstants.UPGRADE_FIRMWARE_TIMEOUT * 2 // REPORT_INTERVAL

    adapter = partial(LibRESTfulAdapter.lib_restful_interface, "Upgrade_New")
    upgrade_lock = RedfishGlobals.high_risk_exclusive_lock
    payload: UpgradePayload
    reporter: UpgradeReportManager

    def __init__(self, request):
        self.request = request

    @FdCommonMethod.fd_operational_log("Execute upgrade service")
    def handler(self) -> Result:
        if self.upgrade_lock.locked():
            run_log.error("The operation is busy.")
            UpgradeReportManager().report_failed_message(error_code=ErrorCode.midware_resource_busy,
                                                         reason="Upgrade service is busy.")
            return Result(False)

        with self.upgrade_lock:
            try:
                self._validate_external_params()
            except ExternalParmaError as err:
                run_log.error("Execute upgrade service error, %s", err)
                UpgradeReportManager().report_failed_message(
                    error_code=ErrorCode.midware_input_parameter_invalid,
                    reason="External params invalid.")
                return Result(False)

            self.reporter = UpgradeReportManager(self.payload.https_server)
            try:
                ret_dict = LibRESTfulAdapter.lib_restful_interface("ExclusiveStatus", "GET", None, False)
                if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                    raise UpgradeError("System is busy, operate failed.")

                if not isinstance(ret_dict.get("message"), dict) or not \
                        isinstance(ret_dict.get("message").get("system_busy"), bool) or \
                        ret_dict.get("message").get("system_busy"):
                    raise UpgradeError("System is busy, operate failed.")
            except UpgradeError as err:
                run_log.error("Execute upgrade service error: catch %s error", err)
                self.reporter.report_failed_message(reason="System is busy, operate failed.")
                return Result(False)

            try:
                self._execute_upgrade()
            except BaseError as err:
                run_log.error("Execute upgrade service error: %s", err)
                self.reporter.report_failed_message(err.CODE, reason="upgrade error")
                return Result(False)
            except Exception as err:
                run_log.error("Execute upgrade service error: catch %s error", err.__class__.__name__)
                self.reporter.report_failed_message(reason="upgrade error")
                return Result(False)
            else:
                run_log.info("Execute upgrade service success")
                return Result(True)
            finally:
                try:
                    FileUtils.delete_file_or_link(self.payload.https_server.local_filename)
                except Exception as err:
                    run_log.error("Clear upgrade package: %s failed, %s", self.payload.https_server.local_filename, err)

    def _validate_external_params(self):
        try:
            self.payload = UpgradePayload.from_payload(json.loads(self.request))
        except Exception as err:
            raise ExternalParmaError("External params invalid.") from err

    def _execute_upgrade(self) -> NoReturn:
        run_log.info("Start upgrade")
        self._download_software()
        self._firmware_upgrade()
        self._await_upgrade_finish()
        run_log.info("Execute upgrade success")

    def _download_software(self) -> NoReturn:
        downloading_reporter = Thread(target=self.reporter.report_process_while_downloading)
        downloading_reporter.start()
        try:
            self._download_software_image_by_https()
        finally:
            # 将downloading设为False,触发上报循环结束
            self.reporter.downloading = False
            downloading_reporter.join()

    def _download_software_image_by_https(self) -> NoReturn:
        run_log.info("Start download firmware image")
        if not FileCheck.is_exists(self.payload.https_server.download_dir):
            if not FileCreate.create_dir(self.payload.https_server.download_dir, 0o1700):
                raise UpgradeError("Create download dir failed.")

        ret, msg = https_download_file(
            self.payload.https_server.url, self.payload.https_server.user_name,
            self.payload.https_server.password, self.payload.https_server.local_filename,
            self.payload.check_code
        )
        if ret != 0:
            raise DownloadError("Use https download software failed.")

        run_log.info(f"Download software image by https success")

    def _firmware_upgrade(self) -> NoReturn:
        run_log.info("start execute firmware upgrade.")
        file_info = self.payload.https_server.upgrade_request
        file_info['_Xip'] = FdCfgManager.get_cur_fd_ip()
        file_info['_User'] = "FD"
        try:
            result = AdapterResult.from_dict(self.adapter("POST", file_info))
        except Exception as err:
            raise UpgradeError(f"Firmware to upgrade error, catch {err.__class__.__name__} error.") from err

        if result.status != AppCommonMethod.OK:
            raise UpgradeError(result.message)

        run_log.info("Firmware upgrade will run in asynchronous. Now return and monitor!")

    def _await_upgrade_finish(self) -> NoReturn:
        last_percentage = ""
        for _ in range(self.REPORT_MAX_TIMES):
            upgrade_info = self._get_upgrade_process()
            self.reporter.report_upgrade_process(upgrade_info)
            # 相同进度只打一次
            if self.reporter.report_model.percentage != last_percentage:
                last_percentage = self.reporter.report_model.percentage
                run_log.info("upgrading...%s", last_percentage)
            if upgrade_info.percentage == 100:
                return
            time.sleep(self.REPORT_INTERVAL)

        # 超时，上报失败
        raise TimeOutError("firmware upgrade timeout")

    def _get_upgrade_process(self) -> UpgradeInfo:
        try:
            result = AdapterResult.from_dict(self.adapter("GET", None))
        except Exception as err:
            raise UpgradeError(f"Get firmware upgrade progress error, catch {err.__class__.__name__} error.") from err

        if result.status != AppCommonMethod.OK:
            raise UpgradeError(result.message)

        try:
            info: UpgradeInfo = UpgradeInfo.from_dict(result.message)
        except Exception as err:
            raise UpgradeError(f"Get firmware upgrade progress error, catch {err.__class__.__name__} error.") from err

        if info.state == "Failed":
            raise UpgradeError(info.msg)

        return info
