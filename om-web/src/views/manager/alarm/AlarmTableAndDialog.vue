<template>
  <div class='tab-title'>
    {{ tableName }}
  </div>
  <div class='container'>
    <el-button :disabled='selectedAlarm.length === 0' @click='isShowDialog = true'>
      {{ buttonText }}
    </el-button>
    <div style='float: right; display: flex; align-items: center;'>
      <el-input
        v-model='alarmQuery'
        @input='searchAlarmTable'
        clearable
        :placeholder="$t('alarm.common.searchPlaceholder')"
        style='width: 300px;'
      />
      <el-button style='width: 32px; height: 32px; margin-left: 20px; ' @click='refreshAlarmTable'>
        <img style='width: 14px; height: 14px' alt='' src='@/assets/img/common/refresh.svg' />
      </el-button>
    </div>
    <el-table
      :data="alarmData"
      style="width: 100%; margin-top: 20px;"
      max-height="630"
      ref="alarmDataRef"
      @selection-change="handleSelectionChange"
      row-key="UniquelyIdentifies"
      @sort-change="sortSeverity"
    >
      <el-table-column type="selection" width="55" :reserve-selection="true" />
      <el-table-column :label="$t('alarm.common.alarmId')" prop="AlarmId" />
      <el-table-column
        :label="$t('alarm.common.severity')"
        sortable="custom"
        prop='PerceivedSeverity'
      >
        <template #default='props'>
          <el-tag v-if="props.row.PerceivedSeverity === '0'" effect='dark' type='danger'>{{ $t('alarm.common.emergency') }}</el-tag>
          <el-tag v-if="props.row.PerceivedSeverity === '1'" effect='dark' type='warning'>{{ $t('alarm.common.serious') }}</el-tag>
          <el-tag v-if="props.row.PerceivedSeverity === '2'" effect='dark' :type='""'>{{ $t('alarm.common.normal') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('alarm.common.alarmObject')" prop='AlarmInstance' />
      <el-table-column :label="$t('alarm.common.alarmName')"  prop='AlarmName' />
      <el-table-column :label="$t('common.operations')">
        <template #default='scope'>
          <el-button
            link
            type='primary'
            size='small'
            @click='clickOperationButton(scope.row)'
          >
            {{ operationText }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      layout='total, sizes, prev, pager, next'
      :current-page='pageNum'
      :total='filterData.length'
      :page-size='pageSize'
      :page-sizes='[10, 20, 50, 100]'
      @size-change='handleSizeChange'
      @current-change='handleCurrentChange'
      style='position: absolute; bottom: 20px;'
    />
  </div>
  <el-dialog
    v-model='isShowDialog'
    :title='dialogTitle'
  >
    <div style='display: flex; align-items: center; word-break: break-word;'>
      <img src='@/assets/img/alarm.svg' alt='' style='margin-right: 10px;'/>
      {{ dialogText }}
    </div>
    <el-table
      :data='selectedAlarm'
      height='300'
      style='width: 100%; margin-top: 30px;'
    >
      <el-table-column :label="$t('alarm.common.alarmId')" prop='AlarmId' />
      <el-table-column :label="$t('alarm.common.severity')" prop='PerceivedSeverity'>
        <template #default='props'>
          <el-tag v-if='props.row.PerceivedSeverity === "0"' effect='dark' type='danger'>{{ $t('alarm.common.emergency') }}</el-tag>
          <el-tag v-if='props.row.PerceivedSeverity === "1"' effect='dark' type='warning'>{{ $t('alarm.common.serious') }}</el-tag>
          <el-tag v-if='props.row.PerceivedSeverity === "2"' effect='dark' :type="''">{{ $t('alarm.common.normal') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('alarm.common.alarmObject')" prop='AlarmInstance' />
      <el-table-column :label="$t('alarm.common.alarmName')" prop='AlarmName' />
    </el-table>
    <template #footer>
      <span class='dialog-footer'>
        <el-button type='primary' @click='confirmDialog'>
          {{ $t('common.confirm') }}
        </el-button>
        <el-button @click='cancelDialog'>
          {{ $t('common.cancel') }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import { ref, defineComponent, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { createAlarmShieldRules, cancelAlarmShieldRules } from '@/api/alarm';
import { AppMessage, handleOperationResponseError, showErrorAlert } from '@/utils/commonMethods';

export default defineComponent({
  name: 'AlarmTableAndDialog',
  props: {
    tableData: Array,
    tabName: String,
  },
  emits: ['refresh'],
  setup(props, ctx) {
    const { t } = useI18n();
    const isShowDialog = ref(false);
    const alarmQuery = ref();
    const alarmData = ref([]);
    const filterData = ref([]);
    const alarmDataRef = ref();
    const pageSize = ref(10);
    const pageNum = ref(1);
    const selectedAlarm = ref([]);
    const tableName = ref();
    const operationText = ref();
    const buttonText = ref();
    const dialogText = ref();
    const dialogTitle = ref();

    const refreshAlarmData = () => {
      alarmData.value = filterData.value.slice(
        pageSize.value * (pageNum.value - 1),
        pageSize.value * pageNum.value
      );
    }

    const searchAlarmTable = () => {
      alarmDataRef.value.clearSort();
      if (!alarmQuery.value) {
        filterData.value = [].concat(props.tableData);
      } else {
        filterData.value = props.tableData.filter((item) => {
          let hasAlarmId = item.AlarmId.indexOf(alarmQuery.value) !== -1;
          let hasAlarmInstance = item.AlarmInstance.indexOf(alarmQuery.value) !== -1;
          let hasAlarmName = item.AlarmName.indexOf(alarmQuery.value) !== -1;
          return Boolean(hasAlarmId || hasAlarmInstance || hasAlarmName)
        });
      }
      refreshAlarmData();
    }

    const refreshAlarmTable = async () => {
      await ctx.emit('refresh')
    }

    watch(
      () => props.tableData,
      () => {
        searchAlarmTable();
        pageNum.value = 1

        tableName.value = {
          unmasked: t('alarm.unmaskedAlarm.alarmList'),
          masked: t('alarm.maskedAlarm.alarmList'),
        }[props.tabName];

        operationText.value = {
          unmasked: t('alarm.unmaskedAlarm.createMask'),
          masked: t('alarm.maskedAlarm.cancelMask'),
        }[props.tabName];

        buttonText.value = operationText.value;
      }
    )

    watch(
      isShowDialog,
      () => {
        if (isShowDialog.value) {
          dialogText.value = {
            unmasked: t('alarm.unmaskedAlarm.createMaskTip', { num: selectedAlarm.value.length }),
            masked: t('alarm.maskedAlarm.cancelMaskTip', { num: selectedAlarm.value.length }),
          }[props.tabName];

          dialogTitle.value = {
            unmasked: t('alarm.common.maskReminder', { num: selectedAlarm.value.length }),
            masked: t('alarm.common.cancelReminder', { num: selectedAlarm.value.length }),
          }[props.tabName];
        }
      }
    )
    const handleSizeChange = (value) => {
      pageSize.value = value
      refreshAlarmData();
    }

    const handleCurrentChange = (value) => {
      pageNum.value = value
      refreshAlarmData();
    }

    let sortSeverityOrder = null;
    const sortSeverity = ({ column, prop, order }) => {
      sortSeverityOrder = order;
      if (order === 'descending') {
        filterData.value.sort((a, b) => parseInt(a[prop]) - parseInt(b[prop]));
      } else if (order === 'ascending') {
        filterData.value.sort((a, b) => parseInt(b[prop]) - parseInt(a[prop]));
      } else {
        searchAlarmTable();
      }
      refreshAlarmData();
    }

    const handleSelectionChange = (value) => {
      selectedAlarm.value = value;
    }

    const clickOperationButton = (row) => {
      selectedAlarm.value = [row];
      isShowDialog.value = true;
    }

    const confirmDialog = async () => {
      if (selectedAlarm.value.length > 256) {
        showErrorAlert(t('alarm.common.exceedLimit'))
        isShowDialog.value = false
        return;
      }

      let paramData = [].concat(selectedAlarm.value)
      paramData.forEach(item => {
        delete item.AlarmName
      })
      let params = {
        AlarmShieldMessages: paramData,
      }
      try {
        if (props.tabName === 'masked') {
          await cancelAlarmShieldRules(params)
        } else {
          await createAlarmShieldRules(params);
        }
        AppMessage.success(t('common.saveSuccessfully'));
        await ctx.emit('refresh');
        alarmDataRef.value?.clearSelection();
      } catch (err) {
        handleOperationResponseError(err)
      }
      isShowDialog.value = false
    }

    const cancelDialog = () => {
      isShowDialog.value = false
    }

    return {
      props,
      isShowDialog,
      selectedAlarm,
      alarmQuery,
      alarmData,
      filterData,
      alarmDataRef,
      pageSize,
      pageNum,
      tableName,
      operationText,
      buttonText,
      dialogText,
      dialogTitle,
      sortSeverity,
      searchAlarmTable,
      refreshAlarmTable,
      handleSizeChange,
      handleCurrentChange,
      handleSelectionChange,
      clickOperationButton,
      confirmDialog,
      cancelDialog,
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