<template>
  <basic-information-container style='height: 70vh' >
    <el-form
      :model='webTimeoutForm'
      label-position='left'
      label-width='140px'
      :rules='webTimeoutFormRule'
      ref='webTimeoutFormRef'
      style='margin-top: 20px; width: 500px;'
    >
      <el-form-item :label='$t("safety.webTimeout.timeoutPeriod")' prop='SessionTimeout'>
        <el-row style='width: 300px;'>
          <el-col :span='22'>
            <el-input-number
              :min='constants.MIN_WEB_TIMEOUT'
              :max='constants.MAX_WEB_TIMEOUT'
              v-model='webTimeoutForm.SessionTimeout'
              style='width: 98%;'
            />
          </el-col>
          <el-col :span='2' style='padding-top: 6px;'>
            <app-popover :text='$t("safety.webTimeout.timeoutTip")' />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item :label='$t("common.password")' prop='Password'>
        <el-input
          style='width: 300px;'
          v-model='webTimeoutForm.Password'
          :placeholder='$t("common.pleaseInput")'
          type='password'
          @keyup.enter='submitChange(webTimeoutFormRef)'
          autocomplete='off'
          @copy.native.capture.prevent='()=>{}'
          @paste.native.capture.prevent='()=>{}'
          @cut.native.capture.prevent='()=>{}'
        />
      </el-form-item>
    </el-form>
    <el-button class='save-btn' type="primary" @click='submitChange(webTimeoutFormRef)'>
      {{ $t('common.save') }}
    </el-button>
  </basic-information-container>
</template>

<script>
import { defineComponent, ref, reactive, onMounted } from 'vue';
import { useI18n } from 'vue-i18n'

import { querySessionService, modifySessionService } from '@/api/account';
import { validateSessionTimeout } from '@/utils/validator';
import BasicInformationContainer from '@/components/BasicInformationContainer.vue';
import AppPopover from '@/components/AppPopover.vue';
import constants from '@/utils/constants';
import { AppMessage, handleOperationResponseError } from '@/utils/commonMethods';

export default defineComponent({
  name: 'WebTimeout',
  components: {
    BasicInformationContainer,
    AppPopover,
  },
  setup() {
    const { t } = useI18n()
    const webTimeoutFormRef = ref();
    const webTimeoutForm = ref({
      SessionTimeout: null,
      Password: null,
    })

    const webTimeoutFormRule = reactive({
      SessionTimeout: [
        {
          required: true,
          message: t('common.wrongInput'),
          trigger: 'blur',
        },
        {
          validator: validateSessionTimeout,
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
          await modifySessionService(webTimeoutForm.value);
          AppMessage.success(t('common.saveSuccessfully'));
        } catch (err) {
          handleOperationResponseError(err);
        } finally {
          formElement.resetFields();
          await fetchAndInitWebTimeout();
        }
      })
    }

    const fetchAndInitWebTimeout = async () => {
      let { data } = await querySessionService();
      webTimeoutForm.value.SessionTimeout = data?.SessionTimeout ?? constants.DEFAULT_WEB_TIMEOUT;
    }

    onMounted(async () => {
      await fetchAndInitWebTimeout();
    })

    return {
      constants,
      webTimeoutForm,
      webTimeoutFormRule,
      webTimeoutFormRef,
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