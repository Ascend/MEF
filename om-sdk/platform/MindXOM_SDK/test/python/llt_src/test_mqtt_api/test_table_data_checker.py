# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from pytest_mock import MockerFixture

from ibma_redfish_globals import RedfishGlobals
from net_manager.checkers.table_data_checker import NetManagerCfgFdChecker, NetManagerCfgChecker
from net_manager.models import NetManager


class TestNetManagerCfgFdChecker:
    @staticmethod
    def test_net_manager_cfg_fd_checker_failed():
        ret = RedfishGlobals.check_external_parameter(NetManagerCfgFdChecker, {"server_name": "12"})
        assert ret == {"status": 400, "message": [100024, "Parameter is invalid."]}

    @staticmethod
    def test_net_manager_cfg_fd_checker_success():
        param_dict = {
            "server_name": "123",
            "server_ip": "1.1.1.1",
            "server_port": 443,
            "cloud_user": "cloudUser",
            "cloud_pwd": "cloudPwd",
            "status": "connecting",
        }
        assert not RedfishGlobals.check_external_parameter(NetManagerCfgFdChecker, param_dict)


class TestNetManagerCfgChecker:
    @staticmethod
    def test_check_web_cfg_failed():
        assert not NetManagerCfgChecker().check_web_cfg(NetManager(port="443")).success

    @staticmethod
    def test_check_web_cfg_success():
        assert NetManagerCfgChecker().check_web_cfg(NetManager()).success

    @staticmethod
    def test_check_fd_cfg_failed():
        assert not NetManagerCfgChecker().check_fd_cfg(NetManager()).success

    @staticmethod
    def test_check_fd_cfg_success():
        net_manager = NetManager(
            server_name="123",
            ip="1.1.1.1",
            port="443",
            cloud_user="cloudUser",
            cloud_pwd="cloudPwd",
            status="connecting",
        )
        assert NetManagerCfgChecker().check_fd_cfg(net_manager).success

    @staticmethod
    def test_check_net_cfg_with_invalid_mgmt_type():
        assert not NetManagerCfgChecker().check_net_cfg(NetManager()).success

    @staticmethod
    def test_check_net_cfg_with_invalid_node_id():
        net_manager = NetManager(net_mgmt_type="FusionDirector", node_id="/*-")
        assert not NetManagerCfgChecker().check_net_cfg(net_manager).success

    @staticmethod
    def test_check_net_cfg_success():
        net_manager = NetManager(net_mgmt_type="Web", node_id="e6a47e30-3a09-11ea-9218-a8494df5f123")
        assert NetManagerCfgChecker().check_net_cfg(net_manager).success

    @staticmethod
    def test_check_dict_with_empty_param():
        assert not NetManagerCfgChecker().check_dict(dict()).success
