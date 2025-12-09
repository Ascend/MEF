<template>
  <div class="title tab-title">{{ $t("certificate.menuName") }}</div>
  <el-card class="box-card" style="width: 700px;">
    <template #header>
      <div class="card-header">
        <span>{{ $t("certificate.serverCert.tabName") }}</span>
        <span>
          <el-button @click="downloadCSR">
            <img src="@/assets/img/manager/download.svg" alt="" style="height: 16px; width: 16px; margin-right: 4px;"/>
            {{ $t("certificate.serverCert.downloadCSR") }}
          </el-button>
          <el-button @click="isShowUploadForm = true">
            {{ $t("certificate.serverCert.customCertUpload") }}
          </el-button>
        </span>
      </div>
    </template>
    <div>
      <app-tip :tip-text="$t('certificate.serverCert.uploadCertTip')" />
      <app-tip v-show="isDefaultCert" tip-type="warn" :tip-text="$t('certificate.serverCert.prefabricatedCertTip')" />
      <el-row style="margin-top: 20px;">
        <el-col :span="6" class="medium-font cert-title">{{ $t("certificate.serverCert.issuer") }}</el-col>
        <el-col :span="18" class="medium-font cert-content">{{ certData?.Issuer }}</el-col>
      </el-row>
      <el-row>
        <el-col :span="6" class="medium-font cert-title">{{ $t("certificate.serverCert.user") }}</el-col>
        <el-col :span="18" class="medium-font cert-content">{{ certData?.Subject }}</el-col>
      </el-row>
      <el-row>
        <el-col :span="6" class="medium-font cert-title">{{ $t("certificate.serverCert.startTime") }}</el-col>
        <el-col :span="18" class="medium-font cert-content">{{ certData?.ValidNotBefore }}</el-col>
      </el-row>
      <el-row>
        <el-col :span="6" class="medium-font cert-title">{{ $t("certificate.serverCert.endTime") }}</el-col>
        <el-col :span="18" class="medium-font cert-content">{{ certData?.ValidNotAfter }}</el-col>
      </el-row>
      <el-row>
        <el-col :span="6" class="medium-font cert-title">{{ $t("certificate.serverCert.sn") }}</el-col>
        <el-col :span="18" class="medium-font cert-content">{{ certData?.SerialNumber }}</el-col>
      </el-row>
      <el-row>
        <el-col :span="6" class="medium-font cert-title">{{ $t("certificate.serverCert.fingerPrint") }}</el-col>
        <el-col :span="18" class="medium-font cert-content">{{ certData?.FingerPrint }}</el-col>
      </el-row>
    </div>
  </el-card>
  <upload-cert :is-show-upload-form="isShowUploadForm" @refresh="refresh" @cancelUpload="isShowUploadForm = false" />
</template>

