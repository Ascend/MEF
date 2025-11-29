# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
import re


class NetCheck:
    @staticmethod
    def get_addr_bin(addr, is_ipv6=False):
        """
            功能描述：地址转换，把点分十进制表示的IP地址字符串转换为二进制形式的字符串
            参数:
                addr 字符串类型，点分十进制形式的IP地址
                isIPv6  true，表示处理IPv6地址； 默认为 false，表示处理IPv4地址
            返回值：字符串类型，二进制形式的IP地址字
            异常描述：NA
            修改记录：2016-12-24 创建
                2017-01-03 修改：可信任入参，去掉不必要的判断
                2017-01-06 修改： 支持 IPv6
        """
        addr_bin_str = ""
        # IPv6 冒分十六进制表示法由冒号分割成 8 段， 每段由 4 个 '0-F' 字符表示
        ipv6_len = 8
        ipv6_sub = 4

        if is_ipv6:
            # 以冒号截取 IPv6 地址
            tmp_list = addr.split(":")
            # 计算由 "::" 省略掉的段
            num = ipv6_len - len(tmp_list)

            for ip_num in tmp_list:
                if not ip_num:
                    # 把省略掉的段和当前段以 16 个零（即："0000" * IPv6sub）补齐到二进制字符串
                    addr_bin_str += "0000" * ipv6_sub * (num + 1)
                    continue

                # 把十六进制字符串格式化为二进制字符串，并追加到地址字符串
                addr_bin_str += '{0:016b}'.format(int(ip_num, 16))
            return addr_bin_str

        # 处理 IPv4 地址
        addr_list = addr.split(".")
        for addr in addr_list:
            # 把十进制数据格式化为8位二进制数据
            addr_bin_str += '{0:08b}'.format(int(addr))

        return addr_bin_str

    @staticmethod
    def get_num_for_mask(mask_str):
        """
            功能描述：子网掩码转换，把 IP 地址后的数字解析成子网掩码
            参数：maskStr 字符串类型，表示点分十进制的子网掩码
            返回值： 子网掩码位数，字符串类型
            异常描述：NA
            修改记录：2017-02-17 创建
        """
        # 校验入参
        if not isinstance(mask_str, str):
            return ""

        # 得到二进制形式的子网掩码
        mask_bin = NetCheck.get_addr_bin(mask_str)

        # 子网掩码由1和0组成，且1和0分别连续，0的位置即为掩码中1的位数
        ret = mask_bin.find("0")
        if 0 < ret < 32:
            return str(ret)
        if ret == -1:
            return '32'
        else:
            return ""

    @staticmethod
    def is_same_network_segment(addr1, addr2, num_mask_str, is_ipv6=False):
        """
            功能描述：判断IP地址是否在同一网段
            参数：
                addr1  第一个IP地址字符串，值类似："192.168.12.10"
                addr2  第二个IP地址 字符串
                numMaskStr  字符串形式的数值，表示子网掩码
                isIPv6  true，表示处理IPv6地址； 默认为 false，表示处理IPv4地址
            返回值：在同一网段返回 True ； 否则，返回 False
            异常描述：NA
            修改记录：2016-12-02 创建
                2016-01-06 修改： 支持 IPv6 地址判断
        """
        # 确定二进制地址长度
        if is_ipv6:
            addr_bin_len = 128
        else:
            addr_bin_len = 32

        # 判断 子网掩码字符串 是否合法
        if num_mask_str == "":
            return False

        if str.isdigit(num_mask_str):
            num_mask = int(num_mask_str)
        else:
            return False

        if num_mask < 0 or num_mask > addr_bin_len:
            return False

        # 得到二进制形式的地址字符串
        addr1_str = NetCheck.get_addr_bin(addr1, is_ipv6)
        addr2_str = NetCheck.get_addr_bin(addr2, is_ipv6)

        # 根据子网掩码截取地址字符串，比较是否是同一网段
        if addr1_str[:num_mask] == addr2_str[:num_mask]:
            return True
        else:
            return False

    @staticmethod
    def net_work_address(addr1, addr2):
        ip_list = addr1.split(".")
        smart_list = addr2.split(".")
        address_list = []
        for i, _ in enumerate(ip_list):
            p = int(ip_list[i]) & int(smart_list[i])
            address_list.append(p)
        address = ".".join(str(s) for s in address_list)

        return address

    @staticmethod
    def mac_addr_single_cast_check(mac_addr):
        pattern = "^([A-Fa-f0-9]{2}[:]){5}[A-Fa-f0-9]{2}$"
        if not re.fullmatch(pattern, mac_addr):
            return False
        if mac_addr.upper() in ['00:00:00:00:00:00', 'FF:FF:FF:FF:FF:FF']:
            return False
        bit40 = mac_addr.split(':')[0]
        try:
            bit40 = int(bit40, 16) % 2
        except Exception:
            return False
        return bit40 == 0
