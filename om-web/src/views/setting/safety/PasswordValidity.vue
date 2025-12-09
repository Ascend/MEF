<template>
  <basic-information-container style='height: 70vh' >
    <el-form
      :model='passwordValidityForm'
      label-position='left'
      label-width='140px'
      :rules='passwordValidityFormRule'
      ref='passwordValidityFormRef'
      style='margin-top: 20px; width: 500px;'
    >
      <el-form-item :label='$t("safety.passwordValidity.validityPeriod")' prop='PasswordExpirationDays'>
        <el-row style='width: 300px;'>
          <el-col :span='22'>
            <el-input-number
              :min='constants.MIN_PASSWORD_VALIDITY'
              :max='constants.MAX_PASSWORD_VALIDITY'
              v-model='passwordValidityForm.PasswordExpirationDays'
              style='width: 98%;'
            />
          </el-col>
          <el-col :span='2' style='padding-top: 6px;'>
            <app-popover :text='$t("safety.passwordValidity.validityTip")' />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item :label='$t("common.password")' prop='Password'>
        <el-input
          style='width: 300px;'
          v-model='passwordValidityForm.Password'
          :placeholder='$t("common.pleaseInput")'
          type='password'
          @keyup.enter='submitChange(passwordValidityFormRef)'
          autocomplete='off'
          @copy.native.capture.prevent='()=>{}'
          @paste.native.capture.prevent='()=>{}'
          @cut.native.capture.prevent='()=>{}'
        />
      </el-form-item>
    </el-form>
    <el-button class='save-btn' type="primary" @click='submitChange(passwordValidityFormRef)'>
      {{ $t('common.save') }}
    </el-button>
  </basic-information-container>
</template>

<script>

import { defineComponent, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n'

import { queryAccountService, modifyAccountService, getUserInfo } from '@/api/account';
import BasicInformationContainer from '@/components/BasicInformationContainer.vue';
import AppPopover from '@/components/AppPopover.vue';
import constants from '@/utils/constants';
import { AppMessage, showErrorAlert, handleOperationResponseError } from '@/utils/commonMethods';

export default defineComponent({
  name: 'PasswordValidity',
  components: {
    BasicInformationContainer,
    AppPopover,
  },
  setup() {
    const { t } = useI18n()
    const passwordValidityFormRef = ref();
    const passwordValidityForm = ref({
      PasswordExpirationDays: null,
      Password: null,
    })

    const passwordValidityFormRule = ref({
      PasswordExpirationDays: [
        {
          required: true,
          message: t('common.wrongInput'),
          trigger: 'blur',
        },
      ],
      Password: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
    })

    const submitChange = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        try {
          await modifyAccountService(passwordValidityForm.value);
          AppMessage.success(t('common.saveSuccessfully'))
        } catch (err) {
          handleOperationResponseError(err)
        } finally {
          formElement.resetFields();
          await fetchAndInitExpirationDay();
          await showPasswordValidDaysInfo();
        }
      })
    }

    const showPasswordValidDaysInfo = async () => {
      let { data } = await getUserInfo(sessionStorage.getItem('userId'));
      if (data?.Oem?.PasswordValidDays) {
        const PasswordValidDays = data.Oem.PasswordValidDays;
        if (PasswordValidDays !== '--' && parseInt(PasswordValidDays, 10) > 0 && parseInt(PasswordValidDays, 10) <= 10) {
          AppMessage.warning(t('safety.passwordValidity.passwordWillExpire', { day: PasswordValidDays }))
        } else if (PasswordValidDays !== '--' && parseInt(PasswordValidDays, 10) === 0) {
          showErrorAlert(t('safety.passwordValidity.passwordHasExpired'))
        }
      }
    }

    const fetchAndInitExpirationDay = async () => {
      let { data } = await queryAccountService();
      passwordValidityForm.value.PasswordExpirationDays = data?.PasswordExpirationDays ?? constants.DEFAULT_PASSWORD_VALIDITY;
    }

    onMounted(async () => {
      await fetchAndInitExpirationDay();
      await showPasswordValidDaysInfo();
    })

    return {
      constants,
      passwordValidityForm,
      passwordValidityFormRule,
      passwordValidityFormRef,
      submitChange,
    }
  },
})
</script>

<style scoped>
.save-btn {
  padding: 12px 24px;
  margin-top: 40px;
}
</style>