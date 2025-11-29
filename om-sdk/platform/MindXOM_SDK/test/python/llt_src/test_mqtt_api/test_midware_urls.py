# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json

import mock
from pytest_mock import MockerFixture

from common.checkers.fd_param_checker import SysAssetTagChecker
from common.file_utils import FileCheck
from common.utils.url_downloader import UrlConnect
from fd_msg_process.common_redfish import CommonRedfish
from fd_msg_process.midware_urls import MidwareErrCode
from fd_msg_process.midware_urls import MidwareUris
from common.utils.result_base import Result
from lib_restful_adapter import LibRESTfulAdapter
from net_manager.manager.fd_cfg_manager import FdCfgManager
from test_mqtt_api.get_log_info import GetLogInfo
from test_mqtt_api.get_log_info import GetOperationLog
from user_manager.user_manager import UserManager
from user_manager.user_manager import SessionManager

getLog = GetLogInfo()
getOplog = GetOperationLog()


class FakeJsonClass:
    def __getattr__(self, item):
        arg = "info"
        setattr(self, item, arg)
        return arg


class FakeRequest:
    def __init__(self, method, url, headers, fields, preload_content):
        self.method = method
        self.url = url
        self.status = 200
        self.headers = headers
        self.fields = fields
        self.preload_content = preload_content

    def release_conn(self):
        return True


class FakePoolManager:
    def __init__(self):
        self.request = FakeRequest


