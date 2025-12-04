# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from common.checkers import AndChecker
from common.checkers import BoolChecker
from common.checkers import BoolEqualChecker
from common.checkers import ExistsChecker
from common.checkers import IPV4Checker
from common.checkers import IntegerChecker
from common.checkers import ListChecker
from common.checkers import ModelChecker
from common.checkers import OrChecker
from common.checkers import RegexStringChecker
from common.checkers import StringBinaryChecker
from common.checkers import StringChoicesChecker
from common.checkers import StringEmptyChecker
from common.checkers import StringExcludeWordsChecker
from common.checkers import StringLengthChecker
from common.checkers import NotExistsChecker, SecurityLoadCfgChecker
from common.checkers import PasswordComplexityChecker
from common.checkers import DateChecker
from common.checkers.base_checker.list_checker import PartitionListChecker
from common.checkers.param_checker import LTEConfigInfoChecker

STRING_INCLUDE_BLACKLIST = ("..",)
CERT_TYPE_WHITELIST = ("FDRootCert",)
CERT_TYPE_CHOICES = ("text",)


class ComputerSystemResetChecker(ModelChecker):
    """复位主机系统-espmanager/ComputerSystemReset"""

    class Meta:
        from common.constants.product_constants import FD_RESTART_TYPE
        fields = (
            StringChoicesChecker("restart_method", choices=FD_RESTART_TYPE)
        )


class FirmwareEffectiveChecker(ModelChecker):
    """复位系统生效固件-espmanager/UpdateService/FirmwareEffective"""

    class Meta:
        fields = (
            StringChoicesChecker("active", choices=("inband",))
        )


class SysAssetTagChecker(ModelChecker):
    """修改自定义电子标签-espmanager/SysAssetTag"""

    class Meta:
        TAG_PATTERN = r'^([\x20-\x7E]){1,255}$'
        fields = (
            RegexStringChecker("asset_tag", match_str=TAG_PATTERN)
        )


class HostnameChecker(ModelChecker):
    """配置主机名-espmanager/Hostname"""

    class Meta:
        HOSTNAME_PATTERN = r"^(?!-)[A-Za-z0-9\-]{1,63}(?<!-)$"
        fields = (
            RegexStringChecker("hostname", match_str=HOSTNAME_PATTERN),
        )


