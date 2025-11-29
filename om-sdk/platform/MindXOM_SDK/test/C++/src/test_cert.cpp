#include <stdio.h>
#include <stdlib.h>
#include <libgen.h>
#include <cstdio>
#include <iostream>
#include "securec.h"
#include "test_cert.h"

#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif
#endif /* __cplusplus */

#include "certproc.h"

#ifdef __cplusplus
#if __cplusplus
}
#endif
#endif /* __cplusplus */

using namespace testing;
using namespace std;

namespace CERT_TEST {
    TEST(CertTest, test_cert_perdiodical_will_return_failed_when_params_not_empty)
    {
        /* certPeriodicalCheck */
        char const *certinfo_file = "xxx.hpm";
        int dd = 10;
        int * day = &dd;
        std::cout << "dt test test_cert_perdiodical_check start: " << certinfo_file << day;
        int ret = certPeriodicalCheck(certinfo_file, day);
        std::cout << "dt test test_cert_perdiodical_check end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(CertTest, test_cert_perdiodical_will_return_failed_when_day_is_empty)
    {
        /* certPeriodicalCheck */
        char const *certinfo_file = "xxx.crt";
        int *day = new int();
        std::cout << "dt test test_cert_perdiodical_check2 start: " << certinfo_file << " " << day;
        int ret = certPeriodicalCheck(certinfo_file, day);
        std::cout << "dt test test_cert_perdiodical_check2 end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(CertTest, test_cert_perdiodical_will_return_failed_when_certinfo_file_is_crt)
    {
        /* certPeriodicalCheck */
        char const *certinfo_file = "xxx.crt";
        int dd = 10;
        int * day = &dd;
        std::cout << "dt test test_cert_perdiodical_check3 start: " << certinfo_file << " " << day;
        int ret = certPeriodicalCheck(certinfo_file, day);
        std::cout << "dt test test_cert_perdiodical_check3 end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(CertTest, test_get_cert_info_will_return_failed_when_params_are_empty)
    {
        /* cert_getcertinfo */
        const char *certinfo_file = "xxx.crt";
        char subj[] = "import cert";
        unsigned long subj_len = 123456789;
        char issuer[] = "issues";
        unsigned long issuer_len = 987223;

        char notbegin[] = "";
        int notbegin_len;
        char notafter[] = "";
        int notafter_len;

        char serialnum[] = "";
        unsigned long serialnum_len;

        std::cout << "dt test test_get_cert_info start: ";
        int ret = cert_getcertinfo(certinfo_file, subj, subj_len, issuer, issuer_len, notbegin, notbegin_len, notafter, notafter_len, serialnum, serialnum_len);
        std::cout << "dt test test_get_cert_info end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }

    TEST(CertTest, test_cert_manage_init_will_return_failed)
    {
        /* cert_manage_init */
        int res = cert_manage_init("/home/data/config/default/om_cert.keystore",
                                   "/home/data/config/default/om_cert_backup.keystore",
                                   "/home/data/config/default/om_alg.json");
        std::cout << "dt test_cert_manage_init1 res: " << res;
        EXPECT_EQ(-1, res);
    }
}