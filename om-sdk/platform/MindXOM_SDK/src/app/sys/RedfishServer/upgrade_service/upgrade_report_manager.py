# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import time
from concurrent.futures import ThreadPoolExecutor
from typing import Optional

from common.constants.error_code_constants import ErrorCode
from common.constants.upgrade_constants import FDResultMsg
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from fd_msg_process.config import Topic
from net_manager.manager.fd_cfg_manager import FdMsgData
from upgrade_service.models import Report, HttpsServer, UpgradeInfo
from wsclient.ws_client_mgr import WsClientMgr

upgrade_thread_pool = ThreadPoolExecutor(max_workers=4)


class UpgradeReportManager:
    TOPIC = Topic.PUB_REPORT_UPGRADE_PROGRESS

    def __init__(self, https_server: Optional[HttpsServer] = None):
        self.https_server = https_server
        self.downloading = True
        self.report_model = Report(name=https_server.filename.split(".")[0]) if https_server else Report()

    def report_failed_message(self, error_code: int = ErrorCode.midware_firmware_upgrade_err, reason: str = ""):
        self.report_model.result = FDResultMsg.FAILED
        self.report_model.reason = f"ERR.{AppCommonMethod.convert_err_code_fd_format(error_code)}, {reason}"
        self._publish_msg()

    def report_upgrade_process(self, upgrade_info: UpgradeInfo):
        self.report_model.update_by_upgrade_info(upgrade_info)
        self._publish_msg()

    def report_process_while_downloading(self, interval: int = 15):
        """下载期间间隔interval上报进度为固定30%。"""
        while self.downloading:
            time.sleep(interval)
            self.report_model.percentage = "30%"
            self._publish_msg()

    def _publish_msg(self):
        msg = FdMsgData.gen_ws_msg_obj(self.report_model.to_dict(), "websocket/upgrade_progress")
        try:
            upgrade_thread_pool.submit(WsClientMgr().send_msg, msg)
        except Exception as ex:
            run_log.error("publish ws msg exception:err_msg=%s, topic=%s", ex, msg.topic)
