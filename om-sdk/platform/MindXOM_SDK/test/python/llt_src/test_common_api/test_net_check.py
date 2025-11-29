from collections import namedtuple

from pytest_mock import MockerFixture


from common.net_check import NetCheck


class TestNetCheck:
    GetAddrBinCase = namedtuple("GetAddrBinCase", "excepted, addr, is_ipv6")
    GetNumForMaskCase = namedtuple("GetNumForMaskCase", "excepted, mask_str, mask_bin")
    IsSameNetworkSegmentCase = namedtuple("GetNumForMaskCase", "excepted, num_mask_str, is_ipv6, get_addr_bin")
    IpAddrCheckCase = namedtuple("IpAddrCheckCase", "excepted, ip_addr, check")
    MacAddrSingleCastCheckCase = namedtuple("MacAddrSingleCastCheckCase", "excepted, mac_addr")

    use_cases = {
        "test_get_addr_bin": {
            "is_ipv6": ("0000000010101011000000000000000000000000000000000000000000000000000000000000000"
                        "0000000001110111100000000000100000000000000010000", "AB::EF:10:10", True),
            "is_ipv4": ("00110011001001100100010100001011", "51.38.69.11", False)
        },
        "test_get_num_for_mask": {
            "not_str": ("", ["test1"], None),
            "find_zero_in_range": ("1", "test2", "10110011001001100100010100001011"),
            "find_zero_failed": ("32", "test3", "11111"),
            "find_zero_not_in_range": ("", "test4", "011111"),
        },
        "test_is_same_network_segment": {
            "num_mask_str_null": (False, "", True, ["", ""]),
            "num_mask_isdigit_false": (False, "ab", False, ["", ""]),
            "num_mask_less_0": (False, "-1", False, ["", ""]),
            "num_mask_more_32": (False, "33", False, ["", ""]),
            "same": (True, "2", False, ["ab", "abc"]),
            "not_same": (False, "2", False, ["ddd", "abc"]),
        },
        "test_mac_addr_single_cast_check": {
            "not_match": (False, "#$1234"),
            "all_zero": (False, "00:00:00:00:00:00"),
            "normal": (True, "10:00:00:00:00:00"),
            "exception": (False, "PP:00:00:00:00:00")
        }
    }

    def test_get_addr_bin(self, model: GetAddrBinCase):
        ret = NetCheck.get_addr_bin(model.addr, model.is_ipv6)
        assert model.excepted == ret

    def test_get_num_for_mask(self, mocker: MockerFixture, model: GetNumForMaskCase):
        mocker.patch.object(NetCheck, "get_addr_bin", return_value=model.mask_bin)
        ret = NetCheck.get_num_for_mask(model.mask_str)
        assert model.excepted == ret

    def test_is_same_network_segment(self, mocker: MockerFixture, model: IsSameNetworkSegmentCase):
        mocker.patch.object(NetCheck, "get_addr_bin", side_effect=model.get_addr_bin)
        ret = NetCheck.is_same_network_segment("test1", "test2", model.num_mask_str, model.is_ipv6)
        assert model.excepted == ret

    def test_net_work_address(self):
        ret = NetCheck.net_work_address("10.10.10.10", "11.11.11.11")
        assert ret == "10.10.10.10"

    def test_mac_addr_single_cast_check(self, model: MacAddrSingleCastCheckCase):
        ret = NetCheck.mac_addr_single_cast_check(model.mac_addr)
        assert model.excepted == ret
