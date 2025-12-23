# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from common.checkers import ExtensionChecker
from common.checkers import PasswordComplexityChecker
from common.checkers.param_checker import FdIpChecker
from common.checkers import BoolChecker
from common.checkers import CheckResult
from common.checkers import ExistsChecker
from common.checkers import ModelChecker
from common.checkers import PortChecker
from common.checkers import RegexStringChecker
from common.checkers import StringBinaryChecker
from common.checkers import StringChoicesChecker
from common.checkers import StringExcludeWordsChecker

FILE_NAME_PATTERN = r"^[a-zA-Z0-9_.-]{1,255}$"
STRING_INCLUDE_BLACKLIST = ("..",)


class CertUploadChecker(ModelChecker):
    """网管证书文件上传"""

    class Meta:
        fields = (
            StringExcludeWordsChecker("imgfile", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("imgfile", match_str=FILE_NAME_PATTERN),
            ExtensionChecker("imgfile", choices=("crt",)),
        )


class CrlUploadChecker(ModelChecker):
    """网管吊销列表文件上传"""

    class Meta:
        fields = (
            StringExcludeWordsChecker("imgfile", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("imgfile", match_str=FILE_NAME_PATTERN),
            ExtensionChecker("imgfile", choices=("crl",)),
        )


def judge_pwd_by_account(password: str, account: str):
    if password in (account, account[:: -1]):
        return False
    return True


class FdConfigChecker(ModelChecker):
    class Meta:
        NODE_ID_MATCH_STR = "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"
        fields = (
            FdIpChecker("NetIP"),
            RegexStringChecker("NetAccount", "^[a-zA-Z0-9-_]{1,256}$"),
            # fd对纳管密码的要求是复杂度3，长度8-32
            PasswordComplexityChecker("NetPassword", min_len=8, max_len=32),
            StringBinaryChecker("NetPassword", "NetAccount", compare_fun=judge_pwd_by_account),
            BoolChecker("test"),
            RegexStringChecker("ServerName", "^[A-Za-z0-9-.]{0,64}$"),
            PortChecker("Port", restrict=True),
            RegexStringChecker("NodeId", NODE_ID_MATCH_STR),
        )


class NetManageConfigChecker(ExistsChecker):
    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(attr_name, required)

    @staticmethod
    def check_config_params(net_manager_params: dict) -> CheckResult:
        net_mgmt_type = net_manager_params.get("ManagerType")
        result = StringChoicesChecker("ManagerType", ("Web", "FusionDirector",)).check({"ManagerType": net_mgmt_type})
        if not result.success:
            return CheckResult.make_failed(f"Net manager config checkers: {result.reason}")

        if net_mgmt_type == "Web":
            return CheckResult.make_success()

        return FdConfigChecker().check(net_manager_params)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        net_manager = self.raw_value(data)
        if not net_manager:
            return CheckResult.make_failed("Net manager config checkers: net_manager is null.")
        return self.check_config_params(net_manager)
