# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
class ErrorCode(object):
    # 系统升级错误
    upgrade_timeout_err_code = 171
    midware_upgrade_fail = 173  # "173": "Upgrade failed for unknown reason."
    # use 600-700 for midware
    midware_common_err = 600
    midware_input_not_json = 601
    midware_input_not_permitted = 602
    midware_input_parameter_invalid = 603
    midware_resource_busy = 604

    # profile
    midware_profile_not_exist = 610

    # ntp 611-620
    midware_config_ntp_common_err = 611

    # dns 626-630
    midware_config_dns_common_err = 626

    # partition 631-639
    midware_partition_dev_not_fount = 631
    midware_partition_is_using = 632
    midware_partition_list_failed = 633
    midware_partition_docker_failed = 634
    midware_partition_out_range = 635
    midware_partition_damaged = 636
    midware_partition_path_not_empty = 637
    midware_partition_fs_not_support = 638

    # firmware 640-649
    midware_firmware_download_err = 640
    midware_firmware_upgrade_err = 641

    # infocollect 650-659
    midware_info_collect_common_err = 650
    midware_info_collect_exec_err = 651
    midware_info_collect_upload_err = 652

    # asset tag 670-675
    midware_assert_tag_common_err = 670

    # password_validity 676-679
    midware_config_passwd_validity_err = 676

    @staticmethod
    def midware_generate_err_msg(err_code, error_info):
        return "ERR." + str(err_code) + ", " + error_info
