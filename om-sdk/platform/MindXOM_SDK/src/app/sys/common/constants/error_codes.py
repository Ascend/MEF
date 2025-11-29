# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from common.constants.messages_codes import MessagesCode


class ErrorCode(MessagesCode):
    def __init__(self, code, message_key):
        super().__init__(code, message_key)

    def __repr__(self) -> str:
        return 'ErrorCode [code:{}, messageKey:{}]'.format(self.code, self.messageKey)


class CommonErrorCodes(ErrorCode):
    # 公共错误代码 code: 100000~109999
    ERROR_ARGUMENT_VALUE_NOT_EXIST = ErrorCode(100000, "Error argument [{}] empty")
    ERROR_FILE_PATH_FORMAT = ErrorCode(100001, "File path error.")
    ERROR_FILE_PATH_NULL = ErrorCode(100002, "File path is null.")
    ERROR_FILE_PATH_LONG = ErrorCode(100003, "File path is too long.")
    ERROR_FILE_PATH_SIZE = ErrorCode(100004, "Path free size error.")
    ERROR_TMP_DIR_SPACE = ErrorCode(100005, "The space of the /run directory is low")  # run目录空间不足
    ERROR_ARGUMENT_FORMAT = ErrorCode(100006, "Argument [{}] format error.")  # 输入的信息错误
    ERROR_FILE_PATH_NOT_EXIST = ErrorCode(100007, "File path does not exist.")  # 文件路径不存在
    ERROR_RETURN_VALUE = ErrorCode(100008, "The return value is empty.")  # 返回值为空
    ERROR_COLLECTION_LENGTH = ErrorCode(100009, "The length of the collection is not as expected")  # 集合长度不符合预期
    ERROR_ARGUMENT_TYPE = ErrorCode(100010, "Argument [{}] type error")  # 参数的类型错误
    ERROR_INTERNAL_SERVER = ErrorCode(100011, "Internal server error")
    ERROR_ARGUMENT_NOT_EXIST = ErrorCode(100012, "Argument [{}] not exist")
    ERROR_ARGUMENT_NUMBER_WRONG = ErrorCode(100013, "Argument number wrong")
    ERROR_REQUEST_ARGUMENT_JSON = ErrorCode(100014, "Request message should json")
    ERROR_ARGUMENT_VALUE_WRONG = ErrorCode(100015, "Argument [{}] value wrong")  # 参数值错误
    ERROR_PARAM_RANGE_WRONG = ErrorCode(100016, "Param [{}] range wrong.")
    ERROR_MODIFY_FAILED = ErrorCode(100017, "Modify failed.")  # 当前用户名被占用
    ERROR_DIR_SPACE_LOW = ErrorCode(100018, "The [{}] space directory is low.")  # 磁盘空间不足
    ERROR_CMD_EXEC_WRONG = ErrorCode(100019, "Command execution error.")  # 命令执行错误
    ERROR_SYSTEM_UPGRADED_WRONG = ErrorCode(100020, "The System upgrade.")  # 系统正在升级
    ERROR_RESTORE_DEFAULTS_WRONG = ErrorCode(100021, "Restore defaults failed.")  # 恢复出厂设置失败
    ERROR_FILE_PATH_SOFTLINK = ErrorCode(100022, "File path is softlink.")  # 文件路径存在软连接
    ERROR_FILE_PATH_CHECK = ErrorCode(100023, "File path check.")  # 文件路径检查抛异常
    ERROR_PARAMETER_INVALID = ErrorCode(100024, "Parameter is invalid.")  # 参数无效
    # root账号的密码是默认密码抛异常
    ERROR_ROOT_PASS_WORD_FLAG_EXIST_CHECK = ErrorCode(100025, "The root account has not changed the default password, "
                                                              "please refer to the information to modify.")
    # 操作锁获取失败，只支持串行操作
    ERROR_ROOT_OPERATE_LOCKED = ErrorCode(100026, "Get operation lock failed.")
    # 操作者非法
    ERROR_ROOT_OPERATE_ILLEGAL = ErrorCode(100027, "The operator is illegal.")
    # 并发冲突
    OPERATE_IS_BUSY = ErrorCode(100028, "The operation is busy.")
    # 服务启动中
    ERROR_SERVICE_IS_STARTING = ErrorCode(100029, "The service is starting.")
    # 服务启动异常
    ERROR_SERVICE_STARTUP_FAILED = ErrorCode(100030, "The service startup failed.")
    # 正在恢复最小系统
    ERROR_RECOVER_OS_IS_RUNNING = ErrorCode(100031, "Recover mini os is running.")
    ERROR_FILE_LINES = ErrorCode(100032, "File lines exceed limit.")


