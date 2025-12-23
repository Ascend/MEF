# -*-coding:utf-8-*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import datetime
import fcntl
import re
import socket
import struct
import sys
from abc import ABC
from typing import Iterable
from typing import Set
from typing import Tuple

from common.checkers.base_checker.basic_attr_checker import ExistsChecker, StringLengthChecker, RegexStringChecker
from common.checkers.base_checker.basic_attr_checker import FloatChecker, StringChoicesChecker, StringEmptyChecker
from common.checkers.base_checker.basic_attr_checker import IntegerChecker, BoolChecker, BoolEqualChecker
from common.checkers.base_checker.ip_checker import IPV4Checker
from common.checkers.base_checker.list_checker import OrChecker, AndChecker
from common.checkers.base_checker.model_checker import ModelChecker
from common.checkers.base_checker.string_checker import StringExcludeWordsChecker
from common.net_check import NetCheck
from common.checkers.base_checker.abc_checker import AttrCheckerBase
from common.checkers.base_checker.abc_checker import CheckResult


class SubLayerRedundancyParaChecker(AttrCheckerBase):
    def __init__(self, attr_name: str, supported_fields: Tuple):
        """

        :param attr_name: 如果要解析当前字典，传入None或者""即可
        :param supported_fields: 支持的参数列表字段
        """
        super().__init__(attr_name)
        self._supported_fields = supported_fields

    def required(self) -> bool:
        return True

    def check_dict(self, data: dict) -> CheckResult:
        if self.name():
            value = self.raw_value(data)
            if value is None:
                return CheckResult.make_failed("RedundancyParaChecker: {} not exists".format(self.name()))
            data = self.raw_value(data)

        ret = [k for k, _ in data.items() if k not in self._supported_fields]
        if ret:
            return CheckResult.make_failed(f"RedundancyParaChecker: {self.name()} has superfluous para.")
        return CheckResult.make_success()


class PartitionDeviceOdataIdChecker(ModelChecker):
    """磁盘分区-路径参数"""

    class Meta:
        STRING_INCLUDE_BLACKLIST = ("..",)
        DEVICE_ID_PATTERN = r"^/dev/[A-Za-z0-9_-]{1,251}$"
        fields = (
            ExistsChecker("@odata.id", required=True),
            StringExcludeWordsChecker("@odata.id", words=STRING_INCLUDE_BLACKLIST),
            StringLengthChecker("@odata.id", max_len=256),
            RegexStringChecker("@odata.id", match_str=DEVICE_ID_PATTERN)
        )


class PartitionDeviceChecker(ModelChecker):
    """磁盘分区"""

    class Meta:
        fields = (
            SubLayerRedundancyParaChecker(
                "Device", supported_fields=("@odata.id",)),
            PartitionDeviceOdataIdChecker("Device"),
        )


class NotExistsChecker(ExistsChecker):
    def __init__(self, attr_name=None, required: bool = True):
        super().__init__(attr_name, required)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if result.success:
            return CheckResult.make_failed("NotExistsChecker: {} exists".format(self.name()))
        else:
            return CheckResult.make_success()


class NumericChecker(FloatChecker):
    def __init__(
            self,
            attr_name: str = None,
            min_value: float = sys.float_info.min,
            max_value: float = sys.float_info.max,
            required: bool = True,
    ):
        super().__init__(attr_name, min_value, max_value, required)

    def check_dict(self, data: dict) -> CheckResult:
        ret = super().check_dict(data)
        if not ret.success:
            return ret

        value = self.raw_value(data)
        if not self.is_multiple_of_half(float(value)):
            return CheckResult.make_failed("must be an integer multiple of 0.5")

        return CheckResult.make_success()

    def is_multiple_of_half(self, num):
        return num % 0.5 == 0


