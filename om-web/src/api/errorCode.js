/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

export default {
  COMMON: {
    ERROR_ARGUMENT_EMPTY: 100000, // Error argument [{}] empty
    ERROR_FILE_PATH_FORMAT: 100001, // File path error.
    ERROR_FILE_PATH_NULL: 100002, // File path is null.
    ERROR_FILE_PATH_LONG: 100003, // File path is too long.
    ERROR_FILE_PATH_SIZE: 100004, // Path free size error.
    ERROR_ARGUMENT_FORMAT: 100006, // Argument [{}] format error.
    ERROR_FILE_PATH_NOT_EXIST: 100007, // File path does not exist.
    ERROR_RETURN_VALUE: 100008, // The return value is empty.
    ERROR_COLLECTION_LENGTH: 100009, // The length of the collection is not as expected
    ERROR_ARGUMENT_TYPE: 100010, // Argument [{}] type error
    ERROR_INTERNAL_SERVER: 100011, // Internal server error
    ERROR_ARGUMENT_NUMBER_WRONG: 100013, // Argument number wrong
    ERROR_REQUEST_ARGUMENT_JSON: 100014, // Request message should json
    ERROR_ARGUMENT_VALUE_WRONG: 100015, // Argument [{}] value wrong
    ERROR_PARAM_RANGE_WRONG: 100016, // Param [{}] range wrong.
    ERROR_MODIFY_FAILED: 100017, // Modify failed.
    ERROR_FILE_PATH_SOFT_LINK: 100022, // File path is softlink.
    ERROR_FILE_PATH_CHECK: 100023, // File path check.
    ERROR_INVALID_PARAMS: 100024, // Parameter is invalid.
    ERROR_OPERATE_IS_BUSY: 100028, // The operation is busy.
    ERROR_CMD_EXEC_ERROR: 100019, // Command execution error.
    ERROR_FILE_OPEN: 110103, // File open error.
    ERROR_OPERATION_ILLEGAL: 100027, // The operator is illegal.
    ERROR_ARGUMENT_NOT_EXIST: 100012, // Argument [{}] not exist
    ERROR_SYS_UPGRADE: 100020, // The System upgrade.
    ERROR_RESTORE_DEFAULTS: 100021, // Restore defaults failed.
    ERROR_DIR_SPACE_LOW: 100018, // The [{}] space directory is low.
    ERROR_RECOVER_OS: 100031, // Recover mini os is running.
  },
  SESSION: {
    ERROR_ROOT_OPERATE_LOCKED: 100026,
    ERROR_SESSION_NOT_FOUND: 110201, // Session not found.
    ERROR_TIME_LIMIT_EXCEEDED: 110202, // 会话超时
    ERROR_DB_OPERATION_FAILED: 110203, // Database operation failed.
    ERROR_PASSWORD_EXPIRE: 110204, // password has expired.
    ERROR_REQUEST_IP_ADDR: 110205, // ip address error.
    ERROR_USER_LOCK_STATE: 110206,
    ERROR_USER_NOT_MATCH_PASSWORD: 110207,
    ERROR_SECURITY_CFG_NOT_MEET: 110208,
    ERROR_PARAM_RANGE: 110209, // Param [{}] wrong range.
    ERROR_USER_NOT_FOUND: 110210, // 用户不存在
    ERROR_DUPLICATE_PASSWORD: 110211, // 与前五次密码一致
    ERROR_USERNAME_OS_MODIFY: 110212, // Modify os username failed.
    ERROR_PASSWORD_OS_MODIFY: 110213, // Modify os password failed.
    ERROR_PASSWORD_REVERSAL: 110214,
    ERROR_PASSWORD_TWO_INCONSISTENT: 110215, // Two passwords are inconsistent.
    ERROR_USER_AUTH_FAILED: 110216,
    ERROR_USERNAME_USED: 110217,
    ERROR_OS_USER_NOT_EXIST: 110218,
    ERROR_CHANGE_PASSWORD: 110219,
    ERROR_PASSWORD_IS_DEFAULT: 110221,
    ERROR_PASSWORD_IS_WEAK: 110222, // 新密码强度太弱
    ERROR_EDGE_CONFIG_NOT_FOUND: 110223, // Edge_config not found.
    ERROR_RESTORE_INIT_WORD_WRONG: 100025, // The root account has not changed the default password
    ERROR_INTERNAL: 110224,
    ERROR_SERVICE_IS_STARTING: 100029, // The service is starting.
    ERROR_SERVICE_STARTUP_FAILED: 100030, // The service startup failed.
  },
  CERT: {
    SUCCESS_BUT_NOT_SAFE: 206,
    ERROR_CERTIFICATE_ABOUT_TO_EXPIRE: 110301, // The certificate is about to expire.
    ERROR_CERTIFICATE_INIT_FAILED: 110302, // Importing a custom certificate init failed
    ERROR_CERTIFICATE_TOO_LARGE: 110303, // Certificate is too large
    ERROR_CERTIFICATE_OR_PASSWORD_VERIFY_FAILED: 110304, // Certificate or password verification failed
    ERROR_CERTIFICATE_IS_NOT_VALID: 110305, // Certificate is not valid
    ERROR_IMPORT_CERTIFICATE_INTERNAL_ERROR: 110306, // Importing a custom certificate failed
    ERROR_IMPORT_CERTIFICATE_LEGALITY_RISKY: 110307, // Importing a custom certificate success but legality is risky
    ERROR_CERTIFICATE_HAS_EXPIRED: 110308, // The certificate has expired
    ERROR_CERTIFICATE_DOES_NOT_MATCH: 110309, // The certificate does not match the private key
    ERROR_CERTIFICATE_WAIT_TIME: 110310, // Please wait 2 hours and download it again
    ERROR_CERTIFICATE_IS_INVALID: 110311, // The certificate is invalid
  },
  UPLOAD: {
    ERROR_EMPTY_FILE: 110104, // 上传空文件
    ERROR_FILE_NOT_EXIST: 110105, // 文件不存在
    ERROR_FILE_SIZE_TOO_LARGE: 110106, // 文件过大
    ERROR_SPACE_INSUFFICIENT: 100005, // 空间不足
    ERROR_UPLOAD_FLAG: 110107, // 10分钟内重复上传文件
  },
  NFS: {
    ERROR_MOUNT_PATH_EXISTED: 110501, // Mount path already exists.
    ERROR_MOUNT_PATH_NOT_IN_WHITE_LIST: 110502, // Mount path not in white list.
    ERROR_MOUNT_PATH_SUBDIRECTORY_RELATIONSHIP: 110503, // There is a subdirectory relationship between Mount Path and Mounted path.
    ERROR_CONFIGURATION_EXCEEDS_LIMIT: 110504, // NFS configuration exceeds limit.
    ERROR_MOUNT_FAILED: 110505, // Operate NFS failed.
    ERROR_MOUNT_TIME_OUT: 110506, // Operate time out.
    ERROR_UNMOUNT_PATH_NOT_EXISTED: 110507, // Unmount path does not exist.
    ERROR_FILESYSTEM_ERROR: 110508, // FileSystem is error
    ERROR_GET_LOCAL_INFO_FAILED: 110509, // NFS local mount path to get mount local info failed
  },
  NIC: {
    ERROR_TAG_CHECK_FAILED: 110401, // Check tag info failed.
    ERROR_DEAL_ETHERNET_INTERFACE_FAILED: 110402, // Deal with EthernetInterface failed.
    ERROR_CONFIG_FILE_NOT_EXIST: 110403, // Config file not exists.
    ERROR_TMP_NETCFG_FILE_INVALID: 110404, // Netcfg file check failed.
    ERROR_LOAD_NET_CONFIG_YAML_FAILED: 110405, // Load net config yaml failed.
    ERROR_MODIFY_NETPLAN_FAILED: 110406, // Modify netplan failed.
    ERROR_WRITE_NET_CONFIG_YAML_FAILED: 110407, // Write net config yaml failed.
    ERROR_NETWORK_APPLY_FAILED: 110408, // Network apply failed, rollback.
    ERROR_PING_TEST_FAILED: 110409, // Ping test failed.
    ERROR_WRITE_TAG_FAILED: 110410, // Write tag ini failed.
    ERROR_WRITE_TAG_LOOK_AFTER_OLD_BER4NIC_FAILED: 110411, // Write tag look after old ber4Nic failed.
    ERROR_SER_NGINX_FAILED: 110412, // Set nginx config failed.
    ERROR_CREATE_MULTIPLE_IP_FILE_FAILED: 110413, // Create multiple ip file errors
    ERROR_WRITE_GATEWAY_FAILED: 110414, // Write gateway failed
    ERROR_RESTART_NETWORK_FAILED: 110415, // Restart network failed
    ERROR_SOURCE_NOT_EXIST: 110416, // System source not exist
  },
  NET_MANAGER: {
    PARTIAL_SUCCESS: 206,
  },
  DRIVE: {
    ERROR_PARTITION_NAME: 110601, // Partition_name is wrong.
    ERROR_PARTITION_IS_SYSTEM_PARTITION: 110602, // Partition is in system partitions.
    ERROR_CONVERSE_SITE_PARTITION_FAILED: 110603, // Converse site partition failed.
    ERROR_UNMOUNT_FAILED: 110604, // Unmount partition failed.
    ERROR_DELETE_PARTITION_FAILED: 110605, // Delete partition failed.
  },
  LOG: {
    ERROR_LOG_COLLECT: 110001, // Collect log failed.
  },
  PARTITION: {
    ERROR_PARTITION_NAME: 110601, // Partition_name is wrong.
    ERROR_PARTITION_IS_SYSTEM_PARTITION: 110602, // Partition is in system partitions.
    ERROR_CONVERSE_SITE_PARTITION_FAILED: 110603, // Converse site partition failed.
    ERROR_UNMOUNT_FAILED: 110604, // Unmount partition failed.
    ERROR_DELETE_PARTITION_FAILED: 110605, // Delete partition failed.
    ERROR_MOUNT_PATH_NOT_IN_WHITE_LIST: 110606, // Mount path not in white list.
    ERROR_DISK_IS_MOUNTED: 110607, // Disk is mounted.
    ERROR_MOUNT_PATH_INVALID: 110608, // Mount path is invalid.
    ERROR_PATH_MOUNTED: 110609, // Path has been mounted
    ERROR_MOUNT_FAILED: 110610, // Mount partition failed.
    ERROR_SPACE_NOT_ENOUGH: 110611, // The disk not enough space.
    ERROR_SYSTEM_BUSY: 110612, // System busy, will not make a filesystem here!
    ERROR_CREATE_PARTITION_FAILED: 110613, // Create partition failed.
    ERROR_DISK_IS_NOT_MOUNTED: 110614, // The disk is not mounted.
    ERROR_MOUNT_IS_EXISTED: 110615, // Mount path is existed.
    ERROR_DISK_TYPE: 110616, // Partition type invalid
    ERROR_DRIVE_LETTER_NOT_EXISTED: 110617, // Drive letter does not exist.
    ERROR_PARTITION_NUM: 110618, // The number of partitions exceeds the limit.
    ERROR_DOCKER_UNMOUNT_FAILED: 110619, // Containers have not been deleted or docker path is occupied.
  },
  REGISTRATION: {
    ERROR_IP_LOCKED: 110225, // ip被锁定 请5分钟后重试
    ERROR_CERTIFICATE_FORMAT: 110312, // The certificate is not X.509 v3 standard format
    ERROR_PUBLIC_KEY: 110313, // The certificate pubkey type is not RSA or EC
    ERROR_LESS_RSA_PUBLIC_KEY: 110314, // The RSA pubkey length is less than 3072
    ERROR_LESS_EC_PUBLIC_KEY: 110315, // The EC pubkey length is less than 265
    ERROR_SIGNATURE_ALGORITHM: 110316,
    // The signature algorithm is not sha256WithRSAEncryption、
    // sha384WithRSAEncryption、sha512WithRSAEncryption or ecdsa-with-SHA256
    ERROR_CERTIFICATE_IS_NOT_CA: 110317, // The cert is not CA
    ERROR_CERTIFICATE_KEY_SIGN: 110318, // Cannot be used for verifying signatures on public key certificates
    ERROR_CERTIFICATE_CHAIN_NUMS_MAX: 110319, // The number of CA certificate chains is greater than 10.
    ERROR_CERTIFICATE_CA_SIGNATURE_INVALID: 110320, // The CA cert signature is invalid
    ERROR_DOCKER_IS_NOT_MOUNTED: 110226, // Docker root dir is not mounted. Please mount it first.
  },
}