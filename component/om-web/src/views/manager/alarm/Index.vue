<template>
  <el-tabs v-model="activeName">
    <el-tab-pane :label="$t('alarm.currentAlarm.tabName')" name="current">
      <curr-alarm />
    </el-tab-pane>
    <el-tab-pane :label="$t('alarm.maskedAlarm.tabName')" name="masked">
      <alarm-table-and-dialog :table-data="maskedData" :tab-name="'masked'" @refresh="refresh" />
    </el-tab-pane>
    <el-tab-pane :label="$t('alarm.unmaskedAlarm.tabName')" name="unmasked">
      <alarm-table-and-dialog :table-data="unmaskedData" :tab-name="'unmasked'" @refresh="refresh" />
    </el-tab-pane>
  </el-tabs>
</template>

<script>
import { ref, defineComponent, onMounted } from 'vue';

import CurrAlarm from '@/views/manager/alarm/CurrAlarm.vue';
import AlarmTableAndDialog from '@/views/manager/alarm/AlarmTableAndDialog.vue';
import { queryAlarmShieldRules } from '@/api/alarm';
import { getLanguageDefaultChinese } from '@/utils/commonMethods';
import { fetchJson } from '@/api/common';

export default defineComponent({
  name: 'MaintenanceAlarm',
  components: {
    CurrAlarm,
    AlarmTableAndDialog,
  },
  setup() {
    const activeName = ref('current')
    const maskedData = ref([]);
    const unmaskedData = ref([]);
    const originSolutionMapper = ref();
    const originAlarmMapper = ref();
    let solutionZh;
    let solutionEn;
    let allAlarm;

    const loadJson = async () => {
      solutionZh = await fetchJson('/config/alarm_info_solution_zh.json');
      solutionEn = await fetchJson('/config/alarm_info_solution_en.json');
      allAlarm = await fetchJson('/config/all_alarm_for_manager.json');
    }

    const generateSolutionMapper = () => {
      let solutionList = getLanguageDefaultChinese() === 'zh' ? solutionZh : solutionEn;
      let mapper = Object();
      solutionList?.EventSuggestion?.forEach(item => {
        mapper[item.id] = item
      })
      originSolutionMapper.value = mapper;
    }

    const generateAlarmMapper = () => {
      let mapper = Object();
      allAlarm?.EventSuggestion.forEach(item => {
        mapper[item.innerid] = {
          UniquelyIdentifies: item.innerid,
          AlarmId: item.id,
          PerceivedSeverity: item.level,
          AlarmInstance: item.AlarmInstance,
          AlarmName: originSolutionMapper.value[item.id].name,
        }
      })
      originAlarmMapper.value = mapper;
    }

    const computeMaskedAndUnmaskedDataByShieldRules = (shieldRules) => {
      let alarmMapper = Object.assign({}, originAlarmMapper.value);
      if (shieldRules) {
        shieldRules.forEach(item => {
          item.AlarmName = alarmMapper[item.UniquelyIdentifies].AlarmName;
          delete alarmMapper[item.UniquelyIdentifies];
        })
        maskedData.value = shieldRules;
      }
      unmaskedData.value = Object.values(alarmMapper);
    }

    const refresh = async () => {
      generateSolutionMapper();
      generateAlarmMapper();

      let { data } = await queryAlarmShieldRules();
      computeMaskedAndUnmaskedDataByShieldRules(data?.AlarmShieldMessages ?? [])
    }

    onMounted(async () => {
      await loadJson();
      await refresh();
    })

    return {
      activeName,
      maskedData,
      unmaskedData,
      refresh,
    }
  },
});
</script>

<style lang="scss" scoped>

</style>