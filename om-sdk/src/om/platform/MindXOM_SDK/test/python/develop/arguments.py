from argparse import ArgumentParser


class Arguments:
    user: str

    @property
    def socket_path(self) -> str:
        return f"/run/iBMA_{self.user}.sock"


def add_common_arguments(parse: ArgumentParser):
    """公共的参数"""
    parse.add_argument("--user", "-u", type=str, required=True, help="调试者名称，同一个调试环境需要唯一，避免冲突")
