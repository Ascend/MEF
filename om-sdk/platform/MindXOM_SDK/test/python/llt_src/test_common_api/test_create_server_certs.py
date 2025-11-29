from collections import namedtuple

from pytest_mock import MockerFixture

from common.file_utils import FileCheck
from common.file_utils import FileConfusion
from common.utils.result_base import Result

from create_server_certs import main, check_parameter, CertMgr, create_server_certs


class TestCreateServerCerts:
    MainCase = namedtuple("MainCase", "excepted, check_param, check_cert, create")
    CheckCertInfoCase = namedtuple("CheckCertInfoCase", "excepted, check_path")
    CheckParamCase = namedtuple("CheckParamCase", "excepted, test_argv, file_exist")
    CreateCase = namedtuple("CreateCase", "excepted, test_argv, is_cert_ok")
    CertMgrCreateCase = namedtuple("CertMgrCreateCase", "get_safe_pwd, check_input_path")
    CertOkCase = namedtuple("CertOkCase", "excepted, cert_dir, check_path, check_ret")

    use_cases = {
        "test_create_certs_main": {
            "check_parameter_failed": (False, False, False, False),
            "create_server_certs_failed": (False, True, True, False),
            "normal": (True, True, True, True),
        },
        "test_check_parameter": {
            "len_wrong": (False, ["test", "1", "1", "/home/data/config/default"], Result(True)),
            "uid_len_wrong": (False, ["test", "10000000000", "1", "/home/data/config/default", "normal"], Result(True)),
            "gid_len_wrong": (False, ["test", "1", "10000000000", "/home/data/config/default", "normal"], Result(True)),
            "uid_str_not_isdigit": (False, ["test", "1a", "1", "/home/data/config/default", "normal"], Result(True)),
            "gid_str_not_isdigit": (False, ["test", "1", "1a", "/home/data/config/default", "normal"], Result(True)),
            "cert_dir_not_in_range": (False, ["test", "1", "1", "/home/data/config/default1", "normal"], Result(True)),
            "type_not_in_range": (False, ["test", "1", "1", "/home/data/config/default", "normal1"], Result(True)),
            "not_exist": (False, ["test", "1", "1", "/home/data/config/default", "normal"], Result(False)),
            "normal": (True, ["test", "1", "1", "/home/data/config/default", "normal"], Result(True)),
        },
        "test_create_server_certs": {
            "is_cert_ok": (True, ["test", "1", "1", "/home/data/config/default", "normal"], True),
            "is_cert_ok_not_ok": (True, ["test", "1", "1", "/home/data/config/default", "normal"], False),
            "not_normal": (True, ["test", "1", "1", "/home/data/config/default", "test"], True),
        },
        "test_is_cert_ok": {
            "check_failed": (False, "/home/data/config/default", False, [False]),
            "check_success": (True, "/usr/local/mindx/MindXOMUpgrade/software/ibma/cert", True, [True, True]),
        }
    }

    @staticmethod
    def create_certs_mock(mocker: MockerFixture, model: CertMgrCreateCase):
        mocker.patch.object(FileCheck, "check_input_path_valid", side_effect=model.check_input_path)
        mocker.patch("os.fdopen")
        mocker.patch("os.open")

    def test_create_certs_main(self, mocker: MockerFixture, model: MainCase):
        mocker.patch("create_server_certs.check_parameter", return_value=model.check_param)
        mocker.patch("create_server_certs.create_server_certs", return_value=model.create)
        ret = main(["test", "test", "test", "test"])
        assert model.excepted == ret

    def test_check_parameter(self, mocker: MockerFixture, model: CheckParamCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.file_exist)
        ret = check_parameter(model.test_argv)
        assert model.excepted == ret

    def test_create_server_certs(self, mocker: MockerFixture, model: CreateCase):
        mocker.patch.object(CertMgr, "is_cert_ok", return_value=model.is_cert_ok)
        mocker.patch.object(CertMgr, "remove_old_cert_files")
        mocker.patch.object(CertMgr, "create_cert")
        mocker.patch.object(CertMgr, "change_server_cert_fils")
        mocker.patch.object(CertMgr, "remove_root_ca")
        ret = create_server_certs(*model.test_argv[1:])
        assert model.excepted == ret

    def test_remove_root_ca(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch.object(FileConfusion, "confusion_path")
        mocker.patch("os.remove")
        cert_mgr = CertMgr("1", "1", "/home/data/config/default", 0)
        ret = cert_mgr.remove_root_ca()
        assert ret is None

    def test_change_server_cert_fils(self, mocker: MockerFixture):
        mocker.patch("os.lchown")
        mocker.patch("os.chmod")
        cert_mgr = CertMgr("1", "1", "/home/data/config/default", 0)
        ret = cert_mgr.change_server_cert_fils()
        assert ret is None

    def test_remove_old_cert_files(self, mocker: MockerFixture):
        mocker.patch("os.path.exists", return_value=True)
        mocker.patch.object(FileConfusion, "confusion_path")
        mocker.patch("os.remove")
        cert_mgr = CertMgr("1", "1", "/home/data/config/default", 0)
        ret = cert_mgr.remove_old_cert_files()
        assert ret is None

    def test_is_cert_ok(self, mocker: MockerFixture, model: CertOkCase):
        mocker.patch.object(FileCheck, "check_path_is_exist_and_valid", return_value=model.check_path)
        mocker.patch("lib.Linux.systems.security_service.security_service_clib.check_cert_expired",
                     return_value=model.check_ret)
        cert_mgr = CertMgr("1", "1", "/home/data/config/default", 0)
        ret = cert_mgr.is_cert_ok(model.cert_dir)
        assert model.excepted == ret