class SysConfigEffectChecker(ModelChecker):
    """配置生效-espmanager/SysConfigEffect"""

    class Meta:
        PROFILE_NAME_PATTERN = r"^[a-z0-9A-Z-_.]+$"
        fields = (
            StringExcludeWordsChecker("profile_name", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("profile_name", match_str=PROFILE_NAME_PATTERN, min_len=1, max_len=32),
        )


class UserInfoChecker(ModelChecker):
    """用户信息修改-espmanager/passthrough/account_modify"""

    class Meta:
        USERNAME_PATTERN = r'^[a-zA-Z0-9]*[a-zA-Z]+[a-zA-Z0-9]*$'
        fields = (
            RegexStringChecker("account", match_str=USERNAME_PATTERN, min_len=1, max_len=16),
            PasswordComplexityChecker("new_password", min_len=8, max_len=20),
        )


class InfoCollectHttpsServerChecker(ModelChecker):
    """信息收集-https_server字段检查"""

    class Meta:
        URL_PATTERN = r"^POST https[^@\n!\\|;&$<>` ]+$"
        USERNAME_PATTERN = r'^[a-zA-Z][a-zA-Z0-9-_]{1,64}[a-zA-Z0-9]$'
        fields = (
            AndChecker(
                StringExcludeWordsChecker(
                    "url", words=STRING_INCLUDE_BLACKLIST),
                RegexStringChecker("url", match_str=URL_PATTERN, max_len=2048),
            ),
            RegexStringChecker("user_name", match_str=USERNAME_PATTERN),
            RegexStringChecker("password", '[0-9a-zA-Z!@#$%^&*()_+\-=~`{[\]}|\\:;\'"<,>.?/\s]{8,20}'),
        )


class InfoCollectChecker(ModelChecker):
    """信息收集-espmanager/InfoCollect"""

    class Meta:
        fields = (
            StringChoicesChecker("type", choices=("all",)),
            StringChoicesChecker("module", choices=("all",)),
            AndChecker(
                ExistsChecker("https_server"),
                InfoCollectHttpsServerChecker("https_server"),
            ),
        )


def string_compare(str1, str2):
    return str1 != str2


class SysConfigNtpServerChecker(ModelChecker):
    """配置导入-config-ntp_server字段检查"""

    class Meta:
        fields = (
            BoolChecker("sync_net_manager"),
            BoolChecker("service_enabled"),
            OrChecker(
                AndChecker(
                    BoolEqualChecker("service_enabled", equal_value=False),
                    BoolEqualChecker("sync_net_manager", equal_value=False),
                    StringEmptyChecker("preferred_server"),
                    StringEmptyChecker("alternate_server"),
                ),
                AndChecker(
                    BoolEqualChecker("service_enabled", equal_value=True),
                    OrChecker(
                        AndChecker(
                            BoolEqualChecker("sync_net_manager", equal_value=True),
                            StringEmptyChecker("preferred_server"),
                            StringEmptyChecker("alternate_server"),
                        ),
                        AndChecker(
                            BoolEqualChecker("sync_net_manager", equal_value=False),
                            IPV4Checker("preferred_server"),
                            OrChecker(
                                StringEmptyChecker("alternate_server"),
                                AndChecker(
                                    IPV4Checker("alternate_server"),
                                    StringBinaryChecker(
                                        "preferred_server", "alternate_server", compare_fun=string_compare),
                                ),
                            ),
                        ),
                    ),
                ),
            ),
        )


class SysConfigPartitionDeviceChecker(ModelChecker):
    """配置导入-config-partitions-device字段检查"""

    class Meta:
        DEVICE_LOCATION_PATTERN = r"^[0-9a-zA-Z-_.]+$"
        fields = (
            StringChoicesChecker("device_type", choices=(
                "SimpleStorage", "Volume")),
            ExistsChecker("device_location"),
            StringExcludeWordsChecker("device_location", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("device_location", match_str=DEVICE_LOCATION_PATTERN, max_len=256)
        )


class SysConfigPartitionChecker(ModelChecker):
    """配置导入-config-partitions字段检查"""

    class Meta:
        MOUNT_PATH_PATTERN = r"^\/[0-9a-zA-Z/_-]{1,256}$"
        fields = (
            AndChecker(
                IntegerChecker("capacity_bytes", restrict=True),
                SysConfigPartitionDeviceChecker("device"),
                OrChecker(
                    NotExistsChecker("file_system"),
                    StringChoicesChecker("file_system", choices=("ext4",)),
                ),
                OrChecker(
                    StringEmptyChecker("mount_path"),
                    AndChecker(
                        StringExcludeWordsChecker("mount_path", words=(" ",)),
                        RegexStringChecker(
                            "mount_path", match_str=MOUNT_PATH_PATTERN, max_len=256)
                    ),
                ),
            ),
        )


class SysConfigStaticHostChecker(ModelChecker):
    """配置导入-config-static_host_list字段检查"""

    class Meta:
        DOMAIN_PATTERN = r"^[A-Za-z0-9\-\.]{1,253}$"
        fields = (
            IPV4Checker("ip_address"),
            RegexStringChecker("name", match_str=DOMAIN_PATTERN)
        )


class SysConfigNameServerChecker(ModelChecker):
    """配置导入-config-name_server字段检查"""

    class Meta:
        fields = (
            IPV4Checker("ip_address")
        )


class SysConfigSecurityPolicyChecker(ModelChecker):
    """配置导入-config-password_validity、web_access、ssh_access字段检查"""

    class Meta:
        fields = (
            AndChecker(
                StringLengthChecker("password_validity", required=False),
                IntegerChecker("password_validity", min_value=0, max_value=365, required=False, restrict=False),
            ),
            BoolChecker("web_access", required=False),
            BoolChecker("ssh_access", required=False),
            AndChecker(
                StringLengthChecker("session_timeout", required=False),
                IntegerChecker("session_timeout", min_value=5, max_value=120, required=False, restrict=False),
            ),
            AndChecker(
                StringLengthChecker("cert_alarm_time", required=False),
                IntegerChecker("cert_alarm_time", min_value=7, max_value=180, required=False, restrict=False),
            ),
            ListChecker("security_load", elem_checker=SecurityLoadCfgChecker(), max_len=30, required=False),
        )


class SysConfigLteInfoChecker(ModelChecker):
    """配置导入-config-lte_info字段检查"""

    class Meta:
        fields = (
            OrChecker(
                AndChecker(
                    NotExistsChecker("state_lte"),
                    NotExistsChecker("state_data"),
                ),
                AndChecker(
                    BoolChecker("state_lte"),
                    BoolChecker("state_data"),
                    OrChecker(
                        BoolEqualChecker("state_lte", equal_value=True),
                        AndChecker(
                            BoolEqualChecker("state_lte", equal_value=False),
                            BoolEqualChecker("state_data", equal_value=False),
                        ),
                    ),
                ),
            ),
            ListChecker("apn_info", elem_checker=LTEConfigInfoChecker(), max_len=1, required=False),

        )


class SysConfigInfoChecker(ModelChecker):
    """配置导入-config字段检查"""

    class Meta:
        fields = (
            SysConfigNtpServerChecker("ntp_server", required=False),
            PartitionListChecker("partitions", elem_checker=SysConfigPartitionChecker(required=False), required=False),
            ListChecker("static_host_list", elem_checker=SysConfigStaticHostChecker(), max_len=128, required=False),
            ListChecker("name_server", elem_checker=SysConfigNameServerChecker(), max_len=3, required=False),
            ListChecker("lte_info", elem_checker=SysConfigLteInfoChecker(), min_len=1, max_len=1, required=False),
            SysConfigSecurityPolicyChecker("security_policy", required=False)
        )


class SysConfigChecker(ModelChecker):
    """配置导入-espmanager/SysConfig"""

    class Meta:
        PRODUCT_PATTERN = r"(^[a-z0-9A-Z]$)|(^[a-z0-9A-Z][a-z0-9A-Z-\s_.]{0,62}[a-z0-9A-Z]$)"
        PROFILE_NAME_PATTERN = r"(^[a-z0-9A-Z]$)|(^[a-z0-9A-Z][a-z0-9A-Z-_.]{0,30}[a-z0-9A-Z]$)"
        fields = (
            StringExcludeWordsChecker("product", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker(
                "product", match_str=PRODUCT_PATTERN, min_len=1, max_len=64),
            AndChecker(
                StringExcludeWordsChecker("profile_name", words=STRING_INCLUDE_BLACKLIST),
                RegexStringChecker("profile_name", match_str=PROFILE_NAME_PATTERN, min_len=1, max_len=32),
                SysConfigInfoChecker("config"),
            ),
        )


class HttpsServerChecker(ModelChecker):
    """固件升级-member字段检查-https_server"""

    class Meta:
        USER_NAME_PATTERN = r"^[a-zA-Z0-9-_]+$"
        HTTPS_URL_PATTERN = r"^\S+ https[^@\n!\\|;&$<>` ]+$"

        fields = (
            RegexStringChecker(
                "user_name", match_str=USER_NAME_PATTERN, min_len=1, max_len=64),
            StringLengthChecker("password", min_len=8, max_len=64),
            AndChecker(
                StringExcludeWordsChecker(
                    "image", words=STRING_INCLUDE_BLACKLIST),
                RegexStringChecker(
                    "image", match_str=HTTPS_URL_PATTERN, max_len=256),
            ),
        )


class FirmwareMemberChecker(ModelChecker):
    """固件升级-member字段检查"""

    class Meta:
        CHECK_CODE_PATTERN = r'^[0-9a-f]{64}$'
        NAME_PATTERN = r"^[a-zA-Z0-9-_.]+$"
        fields = (
            StringChoicesChecker("type", choices=("firmware",)),
            AndChecker(
                StringChoicesChecker("operator", choices=("install",)),
                HttpsServerChecker("https_server", required=True),
                StringChoicesChecker(
                    "install_method", choices=("force", "normal", "")),
                StringChoicesChecker(
                    "enable_method", choices=("now", "delay", "")),
                StringExcludeWordsChecker("name", words=("..",)),
                RegexStringChecker(
                    "name", match_str=NAME_PATTERN, min_len=1, max_len=256),
                StringChoicesChecker(
                    "check_type", choices=("sha256",), required=False),
            ),
            OrChecker(
                StringEmptyChecker("check_code", required=False),
                RegexStringChecker("check_code", match_str=CHECK_CODE_PATTERN),
            ),
        )


class UpdateServiceFirmwareChecker(ModelChecker):
    """固件升级-espmanager/UpdateService/Firmware"""

    class Meta:
        fields = (
            ListChecker("member_list",
                        elem_checker=FirmwareMemberChecker(), min_len=1, max_len=256)
        )


class SetDflcInfoChecker(ModelChecker):
    """写入dflc信息-espmanager/config_dflc"""

    class Meta:
        fields = (
            OrChecker(
                StringEmptyChecker("start_point", required=False),
                DateChecker("start_point"),
            ),
            IntegerChecker("life_span", min_value=0, max_value=255, restrict=True),
        )


class ReAlarmChecker(ModelChecker):
    """重上报告警"""

    class Meta:
        fields = (
            StringChoicesChecker("resource", choices=("alarm",))
        )


class FdCertImportChecker(ModelChecker):
    """FD导入根证书-espmanager/cert_update"""

    class Meta:
        fields = (
            RegexStringChecker("cert_name", match_str=r"^[0-9a-zA-Z_.]+$", min_len=4, max_len=64),
            StringExcludeWordsChecker("cert_name", words=("..",)),
            StringChoicesChecker("cert_type", choices=CERT_TYPE_WHITELIST),
            StringChoicesChecker("type", choices=CERT_TYPE_CHOICES),
            RegexStringChecker("content", match_str=r"^[a-zA-Z0-9=+/]+$", min_len=1, max_len=20480),
        )


class ImportCrlChecker(ModelChecker):
    """FD导入吊销列表-espmanager/crl_update"""

    class Meta:
        fields = (
            StringChoicesChecker("cert_type", choices=CERT_TYPE_WHITELIST),
            StringChoicesChecker("type", choices=CERT_TYPE_CHOICES),
            RegexStringChecker("content", match_str=r"^[a-zA-Z0-9=+/]+$", min_len=1, max_len=8192),
        )


class FdCertDeleteChecker(ModelChecker):
    """FD删除证书-espmanager/cert_delete"""

    class Meta:
        fields = (
            RegexStringChecker("cert_name", match_str=r"^[0-9a-zA-Z_.]+$", min_len=4, max_len=64),
            StringExcludeWordsChecker("cert_name", words=("..",)),
            StringChoicesChecker("cert_type", choices=CERT_TYPE_WHITELIST),
        )
