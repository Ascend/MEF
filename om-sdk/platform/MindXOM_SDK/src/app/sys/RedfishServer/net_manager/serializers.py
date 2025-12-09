# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json

from flask import make_response

from common.utils.app_common_method import AppCommonMethod
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from net_manager.common_methods import Serializer


class GetNodeIdSerializer(Serializer):
    service = Serializer.root.net_manage_service.get_node_id


class GetNetManageConfigSerializer(Serializer):
    service = Serializer.root.net_manage_service.get_net_manage_config


class GetFdCertSerializer(Serializer):
    service = Serializer.root.net_manage_service.get_fd_cert


class ImportFdCrlSerializer(Serializer):

    def make_200_response(self, data: dict, status: int):
        """
        功能描述：正常流程返回逻辑
        """
        resp_json = json.loads(SuccessMessageResourceSerializer().service.get_resource())
        AppCommonMethod.replace_kv_list(resp_json, data)
        return make_response(json.dumps(resp_json), status)


class ImportFdCertSerializer(Serializer):
    service = Serializer.root.net_manage_service.import_fd_cert
