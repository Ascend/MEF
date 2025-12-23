<template>
  <div class='container' v-loading="loading">
    <div class='title'>{{ $t('update.systemUpgrade.updateInfo') }}</div>
    <app-tip :tip-type='"info"' :tip-list='tipList' />
    <app-tip :tip-type='"warn"' :tip-text='$t("update.systemUpgrade.updateProcessTip")' />
    <el-form
      :model='updateForm'
      label-position='right'
      label-width='150px'
      style='width: 96%; margin-top: 20px;'
      ref='updateFormRef'
      :rules='updateFormRule'
    >
      <el-form-item :label='$t("update.systemUpgrade.updateFile")' prop='updateFile'>
        <el-upload
          class='upload-demo'
          action='/redfish/v1/UpdateService/FirmwareInventory'
          name='imgfile'
          :with-credentials='true'
          accept='.zip'
          :headers='uploadHeaders'
          :on-success='uploadSuccess'
          :on-error='uploadError'
          :show-file-list='false'
          :before-upload="beforeUploadFile"
          :auto-upload="false"
          :disabled="isUpgrading"
          ref="upload"
          :on-change="onUploadChange"
          v-model:file-list='fileList'
        >
          <el-input
            v-model='updateForm.updateFile'
            style='width: 400px;'
            :placeholder='$t("common.pleaseSelect")'
            :disabled="isUpgrading"
          >
            <template #append>...</template>
          </el-input>
        </el-upload>
        <el-button type='primary' style='margin-left: 20px;' @click='clickToUpdate' :disabled="isUpgrading">
          {{ $t('update.systemUpgrade.clickToUpdate') }}
        </el-button>
      </el-form-item>
      <el-form-item :label='$t("update.systemUpgrade.updateInfo")'>
        <el-table
          :data="upgradeInfo"
          style="width: 100%;"
        >
          <el-table-column :width="240" :label="$t('update.systemUpgrade.currFirmware')" prop="Module" />
          <el-table-column :width="240" :label="$t('update.systemUpgrade.currVersion')" prop="Version" />
          <el-table-column :label="$t('update.common.packageName')" prop="packageName" />
          <el-table-column :width="240" :label="$t('update.systemUpgrade.updatePackageVersion')" prop="InactiveVersion" />
          <el-table-column :width="240" :label="$t('update.systemUpgrade.updateProgress')" prop="UpgradeProcess" >
            <template #default="scope">
              <el-progress :percentage="scope.row.UpgradeProcess" :status="getUpgradeStatus(scope.row.UpgradeResult)" />
            </template>
          </el-table-column>
        </el-table>
        <el-button type='primary' style='margin-top: 20px;' @click="clickTakeEffect" :disabled="!isTakingEffect">
          {{ $t('update.common.clickToTakeEffect') }}
        </el-button>
      </el-form-item>
    </el-form>
  </div>
  <el-dialog
    v-model="isShowTakeEffectTip"
    :title="$t('common.tips')"
    :width="600"
  >
    <div style="display: flex; align-items: center; word-break: break-word;">
      <img src="@/assets/img/alarm.svg" alt="" style="margin-right: 10px;"/>
        {{ $t('update.systemUpgrade.takeEffectTip') }}
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="confirmTakeEffect">
          {{ $t("common.confirm") }}
        </el-button>
        <el-button @click="cancelTakeEffect">
          {{ $t("common.cancel") }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import { defineComponent, onMounted, reactive, ref, watch, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import AppTip from '@/components/AppTip.vue';
import { querySystemsSourceInfo } from '@/api/information';
import { queryUpdateStatus, updateFirmware, resetFirmware } from '@/api/update';
import constants from '@/utils/constants';
import {
  AppMessage, showErrorAlert, clearSessionAndJumpToLogin,
  handleUploadError, isValidUploadFilename, getAuthToken
} from '@/utils/commonMethods';
import errorTipMapper from '@/api/errorTipMapper';

export default defineComponent({
  name: 'SystemUpdate',
  components: {
    AppTip,
  },
  setup() {
    const { t } = useI18n()
    const uploadHeaders = ref({
      ...getAuthToken(),
      'AutoRefresh': false,
    })
    const router = useRouter();
    const fileList = ref([])
    const upload = ref();
    let isUpgrading = ref(false);
    const isTakingEffect = ref(false);
    const tipList = [
      t('update.systemUpgrade.formatTip'),
      t('update.systemUpgrade.effectTip'),
    ]
    const isShowTakeEffectTip = ref(false);
    const updateFormRef = ref()
    const updateForm = ref({
      updateFile: '',
    })

    const updateFormRule = reactive({
      updateFile: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
    })

    const upgradeInfo = ref([])

    const getUpgradeStatus = (status) => {
      let upgradeResultMapper = {
        true: 'success',
        false: 'exception',
        null: '',
      }
      return upgradeResultMapper[status]
    }

    const loading = ref(false)
    const beforeUploadFile = (rawFile) => {
      if (!isValidUploadFilename(rawFile.name)) {
        showErrorAlert(t('common.uploadFileNameTip'))
        clearUpload();
        return false
      }

      if (rawFile.name.slice(-4).toLowerCase() !== '.zip') {
        showErrorAlert(t('update.systemUpgrade.fileTypeErrorTip'))
        clearUpload();
        return false
      }
      if (rawFile.size > constants.MAX_UPGRADE_FILE_SIZE_BYTE) {
        showErrorAlert(t('update.systemUpgrade.fileSizeErrorTip'))
        clearUpload();
        return false
      }
      loading.value = true;
      return true
    }

    let SimpleUpdateErrorTipMapper = {
      'ERR.01': t('update.systemUpgrade.upgradeConflictErrorTip'),
      'ERR.02': t('update.systemUpgrade.verifyFailedTip'),
      'ERR.03': t('update.systemUpgrade.createFileFailedTip'),
      'ERR.04': t('update.systemUpgrade.upgradeOMErrorTip'),
      'ERR.06': t('update.systemUpgrade.upgradeOSErrorTip'),
      'ERR.07': t('update.systemUpgrade.upgradeMEFErrorTip'),
      'ERR.08': t('update.systemUpgrade.upgradeNpuErrorTip'),
      'ERR.011': t('update.systemUpgrade.fileNotExistErrorTip'),
      'ERR.012': t('update.systemUpgrade.decompressFailedErrorTip'),
      'ERR.010': t('update.systemUpgrade.deleteFileFailedTip'),
      'ERR.013': t('update.systemUpgrade.paramsErrorTip'),
      'ERR.014': t('update.systemUpgrade.upgradingTip'),
      'ERR.015': t('update.systemUpgrade.syncFileFailedTip'),
      'ERR.016': t('update.systemUpgrade.upgradeToolboxFailedTip'),
      'ERR.017': t('update.systemUpgrade.sdkUpgradeFailed'),
      'ERR.018': t('update.systemUpgrade.cannotUpgradeToRc1'),
      'ERR.0555': t('update.systemUpgrade.mefIsStarting'),
    }

    const uploadSuccess = async () => {
      let params = {
        ImageURI: updateForm.value.updateFile,
        TransferProtocol: 'https',
      }

      clearIntervals(operateType.updating);
      loading.value = false;
      isUpgrading.value = true;

      try {
        let { data } = await updateFirmware(params, false);
        if (data?.upgradeState !== 'Failed') {
          AppMessage.success(t('update.common.startUpdate'))
          updatingIntervalId = await updateProgress(operateType.updating);
        } else {
          clearUpload();
        }
      } catch (err) {
        let status = err?.response?.data?.error['@Message.ExtendedInfo'][0]?.Oem?.status;
        if (Object.prototype.hasOwnProperty.call(errorTipMapper, status)) {
          showErrorAlert(t(errorTipMapper[status]))
          return
        }

        let msg = err?.response?.data?.error['@Message.ExtendedInfo'][0]?.Message;
        handleSimpleUpdateError(msg);
        isUpgrading.value = false;
        clearIntervals(operateType.updating)
        clearUpload();
        await queryUpdateStatus(false);
      }
    }

    const uploadError = (error) => {
      loading.value = false;
      handleUploadError(error);
      fileList.value.length = 0;
      clearUpload();
    }

    const onUploadChange = (file) => {
      if (file.status === 'ready') {
        updateForm.value.updateFile = file.name;
      }
    }

    const getSystemVersion = async () => {
      let { data } = await querySystemsSourceInfo(false);
      return data?.Oem;
    }

    const initUpgradeInfoTable = (data) => {
      upgradeInfo.value = data?.Firmware
    }

    let initUpdateIntervalId;
    onMounted(async () => {
      initUpgradeInfoTable(await getSystemVersion());
      initUpdateIntervalId = await updateProgress(operateType.init);
    })

    watch(
      upgradeInfo,
      () => {
        for (let i = 0; i < upgradeInfo.value?.length; i++) {
          if (upgradeInfo.value[i].UpgradeResult) {
            isTakingEffect.value = true
            return
          }
        }
        isTakingEffect.value = false
      },
      { deep: true }
    )

    const clickToUpdate = async () => {
      if (!updateForm.value.updateFile) {
        showErrorAlert(t('common.pleaseUpload'))
        return
      }
      fileList.value.splice(0, fileList.value.length - 1);
      upload.value.submit();
    }

    const updateFirmwareStatus = (data) => {
      isUpgrading.value = true;
      for (let i = 0; i < upgradeInfo.value?.length; i++) {
        if (upgradeInfo.value[i].Module === data?.Module) {
          upgradeInfo.value[i].UpgradeProcess = data?.PercentComplete;
          upgradeInfo.value[i].InactiveVersion = data?.Version;
          upgradeInfo.value[i].packageName = updateForm.value.updateFile;
          if (data?.TaskState === 'Success') {
            upgradeInfo.value[i].UpgradeResult = true
          } else if (data?.TaskState === 'Failed') {
            upgradeInfo.value[i].UpgradeResult = false
          } else if (data?.TaskState === 'Running' || data?.TaskState === 'New') {
            upgradeInfo.value[i].UpgradeResult = null
          }
        }
      }
    }

    const handleSimpleUpdateError = (msg) => {
      let errorNum = msg.split('-')[0];
      if (Object.prototype.hasOwnProperty.call(SimpleUpdateErrorTipMapper, errorNum)) {
        showErrorAlert(SimpleUpdateErrorTipMapper[errorNum])
      } else if (msg.indexOf('ERR.0-1') !== -1) {
        showErrorAlert(t('update.systemUpgrade.notUpgradeErrorTip'))
      } else {
        showErrorAlert(t('update.systemUpgrade.upgradeFailed'))
      }
    }

    let updatingIntervalId;
    const updateProgress = async (operate) => {
      let { data } = await queryUpdateStatus(false);
      updateFirmwareStatus(data)

      if (data?.TaskState === 'Running') {
        isUpgrading.value = true
        return setInterval(intervalUpdateProgressFunc, 5000, operate)
      } else if (data?.TaskState === 'Success') {
        isUpgrading.value = false;
        AppMessage.success(t('update.systemUpgrade.upgradeSuccessfully'))
        clearUpload();
      } else if (data?.TaskState === 'Failed') {
        isUpgrading.value = false;
        let msg = data?.Messages?.upgradeState;
        handleSimpleUpdateError(msg)
        clearUpload();
      } else if (data?.TaskState === 'New') {
        isUpgrading.value = false;
      }
      return null;
    }

    const clearUpload = () => {
      updateForm.value.updateFile = ''
      upload.value?.clearFiles();
    }

    const operateType = {
      init: 'init',
      updating: 'updating',
    }

    const clearIntervals = (operate) => {
      if (operate === operateType.init && initUpdateIntervalId) {
        clearInterval(initUpdateIntervalId);
        initUpdateIntervalId = null;
      }
      if (operate === operateType.updating && updatingIntervalId) {
        clearInterval(updatingIntervalId);
        updatingIntervalId = null;
      }
    }

    const intervalUpdateProgressFunc = async (operate) => {
      let { data } = await queryUpdateStatus(false);
      if (data?.Module) {
        updateFirmwareStatus(data)
      } else {
        isUpgrading.value = false;
        clearIntervals(operate);
      }

      if (data?.TaskState === 'Success') {
        isUpgrading.value = false;
        AppMessage.success(t('update.systemUpgrade.upgradeSuccessfully'))
        clearIntervals(operate);
        clearUpload()
      } else if (data?.TaskState === 'Failed') {
        isUpgrading.value = false;
        let msg = data?.Messages?.upgradeState;
        handleSimpleUpdateError(msg);
        clearIntervals(operate);
        clearUpload();
      }
    }

    const clickTakeEffect = () => {
      isShowTakeEffectTip.value = true
    }

    const confirmTakeEffect = async () => {
      try {
        await resetFirmware({ ResetType: 'GracefulRestart' })
        clearSessionAndJumpToLogin(router);
      } catch (err) {
        let msg = err?.response?.data?.error['@Message.ExtendedInfo'][0]?.Message;
        let effectErrorTipMapper = {
          'ERR.02': t('update.systemUpgrade.invalidParamsErrorTip'),
          'ERR.03': t('update.systemUpgrade.upgradeBusyErrorTip'),
          'ERR.04': t('update.systemUpgrade.upgradeFailedNotEffectErrorTip'),
          'ERR.05': t('update.systemUpgrade.migrateDbFailedErrorTip'),
          'ERR.06': t('update.systemUpgrade.backupLogErrorTip'),
          'ERR.07': t('update.systemUpgrade.registerOmFilesFailedErrorTip'),
          'ERR.08': t('update.systemUpgrade.effectMefFailedErrorTip'),
        }
        let errorNum = msg.split('-')[0];
        if (Object.prototype.hasOwnProperty.call(effectErrorTipMapper, errorNum)) {
          showErrorAlert(t(effectErrorTipMapper[errorNum]))
        } else if (msg.indexOf('ERR.01') !== -1) {
          showErrorAlert(t('update.systemUpgrade.effectBusyErrorTip'))
        } else {
          showErrorAlert(t('common.operationErrorTip'))
        }
        cancelTakeEffect();
      }
    }

    const cancelTakeEffect = () => {
      isShowTakeEffectTip.value = false
    }

    onBeforeUnmount(() => {
      // 清除定时器
      clearIntervals(operateType.init);
      clearIntervals(operateType.updating);
    });

    return {
      fileList,
      tipList,
      updateFormRef,
      updateForm,
      uploadHeaders,
      updateFormRule,
      upgradeInfo,
      isShowTakeEffectTip,
      loading,
      getUpgradeStatus,
      uploadSuccess,
      uploadError,
      clickToUpdate,
      clickTakeEffect,
      confirmTakeEffect,
      cancelTakeEffect,
      beforeUploadFile,
      upload,
      onUploadChange,
      isUpgrading,
      isTakingEffect,
    }
  },
})
</script>

<style scoped>
.container {
  padding: 24px;
  background: var(--el-bg-color-overlay);
  height: 80vh;
  border-radius: 4px;
}

.title {
  font-size: 16px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
  margin-bottom: 10px;
}
</style>