<template>
  <div class='container'>
    <el-row style="margin-bottom: 30px;">
      <el-col :span='2' class="title">{{ $t('network.networkWireless.configApn') }}</el-col>
      <el-col :span='18' class='align-center'>
        <img src='@/assets/img/manager/abnormal.svg' alt=''/>
        <span style='margin-left: 10px;'>{{ $t('network.networkWireless.apnTip') }}</span>
      </el-col>
      <el-col :span='4' style="text-align: end;">
        <span v-if="!isShowEditForm" style="cursor: pointer; float: right;" class='align-center'  @click="clickEdit" >
          <img src='@/assets/img/edit.svg' alt=''/>
          <span style='margin-left: 10px;'>{{ $t('common.edit') }}</span>
        </span>
        <span v-else>
          <el-button size='small' type='primary' @click='saveEditApn(editApnInfoFormRef)'>{{ $t('common.save') }}</el-button>
          <el-button size='small' @click='cancelEditApn'>{{ $t('common.cancel') }}</el-button>
        </span>
      </el-col>
    </el-row>
    <el-row v-if="!isShowEditForm" class='sub-block'>
      <el-col :span='12'>
        <div class='sub-block-title'>{{ $t('network.networkWireless.apn') }}</div>
        <div class='sub-block-content'>{{ editApnInfoForm.apn_name ?? '--' }}</div>
      </el-col>
      <el-col :span='12'>
        <div class='sub-block-title'>{{ $t('network.networkWireless.authType') }}</div>
        <div class='sub-block-content'>{{ authTypeMapper[apnInfo.authType] ?? '--' }}</div>
      </el-col>
    </el-row>
    <el-form
      label-position="left"
      :model='editApnInfoForm'
      :rules='editApnInfoFormRule'
      ref='editApnInfoFormRef'
      :inline="true"
      v-else
    >
      <el-row class='top-form'>
        <el-col :span="6">
          <div class='sub-block-title required'>{{ $t('network.networkWireless.apn') }}</div>
          <el-form-item prop="apn_name">
            <el-input v-model="editApnInfoForm.apn_name" :placeholder='$t("common.pleaseInput")'/>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <div class='sub-block-title required'>{{ $t('network.networkWireless.authType') }}</div>
          <el-form-item prop="auth_type">
            <el-select v-model="editApnInfoForm.auth_type" @change="changeAuthType">
              <el-option
                v-for="item in authTypeOption"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <div class='sub-block-title' style="margin-top: -10px;">{{ authTypeTip }}</div>
        </el-col>
        <el-col :span="6">
          <div class='sub-block-title'>{{ $t('common.username') }}</div>
          <el-form-item prop="apn_user">
            <el-input
              v-model="editApnInfoForm.apn_user"
              type='password'
              autocomplete="off"
              @copy.native.capture.prevent="()=>{}"
              @paste.native.capture.prevent="()=>{}"
              @cut.native.capture.prevent="()=>{}"
              :placeholder='$t("common.pleaseInput")'
            />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <div class='sub-block-title'>{{ $t('common.password') }}</div>
          <el-form-item prop='apn_passwd'>
            <el-input
              v-model='editApnInfoForm.apn_passwd'
              :placeholder='$t("common.pleaseInput")'
              type='password'
              @keyup.enter='saveEditApn(editApnInfoFormRef)'
              autocomplete="off"
              @copy.native.capture.prevent="()=>{}"
              @paste.native.capture.prevent="()=>{}"
              @cut.native.capture.prevent="()=>{}"
            />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>
  </div>
  <div class='container' style="margin-top: 20px; height: 60vh;">
    <div class="title" style="margin-bottom: 20px;">{{ $t('network.networkWireless.mobileNetwork') }}</div>
    <app-tip :tip-text='$t("network.networkWireless.needSimTip")' />
    <app-tip :tip-type='"warn"' v-if="!hasSim" :tip-text='$t("network.networkWireless.noSimTip")' />
    <el-row class="config-row">
      <el-col :span="2" class="mobile-list-title">{{ $t("network.networkWireless.wirelessSwitch") }}</el-col>
      <el-col :span="20">
        <el-switch v-model="editMobileNetworkForm.state_lte" :disabled="!hasSim" :before-change="beforeSwitchStateLte"/>
      </el-col>
    </el-row>
    <el-row class="config-row">
      <el-col :span="2" class="mobile-list-title">{{ $t("network.common.ipAddress") }}</el-col>
      <el-col :span="4">
        <el-input v-model="editMobileNetworkForm.ipAddress" :disabled="true">
        </el-input>
      </el-col>
    </el-row>
    <el-row class="config-row">
      <el-col :span="2" class="mobile-list-title">{{ $t("network.networkWireless.mobileData") }}</el-col>
      <el-col :span="20">
        <el-switch
          v-model="editMobileNetworkForm.state_data"
          :before-change="beforeSwitchStateData"
          :disabled="!hasSim"
        />
      </el-col>
    </el-row>
    <el-row class="config-row">
      <el-col :span="2" class="mobile-list-title">{{ $t("network.networkWireless.networkStatus") }}</el-col>
      <el-col :span="20" class="align-center">
        <img v-bind:src='editMobileNetworkForm.signalLevelImg' alt=''/>
        <span class="signal-text">{{ editMobileNetworkForm.networkType }}</span>
        <el-button :disabled="!editMobileNetworkForm.state_data" size="small" style="margin-left: 20px;" @click="clickTest">
          {{ $t('network.common.test') }}
        </el-button>
      </el-col>
    </el-row>
  </div>
  <el-dialog
    v-model="lteDialog.isShow"
    :title='$t("common.tips")'
    width="600px"
  >
    <div style="display: flex; align-items: center; word-break: break-word;">
      <img src="@/assets/img/alarm.svg" alt="" style="margin-right: 10px;"/>
      {{ lteDialog.tip }}
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="confirmLteDialog">
          {{ $t("common.confirm") }}
        </el-button>
        <el-button @click="cancelLteDialog">
          {{ $t("common.cancel") }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import { defineComponent, ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

import { validateApnName } from '@/utils/validator';
import { queryWirelessConfigInfo, configWirelessStatusInfo, queryWirelessStatusInfo, configWireless } from '@/api/network';
import AppTip from '@/components/AppTip.vue';
import { handleOperationResponseError, showWarningAlert, AppMessage } from '@/utils/commonMethods';
import constants from '@/utils/constants';
import signal0 from '@/assets/img/manager/signal-0.svg';
import signal1 from '@/assets/img/manager/signal-1.svg';
import signal2 from '@/assets/img/manager/signal-2.svg';
import signal3 from '@/assets/img/manager/signal-3.svg';
import signal4 from '@/assets/img/manager/signal-4.svg';
import signal5 from '@/assets/img/manager/signal-5.svg';

export default defineComponent({
  name: 'WirelessView',
  components: {
    AppTip,
  },
  setup() {
    const apnInfo = ref({
      authType: null,
      modeType: null,
    });
    const { t } = useI18n();
    const isShowEditForm = ref(false);
    const hasSim = ref(false);
    const lteDialog = ref({
      isShow: false,
      tip: null,
    })
    const editApnInfoFormRef = ref();
    const editApnInfoForm = ref({
      'apn_name': null,
      'apn_user': null,
      'auth_type': null,
      'apn_passwd': null,
    })
    const authTypeOption = [
      {
        value: '0',
        label: 'None',
      },
      {
        value: '1',
        label: 'PAP',
      },
      {
        value: '2',
        label: 'CHAP',
      },
      {
        value: '3',
        label: 'PAP or CHAP',
      },
    ]
    const authTypeTip = ref('')

    const editMobileNetworkForm = ref({
      'state_lte': null,
      'ipAddress': null,
      'state_data': null,
      signalLevelImg: signal0, // 0-5级
      networkType: null, // 5G、4G、3G、2G
    });
    const editMobileNetworkFormRef = ref();

    const validateApnUserName = (rule, value, callback) => {
      if (!value) {
        if (editApnInfoForm.value.apn_passwd && editApnInfoForm.value.apn_passwd.length > 0) {
          callback(new Error(t('network.networkWireless.fieldEmptyTip')))
          return;
        }
        callback()
        return;
      }
      if (!editApnInfoForm.value.auth_type) {
        callback()
        return;
      }

      let maxLength = 64;
      let pattern = /^[a-zA-Z0-9\-_\.@]{1,64}$/;
      if (!pattern.test(value)) {
        callback(new Error(t('network.networkWireless.userNameErrorTip', { length: maxLength })))
        return;
      }

      callback()
    }

    const validateApnUserPassword = (rule, value, callback) => {
      if (!value) {
        if (editApnInfoForm.value.apn_user && editApnInfoForm.value.apn_user.length > 0) {
          callback(new Error(t('network.networkWireless.fieldEmptyTip')))
          return;
        }
        callback()
        return;
      }
      const pattern = /^[a-zA-Z0-9~`!? .:;-_"\(\)\{\}\[\]\/<>@#\$%\^&\*\+\|\\=\s]{1,64}$/;
      if (!pattern.test(value)) {
        callback(new Error(t('network.networkWireless.userPasswordErrorTip')))
        return;
      }

      callback()
    }

    const editApnInfoFormRule = reactive({
      'apn_name': [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateApnName,
          trigger: 'blur',
        },
      ],
      'apn_user': [
        {
          validator: validateApnUserName,
          trigger: 'blur',
        },
      ],
      'auth_type': [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
      'apn_passwd': [
        {
          validator: validateApnUserPassword,
          trigger: 'blur',
        },
      ],
    })

    const clickEdit = () => {
      isShowEditForm.value = true;
      const chapType = '2'
      editApnInfoForm.value.auth_type = chapType;
      changeAuthType(chapType);
    }

    const saveEditApn = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        editApnInfoForm.value.apn_user = editApnInfoForm.value.apn_user || null
        editApnInfoForm.value.apn_passwd = editApnInfoForm.value.apn_passwd || null
        try {
          await configWireless(editApnInfoForm.value)
          AppMessage.success(t('common.modifySuccessfully'))
        } catch (err) {
          handleOperationResponseError(err)
        } finally {
          formElement.resetFields();
          cancelEditApn()
          await init();
        }
      })
    }

    const cancelEditApn = () => {
      editApnInfoFormRef.value.resetFields();
      isShowEditForm.value = false;
    }

    const beforeSwitchStateLte = () => {
      // state_lte 为 true，尝试关闭无线网络开关
      if (editMobileNetworkForm.value.state_lte) {
        lteDialog.value.isShow = true;
        lteDialog.value.tip = t('network.networkWireless.turnOffWirelessTip')
        return false
      }

      // state_lte 为 false，尝试打开无线网络开关
      // 且已配置了默认网关
      if (defaultGateway.value) {
        lteDialog.value.isShow = true;
        lteDialog.value.tip = t('network.networkWireless.turnOnWirelessTip')
        return false
      }

      // state_lte 为 false，尝试打开无线网络开关
      // 未配置默认网关，直接调用接口
      confirmLteDialog()
      return true
    }

    const beforeSwitchStateData = async () => {
      if (!editMobileNetworkForm.value.state_lte) {
        showWarningAlert(t('network.networkWireless.turnOnDataTip'))
        return false;
      }

      let params = {
        'state_lte': editMobileNetworkForm.value.state_lte,
        'state_data': !editMobileNetworkForm.value.state_data,
      }
      try {
        await configWirelessStatusInfo(params);
        AppMessage.success(t('common.modifySuccessfully'))
        await init();
      } catch (err) {
        handleOperationResponseError(err)
      }
      return false;
    }

    const confirmLteDialog = async () => {
      let params = {
        'state_lte': !editMobileNetworkForm.value.state_lte,
        'state_data': false,
      }
      try {
        await configWirelessStatusInfo(params);
        AppMessage.success(t('common.modifySuccessfully'))
        await init();
      } catch (err) {
        handleOperationResponseError(err)
      } finally {
        cancelLteDialog();
      }
    }

    const cancelLteDialog = () => {
      lteDialog.value.isShow = false;
    }

    const fetchApnInfo = async () => {
      let { data } = await queryWirelessConfigInfo();
      return data;
    }

    const fetchStatusInfo = async () => {
      let { data } = await queryWirelessStatusInfo();
      return data;
    }

    const initApnInfo = (data) => {
      editApnInfoForm.value.apn_name = data?.apn_name;
      editApnInfoForm.value.apn_user = data?.apn_user;
      editApnInfoForm.value.auth_type = data?.auth_type;
      changeAuthType(data?.auth_type)
      apnInfo.value.authType = data?.auth_type;
      apnInfo.value.modeType = data?.mode_type;
    }

    const defaultGateway = ref();
    const initStatusInfo = (data) => {
      hasSim.value = data?.sim_exist;
      editMobileNetworkForm.value.state_lte = data?.state_lte;
      editMobileNetworkForm.value.state_data = data?.state_data;
      editMobileNetworkForm.value.ipAddress = data?.ip_addr || constants.DEFAULT_EMPTY_TEXT;
      editMobileNetworkForm.value.networkType = data?.network_type;
      defaultGateway.value = data?.default_gateway;

      let imgMapper = {
        0: signal0,
        1: signal1,
        2: signal2,
        3: signal3,
        4: signal4,
        5: signal5,
      }
      if (data?.network_signal_level) {
        editMobileNetworkForm.value.signalLevelImg = imgMapper[data?.network_signal_level]
      } else {
        editMobileNetworkForm.value.signalLevelImg = imgMapper[0]
      }
    }

    const authTypeMapper = {
      0: 'NONE',
      1: 'PAP',
      2: 'CHAP',
      3: 'PAP or CHAP',
    }

    const init = async () => {
      initApnInfo(await fetchApnInfo());
      initStatusInfo(await fetchStatusInfo());
    }

    const clickTest = async () => {
      initStatusInfo(await fetchStatusInfo());
    }

    onMounted(async () => {
      await init();
    })

    const changeAuthType = (val) => {
      const authTypeTipsMapper = {
        0: t('network.networkWireless.authTypeTip'),
        1: t('network.networkWireless.authTypeTip'),
        2: '',
        3: t('network.networkWireless.authTypeTip'),
      }
      authTypeTip.value = authTypeTipsMapper[val]
    }

    return {
      lteDialog,
      hasSim,
      apnInfo,
      isShowEditForm,
      editApnInfoFormRef,
      editApnInfoForm,
      editApnInfoFormRule,
      authTypeOption,
      editMobileNetworkForm,
      editMobileNetworkFormRef,
      authTypeMapper,
      authTypeTip,
      confirmLteDialog,
      cancelLteDialog,
      beforeSwitchStateLte,
      clickEdit,
      saveEditApn,
      cancelEditApn,
      beforeSwitchStateData,
      clickTest,
      changeAuthType,
    }
  },
})
</script>

<style scoped>
.container {
  background: var(--el-bg-color-overlay);
  border-radius: 4px;
  padding: 24px;
}

.align-center {
  display: flex;
  align-items: center;
}

.sub-block > div {
  border-left: 1px solid var(--el-border-color);
  padding-left: 14px;
  margin-bottom: 32px;
}

.top-form > div {
  border-left: 1px solid var(--el-border-color);
  padding-left: 14px;
}

.sub-block {
  margin-top: 30px;
  width: 50%;

}

.sub-block-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 200;
  margin-bottom: 8px;
}

.sub-block-content {
  font-size: 14px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
}

.required:after {
  content: "*";
  color: var(--el-color-danger);
  margin-left: 4px;
}

.title {
  font-size: 16px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
}

.mobile-list-title {
  font-size: 14px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 32px;
  font-weight: 400;
}

.ip-block {
  border-right: 0;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  box-shadow: 1px 0 0 0 var(--el-input-border-color) inset,0 1px 0 0 var(--el-input-border-color) inset,0 -1px 0 0 var(--el-input-border-color) inset;
}

.config-row {
  margin-top: 20px;
}

.signal-text {
  font-size: 12px;
  color: #D3DCE9;
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
  margin-left: 6px;
}
</style>