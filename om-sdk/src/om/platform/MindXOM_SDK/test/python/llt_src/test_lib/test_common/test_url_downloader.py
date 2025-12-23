import threading
from collections import namedtuple
from unittest.mock import Mock

from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.utils.exec_cmd import ExecCmd
from common.utils.result_base import Result

from common.utils.app_common_method import AppCommonMethod
from common.utils.url_downloader import MAX_FILE_SIZE
from common.utils.url_downloader import UrlConnect
from common.utils.url_downloader import UrlDownloader
from common.utils.url_downloader import check_file_valid
from common.utils.url_downloader import https_download_file
from net_manager.manager.fd_cfg_manager import FdConfigData
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import NetManager
from ut_utils.mock_utils import mock_path_exists
from ut_utils.mock_utils import mock_read_data
from ut_utils.mock_utils import mock_time_sleep
from ut_utils.mock_utils import mock_write_file_with_os_open

HttpsDownloadFile = namedtuple("HttpsDownloadFile", "expect, ret, valid")
CheckFileValid = namedtuple("CheckFileValid", "expect, exists, cmd")
GetContext = namedtuple("GetContext", "expect, net_manager_obj, ssl_context")
MakeRangeString = namedtuple("MakeRangeString", "expect, start, end")
IsConnected = namedtuple("IsConnected", "expect, resp, fp, closed")
CloseModel = namedtuple("CloseModel", "expect, resp")
ConnectModel = namedtuple("ConnectModel", "expect, url, content, resp")
DownloadFile = namedtuple("DownloadFile",
                          "expect, f_bavail, f_frsize, size, tuple_ret, seek, local_size, connect, make")
GetFileName = namedtuple("GetFileName", "expect, url")
GetFileSize = namedtuple("GetFileSize", "expect, size, tuple_ret")
GetDownFileSeek = namedtuple("GetDownFileSeek", "expect, valid, size")
GetLocalFileSize = namedtuple("GetLocalFileSize", "expect, size")
ConnectIfNeed = namedtuple("ConnectIfNeed", "expect, retry, connect")
MakeDirTree = namedtuple("MakeDirTree", "expect, exists, dirs")
RemoveFileDir = namedtuple("RemoveFileDir", "expect, link, remove")
MakeLocalFile = namedtuple("MakeLocalFile", "expect, tree, size, exists, write, download")
WriteSeek = namedtuple("WriteSeek", "expect, total_size, b_block")


