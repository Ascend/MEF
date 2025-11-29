# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import json
import os.path
from dataclasses import dataclass, field, asdict
from typing import Optional, Dict

from common.checkers.fd_param_checker import UpdateServiceFirmwareChecker
from common.constants.upgrade_constants import FDResultMsg
from common.constants.upload_constants import UploadConstants
from common.schema import BaseModel
from upgrade_service.errors import ExternalParmaError


@dataclass
class HttpsServer(BaseModel):
    """固件升级payload嵌套的https_server数据结构"""
    image: str
    password: str
    user_name: str
    url: str = field(init=False)
    filename: str = field(init=False, default=UploadConstants.DOWNLOAD_IMAGE_NAME)
    download_dir: str = field(init=False, default=UploadConstants.DRIVER_DOWNLOAD_PATH)
    local_filename: str = field(init=False)
    upgrade_request: Dict[str, str] = field(init=False)

    def __post_init__(self):
        self.url = self.image.split(" ")[1] if " " in self.image else ""
        self.local_filename = os.path.join(self.download_dir, self.filename)
        self.upgrade_request = {"TransferProtocol": "https", "ImageURI": self.local_filename}


@dataclass
class UpgradePayload(BaseModel):
    """固件升级payload嵌套的二级资源数据"""
    operator: str
    type: str
    install_method: str
    enable_method: str
    name: str
    https_server: HttpsServer
    check_type: str
    check_code: str

    @classmethod
    def from_payload(cls, payload):
        ret = UpdateServiceFirmwareChecker().check(payload)
        if not ret.success:
            raise ExternalParmaError(ret.reason)

        # 校验成功，member_list必然为[{}]
        content: dict = payload.get("member_list")[0]
        return cls.from_dict(content)


@dataclass
class UpgradeInfo(BaseModel):
    """Upgrade_New GET返回的封装数据"""
    state: str = field(metadata={"alias": "TaskState"})
    percentage: int = field(metadata={"alias": "PercentComplete"})
    version: str = field(metadata={"alias": "Version"})
    msg: str

    @classmethod
    def from_dict(cls, data: dict):
        message: Optional[dict] = data.get("Messages")
        msg = message.get("upgradeState") if message else ""
        data["msg"] = msg
        return super().from_dict(data)


@dataclass
class Report:
    """上报消息结构"""
    operator: str = field(default="install")
    name: str = field(default="firmware")
    version: str = field(default="NA")
    percentage: str = field(default="0%")
    result: str = field(default="processing")
    reason: str = field(default="")

    def dumps(self):
        return json.dumps({"members": [asdict(self)]})

    def to_dict(self):
        return {"members": [asdict(self)]}

    def update_by_upgrade_info(self, info: UpgradeInfo):
        # 运行到此，说明下载已完成，进度至少为35%
        percentage = max(35, info.percentage)
        self.result = FDResultMsg.PROCESSING if percentage != 100 else FDResultMsg.SUCCESS
        self.percentage = f"{percentage}%"
        self.version = info.version


