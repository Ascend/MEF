<template>
  <div class="tab-title" style="margin-top: 10px;">
    {{ $t('reset.restoreFactorySetting.tabName') }}
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
    >
      <el-form-item prop='isRetain' :label="$t('reset.restoreFactorySetting.ipAddress')" class="not-required">
        <el-radio-group v-model="resetForm.isRetain">
          <el-radio :label="true" style="margin-right: 100px;">
            {{ $t("reset.restoreFactorySetting.retain") }}
          </el-radio>
          <el-radio :label="false" style="margin-right: 10px;">
            {{ $t("reset.restoreFactorySetting.notRetain") }}
          </el-radio>
        </el-radio-group>
        <div class="reset-tip" style="margin-top: 20px;">
          <img src="@/assets/img/home/information.svg" alt=""/>
          {{ $t('reset.restoreFactorySetting.resetDeleteConfigTip') }}
        </div>
        <div class="reset-tip">
          <img src="@/assets/img/home/information.svg" alt=""/>
          {{ $t('reset.restoreFactorySetting.resetNotRecoverIpTip') }}
        </div>
        <div class="reset-tip">
          <img src="@/assets/img/home/information.svg" alt=""/>
          {{ $t('reset.restoreFactorySetting.resetIrrevocableTip') }}
        </div>
      </el-form-item>
      <el-form-item prop='ipAddress' v-if="resetForm.isRetain" :label="$t('reset.restoreFactorySetting.chooseRetainIpAddress')" class="required">
        <el-select v-model='resetForm.ipAddress' style='width: 300px; margin-bottom: 10px;' :placeholder='$t("common.pleaseSelect")'>
          <el-option v-for='item in ipAddressOptions' :key='item.value' :label='item.label' :value='item.value' />
        </el-select>
        <div class="reset-tip" v-if="!isSelectDefaultIp">
          <img src="@/assets/img/home/information.svg" alt=""/>
          {{ $t('reset.restoreFactorySetting.setRetainedIpTip') }}
        </div>
      </el-form-item>
      <el-form-item prop='rootPassword' :label="$t('common.password')" class="required">
        <el-input
          style='width: 300px;'
          v-model='resetForm.rootPassword'
          :placeholder='$t("reset.restoreFactorySetting.inputText")'
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
import {defineComponent, ref, onMounted} from 'vue';
import { useI18n } from 'vue-i18n'
import BasicInformationContainer from '@/components/BasicInformationContainer.vue';
import { queryEthernetIpList, resetAction } from '@/api/reset';
import { AppMessage, clearSessionAndJumpToLogin, handleOperationResponseError } from '@/utils/commonMethods';
import { ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';

export default defineComponent({
  name: 'SettingReset',
  components: {
    BasicInformationContainer,
  },
  setup() {
    const { t } = useI18n()
    const resetFormRef = ref();
    const resetForm = ref({
      isRetain: true,
      ipAddress: null,
      rootPassword: null,
    })
    const isSelectDefaultIp = ref(false);

    const resetFormRule = ref({
      ipAddress: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
      rootPassword: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
    })
    const ipAddressOptions = ref([])
    
    const router = useRouter();
    
    const submitReset = async (formElement) => {
      if (!formElement) {
        return
      };
      await formElement.validate(async (valid) => {
        if (!valid) {
          return
        };
        await ElMessageBox.confirm(
          t('reset.restoreFactorySetting.resetIrrevocableTip'),
          t('reset.restoreFactorySetting.resetConfirm'),
          {
            confirmButtonText: t('common.confirm'),
            cancelButtonText: t('common.cancel'),
            type: 'warning',
          }
        );
        let params = {
          ethernet: resetForm.value.isRetain ? resetForm.value.ipAddress : '',
          root_pwd: resetForm.value.rootPassword,
        }
        try {
          await resetAction(params);
          clearSessionAndJumpToLogin(router);
          AppMessage.warning(t('reset.restoreFactorySetting.resetSuccessTip'));
        } catch (err) {
          handleOperationResponseError(err);
        } finally {
          formElement.resetFields();
        }
      })
    }

    const initIpAddressOptions = (ipList) => {
      let thisHost = window.location.host;
      let options = []
      Object.getOwnPropertyNames(ipList).forEach((key) => {
        if (thisHost === ipList[key]) {
          ipList[key] = ipList[key] + ' (' + t('reset.restoreFactorySetting.current') + ')';
          resetForm.value.ipAddress = key;
          isSelectDefaultIp.value = true;
        }
        let option = {};
        option.value = key;
        option.label = ipList[key];
        options.push(option);
      })
      ipAddressOptions.value = options;
    }

    const init = async () => {
      let { data } = await queryEthernetIpList();
      if (!data?.ip_addr_list) {
        return
      };
      initIpAddressOptions(data?.ip_addr_list);
    }

    onMounted(async () => {
      await init();
    })

    return {
      resetFormRef,
      resetForm,
      resetFormRule,
      ipAddressOptions,
      isSelectDefaultIp,
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