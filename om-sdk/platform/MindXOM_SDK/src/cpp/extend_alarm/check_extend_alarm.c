/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 * Description: Module operations.
 * Create 2023-08-14
 */

#include <sys/ioctl.h>
#include <time.h>
#include <linux/mmc/ioctl.h>
#include "fcntl.h"
#include "stdio.h"
#include "stdlib.h"
#include "string.h"
#include "stdint.h"
#include "securec.h"
#include "unistd.h"
#include "extend_alarm.h"
#include "file_checker.h"
#include "check_extend_alarm.h"

/* 存储设备名称 */
char g_store_dev_name[FAULT_STORE_BUTT][32] = {
    "",
    "",
    "",
    "",
    "",
    "",
    "",
    "",
};
char g_store_dev_block[FAULT_STORE_BUTT] = {0};
char g_store_dev_persent[FAULT_STORE_BUTT] = {0};
int g_store_dev_block_state[FAULT_STORE_BUTT] = {0};
char USB_HUB_ID[CONFIG_FIELD_LENGTH] = {0};
int g_fault_minid_present_flag = MINID_EXIST;
DCMI_INF_STRUCT g_dcmi_infs;
unsigned int g_fault_minid_health = 0;
int g_fault_minid_err_num = 0;
unsigned int g_fault_minid_errcode[FAULT_MINID_ERR_CODE_NUM] = {0};
static unsigned int g_fault_minid_errcode_cfg[FAULT_MINID_BUTT] = {
    0x80E78000,
    0x80E78008,
    0x80FA4E00,
    0x80C78008,
    0x8C0E4E00,
    0x8C0A4E00,
    0x8C104E00,
    0x8C0C4E00,
    0x8C044E00,
    0x8C064E00,
    0x8C03A000,
    0x80E20207,
    0x80E24E00,
    0x80E3A201,
    0x80E3A202,
    0x80E3A203,
    0x80E39200,
    0x80E2120D,
    0x80C98008,
    0x80C98002,
    0x80C98009,
    0x813B8008,
    0x813B8002,
    0x81578008,
    0x81578002,
    0x80A18008,
    0x80DE1801,
    0x80F18008,
    0x80F18000,
    0x80DF8000,
    0x80DF8009,
    0x80DF8008,
    0x80DE0207,
    0x81478008,
    0x81478002,
    0x80E58E03,
    0x80E58E02,
    0x80F78009,
    0x80F78008,
    0x80D38009,
    0x80D58000,
    0x80D58009,
    0x80CD8008,
    0x80CD8003,
    0x80A38008,
    0x80A58008,
    0x80BD8008,
    0x80BD8000,
    0x815F8008,
    0x815F8002,
    0x80CF8008,
    0x80CF8009,
    0x80D98008,
    0x80DB800A,
    0x80DB8000,
    0x80DD8000,
    0x80DD8003,
    0x80DD8008
};

static int parse_hard_disk_age_info(unsigned short *status);

// 硬盘在位检查
int get_disk_persent(unsigned short fault_id, unsigned short sub_id, unsigned short *status)
{
    (void)sub_id;

    if (status == NULL) {
        EXTEND_ALARM_LOG_ERR("status is null");
        return EXTEND_ALARM_ERROR;
    }
    if (strlen(g_store_dev_name[fault_id]) == 0) {
        /* 不在位,且没有告警 */
        if (g_store_dev_persent[fault_id] == 1) {
            /* 之前有在位过，现在不在位了，告警 */
            *status = FAULT_STATUS_ERR;
        } else {
            *status = FAULT_STATUS_OK;
        }
    } else {
        /* 在位 */
        if (g_store_dev_persent[fault_id] != 1) {
            /* 置在位标识 */
            g_store_dev_persent[fault_id] = 1;
        }
        *status = FAULT_STATUS_OK;
    }
    return SUCCESS;
}

