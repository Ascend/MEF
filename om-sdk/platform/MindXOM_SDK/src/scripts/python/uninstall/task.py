# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from abc import ABC, abstractmethod
from typing import Generator, Callable, Any


class BaseTask(ABC):
    name: str = "Base task"

    @abstractmethod
    def steps(self) -> Generator[Callable[..., Any], None, None]:
        """任务执行步骤"""
        pass

    def run(self):
        for step in self.steps():
            step()
