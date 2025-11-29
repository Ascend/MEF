# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from collections import namedtuple

from common.token_bucket import TokenBucket

ConsumeTokenCase = namedtuple("ConsumeTokenCase", "expected, consume_token_amount")


class TestUtils:
    @staticmethod
    def locked():
        return True


class TestCopyOmSysFile:
    use_cases = {
        "test_consume_token": {
            "normal": (True, 1),
            "false": (False, 21),
        },
    }

    def test_consume_token(self, model: ConsumeTokenCase):
        ret = TokenBucket.consume_token(TokenBucket(), model.consume_token_amount)
        assert model.expected == ret
