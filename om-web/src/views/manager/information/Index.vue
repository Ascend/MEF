<template>
  <div v-loading="loading">
    <div class='block sys-info'>
      <img class='device-img' :src="$t('img.device')" alt=''/>
      <div v-if='!isEditHostname' style='display: flex; align-items: center;'>
        <el-popover placement="left" trigger="hover">
          <template #reference>
            <span class='hostname'>{{ formatText(systemInfo.hostName, 10) }}</span>
          </template>
          <div style="word-break: break-word;">{{ systemInfo.hostName }}</div>
        </el-popover>
        <el-tag effect="plain" type='info' style='margin-left: 10px; width: 80px;'>{{ $t('information.systemInfo.hostName') }}</el-tag>
        <el-button style='margin-left: 10px;' size='small' @click='clickEditSysInfo'>
          {{ $t('common.edit') }}
        </el-button>
      </div>
      <div v-else style='display: flex; align-items: center;'>
        <el-form
          :model='hostNameForm'
          label-position='right'
          :rules='hostNameFormRule'
          ref='hostNameFormRef'
          @submit.native.prevent
        >
          <el-form-item prop='hostName' style='margin-top: 0; margin-bottom: 0;'>
            <el-input
              style='width: 200px;'
              v-model='hostNameForm.hostName'
              @keyup.enter.native="saveEditHostname(hostNameFormRef)"
            />
          </el-form-item>
        </el-form>
        <span style='margin-left: 10px;'>
          <el-button size='small' type='primary' @click='saveEditHostname(hostNameFormRef)'>{{ $t('common.save') }}</el-button>
          <el-button size='small' @click='cancelEditHostname'>{{ $t('common.cancel') }}</el-button>
        </span>
      </div>
      <div class='info'>
        <div>
          <div class='title'>
            {{ $t('information.systemInfo.os') }}
          </div>
          <div class='content'>
            {{ systemInfo.os }}
          </div>
        </div>
        <div>
          <div class='title'>
            {{ $t('information.systemInfo.firmwareVersion') }}
          </div>
          <div class='content'>
            {{ systemInfo.firmwareVersion || constants.DEFAULT_EMPTY_TEXT }}
          </div>
        </div>
        <div>
          <div class='title'>
            {{ $t('information.systemInfo.uptime') }}
          </div>
          <div class='content'>
            {{ systemInfo.uptime }}
          </div>
        </div>
      </div>
      <div class='info'>
        <div>
          <div class='title'>
            {{ $t('information.systemInfo.productAssetLabel') }}
          </div>
          <div class='content' v-if='!isEditAssetLabel' style='display: flex; align-items: center;'>
            <el-popover placement="left" trigger="hover">
              <template #reference>
                <span>{{ formatText(systemInfo.productAssetLabel, 10) }}</span>
              </template>
              <div style="word-break: break-word;">{{ systemInfo.productAssetLabel }}</div>
            </el-popover>
            <el-button style='margin-left: 10px;' size='small' @click='clickEditAssetLabel'>
              {{ $t('common.edit') }}
            </el-button>
          </div>
          <div class='content' v-else style='display: flex; align-items: center;'>
            <el-form
              :model='assetLabelForm'
              label-position='right'
              :rules='assetLabelFormRule'
              ref='assetLabelFormRef'
              @submit.native.prevent
            >
              <el-form-item prop='assetLabel' style='margin-top: 0; margin-bottom: 0;'>
                <el-input
                  style='width: 150px;'
                  v-model='assetLabelForm.assetLabel'
                  @keyup.enter.native='saveEditAssetLabel(assetLabelFormRef)'
                />
              </el-form-item>
            </el-form>
            <span style='margin-left: 10px;'>
              <el-button size='small' type='primary' @click='saveEditAssetLabel(assetLabelFormRef)'>{{ $t('common.save') }}</el-button>
              <el-button size='small' @click='cancelEditAssetLabel'>{{ $t('common.cancel') }}</el-button>
            </span>
          </div>
        </div>
        <div>
          <div class='title'>
            {{ $t('information.systemInfo.npuVersion') }}
          </div>
          <div class='content'>
            {{ systemInfo.npuVersion || constants.DEFAULT_EMPTY_TEXT}}
          </div>
        </div>
        <div>
          <div class='title'>
            {{ $t('information.systemInfo.model') }}
          </div>
          <div class='content'>
            {{ systemInfo.model }}
          </div>
        </div>
      </div>
      <div class='info'>
        <div>
          <div class='title'>
            {{ $t('information.systemInfo.productSn') }}
          </div>
          <div class='content'>
            {{ systemInfo.productSn }}
          </div>
        </div>
        <div v-if="isA500()">
          <div class='title'>
            {{ $t('information.systemInfo.mcuVersion') }}
          </div>
          <div class='content'>
            {{ systemInfo.mcuVersion || constants.DEFAULT_EMPTY_TEXT }}
          </div>
        </div>
        <div style='border: none;'>
        </div>
      </div>
    </div>

    <div class='block device-info'>
      <div class='tab-title'>{{ $t('information.systemInfo.peripherals') }}</div>
      <div style='float: right; display: flex; align-items: center;'>
        <el-input
          v-model='snQuery'
          @input='handleQuerySn'
          clearable
          :placeholder='$t("information.systemInfo.searchSnPlaceholder")'
        />
        <el-button style='width: 32px; height: 32px; margin-left: 20px;' @click='refreshDeviceTable'>
          <img style='width: 14px; height: 14px' alt='' src='@/assets/img/common/refresh.svg' />
        </el-button>
      </div>
      <el-table :data='deviceData' style='width: 100%; margin-top: 50px; height: 400px;' >
        <el-table-column :label='$t("information.systemInfo.productSn")' prop='productSn' />
        <el-table-column width='100' :label='$t("information.systemInfo.category")' prop='category' />
        <el-table-column :label='$t("information.systemInfo.vendor")' prop='vendor' />
        <el-table-column :label='$t("information.systemInfo.model")' prop='model' />
        <el-table-column
          :label='$t("information.systemInfo.healthStatus")'
          width='140'
          prop='healthStatus'
        >
          <template #default="props">
            <el-tag effect="dark" v-if='isHealthStatusOk(props.row.healthStatus)' type='success'>{{ props.row.healthStatus }}</el-tag>
            <el-tag effect="dark" v-else type='danger'>{{ props.row.healthStatus }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label='$t("information.systemInfo.startupStatus")'
          width='140'
          prop='startupStatus'
        >
          <template #default="props">
            <el-tag effect="dark" v-if='isStartupStatusEnabled(props.row.startupStatus)' type='success'>{{ props.row.startupStatus }}</el-tag>
            <el-tag effect="dark" v-else type='danger'>{{ props.row.startupStatus }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label='$t("information.systemInfo.deviceLocation")' prop='deviceLocation' />
        <el-table-column :label='$t("information.systemInfo.firmwareVersion")' prop='firmwareVersion' />
      </el-table>
      <el-pagination
        layout="total, sizes, prev, pager, next"
        :current-page="pagination.pageNum"
        :total="originDeviceData.length"
        :page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        style="margin-top: 20px;"
      />
    </div>
  </div>
</template>

<script>
import { ref, defineComponent, onMounted, reactive } from 'vue';
import { useI18n } from 'vue-i18n'

import { queryAllExtendedDevicesInfo, querySystemsSourceInfo, modifySystemsSourceInfo, queryAiProcessorInfo } from '@/api/information';
import { getUrls } from '@/api/http';
import { validateHostName, validateAssetTag } from '@/utils/validator';
import constants from '@/utils/constants';
import {
  AppMessage, checkModuleType, handleOperationResponseError, isA500, isFulfilled, checkUrlsResponse
} from '@/utils/commonMethods';

export default defineComponent({
  name: 'Maintenance-Information',
  setup() {
    const { t } = useI18n()
    const isEditHostname = ref(false);
    const isEditAssetLabel = ref(false);
    const originDeviceData = ref([]);
    const deviceData = ref();
    const snQuery = ref();
    const loading = ref(false);
    const pagination = ref({
      pageNum: 1,
      pageSize: 10,
    })

    const systemInfo = ref({
      hostName: '',
      os: '',
      uptime: '',
      productAssetLabel: '',
      productSn: '',
      model: '',
      mcuVersion: '',
      npuVersion: '',
      firmwareVersion: '',
    })

    const hostNameForm = ref({
      hostName: '',
    })
    const hostNameFormRef = ref();
    const hostNameFormRule = reactive({
      hostName: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateHostName,
          trigger: 'blur',
        },
      ],
    })

    const assetLabelForm = ref({
      assetLabel: '',
    })
    const assetLabelFormRef = ref();
    const assetLabelFormRule = reactive({
      assetLabel: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateAssetTag,
          trigger: 'blur',
        },
      ],
    })

    const getSystemInfo = async () => {
      try {
        let { data } = await querySystemsSourceInfo(false);
        return data;
      } catch (e) {
        loading.value = false
        return null;
      }
    }

    const setSystemInfo = (data) => {
      systemInfo.value.hostName = data?.HostName;
      systemInfo.value.os = data?.Oem?.OSVersion ?? constants.DEFAULT_EMPTY_TEXT;
      systemInfo.value.uptime = data?.Oem?.Uptime ?? constants.DEFAULT_EMPTY_TEXT;
      systemInfo.value.productAssetLabel = data?.AssetTag;
      systemInfo.value.productSn = data?.SerialNumber ?? constants.DEFAULT_EMPTY_TEXT;
      systemInfo.value.model = data?.Model ?? constants.DEFAULT_EMPTY_TEXT;
      let versions = data?.Oem?.Firmware;
      for (let i = 0; i < versions?.length; i++) {
        if (checkModuleType(versions[i].Module, 'mcu')) {
          systemInfo.value.mcuVersion = versions[i].Version || constants.DEFAULT_EMPTY_TEXT;
        } else if (checkModuleType(versions[i].Module, 'npu')) {
          systemInfo.value.npuVersion = versions[i].Version || constants.DEFAULT_EMPTY_TEXT;
        } else if (checkModuleType(versions[i].Module, 'om') || checkModuleType(versions[i].Module, 'firmware')) {
          systemInfo.value.firmwareVersion = versions[i].Version || constants.DEFAULT_EMPTY_TEXT;
        }
      }
    }

    const fetchAiProcessorInfo = async () => {
      try {
        let { data } = await queryAiProcessorInfo(false);
        return data;
      } catch (e) {
        loading.value = false
        return null;
      }
    }

    const setNpuInfo = (data) => {
      systemInfo.value.npuVersion = data?.NpuVersion || constants.DEFAULT_EMPTY_TEXT;
    }

    const init = async () => {
      let data = await getSystemInfo();
      setSystemInfo(data);

      let aiProcessorInfo = await fetchAiProcessorInfo();
      setNpuInfo(aiProcessorInfo);
    }

    const initDeviceTable = async () => {
      let { data } = await queryAllExtendedDevicesInfo(false);
      let deviceList = [];

      let params = data?.Members.map(item => (
        {
          url: item['@odata.id'],
          isShowLoading: false,
        }
      ))

      let allResponse = await getUrls(params);
      allResponse.filter(item => isFulfilled(item)).map(item => {
        let { data: singleDevice } = item.value;
        if (!singleDevice) {
          return;
        }
        let device = {
          productSn: singleDevice?.SerialNumber,
          category: singleDevice?.DeviceClass,
          vendor: singleDevice?.Manufacturer,
          model: singleDevice?.Model,
          healthStatus: singleDevice?.Status?.Health,
          startupStatus: singleDevice?.Status?.State,
          deviceLocation: singleDevice?.Location,
          firmwareVersion: singleDevice?.FirmwareVersion,
        }
        if (!Object.values(device).every(value => value === null || value === undefined)) {
          deviceList.push(device);
        }
      })
      checkUrlsResponse(allResponse);
      originDeviceData.value = deviceList;
      handleCurrentChange(pagination.value.pageNum);
    }

    const handleQuerySn = () => {
      deviceData.value = originDeviceData.value.filter(item => item.productSn.indexOf(snQuery.value) !== -1)
    }

    const refreshDeviceTable = async () => {
      loading.value = true
      await initDeviceTable();
      loading.value = false
    }

    onMounted(async () => {
      loading.value = true
      await init();
      await initDeviceTable();
      loading.value = false
    })

    const clickEditSysInfo = () => {
      hostNameForm.value.hostName = systemInfo.value.hostName;
      isEditHostname.value = true;
    }

    const clickEditAssetLabel = () => {
      assetLabelForm.value.assetLabel = systemInfo.value.productAssetLabel;
      isEditAssetLabel.value = true;
    }

    const saveEditHostname = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        isEditHostname.value = false;
        let params = {
          HostName: hostNameForm.value.hostName,
        }
        try {
          await modifySystemsSourceInfo(params, true);
          AppMessage.success(t('common.editSuccess'))
          await init();
        } catch (err) {
          handleOperationResponseError(err)
        }
      })
    }

    const cancelEditHostname = () => {
      isEditHostname.value = false;
    }

    const saveEditAssetLabel = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        isEditAssetLabel.value = false;
        let params = {
          AssetTag: assetLabelForm.value.assetLabel,
        }
        try {
          await modifySystemsSourceInfo(params, true);
          AppMessage.success(t('common.editSuccess'))
          await init();
        } catch (err) {
          handleOperationResponseError(err)
        }
      })
    }

    const cancelEditAssetLabel = () => {
      isEditAssetLabel.value = false;
    }

    const isHealthStatusOk = (value) => value?.toLowerCase() === 'ok'

    const isStartupStatusEnabled = (value) => value?.toLowerCase() === 'enabled'

    const handleSizeChange = (value) => {
      pagination.value.pageSize = value
      deviceData.value = originDeviceData.value.slice(
        pagination.value.pageSize * (pagination.value.pageNum - 1),
        pagination.value.pageSize * pagination.value.pageNum
      );
    }

    const handleCurrentChange = (value) => {
      pagination.value.pageNum = value
      deviceData.value = originDeviceData.value.slice(
        pagination.value.pageSize * (pagination.value.pageNum - 1),
        pagination.value.pageSize * pagination.value.pageNum
      );
    }
    const formatText = (text, length) => text?.length > length ? text.slice(0, length) + '...' : text

    return {
      constants,
      snQuery,
      isEditHostname,
      isEditAssetLabel,
      systemInfo,
      hostNameForm,
      hostNameFormRef,
      hostNameFormRule,
      assetLabelForm,
      assetLabelFormRule,
      assetLabelFormRef,
      deviceData,
      pagination,
      originDeviceData,
      loading,
      handleQuerySn,
      clickEditSysInfo,
      clickEditAssetLabel,
      saveEditHostname,
      cancelEditHostname,
      saveEditAssetLabel,
      cancelEditAssetLabel,
      refreshDeviceTable,
      isHealthStatusOk,
      isStartupStatusEnabled,
      handleSizeChange,
      handleCurrentChange,
      formatText,
      isA500,
    }
  },
});
</script>

<style lang='scss' scoped>
.block {
  padding: 24px;
  background: var(--el-bg-color-overlay);
  border-radius: 4px;
}

.sys-info {
  height: 260px;
}

.device-info {
  margin-top: 20px;
}

.device-img {
  width: 200px;
  height: 140px;
  margin: 70px 20px;
  float: left;
}

.hostname {
  font-size: 24px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 36px;
  font-weight: 700;
}

.info {
  display: flex;
  margin-top: 30px;
}

.info > div {
  flex: 1;
  text-align: left;
  padding-left: 10px;
}

.info > div {
  border-left: 1px solid var(--el-border-color);
}

.title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 200;
  margin-bottom: 10px;
  font-family: 'HarmonyOS_Sans_SC_Light', serif;
}

.content {
  font-size: 14px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
  font-family: 'HarmonyOS_Sans_SC_Regular', serif;
}
</style>