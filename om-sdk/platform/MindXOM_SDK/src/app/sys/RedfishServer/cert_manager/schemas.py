# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from dataclasses import dataclass
from datetime import datetime


@dataclass
class CrlInfoSchema:
    last_update: datetime
    next_update: datetime


@dataclass
class CertInfoSchema:
    subject: str
    issuer: str
    start_time: datetime
    end_time: datetime
    serial_num: str
    signature_algorithm: str
    signature_len: int
    cert_version: int
    pubkey_type: int
    fingerprint: str
    key_cert_sign: int
    is_ca: int
    chain_num: int
    ca_sign_valid: bool