# 日志收集错误代码110000~110099
class LogErrorCodes(CommonErrorCodes):
    ERROR_LOG_COLLECT = ErrorCode(110001, "Collect log failed.")  # 日志收集失败


# 文件错误代码110100~110199
class FileErrorCodes(CommonErrorCodes):
    ERROR_FILE_NO_SECTION = ErrorCode(110100, "Not found section.")  # 没有找到相应的section
    ERROR_FILE_SECTION_EXIST = ErrorCode(110101, "Section exist.")  # section已经存在
    ERROR_FILE_NO_OPTION = ErrorCode(110102, "Not found option.")  # 没有找到对应的option
    ERROR_FILE_OPEN_ERROR = ErrorCode(110103, "File open error.")  # 打开文件错误
    ERROR_FILE_CONTENT_NULL_ERROR = ErrorCode(110104, "File content is null.")
    ERROR_FILE_NOT_EXIST_ERROR = ErrorCode(110105, "File not exist.")
    ERROR_FILE_SIZE_INVALID_ERROR = ErrorCode(110106, "File size is invalid.")
    ERROR_UPLOAD_FLAG = ErrorCode(110107, "Forbidden to upload new file")  # 10分钟内重复上传文件


# 用户错误代码110200~110299
class UserManageErrorCodes(CommonErrorCodes):
    ERROR_SESSION_NOT_EXIST = ErrorCode(110200, "Session not exist.")  # session没有找到。
    ERROR_SESSION_NOT_FOUND = ErrorCode(110201, "Session not found.")  # session没有找到。
    ERROR_SESSION_TIME_LIMIT = ErrorCode(110202, "Session time limit exceed.")  # session过期。
    ERROR_DB_OPERATION_FAILED = ErrorCode(110203, "Database operation failed.")  # 数据库操作失败。
    ERROR_PASSWORD_VALID_DAY = ErrorCode(110204, "The password has expired.")  # 密码已过期。
    ERROR_REQUEST_IP_ADDR = ErrorCode(110205, "ip address error.")  # ip地址无效。
    ERROR_USER_LOCK_STATE = ErrorCode(110206, "User lock state locked, [{}] seconds left.")  # 用户处于锁定状态
    ERROR_USER_NOT_MATCH_PASSWORD = ErrorCode(110207, "The user name or password error.")  # 用户名与密码不匹配
    ERROR_SECURITY_CFG_NOT_MEET = ErrorCode(110208, "Security load check failed.")  # 登录规则不满足
    ERROR_PARAM_RANGE = ErrorCode(110209, "Param [{}] wrong range.")  # session范围错误。
    ERROR_USER_NOT_FOUND = ErrorCode(110210, "User not found.")  # 用户不存在
    ERROR_PASSWORD_DUPLICATE = ErrorCode(110211, "New password is duplicate with the old password.")  # 新密码与老密码重复
    ERROR_USERNAME_OS_MODIFY = ErrorCode(110212, "Modify os username failed.")  # 修改os的用户名失败
    ERROR_PASSWORD_OS_MODIFY = ErrorCode(110213, "Modify os password failed.")  # 修改os的密码失败
    ERROR_PASSWORD_COMPARED_USERNAME_REVERSAL = ErrorCode(110214, "The new password cannot be the same as the user name"
                                                                  " or the user name in reverse order.")  # 密码不能倒写
    ERROR_PASSWORD_TWO_INCONSISTENT = ErrorCode(110215, "Two passwords are inconsistent.")  # 两次密码不一致
    ERROR_USER_AUTH_FAILED = ErrorCode(110216, "User authentication failed.")  # 用户认证失败
    ERROR_USERNAME_USED = ErrorCode(110217, "Currently used by process.")  # 当前用户名被占用
    ERROR_OS_USER_NOT_EXIST = ErrorCode(110218, "OS user does not exist.")
    ERROR_PASSWORD_IS_INITIAL = ErrorCode(110219, "User [{}] log in failed.")
    PARAMETER_IS_WRONG = ErrorCode(110220, "Parameter is wrong.")
    ERROR_PASSWORD_IS_DEFAULT = ErrorCode(110221, "Can not set to default password.")
    ERROR_PASSWORD_IS_WEAK = ErrorCode(110222, "The new password strength is too weak.")
    ERROR_EDGE_CONFIG_NOT_FOUND = ErrorCode(110223, "Edge_config not found.")  # session没有找到。
    INTERNAL_ERROR = ErrorCode(110224, "Internal error.")  # 内部错误。


