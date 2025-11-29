# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from abc import ABC, abstractmethod
from flask import Blueprint


class Route(ABC):
    def __init__(self, blueprint: Blueprint):
        self.blueprint = blueprint

    @abstractmethod
    def add_route(self):
        pass
