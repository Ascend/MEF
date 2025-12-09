#!/usr/bin/env python
# -*- coding:utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
class MidwareRoute(object):
    view_functions = {}

    def add_url_rule(self, rule, view_func, **options):
        MidwareRoute.view_functions[rule] = view_func

    def route(self, rule, **options):
        def decorator(func):
            self.add_url_rule(rule, func, **options)
            return func

        return decorator
