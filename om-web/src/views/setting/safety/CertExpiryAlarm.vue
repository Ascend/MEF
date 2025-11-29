<template>
  <basic-information-container style='height: 70vh' >
    <el-form
      :model='certAlarmForm'
      label-position='left'
      label-width='200px'
      :rules='certAlarmFormRule'
      ref='certAlarmFormRef'
      style='margin-top: 20px; width: 500px;'
    >
      <el-form-item :label='$t("safety.certificateExpiryAlarm.warningPeriod")' prop='CertAlarmTime'>
        <el-row style='width: 300px;'>
          <el-col :span='22'>
            <el-input-number
              :min='constants.MIN_EXPIRY_ALARM'
              :max='constants.MAX_EXPIRY_ALARM'
              v-model='certAlarmForm.CertAlarmTime'
              style='width: 98%;'
            />
          </el-col>
          <el-col :span='2' style='padding-top: 6px;'>
            <app-popover :text='$t("safety.certificateExpiryAlarm.expirationTip")' />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item :label='$t("common.password")' prop='Password'>
        <el-input
          style='width: 300px;'
          v-model='certAlarmForm.Password'
          :placeholder='$t("common.pleaseInput")'
          type='password'
          @keyup.enter='submitChange(certAlarmFormRef)'
          autocomplete='off'
          @copy.native.capture.prevent='()=>{}'
          @paste.native.capture.prevent='()=>{}'
          @cut.native.capture.prevent='()=>{}'
        />
      </el-form-item>
    </el-form>
    <el-button class='save-btn' type="primary" @click='submitChange(certAlarmFormRef)'>
      {{ $t('common.save') }}
    </el-button>
  </basic-information-container>
</template>

<script>

import { defineComponent, ref, reactive, onMounted } from 'vue';
import { useI18n } from 'vue-i18n'

import { queryHttpsCertAlarmTime, modifyHttpsCertAlarmTime } from '@/api/safety';
import { validateCertAlarmTime } from '@/utils/validator';
import BasicInformationContainer from '@/components/BasicInformationContainer.vue';
import AppPopover from '@/components/AppPopover.vue';
import constants from '@/utils/constants';
import { AppMessage, handleOperationResponseError } from '@/utils/commonMethods';

export default defineComponent({
  name: 'CertAlarm',
  components: {
    BasicInformationContainer,
    AppPopover,
  },
  setup() {
    const { t } = useI18n()

    const certAlarmForm = ref({
      CertAlarmTime: null,
      Password: null,
    })
    const certAlarmFormRef = ref();

    const certAlarmFormRule = reactive({
      CertAlarmTime: [
        {
          required: true,
          message: t('common.wrongInput'),
          trigger: 'blur',
        },
        {
          validator: validateCertAlarmTime,
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
          await modifyHttpsCertAlarmTime(certAlarmForm.value);
          AppMessage.success(t('common.saveSuccessfully'));
          await fetchAndInitAlarmTime();
        } catch (err) {
          handleOperationResponseError(err)
        } finally {
          formElement.resetFields();
          await fetchAndInitAlarmTime();
        }
      })
    }

    const fetchAndInitAlarmTime = async () => {
      let { data } = await queryHttpsCertAlarmTime();
      certAlarmForm.value.CertAlarmTime = data?.CertAlarmTime ?? constants.DEFAULT_EXPIRY_ALARM;
    }

    onMounted(async () => {
      await fetchAndInitAlarmTime();
    })

    return {
      constants,
      certAlarmForm,
      certAlarmFormRule,
      certAlarmFormRef,
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