<template>
  <el-button @click='exportWeakDict' style="display: flex; align-items: center;">
    <img style="width: 14px; height: 14px; margin-right: 4px;" alt='' src='@/assets/img/upload.svg' />
    {{ $t('safety.weakPassword.export') }}
  </el-button>
  <div class="container " style='height: 64vh; margin-top: 20px;' >
    <div class='basic-info'>{{ $t('common.basicInfo') }}</div>
    <app-tip :tip-type='"info"' :tip-text='$t("time.time.saveTip")' />
    <el-form
      :model='form'
      :rules='resetFormRule'
      ref='formRef'
      label-position='left'
      label-width="auto"
      style='margin-top: 20px; width: 800px;'
      :hide-required-asterisk="true"
    >
      <el-form-item prop='isRetain' :label="$t('safety.weakPassword.selectOp')" class="not-required">
        <el-radio-group v-model="form.op" @change="onOpChange">
          <el-radio :label="ops.import" style="margin-right: 100px;">
            {{ $t("safety.weakPassword.import") }}
          </el-radio>
          <el-radio :label="ops.delete" style="margin-right: 10px;">
            {{ $t("safety.weakPassword.delete") }}
          </el-radio>
        </el-radio-group>
        <div v-show="isImportOp">
          <div class="reset-tip" style="margin-top: 20px;">
            <img src="@/assets/img/home/information.svg" alt=""/>
            {{ $t("safety.weakPassword.importTip1") }}
          </div>
          <div class="reset-tip">
            <img src="@/assets/img/home/information.svg" alt=""/>
            {{ $t("safety.weakPassword.importTip2") }}
          </div>
          <div class="reset-tip">
            <img src="@/assets/img/home/information.svg" alt=""/>
            {{ $t("safety.weakPassword.importDictTip") }}
          </div>
        </div>
        <div v-show="isDeleteOp">
          <div class="reset-tip" style="margin-top: 20px;">
            <img src="@/assets/img/home/information.svg" alt=""/>
            {{ $t("safety.weakPassword.deleteTip1") }}
          </div>
          <div class="reset-tip">
            <img src="@/assets/img/home/information.svg" alt=""/>
            {{ $t("safety.weakPassword.deleteTip2") }}
          </div>
        </div>
      </el-form-item>
      <el-form-item prop='fileName' v-if="isImportOp" :label="$t('safety.weakPassword.importDict')" class="required">
        <el-upload
          class="upload"
          action="/redfish/v1/UpdateService/FirmwareInventory"
          name="imgfile"
          ref="upload"
          :with-credentials="true"
          accept=".conf"
          :headers="uploadHeaders"
          :before-upload="beforeWeakDictUpload"
          :on-success="onUploadSuccess"
          :on-error="onUploadError"
          :show-file-list="false"
          :auto-upload="false"
          :on-change="onUploadChange"
          v-model:file-list="fileList"
        >
          <el-input style="width: 100%;" v-model="form.fileName" :placeholder="$t('common.pleaseSelect')">
            <template #append>...</template>
          </el-input>
        </el-upload>
      </el-form-item>
      <el-form-item prop='password' :label="$t('safety.loginRules.password')" class="required">
        <el-input
          style='width: 300px;'
          v-model='form.password'
          :placeholder='$t("common.pleaseInput")'
          type='password'
          @keyup.enter='submit(formRef)'
          autocomplete='off'
          @copy.native.capture.prevent='()=>{}'
          @paste.native.capture.prevent='()=>{}'
          @cut.native.capture.prevent='()=>{}'
        />
      </el-form-item>
    </el-form>
    <el-button class='save-btn' type="primary" @click='submit(formRef)'>
      {{ $t('common.save') }}
    </el-button>
  </div>
</template>

<script>
import { defineComponent, ref, reactive, computed } from 'vue';
import { useI18n } from 'vue-i18n'

import {
  AppMessage, getAuthToken, showErrorAlert, handleUploadError,
  isValidUploadFilename, handleOperationResponseError
} from '@/utils/commonMethods';
import constants from '@/utils/constants';
import { deletePunyDict, exportPunyDict, importPunyDict } from '@/api/safety';
import AppTip from '@/components/AppTip.vue';
import { saveFile } from '@/utils/downloadFile';

