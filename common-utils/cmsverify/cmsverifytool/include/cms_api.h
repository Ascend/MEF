/*
 * 版权所有 (c) 华为技术有限公司 2024
 * 文 件 名   : cms_api.c
 * 生成日期   : 2024年2月29日
 * 功能描述   : CMS验签和证书吊销比较
 */
#ifndef CMS_API_H
#define CMS_API_H
#include "cmscbb_cms_vrf.h"

unsigned int VerifyCmsFile(char *crlName, char *cmsName, char *fileName);
unsigned int CompareCrls(const char *pszCrlToUpdate, const char *pszCrlOnDevice, CmscbbCrlPeriodStat *stat);

#endif
