# -*- coding: UTF-8 -*-
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
from argparse import ArgumentParser
from typing import Dict, Set, Iterable, Tuple, NoReturn

from sqlalchemy import inspect, Table, Column, String
from sqlalchemy.ext.automap import automap_base
from sqlalchemy.orm import scoped_session

from common.constants.base_constants import MigrateOperate
from common.db.base_models import Base, Structure, SaveDefaultsMixin, RebuildMixin
from common.db.database import DataBase
from common.file_utils import FileCreate
from common.file_utils import FilePermission
from common.log.logger import run_log


class Migrate(DataBase):
    auto_base = None
    reflect_tables: Dict[str, Table] = {}

    def __init__(self, db_path: str, init_structures: Dict[str, Iterable[str]]):
        super().__init__(db_path)
        self.init_structures = init_structures

    @property
    def new_tables(self) -> Set[str]:
        """新增的表名集合"""
        return set(self.models) - set(self.reflect_tables)

    @property
    def existed_tables(self) -> Set[str]:
        """数据库中已经存在的表名集合"""
        return set(self.models).intersection(self.reflect_tables)

    @property
    def useless_tables(self) -> Set[str]:
        """数据库中无用的表名集合"""
        return set(self.reflect_tables) - {Structure.__tablename__, *self.models}

    @classmethod
    def execute(cls, db_file: str, init_structures: Dict[str, Iterable[str]]) -> int:
        try:
            cls(db_file, init_structures).make_migrate()
        except Exception as error:
            run_log.error("migrate database failed, catch %s", error.__class__.__name__)
            return 1

        run_log.info("migrate database finish.")
        return 0

    @classmethod
    def execute_on_install(cls, db_file: str, init_structures: Dict[str, Iterable[str]]) -> int:
        # 创建数据库目录
        if not FileCreate.create_dir(os.path.dirname(db_file), mode=0o700):
            run_log.error("create database dir failed.")
            return 1

        try:
            cls(db_file, init_structures).init_database()
        except Exception as err:
            run_log.error("migrate db failed, catch %s", err.__class__.__name__)
            return 1
        run_log.info("migrate database finish.")
        return 0

    def init_database(self):
        run_log.info("start init database.")
        self._create_structure_table()
        self._create_all_tables()
        self._save_default_data()
        FilePermission.set_path_permission(self.db_path, 0o600)
        run_log.info("init database finish.")

    def make_migrate(self):
        self._reflect_database()
        run_log.info("new tables: %s", self.new_tables)
        run_log.info("existed tables: %s", self.existed_tables)
        run_log.info("useless tables: %s", self.useless_tables)
        self._create_structure_table()
        self._drop_tables_before_migrate()
        self._create_all_tables()
        self._save_default_data()
        self._migrate()
        FilePermission.set_path_permission(self.db_path, 0o600)
        run_log.info("migrate finish.")

    def _reflect_database(self):
        self.auto_base = automap_base()
        self.auto_base.prepare(self.engine, reflect=True)
        self.reflect_tables: Dict[str, Table] = {table.name: table for table in self.auto_base.metadata.sorted_tables}

    def _save_default_data(self):
        """存默认数据到数据库"""
        with self.session_maker() as session:
            for table_name, model in self.models.items():
                if not issubclass(model, SaveDefaultsMixin):
                    continue
                # 非重建且已存在的表，不需保存默认数据
                if not issubclass(model, RebuildMixin) and table_name in self.reflect_tables:
                    continue
                model.save_defaults(session)

    def _create_structure_table(self):
        """创建结构维护表"""
        Structure.__table__.create(self.engine, True)

    def _new_table_generator(self) -> Iterable[Tuple[Table, Structure]]:
        """生成需要新建的表"""
        for table_name, model in self.models.items():
            # 重建表或数据库已存在的表不冗余字段
            if issubclass(model, RebuildMixin) or table_name in self.reflect_tables:
                continue
            # 需要平滑的表，不在INIT_COLUMNS则被忽略
            if table_name not in self.init_structures:
                continue
            table: Table = model.__table__
            # 当前模型用到的数据表列，需要在append_column之前获取
            columns = self.get_table_column_names(table)
            # 一个表的总的字段，版本间要求不能变化 = 表模型首次用到的字段 + 冗余字段
            init_columns = set(self.init_structures[table_name])
            # 每个表预留三个字段：reserved1、reserved2、reserved3；有则忽略；
            reserved = set(f"reserved{number}" for number in range(1, 4))
            # 需要扩展的字段 = 初始表字段 + 预留字段 - 当前模型使用的字段
            for name in init_columns.union(reserved) - columns:
                # 预留字段统一采用String，使用预留字段时易于兼容
                table.append_column(Column(name, String, nullable=True, comment="预留字段"))

            yield table, Structure(table_name=table_name, columns=",".join(columns))

    def _create_all_tables(self):
        """创建数据表，并扩展冗余字段"""
        # 重建表先删后建，无需记录structure
        tables = [model.__table__ for model in self.models.values() if issubclass(model, RebuildMixin)]
        with self.session_maker() as session:
            structures = []
            for table, structure in self._new_table_generator():
                tables.append(table)
                structures.append(structure)
            # 将表结构信息持久到数据Structure中
            session.bulk_save_objects(structures)
        # 创建表；如果处理结构失败则会上抛异常，不会新建表
        Base.metadata.create_all(self.engine, tables)

    def _drop_tables_before_migrate(self):
        """迁移前先删掉需要重建或不用的表"""
        tables = []
        table_names = []
        for table_name, table in self.reflect_tables.items():
            if table_name == Structure.__tablename__:
                continue
            if table_name not in self.models or issubclass(self.models[table_name], RebuildMixin):
                tables.append(table)
                table_names.append(table_name)
        with self.session_maker() as session:
            session.query(Structure).filter(Structure.table_name.in_(table_names)).delete()
        self.auto_base.metadata.drop_all(bind=self.engine, tables=tables)

    def _migrate_tables(self, session: scoped_session) -> NoReturn:
        """生成需要迁移的表"""
        for table_name, model in self.models.items():
            # 没在反射表中或者是重建表的无需迁移
            if table_name not in self.reflect_tables or issubclass(model, RebuildMixin):
                continue
            reflect_table = self.reflect_tables[table_name]
            table: Table = model.__table__
            # 获取升级前对应表的结构
            structure = session.query(Structure).filter_by(table_name=table_name).first()
            old_columns = set(structure.columns.split(","))
            # 获取库表中的所有字段
            reflect_columns = self.get_table_column_names(reflect_table)
            cur_columns = self.get_table_column_names(table)
            # 弃用的字段 = 反射出来的字段 - 当前model使用的字段；需要设置成null
            useless_columns = reflect_columns - cur_columns
            # 增用的字段 = 当前使用的字段 - 上一版的字段；需要更新成默认值
            added_columns = cur_columns - old_columns

            update_ds = {name: None for name in useless_columns}
            # 获取默认配置
            for column in inspect(table).columns:
                if column.name not in added_columns:
                    continue
                # 如果需要从配置文件中获取默认配置，则需要更新到model中，否则这里需要感知具体模型
                default_value = column.default.arg if column.default else None
                update_ds[column.name] = default_value

            # 更新表数据
            session.query(reflect_table).update(update_ds)
            session.query(Structure).filter_by(table_name=table_name).update({"columns": ",".join(cur_columns)})

    def _migrate(self):
        with self.session_maker() as session:
            self._migrate_tables(session)


def input_args(argv=None):
    parse = ArgumentParser()
    parse.add_argument("operate", type=str, choices=MigrateOperate.values(), help="操作类型")
    return parse.parse_args(argv)
