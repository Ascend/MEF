<template>
  <div class="top-block">
    <span class="tab-name">{{ $t('safety.loginRules.tabName') }}</span>
    <img class="info-icon" src="@/assets/img/common/question.svg" alt=""/>
    <span class="rule-tip">{{ $t('safety.loginRules.ruleTip') }}</span>
  </div>
  <div class="container">
    <div>
      <el-button type='primary' @click='clickCreate'>
        <img src='@/assets/img/manager/add.svg' alt='' style='height: 14px; width: 14px; margin-right: 8px;'/>
        {{ $t('safety.loginRules.addLoginRule') }}
      </el-button>
      <el-button @click='onClickImport'>{{ $t('common.import') }}</el-button>
      <el-button @click='onClickExport'>{{ $t('common.export') }}</el-button>
      <span style="float: right;">
        <el-button style="width: 32px; height: 32px; margin-left: 6px;" @click="refreshTable">
          <img style="width: 14px; height: 14px" alt="" src="@/assets/img/common/refresh.svg" />
        </el-button>
      </span>
    </div>
    <el-table
      :data="rulesList"
      style="width: 100%; margin-top: 30px;"
      height="600"
    >
      <el-table-column :label="$t('safety.loginRules.time')">
        <template #default='scope'>
          <span>{{ ruleFormatter.time(scope.row) }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('safety.loginRules.networkSegment')" prop="ip_addr" />
      <el-table-column :label="$t('safety.loginRules.mac')" prop="mac_addr" />
      <el-table-column
        :label="$t('safety.loginRules.status')"
        prop="enable"
        :filters="statusFilter"
        :filter-method="filterStatus"
      >
        <template #default='scope'>
          <el-switch v-model="scope.row.enable"
                     :active-value="'true'"
                     :before-change="beforeChangeRuleStatus"
                     @click="onClickSwitchBtn(scope.row.index)"
                     :inactive-value="'false'"></el-switch>
        </template>
      </el-table-column>
      <el-table-column prop='operations' :label='$t("common.operations")'>
        <template #default='scope'>
          <el-button link type='primary' size='small' @click='onEditRule(scope.row.index)'>
            {{ $t('common.modify') }}
          </el-button>
          <el-button link type='primary' size='small' @click='onDeleteRuleBtnClick(scope.row.index)'>
            {{ $t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
  <edit-login-rule-dialog
    :title='$t("safety.loginRules.addLoginRule")'
    v-if="isShowCreate"
    :is-show='isShowCreate'
    :rules="rulesList"
    @refresh="onCreateRefresh"
    @cancel="onCreateCancel"
  ></edit-login-rule-dialog>

  <edit-login-rule-dialog
    :title='$t("safety.loginRules.editLoginRule")'
    v-if="isShowEdit"
    :is-show='isShowEdit'
    :rules="rulesList"
    :edit-rule-index="editRuleIndex"
    @refresh="onEditRefresh"
    @cancel="onEditCancel"
  ></edit-login-rule-dialog>

  <import-login-rule-dialog
    v-if="isShowImportDialog"
    :is-show="isShowImportDialog"
    :title='$t("safety.loginRules.importLoginRule")'
    @submit="onImportSubmit"
    @cancel="isShowImportDialog = false"
  ></import-login-rule-dialog>

  <confirm-password-dialog
    v-if="isShowConfirmPasswordSwitch"
    :is-show="isShowConfirmPasswordSwitch"
    :title='$t("common.confirmPassword")'
    :tip='$t("safety.loginRules.switchEditLoginRule")'
    @submitConfirm="onConfirmPasswordSwitchSubmit"
    @cancel="isShowConfirmPasswordSwitch = false"
  ></confirm-password-dialog>

  <confirm-password-dialog
    v-if="isShowConfirmPasswordDialog"
    :is-show="isShowConfirmPasswordDialog"
    :title='$t("common.confirmPassword")'
    :tip='$t("safety.loginRules.deleteLoginRuleTip")'
    @submitConfirm="onConfirmPasswordDeleteSubmit"
    @cancel="isShowConfirmPasswordDialog = false"
  ></confirm-password-dialog>

</template>

<script>
import { defineComponent, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { exportLoginRules, modifyLoginRules, queryLoginRules } from '@/api/safety';
import EditLoginRuleDialog from '@/views/setting/safety/components/EditLoginRuleDialog.vue';
import ConfirmPasswordDialog from '@/views/setting/safety/components/ConfirmPasswordDialog.vue';
import { AppMessage, deepCopyDictList, showErrorAlert, handleOperationResponseError } from '@/utils/commonMethods';
import ImportLoginRuleDialog from '@/views/setting/safety/components/ImportLoginRuleDialog.vue';
import { saveFile } from '@/utils/downloadFile';
import constants from '@/utils/constants';

export default defineComponent({
  name: 'LoginRule',
  components: {
    ImportLoginRuleDialog,
    ConfirmPasswordDialog,
    EditLoginRuleDialog,
  },
  setup() {
    const { t } = useI18n();
    // list
    const rulesList = ref([]);
    const filterStatus = (value, row) => row.enable === value
    const statusFilter = [
      {
        text: t('safety.loginRules.statusOn'),
        value: 'true',
      },
      {
        text: t('safety.loginRules.statusOff'),
        value: 'false',
      },
    ];
    const refreshTable = async () => {
      let resp = await queryLoginRules();
      for (let i = 0; i < resp?.data?.load_cfg?.length; i++) {
        resp.data.load_cfg[i].index = i;
      }
      rulesList.value = resp.data.load_cfg;
    }
    // create
    const isShowCreate = ref(false);
    const clickCreate = () => {
      if (rulesList.value.length + 1 > constants.MAX_LOGIN_RULE_NUM_LIMIT) {
        showErrorAlert(t('safety.loginRules.createNumLimitTip', { limit: constants.MAX_LOGIN_RULE_NUM_LIMIT }));
        return;
      }
      isShowCreate.value = true;
    }
    const onCreateRefresh = async () => {
      isShowCreate.value = false;
      await refreshTable();
    }
    const onCreateCancel = () => {
      isShowCreate.value = false;
    }
    // edit
    const isShowEdit = ref(false);
    const editRuleIndex = ref();
    const onEditRule = (index) => {
      editRuleIndex.value = index
      isShowEdit.value = true;
    }
    const onEditRefresh = async () => {
      isShowEdit.value = false;
      await refreshTable();
    }
    const onEditCancel = () => {
      isShowEdit.value = false;
    }
    // edit by switch
    const beforeChangeRuleStatus = () => {
      return new Promise((resolve) => {
        resolve(false);
      })
    }
    const isShowConfirmPasswordSwitch = ref(false);
    const switchRuleIndex = ref();
    const onClickSwitchBtn = (index) => {
      switchRuleIndex.value = index
      isShowConfirmPasswordSwitch.value = true;
    }
    const onConfirmPasswordSwitchSubmit = async (password) => {
      try {
        let loadCfg = deepCopyDictList(rulesList.value);
        let status = loadCfg[switchRuleIndex.value].enable;
        loadCfg[switchRuleIndex.value].enable = status === 'true' ? 'false' : 'true';
        await modifyLoginRules({
          'Password': password,
          'load_cfg': loadCfg,
        });
        AppMessage.success(t('safety.loginRules.switchSuccessTip'));
        await refreshTable();
      } catch (err) {
        handleOperationResponseError(err);
      } finally {
        isShowConfirmPasswordSwitch.value = false;
      }
    }
    // delete
    const isShowConfirmPasswordDialog = ref(false);
    const deleteRuleIndex = ref();
    const onDeleteRuleBtnClick = (index) => {
      deleteRuleIndex.value = index
      isShowConfirmPasswordDialog.value = true;
    }
    const onConfirmPasswordDeleteSubmit = async (password) => {
      try {
        let loadCfg = deepCopyDictList(rulesList.value);
        loadCfg.splice(deleteRuleIndex.value, 1);
        await modifyLoginRules({
          'Password': password,
          'load_cfg': loadCfg,
        });
        await refreshTable();
        AppMessage.success(t('common.deleteSuccessfully'));
      } catch (err) {
        handleOperationResponseError(err);
      } finally {
        isShowConfirmPasswordDialog.value = false;
      }
    }
    // import
    const isShowImportDialog = ref(false);
    const onClickImport = () => {
      isShowImportDialog.value = true;
    }
    const onImportSubmit = () => {
      isShowImportDialog.value = false;
      refreshTable();
    }
    // export
    const onClickExport = async () => {
      try {
        let resp = await exportLoginRules({});
        let fileName = `rule_${new Date().getTime()}.ini`;
        saveFile(resp.data, fileName);
      } catch (err) {
        handleOperationResponseError(err, t('safety.loginRules.exportFailed'));
      }
    }

    const ruleFormatter = {
      time: (row) => {
        if (!row.start_time) {
          return '-';
        }
        return `${row.start_time} ~ ${row.end_time}`
      },
    }

    onMounted(async () => {
      await refreshTable();
    })

    return {
      ruleFormatter,
      rulesList,
      isShowCreate,
      isShowImportDialog,
      onCreateRefresh,
      onImportSubmit,
      isShowEdit,
      onEditRefresh,
      isShowConfirmPasswordDialog,
      onConfirmPasswordDeleteSubmit,
      editRuleIndex,
      statusFilter,
      onCreateCancel,
      onEditCancel,
      clickCreate,
      onClickImport,
      onClickExport,
      clickExport: onClickExport,
      refreshTable,
      onEditRule,
      onDeleteRuleBtnClick,
      filterStatus,
      beforeChangeRuleStatus,
      onClickSwitchBtn,
      isShowConfirmPasswordSwitch,
      switchRuleIndex,
      onConfirmPasswordSwitchSubmit,
    }
  },
})
</script>

<style lang="scss" scoped>
.top-block {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}


.tab-name {
  font-size: 16px;
  color: #FFFFFE;
  line-height: 24px;
  font-weight: 700;
  margin-right: 20px;
  width: 120px;
}

.info-icon {
  width: 14px;
  height: 14px;
}

.rule-tip {
  font-size: 12px;
  color: var(--el-customize-tip-text-color);
  line-height: 16px;
  font-weight: 400;
  margin-left: 10px;
}

#input-time {
  :deep(.el-form-item__content) {
    display: block;
  }
}

.container {
  padding: 24px;
  background: var(--el-bg-color-overlay);
  border-radius: 4px;
  height: 76vh;
}
</style>