class PasswordComplexityChecker(StringLengthChecker):
    """检测密码复杂度"""

    def __init__(self, attr_name=None, min_len: int = 0, max_len: int = sys.maxsize, required: bool = True):
        super().__init__(attr_name, min_len, max_len, required)

    @staticmethod
    def check_password_ok(password: str):
        """检验密码复杂度是否符合要求"""
        pword_complexity = 0
        if not isinstance(password, (str, bytes)):
            return False
        # 密码复杂度，密码要求包含至少三种字符
        pattern_pword_num = re.compile(r"[0-9]")
        match_pword_num = pattern_pword_num.findall(password)
        if match_pword_num:
            pword_complexity = pword_complexity + 1
        pattern_pword_lowercase = re.compile(r'[a-z]')
        match_pword_lowercase = pattern_pword_lowercase.findall(password)
        if match_pword_lowercase:
            pword_complexity = pword_complexity + 1
        pattern_pword_uppercase = re.compile(r'[A-Z]')
        match_pword_uppercase = pattern_pword_uppercase.findall(password)
        if match_pword_uppercase:
            pword_complexity = pword_complexity + 1
        pattern_pword_symbol = re.compile(r'[!@#$%^&*()_+\-=~`{[\]}|\\:;\'"<,>.?/ ]')
        match_pword_symbol = pattern_pword_symbol.findall(password)
        if match_pword_symbol:
            pword_complexity = pword_complexity + 1
        # 如果密码长度小于6，或密码复杂度小于3则密码格式错误
        password_complexity = 3

        if pword_complexity < password_complexity:
            return False
        return True

    def check_dict(self, data) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if not self.check_password_ok(value):
            msg_format = "Insufficient password complexity!"
            return CheckResult.make_failed(msg_format)
        return super().check_dict(data)


class ExtensionChecker(StringChoicesChecker):
    """检查文件后缀"""

    def __init__(self, attr_name: str, choices: Tuple[str, ...] = (), required: bool = True):
        super().__init__(attr_name, choices, required)
        self.attr_name = attr_name

    def check_dict(self, data: dict) -> CheckResult:
        value = self.raw_value(data)
        if "." not in value:
            msg_format = "ExtensionChecker: . is not in param."
            return CheckResult.make_failed(msg_format)
        split_data = {self.attr_name: value.rsplit(".", 1)[1]}
        result = super().check_dict(split_data)
        if result.success:
            return CheckResult.make_success()
        else:
            return CheckResult.make_failed(
                "ExtensionChecker: Invalid extension of {}".format(self.attr_name))


class GatewayChecker(StringLengthChecker):
    """网关检查"""
    def check_dict(self, data: dict) -> CheckResult:
        value = self.raw_value(data)
        address = value["Address"]
        gateway = value["Gateway"]
        mask_num = NetCheck.get_num_for_mask(value["SubnetMask"])
        if not NetCheck.is_same_network_segment(address, gateway, mask_num):
            msg_format = "The gateway does not match the ip and subnet mask"
            return CheckResult.make_failed(msg_format)
        if gateway == NetCheck.net_work_address(address, value["SubnetMask"]):
            msg_format = "The gateway is the same as network address"
            return CheckResult.make_failed(msg_format)
        return CheckResult.make_success()


class SubnetMaskChecker(RegexStringChecker):
    """子网检查"""

    def __init__(self, attr_name=None, match_str="", required=True):
        super().__init__(attr_name, match_str=match_str, required=required, max_len=15)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        if value is None:
            return CheckResult.make_success()
        if not NetCheck.get_num_for_mask(value):
            return CheckResult.make_failed("SubnetMask is wrong")
        return CheckResult.make_success()


class Ipv4AddressItemChecker(ModelChecker):
    """Ipv4地址信息检查"""

    class Meta:
        SUBNET_MASK_PATTERN = r"^(254|252|248|240|224|192|128|0)\.0\.0\.0" \
                              r"|255\.(254|252|248|240|224|192|128|0)\.0\.0" \
                              r"|255\.255\.(254|252|248|240|224|192|128|0)\.0" \
                              r"|255\.255\.255\.(255|254|252|248|240|224|192|128|0)$"
        TAG_PATTERN = r"^[0-9a-zA-Z_]{1,32}$"
        fields = (
            IPV4Checker("Address"),
            SubnetMaskChecker("SubnetMask", match_str=SUBNET_MASK_PATTERN),
            StringChoicesChecker("AddressOrigin", choices=("Static", )),
            OrChecker(
                StringEmptyChecker("Gateway", required=False),
                AndChecker(
                    IPV4Checker("Gateway"),
                    GatewayChecker(),
                )
            ),
            OrChecker(
                StringEmptyChecker("Tag"),
                RegexStringChecker("Tag", match_str=TAG_PATTERN),
            ),
            IntegerChecker("VlanId", min_value=1, max_value=4094, required=False, restrict=True),
            OrChecker(
                AndChecker(
                    NotExistsChecker("ConnectTest"),
                    OrChecker(
                        NotExistsChecker("RemoteTestIp"),
                        StringLengthChecker("RemoteTestIp", min_len=0, max_len=128),
                    )
                ),
                AndChecker(
                    BoolChecker("ConnectTest"),
                    OrChecker(
                        AndChecker(
                            BoolEqualChecker("ConnectTest", equal_value=False),
                            OrChecker(
                                NotExistsChecker("RemoteTestIp"),
                                StringLengthChecker("RemoteTestIp", min_len=0, max_len=128),
                            )
                        ),
                        AndChecker(
                            BoolEqualChecker("ConnectTest", equal_value=True),
                            StringLengthChecker("RemoteTestIp", min_len=0, max_len=128),
                            IPV4Checker("RemoteTestIp"),
                        )
                    )
                ),
            )
        )


