# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import functools
from concurrent.futures.thread import ThreadPoolExecutor
from typing import Iterable

from common.utils.result_base import Result

from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from net_manager.manager.fd_cfg_manager import FdCfgManager
from net_manager.manager.fd_cfg_manager import FdMsgData
from wsclient.ws_client_mgr import WsClientMgr

message_thread_pool = ThreadPoolExecutor(max_workers=32)


class FdCommonMethod(object):
    # 是否成功执行过安全配置核查
    sec_cfg_check_done: bool = False

    @staticmethod
    def fd_operational_log(operation_message):

        def wrap(fd_func):
            @functools.wraps(fd_func)
            def func(*args, **kwargs):
                operation_subject = "FD@AtlasEdge"
                restart_type = ""
                restart_exist_valid = isinstance(args, tuple) and len(args) > 0 and isinstance(args[0], dict) \
                                      and args[0].get("restart_method", "") in ["Graceful", "Force", "ColdReset"]
                if operation_message == "Restart system" and restart_exist_valid:
                    restart_type = "(" + args[0].get("restart_method") + ")"
                operation = operation_message + restart_type
                message = f"{operation} failed"
                try:
                    fd_ip = FdCfgManager.get_cur_fd_ip()
                    if not fd_ip:
                        run_log.error("Get fd ip failed.")
                    else:
                        operation_subject = "FD@{}".format(fd_ip)

                    operate_log.info("[%s] %s executing.", operation_subject, operation)

                    result = fd_func(*args, **kwargs)
                    is_valid = (isinstance(result, list) and len(result) > 0 and result[0] == 0) or (
                            isinstance(result, Result) and result)
                    if is_valid:
                        message = f"{operation} successful"
                    return result
                except Exception as err:
                    run_log.error("[%s] %s failed, reason: %s.", operation_subject, operation, err)
                    return [-1, f"{operation_message} failed"]
                finally:
                    operate_log.info("[%s] %s.", operation_subject, message)

            return func

        return wrap

    @staticmethod
    def contains_forbidden_domain(domain_name_list: Iterable[str]):
        fd_server_name = FdCfgManager.get_cur_fd_server_name()
        return any(
            True if domain == fd_server_name or domain in AppCommonMethod.FORBIDDEN_DOMAINS else False
            for domain in domain_name_list
        )


def publish_ws_msg(msg: FdMsgData):
    """通过websocket发布消息"""
    try:
        message_thread_pool.submit(WsClientMgr().send_msg, msg)
    except Exception as ex:
        run_log.error("publish ws msg exception:err_msg=%s, topic=%s", ex, msg.topic)


class MidwareErrCode(object):
    # 系统升级错误
    upgrade_timeout_err_code = 171

    # use 300-400
    midware_parma_error = 300
    midware_import_cert_error = 323
    # use 600-700 for midware
    midware_common_err = 600
    midware_input_not_json = 601
    midware_input_not_permitted = 602
    midware_input_parameter_invalid = 603
    midware_resource_busy = 604
    midware_ability_disable = 605

    # profile
    midware_profile_not_exist = 610

    # ntp 611-620
    midware_config_ntp_common_err = 611

    # dns 626-630
    midware_config_dns_common_err = 626

    # partition 631-639
    midware_partition_dev_not_fount = 631
    midware_partition_is_using = 632
    midware_partition_list_failed = 633
    midware_partition_docker_failed = 634
    midware_partition_out_range = 635
    midware_partition_damaged = 636
    midware_partition_path_not_empty = 637
    midware_partition_fs_not_support = 638

    # firmware 640-649
    midware_firmware_download_err = 640
    midware_firmware_upgrade_err = 641

    # infocollect 650-659
    midware_info_collect_common_err = 650
    midware_info_collect_exec_err = 651
    midware_info_collect_upload_err = 652

    # asset tag 670-675
    midware_assert_tag_common_err = 670

    # password_validity 676-679
    midware_config_passwd_validity_err = 676

    @staticmethod
    def midware_generate_err_msg(err_code, error_info):
        return f"ERR.{AppCommonMethod.convert_err_code_fd_format(err_code)}, {error_info}"
