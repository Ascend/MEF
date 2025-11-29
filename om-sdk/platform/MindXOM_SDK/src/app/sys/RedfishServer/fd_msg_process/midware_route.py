#!/usr/bin/env python
# -*- coding:utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
class MidwareRoute(object):
    view_functions = {}

    def add_url_rule(self, rule, view_func, **options):
        MidwareRoute.view_functions[rule] = view_func

    def route(self, rule, **options):
        def decorator(func):
            self.add_url_rule(rule, func, **options)
            return func

        return decorator
