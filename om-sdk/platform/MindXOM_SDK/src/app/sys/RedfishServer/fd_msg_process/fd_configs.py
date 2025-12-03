# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import importlib
from typing import Dict

from common.log.logger import run_log
from fd_extend_interfaces import EXTEND_FD_TOPIC_HANDLER_MAPPING_PATH
from fd_msg_process.config import Topic
from fd_msg_process.fd_msg_handlers import FDMessageHandler


def get_msg_handling_mapping(extend_mapping_path: str = "") -> Dict:
    """
        功能描述：获取云边协同topic与处理函数的映射关系
        extend_mapping_path: 扩展的映射关系的路径
    """

    res = {
        Topic.SUB_CONFIG_NET_MANAGER: FDMessageHandler.handle_config_net_manager_task,
        Topic.SUB_RESET_ALARM: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_PASS_THROUGH_ACCOUNT_MODIFY: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_CONFIG_HOST_NAME: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_COMPUTER_SYSTEM_RESET: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_OPERATE_PROFILE: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_OPERATE_PROFILE_EFFECT: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_OPERATE_INFO_COLLECT: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_OPERATE_FIRMWARE_EFFECTIVE: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_OPERATE_INSTALL: FDMessageHandler.handle_upgrade_service,
        Topic.SUB_CERT_QUERY: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_CERT_UPDATE: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_CRL_UPDATE: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_CERT_DELETE: FDMessageHandler.handle_msg_from_fd_by_mqtt,
        Topic.SUB_HARDWARE_OPERATE_TAG: FDMessageHandler.handle_msg_from_fd_by_mqtt,
    }

    if not extend_mapping_path:
        return res

    try:
        index = extend_mapping_path.rfind(".")
        mapping_name = extend_mapping_path[index + 1:]
        module_path = extend_mapping_path[: index]
        module = importlib.import_module(module_path)
        if module and hasattr(module, mapping_name):
            mapping = getattr(module, mapping_name)
            res.update(mapping)
    except Exception as err:
        run_log.warning("get extend topic with handler mapping failed. reason: %s", err)

    return res


# 云边协同topic与处理函数的映射关系
MSG_HANDLING_MAPPING = get_msg_handling_mapping(EXTEND_FD_TOPIC_HANDLER_MAPPING_PATH)
