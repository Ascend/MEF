<template>
  <div class="tab-title">{{ $t("registration.nms.tabName") }}</div>
  <div class="container info-block">
    <div class="container-title">{{ $t("registration.nms.currentMode") }}</div>
    <div class="info">
      <div>
        <div class="title">
          {{ $t("registration.nms.currentMode") }}
        </div>
        <div class="content">
          {{ currConfig.mode }}
        </div>
      </div>
      <div v-if="!isWebMode(currConfig.mode)">
        <div class="title">
          {{ $t("registration.nms.ipAddress") }}
        </div>
        <div class="content">
          {{ currConfig.ip }}
        </div>
      </div>
      <div v-if="!isWebMode(currConfig.mode)">
        <div class="title">
          {{ $t("registration.nms.account") }}
        </div>
        <div class="content">
          {{ currConfig.account }}
        </div>
      </div>
      <div v-if="!isWebMode(currConfig.mode)">
        <div class="title">
          {{ $t("registration.nms.connectStatus") }}
        </div>
        <div class="content">
          {{ currConfig.connectStatus }}
        </div>
      </div>
    </div>
  </div>
  <div class="container config-block">
    <div class="container-title">{{ $t("registration.nms.configInfo") }}</div>
    <app-tip :tip-text="registrationTip" />
    <div>
      <span class="select-mode">{{ $t("registration.nms.selectMode") }}</span>
      <el-radio-group v-model="selectedMode" style="margin-left: 20px;">
        <el-radio label="web" style="margin-right: 10px;">
          {{ $t("registration.nms.webMode") }}
        </el-radio>
        <app-popover :text="$t('registration.nms.radioWebTip')" />
        <el-radio label="fd" style="margin-right: 10px; margin-left: 50px;">
          {{ $t("registration.nms.fdMode") }}
        </el-radio>
        <app-popover :text="$t('registration.nms.radioFdTip')" />
      </el-radio-group>
    </div>
    <el-form
      ref="fdFormRef"
      :model="fdForm"
      :rules="fdFormRule"
      style="margin-left: 120px; margin-top: 20px; width: 610px;"
      :class="{isEnglish: isEnglish()}"
      label-position='left'
      label-width="auto"
      v-if="selectedMode === 'fd'"
      :hide-required-asterisk="true"
    >
      <el-form-item prop="nodeId" class="required" :label="$t('registration.nms.nodeId')">
        <el-row :gutter="10" style="width: 100%;">
          <el-col :span="23">
            <el-input style="width: 380px;" v-model="fdForm.nodeId"></el-input>
          </el-col>
          <el-col :span="1" style="padding-top: 6px;">
            <app-popover :text="nodeIdTip" />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item prop="serverName" class="not-required" :label="$t('registration.nms.serverName')">
        <el-row :gutter="10" style="width: 100%;">
          <el-col :span="23">
            <el-input style="width: 380px;" v-model="fdForm.serverName"></el-input>
          </el-col>
          <el-col :span="1" style="padding-top: 6px;">
            <app-popover :text="$t('registration.nms.serverNameTip')" />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item prop="ipAddress" class="required" :label="$t('registration.nms.ipAddress')">
        <el-input style="width: 380px;" v-model="fdForm.ipAddress"></el-input>
      </el-form-item>
      <el-form-item prop="port" class="required" :label="$t('registration.nms.port')">
        <el-input style="width: 380px;" v-model="fdForm.port"></el-input>
      </el-form-item>
      <el-form-item prop="account" class="required" :label="$t('registration.nms.account')">
        <el-row :gutter="10" style="width: 100%;">
          <el-col :span="23">
            <el-input style="width: 380px;" v-model="fdForm.account"></el-input>
          </el-col>
          <el-col :span="1" style="padding-top: 6px;">
            <app-popover :text="$t('registration.nms.accountTip')" />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item prop="password" class="required" :label="$t('registration.nms.password')">
        <el-input
          style="width: 380px;"
          v-model="fdForm.password"
          type="password"
          autocomplete="off"
          @copy.native.capture.prevent="()=>{}"
          @paste.native.capture.prevent="()=>{}"
          @cut.native.capture.prevent="()=>{}"
        ></el-input>
      </el-form-item>
      <el-form-item prop="rootCertFile" class="not-required" :label="$t('registration.nms.rootCertFile')">
        <el-row :gutter="10" style="width: 100%;">
          <el-col :span="23">
            <el-upload
              class="upload-demo"
              action="/redfish/v1/NetManager/ImportFdCert"
              name="imgfile"
              :with-credentials="true"
              accept=".crt"
              :headers="uploadHeaders"
              :on-success="uploadCertSuccess"
              :on-error="uploadCertError"
              :show-file-list="false"
              v-model:file-list="certFileList"
              :before-upload="beforeUploadCertFile"
            >
              <el-input style="width: 380px;" v-model="fdForm.rootCertFile" :placeholder="$t('common.pleaseSelect')">
                <template #append>...</template>
              </el-input>
            </el-upload>
          </el-col>
          <el-col :span="1" style="padding-top: 6px;">
            <app-popover :text="rootCertTip" />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item class="not-required" :label="$t('registration.nms.currCertContent')">
        <el-row :gutter="10" style="width: 100%;">
          <el-col :span="15">
            <div class="cert-content">
              <el-row :gutter="10" class="cert-row">
                <el-col :span="6" class="cert-title">{{ $t("registration.nms.issuer") }}</el-col>
                <el-col :span="18">{{ certContent.issuer }}</el-col>
              </el-row>
              <el-row :gutter="10" class="cert-row">
                <el-col :span="6" class="cert-title">{{ $t("registration.nms.user") }}</el-col>
                <el-col :span="18">{{ certContent.user }}</el-col>
              </el-row>
              <el-row :gutter="10" class="cert-row">
                <el-col :span="6" class="cert-title">{{ $t("registration.nms.certValidPeriod") }}</el-col>
                <el-col :span="18">{{ certContent.certValidPeriod }}</el-col>
              </el-row>
              <el-row :gutter="10" class="cert-row">
                <el-col :span="6" class="cert-title">{{ $t("registration.nms.sn") }}</el-col>
                <el-col :span="18">{{ certContent.sn }}</el-col>
              </el-row>
              <el-row :gutter="10" class="cert-row">
                <el-col :span="6" class="cert-title">{{ $t("registration.nms.fingerPrint") }}</el-col>
                <el-col :span="18">{{ certContent.fingerPrint }}</el-col>
              </el-row>
            </div>
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item prop="crlFile" class="not-required" :label="$t('registration.nms.crlFile')">
        <el-row :gutter="10" style="width: 100%;">
          <el-col :span="23">
            <el-upload
              class="upload-demo"
              action="/redfish/v1/NetManager/ImportFdCrl"
              name="imgfile"
              :with-credentials="true"
              accept=".crl"
              :headers="uploadHeaders"
              :on-success="uploadCrlSuccess"
              :on-error="uploadCrlError"
              :show-file-list="false"
              v-model:file-list="crlFileList"
              :before-upload="beforeUploadCrlFile"
            >
              <el-input style="width: 380px;" v-model="fdForm.crlFile" :placeholder="$t('common.pleaseSelect')">
                <template #append>...</template>
              </el-input>
            </el-upload>
          </el-col>
          <el-col :span="1" style="padding-top: 6px;">
            <app-popover :text="$t('registration.nms.crlTip')" />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item prop="connectTest" class="not-required" :label="$t('registration.nms.interconnectionTest')">
        <el-row :gutter="10" style="width: 100%;">
          <el-col :span="23">
            <el-radio-group v-model="fdForm.connectTest">
              <el-radio :label="true" style="margin-right: 10px;">
                {{ $t("registration.nms.test") }}
              </el-radio>
              <el-radio :label="false" style="margin-right: 10px; margin-left: 50px;">
                {{ $t("registration.nms.notTest") }}
              </el-radio>
            </el-radio-group>
          </el-col>
          <el-col :span="1" style="padding-top: 6px;">
            <app-popover :text="$t('registration.nms.connectTestTip')" />
          </el-col>
        </el-row>
      </el-form-item>
    </el-form>
    <el-button type="primary" style="margin-top: 20px;" @click="saveConfig">{{ $t("common.save") }}</el-button>
  </div>
