#!/usr/bin/python
# -*- coding: UTF-8 -*-

"""
功    能：Init处理模块
修改记录：2021-11-18 创建
"""

import os
import sys


class TestEnv:
    @staticmethod
    def init_env():
        """
        功能描述：OM测试用例环境初始化函数
        参数：无
        返回值：无
        异常描述：NA
        """

        # 获取MindXOM目录绝对路径
        root_dir = os.path.dirname(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

        # 安装根目录
        install_root_dir = os.path.join(root_dir, "install")

        # python 模块
        python_path_nginx = os.path.join(install_root_dir, "software/nginx")
        python_path_ibma = os.path.join(install_root_dir, "software/ibma")
        python_path_ibma_bin = os.path.join(install_root_dir, "software/ibma/bin")
        python_path_ibma_lib = os.path.join(install_root_dir, "software/ibma/lib")
        python_path_opensource = os.path.join(install_root_dir, "software/ibma/opensource/python")
        python_path_redfish = os.path.join(install_root_dir, "software/RedfishServer")
        llt_test_path = os.path.dirname(os.path.realpath(__file__))
        os.environ.setdefault('PYTHONPATH', python_path_nginx + ":" + python_path_ibma + ":"
                              + python_path_ibma_bin + ":" + python_path_ibma_lib + ":"
                              + python_path_opensource + ":" + python_path_redfish)

        sys.path.append(python_path_nginx)
        sys.path.append(python_path_ibma)
        sys.path.append(python_path_ibma_bin)
        sys.path.append(python_path_ibma_lib)
        sys.path.append(python_path_opensource)
        sys.path.append(llt_test_path)
        sys.path.append(python_path_redfish)

        # bin路径
        bin_path_ens = os.path.join(install_root_dir, "software/ens/bin")
        os.environ.setdefault('PATH', bin_path_ens + ":" + bin_path_ens)

        # lib路径
        lib_path = os.path.join(install_root_dir, "lib/")
        lib_path_ens_lib = os.path.join(install_root_dir, "software/ens/lib/")
        lib_path_ens_modules = os.path.join(install_root_dir, "software/ens/modules/")
        os.environ.setdefault('LD_LIBRARY_PATH', lib_path + ":"
                              + lib_path_ens_lib + ":" + lib_path_ens_modules)


TestEnv.init_env()
