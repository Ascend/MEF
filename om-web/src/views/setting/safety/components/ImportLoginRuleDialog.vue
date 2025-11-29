<template>
  <el-dialog
    :title='title'
    :modelValue='isShow'
    :width='500'
    @close="onCancel"
    align-center
  >
    <app-tip :tip-text='$t("safety.loginRules.importTip")'/>
    <br>
    <el-form
      ref='formRef'
      label-position='left'
      label-width="auto"
      :model='form'
      :rules='formRules'
      :hide-required-asterisk="true"
    >
      <el-form-item :label='$t("safety.loginRules.importRuleFile")' class='required' prop='fileName'>
        <el-upload
          class="upload"
          action="/redfish/v1/UpdateService/FirmwareInventory"
          name="imgfile"
          :with-credentials="true"
          accept=".ini"
          :headers="uploadHeaders"
          :before-upload="beforeLoginRuleUpload"
          :on-success="onUploadSuccess"
          :on-error="onUploadError"
          :show-file-list="false"
          :auto-upload="false"
          :on-change="onUploadChange"
          ref="upload"
          v-model:file-list="fileList"
        >
          <el-input style="width: 100%;" v-model="form.fileName" :placeholder="$t('common.pleaseSelect')">
            <template #append>...</template>
          </el-input>
        </el-upload>
      </el-form-item>
      <el-form-item :label='$t("safety.loginRules.password")' class='required' prop='password'>
        <el-input v-model='form.password'
                  type="password"
                  @keyup.enter='onConfirm(formRef)'
                  autocomplete="off"
                  @copy.native.capture.prevent="()=>{}"
                  @paste.native.capture.prevent="()=>{}"
                  @cut.native.capture.prevent="()=>{}"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="onConfirm(formRef)">
          {{ $t("common.confirm") }}
        </el-button>
        <el-button @click="onCancel(formRef)">
          {{ $t("common.cancel") }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import AppTip from '@/components/AppTip.vue';
import { useI18n } from 'vue-i18n';
import { reactive, ref } from 'vue';
import {
  AppMessage, getAuthToken, showErrorAlert,
  handleUploadError, isValidUploadFilename, handleOperationResponseError
} from '@/utils/commonMethods';
import { importLoginRules } from '@/api/safety';
import constants from '@/utils/constants';

export default {
  name: 'ImportLoginRuleDialog',
  components: { AppTip },
  props: {
    isShow: Boolean,
    title: String,
  },
  setup(props, ctx) {
    const events = {
      submit: 'submit',
      cancel: 'cancel',
    };
    const uploadHeaders = ref({
      ...getAuthToken(),
      'AutoRefresh': false,
    })
    const { t } = useI18n();
    const formRef = ref();
    const upload = ref();
    const form = reactive({
      fileName: '',
      password: null,
    });
    const formRules = ref({
      fileName: [
        {
          required: true,
          message:  t('common.pleaseUpload'),
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
    });

    const fileSizeLimitBytes = constants.MAX_LOGIN_RULE_FILE_SIZE_KB * 1024;
    const beforeLoginRuleUpload = (rawFile) => {
      if (!isValidUploadFilename(rawFile.name)) {
        showErrorAlert(t('common.uploadFileNameTip'))
        formRef.value.resetFields();
        return false
      }

      if (!/(\.ini)$/i.test(rawFile.name)) {
        showErrorAlert(t('safety.loginRules.uploadTip'))
        formRef.value.resetFields();
        return false
      }

      if (rawFile.size > fileSizeLimitBytes) {
        showErrorAlert(t('safety.loginRules.importTip'));
        formRef.value.resetFields();
        return false;
      }
      return true;
    }
    const fileList=ref([]);
    const onUploadSuccess = async () => {
      try {
        await importLoginRules({
          'Password': form.password,
          'file_name': form.fileName,
        });
        AppMessage.success(t('safety.loginRules.importSuccessTip'));
        ctx.emit(events.submit);
      } catch (err) {
        handleOperationResponseError(err, t('safety.loginRules.importFailedTip'))
        ctx.emit(events.cancel);
      }
    }
    const onUploadError = (error) => {
      handleUploadError(error)
      fileList.value.length = 0;
      upload.value.clearFiles();
      ctx.emit(events.cancel);
    }

    const onUploadChange = (file) => {
      if (file.status === 'ready') {
        form.fileName = file.name;
      }
    }

    const onConfirm = async (formRef_) => {
      if (!formRef_) {
        return
      }
      await formRef_.validate(async (valid) => {
        if (!valid) {
          return
        }

        fileList.value.splice(0, fileList.value.length - 1);
        upload.value.submit();
      });
    }

    const onCancel = async () => {
      ctx.emit(events.cancel);
    }

    return {
      form,
      formRules,
      formRef,
      uploadHeaders,
      fileList,
      beforeLoginRuleUpload,
      onUploadSuccess,
      onUploadError,
      onConfirm,
      onCancel,
      onUploadChange,
      upload,
    };
  },
};
</script>

<style lang="scss" scoped>
.upload {
  width: 100%;
  :deep(.el-upload) {
    display: block;
  }
}

</style>