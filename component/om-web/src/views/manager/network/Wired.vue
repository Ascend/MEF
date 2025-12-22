<template>
  <div v-loading="loading">
    <div style="height: 40vh; padding-bottom: 20px;" v-for='eth in wiredData' v-bind:key='eth.label'>
      <div class='info-block'>
        <span class='tab-title' style="margin-bottom: 0; margin-right: 10px;">{{ $t('network.networkWired.configNetwork') }} ({{ eth.label }})</span>
        <app-popover v-if="!eth.isEth0" :text="$t('network.networkWired.ethOtherTip')" />
        <img class='status-img' alt='' src='@/assets/img/manager/status.svg' />
        <span class="network-status">{{ $t('network.networkWired.checkNetworkStatus') }}</span>
        <span class='status-text' >
          {{ eth.linkStatus }}
          {{ eth.workMode }}
        </span>
        <el-button class='check-btn' size='small' @click='clickTestNetwork(eth)'>
          {{ $t('network.common.test') }}
        </el-button>
      </div>

      <div class='data-block'>
        <el-button type='primary' @click='openAddIpDialog(eth)'>
          <img alt='' src='@/assets/img/manager/add.svg' style='width: 12px; height: 12px; margin-right: 2px;' />
          {{ $t('network.networkWired.configIpAddress') }}
        </el-button>
        <el-button type='primary' @click='clickSave(eth)'>
          {{ $t('common.save') }}
        </el-button>
        <el-table :data='eth.ethData' style='width: 96%; margin-top: 20px;' :empty-text='$t("network.networkWired.empty")'>
          <el-table-column prop='Address' :label='$t("network.common.ipAddress")' />
          <el-table-column prop='SubnetMask' :label='$t("network.networkWired.subnetMask")' />
          <el-table-column prop='Gateway' :label='$t("network.networkWired.defaultGateway")' />
          <el-table-column prop='VlanId' :label='$t("network.networkWired.vlanId")' />
          <el-table-column prop='Tag' :label='$t("network.networkWired.function")' />
          <el-table-column prop='AddressOrigin' :label='$t("network.networkWired.addressOrigin")' />
          <el-table-column prop='operations' :label='$t("common.operations")'>
            <template #default='scope'>
              <el-button link type='primary' size='small' @click='modifyEthIp(eth, scope.$index)'>
                {{ $t('common.modify') }}
              </el-button>
              <el-button link type='primary' size='small' @click='deleteEthIp(eth, scope.$index)'>
                {{ $t('common.delete') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
  <el-dialog
    :title='$t("network.networkWired.deleteReminder")'
    v-model='selectedDeleteEthItem.isShowDeleteDialog'
    :width="500"
  >
    <div style='display: flex; align-items: center;'>
      <img src='@/assets/img/alarm.svg' alt=''/>
      {{ selectedDeleteEthItem.deleteTip }}
    </div>
    <template #footer>
      <span class='dialog-footer'>
        <el-button type='primary' @click='confirmDelete'>
          {{ $t('common.confirm') }}
        </el-button>
        <el-button @click='cancelDelete'>
          {{ $t('common.cancel') }}
        </el-button>
      </span>
    </template>
  </el-dialog>
  <modify-eth-ip :selected-eth-item='selectedEthItem' @cancelAddIp='cancelAddIp' @temperatureSaveIpAddresses="temperatureSaveIpAddresses" />
</template>

<script>
import { defineComponent, onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessageBox } from 'element-plus';

import AppPopover from '@/components/AppPopover.vue';
import { getUrls } from '@/api/http';
import { queryAllEthernetInfo, querySingleEthernetInfo } from '@/api/network';
import { modifyByOdataUrl } from '@/api/common';
import ModifyEthIp from '@/views/manager/network/ModifyEthIp.vue';
import constants from '@/utils/constants';
import { AppMessage, handleOperationResponseError, isFulfilled, checkUrlsResponse } from '@/utils/commonMethods';

export default defineComponent({
  name: 'WiredNetwork',
  components: {
    ModifyEthIp,
    AppPopover,
  },
  setup() {
    const { t } = useI18n()
    const loading = ref(false);
    const selectedEthItem = reactive({
      isShow: false,
      operateType: '',
      currEth: null,
      currIpIndex: null,
      wiredData: null,
    })

    const selectedDeleteEthItem = ref({
      isShowDeleteDialog: false,
      deleteTip: '',
      currEth: null,
      currIpIndex: null,
    })

    const openAddIpDialog = (eth) => {
      if (Object.keys(eth.ethData).length === constants.MAX_ETH_NUM) {
        AppMessage.warning(t('network.networkWired.ipMeetLimitErrorTip'))
      } else {
        selectedEthItem.isShow = true
        selectedEthItem.operateType = 'add'
        selectedEthItem.currEth = eth
        selectedEthItem.currIpIndex = null
        selectedEthItem.wiredData = wiredData
      }
    }

    const modifyEthIp = (eth, ipIndex) => {
      selectedEthItem.isShow = true
      selectedEthItem.operateType = 'edit'
      selectedEthItem.currEth = eth
      selectedEthItem.currIpIndex = ipIndex
      selectedEthItem.wiredData = wiredData
    }

    const deleteEthIp = (eth, ipIndex) => {
      if (Object.keys(eth.ethData).length === constants.MIN_ETH_NUM) {
        AppMessage.warning(t('network.networkWired.ipEmptyErrorTip'))
      } else {
        selectedDeleteEthItem.value.deleteTip = t(
          'network.networkWired.deleteTip',
          { ipAddress: eth.ethData[ipIndex].Address }
        );
        selectedDeleteEthItem.value.isShowDeleteDialog = true;
        selectedDeleteEthItem.value.currEth = eth;
        selectedDeleteEthItem.value.currIpIndex = ipIndex;
      }
    }

    const wiredData = ref({
      eth0: {
        label: t('network.networkWired.eth0'),
      },
      eth1: {
        label: t('network.networkWired.ethOther'),
      },
    })

    const nameMapper = {
      eth0: t('network.networkWired.eth0'),
      eth1: t('network.networkWired.ethOther'),
    }
    const fetchEthernetMembers = async () => {
      let { data } = await queryAllEthernetInfo(false);
      return data?.Members;
    }

    const initWiredData = async (members) => {
      if (!members) {
        loading.value = false;
        return
      };

      let params = members.map(item => (
        {
          url: item['@odata.id'],
          isShowLoading: false,
        }
      ))

      let allResponse = await getUrls(params);
      allResponse.filter(item => isFulfilled(item)).map(item => {
        let { data } = item.value;
        if (!data) {
          return;
        }
        wiredData.value[data.Name] = {
          label: nameMapper[data.Name],
          isEth0: data.Name?.toLowerCase().indexOf('eth0') !== -1,
          id: data?.Id,
          odata: data['@odata.id'],
          isShowAddIpDialog: false,
          linkStatus: data?.LinkStatus === 'LinkUp' ? t('network.networkWired.statusConnected') : t('network.networkWired.statusUnconnected'),
          ethData: data?.IPv4Addresses || [],
          workMode: '',
        }
      })
      checkUrlsResponse(allResponse);
    }

    const init = async () => {
      loading.value = true;
      let members = await fetchEthernetMembers();
      await initWiredData(members);
      loading.value = false;
    }

    onMounted(async () => {
      await init();
    })

    const clickTestNetwork = async (eth) => {
      let { data } = await querySingleEthernetInfo(eth.id);
      eth.workMode = data?.Oem?.WorkMode;
    }

    const cancelAddIp = async () => {
      selectedEthItem.isShow = false;
    }

    const confirmDelete = async () => {
      let ipIndex = selectedDeleteEthItem.value.currIpIndex;
      let ipv4 = [].concat(selectedDeleteEthItem.value.currEth.ethData);
      ipv4.splice(ipIndex, 1);

      if (selectedDeleteEthItem.value.currEth.isEth0) {
        wiredData.value.eth0.ethData = ipv4
      } else {
        wiredData.value.eth1.ethData = ipv4
      }
      cancelDelete()
    }

    const cancelDelete = () => {
      selectedDeleteEthItem.value = {
        isShowDeleteDialog: false,
        deleteTip: '',
        currEth: null,
        currIpIndex: null,
      }
    }

    const clickSave = async (eth) => {
      try {
        await ElMessageBox.confirm(
          t('network.networkWired.modifyTip'),
          t('common.reminder'),
          { type: 'warning' }
        )
        try {
          loading.value = true;
          let { data } = await modifyByOdataUrl(
            eth.odata,
            { IPv4Addresses: eth.ethData },
            constants.MINUTE_TIMEOUT * 3,
            false
          );
          loading.value = false;
          if (data?.Oem?.TaskPercentage.toLowerCase() === 'ok') {
            AppMessage.success(t('common.saveSuccessfully') + t('network.networkWired.restartNginxTip'));
          } else {
            AppMessage.success(t('common.saveSuccessfully') + t('network.networkWired.configGatewayFailedTip'));
          }
        } catch (err) {
          loading.value = false;
          handleOperationResponseError(err);
        }
      } finally {
        await init();
      }
    }

    const temperatureSaveIpAddresses = (ipv4Addresses) => {
      if (selectedEthItem.currEth.isEth0) {
        wiredData.value.eth0.ethData = ipv4Addresses
      } else {
        wiredData.value.eth1.ethData = ipv4Addresses
      }
      selectedEthItem.isShow = false;
    }

    return {
      selectedEthItem,
      selectedDeleteEthItem,
      wiredData,
      loading,
      openAddIpDialog,
      modifyEthIp,
      deleteEthIp,
      clickTestNetwork,
      cancelAddIp,
      confirmDelete,
      cancelDelete,
      clickSave,
      temperatureSaveIpAddresses,
    }
  },
})
</script>

<style lang='scss' scoped>
.info-block {
  margin-top: 16px;
  display: flex;
  align-items: center;
}

.status-img {
  margin-left: 26px;
  margin-right: 10px;
}

.status-text {
  font-size: 14px;
  margin-left: 8px;
  color: var(--el-text-customize-color-primary);
  margin-right: 16px;
  width: 140px;
}

.data-block {
  width: 96%;
  height: 340px;
  background-color: var(--el-bg-color-overlay);
  margin-top: 2vh;
  padding: 1%;
  border-radius: 4px;
}

.network-status {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  line-height: 16px;
  font-weight: 400;
}
</style>