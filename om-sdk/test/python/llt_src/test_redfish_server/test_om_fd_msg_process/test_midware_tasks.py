from collections import namedtuple

from om_fd_msg_process.midware_tasks import get_digital_warranty

from lib_restful_adapter import LibRESTfulAdapter

from pytest_mock import MockerFixture

GetDigitalWarranty = namedtuple("GetDigitalWarranty", "expect, lib_interface")


class TestMidwareTask:
    use_cases = {
        "test_get_digital_warranty": {
            "get_dflc_failed": ({}, {"status": 400, "message": {}}),
            "normal": ({"digital_warranty": {"manufacture_date": 1, "start_point": 1, "life_span": 1}},
                       {"status": 200, "message": {"ManufactureDate": 1, "StartPoint": 1, "LifeSpan": 1}}),
        },
    }

    @staticmethod
    def test_get_digital_warranty(mocker: MockerFixture, model: GetDigitalWarranty):
        mocker.patch.object(LibRESTfulAdapter, "lib_restful_interface", return_value=model.lib_interface)
        ret = get_digital_warranty()
        assert ret == model.expect
