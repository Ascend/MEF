<template>
  <div class="title tab-title">{{ $t("journal.logCollection.tabName") }}</div>
  <div class="container">
    <div class='basic-info'>{{ $t('common.basicInfo') }}</div>
    <app-tip :tip-text="$t('journal.logCollection.collectTopTip')" />
    <div class="collect-block">
      <div class="sub-title">
        <span>
          {{ $t("journal.logCollection.tabName") }}
        </span>
        <el-button class="collect-btn" type="primary" :disabled="isStartCollect" @click="startCollect">
          {{ $t("journal.logCollection.collect") }}
        </el-button>
      </div>
      <div>{{ $t("journal.logCollection.collectSecondTip") }}</div>
      <el-divider />
      <el-progress
        v-if="isStartCollect"
        :percentage="logCollectState.percentage"
        :status="logCollectState.state"
        :stroke-width="8"
      />
    </div>
  </div>
</template>

<script>
import {defineComponent, ref, onUnmounted, watch} from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessageBox } from "element-plus";

import AppTip from '@/components/AppTip.vue';
import { queryLogsInfo, downloadLogInfo, queryLogCollectProgress } from '@/api/journal';
import { querySystemsSourceInfo } from '@/api/information';
import { dateFormat } from '@/utils/commonMethods';
import { showErrorAlert, handleOperationResponseError } from '@/utils/commonMethods';
import { saveFile } from '@/utils/downloadFile';

export default defineComponent({
  name: 'MaintenanceJournal',
  components: {
    AppTip,
  },
  setup() {
    const { t } = useI18n();
    const progressTimer = ref();
    const isStartCollect = ref(false);
    const logCollectState = ref({
      percentage: 0,
      state: '',
    });

    const getLogNames = async () => {
      let { data } = await queryLogsInfo();
      if (data['Members@odata.count'] === 0) {
        showErrorAlert(t('journal.logCollection.collectError'))
        return '';
      }

      let logNames = '';
      for (let i = 0; i < data['Members@odata.count']; i++) {
        let odata = data?.Members[i]['@odata.id'];
        if (odata === undefined) {
          continue
        }
        logNames += odata.split('/').pop() + ' ';
      }
      return logNames;
    }

    const getHostnameAndModel = async () => {
      let { data } = await querySystemsSourceInfo();
      return {
        hostName: data?.HostName,
        model: data?.Model,
      }
    }

    const generateCollectTarName = async () => {
      let { hostName, model } = await getHostnameAndModel();

      let now = dateFormat(new Date(), 'yyyyMMddhhmmss');
      return `${hostName}_${model}_${now}.tar.gz`;
    }

    const startCollect = async () => {
      await ElMessageBox.confirm(
        t('journal.logCollection.collectSecondTip'),
        t('journal.logCollection.collectConfirm'),
        {
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel'),
          type: 'warning',
        }
      );
      isStartCollect.value = true;
      let logFileName = await generateCollectTarName();

      let logNames = await getLogNames();
      if (!logNames) {
        return;
      }

      setProgress();

      let params = { name: logNames.trim() }
      try {
        let { data } = await downloadLogInfo(params);
        saveFile(data, logFileName);
      } catch (err) {
        handleOperationResponseError(err, t('journal.logCollection.collectError'))
      }

      isStartCollect.value = false;
    }

    const setProgress = () => {
      progressTimer.value = setInterval(async () => {
        let { data } = await queryLogCollectProgress();
        if (data === undefined) {
          return;
        }
        if (data?.TaskState === 'failed') {
          logCollectState.value.state = 'exception'
          showErrorAlert(t('journal.logCollection.collectError'))
          clearProgressInterval();
          return
        }
        logCollectState.value.percentage = data?.PercentComplete ?? 0;
      }, 5 * 1000)
    }

    const clearProgressInterval = () => {
      clearInterval(progressTimer.value);
      progressTimer.value = null;
    }

    watch(isStartCollect, () => {
      if (!isStartCollect.value) {
        clearProgressInterval();
      }
    })

    onUnmounted(() => {
      clearProgressInterval();
    })

    return {
      logCollectState,
      isStartCollect,
      startCollect,
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

.collect-block {
  width: calc(100% - 48px);
  height: 150px;
  background: var(--dialog-div-background);
  padding: 24px;
}

.sub-title {
  font-size: 20px;
  color: var(--el-text-color-regular);
  letter-spacing: 0;
  line-height: 30px;
  font-weight: 400;
  margin-bottom: 10px;
}

.collect-btn {
  float: right;
  right: 0;
  border: none;
}

.basic-info {
  font-size: 16px;
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
  margin-bottom: 8px;
}
</style>