class SecurityServiceErrorCodes(CommonErrorCodes):
    ERROR_CERTIFICATE_ABOUT_TO_EXPIRE = ErrorCode(110301, "The certificate is about to expire.")
    ERROR_CERTIFICATE_INIT_FAILED = ErrorCode(110302, "Importing a custom certificate init failed")
    ERROR_CERTIFICATE_TOO_LARGE = ErrorCode(110303, "Certificate is too large")
    ERROR_CERTIFICATE_OR_PASSWORD_VERIFY_FAILED = ErrorCode(110304, "Certificate or password verification failed")
    ERROR_CERTIFICATE_IS_NOT_VALID = ErrorCode(110305, "Certificate is not valid")
    ERROR_IMPORT_CERTIFICATE_INTERNAL_ERROR = ErrorCode(110306, "Importing a custom certificate failed")
    ERROR_IMPORT_CERTIFICATE_LEGALITY_RISKY = ErrorCode(110307,
                                                        "Importing a custom certificate success but legality is risky")
    ERROR_CERTIFICATE_HAS_EXPIRED = ErrorCode(110308, "The certificate has expired")
    ERROR_CERTIFICATE_DOES_NOT_MATCH = ErrorCode(110309, "The certificate does not match the private key")
    ERROR_CERTIFICATE_WAIT_TIME = ErrorCode(110310, "Please wait 2 hours and download it again")
    ERROR_CERTIFICATE_IS_INVALID = ErrorCode(110311, "The certificate is invalid")
    ERROR_CERTIFICATE_IS_NOT_X509V3 = ErrorCode(110312, "The certificate is not X.509 v3 format")
    ERROR_CERTIFICATE_IS_NOT_RSA_EC = ErrorCode(110313, "The certificate pubkey type is not RSA or EC")
    ERROR_CERTIFICATE_RSA_LEN_INVALID = ErrorCode(110314, "The RSA pubkey length is not greater or equal to 3072")
    ERROR_CERTIFICATE_EC_LEN_INVALID = ErrorCode(110315, "The EC pubkey length is not greater or equal to 256")
    ERROR_CERTIFICATE_SIGN_ALG_INVALID = ErrorCode(110316, "The signature algorithm is not sha256WithRSAEncryption or"
                                                           "sha384WithRSAEncryption or sha512WithRSAEncryption or"
                                                           "ecdsa-with-SHA256")
    ERROR_CERTIFICATE_IS_NOT_CA = ErrorCode(110317, "The cert is not CA")
    ERROR_CERTIFICATE_KEY_SIGN = ErrorCode(110318, "Cannot used for verifying signatures on public key certificates")
    ERROR_CERTIFICATE_CHAIN_NUMS_MAX = ErrorCode(110319, "The number of CA certificate chains is greater than 10.")
    ERROR_CERTIFICATE_CA_SIGNATURE_INVALID = ErrorCode(110320, "The signature of CA certificate is invalid")


