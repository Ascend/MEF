# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import os


class UploadConstants:
    CONF_ALLOWED_EXTENSIONS = ("conf",)
    INI_ALLOWED_EXTENSIONS = ("ini",)
    CERT_ALLOWED_EXTENSIONS = ("crt", "crl", "cer")
    ZIP_ALLOWED_EXTENSIONS = ("zip",)

    UPLOAD_BASE_DIR = "/run/web"
    ZIP_UPLOAD_DIR = os.path.join(UPLOAD_BASE_DIR, "zip")
    CONF_UPLOAD_DIR = os.path.join(UPLOAD_BASE_DIR, "conf")
    INI_UPLOAD_DIR = os.path.join(UPLOAD_BASE_DIR, "ini")
    CERT_UPLOAD_DIR = os.path.join(UPLOAD_BASE_DIR, "cert")
    CSR_UPLOAD_DIR = os.path.join(UPLOAD_BASE_DIR, "csr")
    NET_MANAGER_UPLOAD_DIR = os.path.join(UPLOAD_BASE_DIR, "net_manager")
    MARK_FILE_DIR = os.path.join(UPLOAD_BASE_DIR, "flag")

    EXT_MARK_TYPE = {
        "conf": "conf",
        "ini": "ini",
        "crt": "cert",
        "cer": "cert",
        "crl": "crl",
        "zip": "zip",
    }

    MARK_FILES = {
        "conf": os.path.join(MARK_FILE_DIR, "conf_upload_flag"),
        "ini": os.path.join(MARK_FILE_DIR, "ini_upload_flag"),
        "cert": os.path.join(MARK_FILE_DIR, "cert_upload_flag"),
        "crl": os.path.join(MARK_FILE_DIR, "cert_upload_flag"),
        "zip": os.path.join(MARK_FILE_DIR, "zip_upload_flag")
    }

    FILE_SAVE_DIR = {
        "conf": CONF_UPLOAD_DIR,
        "ini": INI_UPLOAD_DIR,
        "crt": CERT_UPLOAD_DIR,
        "cer": CERT_UPLOAD_DIR,
        "crl": CERT_UPLOAD_DIR,
        "zip": ZIP_UPLOAD_DIR
    }

    CONF_MAX_SIZE = 20 * 1024 * 1024
    INI_MAX_SIZE = 10 * 1024
    CERT_MAX_SIZE = 10 * 1024
    ZIP_MAX_SIZE = 512 * 1024 * 1024

    FILE_MAX_SIZE = {
        "conf": CONF_MAX_SIZE,
        "ini": INI_MAX_SIZE,
        "crt": CERT_MAX_SIZE,
        "cer": CERT_MAX_SIZE,
        "crl": CERT_MAX_SIZE,
        "zip": ZIP_MAX_SIZE
    }

    TIMEOUT = 600  # 长传文件超时时间，600s = 10min

    # FD下载只支持ZIP包下载
    DRIVER_DOWNLOAD_PATH = os.path.join(UPLOAD_BASE_DIR, "fd_zip")
    DOWNLOAD_IMAGE_NAME = "A500-3000-3010-firmware.zip"

    # web保存zip升级包固定名字
    WEB_DOWNLOAD_PACKAGE_NAME = "web_upgrade.zip"

    # 弱口令文件固定文件名
    WEAK_DICT_NAME = "weak_dict.conf"
