# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import sys
from dataclasses import is_dataclass, fields, dataclass, asdict, MISSING, Field
from typing import Dict, Any, Union, Optional


class BaseModel:

    @classmethod
    def from_dict(cls, data: dict):
        """
        将输入数据转化成数据模型实例：
            目前仅支持嵌套字典结构；
            data中多余字段将被忽略；
            模型中相较data多余的字段，若没有设置默认值且field(init=True)则会触发TypeError异常；
            支持alias别名，alias对应data中的key；
        示例1：正常创建模型
            @dataclass
            class Example(BaseModel):
                age: int = field(alias="Age")
                name: str = field(default="example")    外部参数未传，则使用默认值，init=True
            print(Example.from_dict({"Age": 18}))
            >> Example(age=18, name='example')

        示例2：错误的模型
            @dataclass
            class Example(BaseModel):
                age: int = field(alias="Age")
                name: str = field()    外部参数未传，未设默认值，init=True，则会触发异常
            print(Example.from_dict({"Age": 18}))
            >> TypeError: __init__() missing 1 required positional argument: 'name'

        示例3：嵌套模型
            @dataclass
            class SubExample(BaseModel):
                age: int = field(alias="Age")
                name: str = field(default="sub")

            @dataclass
            class Example(BaseModel):
                age: int = field(alias="Age")
                sub: SubExample = field(default=SubExample)

            print(Example.from_dict({"Age": 18, "sub": {"Age": 20}}))
            >> Example(age=18, sub=SubExample(age=20, name='sub'))
        """
        if not is_dataclass(cls):
            raise TypeError(f"{cls.__name__} is not dataclass.")

        _items: Dict[str, Union[BaseModel, Any]] = {}
        for _field in fields(cls):
            if not _field.init:
                continue
            name = _field.metadata.get("alias") or _field.name
            # 字段不在输入数据，可能是赋有默认值或由实例化时触发字段异常
            if name not in data:
                continue
            value = data[name]
            # 嵌套模型
            nested = not hasattr(_field.type, "__origin__") and issubclass(_field.type, BaseModel)
            if nested and isinstance(value, dict):
                value = _field.type.from_dict(value)
            _items[_field.name] = value
        return cls(**_items)

    def to_dict(self):
        if not is_dataclass(self):
            raise TypeError(f"{self.__class__.__name__} is not dataclass.")

        return asdict(self)


def field(*, default=MISSING, alias: Optional[str] = None, comment: str = "", init=True, to_repr=True,
          to_hash=None, compare=True, kw_only=MISSING):
    """自封装的field函数，支持alias与备注"""
    metadata = {"alias": alias, "comment": comment} if alias or comment else None
    default_factory = MISSING
    if callable(default):
        default_factory = default
        default = MISSING

    # 适配python 3.10以及以上版本 dataclasses Field
    if sys.version_info[:2] >= (3, 10):
        return Field(default, default_factory, init, to_repr, to_hash, compare, metadata, kw_only)

    return Field(default, default_factory, init, to_repr, to_hash, compare, metadata)


@dataclass
class AdapterResult(BaseModel):
    """Adopter接口返回的数据结构"""
    status: int
    message: Union[dict, str]
