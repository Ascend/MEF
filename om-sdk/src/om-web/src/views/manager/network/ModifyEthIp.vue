<template>
  <el-dialog
    :modelValue='isShow'
    :title='dialogTitle'
    @close='handleCancelAddIp'
    style="width: 600px; "
    :center="true"
    :close-on-click-modal="false"
  >
    <div class='same-segment-tip'>
      <img src='@/assets/img/common/info.svg' alt=''/>
      <span>
        {{ $t('network.networkWired.sameSegmentTip', { userGuide: configJson?.userGuide[getLanguageDefaultChinese()] }) }}
      </span>
    </div>
    <el-form
      ref='addIpFormRef'
      :model='addIpForm'
      :rules='addIpFormRule'
      style='width: 460px; '
      class='add-ip-form'
      label-position='left'
      label-width="auto"
      :hide-required-asterisk="true"
    >
      <el-form-item prop='Tag' class="not-required" :label="$t('network.networkWired.function')">
        <el-input v-model='addIpForm.Tag'/>
      </el-form-item>
      <el-form-item prop='Address' class="required" :label="$t('network.common.ipAddress')">
        <el-input v-model='addIpForm.Address' @blur="blurOnAddress"/>
      </el-form-item>
      <el-form-item prop='SubnetMask' class="required" :label="$t('network.networkWired.subnetMask')">
        <el-input v-model='addIpForm.SubnetMask'/>
      </el-form-item>
      <el-form-item
        prop='Gateway'
        class="not-required"
        :label="$t('network.networkWired.defaultGateway')"
        style="margin-top: 40px;"
      >
        <el-tooltip
          effect='dark'
          :content='$t("network.networkWired.modifyDefaultGatewayTip")'
          placement='right'
        >
          <el-input v-model='addIpForm.Gateway' />
        </el-tooltip>
      </el-form-item>
      <el-form-item prop='VlanId' class="not-required" :label="$t('network.networkWired.vlanId')">
        <el-tooltip
          effect='dark'
          :content='$t("network.networkWired.modifyVlanIdTip")'
          placement='right'
        >
          <el-input v-model='addIpForm.VlanId' />
        </el-tooltip>
      </el-form-item>
      <el-form-item prop='AddressOrigin' class="not-required" :label="$t('network.networkWired.addressOrigin')" >
        <el-input v-model='addIpForm.AddressOrigin' disabled />
      </el-form-item>
      <el-form-item prop='ConnectTest' class="required" :label="$t('network.networkWired.connectTest')">
        <el-radio-group v-model='addIpForm.ConnectTest'>
          <el-radio :label='true'>{{ $t('network.networkWired.testText') }}</el-radio>
          <el-radio :label='false'>{{ $t('network.networkWired.notTestText') }}</el-radio>
        </el-radio-group>
        <div v-if='!addIpForm.ConnectTest' class="not-test-tip">
          {{ $t("network.networkWired.notTestTip") }}
        </div>
      </el-form-item>
      <el-form-item
        v-if='addIpForm.ConnectTest'
        prop='RemoteTestIp'
        class="required"
        :label="$t('network.networkWired.remoteTestIp')">
        <el-input v-model='addIpForm.RemoteTestIp' />
      </el-form-item>
    </el-form>
    <div style="text-align: center; margin-top: 20px;">
      <el-button type='primary' @click='handleConfirmAddIp'>
        {{ $t('common.confirm') }}
      </el-button>
      <el-button @click='handleCancelAddIp'>{{ $t('common.cancel') }}</el-button>
    </div>
  </el-dialog>
</template>

