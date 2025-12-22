<template>
  <el-drawer
    :modelValue="props.alarmItem.isShowDetailDrawer"
    :title="props.alarmItem.item?.AlarmId"
    @close="closeDrawer"
  >
    <div class="block basic-info">
      <div class="detail-block-title">{{ $t("common.basicInfo") }}</div>
      <el-row class="basic-info-row">
        <el-col :span="6">{{ $t("alarm.common.alarmId") }}</el-col>
        <el-col :span="18">{{ props.alarmItem.item?.AlarmId }}</el-col>
      </el-row>
      <el-row class="basic-info-row">
        <el-col :span="6">{{ $t("alarm.common.severity") }}</el-col>
        <el-col :span="18">
          <el-tag effect="dark" v-if="props.alarmItem.item?.PerceivedSeverity === '0'" type="danger">{{ $t("alarm.common.emergency") }}</el-tag>
          <el-tag effect="dark" v-if="props.alarmItem.item?.PerceivedSeverity === '1'" type="warning">{{ $t("alarm.common.serious") }}</el-tag>
          <el-tag effect="dark" v-if="props.alarmItem.item?.PerceivedSeverity === '2'">{{ $t("alarm.common.normal") }}</el-tag>
        </el-col>
      </el-row>
      <el-row class="basic-info-row">
        <el-col :span="6">{{ $t("alarm.common.alarmObject") }}</el-col>
        <el-col :span="18">{{ props.alarmItem.item?.AlarmInstance }}</el-col>
      </el-row>
      <el-row class="basic-info-row">
        <el-col :span="6">{{ $t("alarm.common.alarmName") }}</el-col>
        <el-col :span="18">{{ props.alarmItem.item?.AlarmName }}</el-col>
      </el-row>
      <el-row class="basic-info-row">
        <el-col :span="6">{{ $t("alarm.currentAlarm.generateAt") }}</el-col>
        <el-col :span="18">{{ props.alarmItem.item?.Timestamp }}</el-col>
      </el-row>
    </div>
    <div class="block suggestion">
      <div class="detail-block-title">{{ $t("alarm.common.suggestion") }}</div>
      <div class="suggestion-row" v-for="line in currAlarmSolution" :key="line" >{{ line }}</div>
    </div>
  </el-drawer>
</template>

<script>
import { defineComponent, ref, watch, onMounted } from 'vue';

import { getLanguageDefaultChinese } from '@/utils/commonMethods';
import { fetchJson } from '@/api/common';

export default defineComponent({
  name: 'AlarmDetail',
  props: {
    alarmItem: Object,
  },
  emits: ['closeDetailDrawer'],
  setup(props, ctx) {
    const currAlarmSolution = ref();
    let solutionZh;
    let solutionEn;

    const loadJson = async () => {
      solutionZh = await fetchJson('/config/alarm_info_solution_zh.json');
      solutionEn = await fetchJson('/config/alarm_info_solution_en.json');
    }

    onMounted(async () => {
      await loadJson();
    })

    const initDetailContent = () => {
      let solutionJson = getLanguageDefaultChinese() === 'zh' ? solutionZh : solutionEn
      let detailItem = solutionJson?.EventSuggestion.filter(item => item.id === props.alarmItem.item?.AlarmId)[0]
      if (!detailItem) {
        return;
      }
      let suggestion = detailItem.dealSuggestion.split('@#AB')
      currAlarmSolution.value = suggestion ?? [];
    }

    watch(
      () => props.alarmItem.isShowDetailDrawer,
      (isShowDetailDrawer) => {
        if (isShowDetailDrawer) {
          initDetailContent();
        } else {
          currAlarmSolution.value = [];
        }
      }
    )

    const closeDrawer = () => {
      ctx.emit('closeDetailDrawer')
    }

    return {
      props,
      currAlarmSolution,
      closeDrawer,
    }
  },
})
</script>

<style scoped>
.block {
  background: var(--el-bg-color-overlay);
  border-radius: 4px;
  padding: 20px;
}

.detail-block-title {
  font-size: 16px;
  line-height: 24px;
  font-weight: 700;
  margin-bottom: 20px;
}

.basic-info {
  margin-bottom: 20px;
  min-height: 280px;
}

.suggestion{
  min-height: 160px;
}

.basic-info-row {
  margin-top: 20px;
  color: var(--el-text-color-secondary);
}

.suggestion-row {
  color: var(--el-text-color-secondary);
  margin-bottom: 10px;
}
</style>