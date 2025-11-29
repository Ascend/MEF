# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
from om_fd_msg_process.om_config import OMTopic
from om_fd_msg_process.om_fd_msg_handlers import OMFDMessageHandler

OM_MSG_HANDLING_MAPPING = {
    OMTopic.SUB_CONFIG_DFLC: OMFDMessageHandler.handle_msg_config_dflc,
    OMTopic.SUB_COMPUTER_SYSTEM_RESET: OMFDMessageHandler.handle_computer_system_reset_msg_from_fd_by_mqtt,
    OMTopic.SUB_RECOVER_MINI_OS: OMFDMessageHandler.handle_recover_mini_os,
}
