# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os
from typing import Tuple, Iterable, Any, Dict, NoReturn

from sqlalchemy.orm import scoped_session

from common.constants.base_constants import CommonConstants
from common.db.base_models import Base
from common.file_utils import FileCheck
from common.kmc_lib.update_kmc_key import Adapter
from common.log.logger import run_log


class TableAdapter(Adapter):
    """数据库表的Kmc密钥更新基类"""

    model: Base
    session: scoped_session
    filter_by: str
    cols: Tuple[str]

    def decrypted_generator(self) -> Iterable[Tuple[Any, Dict[str, Any]]]:
        with self.session() as session:
            for obj in session.query(self.model).all():
                data = {col: self.kmc.decrypt(getattr(obj, col)) for col in self.cols if getattr(obj, col)}
                if not data:
                    continue
                yield getattr(obj, self.filter_by), data

    def encrypt(self) -> NoReturn:
        with self.session() as session:
            for filter_value, data in self.decrypted.items():
                values = {col: self.kmc.encrypt(value) for col, value in data.items()}
                session.query(self.model).filter_by(**{self.filter_by: filter_value}).update(values)


class FilesAdapter(Adapter):
    FLAGS = os.O_WRONLY | os.O_TRUNC | os.O_CREAT
    decrypted: Dict[str, str]

    def encrypted_files(self) -> Iterable[str]:
        pass

    def decrypted_generator(self) -> Iterable[Tuple[str, str]]:
        for filename in self.encrypted_files():
            ret = FileCheck.check_path_is_exist_and_valid(filename)
            if not ret:
                run_log.error("%s invalid, because of %s", os.path.basename(filename), ret.error)
                continue

            if os.path.getsize(filename) > CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES:
                run_log.error("%s out of max size.", os.path.basename(filename))
                continue

            with open(filename) as file_obj:
                yield filename, self.kmc.decrypt(file_obj.read())

    def encrypt(self) -> NoReturn:
        for filename, content in self.decrypted.items():
            with os.fdopen(os.open(filename, self.FLAGS, 0o600), "w") as file_obj:
                file_obj.write(self.kmc.encrypt(content))
