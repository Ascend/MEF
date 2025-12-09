# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
import re
from hashlib import sha256
from pathlib import Path
from typing import Iterable

from sqlalchemy import Column
from sqlalchemy import String

from common.constants.base_constants import CommonConstants
from common.db.base_models import Base
from common.db.base_models import RebuildMixin
from common.db.base_models import SaveDefaultsMixin
from common.file_utils import FileCheck
from common.log.logger import run_log
from devm import config
from common.kmc_lib.kmc import Kmc


class CertInfo:
    CERTS_DIR = "software/ibma/cert"

    def __init__(self, work_dir: str):
        certs_dir = os.path.join(work_dir, self.CERTS_DIR)
        self.primary_keystore_name = os.path.join(certs_dir, "om_cert.keystore")
        self.backup_keystore_name = os.path.join(certs_dir, "om_cert_backup.keystore")
        self.alg_json_name = os.path.join(certs_dir, "om_alg.json")


class DevmConfig(Base, RebuildMixin, SaveDefaultsMixin):
    __tablename__ = "devm_config"

    cfg_sha256 = Column(String(256), comment="配置文件的sha256加密值")
    filename = Column(String(256), unique=True, primary_key=True, comment="配置文件的文件名")

    @classmethod
    def default_instance_generator(cls) -> Iterable["DevmConfig"]:
        yield from DevmConfigMgr(CommonConstants.OM_UPGRADE_DIR_PATH).config_generator()


class ConfigFile:
    MAX_SIZE = 1 * 1024 * 1024

    def __init__(self, cfg: Path):
        self.cfg = cfg

    def read_bytes(self) -> bytes:
        if not self._is_valid(self.cfg):
            run_log.error("%s invalid, ignored.", self.cfg.name)
            return b""

        # 获取配置文件内容
        try:
            return self.cfg.read_bytes()
        except Exception as err:
            msg = f"{self.cfg.name} invalid. catch {err.__class__.__class__} while reading the content."
            run_log.error(msg)
            return b""

    def _is_valid(self, file: Path):
        return FileCheck.check_path_is_exist_and_valid(file.as_posix()) and file.stat().st_size < self.MAX_SIZE


class DevmConfigMgr:
    CFG_DIR: str = "software/ibma/config/devm_configs"
    EXT: str = ".json"
    FILE_NAME_PATTERN = r"^[a-zA-Z0-9_.-]{1,255}$"

    def __init__(self, work_dir: str):
        self.cfg_dir = os.path.join(work_dir, self.CFG_DIR)
        cert = CertInfo(work_dir)
        self.kmc = Kmc(cert.primary_keystore_name, cert.backup_keystore_name, cert.alg_json_name)

    def config_generator(self) -> Iterable[DevmConfig]:
        """配置生成器，遇到异常则忽略，确保om-init运行成功。"""
        for index, file in enumerate(Path(self.cfg_dir).glob(f"*{self.EXT}")):
            if index > config.MAX_CFG_FILES:
                break

            if not self._check_filename_valid(file.name):
                run_log.error("%s invalid, ignored.", file.name)
                continue

            cfg_file = ConfigFile(file)
            cfg_bytes = cfg_file.read_bytes()
            if not cfg_bytes:
                continue

            try:
                yield DevmConfig(
                    filename=file.name,
                    cfg_sha256=self.kmc.encrypt(sha256(cfg_bytes).hexdigest())
                )
            except Exception as err:
                run_log.error("encrypt %s error, catch %s", file.name, err.__class__.__name__)

    def check_config(self, cfg_file) -> bool:
        file_name = os.path.basename(cfg_file)
        if not self._check_filename_valid(file_name):
            run_log.error("%s file name invalid.", file_name)
            raise ValueError(f"{file_name} file name invalid")

        cfg_bytes = ConfigFile(Path(cfg_file)).read_bytes()
        if not cfg_bytes:
            run_log.error("%s invalid, ignored.", file_name)
            raise ValueError(f"{file_name} file content invalid")

        return sha256(cfg_bytes).hexdigest() == self._get_cfg_info(file_name)

    def _check_filename_valid(self, file_name) -> bool:
        find = re.fullmatch(self.FILE_NAME_PATTERN, file_name)
        return find and find.group(0) == file_name

    def _get_cfg_info(self, file_name) -> str:
        from monitor_db.session import session_maker
        with session_maker() as session:
            cfg_info: DevmConfig = session.query(DevmConfig).filter_by(filename=file_name).first()
            if not cfg_info:
                raise ValueError(f"{file_name} config info not in db")
            cfg_sha256 = self.kmc.decrypt(cfg_info.cfg_sha256)
            return cfg_sha256
