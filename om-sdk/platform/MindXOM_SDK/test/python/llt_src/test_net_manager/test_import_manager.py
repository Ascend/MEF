#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
from datetime import datetime, timedelta
from pathlib import Path
from typing import NamedTuple, Literal, Iterable

import pytest
from pytest_mock import MockerFixture

from common.constants.error_codes import SecurityServiceErrorCodes
from common.db.database import DataBase
from net_manager.constants import NetManagerConstants
from net_manager.exception import DataCheckException
from net_manager.manager.import_manager import ImportCert, ImportCrl
from net_manager.manager.net_cfg_manager import NetCfgManager
from net_manager.models import CertManager, CertInfo
from test_net_manager.conftest import cert_p, chain_contents


class Necessary(NamedTuple):
    expect: str = ""
    source: Literal["Web", "FusionDirector"] = NetManagerConstants.WEB
    status: str = ""


class ImportDeal(NamedTuple):
    chain_num: int = 1
    err_msg: str = ""
    time_now: datetime = datetime.utcnow() + timedelta(days=1)
    certs: Iterable[Path] = []

    def cert_contents(self):
        return chain_contents(self.certs)


class CrlImportDeal(NamedTuple):
    crl_it: Iterable[Path] = []
    cert_num: int = 1
    err_msg: str = ""
    time_now: datetime = datetime.utcnow() + timedelta(days=1)
    cert_chains: Iterable[Iterable[Path]] = []

    def crl_contents(self):
        return chain_contents(self.crl_it)

    def cert_chain_contents(self) -> Iterable[str]:
        yield from (chain_contents(certs) for certs in self.cert_chains)


class TestImportCert:
    use_cases = {
        "test_necessary": {
            "web necessary": Necessary(),
            "web not necessary": Necessary("not allowed", status="ready"),
            "Fd necessary": Necessary(source=NetManagerConstants.FUSION_DIRECTOR),
            "Fd necessary ready": Necessary(source=NetManagerConstants.FUSION_DIRECTOR, status="ready")
        },
        "test_import_deal": {
            "import single root ca should be success":
                ImportDeal(certs=cert_p.single_root_ca()),
            "import ec 256 should be success":
                ImportDeal(certs=cert_p.ec_256_ca(1)),
            "import ec 128 should be error":
                ImportDeal(certs=cert_p.ec_128_ca(1),
                           err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_EC_LEN_INVALID.messageKey),
            "import single inter ca should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_CA_SIGNATURE_INVALID.messageKey,
                           certs=cert_p.single_inter_ca()),
            "import ca chain with more than one root should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.messageKey,
                           certs=cert_p.more_than_one_root_ca()),
            "import chain with unsafe should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_RSA_LEN_INVALID.messageKey,
                           certs=cert_p.chain_with_unsafe_ca()),
            "import cert chain should be success":
                ImportDeal(chain_num=5, certs=cert_p.normal_chain()),
            "import break chain should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.messageKey,
                           certs=cert_p.break_chain()),
            "start time le now should be expired":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_HAS_EXPIRED.messageKey,
                           certs=cert_p.single_root_ca(), time_now=datetime.utcnow() + timedelta(days=-10)),
            "end time gt now should be expired":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_HAS_EXPIRED.messageKey,
                           certs=cert_p.single_root_ca(), time_now=datetime.utcnow() + timedelta(days=8300)),
            "duplicate cert should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.messageKey,
                           certs=cert_p.duplicate_ca()),
            "chain num ge limit should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_CHAIN_NUMS_MAX.messageKey,
                           certs=cert_p.chain_num_ge_limit()),
            "inter ca with no root should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.messageKey,
                           certs=cert_p.chain_1_with_rt_2_inter_ca()),
            "shuffled normal chain should be success":
                ImportDeal(chain_num=5, certs=cert_p.shuffle_chain()),
            "cert chain with crl should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.messageKey,
                           certs=cert_p.cert_chain_with_crl()),
            "not cert should be error":
                ImportDeal(err_msg=SecurityServiceErrorCodes.ERROR_CERTIFICATE_IS_INVALID.messageKey,
                           certs=cert_p.serial())
        }
    }

    @staticmethod
    def test_necessary(mocker: MockerFixture, model: Necessary):
        mocker.patch.object(NetCfgManager, "get_net_cfg_info").return_value.status = model.status
        if model.expect:
            with pytest.raises(DataCheckException, match=model.expect):
                ImportCert(model.source)
        else:
            ImportCert(model.source)

    @staticmethod
    def test_import_deal(init_certs_db: DataBase, mocker: MockerFixture, model: ImportDeal):
        mocker.patch("net_manager.checkers.contents_checker.datetime").utcnow.return_value = model.time_now
        if model.err_msg:
            with pytest.raises(Exception, match=model.err_msg):
                ImportCert().import_deal(model.cert_contents())
        else:
            ImportCert().import_deal(model.cert_contents())
            with init_certs_db.session_maker() as session:
                assert model.chain_num == session.query(CertInfo).first().chain_num

    @staticmethod
    def test_import_deal_by_fd(init_certs_db: DataBase):
        """测试FD导入时的重复判断"""
        contents = chain_contents(cert_p.normal_chain(1))
        import_obj = ImportCert("FusionDirector", "any")
        import_obj.import_deal(contents)
        with pytest.raises(Exception, match="finger already existed"):
            import_obj.import_deal(contents)


