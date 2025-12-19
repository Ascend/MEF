# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
"""
功 能：外部参数校验
"""
from common.checkers import AlarmShieldMessageChecker
from common.checkers import DateChecker
from common.checkers import ExtensionChecker
from common.checkers import Ipv4AddressItemChecker
from common.checkers import LocalIpChecker
from common.checkers import LogNameChecker
from common.checkers import NotExistsChecker
from common.checkers import NumericChecker
from common.checkers import PartitionDeviceChecker
from common.checkers import PasswordComplexityChecker
from common.checkers import SecurityLoadCfgChecker
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

FILE_NAME_PATTERN = r"^[a-zA-Z0-9_.-]{1,255}$"
STRING_INCLUDE_BLACKLIST = ("..",)


class CreateSessionServiceChecker(ModelChecker):
    """创建新会话-/redfish/v1/SessionService/Sessions-POST"""

    class Meta:
        USERNAME_PATTERN = r"^[a-zA-Z0-9]*[a-zA-Z]+[a-zA-Z0-9]*$"
        fields = (
            RegexStringChecker("UserName", match_str=USERNAME_PATTERN, min_len=1, max_len=16),
            ExistsChecker("Password")
        )


class MemberIdChecker(ModelChecker):
    """校验MemberId-/redfish/v1/AccountService/Accounts/member_id"""

    class Meta:
        ID_PATTERN = r"^\d*$"
        fields = (
            RegexStringChecker("member_id", match_str=ID_PATTERN, min_len=1, max_len=16),
        )


def pword_compare(str1, str2):
    if str1 == str2:
        return True
    else:
        return False


class ChangeNameAndPasswordChecker(ModelChecker):
    """校验修改用户名和密码-/redfish/v1/AccountService/Accounts/member_id-PATCH"""

    class Meta:
        USERNAME_PATTERN = r'^[a-zA-Z0-9]*[a-zA-Z]+[a-zA-Z0-9]*$'
        fields = (
            RegexStringChecker("UserName", match_str=USERNAME_PATTERN, min_len=1, max_len=16),
            ExistsChecker("old_password"),
            PasswordComplexityChecker("Password", min_len=8, max_len=20),
            StringBinaryChecker("Password", "new_password_second", compare_fun=pword_compare),
            StringLengthChecker("new_password_second", min_len=8, max_len=20),
        )


class DeleteServiceChecker(ModelChecker):
    """删除指定会话-/redfish/v1/SessionService/Sessions/index-DELETE"""

    class Meta:
        PARAM_PATTERN = r"^[a-f0-9]+$"
        fields = (
            RegexStringChecker("index", match_str=PARAM_PATTERN, min_len=48, max_len=48),
        )


class AccountServiceInfoChecker(ModelChecker):
    """修改用户服务信息-/redfish/v1/AccountService-PATCH"""

    class Meta:
        fields = (
            IntegerChecker("PasswordExpirationDays", min_value=0, max_value=365, restrict=True),
            ExistsChecker("Password")
        )


class AlarmShieldChecker(ModelChecker):
    """
    新增告警屏蔽规则-/redfish/v1/Systems/Alarm/AlarmShield/Increase-PATCH
    减少告警屏蔽规则-/redfish/v1/Systems/Alarm/AlarmShield/Decrease-PATCH
    """

    class Meta:
        fields = (
            ListChecker("AlarmShieldMessages", elem_checker=AlarmShieldMessageChecker(), min_len=1, max_len=256),
        )


class SessionServiceInfoChecker(ModelChecker):
    """修改会话服务信息-/redfish/v1/SessionService-PATCH"""

    class Meta:
        fields = (
            IntegerChecker("SessionTimeout", min_value=5, max_value=120, restrict=True),
            ExistsChecker("Password")
        )


class UpdateServiceChecker(ModelChecker):
    """升级固件-/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate-POST"""

    class Meta:
        fields = (
            AndChecker(
                StringExcludeWordsChecker("ImageURI", words=STRING_INCLUDE_BLACKLIST),
                RegexStringChecker("ImageURI", match_str=FILE_NAME_PATTERN)
            ),
            StringChoicesChecker(
                "TransferProtocol", choices=("https",)),
        )


