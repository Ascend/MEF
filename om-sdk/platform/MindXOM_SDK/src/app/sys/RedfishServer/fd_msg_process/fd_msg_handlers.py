# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
import time
from queue import Queue
from typing import List

from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from fd_msg_process.config import Topic
from fd_msg_process.fd_common_methods import FdCommonMethod
from fd_msg_process.fd_common_methods import message_thread_pool, publish_ws_msg
from fd_msg_process.midware_route import MidwareRoute
from fd_msg_process.om_msg_queue import fd_message_que, hisec_event_message_que
from common.utils.result_base import Result
from mef_msg_process.msg_que import msg_que_to_mef, msg_que_from_mef
from mef_msg_process.mef_proc import MefProc
from net_manager.manager.fd_cfg_manager import FdMsgData
from upgrade_service.upgrade_entry import UpgradeService
from wsclient.connect_status import FdConnectStatus
from wsclient.ws_client_mgr import WsClientMgr


class FDMessageHandler:
    NTP_SYNC_MAX_RETRY_CNT = 10

    @staticmethod
    def check_ack_response(msg: FdMsgData) -> bool:
        """
        是否为ACK应答消息类型
        :return: True: ACK应答消息，False:非ACK应答消息
        """
        return msg.topic == Topic.SUB_OPERATE_PROFILE

    @staticmethod
    def handle_config_net_manager_task(msg_obj: FdMsgData):
        run_log.info("Receive topic: %s", msg_obj.topic)

        ret = FDMessageHandler.call_view_func(msg_obj.deal_func_name, msg_obj)
        ret_valid = isinstance(ret.data, list) and len(ret.data) == 2 and ret.data[0] == 0
        if not ret_valid:
            run_log.error("func deal failed: %s, %s", msg_obj.deal_func_name, ret.error)
            payload_failed = {
                "topic": "netmanager", "percentage": "0%", "result": "failed",
                "reason": AppCommonMethod.COM_ERR_REASON_TO_FD
            }
            pub_msg = FdMsgData.gen_ws_msg_obj(payload_failed, msg_obj.up_resource)
            ret = WsClientMgr().send_msg(pub_msg)
            if not ret:
                run_log.error("publish msg failed: %s", ret.error)
            return [-1, f"call_view_func failed: {ret.error}"]

        # 同步ntp任务失败不影响纳管流程
        message_thread_pool.submit(FDMessageHandler.handle_ntp_sync_task)

        pub_msg = FdMsgData.gen_ws_msg_obj(ret.data[1], msg_obj.up_resource)
        WsClientMgr().send_msg(pub_msg)

        from wsclient.ws_monitor import WsMonitor
        WsMonitor.stop_fd_connect_monitor()
        FdConnectStatus().trans_to_connecting()
        time.sleep(15)
        # 停止之前的所有协程任务
        run_log.info("Try to stop current loops when configuring netmanager resource.")
        WsClientMgr().stop_current_fd_loops()
        MefProc.stop_mef_connect_timer()

        # 清空om、mef消息队列
        FDMessageHandler.clear_fd_msg_queue([msg_que_to_mef, msg_que_from_mef, fd_message_que, hisec_event_message_que])
        WsMonitor.start_fd_connect_monitor()
        # 切换FD后，重新与MEF建立连接，保证MEF获取的是新的FD地址
        MefProc.start_mef_connect_timer()
        run_log.info("Done topic: %s", msg_obj.topic)
        return [0, ""]

    @staticmethod
    def handle_msg_from_fd_by_mqtt(msg: FdMsgData):
        handle_topic_func_map = {
            Topic.SUB_RESET_ALARM: Topic.SUB_RESET_ALARM,
            Topic.SUB_OPERATE_PROFILE: "espmanager/SysConfig",
            Topic.SUB_CONFIG_HOST_NAME: "espmanager/Hostname",
            Topic.SUB_PASS_THROUGH_ACCOUNT_MODIFY: "espmanager/passthrough/account_modify",
            Topic.SUB_HARDWARE_OPERATE_TAG: "espmanager/SysAssetTag",
            Topic.SUB_CERT_QUERY: "espmanager/cert_query",
            Topic.SUB_CERT_UPDATE: "espmanager/cert_update",
            Topic.SUB_CRL_UPDATE: "espmanager/crl_update",
            Topic.SUB_CERT_DELETE: "espmanager/cert_delete",
            Topic.SUB_COMPUTER_SYSTEM_RESET: FDMessageHandler.handle_restart_operation_task,
            Topic.SUB_OPERATE_PROFILE_EFFECT: FDMessageHandler.handle_profile_effect_task,
            Topic.SUB_OPERATE_INFO_COLLECT: FDMessageHandler.handle_info_collect_progress_task,
            Topic.SUB_OPERATE_FIRMWARE_EFFECTIVE: FDMessageHandler.handle_firmware_effective_task
        }
        run_log.info("Receive topic: %s", msg.topic)

        if msg.topic in (
                Topic.SUB_RESET_ALARM,
                Topic.SUB_CONFIG_HOST_NAME,
                Topic.SUB_OPERATE_PROFILE,
                Topic.SUB_HARDWARE_OPERATE_TAG,
                Topic.SUB_PASS_THROUGH_ACCOUNT_MODIFY,
                Topic.SUB_CERT_QUERY,
                Topic.SUB_CERT_UPDATE,
                Topic.SUB_CRL_UPDATE,
                Topic.SUB_CERT_DELETE,
        ):
            FDMessageHandler.call_func_and_report(handle_topic_func_map.get(msg.topic), msg)
        elif msg.topic in (Topic.SUB_COMPUTER_SYSTEM_RESET,
                           Topic.SUB_OPERATE_PROFILE_EFFECT,
                           Topic.SUB_OPERATE_INFO_COLLECT,
                           Topic.SUB_OPERATE_FIRMWARE_EFFECTIVE):
            handle_func_name = handle_topic_func_map.get(msg.topic)
            try:
                message_thread_pool.submit(handle_func_name, msg)
            except Exception as ex:
                run_log.error("Exception:err_msg=%s,topic:%s", ex, msg.topic)
                return

        run_log.info("Done topic: %s", msg.topic)

    @staticmethod
    def handle_upgrade_service(msg: FdMsgData):
        message_thread_pool.submit(UpgradeService(msg.payload).handler)

    @staticmethod
    def call_view_func(topic, payload_dict):
        try:
            res = MidwareRoute.view_functions[topic](payload_dict)
        except Exception as ex:
            run_log.error("Exception:err_msg=%s,topic:%s", ex, topic)
            return Result(result=False, err_msg=f"Exception:err_msg={ex},topic:{topic}")
        return Result(True, data=res)

    @staticmethod
    def handle_restart_operation_task(msg: FdMsgData):
        res = MidwareRoute.view_functions["espmanager/ComputerSystemReset"](msg.content)
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(res[1], msg.up_resource)
        publish_ws_msg(resp_msg_obj)

    @staticmethod
    def handle_profile_effect_task(msg: FdMsgData):
        res = MidwareRoute.view_functions["espmanager/SysConfigEffect"](msg.content)
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(res[1], msg.up_resource)
        publish_ws_msg(resp_msg_obj)
        try:
            # 配置核查生效成功之后，立即触发更新系统信息任务，并上报FD
            if res[0] == 0 and FdCommonMethod.sec_cfg_check_done:
                sys_info = MidwareRoute.view_functions["espmanager/SysInfo"]()[1]
                msg: FdMsgData = FdMsgData.gen_ws_msg_obj(sys_info, "websocket/sys_info")
                publish_ws_msg(msg)
        finally:
            FdCommonMethod.sec_cfg_check_done = False

    @staticmethod
    def handle_reset_alarm_from_fd(msg):
        try:
            payload_dict = json.loads(msg.payload)
        except Exception as err:
            run_log.error("Exception:err_msg=%s,topic:%s", err, msg.topic)
            return

        run_log.info("Receive topic: %s", msg.topic)
        ret = FDMessageHandler.call_view_func(msg.topic, payload_dict)
        if not ret:
            return
        run_log.info("Done topic: %s", msg.topic)

    @staticmethod
    def handle_info_collect_progress_task(msg: FdMsgData):
        res = MidwareRoute.view_functions["espmanager/InfoCollect"](msg.content)
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(res[1], msg.up_resource)
        publish_ws_msg(resp_msg_obj)

    @staticmethod
    def handle_firmware_effective_task(msg: FdMsgData):
        res = MidwareRoute.view_functions["espmanager/UpdateService/FirmwareEffective"](msg.content)
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(res[1], msg.up_resource)
        publish_ws_msg(resp_msg_obj)

    @staticmethod
    def handle_ntp_sync_task():
        count = 1
        while count <= FDMessageHandler.NTP_SYNC_MAX_RETRY_CNT:
            from fd_msg_process.midware_urls import MidwareUris
            ret = MidwareUris.snyc_ntp_config()
            if ret[0] == 0:
                run_log.info("Times of set ntp is %s, fd set ntp is succeed", count)
                return

            run_log.warning("Times of set ntp is %s, fd set ntp is failed: %s", count, ret[1])
            time.sleep(5)
            count += 1

    @staticmethod
    def call_func_and_report(handle_func_name, msg):
        if not handle_func_name:
            run_log.warning("Topic %s is wrong", msg.topic)
            return

        ret = FDMessageHandler.call_view_func(handle_func_name, msg.content)
        ret_valid = ret and ret.data and isinstance(ret.data, list) and len(ret.data) == 2
        if not ret_valid:
            run_log.error("result of deal func is invalid")

        if not msg.up_resource:
            # 不需要上报消息处理结果
            return

        parent_msg_id = msg.header.msg_id if FDMessageHandler.check_ack_response(msg) else ""
        resp_msg_obj = FdMsgData.gen_ws_msg_obj(ret.data[1], msg.up_resource, parent_msg_id)
        publish_ws_msg(resp_msg_obj)

    @staticmethod
    def clear_fd_msg_queue(msg_queues: List[Queue]):
        for msg_queue in msg_queues:
            if not msg_queue or msg_queue.empty():
                continue
            msg_queue.queue.clear()
        run_log.info("Clear fd msg queue success.")
