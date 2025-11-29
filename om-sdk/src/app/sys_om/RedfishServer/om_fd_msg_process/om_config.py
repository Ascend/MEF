# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

class OMTopic:
    # dflc生命周期信息写入
    SUB_CONFIG_DFLC = "$hw/edge/v1/hardware/operate/config_dflc"
    # dflc生命周期信息写入的结果
    PUB_CONFIG_DFLC_TO_FD = r"$hw/edge/v1/hub/report/config_dflc_result"
    # 复位主机系统
    SUB_COMPUTER_SYSTEM_RESET = "$hw/edge/v1/hardware/operate/restart"
    # 恢复最小系统
    SUB_RECOVER_MINI_OS = "$hw/edge/v1/hardware/operate/min_recovery"