class LogNameChecker(StringEmptyChecker):
    """日志服务资源检测"""

    def __init__(self, attr_name=None, required: bool = True):
        super().__init__(attr_name, required)
        self.attr_name = attr_name

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if result.success:
            return CheckResult.make_failed("Log names no values")
        else:
            from common.constants.product_constants import LOG_COLLECT_LIST
            log_names = data.get(self.attr_name).strip(" ").split(" ")
            log_names_set = set(log_names)
            if len(log_names) != len(log_names_set):
                return CheckResult.make_failed("Log names have duplicate values")
            if not log_names_set.issubset(set(LOG_COLLECT_LIST)):
                return CheckResult.make_failed("Log names wrong")
            else:
                return CheckResult.make_success()


class TimeHourMinChecker(StringLengthChecker):
    def __init__(self, attr_name: str, min_len: int = 0, max_len: int = sys.maxsize, required: bool = True):
        super().__init__(attr_name, min_len, max_len, required)

    @staticmethod
    def check_time_hour_min(hh_mm):
        if ':' not in hh_mm:
            return False
        hh_mm = hh_mm.split(':')
        if len(hh_mm) != 2:
            return False
        if any(len(item) > 128 for item in hh_mm):
            return False
        try:
            hh, mm = int(hh_mm[0]), int(hh_mm[1])
        except Exception:
            return False
        if hh < 0 or hh > 24:
            return False
        if mm < 0 or mm > 59:
            return False
        if hh == 24 and mm != 0:
            return False
        return True

    def check_dict(self, data: dict) -> CheckResult:
        value = self.raw_value(data)
        if not TimeHourMinChecker.check_time_hour_min(value):
            err_msg = "Invalid time format"
            return CheckResult.make_failed(err_msg)
        return CheckResult.make_success()


class Ipv4WithMaskChecker(StringLengthChecker):
    def __init__(self, attr_name: str, min_len: int = 0, max_len: int = sys.maxsize, required: bool = True):
        super().__init__(attr_name, min_len, max_len, required)

    def check_dict(self, data: dict) -> CheckResult:
        ip_addr = self.raw_value(data)
        if "/" in ip_addr:
            ip_raw = ip_addr.split('/')[0]
            if not IPV4Checker("ip").check({"ip": ip_raw}):
                err_msg = "Parameter ip_addr with mask invalid"
                return CheckResult.make_failed(err_msg)
            mask_num = ip_addr.split('/')[1]
            if len(mask_num) > 128:
                return CheckResult.make_failed("Parameter mask invalid")
            try:
                mask_num = int(mask_num)
            except Exception:
                err_msg = "extract ip/mask info fail"
                return CheckResult.make_failed(err_msg)
            if mask_num <= 0 or mask_num > 32:
                err_msg = "Parameter mask invalid"
                return CheckResult.make_failed(err_msg)
        else:
            if not IPV4Checker("ip").check({"ip": ip_addr}):
                err_msg = "Parameter single ip_addr invalid"
                return CheckResult.make_failed(err_msg)
        if '127' == ip_addr[:3]:
            err_msg = "Parameter single ip_addr reserved"
            return CheckResult.make_failed(err_msg)
        return CheckResult.make_success()


class MacAddrChecker(StringLengthChecker):
    def __init__(self, attr_name: str, min_len: int = 0, max_len: int = sys.maxsize, required: bool = True):
        super().__init__(attr_name, min_len, max_len, required)

    def check_dict(self, data: dict) -> CheckResult:
        mac_addr = self.raw_value(data)
        if len(mac_addr) != 8 and len(mac_addr) != 17:
            err_msg = "Parameter mac_addr len invalid"
            return CheckResult.make_failed(err_msg)
        if len(mac_addr) == 8:
            mac_addr = mac_addr + ':00:00:01'
        if NetCheck.mac_addr_single_cast_check(mac_addr) is False:
            err_msg = "Parameter mac_addr error"
            return CheckResult.make_failed(err_msg)
        return CheckResult.make_success()


