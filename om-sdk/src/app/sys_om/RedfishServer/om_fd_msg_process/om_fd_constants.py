# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from om_fd_msg_process.om_config import OMTopic
from om_fd_msg_process.om_fd_msg_handlers import OMFDMessageHandler

OM_MSG_HANDLING_MAPPING = {
    OMTopic.SUB_CONFIG_DFLC: OMFDMessageHandler.handle_msg_config_dflc,
    OMTopic.SUB_COMPUTER_SYSTEM_RESET: OMFDMessageHandler.handle_computer_system_reset_msg_from_fd_by_mqtt,
    OMTopic.SUB_RECOVER_MINI_OS: OMFDMessageHandler.handle_recover_mini_os,
}
