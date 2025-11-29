#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

from flask import Flask, Blueprint

app = Flask(__name__)


class TestUtils:
    user_id = "12345"


def get_client(blueprint: Blueprint):
    app.register_blueprint(blueprint)
    app.testing = True
    return app.test_client()
