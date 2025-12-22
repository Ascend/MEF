import contextlib
from unittest.mock import patch

from bin.monitor import main
from common.log.log_constant import LogConstant
from debug_logger import init_logger
from lib.common.common_methods import CommonMethods
from monitor.monitor_arguments import MonitorArguments
from monitor.monitor_arguments import args_parse


def debug_or_run_monitor(args: MonitorArguments):
    init_logger(LogConstant.MONITOR_MODULE_NAME, args)
    with contextlib.ExitStack() as stack:
        # 将monitor进程与redfish交互的iBMA.sock文件替换成f"iBMA_{args.user}.sock"
        stack.enter_context(patch.object(CommonMethods, "get_config_value", return_value=args.socket_path))

        # 启动monitor
        main()


if __name__ == '__main__':
    debug_or_run_monitor(args_parse())