class TestMidwareUrls:

    @staticmethod
    def test_check_external_param_should_ok():
        param_data = {"asset_tag": "abcd123A[]"}
        payload_publish = {
            "topic": "tag",
            "percentage": "0%",
            "result": "failed",
            "reason": ""
        }
        err_info = "Set system tag failed. %s"
        ret = MidwareUris.check_external_param(SysAssetTagChecker, param_data, payload_publish, err_info)
        assert ret[0] == 0

    @staticmethod
    def test_check_external_param_should_failed():
        param_data = {"asset_tag": "abcd123A[]。"}
        payload_publish = {
            "topic": "tag",
            "percentage": "0%",
            "result": "failed",
            "reason": ""
        }
        err_info = "Set system tag failed. %s"
        ret = MidwareUris.check_external_param(SysAssetTagChecker, param_data, payload_publish, err_info)
        assert ret[0] == -1

    @staticmethod
    def test_check_capacity_bytes_valid_pass_empty_param():
        ret = MidwareUris.check_capacity_bytes_valid([])
        assert ret

    @staticmethod
    def test_check_capacity_bytes_valid_pass_param_should_ok():
        permit_unit = 512 * 1024 * 1024
        partition_cfg = [
            {
                "capacity_bytes": permit_unit,
            }
        ]
        ret = MidwareUris.check_capacity_bytes_valid(partition_cfg)
        assert ret

    @staticmethod
    def test_check_capacity_bytes_valid_pass_param_should_failed():
        permit_unit = 512 * 1024 * 1024 + 1024
        partition_cfg = [
            {
                "capacity_bytes": permit_unit,
            }
        ]
        ret = MidwareUris.check_capacity_bytes_valid(partition_cfg)
        assert not ret

    @staticmethod
    @mock.patch.object(SessionManager, "get_all_info", mock.Mock(return_value={"message": {"SessionTimeout": 40}}))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_session_timeout():
        MidwareUris.resp_json_sys_info["security_policy"] = dict()
        MidwareUris.get_session_timeout()
        assert MidwareUris.resp_json_sys_info["security_policy"]["session_timeout"] == "40"

    @staticmethod
    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface',
                       mock.Mock(return_value={"message": {"CertAlarmTime": 40}}))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_cert_alarm_time():
        MidwareUris.resp_json_sys_info["security_policy"] = dict()
        MidwareUris.get_cert_alarm_time()
        assert MidwareUris.resp_json_sys_info["security_policy"]["cert_alarm_time"] == "40"

    @staticmethod
    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface',
                       mock.Mock(return_value={"message": {"load_cfg": []}}))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_security_load_config():
        MidwareUris.resp_json_sys_info["security_policy"] = dict()
        MidwareUris.get_security_load_config()
        assert not MidwareUris.resp_json_sys_info["security_policy"]["security_load"]

    @staticmethod
    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface',
                       mock.Mock(return_value={"message": {}}))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_lte_info_failed():
        MidwareUris.get_lte_info()
        assert not MidwareUris.resp_json_sys_status["lte_info"]

    @staticmethod
    @mock.patch("json.loads", mock.Mock(
        return_value={"default_gateway": "default gateway", "lte_enable": "lte enable", "sim_exist": "sim exist",
                      "state_lte": "state lte", "state_data": "state data",
                      "network_signal_level": "network signal level",
                      "network_type": "network type", "ip_addr": "ip addr"}))
    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface',
                       mock.Mock(return_value={"message": {"lte_enable": True}}))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(MidwareUris, 'get_apn_info', mock.Mock(
        return_value={"apn_name": "apn name", "apn_user": "apn user",
                      "auth_type": "auth type", "mode_type": "mode type"}))
    def test_get_lte_info_ok():
        MidwareUris.get_lte_info()
        assert MidwareUris.resp_json_sys_status["lte_info"] == [
            {"default_gateway": "default gateway", "lte_enable": "lte enable", "sim_exist": "sim exist",
             "state_lte": "state lte", "state_data": "state data",
             "network_signal_level": "network signal level",
             "network_type": "network type", "ip_addr": "ip addr",
             "apn_info": [{"apn_name": "apn name", "apn_user": "apn user",
                           "auth_type": "auth type", "mode_type": "mode type"}]}
        ]

    @staticmethod
    @mock.patch("json.loads", mock.Mock(
        return_value={"apn_name": "apn name", "apn_user": "apn user",
                      "auth_type": "auth type", "mode_type": "mode type"}))
    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface', mock.Mock(return_value={"message": None}))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    def test_get_apn_info():
        ret = MidwareUris.get_apn_info()
        assert ret == {"apn_name": "apn name", "apn_user": "apn user",
                       "auth_type": "auth type", "mode_type": "mode type"}

    @staticmethod
    @mock.patch("json.loads", mock.Mock(return_value={"web_access": "web access", "ssh_access": "ssh access"}))
    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface', mock.Mock(return_value={"message": None}))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    def test_get_access_control_info():
        MidwareUris.resp_json_sys_info["security_policy"] = dict()
        MidwareUris.get_access_control_info()
        assert MidwareUris.resp_json_sys_info["security_policy"]["web_access"] == "web access" and \
               MidwareUris.resp_json_sys_info["security_policy"]["ssh_access"] == "ssh access"

    @staticmethod
    @mock.patch("json.loads", mock.Mock(
        return_value={"ClientEnabled": "ClientEnabled", "ServerEnabled": "ServerEnabled", "Port": "Port",
                      "NTPRemoteServers": "NTPRemoteServers", "NTPRemoteServersbak": "NTPRemoteServersbak",
                      "NTPLocalServers": "NTPLocalServers"}))
    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface', mock.Mock(return_value={"message": None}))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_ntp_config_ok():
        assert MidwareUris.get_ntp_config()[0] == 0

    @staticmethod
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=False))
    def test_get_ntp_config_failed():
        ret = MidwareUris.get_ntp_config()
        assert ret[0] == -1 and getLog.get_log() is not None

    @staticmethod
    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface', mock.Mock(return_value={"message": None}))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_set_ntp_config_ok():
        ntp_server_ip = "1.1.1.1"
        ret = MidwareUris.set_ntp_config(ntp_server_ip)
        assert ret[0] == 0 and ntp_server_ip in ret[1]

    @staticmethod
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=False))
    def test_set_ntp_config_failed():
        ntp_server_ip = "1.1.1.1"
        ret = MidwareUris.set_ntp_config(ntp_server_ip)
        assert ret[0] == -1 and \
               str(MidwareErrCode.midware_config_ntp_common_err) in ret[1] and \
               getLog.get_log() is not None

    @staticmethod
    @mock.patch("os.path.exists", mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, "get_ntp_config", mock.Mock(return_value=[0, {"NTPRemoteServers": "1.1.1.1"}]))
    @mock.patch.object(FdCfgManager, "get_cur_fd_ip", mock.Mock(return_value="1.1.1.1"))
    def test_snyc_ntp_config():
        assert MidwareUris.snyc_ntp_config()[0] == 0

    def test_midware_generate_err_msg(self):
        ret = MidwareErrCode.midware_generate_err_msg(1, "info")
        assert ret == 'ERR.1, info'

    @getOplog.clear_common_log
    @mock.patch.object(FdCfgManager, "get_cur_fd_ip", mock.Mock(return_value=""))
    def test_set_operation_log(self):
        MidwareUris.set_operation_log("message")
        ret = getOplog.get_log()
        assert ret is not None

    def test_check_json_request(self):
        ret = MidwareUris.check_json_request(json.dumps({"req": 1}))
        assert ret[0] == 0

    def test_get_log_collect_publish_template(self):
        ret = MidwareUris.get_log_collect_publish_template("all")
        assert ret == {'module': 'all',
                       'percentage': '0%',
                       'reason': '',
                       'result': 'processing',
                       'type': 'all'}

    @mock.patch.object(UrlConnect, 'get_context', mock.Mock(return_value=Result(False)))
    @mock.patch("os.path.getsize", mock.Mock(return_value=1))
    def test_https_upload_should_failed_when_ssl_error(self):
        ret = MidwareUris.https_upload({"https_server": {"url": "123 4", }}, "path")
        assert ret[0] == -1 and ret[1] == "get client ssl context error."

    @mock.patch.object(FileCheck, 'check_path_is_exist_and_valid', mock.Mock(return_value=Result(False)))
    @mock.patch.object(UrlConnect, 'get_context', mock.Mock(return_value=True))
    @mock.patch("os.path.getsize", mock.Mock(return_value=1))
    def test_https_upload_should_failed_when_path_invalid(self):
        ret = MidwareUris.https_upload({"https_server": {"url": "123 4", }}, "path")
        assert ret[0] == -1 and "path invalid" in ret[1]

    @mock.patch('urllib3.PoolManager', mock.Mock(return_value=FakePoolManager()))
    @mock.patch('builtins.open', mock.MagicMock())
    @mock.patch.object(FileCheck, 'check_path_is_exist_and_valid', mock.Mock(return_value=Result(True)))
    @mock.patch.object(UrlConnect, 'get_context', mock.Mock(return_value=Result(True)))
    @mock.patch("os.path.getsize", mock.Mock(return_value=1))
    def test_https_upload_should_ok(self):
        ret = MidwareUris.https_upload({"https_server": {"url": "123 4", "user_name": "a", "password": "b"}}, "path")
        assert ret[0] == 0 and ret[1] == "upload log to fd success"

    def test_get_alarm_health_status(self):
        ret = MidwareUris.get_alarm_health_status()
        assert ret == "OK"

    def test_health_string_to_bool_should_failed(self):
        ret = MidwareUris.health_string_to_bool("failed")
        assert not ret

    def test_health_string_to_bool_should_ok(self):
        ret = MidwareUris.health_string_to_bool("OK")
        assert ret

    def test_xb_to_b_when_string_is_int(self):
        ret = MidwareUris.xb_to_b(1)
        assert ret == 1

    def test_xb_to_b(self):
        ret = MidwareUris.xb_to_b("1MB")
        assert ret == 1048576

    def test_b_to_gb(self):
        ret = MidwareUris.b_to_gb("1GB")
        assert ret == "1GB"

    def test_b_to_gb_when_parm_is_not_string(self):
        ret = MidwareUris.b_to_gb(1024 * 1024 * 1024)
        assert ret == '1.0GB'

    @mock.patch.object(FileCheck, 'check_path_is_exist_and_valid', mock.Mock(return_value=Result(False)))
    def test_get_inactive_profile_should_failed(self):
        ret = CommonRedfish.get_inactive_profile()
        assert not ret

    @mock.patch('glob.iglob', mock.Mock(return_value=["/home/data/config/redfish/112.prf"]))
    @mock.patch.object(FileCheck, 'check_path_is_exist_and_valid', mock.Mock(return_value=Result(True)))
    def test_get_inactive_profile_should_ok(self):
        ret = CommonRedfish.get_inactive_profile()
        assert ret == "112"

    @getLog.clear_common_log
    def test_get_edge_system_info_should_failed(self):
        MidwareUris.get_edge_system_info()
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_edge_system_info_should_ok(self):
        MidwareUris.get_edge_system_info()
        assert MidwareUris.resp_json_sys_status["system"]['health_status'] == 'unknown'

    @getLog.clear_common_log
    def test_get_cpu_summary_info_should_failed(self):
        MidwareUris.get_cpu_summary_info()
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_cpu_summary_info_should_ok(self):
        MidwareUris.get_cpu_summary_info()
        assert MidwareUris.resp_json_sys_info["system"]["net_manager_domain"] is None

    @getLog.clear_common_log
    def test_get_memory_summary_info_should_failed(self):
        MidwareUris.get_memory_summary_info()
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_memory_summary_info_should_ok(self):
        MidwareUris.get_cpu_summary_info()
        assert MidwareUris.resp_json_sys_info["system"]["net_manager_domain"] is None

    @getLog.clear_common_log
    def test_get_every_extend_info_should_failed(self):
        MidwareUris.get_every_extend_info(1)
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch("json.loads", mock.Mock(
        return_value={"Name": "name", "DeviceClass": "DeviceClass", "DeviceName": "DeviceName",
                      "Manufacturer": "Manufacturer", "Model": "Model", "SerialNumber": "SerialNumber",
                      "FirmwareVersion": "FirmwareVersion", "Location": "Location",
                      "Status": {"State": 12, "Health": "bad", }, }))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_every_extend_info_should_ok(self):
        MidwareUris.get_every_extend_info("disk0")
        assert MidwareUris.resp_json_sys_status["extended_devices"] == [
            {'name': 'name', 'status': {'health': False, 'state': 12}}]

    @getLog.clear_common_log
    def test_get_extend_info_should_failed(self):
        MidwareUris.get_extend_info()
        ret = getLog.get_log()
        assert ret is not None

    @getLog.clear_common_log
    def test_get_ntp_info_should_failed(self):
        MidwareUris.get_ntp_info()
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch("os.path.exists", mock.Mock(return_value=True))
    @mock.patch("json.loads", mock.Mock(return_value={"ClientEnabled": None, }))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_ntp_info_should_ok(self):
        MidwareUris.get_ntp_info()
        assert MidwareUris.resp_json_sys_info["ntp_server"] == {'alternate_server': None,
                                                                'preferred_server': None,
                                                                'service_enabled': False,
                                                                'sync_net_manager': None}

    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface',
                       mock.Mock(return_value={"message": {"AiTemperature": 700, }, }))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_ai_temp_info_should_ok(self):
        ret = MidwareUris.get_ai_temp_info()
        assert ret == 700

    @mock.patch.object(LibRESTfulAdapter, 'lib_restful_interface',
                       mock.Mock(return_value={"message": {"AiTemperature": 700, }, }))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=False))
    def test_get_ai_temp_info_should_failed(self):
        ret = MidwareUris.get_ai_temp_info()
        assert not ret

    def test_get_extend_location_info(self):
        ret = MidwareUris.get_extend_location_info(1)
        assert not ret

    @getLog.clear_common_log
    def test_get_npu_info_should_failed(self):
        MidwareUris.get_npu_info()
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch("os.path.exists", mock.Mock(return_value=True))
    @mock.patch("json.loads", mock.Mock(
        return_value={"Manufacturer": "Manufacturer", "Status": {"State": 12, "Health": "bad", },
                      "Oem": {"Count": 1, "Capability": {"Calc": 1, "Ddr": 2, },
                              "OccupancyRate": {"AiCore": "AiCore", "AiCpu": "AiCpu", "CtrlCpu": "CtrlCpu",
                                                "DdrUsage": "DdrUsage", "DdrBandWidth": "DdrBandWidth", }}}))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'get_extend_location_info', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'health_string_to_bool', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'get_ai_temp_info', mock.Mock(return_value=True))
    def test_get_npu_info_should_ok(self):
        MidwareUris.get_npu_info()
        assert MidwareUris.resp_json_sys_status["ai_processors"] == []

    @getLog.clear_common_log
    def test_get_every_simple_storage_info_should_failed(self):
        MidwareUris.get_every_simple_storage_info(1)
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch("json.loads", mock.Mock(
        return_value={"Name": "namev", "Description": "dDescriptionv",
                      "Devices": [{"Name": "damev", "Manufacturer": "danufacturerv", "Model": "dodelv",
                                   "CapacityBytes": "dapacityBytesv", "PartitionStyle": "dartitionStylev",
                                   "Location": "docationv", "LeftBytes": "deftBytesv",
                                   "Status": {"State": 12, "Health": "bad", }, }], }))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'health_string_to_bool', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'xb_to_b', mock.Mock(return_value=True))
    def test_get_every_simple_storage_info_should_ok(self):
        MidwareUris.get_every_simple_storage_info(1)
        assert MidwareUris.resp_json_sys_info["simple_storages"] == [{'description': 'dDescriptionv',
                                                                      'devices': [{'capacity_bytes': 1,
                                                                                   'location': 'docationv',
                                                                                   'manufacturer': 'danufacturerv',
                                                                                   'model': 'dodelv',
                                                                                   'name': 'damev',
                                                                                   'partition_style': 'dartitionStylev',
                                                                                   'reserved_bytes': 104857600}],
                                                                      'name': 1,
                                                                      'type': 'namev'}]

    @getLog.clear_common_log
    def test_get_simple_storages_info_should_failed(self):
        MidwareUris.get_simple_storages_info()
        ret = getLog.get_log()
        assert ret is not None

    @getLog.clear_common_log
    def test_get_every_partition_info_should_failed(self):
        MidwareUris.get_every_partition_info(2)
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch("json.loads", mock.Mock(
        return_value={"Name": "name",
                      "Links": [{"DeviceName": "DeviceNamev", "Location": "Locationv", "Device": "Devicev"}],
                      "CapacityBytes": "CapacityBytesv", "FileSystem": "FileSystemv", "MountPath": "MountPathv",
                      "Primary": "Primaryv", "FreeBytes": "FreeBytesv", "Status": {"Health": "Health", }, }))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'health_string_to_bool', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'xb_to_b', mock.Mock(return_value=True))
    def test_get_every_partition_info_should_ok(self):
        MidwareUris.get_every_partition_info(2)
        assert MidwareUris.resp_json_sys_status["partitions"] == [
            {'free_bytes': 1, 'health': True, 'logic_name': None, 'name': 'name'}]

    @getLog.clear_common_log
    def test_get_partitions_info_should_failed(self):
        MidwareUris.get_partitions_info()
        ret = getLog.get_log()
        assert ret is not None

    @getLog.clear_common_log
    def test_get_every_eth_info_should_failed(self):
        MidwareUris.get_every_eth_info(1)
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch("json.loads", mock.Mock(
        return_value={"Name": "name", "Description": "Description", "PermanentMACAddress": "PermanentMACAddress",
                      "MACAddress": "MACAddress", "InterfaceEnabled": "InterfaceEnabled",
                      "Oem": {"WorkMode": "WorkMode", "DeviceName": "DeviceName", "Location": "Location",
                              "AdapterType": "AdapterType", "LteDataSwitchOn": "LteDataSwitchOn",
                              "Statistic": {"SendPackages": 1, "RecvPackages": 2,
                                            "ErrorPackages": 3, "DropPackages": 4, }},
                      "IPv4Addresses": [{"Address": "Address", "SubnetMask": "SubnetMask", "Gateway": "Gateway",
                                         "AddressOrigin": "AddressOrigin", "Tag": "Tag"}], "NameServers": "NameServers",
                      "CapacityBytes": "CapacityBytes", "FileSystem": "FileSystem", "MountPath": "MountPath",
                      "Primary": "Primary", "FreeBytes": "FreeBytes", "Status": {"Health": "Health", },
                      "LinkStatus": "LinkStatus", }))
    @mock.patch.object(CommonRedfish, 'update_json_of_list', mock.Mock(return_value=Result(True)))
    @mock.patch.object(MidwareUris, 'get_extend_location_info', mock.Mock(return_value=Result(True)))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'health_string_to_bool', mock.Mock(return_value=True))
    @mock.patch.object(MidwareUris, 'xb_to_b', mock.Mock(return_value=True))
    def test_get_every_eth_info_should_ok(self):
        MidwareUris.get_every_eth_info(2)
        assert MidwareUris.resp_json_sys_status["eth_statistics"] == [{'id': 2,
                                                                       'link_status': 'LinkDown',
                                                                       'statistics': {'drop_packages': 4,
                                                                                      'error_packages': 3,
                                                                                      'recv_packages': 2,
                                                                                      'send_packages': 1},
                                                                       'work_mode': 'WorkMode'}]

    @getLog.clear_common_log
    def test_get_eth_info_should_failed(self):
        MidwareUris.get_eth_info()
        ret = getLog.get_log()
        assert ret is not None

    @getLog.clear_common_log
    def test_get_dns_and_host_map_info_should_failed(self):
        MidwareUris.get_dns_and_host_map_info()
        ret = getLog.get_log()
        assert ret is not None

    def test_get_dns_and_host_map_info_should_ok(self, mocker: MockerFixture):
        mock_rest_ret = {
            "status": 200,
            "message": {
                "static_host_list": [{"name": "static_host_list", "ip_address": "10.10.10.10"}],
                "name_server": [{"ip_address": "10.10.10.10"}],
            },
        }
        mocker.patch.object(LibRESTfulAdapter, 'lib_restful_interface', return_value=mock_rest_ret)
        mocker.patch.object(CommonRedfish, 'check_status_is_ok', return_value=True)
        mocker.patch.object(FdCfgManager, 'get_cur_fd_server_name', return_value="fd.fusiondirector.huawei.com")
        MidwareUris.get_dns_and_host_map_info()
        assert MidwareUris.resp_json_sys_info["name_server"] == [{"ip_address": "10.10.10.10"}]

    @getLog.clear_common_log
    @mock.patch.object(UserManager, "get_all_info", mock.Mock(return_value={"status": 400, "message": {}}))
    def test_get_passwd_validity_info_should_failed(self):
        MidwareUris.get_passwd_validity_info()
        ret = getLog.get_log()
        assert ret is not None

    @getLog.clear_common_log
    @mock.patch.object(UserManager, "get_all_info", mock.Mock(return_value={"status": 400, "message": {}}))
    def test_get_accounts_info_should_failed(self):
        MidwareUris.get_accounts_info()
        ret = getLog.get_log()
        assert ret is not None

    @mock.patch.object(UserManager, "get_all_info", mock.Mock(
        return_value={"message": {"result": "result", }, }))
    @mock.patch.object(CommonRedfish, 'check_status_is_ok', mock.Mock(return_value=True))
    def test_get_accounts_info_should_ok(self):
        MidwareUris.get_accounts_info()
        assert MidwareUris.resp_json_sys_info["accounts"] == "result"

    def test__get_alarm_info_should_failed_when_dict_is_none(self):
        ret = MidwareUris._get_alarm_info(1)
        assert not ret

    def test__get_alarm_info_should_ok(self):
        alarm_item_info = {
            "id": "00000000",
            "name": "Drive Overtemperature",
            "dealSuggestion": "1. Check whether a TEC alarm is generated. @#AB"
                              "2. Check whether the ambient temperature of the device exceeds 60°C.@#AB"
                              "3. Restart the system. Then check whether the alarm is cleared.@#AB"
                              "4. Contact Vendor technical support.",
            "detailedInformation": "The component temperature exceeds the threshold.",
            "reason": " The ambient temperature is excessively high.",
            "impact": " The system reliability may be affected."
        }

        temp_dict = {
            "MAJOR_VERSION": "2",
            "MINOR_VERSION": "8",
            "AUX_VERSION": "0",
            "EventSuggestion": [
                alarm_item_info,
            ]
        }
        MidwareUris.alarm_info_dict = temp_dict
        ret = MidwareUris._get_alarm_info("00000000")
        assert ret == alarm_item_info