</template>

<script>
import { defineComponent, ref, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { queryNetManagerInfo, queryNetManagerNodeID, queryFdRootCert, modifyNetManagerInfo } from '@/api/registration';
import AppTip from '@/components/AppTip.vue';
import AppPopover from '@/components/AppPopover.vue';
import { validateIp, validateNodeId, validatePort, validateServerName, validateAccount } from '@/utils/validator';
import {
  AppMessage, showErrorAlert, getLanguageDefaultChinese,
  handleUploadError, isValidUploadFilename, getAuthToken
} from '@/utils/commonMethods';
import constants from '@/utils/constants';
import errorCode from '@/api/errorCode';
import errorTipMapper from '@/api/errorTipMapper';

export default defineComponent({
  name: 'MaintenanceRegistration',
  components: {
    AppTip,
    AppPopover,
  },
  setup() {
    const { t } = useI18n();
    const uploadHeaders = ref({
      ...getAuthToken(),
      'AutoRefresh': false,
    })
    const registrationTip = t('registration.nms.registrationTip', { model: sessionStorage.getItem('model') });
    const nodeIdTip = t('registration.nms.nodeIdTip', { model: sessionStorage.getItem('model') });
    const rootCertTip = t('registration.nms.rootCertTip', { model: sessionStorage.getItem('model') })
    const selectedMode = ref()
    const currConfig = ref({
      mode: '',
      ip: '',
      account: '',
      connectStatus: '',
    })

    const fdForm = ref({
      nodeId: '',
      serverName: '',
      ipAddress: '',
      port: '',
      account: '',
      password: null,
      connectTest: true,
      rootCertFile: '',
      crlFile: '',
    });

    const certContent = ref({
      issuer: '',
      user: '',
      certValidPeriod: '',
      sn: '',
      fingerPrint: '',
    })

    const certFileList = ref([])
    const crlFileList = ref([])

    const fdFormRule = ref({
      nodeId: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateNodeId,
          trigger: 'blur',
        },
      ],
      serverName: [
        {
          validator: validateServerName,
          trigger: 'blur',
        },
      ],
      ipAddress: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateIp,
          trigger: 'blur',
        },
      ],
      port: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validatePort,
          trigger: 'blur',
        },
      ],
      account: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateAccount,
          trigger: 'blur',
        },
      ],
      password: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
    });

    const fdFormRef = ref();

    const uploadCertSuccess = async (response, uploadFile) => {
      fdForm.value.rootCertFile = uploadFile.name;
      let { data: certData } = await queryFdRootCert();
      setCertInfo(certData);
      AppMessage.success(t('common.uploadSuccessfullyTip'))
    }

    const uploadCrlSuccess = (response, uploadFile) => {
      fdForm.value.crlFile = uploadFile.name;
      AppMessage.success(t('common.uploadSuccessfullyTip'))
    }

    const uploadCertError = (error) => {
      handleUploadError(error)
    }

    const uploadCrlError = (error) => {
      handleUploadError(error)
    }

    const setCertInfo = (data) => {
      certContent.value = {
        issuer: data?.Issuer,
        user: data?.Subject,
        certValidPeriod: data?.Date,
        sn: data?.SerialNum,
        fingerPrint: data?.Fingerprint,
      }
    }

    watch(selectedMode, async () => {
      fdForm.value.password = ''
      if (selectedMode.value === 'fd') {
        let { data: queryData } = await queryNetManagerNodeID();
        fdForm.value.nodeId = queryData?.NodeConnectID;

        let { data: certData } = await queryFdRootCert();
        setCertInfo(certData);
      }
    })

    const configFd = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        let params = {
          NodeId: fdForm.value.nodeId,
          ManagerType: 'FusionDirector',
          NetIP: fdForm.value.ipAddress,
          NetAccount: fdForm.value.account,
          NetPassword: fdForm.value.password,
          ServerName: fdForm.value.serverName,
          Port: parseInt(fdForm.value.port),
          test: fdForm.value.connectTest,
        }
        try {
          let { status } = await modifyNetManagerInfo(params);
          if (status === errorCode.NET_MANAGER.PARTIAL_SUCCESS) {
            AppMessage.success(t('registration.nms.configSuccessTip'))
          } else {
            AppMessage.success(t('registration.nms.switchSuccessTip'))
          }
          formElement.resetFields();
          await init();
        } catch (err) {
          fdForm.value.password = ''
          if (!err?.response?.data?.error) {
            return;
          }
          let status = err.response.data.error['@Message.ExtendedInfo'][0].Oem.status;
          if (Object.prototype.hasOwnProperty.call(errorTipMapper, status)) {
            showErrorAlert(t(errorTipMapper[status]))
          } else {
            showErrorAlert(t('registration.nms.switchFailedTip'))
          }
        }
      })
    }

    const configWeb = async () => {
      let params = {
        ManagerType: 'Web',
      }
      try {
        await modifyNetManagerInfo(params);
        AppMessage.success(t('registration.nms.switchSuccessTip'))
      } catch (err) {
        if (!err?.response?.data?.error) {
          return;
        }
        let status = err.response.data.error['@Message.ExtendedInfo'][0].Oem.status;
        if (status === errorCode.SESSION.ERROR_USER_NOT_MATCH_PASSWORD) {
          showErrorAlert(t('registration.nms.userNotMatchPassword'))
        } else {
          showErrorAlert(t('registration.nms.switchFailedTip'))
        }
      } finally {
        await init();
      }
    }

    const saveConfig = async () => {
      if (selectedMode.value === 'fd') {
        await configFd(fdFormRef.value);
      } else {
        await configWeb();
      }
    }

    const queryNetManagerConfig = async () => {
      let { data } = await queryNetManagerInfo();
      return data;
    }

    const getMode = (data) => data.toLowerCase().indexOf('web') === -1 ? 'fd' : 'web';

    const isWebMode = (data) => getMode(data) === 'web';

    const initCurrConfig = (data) => {
      currConfig.value.mode = data?.NetManager;
      currConfig.value.ip = data?.NetIP;
      currConfig.value.account = data?.NetAccount;
      let statusMapper = {
        not_configured: t('registration.nms.notConfig'),
        connecting: t('registration.nms.connecting'),
        connected: t('registration.nms.connected'),
        error_configured: t('registration.nms.errorConfig'),
      }
      currConfig.value.connectStatus = statusMapper[data?.ConnectStatus];
      selectedMode.value = getMode(currConfig.value.mode);
    }

    const initSelectedMode = (data) => {
      selectedMode.value = getMode(data?.NetManager);
    }

    const initFdForm = (data) => {
      fdForm.value.serverName = data?.ServerName;
      fdForm.value.ipAddress = data?.NetIP;
      fdForm.value.port = data?.Port;
      fdForm.value.account = data?.NetAccount;
    }

    const init = async () => {
      let data = await queryNetManagerConfig();
      initCurrConfig(data);
      initSelectedMode(data);
      initFdForm(data);
    }

    const beforeUploadCertFile = (rawFile) => {
      if (!isValidUploadFilename(rawFile.name)) {
        showErrorAlert(t('common.uploadFileNameTip'))
        return false
      }

      if (!/(\.crt)$/i.test(rawFile.name)) {
        showErrorAlert(t('registration.nms.crtFileErrorTip'))
        return false
      }
      if (rawFile.size > constants.FD_CERT_FILE_SIZE_MB * 1024 * 1024) {
        showErrorAlert(t('certificate.serverCert.formatErrorTip', { size: `${constants.FD_CERT_FILE_SIZE_MB} MB`}))
        return false
      }
      return true
    }

    const beforeUploadCrlFile = (rawFile) => {
      if (!isValidUploadFilename(rawFile.name)) {
        showErrorAlert(t('common.uploadFileNameTip'))
        return false
      }

      if (!/(\.crl)$/i.test(rawFile.name)) {
        showErrorAlert(t('registration.nms.crlFileErrorTip'))
        return false
      }
      if (rawFile.size > constants.FD_CRL_FILE_SIZE_KB * 1024) {
        showErrorAlert(t('certificate.serverCert.formatErrorTip', { size: `${constants.FD_CRL_FILE_SIZE_KB} KB`}))
        return false
      }
      return true
    }

    onMounted(async () => {
      await init();
    })

    const isEnglish = () => getLanguageDefaultChinese() !== 'zh'

    return {
      selectedMode,
      currConfig,
      uploadHeaders,
      certContent,
      certFileList,
      crlFileList,
      fdForm,
      fdFormRule,
      fdFormRef,
      registrationTip,
      nodeIdTip,
      rootCertTip,
      isWebMode,
      uploadCertSuccess,
      uploadCrlSuccess,
      uploadCertError,
      uploadCrlError,
      saveConfig,
      beforeUploadCertFile,
      beforeUploadCrlFile,
      isEnglish,
    }
  },
});
</script>

<style lang='scss' scoped>
.tab-title {
  margin-top: 10px;
}

.container {
  margin-top: 20px;
  padding: 24px;
  background: var(--el-bg-color-overlay);
  border-radius: 4px;
}

.title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 200;
  margin-bottom: 10px;
}

.content {
  font-size: 14px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
}

.info-block {
  height: 10vh;
}

.config-block {
  height: 62vh;
  overflow: auto;
}

.save-block {
  height: 5vh;
}

.info {
  display: flex;
  margin-top: 20px;
}
.info > div {
  flex: 1;
  text-align: left;
  padding-left: 10px;
}

.info > div {
  border-left: 1px solid var(--el-border-color);
}

.cert-content {
  width: 600px;
  word-break: break-word;
  border: 1px solid var(--el-border-color);
  padding: 10px;
  border-radius: 4px;
}

.cert-row {
  display: flex;
  align-items: baseline;
}

.cert-title {
  font-family: 'HarmonyOS_Sans_SC_Regular', serif;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
}

.select-mode {
  color: #D3DCE9;
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
}

.isEnglish {
  width: 660px !important;
}
</style>