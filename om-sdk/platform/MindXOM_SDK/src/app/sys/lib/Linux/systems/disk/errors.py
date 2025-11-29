# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from common.constants.base_constants import CommonConstants
from common.constants.error_codes import ErrorCode, PartitionErrorCodes


class DevError(Exception):
    code: int = CommonConstants.ERR_CODE_400
    error: ErrorCode = PartitionErrorCodes.ERROR_OPERATE_DEVICE

    @classmethod
    def out(cls):
        """封装成返回结果需要的结构"""
        return [cls.code, [cls.error.code, cls.error.messageKey]]


class DevNotFound(DevError):
    pass


class ConvertGptError(DevError):
    error = PartitionErrorCodes.ERROR_CONVERSE_SITE_PARTITION_FAILED


class CreatePartFailed(DevError):
    error = PartitionErrorCodes.ERROR_CREATE_PARTITION_FAILED


class DelPartFailed(DevError):
    error = PartitionErrorCodes.ERROR_DELETE_PARTITION_FAILED


class SpaceNotEnough(DevError):
    error = PartitionErrorCodes.ERROR_SPACE_NOT_ENOUGH


class PartMounted(DevError):
    error = PartitionErrorCodes.ERROR_DISK_IS_MOUNTED


class PartNotMount(DevError):
    error = PartitionErrorCodes.ERROR_DISK_IS_NOT_MOUNTED


class MountFailed(DevError):
    error = PartitionErrorCodes.ERROR_MOUNT_FAILED


class MountPathInvalid(DevError):
    error = PartitionErrorCodes.ERROR_MOUNT_PATH_INVALID


class MountPathExisted(DevError):
    error = PartitionErrorCodes.ERROR_MOUNT_IS_EXISTED


class PathNotInWhite(DevError):
    error = PartitionErrorCodes.ERROR_MOUNT_PATH_NOT_IN_WHITE_LIST


class UmountDockerFailed(DevError):
    error = PartitionErrorCodes.ERROR_DOCKER_UNMOUNT_FAILED


class UmountFailed(DevError):
    error = PartitionErrorCodes.ERROR_UNMOUNT_FAILED
