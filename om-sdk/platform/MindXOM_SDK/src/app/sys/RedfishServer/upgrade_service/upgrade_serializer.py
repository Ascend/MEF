# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from system_service.systems_serializer import Serializer


class GetUpgradeServiceResourceSerializer(Serializer):
    """
    功能描述：查询升级资源集合
    """
    service = Serializer.root.upgrade_service_common.get_resource_collection


class GetUpgradeServiceActionSerializer(Serializer):
    """
    功能描述：查询升级资源集合
    """
    service = Serializer.root.upgrade_service_common.actions
