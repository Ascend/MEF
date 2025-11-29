# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from common.constants.error_codes import NfsServiceErrorCodes, ErrorCode
from common.common_methods import CommonMethods


class MountError(Exception):
    code: int = CommonMethods.ERROR
    error: ErrorCode

    @classmethod
    def out(cls):
        """封装成返回结果需要的结构"""
        return [cls.code, [cls.error.code, cls.error.messageKey]]


class OperateBusy(MountError):
    error = NfsServiceErrorCodes.OPERATE_IS_BUSY


class ParmaError(MountError):
    error = NfsServiceErrorCodes.ERROR_PARAMETER_INVALID


class MountPathExisted(MountError):
    error = NfsServiceErrorCodes.ERROR_MOUNT_PATH_EXISTED


class MountPathInvalid(MountError):
    error = NfsServiceErrorCodes.ERROR_MOUNT_PATH_NOT_IN_WHITE_LIST


class SubdirRelation(MountError):
    error = NfsServiceErrorCodes.ERROR_MOUNT_PATH_SUBDIRECTORY_RELATIONSHIP


class CfgExceedsLimit(MountError):
    error = NfsServiceErrorCodes.ERROR_CONFIGURATION_EXCEEDS_LIMIT


class OperateFailed(MountError):
    error = NfsServiceErrorCodes.ERROR_MOUNT_FAILED


class TimeOut(MountError):
    error = NfsServiceErrorCodes.ERROR_MOUNT_TIME_OUT


class UmountPathNotExists(MountError):
    error = NfsServiceErrorCodes.ERROR_UNMOUNT_PATH_NOT_EXISTED


class FilesystemWrong(MountError):
    error = NfsServiceErrorCodes.ERROR_FILESYSTEM_ERROR


class LocalInfoError(MountError):
    error = NfsServiceErrorCodes.ERROR_GET_LOCAL_INFO_FAILED
