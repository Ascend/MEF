// Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <errno.h>
#include <time.h>
#include <sys/ioctl.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/inotify.h>
#include <sys/prctl.h>
#include <linux/limits.h>
#include <linux/rtc.h>
#include "securec.h"
#include "test_alarm_process.h"

#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif
#endif /* __cplusplus */

#include "alarm_process.h"
#include "file_checker.h"
#include "ens_api.h"

#ifdef __cplusplus
#if __cplusplus
}
#endif
#endif /* __cplusplus */

using namespace testing;
using namespace std;

namespace ALARM_PROCESS_TEST {
    TEST(AlarmProcessTest, test_fault_record_to_file_param_invalid1)
    {
        /* fault_record_to_file */
        unsigned int *activeNum;
        int *alarmLevel = NULL;
        std::cout << "dt test_fault_record_to_file_param_invalid1 start: " << activeNum << alarmLevel;
        int ret = fault_record_to_file(activeNum, alarmLevel);
        std::cout << "dt test_fault_record_to_file_param_invalid1 end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(AlarmProcessTest, test_fault_record_to_file_param_invalid2)
    {
        /* fault_record_to_file */
        unsigned int *activeNum = NULL;
        int *alarmLevel;
        std::cout << "dt test_fault_record_to_file_param_invalid2 start: " << activeNum << alarmLevel;
        int ret = fault_record_to_file(activeNum, alarmLevel);
        std::cout << "dt test_fault_record_to_file_param_invalid2 end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(AlarmProcessTest, test_search_item_by_info_param_null)
    {
        /* search_item_by_info */
        const ALARM_MSG_FAULT_INFO *curFaultInfo = NULL;
        int isShield = 1;
        std::cout << "dt test_search_item_by_info_param_null start: " << curFaultInfo << isShield;
        ALARM_LV1_STRU *ret = search_item_by_info(curFaultInfo, isShield);
        std::cout << "dt test_search_item_by_info_param_null end:" << std::endl;
        EXPECT_EQ(NULL, ret);
    }

    TEST(AlarmProcessTest, test_search_item_by_info_fault_id_wrong)
    {
        /* search_item_by_info */
        ALARM_MSG_FAULT_INFO tempFaultInfo = {100, 1, 1, 1, 1011111111111111, "TEST", "TEST1"};
        const ALARM_MSG_FAULT_INFO *curFaultInfo = &tempFaultInfo;
        int isShield = 1;
        std::cout << "dt test_search_item_by_info_fault_id_wrong start: " << isShield;
        ALARM_LV1_STRU *ret = search_item_by_info(curFaultInfo, isShield);
        std::cout << "dt test_search_item_by_info_fault_id_wrong end:" << std::endl;
        EXPECT_EQ(NULL, ret);
    }

    TEST(AlarmProcessTest, test_alarm_report_param_null)
    {
        /* alarm_report */
        const unsigned char *info = NULL;
        std::cout << "dt test_alarm_report_param_null start: ";
        int ret = alarm_report(info);
        std::cout << "dt test_alarm_report_param_null end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(AlarmProcessTest, test_alarm_report_len_wrong)
    {
        /* alarm_report */
        ALARM_MSG_INFO_HEAD msg = {100, 100, 100};
        const unsigned char *info = (const unsigned char *)&msg;
        std::cout << "dt test_alarm_report_len_wrong start: " << info;
        int ret = alarm_report(info);
        std::cout << "dt test_alarm_report_len_wrong end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(AlarmProcessTest, test_alarm_process_load)
    {
        /* alarm_process_load */
        std::cout << "dt test_alarm_process_load start: ";
        AMOCKER(ens_intf_export).will(returnValue(NULL));
        int ret = alarm_process_load();
        std::cout << "dt test_alarm_process_load end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(AlarmProcessTest, test_alarm_process_unload)
    {
        /* alarm_process_unload */
        std::cout << "dt test_alarm_process_unload start: ";
        int ret = alarm_process_unload();
        std::cout << "dt test_alarm_process_unload end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }

    TEST(AlarmProcessTest, test_alarm_process_start_create_thread_failed)
    {
        /* alarm_process_start */
        std::cout << "dt test_alarm_process_start_create_thread_failed start: ";
        AMOCKER(set_clear_shield_alarm_flag).will(returnValue(NULL));
        AMOCKER(pthread_create).will(returnValue(-1));
        int ret = alarm_process_start();
        std::cout << "dt test_alarm_process_start_create_thread_failed end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(AlarmProcessTest, test_alarm_process_start_success)
    {
        /* alarm_process_start */
        std::cout << "dt test_alarm_process_start_success start: ";
        AMOCKER(set_clear_shield_alarm_flag).will(returnValue(NULL));
        AMOCKER(pthread_create).will(returnValue(0));
        int ret = alarm_process_start();
        std::cout << "dt test_alarm_process_start_success end:" << ret << std::endl;
        EXPECT_EQ(0, ret);
    }
}