# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

