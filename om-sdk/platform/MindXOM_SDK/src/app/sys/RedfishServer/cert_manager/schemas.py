# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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