// 硬盘温度检查
int get_hard_disk_temp(unsigned short fault_id, unsigned short sub_id, unsigned short *status)
{
    FILE *temp_fd = NULL;
    char file_buff[32] = {0};
    size_t result_len;
    char cmd_temp[] = "smartctl -l scttemp %s | grep 'Current Temperature' | grep Celsius | awk '{print $3}' > /run/hard_disk_temp_info";
    char cmd_info[CMD_SIZE] = {0};
    char *endptr;
    long int temp_curr;

    (void)sub_id;

    if (status == NULL) {
        EXTEND_ALARM_LOG_ERR("status is null");
        return EXTEND_ALARM_ERROR;
    }

    if (fault_id >= FAULT_STORE_BUTT) {
        EXTEND_ALARM_LOG_ERR("get hard disk temperature failed, fault_id(%d)", fault_id);
        return EXTEND_ALARM_ERROR;
    }

    if (strlen(g_store_dev_name[fault_id]) == 0) {
        *status = FAULT_STATUS_OK;
        return SUCCESS;
    }

    if (sprintf_s(cmd_info, CMD_SIZE, cmd_temp, g_store_dev_name[fault_id]) < 0) {
        EXTEND_ALARM_LOG_ERR("get hard disk temperature failed, call sprintf_s cmd_info failed!");
        return EXTEND_ALARM_ERROR;
    }

    if (check_file_path_valid("/run/hard_disk_temp_info") != EDGE_OK) {
        EXTEND_ALARM_LOG_ERR("hard_disk_temp_info path is not valid");
        return EXTEND_ALARM_ERROR;
    }

    (void)system(cmd_info);
    temp_fd = safety_fopen("/run/hard_disk_temp_info", "r");
    if (temp_fd == NULL) {
        EXTEND_ALARM_LOG_ERR("safety_fopen file failed, fd is null!");
        return EXTEND_ALARM_ERROR;
    }
    (void)fgets(file_buff, sizeof(file_buff) - 1, temp_fd);
    (void)fclose(temp_fd);
    (void)unlink("/run/hard_disk_temp_info");

    result_len = strlen(file_buff);
    if (result_len == 0) {
        return EXTEND_ALARM_ERROR;
    }

    temp_curr = strtol(file_buff, &endptr, DECIMAL);
    if (temp_curr > FAULT_HARD_DISK_UP_TEMP_THERSH) {
        *status = FAULT_STATUS_ERR;
    } else {
        *status = FAULT_STATUS_OK;
    }
    return SUCCESS;
}

// 获取硬盘年限
int get_hard_disk_age(unsigned short fault_id, unsigned short sub_id, unsigned short *status)
{
    char cmd_temp[] = "smartctl -H %s > /run/hard_disk_age_info";
    char cmd_info[256] = {0};
    int ret = 0;

    (void)sub_id;

    if (status == NULL) {
        EXTEND_ALARM_LOG_ERR("status is null");
        return EXTEND_ALARM_ERROR;
    }

    if (fault_id >= FAULT_STORE_BUTT) {
        EXTEND_ALARM_LOG_ERR("get hard disk age failed, fault_id(%d)!", fault_id);
        return EXTEND_ALARM_ERROR;
    }

    if (strlen(g_store_dev_name[fault_id]) == 0) {
        *status = FAULT_STATUS_OK;
        return SUCCESS;
    }

    ret = sprintf_s(cmd_info, sizeof(cmd_info), cmd_temp, g_store_dev_name[fault_id]);
    if (ret < 0) {
        EXTEND_ALARM_LOG_ERR("get hard disk age failed, cause by call sprintf_s cmd_info failed!");
        return EXTEND_ALARM_ERROR;
    }

    if (check_file_path_valid("/run/hard_disk_age_info") != SUCCESS) {
        EXTEND_ALARM_LOG_ERR("hard_disk_age_info path is not valid");
        return EXTEND_ALARM_ERROR;
    }

    (void)system(cmd_info);
    ret = parse_hard_disk_age_info(status);
    (void)unlink("/run/hard_disk_age_info");

    return ret;
}

