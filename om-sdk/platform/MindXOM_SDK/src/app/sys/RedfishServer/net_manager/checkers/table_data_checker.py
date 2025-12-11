# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from common.checkers import DateChecker
from common.checkers.param_checker import FdIpChecker
from common.checkers import CheckResult
from common.checkers import ExistsChecker
from common.checkers import ModelChecker
from common.checkers import PortChecker
from common.checkers import RegexStringChecker
from common.checkers import StringChoicesChecker
from net_manager.checkers.contents_checker import CertContentsChecker
from net_manager.checkers.contents_checker import CrlContentsChecker
from net_manager.constants import NetManagerConstants
from net_manager.models import CertManager
from net_manager.models import NetManager

NAME_MIN_LEN = 4
NAME_MAX_LEN = 64


class CertManagerDataChecker(ModelChecker):

    class Meta:
        fields = (
            RegexStringChecker("name", "^[A-Za-z0-9-_.]+$", min_len=NAME_MIN_LEN, max_len=NAME_MAX_LEN),
            StringChoicesChecker("source", ("Web", "FusionDirector",)),
            DateChecker("update_time"),
            CertContentsChecker("cert_contents", min_len=0),
            CrlContentsChecker("crl_contents", min_len=0),
        )


class CertManagerChecker(ExistsChecker):
    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(attr_name, required)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        cert_manager: CertManager = self.raw_value(data)
        if not cert_manager:
            msg_format = f"Cert manager checkers: cert_manager is null."
            return CheckResult.make_failed(msg_format)

        check_ret = CertManagerDataChecker().check(cert_manager.to_dict())
        if not check_ret.success:
            msg_format = f"Cert manager checkers: {check_ret.reason}."
            return CheckResult.make_failed(msg_format, err_code=check_ret.err_code)

        return CheckResult.make_success()


class NetManagerCfgFdChecker(ModelChecker):

    class Meta:
        fields = (
            RegexStringChecker("server_name", "^[A-Za-z0-9-.]{0,64}$"),
            FdIpChecker("server_ip"),
            PortChecker("server_port"),
            RegexStringChecker("cloud_user", "^[a-zA-Z0-9-_]{1,256}$"),
            RegexStringChecker("cloud_pwd", "^[0-9a-zA-Z/+=]{1,256}$"),
            StringChoicesChecker("status", ("", "connecting", "connected", "ready")),
        )


class NetManagerCfgChecker(ExistsChecker):
    NODE_ID_MATCH_STR = "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"

    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(attr_name, required)

    @staticmethod
    def check_web_cfg(net_manager: NetManager) -> CheckResult:
        check_map = {
            "server_name": net_manager.server_name,
            "server_ip": net_manager.ip,
            "server_port": net_manager.port,
            "cloud_user": net_manager.cloud_user,
            "cloud_pwd": net_manager.cloud_pwd,
            "status": net_manager.status,
        }
        for key, value in check_map.items():
            if value:
                msg_format = f"Net manager config checkers: check {key}'s value is invalid."
                return CheckResult.make_failed(msg_format)

        return CheckResult.make_success()

    @staticmethod
    def check_fd_cfg(net_manager: NetManager) -> CheckResult:
        check_data = {
            "server_name": net_manager.server_name,
            "server_ip": net_manager.ip,
            "server_port": net_manager.port,
            "cloud_user": net_manager.cloud_user,
            "cloud_pwd": net_manager.cloud_pwd,
            "status": net_manager.status,
        }
        check_ret = NetManagerCfgFdChecker().check(check_data)
        if not check_ret.success:
            msg_format = f"Net manager config checkers: {check_ret.reason}."
            return CheckResult.make_failed(msg_format)

        return CheckResult.make_success()

    def check_net_cfg(self, net_manager: NetManager) -> CheckResult:
        net_mgmt_type = net_manager.net_mgmt_type
        result = StringChoicesChecker("net_type", ("Web", "FusionDirector",)).check({"net_type": net_mgmt_type})
        if not result.success:
            msg_format = f"Net manager config checkers: {result.reason}"
            return CheckResult.make_failed(msg_format)

        result = RegexStringChecker("node_id", self.NODE_ID_MATCH_STR).check({"node_id": net_manager.node_id})
        if not result.success:
            msg_format = f"Net manager config checkers: {result.reason}"
            return CheckResult.make_failed(msg_format)

        check_map = {
            NetManagerConstants.WEB: self.check_web_cfg(net_manager),
            NetManagerConstants.FUSION_DIRECTOR: self.check_fd_cfg(net_manager),
        }
        return check_map.get(net_mgmt_type)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        net_manager = self.raw_value(data)
        if not net_manager:
            msg_format = f"Net manager config checkers: net_manager is null."
            return CheckResult.make_failed(msg_format)

        return self.check_net_cfg(net_manager)