class SecurityLoadCfgChecker(ModelChecker):
    """配置登录规则信息配置项检查"""

    class Meta:
        fields = (
            StringChoicesChecker("enable", choices=("true", "false")),
            OrChecker(
                AndChecker(
                    NotExistsChecker("start_time"),
                    NotExistsChecker("end_time"),
                ),
                AndChecker(
                    TimeHourMinChecker("start_time"),
                    TimeHourMinChecker("end_time"),
                ),
            ),
            OrChecker(
                NotExistsChecker("ip_addr"),
                Ipv4WithMaskChecker("ip_addr"),
            ),
            OrChecker(
                NotExistsChecker("mac_addr"),
                MacAddrChecker("mac_addr"),
            ),
        )


class AlarmShieldMessageChecker(ModelChecker):
    """告警屏蔽信息检查"""

    class Meta:
        ID_PATTERN = r'^[0-9a-zA-Z]{1,32}$'
        SEVERITY_PATTERN = r"^\d$"
        INSTANCE_PATTERN = r'^[0-9a-zA-Z_.\-\s]{1,32}$'
        fields = (
            RegexStringChecker("UniquelyIdentifies", match_str=ID_PATTERN),
            RegexStringChecker("AlarmId", match_str=ID_PATTERN),
            RegexStringChecker("PerceivedSeverity", match_str=SEVERITY_PATTERN),
            RegexStringChecker("AlarmInstance", match_str=INSTANCE_PATTERN),
        )


class DateChecker(StringLengthChecker):
    """日期检查"""

    def __init__(self, attr_name=None, min_len: int = 10, max_len: int = 64, required: bool = True):
        super().__init__(attr_name, min_len, max_len, required)

    def check_dict(self, data) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        value = self.raw_value(data)
        if not self.required() and not value:
            return CheckResult.make_success()

        check_ret = self._check_format_time(value)
        if not check_ret[0]:
            msg_format = f"check format time failed! {check_ret[1]}"
            return CheckResult.make_failed(msg_format)

        return CheckResult.make_success()

    def _check_format_time(self, time):
        try:
            if ":" in time:
                datetime.datetime.strptime(time, "%Y-%m-%d %H:%M:%S")
            else:
                datetime.datetime.strptime(time, "%Y-%m-%d")
        except Exception:
            return [False, "because time does not match format."]

        return [True, ""]


class LocalIpChecker(ExistsChecker, ABC):
    # 操作类型，对应Linux Socket configuration controls中SIOCGIFADDR，用于获取ipv4地址
    REQ_GET_IPV4 = 0x8915
    DEFAULT_ETH_NAMES = b"eth0", b"eth1"

    def get_if_names(self) -> Iterable[bytes]:
        try:
            for _, name in socket.if_nameindex():
                yield name.encode(encoding="utf-8")
        except OSError:
            yield from self.DEFAULT_ETH_NAMES

    def get_ip_address(self, if_name: bytes) -> str:
        # 根据网卡名本机网卡的ipv4地址
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as obj:
            # 获取这个socket对象的文件描述符
            socket_fd = obj.fileno()
            # 填充if_name字节的256字节的二进制空间，未指定内容的空间会用字节0x00填充
            if_name_binary = struct.pack('256s', if_name)
            # 取fcntl.ioctl结果的[20:24]内容，即网卡的4字节IP地址
            return socket.inet_ntoa(fcntl.ioctl(socket_fd, self.REQ_GET_IPV4, if_name_binary)[20:24])

    def get_local_ips(self) -> Set[str]:
        ip_info_set = {"127.0.0.1"}
        for eth in self.get_if_names():
            try:
                eth_ip = self.get_ip_address(eth)
            except OSError:  # 若网卡无法分配请求的地址，会抛出OSError
                continue
            ip_info_set.add(eth_ip)
        return ip_info_set

    def check_local_ip(self, value) -> CheckResult:
        if value in self.get_local_ips():
            return CheckResult.make_failed("Local ip checkers: net ip is the same as local ip")

        return CheckResult.make_success()

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result

        value = self.raw_value(data)
        if not value:
            msg_format = "Local ip checkers: value is null."
            return CheckResult.make_failed(msg_format)

        if not isinstance(value, str):
            msg_format = f"Local ip checkers: invalid value type '{type(value)}' of {self.name()}"
            return CheckResult.make_failed(msg_format)

        return self.check_local_ip(value)
