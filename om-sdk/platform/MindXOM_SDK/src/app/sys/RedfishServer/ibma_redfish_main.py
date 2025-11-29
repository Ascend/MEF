#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Redfish Server主程序，提供httpServer启动参数解析运行，资源对象初始化功能
修改记录：2016-10-18 创建
"""
import importlib
import os
import signal
import sys
import threading
import time

from common.ResourceDefV1.service_versions import RfServiceVersions
from common.backup_restore_service.backup import Backup
from common.constants.base_constants import CommonConstants
from common.constants.product_constants import SERVICE_ROOT
from common.log.logger import operate_log
from common.log.logger import run_log
from common.utils import ability_policy
from common.utils.ability_policy import AbilityConfig
from common.utils.app_common_method import AppCommonMethod
from common.utils.timer import RepeatingTimer
from extend_interfaces import EXTEND_COMPONENTS_INFO_FUNC_PATHS
from fd_msg_process.fd_add_route import add_midware_route
from fd_msg_process.midware_proc import MidwareProc
from fd_msg_process.midware_route import MidwareRoute
from ibma_redfish_globals import RedfishGlobals
from redfish_kmc.kmc_updater import start_keystore_update_tasks
from lib_restful_adapter import LibRESTfulAdapter
from mef_msg_process.mef_proc import MefProc
from net_manager.manager.fd_cfg_manager import FdCfgManager
from redfish_database_monitor import RedfishDatabaseMonitor
from redfish_db.register import register_models
from redfish_db.session import database
from register_redfish_funcs.register_redfish_funcs import start_om_extend_funcs
from upload_mark_file import UploadMarkFile
from wsclient.ws_monitor import WsMonitor


class RedfishMain:
    """
    功能描述：HTTP服务的主入口函数
    接口：NA
    修改记录：2016-10-31 创建
    """

    @staticmethod
    def rf_start_server(rf_host, rf_port, rf_profile):
        """
        功能描述：启动httpServer
        参数：rfHost 主机IP
        rfPort 主机端口
        rfUser 用户名
        rfProfile JSON模板路径
        返回值：无
        异常描述：NA
        修改记录：2016-10-18 创建
        """
        if rf_profile == "ResourceDefV1":

            # 获取工程目录；RedfishServer独立进程，目录分割
            project_dir = AppCommonMethod.get_project_absolute_path()
            profile_mockup_path = os.path.join(project_dir, "common/MockupData/iBMAServerV1")

            versions = RfServiceVersions(profile_mockup_path, "redfish")
            from ibma_redfish_urls import RedfishURIs
            RedfishURIs.rf_api_ibma_server(SERVICE_ROOT, versions, host=rf_host, port=rf_port, silence=False)

        else:
            run_log.error("invalid profile")
            sys.stdout.flush()
            raise ValueError("invalid profile")

    @staticmethod
    def stop_handler(sig=None, frame=None):
        """
        功能描述：进程退出的处理函数,
        用于关闭采样线程. 正在升级的时候不做处理.
        参数：sig 信号值, frame 栈帧. (暂时都未使用)
        返回值：无.
        异常描述：NA
        修改记录：2022-05-25 创建
        """
        os.kill(os.getpid(), signal.SIGKILL)

    @staticmethod
    def start_monitor_timer():
        """
        功能描述：定时启动Monitor定时器和心跳
        """
        start_timer_success = False
        while True:
            if not start_timer_success:
                ret = LibRESTfulAdapter.start_timer()
                if not isinstance(ret, list) or 0 != ret[0]:
                    run_log.error(f"Run start_timer failed.(ret:{ret})")
                else:
                    start_timer_success = True
            if start_timer_success:
                break

            time.sleep(5)

    @staticmethod
    def config_db():
        """
        注册表，redfish进程启动过程中使用统一的clear_data_on_start处理数据；
        该步骤不能放在session模块，因为session可能被其他脚本（进程调用）
        """
        register_models()
        database.clear_data_on_start()

    @staticmethod
    def start_check_user_lock():
        """
        query user info;
        update user lock status
        record operate log
        """
        try:
            from user_manager.user_manager import UserManager, EdgeConfigManage, SessionManager
            user_one = UserManager.find_user_by_id(1)
            # 锁定时间300S，每60s查询更新一次
            if user_one.lock_state:
                edge_config = EdgeConfigManage.find_edge_config()
                # 判断系统是否重启过, 是的话清零锁定时间
                if int(time.perf_counter()) < user_one.start_lock_time:
                    column_map = {'start_lock_time': 0}
                    UserManager.update_user_specify_column(1, column_map)
                # 校验用户是否到达解锁状态
                if int(time.perf_counter()) - user_one.start_lock_time > edge_config.lock_duration:
                    # 解锁用户的锁定状态
                    column_map = {'lock_state': False, 'start_lock_time': 0, 'pword_wrong_times': 0}
                    UserManager.update_user_specify_column(1, column_map)
                    run_log.info("The user is unlocked.")
                    operate_log.info("[%s@%s] %s" % (user_one.username_db, "LOCAL", "The user is unlocked."))
        except Exception as ex:
            run_log.error(f"check user lock failed {ex}")

    @staticmethod
    def check_and_delete_overtime_session():
        """
        功能描述：每30s检查会话是否超时，超时后，主动删除会话
        """
        try:
            from user_manager.user_manager import SessionManager

            if SessionManager.delete_timeout_session():
                run_log.warning("The session has expired and been deleted.")

        except Exception as err:
            run_log.error("check and delete overtime session failed %s", err)

    @staticmethod
    def start_database_monitor_task():
        db_monitor_instance = RedfishDatabaseMonitor()
        while True:
            try:
                db_monitor_instance.monitor_database_status()
            except Exception as err:
                run_log.error("Data base monitor failed: caught exception: %s", err)

            time.sleep(60)

    @staticmethod
    def do_work(do_list, args=None):
        try:
            # 根据环境变量 重新初始化时间相关设置，保证时区更改后日志时间正常。
            time.tzset()
            for to_do in do_list:
                if to_do not in MidwareRoute.view_functions:
                    run_log.error("%s not found in view_functions", to_do)
                    continue
                if args:
                    MidwareRoute.view_functions[to_do](args)
                else:
                    MidwareRoute.view_functions[to_do]()
        except Exception as ex:
            run_log.error("do_work Caught exception: %s", ex)

    @staticmethod
    def mid_ware_task():
        do_list60s = [r'espmanager/SysStatusProc']

        RepeatingTimer(60, RedfishMain.do_work, args=(do_list60s, )).start()

        do_list120s = [r'espmanager/SysInfoProc']
        extend_funcs = []
        try:
            for extend_thread_func_path in EXTEND_COMPONENTS_INFO_FUNC_PATHS:
                func_index = extend_thread_func_path.rfind(".")
                func_name = extend_thread_func_path[func_index + 1:]
                model_path = extend_thread_func_path[:func_index]
                model = importlib.import_module(model_path)
                if not model or not hasattr(model, func_name):
                    run_log.warning("register %s failed. reason: model or function not exists", func_name)
                    continue
                extend_register_func = getattr(model, func_name)
                extend_funcs.append(extend_register_func)
        except Exception as err:
            run_log.warning("register sys info func failed. reason: %s", err)

        RepeatingTimer(120, RedfishMain.do_work, args=(do_list120s, extend_funcs)).start()

    @staticmethod
    def init_ability():
        try:
            ability_policy.init(AbilityConfig.CONFIG_FILE)
        except Exception as err:
            run_log.error("ability policy init failed, find exception: %s", err)
            raise err

    @staticmethod
    def wait_for_ibma_socket_ready(socket_path=CommonConstants.IBMA_SOCK_PATH):
        """等待与Monitor通信的socket就绪，最多等待1分钟"""
        for _ in range(60):
            time.sleep(1)
            if not os.path.exists(os.path.realpath(socket_path)):
                continue

            ret_dict = LibRESTfulAdapter.lib_restful_interface("System", "GET", None, False)
            if AppCommonMethod.check_status_is_ok(ret_dict):
                break

    @staticmethod
    def start_connect_tasks():
        """启动对接任务"""
        run_log.info("start monitor tasks begin")
        RedfishMain.wait_for_ibma_socket_ready()
        RedfishMain.mid_ware_task()
        threading.Thread(target=RedfishMain.start_monitor_timer).start()
        threading.Thread(target=MidwareProc.dispatch_fd_messages).start()
        RepeatingTimer(5, MidwareProc.start_hisec_event_task).start()
        threading.Thread(target=MidwareProc.pop_and_dispatch_event_task).start()
        # 启动fd对接监控任务
        WsMonitor.start_fd_connect_monitor()
        MefProc.start_mef_connect_timer()
        # 定期检查证书是否即将过期
        FdCfgManager().start_cert_status_monitor()
        run_log.info("start monitor tasks done")

    @staticmethod
    def backup():
        Backup(CommonConstants.REDFISH_BACKUP_DIR, CommonConstants.REDFISH_EDGE_DB_FILE_PATH).entry()

    @staticmethod
    def main():
        """
        功能描述：主入口函数
        参数：无
        返回值：无
        异常描述：NA
        修改记录：2016-10-18 创建
        """
        # 注册退出信号的中断处理函数. systemctl stop 服务时可杀死进程.
        signal.signal(signal.SIGINT, RedfishMain.stop_handler)
        signal.signal(signal.SIGTERM, RedfishMain.stop_handler)

        # 能力项初始化
        RedfishMain.init_ability()

        add_midware_route()

        threading.Thread(target=RedfishMain.start_database_monitor_task).start()

        # 启动Redfish扩展功能
        start_om_extend_funcs()

        rf_profile = "ResourceDefV1"

        RedfishGlobals.init_http_server_param()

        # 启动密钥更新任务
        start_keystore_update_tasks()

        # HtteServer启动参数默认按照配置文件，支持命令行指定

        rf_host = '127.0.0.1'
        rf_port = RedfishGlobals.get_http_port()
        try:
            RedfishMain.config_db()
            UploadMarkFile.clear_upload_mark_file_all()
            # 执行循环备份
            RepeatingTimer(10, RedfishMain.backup).start()
            # start check user lock
            RepeatingTimer(60, RedfishMain.start_check_user_lock).start()
            # 定期清除过期会话
            RepeatingTimer(30, RedfishMain.check_and_delete_overtime_session).start()
            # 检查本次启动是否由恢复最小系统触发
            FdCfgManager.restore_mini_os_config()
            # 检查是否需要重置数据库纳管状态
            FdCfgManager.check_and_reset_status()
            # 启动需要与Monitor通信的相关任务
            threading.Thread(target=RedfishMain.start_connect_tasks).start()
            RedfishMain.rf_start_server(rf_host, rf_port, rf_profile)
            sys.stdout.flush()
        except Exception as ex:
            run_log.error("HTTP server exception.")
            sys.stdout.flush()
            raise Exception from ex


if __name__ == "__main__":
    try:
        RedfishMain.main()
    except Exception:
        sys.exit(2)