<script>
import { defineComponent, ref, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus'

import UploadCert from '@/views/setting/safety/components/UploadCert.vue';
import { downloadCSRFile, queryHttpsCertInfo } from '@/api/certificate';
import AppTip from '@/components/AppTip.vue';
import {saveFile} from '@/utils/downloadFile';
import {handleOperationResponseError} from '@/utils/commonMethods';

export default defineComponent({
  name: 'CertInfo',
  components: {
    UploadCert,
    AppTip,
  },
  setup() {
    const { t } = useI18n()
    const certData = ref();
    const isShowUploadForm = ref(false);
    const isDownloadCSR = ref(false);

    const fetchCertData = async () => {
      let { data } = await queryHttpsCertInfo();
      return data?.X509CertificateInformation?.ServerCert;
    }

    let msgInstance;
    const checkIsExpire = (certResponse) => {
      if (!certResponse) {
        return;
      }

      if (certResponse?.ExpiredDayRemaining === null || certResponse?.ExpiredDayRemaining === undefined) {
        if (msgInstance) {
          msgInstance.close();
          msgInstance = null;
        }
        return;
      }

      let warningConfig = {
        type: 'warning',
        duration: 0,
        grouping: true,
        showClose: true,
        offset: 60,
      }
      if (certResponse?.HttpsCertEable === 'Certificate has not expired') {
        let expiredDay = parseFloat(certResponse?.ExpiredDayRemaining, 10);
        if (expiredDay === 0) {
          msgInstance = ElMessage ({
            message: t('certificate.serverCert.expiredToday'),
            ...warningConfig,
          })
        } else if (expiredDay > 0 && expiredDay <= 90) {
          msgInstance = ElMessage ({
            message: t('certificate.serverCert.willTimeout', { 'day': certResponse?.ExpiredDayRemaining }),
            ...warningConfig,
          })
        }
      } else if (certResponse?.HttpsCertEable === 'Certificate has expired') {
        msgInstance = ElMessage ({
          message: t('certificate.serverCert.alreadyTimeout', { 'day': certResponse?.ExpiredDayRemaining }),
          ...warningConfig,
        })
      } else if (certResponse?.HttpsCertEable === 'Certificate has not effected') {
        msgInstance = ElMessage ({
          message: t('certificate.serverCert.notEffect', { 'day': certResponse?.ExpiredDayRemaining }),
          ...warningConfig,
        })
      }
    }

    const refresh = async () => {
      let ServerCert = await fetchCertData();
      if (!ServerCert) {
        return;
      }
      checkIsDefaultCert(ServerCert);
      certData.value = ServerCert;
      checkIsExpire(ServerCert);
    }

    const downloadCSR = async () => {
      isDownloadCSR.value = true;

      // 下载csr文件
      try {
        let resp = await downloadCSRFile();
        let fileName = `X509_pem_${new Date().getTime()}.csr`;
        saveFile(resp.data, fileName);
      } catch (err) {
        handleOperationResponseError(err, t('certificate.serverCert.downloadErrorTip'))
      }
    }

    onMounted(async () => {
      await refresh();
    })

    onBeforeUnmount(() => {
      if (msgInstance) {
        msgInstance.close();
      }
    })
    
    const isDefaultCert = ref(false);
    
    const checkIsDefaultCert = (serverCert) => {
      if (!serverCert) {
        return;
      }
      if (checkIssuer(serverCert?.Issuer) && checkSubject(serverCert?.Subject)) {
        isDefaultCert.value = true;
      }
    }
    
    const checkIssuer = (issuerStr) => {
      if (!issuerStr) {
        return false;
      }
      
      const issuer = parsingStrings(issuerStr)
      
      return issuer?.C === 'CN' && issuer?.O === 'Huawei' && issuer?.CN === 'Huawei IT Product CA';
    }
    
    const checkSubject = (subjectStr) => {
      if (!subjectStr) {
        return false;
      }
      
      const subject = parsingStrings(subjectStr)
      
      return subject?.C === 'CN' && subject?.O === 'Huawei'
          && (
              (subject?.OU === 'CPL Ascend' && subject?.CN.startsWith('MindXOM-'))
              || (subject?.OU === 'Ascend' && subject?.CN === 'huawei')
          );
    }
    
    const parsingStrings = (str) => {
      const dict = {};
      if (!str.includes(', ')) {
        return dict;
      }
      const pairs = str.split(', ');
      pairs.forEach(pair => {
        if (pair.includes('=')) {
          const [key, value] = pair.split('=');
          dict[key] = value;
        }
      });
      return dict;
    }

    return {
      certData,
      isDownloadCSR,
      isShowUploadForm,
      isDefaultCert,
      refresh,
      downloadCSR,
    }
  },
});
</script>

<style lang="scss" scoped>
.el-card {
  color: var(--el-text-color-regular)
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.box-card {
  margin-top: 20px;
  width: 42%;
}

.el-row {
  margin-bottom: 20px;
}

.title {
  margin-top: 10px;
}

.cert-title {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 500;
}

.cert-content {
  font-size: 14px;
  color: #D3DCE9;
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 500;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
</style>