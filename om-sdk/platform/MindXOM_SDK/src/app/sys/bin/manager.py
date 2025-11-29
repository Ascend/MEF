#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Manger类，该类主要涉及进程管理.
修改记录：2016-11-03 创建
"""
import signal
import os
import time

from bin.lib_adapter import LibAdapter as LibAdapter
from common.constants.base_constants import CommonConstants
from common.file_utils import FileUtils
from common.log.logger import run_log


class Manager(object):
    """
    功能描述：系统 Manager 类.
    接口：NA
    修改记录：
    2016-12-14 创建
    """

    # Manager 继续运行的标记.
    keepRunning = False

    # Manager 退出的标记.
    doExit = False

    # 保存Monitor/Redfish/启动的句柄.
    monitorProcess = None
    redfishProcess = None

    # 保存Monitor/Redfish的文件路径.
    monitorpath = ""
    redfishPath = ""

    # 解释器路径.
    interpreter = None

    # 提供给扫描线程的PID列表.
    pid_list = [0, 0]

    # 是否需要重启被监控的进程.
    taskList = []
    # 是否使用子线程等待输出.
    startChildThread = True
    # 用来保存查到进程的结果.
    startResult = False

    @staticmethod
    def stop_handler(sig=None, frame=None):
        """
        功能描述：进程退出的处理函数,
        用于关闭采样线程. 正在升级的时候不做处理.
        参数：sig 信号值, frame 栈帧. (暂时都未使用)
        返回值：无.
        异常描述：NA
        修改记录：2016-12-14 创建
        """
        # 其他情况, 执行退出操作.
        run_log.info("Manager is ready to exit.")
        try:
            FileUtils.delete_file_or_link(CommonConstants.IBMA_SOCK_PATH)
        except Exception as err:
            run_log.warning("remove ibma socket path failed: %s", err)

        # 强制自己退出. 发送信号KILL.
        Manager.doExit = True
        os.kill(os.getpid(), signal.SIGKILL)

    @staticmethod
    def manager_trace_process():
        """
        功能描述：监视进程运行状态, 拉起和杀死 Monitor/Redfish.
        参数：NA.
        返回值：NA.
        异常描述：NA
        修改记录：
        2017-01-05 创建;
        2017-05-20:如果Monitor在运行, 并且已经注册的时候, 检查HTTP通道;
        """
        while Manager.keepRunning:
            # 这个实际上触发条件是：当检测到内存和CPU的占用率过高时，
            # 就会将Moniter和redfish都杀死然后重新拉起来
            # 当检测机制删除时，不会触发
            # 正常情况下5秒轮询一次进程运行状态.
            time.sleep(5)
            # 根据环境变量 重新初始化时间相关设置,保证时区更改后日志时间正常。
            time.tzset()

        # 进程退出.
        run_log.info("Trace process done.")

    @staticmethod
    def start_manger():
        """
        功能描述：启动 Manager.
        参数：无.
        返回值：无
        异常描述：NA
        修改记录：2016-12-14 创建
                2017-01-09 根据检视意见, 将起进程和扫描的放在两个线程里面.
        """
        # 初始化 LibAdapter.
        LibAdapter.init_resources()

        # 设置启动标记.
        Manager.keepRunning = True

        # 注册退出信号的中断处理函数. service 脚本使用 SIGTERM 杀死进程.
        signal.signal(signal.SIGINT, Manager.stop_handler)
        signal.signal(signal.SIGTERM, Manager.stop_handler)

        # 监控进程运行状态.
        Manager.manager_trace_process()
        # 等待退出标记, 在Windows才用, 如果接收到信号,
        # 在 stop_handler 就会kill掉所有进程.
        while not Manager.doExit:
            time.sleep(0.1)