class NfsServiceErrorCodes(CommonErrorCodes):
    ERROR_MOUNT_PATH_EXISTED = ErrorCode(110501, "Mount path already exists.")
    ERROR_MOUNT_PATH_NOT_IN_WHITE_LIST = ErrorCode(110502, "Mount path not in white list.")
    ERROR_MOUNT_PATH_SUBDIRECTORY_RELATIONSHIP = ErrorCode(
        110503, "There is a subdirectory relationship between Mount Path and Mounted path.")
    ERROR_CONFIGURATION_EXCEEDS_LIMIT = ErrorCode(110504, "NFS configuration exceeds limit.")
    ERROR_MOUNT_FAILED = ErrorCode(110505, "Operate NFS failed.")
    ERROR_MOUNT_TIME_OUT = ErrorCode(110506, "Operate time out.")
    ERROR_UNMOUNT_PATH_NOT_EXISTED = ErrorCode(110507, "Unmount path does not exist.")
    ERROR_FILESYSTEM_ERROR = ErrorCode(110508, "FileSystem is error")
    ERROR_GET_LOCAL_INFO_FAILED = ErrorCode(110509, "NFS local mount path to get mount local info failed")


class PartitionErrorCodes(CommonErrorCodes):
    ERROR_OPERATE_DEVICE = ErrorCode(110600, "Operate device failed.")
    ERROR_PARTITION_NAME = ErrorCode(110601, "Partition name is wrong.")
    ERROR_PARTITION_IS_SYSTEM_PARTITION = ErrorCode(110602, "Partition is in system partitions.")
    ERROR_CONVERSE_SITE_PARTITION_FAILED = ErrorCode(110603, "Converse site partition failed.")
    ERROR_UNMOUNT_FAILED = ErrorCode(110604, "Unmount partition failed.")
    ERROR_DELETE_PARTITION_FAILED = ErrorCode(110605, "Delete partition failed.")
    ERROR_MOUNT_PATH_NOT_IN_WHITE_LIST = ErrorCode(110606, "Mount path not in white list.")
    ERROR_DISK_IS_MOUNTED = ErrorCode(110607, "The disk is mounted.")
    ERROR_MOUNT_PATH_INVALID = ErrorCode(110608, "Mount path is invalid.")
    ERROR_PATH_MOUNTED = ErrorCode(110609, "Path has been mounted")
    ERROR_MOUNT_FAILED = ErrorCode(110610, "Mount partition failed.")
    ERROR_SPACE_NOT_ENOUGH = ErrorCode(110611, "The disk not enough space.")
    ERROR_SYSTEM_BUSY = ErrorCode(110612, "System busy, will not make a filesystem here!")
    ERROR_CREATE_PARTITION_FAILED = ErrorCode(110613, "Create partition failed.")
    ERROR_DISK_IS_NOT_MOUNTED = ErrorCode(110614, "The disk is not mounted.")
    ERROR_MOUNT_IS_EXISTED = ErrorCode(110615, "Mount path is existed.")
    ERROR_DISK_TYPE = ErrorCode(110616, "The file system type of the partition is not ext4.")
    ERROR_DRIVE_LETTER_NOT_EXISTED = ErrorCode(110617, "Drive letter does not exist.")
    ERROR_PARTITION_NUM = ErrorCode(110618, "The number of partitions exceeds the limit.")
    ERROR_DOCKER_UNMOUNT_FAILED = ErrorCode(110619, "Containers have not been deleted or docker path is occupied.")