class TestUrlDownloader:
    use_cases = {
        "test_https_download_file": {
            "Download file failed": ([-1, "download file failed"], [0, "", 1], None),
            "Download file failed-1": ([-1, "Download file failed: err."], [1, ], [0, "err"]),
            "normal": ([0, ""], [1, ], [1]),
        },
        "test_check_file_valid": {
            "check file valid failed, file not exist": (
                [False, "check file valid failed, file not exist"], False, None
            ),
            "Exec cmd failed.": ([False, "Exec cmd failed."], True, ValueError),
            "check code verify failed": ([False, "check code verify failed"], True, [[1, ], ])
        },
        "test_get_context": {
            "invalid": (False, NetManager(), Result(False)),
            "normal": (
                False,
                NetManager(),
                Result(False),
            ),
        },
        "test_make_range_string": {
            "bytes=0-": ("bytes=0-", None, None),
            "null end": ("bytes=1-", 1, None),
            "normal": ("bytes=1-2", 1, 2),
        },
        "test_is_connected": {
            "null resp": (False, None, None, None),
            "null fp": (False, True, None, None),
            "not close": (True, True, True, False),
        },
        "test_close": {
            "true": (True, True),
            "false": (False, False),
        },
        "test_connect": {
            "URL is None.": ((False, "URL is None.", 102400), "", None, None),
            "null context": ((False, "err", 102499), "1", Result(False, err_msg="err"), None),
            "Username or Pword error": (
                (False, "Username or Pword error.", 102401), "1", Result(True, ),
                {"download_error": "please reinput username"}
            ),
            "Open URL download_error": (
                (False, "Open URL download_error", 102499), "1", Result(True, ),
                {"download_error": "err"}
            )
        },
        "test_download_file": {
            "negative file size": ((False, "File size invalid.", 102502), 1, 1, [-1, ], -1, None, None, None, None),
            "File Size Oversize": (
                (False, "File size oversize.", 102499), 10, 10, [200, ], -1, None, None, None, None
            ),
            "file already down Success": (
                (True, "file already down Success.", 0), 10, 10, [90, 90], -1, 80, [90, 90], None, None
            ),
            "not connect": (
                [False, ""], 10, 10, [90, 90, 80], -1, 80, [70, 70, 70], [False, ""], None
            ),
            "normal": (
                "make", 10, 10, [90, 90, 80], -1, 80, [70, 70, 70], [True, ""], "make"
            ),
            "Download meet exception": (
                (False, "Download meet exception", 102499), 10, 10, [70, 90, 80], -1, 80, [70, 70, 70],
                [True, ""], (False, 'Download meet exception', 102499)
            )
        },
        "test_get_file_name": {
            "GetFileName meet Exception": ("", None),
            "filename": ("name", "fileName=name"),
            "filename-1": ("name", "/name"),

        },
        "test_get_file_size": {
            "existed": (1, 1, [False, 1]),
            "false tuple result": (-1, -1, [False, 1]),
            "normal": (-1, -1, [True, 1]),
            "exception": (-1, -1, None),
        },
        "test_get_down_file_seek": {
            "exception": (0, ValueError, None),
            "normal": (1, [True, ], "1/")
        },
        "test_get_local_file_size": {
            "exists": (10, 10)
        },
        "test_connect_if_need": {
            "102401": ([True, "", 102401], None, [[True, "", 102401]]),
            "503": ([False, "", 504], None, [[False, "", 503], [False, "", 504]])
        },
        "test_make_dir_tree": {
            "normal": (True, False, [True]),
            "exception": (False, False, ValueError),
            "not exists": (True, False, [True, ]),
        },
        "test_make_local_file": {
            "Make download dir error": (
                (False, "Make download dir error.", 102500), False, None, None, None, None
            ),
            "Download File size less than 0 or bigger than": (
                (False, f"Download File size less than 0 or bigger than {MAX_FILE_SIZE}.", 102502), True, [-1, ], None,
                None, None
            ),
            "write seek err": (
                [0, "write seek err"], True, [10, ], [True, True, ], [0, "write seek err"], None
            ),
            "write seek err-1": (
                [0, "write seek err-1"], True, [10, ], [True, False, ], [0, "write seek err-1"], None
            ),
            "Open file for write error": (
                (False, "Open file for write error.", 102501), True, [10, ], FileNotFoundError, [0, ], None
            ),
            "retrieval incomplete": (
                (False, "retrieval incomplete: got only 10 out of 10 bytes, please download again.", 102504), True,
                [10, 100], [True, True], [1, ], [10, 10]
            ),
            "download file is larger than expected": (
                (False,
                 "download file is larger than expected: got 1000 bigger than 10 bytes, please download again.",
                 102504), True, [10, 100, 100], [True, True], [1, ], [1000, 1000, 1000]
            ),
            "Success": (
                (True, "Success", 0), True, [10, 100, 100], [True, True], [1], [100, 100, 100]
            )
        },
        "test_write_seek": {
            "Write file success": ([True, "Write file success"], 100, ["1", ""]),
            "Write file error": ([False, "Write file meet exception.", 102508], 100, ValueError)
        }
    }

    @staticmethod
    def test_https_download_file(mocker: MockerFixture, model: HttpsDownloadFile):
        mocker.patch.object(UrlDownloader, "download_file", return_value=model.ret)
        mocker.patch("common.utils.url_downloader.check_file_valid", return_value=model.valid)
        assert https_download_file("", "", "", "", True) == model.expect

    @staticmethod
    def test_check_file_valid(mocker: MockerFixture, model: CheckFileValid):
        mock_path_exists(mocker, return_value=model.exists)
        mocker.patch.object(ExecCmd, "exec_cmd_use_pipe_symbol", side_effect=model.cmd)
        mocker.patch("os.remove")
        assert check_file_valid("", "") == model.expect

    @staticmethod
    def test_get_context(mocker: MockerFixture, model: GetContext):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info", return_value=model.net_manager_obj)
        mocker.patch.object(FdConfigData, "gen_client_ssl_context", return_value=model.ssl_context)
        assert bool(UrlConnect.get_context()) == model.expect

    @staticmethod
    def test_make_range_string(model: MakeRangeString):
        assert UrlConnect._make_range_string(model.start, model.end) == model.expect

    @staticmethod
    def test_is_connected(model: IsConnected):
        connect = UrlConnect("", "", "")
        connect._obj_response = None
        if model.resp:
            connect._obj_response = Mock(return_value=model.resp)
            connect._obj_response.fp = model.fp
        if model.fp:
            connect._obj_response.fp = Mock(return_value=model.fp)
            connect._obj_response.fp.closed = model.closed

        assert connect.is_connected() == model.expect

    @staticmethod
    def test_read():
        connect = UrlConnect("", "", "")
        connect._obj_response = Mock()
        connect._obj_response.read.return_value = "data"
        assert connect.read(1024) == "data"

    @staticmethod
    def test_close(model: CloseModel):
        connect = UrlConnect("", "", "")
        connect._obj_response = Mock(return_value=model.resp) if model.resp else None
        assert connect.close() == model.resp

    @staticmethod
    def test_connect(mocker: MockerFixture, model: ConnectModel):
        mocker.patch("urllib.request.Request")
        mocker.patch.object(UrlConnect, "_make_range_string")
        mocker.patch.object(AppCommonMethod, "make_authentication_string")
        mocker.patch.object(threading, "Thread")
        mocker.patch.object(UrlConnect, "get_context", return_value=model.content)
        mocker.patch("urllib.request.urlopen").return_value.info.return_value = model.resp
        connect = UrlConnect("", "", "")
        connect._str_url = model.url
        assert connect.connect() == model.expect

    @staticmethod
    def test_download_file(mocker: MockerFixture, model: DownloadFile):
        rest = mocker.patch("os.statvfs").return_value
        rest.f_bavail = model.f_bavail
        rest.f_frsize = model.f_frsize
        mocker.patch.object(UrlDownloader, "get_file_size", side_effect=model.size)
        mocker.patch.object(UrlDownloader, "get_down_file_seek", return_value=model.seek)
        mocker.patch.object(UrlDownloader, "get_local_file_size", side_effect=model.local_size)
        mock_path_exists(mocker, return_value=True)
        mocker.patch("os.remove")
        mocker.patch.object(UrlDownloader, "_UrlDownloader__make_local_file", return_value=model.make)
        downloader = UrlDownloader("", "", "")
        downloader._tuple_result = model.tuple_ret
        downloader._obj_url_connect = Mock()
        downloader._obj_url_connect.connect.return_value = model.connect
        assert downloader.download_file("") == model.expect

    @staticmethod
    def test_get_file_name(model: GetFileName):
        downloader = UrlDownloader("", "", "")
        downloader._str_url = model.url
        assert downloader.get_file_name() == model.expect

    @staticmethod
    def test_get_file_size(mocker: MockerFixture, model: GetFileSize):
        downloader = UrlDownloader("", "", "")
        downloader._n_sum_size = model.size
        downloader._tuple_result = model.tuple_ret
        if model.tuple_ret:
            downloader._tuple_result[1] = Mock()
            downloader._tuple_result[1].headers = {"content-length": model.tuple_ret[1],
                                                   "Content-Length": model.tuple_ret[1]}
            mocker.patch.object(UrlDownloader, "_connect_if_need")
        assert downloader.get_file_size() == model.expect

    @staticmethod
    def test_get_down_file_seek(mocker: MockerFixture, model: GetDownFileSeek):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", side_effect=model.valid)
        mock_read_data(mocker, read_data=model.size)
        assert UrlDownloader("", "", "").get_down_file_seek("file") == model.expect

    @staticmethod
    def test_get_local_file_size(mocker: MockerFixture, model: GetLocalFileSize):
        mock_path_exists(mocker, return_value=True)
        mocker.patch("os.path.getsize", return_value=model.size)
        assert UrlDownloader("", "", "").get_local_file_size("") == model.expect

    @staticmethod
    def test_get_download_size():
        downloader = UrlDownloader("", "", "")
        downloader._n_download_size = 10
        assert downloader.get_download_size() == 10

    @staticmethod
    def test_get_download_progress(mocker: MockerFixture):
        downloader = UrlDownloader("", "", "")
        downloader._n_sum_size = 10
        mocker.patch.object(UrlDownloader, "get_download_size", return_value=11)
        assert downloader.get_download_progress() == (10, 11)

    @staticmethod
    def test_set_download_size():
        downloader = UrlDownloader("", "", "")
        downloader._set_download_size(10)
        assert downloader._n_download_size == 10

    @staticmethod
    def test_connect_if_need(mocker: MockerFixture, model: ConnectIfNeed):
        mocker.patch.object(UrlConnect, "connect", side_effect=model.connect)
        mocker.patch("gc.collect")
        mock_time_sleep(mocker)
        assert UrlDownloader("", "", "")._connect_if_need() == model.expect

    @staticmethod
    def test_make_dir_tree(mocker: MockerFixture, model: MakeDirTree):
        mock_path_exists(mocker, return_value=model.exists)
        mocker.patch("os.makedirs", side_effect=model.dirs)
        assert UrlDownloader("", "", "")._make_dir_tree("") == model.expect

    @staticmethod
    def test_make_local_file(mocker: MockerFixture, model: MakeLocalFile):
        mocker.patch.object(UrlDownloader, "_make_dir_tree", return_value=model.tree)
        mocker.patch.object(UrlDownloader, "get_down_file_seek")
        mocker.patch.object(UrlDownloader, "get_file_size", side_effect=model.size)
        mock_path_exists(mocker, side_effect=model.exists)
        mocker.patch("os.remove")
        mocker.patch.object(FileCheck, "check_is_link")
        mocker.patch("builtins.open")
        mock_write_file_with_os_open(mocker)
        mocker.patch.object(UrlDownloader, "_UrlDownloader__write_seek", return_value=model.write)
        mocker.patch.object(UrlDownloader, "dis_connect")
        mocker.patch.object(UrlDownloader, "get_download_size", side_effect=model.download)
        mocker.patch.object(UrlDownloader, "_UrlDownloader__clear_download_file")
        assert UrlDownloader("", "", "")._UrlDownloader__make_local_file("") == model.expect

    @staticmethod
    def test_write_seek(mocker: MockerFixture, model: WriteSeek):
        mocker.patch.object(UrlDownloader, "get_file_size", return_value=model.total_size)
        downloader = UrlDownloader("", "", "")
        downloader._obj_url_connect = UrlConnect("", "", "")
        mocker.patch.object(UrlConnect, "read", side_effect=model.b_block)
        mocker.patch.object(UrlDownloader, "_set_download_size")
        assert downloader._UrlDownloader__write_seek(Mock(), Mock(), 10) == model.expect
