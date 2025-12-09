#include <stdio.h>
#include <stdlib.h>
#include <libgen.h>
#include <cstdio>
#include <iostream>
#include "securec.h"
#include "test_cms.h"

#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif
#endif /* __cplusplus */

#include "cms_api.h"

#ifdef __cplusplus
#if __cplusplus
}
#endif
#endif /* __cplusplus */

using namespace testing;
using namespace std;

namespace CMS_TEST {

    TEST(CmsTest, test_prepare_upgrade_image_cms)
    {
        const char *pathname_cms = "xxxx.cms";
        const char *pathname_crl = "xxxx.crl";
        const char *pathname_tar = "xxxx.tar.gz";

        int ret = prepareUpgradeImageCms(pathname_cms, pathname_crl, pathname_tar);
        std::cout << "dt test test_prepare_upgrade_image_cms1 end:" << ret << std::endl;
        EXPECT_EQ(-1, ret);
    }
}