class FirmwareInventoryChecker(ModelChecker):
    """文件上传-/redfish/v1/UpdateService/FirmwareInventory-POST"""

    class Meta:
        fields = (
            StringExcludeWordsChecker("imgfile", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("imgfile", match_str=FILE_NAME_PATTERN),
            ExtensionChecker("imgfile", choices=(
                "conf", "ini", "crt", "crl", "zip", "cer")),
        )


class SystemChecker(ModelChecker):
    """修改指定系统资源属性-/redfish/v1/System-PATCH"""

    class Meta:
        ASSERT_TAG_PATTERN = r"^([\x20-\x7E]){1,255}$"
        DATE_TIME_OFFSET_PATTERN = r"^[\+\-_a-z:A-Z0-9\/]{0,100}$"
        HOSTNAME_PATTERN = r"^(?!-)[A-Za-z0-9\-]{1,63}(?<!-)$"
        fields = (
            RegexStringChecker("AssetTag", match_str=ASSERT_TAG_PATTERN, required=False),
            DateChecker("DateTime", min_len=10, required=False),
            AndChecker(
                StringExcludeWordsChecker(
                    "DateTimeLocalOffset", words=STRING_INCLUDE_BLACKLIST, required=False),
                RegexStringChecker(
                    "DateTimeLocalOffset", match_str=DATE_TIME_OFFSET_PATTERN, required=False),
            ),
            RegexStringChecker(
                "HostName", match_str=HOSTNAME_PATTERN, required=False),
        )


class ProductNameChecker(ModelChecker):
    """读取前端配置文件中的model--即上报给FD的产品名校验"""

    class Meta:
        PATTERN = r"^[a-zA-Z0-9\s]+$"
        fields = (
            RegexStringChecker("model", match_str=PATTERN, min_len=1, max_len=64),
        )


class ModuleChecker(ModelChecker):
    """查询模组资源属性-/redfish/v1/System/Modules/<module_id>-GET"""

    class Meta:
        NAME_PATTERN = r"^[a-zA-Z0-9\-\_\s]+$"
        fields = (
            RegexStringChecker("module_id", match_str=NAME_PATTERN, min_len=1, max_len=127),
        )


class DeviceChecker(ModelChecker):
    """查询/修改指定设备资源属性-/redfish/v1/System/Modules/<module_id>/<device_id>-GET/PATCH"""

    class Meta:
        NAME_PATTERN = r"^[a-zA-Z0-9\-\_\s]+$"
        fields = (
            RegexStringChecker("module_id", match_str=NAME_PATTERN, min_len=1, max_len=127),
            RegexStringChecker("device_id", match_str=NAME_PATTERN, min_len=1, max_len=127),
        )


class ComputerSystemResetChecker(ModelChecker):
    """系统复位操作-/redfish/v1/Systems/Actions/ComputerSystem.Reset-POST"""

    class Meta:
        from common.constants.product_constants import REDFISH_RESTART_TYPE
        fields = (
            StringChoicesChecker("ResetType", choices=REDFISH_RESTART_TYPE),
        )


def ip_string_compare(str1, str2):
    return str1 != str2


class NTPServiceChecker(ModelChecker):
    """添加NTP配置信息-/redfish/v1/Systems/NTPService-PATCH"""

    class Meta:
        fields = (
            AndChecker(
                BoolChecker("ClientEnabled"),
                BoolChecker("ServerEnabled"),
                BoolEqualChecker("ServerEnabled", equal_value=False),
                OrChecker(
                    StringEmptyChecker("NTPLocalServers"),
                    IPV4Checker("NTPLocalServers", required=False),
                ),
                OrChecker(
                    BoolEqualChecker("ClientEnabled", equal_value=False),
                    AndChecker(
                        BoolEqualChecker("ClientEnabled", equal_value=True),
                        IPV4Checker("NTPRemoteServers"),
                        OrChecker(
                            StringEmptyChecker("NTPRemoteServersbak"),
                            AndChecker(
                                IPV4Checker("NTPRemoteServersbak", required=False),
                                StringBinaryChecker("NTPRemoteServers", "NTPRemoteServersbak",
                                                    compare_fun=ip_string_compare),
                            ),
                        ),
                    ),
                ),
                StringChoicesChecker("Target", choices=("Client",), required=False),
            )
        )


class ExtendedDeviceChecker(ModelChecker):
    """查询外部设备资源的信息-/redfish/v1/Systems/ExtendedDevices/extendeddevice_id-GET"""

    class Meta:
        EXTEND_ID_PATTERN = r"^[a-zA-Z0-9\-\_\s\.]{2,64}$"
        fields = (
            RegexStringChecker("extend_id", match_str=EXTEND_ID_PATTERN),
        )


class LTEStatusInfoChecker(ModelChecker):
    """配置LTE接口资源信息-/redfish/v1/Systems/LTE/StatusInfo-PATCH"""

    class Meta:
        fields = (
            BoolChecker("state_lte"),
            BoolChecker("state_data"),
            OrChecker(
                BoolEqualChecker("state_lte", equal_value=True),
                AndChecker(
                    BoolEqualChecker("state_lte", equal_value=False),
                    BoolEqualChecker("state_data", equal_value=False),
                )
            )
        )


class LTEConfigInfoChecker(ModelChecker):
    """配置LTE APN接口资源信息-/redfish/v1/Systems/LTE/ConfigInfo-PATCH"""

    class Meta:
        CONFIG_PATTERN = r"^[a-zA-Z0-9~`!\?, .:;\-_\'\"\(\)\{\}\[\]\/<>@#\$%\^&\*\+\|\\=\s]{1,64}$"
        APN_NAME_PATTERN = r"^[a-zA-Z0-9\-_\.@]{1,39}$"
        APN_USER_PATTERN = r"^[a-zA-Z0-9\-_\.@]{1,64}$"
        fields = (
            RegexStringChecker(
                "apn_name", match_str=APN_NAME_PATTERN, min_len=1, max_len=39),
            OrChecker(
                AndChecker(
                    RegexStringChecker("apn_user", match_str=APN_USER_PATTERN,
                                       min_len=1, max_len=64, required=False),
                    RegexStringChecker("apn_passwd", match_str=CONFIG_PATTERN,
                                       min_len=1, max_len=64, required=False),
                ),
                AndChecker(
                    NotExistsChecker("apn_user"),
                    NotExistsChecker("apn_passwd"),
                ),
                AndChecker(
                    StringEmptyChecker("apn_user"),
                    StringEmptyChecker("apn_passwd"),
                ),
            ),
            StringChoicesChecker("auth_type", choices=("0", "1", "2", "3")),
        )


class EthernetInterfaceIdChecker(ModelChecker):
    """根据网口查询指定主机以太网资源信息
    /redfish/v1/Systems/EthernetInterfaces/<ethId>-GET
    """

    class Meta:
        fields = (
            StringChoicesChecker("eth_id", choices=("GMAC0", "GMAC1")),
        )


class EthernetInterfaceChecker(ModelChecker):
    """配置以太网接口-/redfish/v1/Systems/EthernetInterfaces/<ethId>-PATCH"""

    class Meta:
        fields = (
            ListChecker(
                "IPv4Addresses", elem_checker=Ipv4AddressItemChecker(), min_len=1, max_len=4),
        )


class StoragesInfoChecker(ModelChecker):
    """查询指定存储资源信息-/redfish/v1/EdgeSystem/SimpleStorages/<storage_id>-GET"""

    class Meta:
        fields = (
            StringChoicesChecker("storage_id", choices=("1", "2", "3",)),
        )


class CreatePatitionChecker(ModelChecker):
    """创建磁盘分区-/redfish/v1/Systems/Partitions-POST"""

    class Meta:
        CAPACITY_BYTES_PATTERN = r"^((\d*(\.)?\d{1})|(\d+)(\.\d{0,1})?)$"
        fields = (
            IntegerChecker("Number", min_value=1, max_value=16,
                           required=True, restrict=True),
            NumericChecker("CapacityBytes", min_value=0.5, required=True),
            RegexStringChecker("CapacityBytes", match_str=CAPACITY_BYTES_PATTERN),
            ListChecker("Links", elem_checker=PartitionDeviceChecker(),
                        min_len=1, max_len=1),
            StringChoicesChecker("FileSystem", choices=("ext4",)),
        )


class PartitionIdChecker(ModelChecker):
    """获取指定磁盘分区资源信息-/redfish/v1/Systems/Partitions/<patition_id>-GET"""

    class Meta:
        PARAM_PATTERN = r"^[a-z0-9A-Z_]+$"
        fields = (
            StringExcludeWordsChecker("partition_id", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("partition_id", match_str=PARAM_PATTERN, min_len=1, max_len=128),
        )


class MountPartitionChecker(ModelChecker):
    """挂载磁盘分区-/redfish/v1/Systems/Partitions/Mount-PATCH"""

    class Meta:
        PARAM_PATTERN = r"^[a-z0-9A-Z_]+$"
        PARTITION_PATTERN = r"^\/[0-9a-zA-Z/_-]{1,255}$"
        fields = (
            StringExcludeWordsChecker("PartitionID", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("PartitionID", match_str=PARAM_PATTERN, min_len=1, max_len=128),
            StringExcludeWordsChecker("MountPath", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("MountPath", match_str=PARTITION_PATTERN),
        )


class UnmountPartitionChecker(ModelChecker):
    """解挂磁盘分区-/redfish/v1/Systems/Partitions/Unmount-PATCH"""

    class Meta:
        PARAM_PATTERN = r"^[a-z0-9A-Z_]+$"
        fields = (
            StringExcludeWordsChecker("PartitionID", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("PartitionID", match_str=PARAM_PATTERN, min_len=1, max_len=128),
        )


class NfsManagerChecker(ModelChecker):
    """
    挂载NFS分区-/redfish/v1/Systems/NfsManage/Actions/NfsManage.Mount-POST
    解挂NFS分区-/redfish/v1/Systems/NfsManage/Actions/NfsManage.Unmount-POST
    """

    class Meta:
        PATH_PATTERN = r"^\/[0-9a-zA-Z/_-]{1,255}$"
        fields = (
            IPV4Checker("ServerIP"),
            AndChecker(
                StringExcludeWordsChecker("ServerDir", words=STRING_INCLUDE_BLACKLIST),
                RegexStringChecker("ServerDir", match_str=PATH_PATTERN),
            ),
            StringChoicesChecker("FileSystem", choices=("nfs4",)),
            AndChecker(
                StringExcludeWordsChecker("MountPath", words=STRING_INCLUDE_BLACKLIST),
                RegexStringChecker("MountPath", match_str=PATH_PATTERN),
            )
        )


class ImportServerCertificateChecker(ModelChecker):
    """导入服务器证书
    /redfish/v1/Systems/SecurityService/HttpsCert/Actions/HttpsCert.ImportServerCertificate-POST
    """

    class Meta:
        fields = (
            StringExcludeWordsChecker("FileName", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("FileName", match_str=FILE_NAME_PATTERN),
            ExtensionChecker("FileName", choices=("crt", "cer")),
            ExistsChecker("Password"),
        )


class LogServiceDownloadChecker(ModelChecker):
    """下载日志信息-/redfish/v1/Systems/LogServices/Actions/download-POST"""

    class Meta:
        fields = (
            LogNameChecker("name"),
        )


class UpdateServiceResetChecker(ModelChecker):
    """将已升级的固件文件生效-/redfish/v1/EdgeSystem/Actions/UpdateService.Reset-POST"""

    class Meta:
        fields = (
            StringChoicesChecker("ResetType", choices=("GracefulRestart",)),
        )


class PunyDictImportChecker(ModelChecker):
    """导入弱字典-/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictImport-POST"""

    class Meta:
        fields = (
            ExistsChecker("FileName"),
            StringExcludeWordsChecker("FileName", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("FileName", match_str=FILE_NAME_PATTERN),
            ExtensionChecker("FileName", choices=("conf",)),
            ExistsChecker("Password")
        )


class PunyDictDeleteChecker(ModelChecker):
    """删除弱字典-/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictDelete-POST"""

    class Meta:
        fields = (
            ExistsChecker("Password")
        )


class RestoreDefaultConfigChecker(ModelChecker):
    """配置登录规则信息-/redfish/v1/Systems/SecurityService/SecurityLoad-PATCH"""

    class Meta:
        fields = (
            BoolChecker("ReserveIP"),
            ExistsChecker("Password")
        )


class SecurityLoadChecker(ModelChecker):
    """配置登录规则信息-/redfish/v1/Systems/SecurityService/SecurityLoad-PATCH"""

    class Meta:
        fields = (
            ListChecker("load_cfg", elem_checker=SecurityLoadCfgChecker(), max_len=30),
            ExistsChecker("Password")
        )


class SecurityLoadImportChecker(ModelChecker):
    """导入登录规则信息-/redfish/v1/Systems/SecurityService/SecurityLoad.Import-POST"""

    class Meta:
        fields = (
            ExistsChecker("Password"),
            StringExcludeWordsChecker("file_name", words=STRING_INCLUDE_BLACKLIST),
            RegexStringChecker("file_name", match_str=FILE_NAME_PATTERN),
            ExtensionChecker("file_name", choices=("ini",))
        )


class CertAlarmTimeInfoChecker(ModelChecker):
    """修改证书报警时间-/redfish/v1/Systems/SecurityService/HttpsCertAlarmTime-PATCH"""

    class Meta:
        fields = (
            IntegerChecker("CertAlarmTime", min_value=7, max_value=180, restrict=True),
            ExistsChecker("Password")
        )


class ResetOpChecker(ModelChecker):
    """远程恢复出厂设置操作-/redfish/v1/Systems/RestoreDefaults.Reset-POST"""

    class Meta:
        ETH_NAME_PATTERN = r"^[a-z0-9][a-z0-9:]{0,31}$"
        fields = (
            OrChecker(
                StringEmptyChecker("ethernet"),
                RegexStringChecker("ethernet", match_str=ETH_NAME_PATTERN)
            ),
            ExistsChecker("root_pwd")
        )


class FdIpChecker(AndChecker):
    """网管FD的服务IP校验规则"""

    def __init__(self, attr_name: str = None, required: bool = True):
        super().__init__(IPV4Checker(attr_name, required), LocalIpChecker(attr_name, required))
