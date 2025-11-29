#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import itertools
import os
import random
import subprocess
from pathlib import Path
from tempfile import NamedTemporaryFile
from typing import Iterable

import pytest
from pytest_mock import MockerFixture

from cert_manager.cert_clib_mgr import CertClibMgr
from cert_manager.parse_tools import CertChainParser
from common.db.database import DataBase
from common.db.migrate import Migrate
from common.utils.singleton import Singleton
from net_manager.manager import net_cfg_manager
from net_manager.manager.net_cfg_manager import CertMgr
from net_manager.models import NetManager, CertManager, CertInfo, FdPreCert, PreCertInfo
from redfish_db.init_structure import INIT_COLUMNS


class TestMigrate(Migrate):

    @classmethod
    def instance(cls):
        return cls._instances.get(cls)


@pytest.fixture(scope="package")
def database(package_mocker: MockerFixture) -> DataBase:
    with NamedTemporaryFile(suffix=".db") as tmp_file:
        db_path = tmp_file.name
    TestMigrate.register_models(NetManager, CertManager, CertInfo, FdPreCert, PreCertInfo)
    TestMigrate.execute_on_install(db_path, INIT_COLUMNS)
    test_db = TestMigrate.instance()
    package_mocker.patch.object(CertClibMgr, "CLIB_PATH", Path("lib", Path(CertClibMgr.CLIB_PATH).name).as_posix())
    package_mocker.patch.object(net_cfg_manager, "session_maker", test_db.session_maker)
    package_mocker.patch.object(net_cfg_manager, "simple_session_maker", test_db.simple_session_maker)
    package_mocker.patch.object(CertMgr, "session_maker", test_db.simple_session_maker)
    yield test_db
    os.remove(db_path)


@pytest.fixture()
def init_certs_db(database: DataBase) -> DataBase:
    with database.session_maker() as session:
        session.query(CertMgr.cert).delete()
        session.query(CertMgr.cert_info).delete()
        session.query(CertMgr.pre_cert).delete()
        session.query(CertMgr.pre_cert_info).delete()
    return database


class CertPath(Singleton):
    GEN_SH = Path(__file__).parent.joinpath("make_certs.sh")
    ROOT_FMT = "/tmp/certs_{}"

    def __init__(self):
        for num in range(1, 3):
            self.create_certs(num)

    def create_certs(self, num: int):
        ret = subprocess.run(("bash", self.GEN_SH.as_posix(), self.ROOT_FMT.format(num)), capture_output=True)
        if ret.returncode != 0:
            raise RuntimeError(f"prepare certs failed. {ret.stderr}")

    def root(self, num: int) -> Path:
        return Path(self.ROOT_FMT.format(num)).joinpath("rootCa")

    def root_ca(self, num: int) -> Path:
        return self.root(num).joinpath("cacert.pem")

    def ec_256_ca(self, rt_num: int) -> Iterable[Path]:
        yield self.root(rt_num).joinpath("cacert_ec_256.pem")

    def ec_128_ca(self, rt_num: int) -> Iterable[Path]:
        yield self.root(rt_num).joinpath("cacert_ec_128.pem")

    def server_crt(self, rt_num: int, in_num: str) -> Path:
        return self.root(rt_num).joinpath(f"server_{in_num}.crt")

    def server_keyfile(self, rt_num: int, in_num: str) -> Path:
        return self.root(rt_num).joinpath(f"server_{in_num}").joinpath("server.key")

    def root_crl(self, num: int) -> Path:
        return self.root(num).joinpath("rootca.crl")

    def inter_ca(self, rt_num: int, in_num: str) -> Path:
        return self.root(rt_num).joinpath(f"interCa_{in_num}", "subcacrt.pem")

    def inter_crl(self, rt_num: int, in_num: str) -> Path:
        return self.root(rt_num).joinpath(f"interCa_{in_num}.crl")

    def single_root_ca(self, rt_num: int = 1) -> Iterable[Path]:
        yield self.root_ca(rt_num)

    def single_root_crl(self, rt_num: int = 1) -> Iterable[Path]:
        yield self.root_crl(rt_num)

    def single_root_chain_for_crl(self) -> Iterable[Iterable[Path]]:
        yield self.single_root_ca(1)
        yield self.single_root_ca(2)

    def single_multi_root_chain_for_crl(self) -> Iterable[Iterable[Path]]:
        yield from self.single_root_chain_for_crl()
        yield self.single_root_ca(1)

    def normal_cert_chain_for_crl(self, chains: int = 1) -> Iterable[Iterable[Path]]:
        yield from (self.normal_chain(num) for num in range(1, chains + 1))

    def more_than_one_root_ca(self) -> Iterable[Path]:
        yield self.root_ca(1)
        yield self.inter_ca(1, "01")
        yield self.root_ca(2)

    def single_inter_ca(self) -> Iterable[Path]:
        yield self.inter_ca(1, "01")

    def chain_with_unsafe_ca(self) -> Iterable[Path]:
        yield self.root_ca(1)
        yield from (self.inter_ca(1, f"0{num}") for num in range(1, 6))

    def normal_chain(self, rt_num: int = 1) -> Iterable[Path]:
        yield self.root_ca(rt_num)
        yield from (self.inter_ca(rt_num, f"0{num}") for num in range(1, 5))

    def normal_crl_chain(self, rt_num: int = 1) -> Iterable[Path]:
        yield self.root_crl(rt_num)
        yield from (self.inter_crl(rt_num, f"0{num}") for num in range(1, 5))

    def break_chain(self) -> Iterable[Path]:
        yield self.root_ca(1)
        yield from (self.inter_ca(1, num) for num in ("01", "03", "04"))

    def duplicate_ca(self) -> Iterable[Path]:
        yield self.root_ca(1)
        yield from itertools.repeat(self.inter_ca(1, "01"), CertChainParser.MAX_CHAIN_NUMS)

    def duplicate_crl(self) -> Iterable[Path]:
        yield self.root_crl(1)
        yield from itertools.repeat(self.inter_crl(1, "01"), CertChainParser.MAX_CHAIN_NUMS)

    def chain_num_ge_limit(self) -> Iterable[Path]:
        yield from self.normal_chain(1)
        yield from self.normal_chain(2)
        yield self.inter_ca(2, "05")

    def crl_chain_num_ge_limit(self) -> Iterable[Path]:
        yield from self.normal_crl_chain(1)
        yield from self.normal_crl_chain(2)
        yield self.inter_crl(2, "05")

    def chain_1_with_rt_2_inter_ca(self) -> Iterable[Path]:
        yield from self.normal_chain(1)
        yield self.inter_ca(2, "02")

    def shuffle_chain(self) -> Iterable[Path]:
        certs = list(self.normal_chain(1))
        random.shuffle(certs)
        yield from certs

    def shuffle_crl_chain(self) -> Iterable[Path]:
        crl_chain = list(self.normal_crl_chain(1))
        random.shuffle(crl_chain)
        yield from crl_chain

    def cert_chain_with_crl(self) -> Iterable[Path]:
        yield from self.normal_chain(1)
        yield self.root_crl(1)

    def serial(self) -> Iterable[Path]:
        yield self.root(1).joinpath("serial.txt")


def chain_contents(files: Iterable[Path]) -> str:
    return "".join(f.read_text() for f in files)


cert_p = CertPath()