static int parse_hard_disk_age_info(unsigned short *status)
{
    FILE *fp = NULL;
    char line_info[LINE_INFO_SIZE] = {0};
    char last_line[LINE_INFO_SIZE] = {0};
    int line_count = 0;

    fp = safety_fopen("/run/hard_disk_age_info", "r");
    if (fp == NULL) {
        EXTEND_ALARM_LOG_ERR("get hard disk age failed, safety_fopen file failed!");
        return EXTEND_ALARM_ERROR;
    }

    while (fgets(line_info, sizeof(line_info) - 1, fp) != NULL) {
        if (line_count > MAX_LINE_COUNT) {
            EXTEND_ALARM_LOG_ERR("/run/hard_disk_age_info is too large.");
            (void)fclose(fp);
            return EXTEND_ALARM_ERROR;
        }
        if (strstr(line_info, "self-assessment test result") || strstr(line_info, "PASSED")) {
            *status = FAULT_STATUS_OK;
            (void)fclose(fp);
            return SUCCESS;
        }
        if (strcmp(line_info, "") != 0 && !strstr(line_info, "Spin_Retry_Count")) {
            if (strcpy_s(last_line, LINE_INFO_SIZE, line_info) != 0) {
                EXTEND_ALARM_LOG_ERR("get hard disk age failed, cause by strcpy_s failed.");
                (void)fclose(fp);
                return EXTEND_ALARM_ERROR;
            }
        }
        line_count++;
    }
    if (line_count == 0) {
        *status = FAULT_STATUS_OK;
    }
    (void)fclose(fp);

    if (strstr(last_line, "ATTRIBUTE_NAME") != NULL) {
        *status = FAULT_STATUS_OK;
    } else {
        *status = FAULT_STATUS_ERR;
    }

    return SUCCESS;
}

/* usb hua告警 */
int get_usb_hub_alarm(unsigned short fault_id, unsigned short sub_id, unsigned short *status)
{
    int ret;
    FILE *temp_fd = NULL;
    char file_buff[32] = {0};
    char usb_hub_info[256] = {0};

    (void)sub_id;
    (void)fault_id;

    if (status == NULL) {
        return EXTEND_ALARM_ERROR;
    }

    if (strlen(USB_HUB_ID) == 0) {
        return SUCCESS;
    }

    ret = sprintf_s(usb_hub_info, CMD_SIZE, "lsusb |grep %s |awk \'{print $6}\' >/run/usb_hub_info_temp", USB_HUB_ID);
    if (ret < 0) {
        EXTEND_ALARM_LOG_ERR("call sprintf_s fail.");
        return EXTEND_ALARM_ERROR;
    }

    if (check_file_path_valid("/run/usb_hub_info_temp") != SUCCESS) {
        EXTEND_ALARM_LOG_ERR("usb_hub_info_temp path is not valid");
        return EXTEND_ALARM_ERROR;
    }

    (void)system(usb_hub_info);
    temp_fd = safety_fopen("/run/usb_hub_info_temp", "r");
    if (temp_fd == NULL) {
        EXTEND_ALARM_LOG_ERR("safety_fopen /run/usb_hub_info_temp file failed, fd is null!");
        return EXTEND_ALARM_ERROR;
    }

    (void)fread(file_buff, 1, sizeof(file_buff) - 1, temp_fd);
    (void)fclose(temp_fd);
    (void)unlink("/run/usb_hub_info_temp");
    if (strlen(file_buff) == 0) {
        *status = FAULT_STATUS_ERR;
        return SUCCESS;
    }

    *status = (strncmp(file_buff, USB_HUB_ID, strlen(USB_HUB_ID)) != 0) ? FAULT_STATUS_ERR : FAULT_STATUS_OK;
    return SUCCESS;
}

