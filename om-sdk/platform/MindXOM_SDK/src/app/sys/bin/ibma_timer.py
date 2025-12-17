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
功    能：定时器类，该类主要涉及定时器的成员定义
"""

import threading
import time
import bin.lib_adapter as adapter
from common.log.logger import run_log


class IbmaTimer(object):
    """
    功能描述：定时器类
    接口：NA
    """
    ALL_TIMERS_INIT_COMPLETE = False

    def __init__(self):
        """
        功能描述：初始化类
        参数：无
        返回值：无
        异常描述：NA
        """
        super(IbmaTimer, self).__init__()
        self.timer = None
        self.isLocal = False
        self.actions = []
        self.name = ""
        self.description = ""
        self.enabled = False
        self.minIntervalTime = 5
        self.maxIntervalTime = 86400
        self.intervalTime = 10
        self.runTimes = 0
        self.errorMsg = ""
        self.interrupt = False
        self.isActive = True
        self.complete = False
        # 初始化是否完成，用于判断全部完成后，才进行iBMC注册
        self.initComplete = False
        # 用于延迟执行操作
        self.delay = False
        # 循环次数，用于控制底层的循环次数，在同一循环周期内缓存数据，降低CPU使用率
        self.CyclesTimes = 1
        # 判断是否第一次启动
        self.FirstRun = True

    @staticmethod
    def set_all_timers_init_complete(flag):
        """
        功能描述：设置定时器是否已经初始化完成
        参数：flag 是否初始化完成
        返回值：无
        异常描述：NA
        """
        IbmaTimer.ALL_TIMERS_INIT_COMPLETE = flag

    def start_timer(self):
        """
        功能描述：启动定时器
        参数：无
        返回值：无
        异常描述：NA
        """
        actions = (self._start_timer, self.dowork)
        try:
            for action in actions:
                action()
        except Exception:
            self.isActive = False
            run_log.error("Timer:%r start failed." % self.name)
        finally:
            self.initComplete = True

    def stop_timer(self):
        """
        功能描述：停止定时器
        参数：无
        返回值：无
        异常描述：NA
        """
        try:
            if self.timer is not None:
                self.interrupt = True
                # 停止定时器
                self.timer.cancel()
        except Exception:
            run_log.error("Timer:%r stop failed." % self.name)

        return

    def judge_effective(self):
        """
        功能描述：判断定时器有效性
        参数：定时器对象
        返回值：是否有效
        异常描述：NA
        """
        # 已经被停止
        if self.interrupt:
            self.errorMsg = "Timer is interrupted."
            return False

        # 没有使能
        if not self.enabled:
            self.errorMsg = "Timer is disabled."
            return False

        # 判断最小间隔时间（不小于1）、最大间隔时间（不大于86400,24小时）是否有效；失效返回 False
        if self.minIntervalTime < 1:
            self.errorMsg = "MinIntervalTime must be greater than 0."
            return False

        if self.maxIntervalTime > 86400:
            self.errorMsg = "MaxIntervalTime must be less than 86400."
            return False

        # 判断间隔时间是否介于最小、最大间隔时间内，包括等于；不满足要求，返回 False
        if self.intervalTime < self.minIntervalTime or self.intervalTime > self.maxIntervalTime:
            self.errorMsg = "IntervalTime must be between minIntervalTime and maxIntervalTime."
            return False

        # 判断执行次数是否有效（大于等于0，小于3000）；不满足要求返回 False
        if self.runTimes < 0:
            self.errorMsg = "Timer is stopped."
            return False

        return True

    def dowork(self):
        """
        功能描述：遍历节点获取数据信息
        参数：无
        返回值：无
        异常描述：NA
        """
        self.complete = False
        action = TimerAction()
        try:
            for index, _ in enumerate(self.actions):
                if self.interrupt:
                    # 被中断
                    break

                action = self.actions[index]
                if action.actionModule == "local":
                    # 检查事件变更类的操作

                    return

                if action.params is not None:
                    # 解析参数列表
                    p_list = adapter.LibAdapter.get_ibma_resources_value(action.params)
                    if not p_list or not isinstance(p_list, list) or len(p_list) == 0:
                        # 没有获取到列表信息，遍历下一个节点
                        continue

                    # 遍历列表，循环获取信息
                    for p_index, _ in enumerate(p_list):
                        if self.interrupt:
                            # 被中断
                            break

                        item1 = p_list[p_index]
                        self.get_resource_info(action.actionModule,
                                               action.actionFuncKey,
                                               action.hasList,
                                               action.parentResourcePath,
                                               action.childResourcePath, item1)
                else:
                    # 没有参数，可以直接进行循环，获取信息
                    self.get_resource_info(action.actionModule,
                                           action.actionFuncKey,
                                           action.hasList,
                                           action.parentResourcePath,
                                           action.childResourcePath)
        except Exception:
            run_log.error("Timer:%r,action:%r run failed." % (self.name, action.actionModule))
            self.stop_timer()
        finally:
            self.complete = True

    def get_resource_info(self, model_name, func_key, need_list=False, p_path=None,
                          c_path=None, item1=None, item2=None, item3=None,
                          item4=None):
        """
        功能描述：遍历获取资源信息
        参数：modelName 模块名称，用于加载具体的对象
             funcKey 函数关键字
             needList 是否是一个列表，True 或 False
             p_path 父资源的路径，如果发生变动时，需要按照父节点来上报
             cPath 子资源的路径，删除资源时，需要同步删除子节点的信息
             item1 第一级资源
             item2 第二级资源
             item3 第三级资源
             item4 第四级资源
        返回值：无
        异常描述：NA
        """
        ret = {"status": adapter.CommonMethods.OK, "message": ""}
        if need_list:
            # 需要先获取列表，然后遍历列表，获取底层资源信息
            ret = adapter.LibAdapter.lib_timer_interface(model_name, "all",
                                                         need_list, p_path, c_path,
                                                         item1, item2, item3,
                                                         item4, self.CyclesTimes)
            items = ret["message"]
            if ret["status"] != adapter.CommonMethods.OK or not isinstance(items, list):
                run_log.error("Timer:%s get list failed, message:%s." % (self.name, ret))
                return

            # 创建一份临时变量，用于判断是否为有效的
            it1 = item1
            it2 = item2
            it3 = item3
            it4 = item4

            for index, _ in enumerate(items):
                if self.interrupt:
                    # 被中断
                    break

                if it1 is None:
                    item1 = items[index]
                elif it2 is None:
                    item2 = items[index]
                elif it3 is None:
                    item3 = items[index]
                elif it4 is None:
                    item4 = items[index]
                else:
                    # 参数过多，不支持
                    run_log.error("Timer:%s has more than 4 parameters." % self.name)
                    return

                # 停留一点时间
                time.sleep(0.1)

                if self.isLocal:
                    ret = adapter.LibAdapter.get_resource_info(model_name,
                                                               func_key,
                                                               False, False,
                                                               item1, item2,
                                                               item3, item4)
                else:
                    ret = adapter.LibAdapter.lib_timer_interface(model_name,
                                                                 func_key, False,
                                                                 p_path, c_path,
                                                                 item1, item2,
                                                                 item3, item4,
                                                                 self.
                                                                 CyclesTimes)
                if ret["status"] != adapter.CommonMethods.OK \
                        and ret["status"] != adapter.CommonMethods.NOT_EXIST:
                    err_msg = "Timer:%s(item1=%s, item2=%s, item3=%s, item4=%s) get information failed, message:%s."
                    run_log.error(err_msg % (self.name, item1, item2, item3, item4, ret))
        else:
            # 直接获取资源信息
            if self.isLocal:
                ret = adapter.LibAdapter.get_resource_info(model_name, func_key, False, False,
                                                           item1, item2, item3, item4)
            else:
                ret = adapter.LibAdapter.lib_timer_interface(model_name, func_key, need_list, p_path,
                                                             c_path, item1, item2, item3, item4,
                                                             self.CyclesTimes)
            if ret["status"] != adapter.CommonMethods.OK and ret["status"] != adapter.CommonMethods.NOT_EXIST:
                err_msg = "Timer:%s(item1=%s, item2=%s, item3=%s, item4=%s) get information failed, message:%s."
                run_log.error(err_msg % (self.name, item1, item2, item3, item4, ret))

    def _start_timer(self):
        if not self.FirstRun:
            time.sleep(self.intervalTime)

        if not self.judge_effective():
            self.isActive = False
            self.complete = True
            # 记录日志并退出循环
            run_log.warning("Timer:%s start failed. Message: %s" % (self.name, self.errorMsg))
            return

        self.timer = threading.Timer(0, self.start_timer)
        self.FirstRun = False
        self.timer.start()

        if not self.complete:
            return

        if not self.initComplete and self.delay:
            # 延迟到下一次执行
            run_log.info("Timer:%s delay to the next execution." % self.name)
            return

        if self.initComplete and not IbmaTimer.ALL_TIMERS_INIT_COMPLETE:
            # 自身已经初始化完成，但是所有的定时器尚未完成的时候，不执行定时任务
            return

        # 将执行的次数 -1，如果执行次数是 0，表示永久生效
        if self.runTimes != 0:
            if self.runTimes == 1:
                self.runTimes = -1
            else:
                self.runTimes -= 1

        # 增加循环次数判断，1~1000
        if self.CyclesTimes % 1000 == 0:
            self.CyclesTimes = 1
        else:
            self.CyclesTimes += 1


class TimerAction(object):
    """
    功能描述：定义动作类
    接口：NA
    """

    def __init__(self):
        """
        功能描述：初始化类
        参数：无
        返回值：无
        异常描述：NA
        """
        super(TimerAction, self).__init__()

        self.actionModule = None
        self.actionFuncKey = None
        self.hasList = False
        self.params = None
        self.description = ""
        # 如果发生变动时，需要按照父节点来上报
        self.parentResourcePath = None
        # 删除资源时，需要同步删除子节点的信息
        self.childResourcePath = None
