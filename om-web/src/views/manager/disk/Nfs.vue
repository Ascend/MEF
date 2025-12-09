<template>
  <div v-loading="loading">
    <div class='tab-title'>{{ $t('disk.nfsManagement.sharedDirList') }}</div>
    <div class='container'>
      <el-button type='primary' class="medium-font" @click='clickCreate'>
        <img src='@/assets/img/manager/add.svg' alt='' style='height: 14px; width: 14px; margin-right: 8px;'/>
        {{ $t('disk.nfsManagement.createSharedDir') }}
      </el-button>
      <div style='float: right; display: flex; align-items: center;'>
        <el-input
          v-model='nfsQuery'
          @input='searchNfsTable'
          clearable
          :placeholder='$t("disk.nfsManagement.searchPlaceholder")'
          style='width: 300px;'
        />
        <el-button style='width: 32px; height: 32px; margin-left: 10px;' @click='initTableDataByNfsInfo(false)'>
          <img style='width: 14px; height: 14px' alt='' src='@/assets/img/common/refresh.svg' />
        </el-button>
      </div>
      <el-table :data='diskData' style='width: 100%; margin-top: 20px;' :empty-text='$t("common.empty")'>
        <el-table-column width='200' prop='ServerIP' :label='$t("disk.nfsManagement.serverIp")' />
        <el-table-column width='400' prop='ServerDir' :label='$t("disk.nfsManagement.sharedDir")' />
        <el-table-column width='240' prop='MountPath' :label='$t("disk.common.mountPath")' />
        <el-table-column width='200' prop='CapacityBytes' :label='$t("disk.common.capacityBytes")' >
          <template #default='props'>
            {{ convertToGB(props.row.CapacityBytes) }}
          </template>
        </el-table-column>
        <el-table-column width='200' prop='FreeBytes' :label='$t("disk.common.freeBytes")'>
          <template #default='props'>
            {{ convertToGB(props.row.FreeBytes) }}
          </template>
        </el-table-column>
        <el-table-column width='140' prop='Status' :label='$t("disk.common.healthStatus")'
          :filters='healthStatusFilter'
          :filter-method='filterHealthStatus'
        >
          <template #default='props'>
            <el-tag effect="dark" v-if='isHealthStatusOk(props.row.Status)' type='success'>{{ $t('common.statusOk') }}</el-tag>
            <el-tag effect="dark" v-else type='danger'>{{ $t('common.statusNotOk') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop='operations' :label='$t("common.operations")'>
          <template #default='props'>
            <el-button link type='primary' size='small' @click='showUnmountDialog(props.row)'>
              {{ $t('disk.common.unmount') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
  <el-dialog
    v-model='isShowCreate'
    width='600px'
    :before-close='handleCloseCreateDialog'
    :title='$t("disk.nfsManagement.createSharedDir")'
    :close-on-click-modal="false"
  >
    <app-tip :tip-text='$t("disk.nfsManagement.createTip")' />
    <el-form
      :model='nfsForm'
      label-position='right'
      label-width='auto'
      :rules='nfsFormRule'
      ref='nfsFormRef'
      style='width: 500px;'
      @submit.native.prevent
    >
      <el-form-item :label='$t("disk.nfsManagement.serverIp")' prop='ServerIP' style="margin-bottom: 40px;">
        <el-input v-model='nfsForm.ServerIP' />
      </el-form-item>
      <el-form-item :label='$t("disk.nfsManagement.sharedDir")' prop='ServerDir' style="margin-bottom: 40px;">
        <el-input v-model='nfsForm.ServerDir' />
      </el-form-item>
      <el-form-item :label='$t("disk.nfsManagement.localMountPath")' prop='MountPath' style="margin-bottom: 40px;">
        <el-input v-model='nfsForm.MountPath' @keyup.enter.native='confirmCreate(nfsFormRef)'/>
      </el-form-item>
      <el-form-item :label='$t("disk.nfsManagement.protocolVersion")' prop='FileSystem' >
        <el-select v-model='nfsForm.FileSystem' :placeholder='$t("common.pleaseSelect")' disabled>
          <el-option v-for='item in protocolVersionOptions' :key='item.value' :label='item.label' :value='item.value' />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class='dialog-footer'>
        <el-button type='primary' @click='confirmCreate(nfsFormRef)'>
          {{ $t('common.confirm') }}
        </el-button>
        <el-button @click='cancelCreate(nfsFormRef)'>
          {{ $t('common.cancel') }}
        </el-button>
      </span>
    </template>
  </el-dialog>
  <el-dialog
    v-model='isShowUnmount'
    width='420px'
    :title='$t("disk.common.unmountReminder")'
  >
    <div style='display: flex; align-items: center; word-break: break-word;'>
      <img src='@/assets/img/alarm.svg' alt='' style='margin-right: 10px;'/>
      {{ $t('disk.nfsManagement.unmountTip') }}
    </div>
    <template #footer>
      <span class='dialog-footer'>
        <el-button type='primary' @click='confirmUnmount'>
          {{ $t('common.confirm') }}
        </el-button>
        <el-button @click='cancelUnmount'>
          {{ $t('common.cancel') }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import { defineComponent, ref, reactive, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';

import { validateIp, validateServerPath } from '@/utils/validator';
import { queryNfsInfo, mountNfs, unmountNfs } from '@/api/drive';
import AppTip from '@/components/AppTip.vue';
import { convertToGB, AppMessage, handleOperationResponseError, showErrorAlert } from '@/utils/commonMethods';
import constants from '@/utils/constants';

export default defineComponent({
  name: 'Nfs',
  components: {
    AppTip,
  },
  setup() {
    const diskData = ref([]);
    const originData = ref([]);
    const isShowCreate = ref(false);
    const isShowUnmount = ref(false);
    const loading = ref(false);
    const selectedUnmountRecord = ref();
    const { t } = useI18n();
    const nfsForm = ref({
      ServerIP: '',
      ServerDir: '',
      MountPath: '',
      FileSystem: 'nfs4',
    })

    const nfsFormRef = ref();
    const nfsQuery = ref();
    const nfsFormRule = reactive({
      ServerIP: [
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
      ServerDir: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateServerPath,
          trigger: 'blur',
        },
      ],
      MountPath: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateServerPath,
          trigger: 'blur',
        },
      ],
    })

    const protocolVersionOptions = [
      {
        value: 'nfs4',
        label: 'NFS4',
      },
    ]

    const clickCreate = () => {
      isShowCreate.value = true;
    }

    const healthStatusFilter = [
      {
        text: 'OK',
        value: 'OK',
      },
      {
        text: 'Critical',
        value: 'Critical',
      },
    ];
    const filterHealthStatus = (value, row) => row.healthStatus === value

    const confirmCreate = async (formElement) =>{
      if (!formElement) {
        return
      };
      await formElement.validate(async (valid) => {
        if (!valid) {
          return
        };
        isShowCreate.value = false;
        try {
          let params = nfsForm.value;
          loading.value = true;
          await mountNfs(params, false);
          await initTableDataByNfsInfo();
          loading.value = false;
          AppMessage.success(t('common.saveSuccessfully'))
        } catch (err) {
          loading.value = false;
          handleOperationResponseError(err);
        } finally {
          cancelCreate(nfsFormRef.value);
        }
      })
    }

    const handleCloseCreateDialog = () => {
      cancelCreate(nfsFormRef.value);
    }

    const cancelCreate = (formElement) => {
      formElement.resetFields();
      isShowCreate.value = false;
    }

    const initTableDataByNfsInfo = async (AutoRefresh = false) => {
      try {
        let { data } = await queryNfsInfo(false, AutoRefresh);
        originData.value = data.nfsList;
        diskData.value = originData.value;
      } catch (e) {
        showErrorAlert(t('common.requestError'))
      }
    }

    const isHealthStatusOk = (value) => value.toLowerCase() === 'ok'

    const searchNfsTable = () => {
      if (nfsQuery.value.length === 0) {
        diskData.value = originData.value;
      } else {
        diskData.value = originData.value.filter(item => Boolean(item.ServerIP.indexOf(nfsQuery.value) !== -1))
      }
    }

    const showUnmountDialog = (row) => {
      isShowUnmount.value = true;
      selectedUnmountRecord.value = row;
    }

    const confirmUnmount = async () => {
      isShowUnmount.value = false;
      try {
        loading.value = true;
        await unmountNfs(selectedUnmountRecord.value, false);
        await initTableDataByNfsInfo();
        loading.value = false;
        AppMessage.success(t('disk.nfsManagement.unmountSuccess'))
      } catch (err) {
        loading.value = false;
        handleOperationResponseError(err);
      } finally {
        cancelUnmount();
      }
    }

    const cancelUnmount = () => {
      isShowUnmount.value = false;
      selectedUnmountRecord.value = null;
    }

    onMounted(async () => {
      loading.value = true;
      await initTableDataByNfsInfo();
      loading.value = false;
      startRefreshTimer();
    })

    let autoRefreshTimer;

    const startRefreshTimer = () => {
      autoRefreshTimer = setInterval(async () => {
        await initTableDataByNfsInfo(true)
      }
      , constants.DEFAULT_TIMEOUT)
    }

    const stopRefreshTimer = () => {
      clearInterval(autoRefreshTimer);
      autoRefreshTimer = null;
    }

    onUnmounted(() => {
      stopRefreshTimer();
    })

    return {
      diskData,
      isShowCreate,
      nfsForm,
      nfsFormRef,
      nfsFormRule,
      protocolVersionOptions,
      healthStatusFilter,
      nfsQuery,
      isShowUnmount,
      loading,
      clickCreate,
      confirmCreate,
      cancelCreate,
      searchNfsTable,
      initTableDataByNfsInfo,
      filterHealthStatus,
      isHealthStatusOk,
      convertToGB,
      handleCloseCreateDialog,
      showUnmountDialog,
      confirmUnmount,
      cancelUnmount,
    }
  },
})
</script>

<style scoped>
.container {
  background: var(--el-bg-color-overlay);
  height: 76vh;
  padding: 20px;
  overflow: auto;
  border-radius: 4px;
}
</style>