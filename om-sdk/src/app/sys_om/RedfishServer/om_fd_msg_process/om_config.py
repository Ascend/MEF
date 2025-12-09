# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

class OMTopic:
    # dflc生命周期信息写入
    SUB_CONFIG_DFLC = "$hw/edge/v1/hardware/operate/config_dflc"
    # dflc生命周期信息写入的结果
    PUB_CONFIG_DFLC_TO_FD = r"$hw/edge/v1/hub/report/config_dflc_result"
    # 复位主机系统
    SUB_COMPUTER_SYSTEM_RESET = "$hw/edge/v1/hardware/operate/restart"
    # 恢复最小系统
    SUB_RECOVER_MINI_OS = "$hw/edge/v1/hardware/operate/min_recovery"
