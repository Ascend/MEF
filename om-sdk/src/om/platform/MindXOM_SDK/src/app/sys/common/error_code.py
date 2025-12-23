# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.


class ErrorCode(object):
    DEFAULT_ERROR_CODE_DIC = "Failure with unknown error code"

    ERR_SCRIPT_PARAMETER = "166"
    ERR_UPGRADE_FAILED = "173"

    CERT_NEED_SERVER_NAME = '314'

    ERR_NO_SPACE_SIZE = "1001"
    ERR_DOWNLAOD_FILES_FAILED = "1002"
    ERR_DEL_MODELFILES_FAILED = "1003"
    ERR_DOWNLOAD_MODELFILES_EXCEPTION = "1004"
    ERR_FILE_NOT_EXIST = "1005"
    ERR_CONTAINER_EXIST = "1006"
    ERR_FD_CERT_SIGNATURE_ALGORITHM = "1007"
    ERR_FD_CERT_SIGNATURE_ALGORITHM_SIZE = "1008"
    ERR_USE_PRE_FD_CERT = "1009"
    ERR_ABILITY_DISABLE = "1010"

    ExecCmdErr = "643"
    mqttInputParameterErr = "603"
    mqttResourceBusyErr = "604"
    downloadFileErr = "642"

    ErrorCodeDic = {
        "166": "Upgrade script parameter error.",
        "173": "Upgrade failed for unknown reason.",
        "314": "server name is none",
        "352": "inner exception",
        "603": "Input Parameter Err.",
        "604": "Resource is busy.",
        "642": "Download file Failed",
        "643": "exec upgrade cmd Failed",
        "1001": "no space size",
        "1002": "download file failed",
        "1003": "del file failed",
        "1004": "exception ",
        "1005": "file not exist",
        "1006": "please delete exist container before switch IEF",
        "1007": "FD certificate Signature algorithm invialid",
        "1008": "FD certificate Signature algorithm length invialid",
        "1009": "FD certificate is preset Preset, please use certificate issued by a trusted authority",
        "1010": "Ability is disable",
    }

    @staticmethod
    def generate_err_msg(err_code):
        return "ERR." + str(err_code) + ", " + ErrorCode.ErrorCodeDic.get(err_code, ErrorCode.DEFAULT_ERROR_CODE_DIC)

    @staticmethod
    def generate_err_msg_add_body(err_code, error_body):
        return "ERR." + str(err_code) + ", " + ErrorCode.ErrorCodeDic.get(err_code, ErrorCode.DEFAULT_ERROR_CODE_DIC) \
               + error_body
