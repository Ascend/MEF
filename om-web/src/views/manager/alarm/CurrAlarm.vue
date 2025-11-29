<template>
  <div class="tab-title">
    {{ $t("alarm.currentAlarm.alarmList") }}
  </div>
  <div class="container">
    <el-button type="primary" v-if="isStartRefresh" @click="clickStopRefresh">
      <img src="@/assets/img/maintenance/stop.svg" style="height: 14px; width: 14px;" alt="" />
      {{ $t("alarm.currentAlarm.stopRefresh") }}
    </el-button>
    <el-button type="primary" v-else @click="clickStartRefresh">
      <img src="@/assets/img/maintenance/play.svg" style="height: 14px; width: 14px;" alt="" />
      {{ $t("alarm.currentAlarm.startRefresh") }}
    </el-button>
    <div style="float: right; display: flex; align-items: center;">
      <el-input
        v-model="alarmQuery"
        @input="searchAlarmTable"
        clearable
        :placeholder="$t('alarm.common.searchPlaceholder')"
        style="width: 300px;"
      />
      <el-button style="width: 32px; height: 32px; margin-left: 20px;" @click="clickRefresh">
        <img style="width: 14px; height: 14px" alt="" src="@/assets/img/common/refresh.svg" />
      </el-button>
    </div>
    <el-table
      :data="alarmData"
      max-height="630"
      style="width: 100%; margin-top: 30px;"
      ref="alarmDataRef"
      @sort-change="sortSeverity"
    >
      <el-table-column :label="$t('alarm.common.alarmId')" prop="AlarmId">
        <template #default="props">
          <el-button link type="primary" size="small" @click="clickAlarmDetail(props.row)">
            {{ props.row?.AlarmId }}
          </el-button>
        </template>
      </el-table-column>
      <el-table-column
        :label="$t('alarm.common.severity')"
        sortable="custom"
        prop="PerceivedSeverity"
      >
        <template #default="props">
          <el-tag effect="dark" v-if="props.row.PerceivedSeverity === '0'" type="danger">{{ $t("alarm.common.emergency") }}</el-tag>
          <el-tag effect="dark" v-if="props.row.PerceivedSeverity === '1'" type="warning">{{ $t("alarm.common.serious") }}</el-tag>
          <el-tag effect="dark" v-if="props.row.PerceivedSeverity === '2'">{{ $t("alarm.common.normal") }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('alarm.common.alarmObject')" prop="AlarmInstance" />
      <el-table-column :label="$t('alarm.common.alarmName')" prop="AlarmName" />
      <el-table-column :label="$t('alarm.currentAlarm.generateAt')" prop="Timestamp" />
    </el-table>
    <el-pagination
      layout="total, sizes, prev, pager, next"
      :current-page="pagination.pageNum"
      :total="filterData.length"
      :page-size="pagination.pageSize"
      :page-sizes="[10, 20, 50, 100]"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
      style="position: absolute; bottom: 20px;"
    />
  </div>
  <el-dialog
    :title="$t('alarm.currentAlarm.stopReminder')"
    v-model="isShowDetailDialog"
  >
    <div style="display: flex; align-items: center; word-break: break-word;">
      <img src="@/assets/img/alarm.svg" alt="" style="margin-right: 10px;"/>
      {{ $t("alarm.currentAlarm.stopReminderText") }}
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="confirmStop">
          {{ $t("common.confirm") }}
        </el-button>
        <el-button @click="cancelStop">
          {{ $t("common.cancel") }}
        </el-button>
      </span>
    </template>
  </el-dialog>
  <alarm-detail :alarm-item="selectedAlarmItem" @closeDetailDrawer="closeDetailDrawer"/>
</template>

<script>
import { ref, defineComponent, onMounted, onUnmounted } from 'vue';
import { queryAlarmSourceService } from '@/api/alarm';

import AlarmDetail from '@/views/manager/alarm/AlarmDetail.vue';
import constants from '@/utils/constants';
import { getLanguageDefaultChinese } from '@/utils/commonMethods';
import { fetchJson } from '@/api/common';

export default defineComponent({
  name: 'CurrAlarm',
  components: {
    AlarmDetail,
  },
  setup() {
    const isStartRefresh = ref(true);
    const alarmQuery = ref();
    const alarmData = ref();
    const originAlarmData = ref([]);
    const filterData = ref([]);

    let solutionZh;
    let solutionEn;
    const loadJson = async () => {
      solutionZh = await fetchJson('/config/alarm_info_solution_zh.json');
      solutionEn = await fetchJson('/config/alarm_info_solution_en.json');
    }
    const selectedAlarmItem = ref({
      item: null,
      isShowDetailDrawer: false,
    })
    const isShowDetailDialog = ref(false)
    const pagination = ref({
      pageNum: 1,
      pageSize: 10,
    })

    let originSolutionMapper = Object();
    const generateSolutionMapper = () => {
      let solutionList = getLanguageDefaultChinese() === 'zh' ? solutionZh : solutionEn;
      solutionList.EventSuggestion.forEach(item => {
        originSolutionMapper[item.id] = item
      })
    }

    const initAlarmTable = async (isShowLoading = true, AutoRefresh = false) => {
      let { data } = await queryAlarmSourceService(isShowLoading, AutoRefresh);
      let alarm = [];
      for (let i = 0; i < data?.AlarMessages?.length; i++) {
        let item = data?.AlarMessages[i];
        item.AlarmName = originSolutionMapper[item.AlarmId]?.name;
        alarm.push(item)
      }
      originAlarmData.value = alarm;
      filterData.value = originAlarmData.value
      alarmData.value = originAlarmData.value;
      handleCurrentChange(pagination.value.pageNum);
    }

    onMounted(async () => {
      await loadJson();
      generateSolutionMapper();
      await initAlarmTable();
      await startRefreshTimer();
    })

    let autoRefreshTimer;
    const startRefreshTimer = () => {
      autoRefreshTimer = setInterval(async () => {
        await initAlarmTable(false, true)
        searchAlarmTable();
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

    const clickAlarmDetail = (row) => {
      if (isStartRefresh.value) {
        isShowDetailDialog.value = true
        selectedAlarmItem.value = {
          item: row,
        }
      } else {
        selectedAlarmItem.value = {
          item: row,
          isShowDetailDrawer: true,
        }
      }
    }

    const closeDetailDrawer = () => {
      selectedAlarmItem.value = {
        item: null,
        isShowDetailDrawer: false,
      }
    }

    const cancelStop = () => {
      isShowDetailDialog.value = false
    }

    const confirmStop = () => {
      stopRefreshTimer();
      selectedAlarmItem.value.isShowDetailDrawer = true
      isStartRefresh.value = false
      isShowDetailDialog.value = false
    }

    const clickStopRefresh = () => {
      stopRefreshTimer();
      isStartRefresh.value = false
    }

    const clickStartRefresh = () => {
      startRefreshTimer();
      isStartRefresh.value = true;
    }

    const refreshAlarmData = () => {
      alarmData.value = filterData.value.slice(
        pagination.value.pageSize * (pagination.value.pageNum - 1),
        pagination.value.pageSize * pagination.value.pageNum
      );
    }

    const alarmDataRef = ref();
    const searchAlarmTable = () => {
      alarmDataRef.value.clearSort();
      if (!alarmQuery.value) {
        filterData.value = [].concat(originAlarmData.value);
      } else {
        filterData.value = originAlarmData.value.filter(item => {
          let hasAlarmId = item?.AlarmId?.indexOf(alarmQuery.value) !== -1;
          let hasAlarmInstance = item?.AlarmInstance?.indexOf(alarmQuery.value) !== -1;
          let hasAlarmName = item?.AlarmName?.indexOf(alarmQuery.value) !== -1;
          return Boolean(hasAlarmId || hasAlarmInstance || hasAlarmName)
        })
      }
      refreshAlarmData();
    }

    const clickRefresh = async () => {
      await initAlarmTable();
      searchAlarmTable();
    }

    const handleSizeChange = (value) => {
      pagination.value.pageSize = value
      refreshAlarmData();
    }

    const handleCurrentChange = (value) => {
      pagination.value.pageNum = value
      refreshAlarmData();
    }

    const sortSeverity = ({ column, prop, order }) => {
      if (order === 'descending') {
        filterData.value.sort((a, b) => parseInt(a[prop]) - parseInt(b[prop]));
      } else if (order === 'ascending') {
        filterData.value.sort((a, b) => parseInt(b[prop]) - parseInt(a[prop]));
      } else {
        searchAlarmTable();
      }
      refreshAlarmData();
    }
    return {
      pagination,
      selectedAlarmItem,
      isStartRefresh,
      alarmQuery,
      alarmData,
      isShowDetailDialog,
      originAlarmData,
      filterData,
      alarmDataRef,
      initAlarmTable,
      clickRefresh,
      clickAlarmDetail,
      cancelStop,
      confirmStop,
      clickStopRefresh,
      clickStartRefresh,
      closeDetailDrawer,
      searchAlarmTable,
      handleSizeChange,
      handleCurrentChange,
      sortSeverity,
    }
  },
})
</script>

<style scoped>
.container {
  background: var(--el-bg-color-overlay);
  height: 76vh;
  padding: 20px;
  border-radius: 4px;
  overflow: auto;
}
</style>