/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   OMSDK is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */

import i18n from '@/utils/locale';
import {ElMessageBox} from 'element-plus';
import errorCode from '@/api/errorCode';
import {clearSessionStorage} from '@/utils/commonMethods';

export function handleResponseError(response) {
  let code = response?.data?.error?.code ?? '';
  if (code.indexOf('SessionTimeLimitExceeded') > -1) {
    clearSessionStorage();
    ElMessageBox.alert(
      i18n.global.t('errorTip.timeLimitExceeded'),
      i18n.global.t('common.errorReminder'),
      { type: 'error'}
    )
      .then(() => {
        window.location.href = '/login';
      })
  }
  if (code.indexOf('AccountForSessionNoLongerExists') > -1) {
    clearSessionStorage();
    ElMessageBox.alert(
      i18n.global.t('errorTip.accountForSessionNoLongerExists'),
      i18n.global.t('common.errorReminder'),
      { type: 'error' }
    )
      .then(() => {
        window.location.href = '/login';
      })
  }

  let status = response?.data?.error['@Message.ExtendedInfo'][0].Oem.status;
  let message = response?.data?.error['@Message.ExtendedInfo'][0].Message;
  if (status === errorCode.SESSION.ERROR_PASSWORD_EXPIRE || message.indexOf('insecure account') > -1) {
    ElMessageBox.alert(
      i18n.global.t('common.passwordHasExpired'),
      i18n.global.t('common.errorReminder'),
      { type: 'error' }
    )
      .then(() => {
        window.location.href = '/changePassword';
      })
  }
}
