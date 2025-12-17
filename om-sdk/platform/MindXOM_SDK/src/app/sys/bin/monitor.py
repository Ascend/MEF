#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

"""
功    能：Monitor类，该类主要涉及定时器的创建，校验，执行
"""
import json
import os
import sys
import threading
import time

import bin.ibma_timer as IbmaTimer
import bin.lib_adapter as LibAdapter
import common.common_methods as CommonMethods
import bin.ibma_server as iBMASocket
from bin.environ import Env
from bin.log_rotate import LogRotate
from bin.monitor_config import SystemSetting, MonitorBackupRestoreCfg
from common.backup_restore_service.backup import Backup
from common.file_utils import FileCheck, FilePermission
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from common.utils.exception_utils import OperationCode
from lib.Linux.EdgeSystem.event import Event
from lib.Linux.systems.disk.partition import Partition
from lib.Linux.systems.nfs.nfs_manage import NfsManage
from monitor_db.backup import DatabaseBackup
from monitor_db.restore import DatabaseRestore
from monitor_kmc.kmc_updater import start_keystore_update_tasks


class Monitor(object):
    """
    功能描述：Monitor类
    接口：NA
    """

    @staticmethod
    def start_socket_server():
        # 打开定时器配置文件
        cfg_file_path = os.path.join(AppCommonMethod.get_project_absolute_path(), "config", "Monitor.json")
        ret = FileCheck.check_path_is_exist_and_valid(cfg_file_path)
        if not ret:
            err_msg = f"{cfg_file_path} path invalid : {ret.error}"
            run_log.error(f"Server initialize failed, {err_msg}")
            return [1, "Server start failed."]

        # 获取配置文件内容
        try:
            with open(cfg_file_path, "r") as file:
                ibma_timers_cfg = json.load(file)
        except Exception as err:
            run_log.error("read ClassMap.json failed: %s", err)
            return [1, "read ClassMap.json failed."]
        LibAdapter.LibAdapter.set_ibma_timers_cfg(ibma_timers_cfg)

        # 初始化资源列表
        LibAdapter.LibAdapter.init_resources()
        try:
            # 初始化资源互斥锁，避免出现不一致的情况
            LibAdapter.LibAdapter.init_resource_lock(ibma_timers_cfg)
        except Exception as err:
            run_log.error("initialize resource lock failed, %s", err)
            return [1, "Server start failed."]

        # 初始化 Socket 服务端
        socket_path = CommonMethods.CommonMethods.get_config_value("iBMA_System", "iBMA_socket_path")

        ret = iBMASocket.Server.init_server(socket_path, LibAdapter.LibAdapter.
                                            lib_socket_call_function)
        FilePermission.set_path_permission(socket_path, 0o750)
        if ret[0] != 0:
            run_log.error("Server start failed, message: %s." % str(ret))
            return ret
        run_log.info("Server start successfully, port [%s]." % socket_path)
        return [0, "Server start completed."]

    @staticmethod
    def start_timer():
        """
        功能描述：创建定时器信息
        参数：无
        返回值：Monitor对象
        异常描述：NA
        """
        ibma_timers = {}
        i_bma_timers_cfg = LibAdapter.LibAdapter.get_ibma_timers_cfg()
        run_log.info("Timer initialize start.")

        try:
            # ClassMap.json中将BeforeTimer删除，故而此处相关调用

            # 遍历定时器配置文件
            for key in i_bma_timers_cfg:
                timer_cfg = i_bma_timers_cfg[key]
                timer = IbmaTimer.IbmaTimer()
                parent_name = Monitor.get_params(timer_cfg, None, "parent", None)
                parent = None
                if parent_name is not None:
                    parent = Monitor.get_params(i_bma_timers_cfg, None, parent_name, None)

                timer.name = key
                timer.delay = Monitor.get_params(timer_cfg, parent, "delay", False)
                timer.description = Monitor.get_params(timer_cfg, parent, "description", "")
                timer.isLocal = Monitor.get_params(timer_cfg, parent, "isLocal", False)
                timer.enabled = Monitor.get_params(timer_cfg, parent, "enabled", False)
                timer.minIntervalTime = Monitor.get_params(timer_cfg, parent, "minIntervalTime", 5)
                timer.maxIntervalTime = Monitor.get_params(timer_cfg, parent, "maxIntervalTime", 86400)
                timer.intervalTime = Monitor.get_params(timer_cfg, parent, "intervalTime", 86400)
                timer.runTimes = Monitor.get_params(timer_cfg, parent, "runTimes", 1)
                action = Monitor.get_params(timer_cfg, None, "action", None)
                if timer.actions is None:
                    timer.actions = Monitor.get_actions(action)
                else:
                    timer.actions.extend(Monitor.get_actions(action))

                ibma_timers[timer.name] = timer

                # 启动定时器，完成的初始值先设置为 True
                timer.complete = True
                tm = threading.Timer(0, timer.start_timer)
                timer.timer = tm
                tm.start()
        except Exception as err:
            run_log.error("Timer initialize failed, %s", err)
            return [1, "Timer initialize failed."]

        # 设置定时器列表
        LibAdapter.LibAdapter.set_ibma_timers(ibma_timers)

        # 等待资源初始化完成后，进行注册
        while True:
            time.sleep(2)
            init_complete = True
            for key in ibma_timers:
                if not ibma_timers.get(key).initComplete:
                    init_complete = False
                    break

            if init_complete:
                break

        # 设置所有的定时器已经完成初始化动作，可以进行下一次遍历
        IbmaTimer.IbmaTimer.set_all_timers_init_complete(True)

        # 启动时初始化Lte init模块
        try:
            from lib.Linux.systems.lte_info import LteInfo
            threading.Thread(target=LteInfo.lte_base_init).start()
        except Exception as err:
            run_log.warning("LteInfo not support. %s", err)

        threading.Thread(target=Partition.init_whitelist).start()
        run_log.info("Timer initialize completed.")
        return [0, "Timer initialize completed."]

    @staticmethod
    def get_actions(cfg_actions):
        """
        功能描述：获取动作列表
        参数：cfgActions 配置的动作
        返回值：动作列表
        异常描述：NA
        """
        actions = []

        if cfg_actions is None:
            return actions

        # 遍历动作清单
        for key in cfg_actions.keys():
            a_list = key.split("_")
            if len(a_list) < 2:
                # 动作的定义类似 module_funckey，如果长度小于两个，认为不合法
                run_log.error("Action key:%s is incorrect.", key)
                continue

            act_cfg = cfg_actions[key]
            timer_action = IbmaTimer.TimerAction()
            timer_action.actionModule = a_list[0]
            timer_action.actionFuncKey = a_list[1]
            timer_action.hasList = Monitor.get_params(act_cfg, None, "hasList", False)
            timer_action.params = Monitor.get_params(act_cfg, None, "params", None)
            timer_action.description = Monitor.get_params(act_cfg, None, "description", "")
            timer_action.parentResourcePath = Monitor.get_params(act_cfg, None, "parentResourcePath", None)
            timer_action.childResourcePath = Monitor.get_params(act_cfg, None, "childResourcePath", None)
            actions.append(timer_action)

        return actions

    @staticmethod
    def get_params(cfg_obj, parent, para_name, default_value=None):
        """
        功能描述：获取配置的值
        参数：cfgObj 配置文件对象
        parent 父项对象
        paraName 配置项节点名称
        default_value 默认值
        返回值：配置的值
        异常描述：NA
        """
        para = cfg_obj.get(para_name, None)
        if para is None:
            if parent is not None:
                return parent.get(para_name, default_value)
            else:
                return default_value
        else:
            if para_name != "enabled" or not para:
                # 属性不是使能状态，或者未使能的情况下，直接返回
                return para

            # 使能情况下，需要查看是否有父节点，如果有，则返回父节点的使能状态
            if parent is not None:
                return parent.get(para_name, False)

            return para

    @staticmethod
    def backup():
        """
        安装启动场景需要等所有服务都起来之后备份一次，初始化时部分配置文件权限不满足要求，
        保证配置文件权限符合要求时能备份到备份目录
        """
        # 首次先检查并备份
        for file_path in MonitorBackupRestoreCfg.BACKUP_FILES:
            Backup(MonitorBackupRestoreCfg.BACKUP_DIR, file_path).entry()