// 获取硬盘阻塞状态
int get_hard_disk_block(unsigned short fault_id, unsigned short sub_id, unsigned short *status)
{
    (void)sub_id;

    if (status == NULL) {
        EXTEND_ALARM_LOG_ERR("Input args is NULL.");
        return EXTEND_ALARM_ERROR;
    }

    if (fault_id >= FAULT_STORE_BUTT) {
        EXTEND_ALARM_LOG_ERR("get hard disk block status failed, fault_id(%d)!", fault_id);
        return EXTEND_ALARM_ERROR;
    }

    if (strlen(g_store_dev_name[fault_id]) == 0) {
        *status = FAULT_STATUS_OK;
        return SUCCESS;
    }

    if (g_store_dev_block[fault_id] == FAULT_HARD_DISK_STATUS_BLOCK) {
        *status = FAULT_STATUS_ERR;
    } else {
        *status = FAULT_STATUS_OK;
    }
    return SUCCESS;
}

int read_extcsd(int fd, unsigned char *ext_csd, int len)
{
    int ret;
    struct mmc_ioc_cmd idata;

    if (ext_csd == NULL) {
        EXTEND_ALARM_LOG_ERR("Input ext_csd is NULL.");
        return EXTEND_ALARM_ERROR;
    }

    memset_s(&idata, sizeof(idata), 0, sizeof(idata));
    memset_s(ext_csd, len, 0, len);
    idata.write_flag = 0;
    idata.opcode = MMC_SEND_EXT_CSD;
    idata.arg = 0;
    idata.flags = MMC_RSP_SPI_R1 | MMC_RSP_R1 | MMC_CMD_ADTC;
    idata.blksz = EXT_CSD_REG_NUM;
    idata.blocks = 1;
    mmc_ioc_cmd_set_data(idata, (uintptr_t)ext_csd);

    ret = ioctl(fd, MMC_IOC_CMD, &idata);
    if (ret) {
        EXTEND_ALARM_LOG_ERR("ioctl failed ret(%d)!", ret);
    }

    return ret;
}

static int count = 0;

// 校验emmc生命周期
int check_extcsd_info(unsigned short fault_id, unsigned short sub_id, unsigned short *status)
{
    int fd = 0;
    int ret;
    char device[] = "/dev/mmcblk0";
    unsigned char ext_csd[EXT_CSD_REG_NUM];
    (void)sub_id;
    (void)fault_id;
    count++;

    if (status == NULL) {
        return EXTEND_ALARM_ERROR;
    }

    if (access(device, F_OK) == -1) {
        *status = FAULT_STATUS_OK;
        return SUCCESS;
    }

    if (check_file_path_valid(device) != 0) {
        EXTEND_ALARM_LOG_ERR("check file /dev/mmcblk0 failed, file is invalid");
        return EXTEND_ALARM_ERROR;
    }

    fd = open(device, 0);
    if (fd < 0) {
        EXTEND_ALARM_LOG_ERR("open device(%s) failed!", device);
        return EXTEND_ALARM_ERROR;
    }

    ret = read_extcsd(fd, ext_csd, EXT_CSD_REG_NUM);
    if (ret) {
        EXTEND_ALARM_LOG_ERR("read extcsd info from device(%s) failed!", device);
        close(fd);
        return EXTEND_ALARM_ERROR;
    }

    if ((ext_csd[EXT_CSD_DEVICE_LIFE_TIME_EST_TYP_A] > MAX_PE_NUM) ||
        (ext_csd[EXT_CSD_DEVICE_LIFE_TIME_EST_TYP_B] > MAX_PE_NUM)) {
        *status = FAULT_STATUS_ERR;
    } else {
        *status = FAULT_STATUS_OK;
    }

    close(fd);
    return SUCCESS;
}

