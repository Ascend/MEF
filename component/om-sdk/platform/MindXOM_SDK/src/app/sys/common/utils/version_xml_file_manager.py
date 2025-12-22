# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import hashlib
import os
from typing import List
from xml.etree.ElementTree import parse as xml_parser

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.utils.exception_utils import OperateBaseError
from common.utils.result_base import Result


class XMLParseError(OperateBaseError):
    pass


class XmlManger:
    XML_DTD_KEY = "<!DOCTYPE"

    def __init__(self, xml_path=""):
        self.xml_path = xml_path
        if not self.xml_path:
            self.xml_path = CommonConstants.OM_VERSION_XML_FILE
        self._parse_xml()

    @staticmethod
    def _check_xml_is_safe(xml_path):
        xml_dtd_key = "<!DOCTYPE"
        xml_max_size = 512 * 1024
        if os.path.getsize(xml_path) > xml_max_size:
            return Result(result=False, err_msg="This xml file is too large.")

        try:
            with open(xml_path, "r") as fw:
                content = fw.read()
                if xml_dtd_key in content:
                    return Result(result=False, err_msg="This is not a valid xml, it exist dtd content")
                if len(content.splitlines()) > CommonConstants.MAX_ITER_LIMIT:
                    return Result(result=False, err_msg="This is not a valid xml, its lines exceed limit.")
                return Result(result=True)
        except Exception as ex:
            return Result(result=False, err_msg=f"Read xml file failed, find exception: {ex}")

    def _parse_xml(self):
        res = FileCheck.check_path_is_exist_and_valid(self.xml_path)
        if not res:
            raise ValueError("Not a valid version xml")

        res = self._check_xml_is_safe(self.xml_path)
        if not res:
            raise ValueError(f"Not a valid version xml, {res.error}")

        self.xml_tree = xml_parser(self.xml_path)


class VersionXmlManager(XmlManger):

    @property
    def version(self) -> str:
        return self.xml_tree.findtext("Package/Version") or ""

    @property
    def module(self) -> str:
        return self.xml_tree.findtext("Package/Module") or ""

    @property
    def processor_architecture(self) -> str:
        return self.xml_tree.findtext("Package/ProcessorArchitecture") or ""

    @property
    def firmware_version(self) -> str:
        return self.xml_tree.findtext("Package/FirmwareVersion") or ""

    @property
    def vendor(self) -> str:
        return self.xml_tree.findtext("Package/Vendor") or ""


class VercfgManager(XmlManger):

    # 升级包中单个文件的大小上限，单位字节
    FILE_MAX_SIZE_BYTES = 300 * 1024 * 1024
    # 升级包中的固件数量
    MAX_FIRMWARE_COMPONENTS = 10

    def __init__(self, vercfg_path):
        super(VercfgManager, self).__init__(vercfg_path)
        self.dir_path = os.path.dirname(self.xml_path)
        self.upgrade_module_list = []

    def verify_sha256(self) -> List[str]:
        root = self.xml_tree.getroot()
        for module in root.iterfind("File"):
            filename = os.path.basename(module.find("FilePath").text)
            file_sha256 = module.find("SHAValue").text
            filepath = os.path.join(self.dir_path, filename)
            # 计算文件sha256值，与vercfg.xml中的sha256值进行对比
            cal_sha256 = self._read_file_sha256(filepath)
            if file_sha256 != cal_sha256:
                raise XMLParseError("sha256 verify failed")
            # 记录所有的升级包文件
            if filepath.endswith("tar.gz") or filepath.endswith("hpm"):
                self.upgrade_module_list.append(filepath)

        if len(self.upgrade_module_list) > self.MAX_FIRMWARE_COMPONENTS:
            raise XMLParseError("the number of components to be upgraded exceeds the maximum limit")

        return self.upgrade_module_list

    def _read_file_sha256(self, path) -> str:
        sha256 = hashlib.sha256()
        # 检查路径
        ret = FileCheck.check_path_is_exist_and_valid(path)
        if not ret:
            raise XMLParseError(f"check xml path failed, {ret.error}")
        # 检查大小
        if os.path.getsize(path) > self.FILE_MAX_SIZE_BYTES:
            raise XMLParseError(f"{path} is too large")
        try:
            with open(path, "rb") as file:
                sha256.update(file.read())
                return sha256.hexdigest()
        except Exception as err:
            raise XMLParseError(f"Read {path} failed, because {err}") from err