class TestImportCrl:
    use_cases = {
        "test_import_deal": {
            "import single crl should be success":
                CrlImportDeal(crl_it=cert_p.single_root_crl(1),
                              cert_chains=cert_p.single_root_chain_for_crl()),
            "import crl for multi certs should be success":
                CrlImportDeal(crl_it=cert_p.single_root_crl(1),
                              cert_num=2,
                              cert_chains=cert_p.single_multi_root_chain_for_crl()),
            "import crl not necessary":
                CrlImportDeal(crl_it=cert_p.single_root_crl(1),
                              err_msg="Check certificate is null"),
            "not crl should be error":
                CrlImportDeal(crl_it=cert_p.serial(),
                              err_msg="load verify locations failed",
                              cert_chains=cert_p.single_root_chain_for_crl()),
            "crl and cert concatenation file should be error":
                CrlImportDeal(crl_it=cert_p.cert_chain_with_crl(),
                              err_msg="Certificate and revocation list concatenation file not supported",
                              cert_chains=cert_p.single_root_chain_for_crl()),
            "duplicate crl in chain should be error":
                CrlImportDeal(crl_it=cert_p.duplicate_crl(),
                              err_msg="Duplicate chain in crl list",
                              cert_chains=cert_p.single_root_chain_for_crl()),
            "last update ge now should be error":
                CrlImportDeal(crl_it=cert_p.single_root_crl(1),
                              err_msg="check last update time or next update time is error",
                              time_now=datetime.utcnow() - timedelta(200),
                              cert_chains=cert_p.single_root_chain_for_crl()),
            "next update le now should be error":
                CrlImportDeal(crl_it=cert_p.single_root_crl(1),
                              err_msg="check last update time or next update time is error",
                              time_now=datetime.utcnow() + timedelta(8000),
                              cert_chains=cert_p.single_root_chain_for_crl()),
            "chain num ge limit should be error":
                CrlImportDeal(crl_it=cert_p.crl_chain_num_ge_limit(),
                              err_msg="The number of crl is greater than",
                              cert_chains=cert_p.single_root_chain_for_crl()),
            "crl chain does not match cert should be error":
                CrlImportDeal(crl_it=cert_p.normal_crl_chain(2),
                              err_msg="Verify CRL does not match against the CA certificate.",
                              cert_chains=cert_p.normal_cert_chain_for_crl(1)),
            "crl num not equal cert num should be error":
                CrlImportDeal(crl_it=cert_p.normal_crl_chain(2),
                              err_msg="Verify CRL does not match against the CA certificate.",
                              cert_chains=cert_p.single_root_chain_for_crl()),
            "shuffled normal chain should be success":
                CrlImportDeal(crl_it=cert_p.shuffle_crl_chain(),
                              cert_num=1,
                              cert_chains=cert_p.normal_cert_chain_for_crl(2)),
            "crl chain should be success":
                CrlImportDeal(crl_it=cert_p.normal_crl_chain(1),
                              cert_num=1,
                              cert_chains=cert_p.normal_cert_chain_for_crl(2))
        }
    }

    @staticmethod
    def test_import_deal(mocker: MockerFixture, init_certs_db: DataBase, model: CrlImportDeal):
        mocker.patch("cert_manager.parse_tools.datetime").utcnow.return_value = model.time_now
        for index, contents in enumerate(model.cert_chain_contents()):
            ImportCert(cert_name=str(index)).import_deal(contents)

        if model.err_msg:
            with pytest.raises(Exception, match=model.err_msg):
                ImportCrl().import_deal(model.crl_contents())
        else:
            contents = model.crl_contents()
            ImportCrl().import_deal(contents)
            with init_certs_db.session_maker() as session:
                assert model.cert_num == session.query(CertManager).filter(CertManager.crl_contents == contents).count()
