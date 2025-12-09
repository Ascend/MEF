# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json
from typing import NamedTuple, Optional, Tuple

import pytest
from mock.mock import Mock
from pytest_mock import MockerFixture

from common.file_utils import FileCheck, FileCreate, FileUtils
from common.utils.result_base import Result
from lib_restful_adapter import LibRESTfulAdapter
from upgrade_service.errors import UpgradeError, TimeOutError, BaseError
from upgrade_service.models import UpgradeInfo
from upgrade_service.upgrade_entry import UpgradeService


class UpgradeModel(NamedTuple):
    adapter: Optional[dict] = None
    error: str = "Get firmware upgrade progress error, catch TypeError error."


class DownloadImgByHttpsModel(NamedTuple):
    dir_exists: bool = True
    create_dir: bool = True
    error: str = ""
    download: Tuple = (-1, "")


class HandlerModel(NamedTuple):
    locked: bool = False
    expect: bool = False
    request: dict = {}
    sys_busy: dict = {}
    upgrade: Optional[Exception] = None


class TestUpgradeEntry:
    use_cases = {
        "test_get_upgrade_process_err": {
            "adapter get none": UpgradeModel(),
            "adapter get empty dict": UpgradeModel({}),
            "adapter not ok": UpgradeModel({"status": 400, "message": "not ok"}, "not ok"),
            "null message": UpgradeModel({"status": 200, "message": {"Messages": None}}),
            "upgrade failed": UpgradeModel(
                adapter={
                    "status": 200,
                    "message": {
                        "Messages": {"upgradeState": "upgrade failed"},
                        "TaskState": "Failed",
                        "PercentComplete": 0,
                        "Version": ""
                    },
                },
                error="upgrade failed",
            ),
        },
        "test_firmware_upgrade": {
            "adapter get none": UpgradeModel(error="Firmware to upgrade error, catch TypeError error."),
            "adapter not ok": UpgradeModel(error="not ok", adapter={"status": 400, "message": "not ok"}),
            "upgrade ok": UpgradeModel(error="", adapter={"status": 200, "message": ""}),
        },
        "test_download_software_image_by_https": {
            "dir create failed": DownloadImgByHttpsModel(False, False, "Create download dir failed."),
            "download failed": DownloadImgByHttpsModel(create_dir=True, error="Use https download software failed."),
            "download success": DownloadImgByHttpsModel(create_dir=True, download=(0, "")),
        },
        "test_handler": {
            "locked": HandlerModel(locked=True),
            "null request": HandlerModel(),
        }
    }

    @staticmethod
    def test_get_upgrade_process_err(mocker: MockerFixture, model: UpgradeModel):
        mocker.patch.object(UpgradeService, "adapter", return_value=model.adapter)
        with pytest.raises(UpgradeError, match=model.error):
            UpgradeService("")._get_upgrade_process()

    @staticmethod
    def test_firmware_upgrade(mocker: MockerFixture, model: UpgradeModel):
        mocker.patch.object(UpgradeService, "adapter", return_value=model.adapter)
        upgrade_service = UpgradeService("")
        upgrade_service.payload = Mock()
        upgrade_service.payload.https_server.upgrade_request = {}
        if not model.error:
            upgrade_service._firmware_upgrade()
            return

        with pytest.raises(UpgradeError, match=model.error):
            upgrade_service._firmware_upgrade()

    @staticmethod
    def test_get_upgrade_process(mocker: MockerFixture):
        mocker.patch.object(UpgradeService, "adapter", return_value={
            "status": 200,
            "message": {
                "Messages": {"upgradeState": "Success"},
                "TaskState": "Success",
                "PercentComplete": 100,
                "Version": "123"
            },
        })
        assert UpgradeService("")._get_upgrade_process() == UpgradeInfo("Success", 100, "123", "Success")

    @staticmethod
    def test_await_upgrade_finish(mocker: MockerFixture):
        mocker.patch.object(UpgradeService, "_get_upgrade_process",
                            return_value=UpgradeInfo("Success", 100, "123", "Success"))
        upgrade_service = UpgradeService("")
        upgrade_service.reporter = Mock()
        upgrade_service._await_upgrade_finish()

    @staticmethod
    def test_await_upgrade_timeout(mocker: MockerFixture):
        mocker.patch.object(UpgradeService, "_get_upgrade_process",
                            return_value=UpgradeInfo("Upgrading", 90, "123", "Upgrading"))
        mocker.patch("time.sleep")
        upgrade_service = UpgradeService("")
        upgrade_service.REPORT_MAX_TIMES = 1
        upgrade_service.reporter = Mock()
        with pytest.raises(TimeOutError, match="firmware upgrade timeout"):
            upgrade_service._await_upgrade_finish()

    @staticmethod
    def test_download_software_image_by_https(mocker: MockerFixture, model: DownloadImgByHttpsModel):
        mocker.patch.object(FileCheck, "is_exists", return_value=model.dir_exists)
        mocker.patch.object(FileCreate, "create_dir", return_value=model.create_dir)
        mocker.patch("upgrade_service.upgrade_entry.https_download_file", return_value=model.download)
        upgrade_service = UpgradeService("")
        upgrade_service.payload = Mock()
        if not model.error:
            upgrade_service._download_software_image_by_https()
            return

        with pytest.raises(BaseError, match=model.error):
            upgrade_service._download_software_image_by_https()

    @staticmethod
    def test_download_software_join(mocker: MockerFixture):
        upgrade_service = UpgradeService("")
        upgrade_service.reporter = Mock()
        mocker.patch.object(UpgradeService, "_download_software_image_by_https")
        upgrade_service._download_software()
        assert not upgrade_service.reporter.downloading

    @staticmethod
    def test_handler(mocker: MockerFixture, model: HandlerModel):
        mocker.patch.object(UpgradeService, "upgrade_lock").locked.return_value = model.locked
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.sys_busy)
        if isinstance(model.upgrade, Exception):
            mocker.patch.object(UpgradeService, "_execute_upgrade", side_effect=model.upgrade)
        else:
            mocker.patch.object(UpgradeService, "_execute_upgrade")
        mocker.patch.object(FileUtils, "delete_file_or_link")
        ret = UpgradeService(json.dumps(model.request)).handler()
        assert isinstance(ret, Result), bool(ret) == model.expect
