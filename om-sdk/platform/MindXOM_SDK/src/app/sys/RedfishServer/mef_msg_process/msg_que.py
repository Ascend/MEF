# coding: utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import queue

msg_que_to_mef = queue.Queue(maxsize=64)
msg_que_from_mef = queue.Queue(maxsize=64)
alarm_que_from_mef = queue.Queue(maxsize=1)