def main():
    # 先实例化环境信息，如果异常则拉不起进程
    try:
        run_log.info("System start from %s", "m.2" if Env().start_from_m2 else "emmc")
    except Exception as err:
        run_log.error("Load environ info failed, because of %s", err)
        return OperationCode.FAILED_OPERATION

    # 需要将hdd的serial_number先存一下
    Event().on_start()

    # 数据库备份与恢复
    DatabaseBackup().entry()
    DatabaseRestore().monitor()

    NfsManage().start_monitor()

    start_socket_server_ret_val = Monitor.start_socket_server()

    sys.stdout.flush()

    # 启动密钥更新任务
    start_keystore_update_tasks()

    # 初始化服务配置，加载必要系统配置失败，monitor进程退出
    try:
        SystemSetting()
    except Exception as err:
        run_log.critical("load OM system settings failed: %s, exit process.", err)
        return OperationCode.FAILED_OPERATION

    # 初始化配置文件
    LibAdapter.LibAdapter.init_resources()

    # 启动日志空间检测服务
    threading.Thread(target=LogRotate().start_logrotate).start()

    # 执行一次备份配置文件操作
    Monitor.backup()

    # 启动 Manager
    from bin.manager import Manager
    try:
        Manager.start_manger()
    except Exception as err:
        run_log.error("Manager runtime exception: %s.", err)

    if 0 == start_socket_server_ret_val[0]:
        run_log.info("Monitor start successfully.")
    else:
        run_log.error("Monitor start failed(%r)." % start_socket_server_ret_val)

    return start_socket_server_ret_val[0]


if __name__ == "__main__":
    # start_socket_server_ret_val[0] 为错误码 0 或 1.
    sys.exit(main())
