import contextlib
import os
from functools import partial
from unittest.mock import patch

from flask import Flask

from common.constants.base_constants import CommonConstants
from common.file_utils import FileCheck
from common.log.log_constant import LogConstant
from common.utils.app_common_method import AppCommonMethod
from debug_logger import init_logger
from ibma_client import Client
from redfish.redfish_arguments import RedfishArguments, args_parse

# REDFISH_ROOT_DIR：/<home_path>/<username>/scr/app/sys/RedfishServer
ROOT_DIR = os.path.join(os.path.realpath(__file__), os.pardir, os.pardir, os.pardir, os.pardir, os.pardir)
REDFISH_ROOT_DIR = os.path.join(os.path.realpath(ROOT_DIR), 'src', 'app', 'sys', 'RedfishServer')


def start_flask(app: Flask, ssl_ctx, args: RedfishArguments):
    """RedfishURIS._start_flask的替换函数"""
    app.run(args.host, port=args.port, ssl_context=ssl_ctx)


def debug_or_run_redfish(args: RedfishArguments):
    init_logger(LogConstant.REDFISH_MODULE_NAME, args)
    # 修改数据库地址
    CommonConstants.REDFISH_EDGE_DB_FILE_PATH = args.ibma_edge_db_path
    # 局部导入，保证CommonConstants.REDFISH_EDGE_DB_FILE_PATH先被赋值
    from ibma_redfish_urls import RedfishURIs
    from ibma_redfish_main import RedfishMain
    from user_manager.token_verification import TokenVerification

    with contextlib.ExitStack() as stack:
        # 调试时，是否启用https
        if not args.https:
            stack.enter_context(patch.object(RedfishURIs, "_get_ssl_context", return_value=None))

        # 调试时，更改启动ip与port
        start_flask_mocker = partial(start_flask, args=args)
        stack.enter_context(patch.object(RedfishURIs, "_start_flask", start_flask_mocker))

        # 调试接口时，是否启用鉴权
        if not args.auth:
            user_info = {
                "status": AppCommonMethod.OK,
                "message": {
                    "user_name": "root",
                    "user_id": 1,
                    "account_insecure_prompt": False
                }
            }
            stack.enter_context(patch.object(TokenVerification, "get_all_info", return_value=user_info))

        # 修改redfish与monitor交互的sock文件
        send_msg = partial(Client.send_msg, socket_path=args.socket_path)
        stack.enter_context(patch.object(Client, "send_msg", send_msg))
        stack.enter_context(patch.object(FileCheck, "check_path_mode_owner_group", return_value=True))

        RedfishMain.main()


if __name__ == '__main__':
    debug_or_run_redfish(args_parse())
