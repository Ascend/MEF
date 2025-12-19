# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os.path

from common.ResourceDefV1.resource import RfResource


class RfNetManage(RfResource):
    GET_NODE_ID_DIR = os.path.normpath("redfish/v1/NetManager/NodeID")
    GET_NET_MANAGE_CONFIG = os.path.normpath("redfish/v1/NetManager")
    GET_FD_CERT = os.path.normpath("redfish/v1/NetManager/QueryFdCert")
    IMPORT_FD_CERT = os.path.normpath("redfish/v1/NetManager/ImportFdCert")

    get_node_id: RfResource
    get_net_manage_config: RfResource
    get_fd_cert: RfResource
    import_fd_cert: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.get_node_id = RfResource(base_path, self.GET_NODE_ID_DIR)
        self.get_net_manage_config = RfResource(base_path, self.GET_NET_MANAGE_CONFIG)
        self.get_fd_cert = RfResource(base_path, self.GET_FD_CERT)
        self.import_fd_cert = RfResource(base_path, self.IMPORT_FD_CERT)

