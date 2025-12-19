# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import os
from contextlib import contextmanager
from typing import Dict, Type, Set

from sqlalchemy import create_engine, Table, inspect
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.orm import scoped_session, sessionmaker

from common.db.base_models import Base, ClearOnStartMixin
from common.exception.biz_exception import BizException
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.exception_utils import OperateBaseError
from common.utils.singleton import Singleton


def session_maker(session: scoped_session):
    try:
        yield session
        session.commit()
    except BizException as err:
        session.rollback()
        raise err
    except OperateBaseError as err:
        session.rollback()
        raise err
    # 防止发生数据库异常时，返回敏感信息
    except SQLAlchemyError as err:
        session.rollback()
        raise Exception("Operate database error") from err
    except Exception as err:
        session.rollback()
        raise Exception("Operate database occur normal error") from err
    finally:
        session.close()


class DataBase(Singleton):
    models: Dict[str, Type[Base]] = {}

    def __init__(self, db_path: str):
        self.db_path = db_path
        url = f"sqlite:////{self.db_path}?check_same_thread=False"
        self._validate_db_path()
        self.engine = create_engine(url, echo=False, encoding="utf-8")
        self._scoped_session = scoped_session(sessionmaker(bind=self.engine))

    @staticmethod
    def get_table_column_names(table: Table) -> Set[str]:
        """获取数据表在数据库中的列名"""
        return set(c.name for c in inspect(table).columns)

    @classmethod
    def register_models(cls, *models: Type[Base]):
        """
        将项目中用到的表显示地注册到数据库处理类中，确保数据表创建：
            例如：Database.registers_table(Person)
        """
        for model in models:
            cls.models[model.__tablename__] = model

    @contextmanager
    def session_maker(self):
        yield from session_maker(self._scoped_session)

    @contextmanager
    def simple_session_maker(self, session: scoped_session = None):
        """嵌套的session"""
        if session is not None:
            yield session
        else:
            yield from session_maker(self._scoped_session)

    def clear_data_on_start(self):
        """启动时，清理历史数据的入口函数"""
        model_generator = (m for m in self.models.values() if issubclass(m, ClearOnStartMixin))
        with self.session_maker() as session:
            for model in model_generator:
                model.clear_table(session)

    def _validate_db_path(self):
        if os.path.exists(self.db_path):
            run_log.info("%s has already exist!", os.path.basename(self.db_path))
        else:
            run_log.info("%s does not exist, create new!", os.path.basename(self.db_path))

        res = FileCheck.check_input_path_valid(self.db_path)
        if not res:
            run_log.error("%s path invalid : %s", self.db_path, res.error)
            raise ValueError(f"{self.db_path} path invalid : {res.error}")