export default defineComponent({
  name: 'SettingReset',
  components: {
    AppTip,
  },
  setup() {
    const { t } = useI18n()
    const uploadHeaders = ref({
      ...getAuthToken(),
      'AutoRefresh': false,
    })
    const formRef = ref();
    const upload = ref();
    const ops = {
      import: 'import',
      delete: 'delete',
    }
    const form = reactive({
      op: ops.import,
      fileName: '',
      ipAddress: null,
      password: null,
    })
    const isImportOp = computed(() => form.op === ops.import)
    const isDeleteOp = computed(() => form.op === ops.delete)
    const onOpChange = () => {
      form.password = '';
      formRef.value.resetFields();
    }

    const fileList=ref([]);
    const fileSizeLimitBytes = constants.MAX_PENNY_DICT_FILE_SIZE_MB * 1024 * 1024;
    const beforeWeakDictUpload = (rawFile) => {
      if (!isValidUploadFilename(rawFile.name)) {
        showErrorAlert(t('common.uploadFileNameTip'))
        formRef.value.resetFields();
        return false
      }

      if (rawFile.name.slice(rawFile.name.length - 5) !== '.conf') {
        showErrorAlert(t('safety.weakPassword.confFileErrorTip'));
        formRef.value.resetFields();
        return false;
      }
      if (rawFile.size > fileSizeLimitBytes) {
        showErrorAlert(t('safety.weakPassword.importErrorTip'));
        formRef.value.resetFields();
        return false;
      }
      return true;
    }
    const onUploadSuccess = async () => {
      try {
        await importPunyDict({
          FileName: form.fileName,
          Password: form.password,
        });
        AppMessage.success(t('safety.weakPassword.importSuccess'));
      } catch (err) {
        handleOperationResponseError(err, t('safety.weakPassword.importFailed'));
      } finally {
        resetUploadForm();
      }
    }

    const onUploadError = (error) => {
      handleUploadError(error);
      resetUploadForm();
    }

    const onUploadChange = (file) => {
      if (file.status === 'ready') {
        form.fileName = file.name;
      }
    }

    const resetFormRule = ref({
      fileName: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'change',
        },
      ],
      password: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
    })

    const submitImport = async (formRef_) => {
      if (!formRef_) {
        return;
      }
      await formRef_.validate(async (valid) => {
        if (!valid) {
          return;
        }
        fileList.value.splice(0, fileList.value.length - 1);
        upload.value.submit();
      });
    };

    const submitDelete = async (formRef_) => {
      if (!formRef_) {
        return;
      }
      await formRef_.validate(async (valid) => {
        if (!valid) {
          return;
        }
        try {
          await deletePunyDict({
            Password: form.password,
          });
          AppMessage.success(t('safety.weakPassword.deleteSuccess'));
        } catch (err) {
          handleOperationResponseError(err, t('safety.weakPassword.deleteFailed'));
        } finally {
          formRef_.resetFields()
        }
      });
    };

    const exportWeakDict = async () => {
      try {
        let resp = await exportPunyDict();
        let fileName = `WeakPas_${new Date().getTime()}.conf`;
        saveFile(resp.data, fileName);
      } catch (err) {
        handleOperationResponseError(err, t('safety.weakPassword.exportFailed'))
      }
    }

    const submit = async (formRef_) => {
      switch (form.op) {
        case ops.import:
          await submitImport(formRef_);
          break;
        case ops.delete:
          await submitDelete(formRef_);
          break;
        default:
          break;
      }
    }

    const resetUploadForm = () => {
      fileList.value = [];
      formRef.value.resetFields();
      upload.value.clearFiles();
    };

    return {
      ops,
      isImportOp,
      isDeleteOp,
      formRef,
      form,
      uploadHeaders,
      onOpChange,
      beforeWeakDictUpload,
      fileList,
      onUploadSuccess,
      onUploadError,
      resetFormRule,
      submit,
      exportWeakDict,
      upload,
      onUploadChange,
    }
  },
});
</script>

<style lang="scss" scoped>
.container {
  background: var(--el-bg-color-overlay);
  border-radius: 4px;
  padding: 24px 48px;
}

.basic-info {
  font-size: 16px;
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
  margin-bottom: 8px;
}

.save-btn {
  padding: 12px 24px;
  margin-top: 40px;
}

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
  width: 720px;
}

.upload {
  width: 300px;
  :deep(.el-upload) {
    display: block;
  }
}

</style>
