<template>
  <div class="tab-title" style="margin-top: 10px;">
    {{ $t('restore.restoreFactorySetting.tabName') }}
  </div>
  <basic-information-container style='height: 70vh' >
    <el-form
      :model='resetForm'
      :rules='resetFormRule'
      ref='resetFormRef'
      label-position='left'
      label-width="auto"
      style='margin-top: 20px; width: 800px;'
      :hide-required-asterisk="true"
      @submit.prevent
    >
      <el-form-item prop='isRetain' :label="$t('restore.restoreFactorySetting.ipAddress')" class="not-required">
        <el-radio-group v-model="resetForm.isRetain">
          <el-radio :label="true" style="margin-right: 100px;">
            {{ $t("restore.restoreFactorySetting.retain") }}
          </el-radio>
          <el-radio :label="false" style="margin-right: 10px;">
            {{ $t("restore.restoreFactorySetting.notRetain") }}
          </el-radio>
        </el-radio-group>
        <div class="reset-tip" style="margin-top: 20px;">
          <img src="@/assets/img/home/information.svg" alt=""/>
          {{ $t('restore.restoreFactorySetting.resetDeleteConfigTip') }}
        </div>
        <div class="reset-tip">
          <img src="@/assets/img/home/information.svg" alt=""/>
          {{ $t('restore.restoreFactorySetting.resetIrrevocableTip') }}
        </div>
        <div class="reset-tip">
          <img src="@/assets/img/home/information.svg" alt=""/>
          {{ $t('restore.restoreFactorySetting.setRetainedIpTip') }}
        </div>
      </el-form-item>
      <el-form-item prop='Password' :label="$t('common.password')" class="required">
        <el-input
          style='width: 300px;'
          v-model='resetForm.Password'
          :placeholder='$t("common.pleaseInput")'
          type='password'
          @keyup.enter='submitReset(resetFormRef)'
          autocomplete='off'
          @copy.native.capture.prevent='()=>{}'
          @paste.native.capture.prevent='()=>{}'
          @cut.native.capture.prevent='()=>{}'
        />
      </el-form-item>
    </el-form>
    <el-button style="margin-top: 20px;" type="primary" @click="submitReset(resetFormRef)">
      {{ $t("common.save") }}
    </el-button>
  </basic-information-container>
</template>

<script>
import {defineComponent, ref} from 'vue';
import { useI18n } from 'vue-i18n'
import BasicInformationContainer from '@/components/BasicInformationContainer.vue';
import { restoreDefaults } from '@/api/reset';
import { AppMessage, showErrorAlert, clearSessionAndJumpToLogin } from '@/utils/commonMethods';
import errorTipMapper from '@/api/errorTipMapper';
import { ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';
import errorCode from '@/api/errorCode';

export default defineComponent({
  name: 'RestoreDefaults',
  components: {
    BasicInformationContainer,
  },
  setup() {
    const router = useRouter();
    const { t } = useI18n()
    const resetFormRef = ref();
    const resetForm = ref({
      isRetain: true,
      Password: null,
    })

    const resetFormRule = ref({
      Password: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
    })

    const submitReset = async (formElement) => {
      if (!formElement) {
        return
      };
      await formElement.validate(async (valid) => {
        if (!valid) {
          return
        };
        await ElMessageBox.confirm(
          t('restore.restoreFactorySetting.resetIrrevocableTip'),
          t('restore.restoreFactorySetting.resetConfirm'),
          {
            confirmButtonText: t('common.confirm'),
            cancelButtonText: t('common.cancel'),
            type: 'warning',
          }
        );
        let params = {
          ReserveIP: resetForm.value.isRetain,
          Password: resetForm.value.Password,
        }
        try {
          await restoreDefaults(params);
          AppMessage.success(t('restore.restoreFactorySetting.restoreSuccessTip'));
          clearSessionAndJumpToLogin(router)
        } catch (err) {
          let status = err?.response?.data?.error['@Message.ExtendedInfo'][0]?.Oem?.status;
          let msg = err?.response?.data?.error['@Message.ExtendedInfo'][0]?.Message;
          if (Object.prototype.hasOwnProperty.call(errorTipMapper, status)) {
            showErrorAlert(t(errorTipMapper[status]))
          } else if (status === errorCode.SESSION.ERROR_USER_LOCK_STATE) {
            showErrorAlert(t('session.lockTip', { seconds: msg.split('[')[1].split(']')[0] }));
            clearSessionAndJumpToLogin(router);
          } else {
            handleRestoreDefaultsError(msg);
          }
        } finally {
          formElement.resetFields();
        }
      })
    }

    const handleRestoreDefaultsError = (msg) => {
      let errorNum = msg?.split('-')[0];
      if (Object.prototype.hasOwnProperty.call(RestoreDefaultsErrorTipMapper, errorNum)) {
        showErrorAlert(RestoreDefaultsErrorTipMapper[errorNum])
      } else {
        showErrorAlert(t('restore.restoreFactorySetting.unknownErrTip'))
      }
    }

    let RestoreDefaultsErrorTipMapper = {
      'ERR.010': t('restore.restoreFactorySetting.restoreSuccessTip'),
      'ERR.011': t('restore.restoreFactorySetting.invalidParamTip'),
      'ERR.012': t('restore.restoreFactorySetting.mefRestoreFailedTip'),
      'ERR.013': t('restore.restoreFactorySetting.omRestoreFailedTip'),
      'ERR.014': t('restore.restoreFactorySetting.networkRestoreFailedTip'),
      'ERR.015': t('restore.restoreFactorySetting.rebootFailedTip'),
      'ERR.016': t('restore.restoreFactorySetting.unknownErrTip'),
      'ERR.0555': t('restore.restoreFactorySetting.mefIsStarting'),
    }

    return {
      resetFormRef,
      resetForm,
      resetFormRule,
      submitReset,
    }
  },
});
</script>

<style lang="scss" scoped>
.reset-tip {
  display: flex;
  align-items: center;
  img {
    width: 14px;
    height: 14px;
    margin-right: 10px;
  };
  font-size: 12px;
  color: var(--el-customize-tip-text-color);
  line-height: 16px;
  font-weight: 400;
  margin-bottom: 10px;
  width: 700px;
}
</style>