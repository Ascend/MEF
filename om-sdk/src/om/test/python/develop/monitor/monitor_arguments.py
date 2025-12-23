from argparse import ArgumentParser

from arguments import Arguments, add_common_arguments


class MonitorArguments(Arguments):
    # .so目录，需要配置到 Run/Debug Configurations > run_monitor > Environment variables中。否则加载*.so文件失败
    LD_LIBRARY_PATH = "/usr/local/mindx/MindXOM/lib"


def args_parse() -> MonitorArguments:
    parse = ArgumentParser(description="开发调试入参")
    add_common_arguments(parse)
    return parse.parse_args(namespace=MonitorArguments())
