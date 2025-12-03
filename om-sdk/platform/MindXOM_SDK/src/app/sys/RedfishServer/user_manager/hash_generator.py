# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import base64
import hashlib
import hmac

from common.log.logger import run_log


class HashGenerator:
    KEY_DERIVATION_ALGORITHM = "pbkdf2"
    HASH_ALGORITHM = "sha512"
    ITERATOR_COUNT = 100000
    SALT_LENGTH = 24

    @staticmethod
    def generate_password_hash(pass_word):
        """
        根据pass_word计算hash值
        :param pass_word: 待计算的pass_word
        :return: 第一个返回值返回hash值生成结果；当生成成功时，第二个返回值返回生成的hash值，否则返回错误描述
        """
        salt = HashGenerator._get_safe_salt()
        encode_value = hashlib.pbkdf2_hmac(HashGenerator.HASH_ALGORITHM,
                                           pass_word.encode("utf-8"),
                                           salt,
                                           HashGenerator.ITERATOR_COUNT,)
        safe_salt = base64.b64encode(salt).decode("utf-8")
        hash_value = base64.b64encode(encode_value).decode("utf-8")
        return f"{HashGenerator.KEY_DERIVATION_ALGORITHM}:{HashGenerator.HASH_ALGORITHM}:" \
               f"{HashGenerator.ITERATOR_COUNT}${safe_salt}${hash_value}"

    @staticmethod
    def check_password_hash(hash_pass_word, pass_word):
        """
        将传入的pass_word计算hash值之后，和hash_pass_word进行比较，以判断密码是否正确
        :param hash_pass_word:保存在数据库中的hash值
        :param pass_word:待比较的密码
        :return:比较结果
        """
        if hash_pass_word.count("$") < 2:
            run_log.error("parse hash pword failed!")
            return False

        method, salt, hash_value = hash_pass_word.split("$", 2)
        if method.count(":") != 2:
            run_log.error("parse hash pword failed!")
            return False

        pb_algorithm, hash_name, iterations = method.split(":", 2)
        if pb_algorithm != HashGenerator.KEY_DERIVATION_ALGORITHM:
            run_log.error("parse hash pword failed!")
            return False

        try:
            iterations = int(iterations)
        except Exception:
            run_log.error("get iterations failed.")
            return False

        encode_value = hashlib.pbkdf2_hmac(hash_name,
                                           pass_word.encode("utf-8"),
                                           base64.b64decode(salt.encode("utf-8")),
                                           iterations)
        new_hash_value = base64.b64encode(encode_value).decode("utf-8")
        return HashGenerator.safe_string_compare(hash_value, new_hash_value)

    @staticmethod
    def safe_string_compare(a, b):
        if isinstance(a, str):
            a = a.encode("utf-8")
        if isinstance(b, str):
            b = b.encode("utf-8")

        return hmac.compare_digest(a, b)

    @staticmethod
    def _get_safe_salt():
        with open("/dev/random", "rb") as file_read:
            salt = file_read.read(HashGenerator.SALT_LENGTH)
            return salt
