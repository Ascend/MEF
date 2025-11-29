<template>
  <el-dialog
    :modelValue="props.isShowUploadForm"
    :title="$t('certificate.serverCert.customCertUpload')"
    :close-on-click-modal="false"
    style="width: 500px;"
    @close="cancelUpload"
  >
    <app-tip :tip-type="'info'" :tip-text="$t('certificate.serverCert.uploadTip')" />
    <el-form
      ref="uploadFormRef"
      label-position="left"
      label-width="180px"
      :model="uploadForm"
      :rules="uploadFormRule"
      style="width: 400px; "
      class="add-ip-form"
    >
      <el-form-item :label="$t('certificate.serverCert.file')" prop="FileName">
        <el-upload
          class="upload-demo"
          action="/redfish/v1/UpdateService/FirmwareInventory"
          name="imgfile"
          :with-credentials="true"
          accept=".crt, .cer"
          :headers="uploadHeaders"
          :on-success="uploadSuccess"
          :on-error="uploadError"
          :show-file-list="false"
          v-model:file-list="fileList"
          :before-upload="beforeUploadFile"
          :auto-upload="false"
          ref="upload"
          :on-change="onUploadChange"
        >
          <el-input v-model="uploadForm.FileName" :placeholder="$t('common.pleaseSelect')">
            <template #append>...</template>
          </el-input>
        </el-upload>
      </el-form-item>
      <el-form-item :label="$t('certificate.serverCert.password')" prop="Password">
        <el-input
            v-model="uploadForm.Password"
            :placeholder="$t('common.pleaseInput')"
            type="Password"
            @keyup.enter='confirmUpload(uploadFormRef)'
            autocomplete="off"
            @copy.native.capture.prevent="()=>{}"
            @paste.native.capture.prevent="()=>{}"
            @cut.native.capture.prevent="()=>{}"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="confirmUpload(uploadFormRef)">
          {{ $t("common.confirm") }}
        </el-button>
        <el-button @click="cancelUpload">
          {{ $t("common.cancel") }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import { ref, defineComponent, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus'

import { importServerCertificate } from '@/api/certificate';
import errorCode from '@/api/errorCode';
import AppTip from '@/components/AppTip.vue';
import {
  getAuthToken, AppMessage, showErrorAlert,
  handleUploadError, isValidUploadFilename
} from '@/utils/commonMethods';
import constants from '@/utils/constants';
import errorTipMapper from '@/api/errorTipMapper';

export default defineComponent({
  name: 'UploadCert',
  components: {
    AppTip,
  },
  props: {
    isShowUploadForm: Boolean,
  },
  setup(props, ctx) {
    const { t } = useI18n()
    const uploadHeaders = ref({
      ...getAuthToken(),
      'AutoRefresh': false,
    })
    const uploadFormRef = ref()
    const fileList = ref([])
    const upload = ref();

    let uploadForm = ref({
      FileName: '',
      Password: null,
    })

    const uploadFormRule = reactive({
      FileName: [
        {
          required: true,
          message: t('common.pleaseUpload'),
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

    const uploadSuccess = async () => {
      try {
        let { status } = await importServerCertificate(uploadForm.value);
        if (status === errorCode.CERT.SUCCESS_BUT_NOT_SAFE) {
          msgInstance = ElMessage({
            message: t('certificate.serverCert.importSuccessButNotSafeTip'),
            type: 'warning',
            duration: 0,
            grouping: true,
            showClose: true,
            offset: 60,
          })
        } else {
          if (msgInstance) {
            msgInstance.close();
            msgInstance = null;
          }
          AppMessage.success(t('certificate.serverCert.importSuccessTip'))
        }
        ctx.emit('cancelUpload')
        ctx.emit('refresh')
      } catch (err) {
        let status = err?.response?.data?.error['@Message.ExtendedInfo'][0]?.Oem?.status;
        if (Object.prototype.hasOwnProperty.call(errorTipMapper, status)) {
          showErrorAlert(t(errorTipMapper[status]))
        } else if (status === errorCode.SESSION.ERROR_USER_LOCK_STATE) {
          let msg = err?.response?.data?.error['@Message.ExtendedInfo'][0]?.Message;
          showErrorAlert(t('session.lockTip', {seconds: msg.split('[')[1].split(']')[0]}))
        } else {
          showErrorAlert(t('common.uploadErrorTip'))
        }
        ctx.emit('cancelUpload')
      } finally {
        resetUploadForm();
      }
    }

    const uploadError = (error) => {
      handleUploadError(error)
      cancelUpload();
    }

    const onUploadChange = (file) => {
      if (file.status === 'ready') {
        uploadForm.value.FileName = file.name;
      }
    }

    const cancelUpload = () =>{
      ctx.emit('cancelUpload')
      resetUploadForm();
    }

    let msgInstance;
    const confirmUpload = async (formElement) =>{
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        fileList.value.splice(0, fileList.value.length - 1);
        upload.value.submit();
      })
    }

    const beforeUploadFile = (rawFile) => {
      if (!isValidUploadFilename(rawFile.name)) {
        showErrorAlert(t('common.uploadFileNameTip'))
        uploadFormRef.value.resetFields();
        return false
      }
      if (!/(\.cer|\.crt)$/i.test(rawFile.name)) {
        showErrorAlert(t('certificate.serverCert.uploadTip') + t('common.uploadAgainFileTip'))
        uploadFormRef.value.resetFields();
        return false
      }
      if (rawFile.size > constants.CERT_FILE_SIZE) {
        showErrorAlert(t('certificate.serverCert.formatErrorTip', { size: `${constants.CERT_FILE_SIZE_KB} KB`}))
        uploadFormRef.value.resetFields();
        return false
      }
      return true
    }

    const resetUploadForm = () => {
      fileList.value = [];
      uploadFormRef.value.resetFields();
      upload.value.clearFiles();
      uploadForm.value.FileName = '';
      uploadForm.value.Password = '';
    };

    return {
      props,
      uploadHeaders,
      fileList,
      uploadFormRef,
      uploadForm,
      uploadFormRule,
      cancelUpload,
      confirmUpload,
      uploadSuccess,
      uploadError,
      beforeUploadFile,
      upload,
      onUploadChange,
    }
  },
})
</script>

<style scoped>
</style>