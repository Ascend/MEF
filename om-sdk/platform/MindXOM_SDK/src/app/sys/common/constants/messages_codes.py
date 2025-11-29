# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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