<script>
import { ref, defineComponent, reactive, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';

import { fetchJson } from '@/api/common';
import { validateIp, validateNetworkUsage, validateVlanId } from '@/utils/validator';
import { getLanguageDefaultChinese, showWarningAlert } from '@/utils/commonMethods';

export default defineComponent({
  name: 'ModifyEthIp',
  props: {
    selectedEthItem: {
      type: Object,
      default: null,
    },
  },
  setup(props, ctx) {
    const { t } = useI18n()
    const addIpFormRef = ref()
    let dialogTitle = ref('')
    let isShow = ref(props.selectedEthItem.isShow)
    const configJson = ref();
    let addIpForm = ref({
      Tag: 'web',
      Address: '',
      SubnetMask: '',
      Gateway: '',
      VlanId: null,
      ConnectTest: true,
      RemoteTestIp: '',
      AddressOrigin: 'Static',
    })

    watch(
      props.selectedEthItem,
      () => {
        isShow.value = props.selectedEthItem.isShow;
        let currItem = props.selectedEthItem.currEth.ethData[props.selectedEthItem.currIpIndex];
        addIpForm.value = {
          Tag: currItem?.Tag || 'web',
          Address: currItem?.Address,
          SubnetMask: currItem?.SubnetMask,
          Gateway: currItem?.Gateway,
          VlanId: currItem?.VlanId,
          ConnectTest: true,
          RemoteTestIp: '',
          AddressOrigin: 'Static',
        }
        dialogTitle.value = {
          add: t('common.add'),
          edit: t('common.modify'),
        }[props.selectedEthItem.operateType] + getLanguageDefaultChinese() === 'zh' ? '' : ' ' + t('network.common.ipAddress');
      }
    )

    const handleConfirmAddIp = () => {
      let sameIpList = getSameIpList()
      if (sameIpList.length > 0) {
        let otherEthItem = props.selectedEthItem.currEth.isEth0 ? props.selectedEthItem.wiredData.eth1 : props.selectedEthItem.wiredData.eth0;
        showWarningAlert(
          t('network.networkWired.sameSegmentIpTip',
            {
              ethNum: otherEthItem.label,
              ipAddress: sameIpList.join(', '),
            }
          )
        )
      }

      if (!addIpFormRef.value) {
        return
      }
      addIpFormRef.value.validate(async (valid) => {
        if (!valid) {
          return
        }
        addIpForm.value.AddressOrigin = 'Static'
        addIpForm.value.Tag = addIpForm.value.Tag || ''
        let originAddresses = props.selectedEthItem.currEth.ethData.concat([]);

        if (addIpForm.value.Gateway) {
          originAddresses.forEach(item => {
            item.Gateway = '';
          })
        }

        if (props.selectedEthItem.operateType === 'add') {
          originAddresses.push(addIpForm.value)
        } else {
          originAddresses.splice(props.selectedEthItem.currIpIndex, 1, addIpForm.value)
        }

        originAddresses.forEach(item => {
          if (!item.VlanId) {
            delete item.VlanId
          } else {
            item.VlanId = parseInt(item.VlanId)
          }
          if (item.Gateway === null || item.Gateway === undefined) {
            item.Gateway = ''
          }
        })
        ctx.emit('temperatureSaveIpAddresses', originAddresses)
        ctx.emit('refresh')
      })
    }

    const handleCancelAddIp = () => {
      addIpFormRef.value.resetFields();
      ctx.emit('cancelAddIp')
    }

    const convertToAllBinary = (ipAddress) => {
      let ipSplit = ipAddress?.split('.');
      let ipBinaryStr = '';
      for (let index = 0; index < 4; index++) {
        let ipBinarySegment = parseInt(ipSplit[index], 10).toString(2);
        for (let count = 0; count < (8 - ipBinarySegment.length); count++) {
          ipBinaryStr += '0';
        }
        ipBinaryStr += ipBinarySegment + '.';
      }
      return ipBinaryStr;
    }

    const checkSameNetMask = (ipv4, gateway, netMask) => {
      let ipv4Binary = convertToAllBinary(ipv4);
      let gatewayBinary = convertToAllBinary(gateway);
      let netMaskBinary = convertToAllBinary(netMask);

      for (let index = 0; index < netMaskBinary.length; index++) {
        if (netMaskBinary[index] === '1') {
          if (ipv4Binary[index] !== gatewayBinary[index]) {
            return false;
          }
        }
      }
      return true;
    };

    const isEqualIPAddress = (Addresses1, Addresses2) => {
      let res1 = [];
      let res2 = [];
      let addr1 = Addresses1.Address?.split('.');
      let addr2 = Addresses2.Address?.split('.');
      let mask1 = Addresses1.SubnetMask?.split('.');
      let mask2 = Addresses2.SubnetMask?.split('.');

      for (let i = 0, ilen = addr1.length; i < ilen; i++) {
        res1.push(parseInt(addr1[i]) & parseInt(mask1[i]));
        res2.push(parseInt(addr2[i]) & parseInt(mask2[i]));
      }
      return res1.join('.') === res2.join('.');
    }

    const validateGateway = (rule, value, callback) => {
      if (!value) {
        callback();
        return;
      }

      if (!addIpForm.value.Address) {
        callback(new Error(t('network.networkWired.ipAddressFirstTip')))
        return;
      }

      if (!addIpForm.value.SubnetMask) {
        callback(new Error(t('network.networkWired.subnetMaskFirstTip')))
        return;
      }

      if (checkSameNetMask(addIpForm.value.Address, value, addIpForm.value.SubnetMask)) {
        callback();
      } else {
        callback(new Error(t('network.networkWired.gatewayNotMatchTip')));
      }
    }

    const getSameIpList = () => {
      let sameIpList = [];
      if (!addIpForm.value.Address || !addIpForm.value.SubnetMask) {
        return sameIpList;
      }
      let otherEthItem = props.selectedEthItem.currEth.isEth0 ? props.selectedEthItem.wiredData.eth1 : props.selectedEthItem.wiredData.eth0;
      for (let i = 0; i < otherEthItem.ethData?.length; i++) {
        if (isEqualIPAddress(otherEthItem.ethData[i], addIpForm.value)) {
          sameIpList.push(otherEthItem.ethData[i].Address)
        }
      }
      return sameIpList
    }

    const isExistIp = (ethItem) => {
      if (!ethItem.ethData) {
        return;
      }
      for (let i = 0; i < ethItem.ethData.length; i++) {
        if (ethItem.ethData[i].Address === addIpForm.value.Address) {
          showWarningAlert(t('network.networkWired.existOtherEthTip'))
          return;
        }
      }
    }
    const blurOnAddress = () => {
      // 校验当前配置的ip地址是否存在
      isExistIp(props.selectedEthItem.wiredData.eth0)
      isExistIp(props.selectedEthItem.wiredData.eth1)
    }

    const validateSubnetMask = (rule, value, callback) => {
      if (!value) {
        callback()
        return;
      }
      const pattern = /^(254|252|248|240|224|192|128|0)\.0\.0\.0|255\.(254|252|248|240|224|192|128|0)\.0\.0|255\.255\.(254|252|248|240|224|192|128|0)\.0|255\.255\.255\.(255|254|252|248|240|224|192|128|0)$/;
      if (!pattern.test(value)) {
        callback(new Error(t('common.errorFormatTip')))
        return;
      }
      callback();
    }

    const addIpFormRule = reactive({
      Tag: [
        {
          validator: validateNetworkUsage,
          trigger: 'blur',
        },
      ],
      Address: [
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
      SubnetMask: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateSubnetMask,
          trigger: 'blur',
        },
      ],
      Gateway: [
        {
          validator: validateIp,
          trigger: 'blur',
        },
        {
          validator: validateGateway,
          trigger: 'blur',
        },
      ],
      RemoteTestIp: [
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
      VlanId: [
        {
          validator: validateVlanId,
          trigger: 'blur',
        },
      ],
    })

    onMounted(async () => {
      configJson.value = await fetchJson('/WhiteboxConfig/json/config.json');
    })

    return {
      configJson,
      getLanguageDefaultChinese,
      addIpForm,
      addIpFormRef,
      addIpFormRule,
      isShow,
      dialogTitle,
      handleConfirmAddIp,
      handleCancelAddIp,
      blurOnAddress,
    }
  },
})
</script>

<style scoped>
.same-segment-tip {
  padding: 10px 20px;
  margin-bottom: 20px;
  background-color: var(--el-bg-color-tip);
}

.same-segment-tip img {
  margin-right: 10px;
  height: 16px;
  width: 16px;
}

.same-segment-tip span {
  word-break: break-word;
}

.not-test-tip {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  line-height: 16px;
}
</style>