// 获取minid告警信息
int fault_get_minid_info(void)
{
    int ret;
    int nums = 0;
    unsigned int health = 0;
    static unsigned int err_num = 0;

    ret = g_dcmi_infs.pfn_get_device_num_in_card(0, &nums);
    if (ret < 0) {
        if ((err_num % 1000) == 0) { // 获取Atlas 200中芯片个数失败1000次记录一次日志
            EXTEND_ALARM_LOG_ERR("init minid error. ret = %d.", ret);
            err_num = 0;
        }
        err_num++;
        return EXTEND_ALARM_ERROR;
    }

    if (nums < 1) {
        EXTEND_ALARM_LOG_ERR("not found minid error. nums = %d.", nums);
        return EXTEND_ALARM_ERROR;
    }
    ret = g_dcmi_infs.pfn_get_device_health(0, 0, &health);
    if (ret < 0) {
        EXTEND_ALARM_LOG_ERR("get device health error. ret = %d.", ret);
        return EXTEND_ALARM_ERROR;
    }

    g_fault_minid_health = health;

    if (health == 0) {
        return SUCCESS;
    }
    ret = g_dcmi_infs.pfn_get_device_errorcode(0, 0, &g_fault_minid_err_num, g_fault_minid_errcode,
        FAULT_MINID_ERR_CODE_NUM);
    if ((ret < 0) || (g_fault_minid_err_num > FAULT_MINID_ERR_CODE_NUM) || g_fault_minid_err_num < 0) {
        EXTEND_ALARM_LOG_ERR("get errorcode error. err_num=%d, ret = %d.", g_fault_minid_err_num, ret);
        g_fault_minid_err_num = 0;
        return EXTEND_ALARM_ERROR;
    }

    return SUCCESS;
}

// 更新minid告警状态
int get_minid_alarm(unsigned short fault_id, unsigned short sub_id, unsigned short *status)
{
    unsigned int index;
    unsigned int err_code;
    unsigned int err_code_curr;
    unsigned int has_find = 0;
    (void)sub_id;
    (void)fault_id;
    if (status == NULL) {
        EXTEND_ALARM_LOG_ERR("Input args is NULL.");
        return EXTEND_ALARM_ERROR;
    }

    // minid不存在时，不上报告警，不检测
    if (g_fault_minid_present_flag != MINID_EXIST) {
        *status = 0;
        return SUCCESS;
    }

    if (sub_id == 0) {
        (void)fault_get_minid_info();
    }

    if (g_fault_minid_health == 0) {
        *status = 0;
        return SUCCESS;
    }

    err_code = g_fault_minid_errcode_cfg[sub_id];
    for (index = 0; index < (unsigned int)g_fault_minid_err_num; index++) {
        err_code_curr = g_fault_minid_errcode[index];
        if (err_code == err_code_curr) {
            has_find = 1;
            break;
        }
    }

    if (has_find == 1) {
        *status = 1;
    } else {
        *status = 0;
    }

    return SUCCESS;
}

int check_minid_temperature(unsigned short fault_id, unsigned short sub_id, unsigned short *status)
{
    int temp = 0;
    static unsigned short last_status = 0;
    int ret;
    (void)sub_id;
    (void)fault_id;
    // minid在位检测
    if (g_fault_minid_present_flag != MINID_EXIST) {
        return EXTEND_ALARM_ERROR;
    }

    ret = g_dcmi_infs.pfn_get_device_temperature(0, 0, &temp);
    if (ret < 0) {
        EXTEND_ALARM_LOG_ERR("get minid's temperature error, ret=%d.", ret);
        return EXTEND_ALARM_ERROR;
    }

    if (temp > FAULT_MINID_TEMP_THERSH) { // 高于95告警
        *status = 1;
    } else {
        if ((last_status == 1) && (temp > FAULT_MINID_TEMP_RECOVER_THERSH) &&
            (temp <= FAULT_MINID_TEMP_THERSH)) { // 如果在90-95度区间且last_status是告警状态，保持不变，90度下告警恢复
            *status = 1;
        } else {
            *status = 0;
        }
    }

    if (*status != last_status) {
        EXTEND_ALARM_LOG_INFO("record hilens minid temperature, last status(%d) to now status(%d) , temperature = %d.",
            last_status, *status, temp);
        last_status = *status;
    }
    return SUCCESS;
}