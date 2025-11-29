<template>
  <div class='title tab-title'>{{ $t('reload.powerControl.tabName') }}</div>
  <div class='container'>
    <app-tip :tip-type="'info'" :tip-text="$t('reload.powerControl.reloadTip')" />
    <div class='btn-div'>
      <el-button type='primary' style='margin-right: 10px;' @click="confirmRestartType('graceful')">
        <img src='@/assets/img/common/refresh.svg' alt='' style='height: 14px; width: 16px;'/>
        {{ $t('reload.powerControl.restartSystem') }}
      </el-button>
      <app-popover :text="$t('reload.powerControl.restartTip')" />
    </div>
    <div class='btn-div' v-if="isA500()">
      <el-button type='primary' style='margin-right: 10px;' @click="confirmRestartType('restart')">
        <img src='@/assets/img/maintenance/upload.svg' alt='' style='height: 14px; width: 16px;'/>
        {{ $t('reload.powerControl.forceRestart') }}
      </el-button>
      <app-popover :text="$t('reload.powerControl.forceRestartTip')" />
    </div>
  </div>

  <el-dialog
    :title='confirmDialogTitle'
    v-model='isConfirmRestart'
    style='width: 400px;'
  >
    <div style='display: flex; align-items: center; word-break: break-word;'>
      <img src='@/assets/img/alarm.svg' alt=''/>
      {{ confirmDialogContent }}
    </div>
    <template #footer>
      <span class='dialog-footer'>
        <el-button type='primary' @click='confirmRestart'>
          {{ $t('common.confirm') }}
        </el-button>
        <el-button @click='cancelRestart'>
          {{ $t('common.cancel') }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import { defineComponent, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import AppPopover from '@/components/AppPopover.vue';
import AppTip from '@/components/AppTip.vue';
import { resetSystem } from '@/api/reload';
import { useRouter } from 'vue-router';
import { clearSessionAndJumpToLogin, handleOperationResponseError, isA500 } from '@/utils/commonMethods'

export default defineComponent({
  name: 'MaintenanceReload',
  components: {
    AppTip,
    AppPopover,
  },
  setup() {
    const { t } = useI18n()
    const isConfirmRestart = ref(false);
    const isConfirmForceRestart = ref(false);
    const confirmDialogTitle = ref();
    const confirmDialogContent = ref();
    const restartType = ref();
    const router = useRouter();

    const confirmRestart = async () => {
      isConfirmRestart.value = false;
      let params = {
        ResetType: restartType.value === 'graceful' ? 'GracefulRestart' : 'Restart',
      }
      try {
        await resetSystem(params);
        clearSessionAndJumpToLogin(router);
      } catch (err) {
        if (err?.response?.status !== 400) {
          clearSessionAndJumpToLogin(router);
        } else {
          handleOperationResponseError(err);
        }
      }
    }

    const cancelRestart = () => {
      isConfirmRestart.value = false;
    }

    const confirmRestartType = (type) => {
      restartType.value = type;
      if (type === 'graceful') {
        confirmDialogTitle.value = t('reload.powerControl.restartReminder')
        confirmDialogContent.value = t('reload.powerControl.restartConfirmTip')
      } else {
        confirmDialogTitle.value = t('reload.powerControl.forceRestartReminder')
        confirmDialogContent.value = t('reload.powerControl.forceRestartConfirmTip')
      }
      isConfirmRestart.value = true;
    }

    return {
      isConfirmRestart,
      isConfirmForceRestart,
      confirmDialogTitle,
      confirmDialogContent,
      confirmRestart,
      cancelRestart,
      confirmRestartType,
      isA500,
    }
  },
});
</script>

<style scoped>
.title {
  margin-top: 10px;
}

.container {
  margin-top: 20px;
  padding: 24px;
  background: var(--el-bg-color-overlay);
  height: 80vh;
  border-radius: 4px;
}

.btn-div {
  display: flex;
  align-items: center;
  margin-top: 20px;
}
</style>