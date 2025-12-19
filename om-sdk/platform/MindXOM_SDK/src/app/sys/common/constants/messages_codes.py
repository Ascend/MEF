# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
class MessagesCode(object):
    code = None
    messageKey = None

    def __init__(self, code, message_key):
        # 初始化父项的属性
        super(MessagesCode, self).__init__()
        self.code = code
        self.messageKey = message_key

    def __repr__(self) -> str:
        return 'MessagesCode [code:{}, messageKey:{}]'.format(self.code, self.messageKey)
