<template>
  <el-dialog
    :modelValue="props.mountPartitionData.isShow"
    :close-on-click-modal="false"
    style="width: 550px;"
    @close="cancelUnmountPartition"
  >
    <app-tip :tip-type='"warn"' v-if="isEmmc()" :tip-text='$t("disk.localDiskManagement.emmcTip")' />
    <el-form
      :model="partitionForm"
      label-position="right"
      label-width="auto"
      :rules="partitionFormRule"
      ref="partitionFormRef"
      style="width: 450px; margin-top: 20px;"
      v-if="!props.mountPartitionData.mountPath"
      @submit.native.prevent
    >
      <el-form-item :label="$t('disk.common.mountPath')" prop="mountPath">
        <el-input v-model="partitionForm.mountPath" @keyup.enter.native="confirmMountPartition(partitionFormRef)" />
      </el-form-item>
      <div style="text-align: center; margin-top: 60px;">
        <el-button type="primary" @click="confirmMountPartition(partitionFormRef)">
          {{ $t('common.confirm') }}
        </el-button>
        <el-button @click="cancelMountPartition(partitionFormRef)">
          {{ $t('common.cancel') }}
        </el-button>
      </div>
    </el-form>
    <div v-else>
      <el-row style="margin-bottom: 40px; word-break: break-word;">
        <el-col :span="2">
          <img class="alarm-img" alt="" src="@/assets/img/alarm.svg" />
        </el-col>
        <el-col :span="20">
          {{ $t('disk.localDiskManagement.mountPartitionTip.unmountTip', { 'partition': props.mountPartitionData?.partitionId }) }}
        </el-col>
      </el-row>
      <div style="margin-left: calc(50% - 86px);">
        <el-button type="primary" @click="confirmUnmountPartition">
          {{ $t('common.confirm') }}
        </el-button>
        <el-button @click="cancelUnmountPartition">
          {{ $t('common.cancel') }}
        </el-button>
      </div>
    </div>
  
  </el-dialog>
</template>

<script>
import { defineComponent, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { handleOperationResponseError } from '@/utils/commonMethods';
import { mountPartitions, unmountPartitions } from '@/api/drive';
import AppTip from '@/components/AppTip.vue';
import { validateServerPath } from '@/utils/validator';

export default defineComponent({
  name: 'MountAndUnmount',
  components: {
    AppTip,
  },
  props: {
    mountPartitionData: Object,
  },
  setup(props, ctx) {
    const { t } = useI18n();
    const partitionFormRef = ref();
    
    const partitionForm = ref({
      mountPath: '',
    });
    
    const partitionFormRule = reactive({
      mountPath: [
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
    });
    
    const confirmMountPartition = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        
        let params = {
          MountPath: partitionForm.value.mountPath,
          PartitionID: props.mountPartitionData.partitionId,
        };
        try {
          await mountPartitions(params);
          ctx.emit('mountPartitionSuccessfully', 'mount');
        } catch (err) {
          ctx.emit('cancelMountAndUnmountPartition');
          handleOperationResponseError(err);
        }
        
        formElement.resetFields();
      });
    };
    
    const cancelMountPartition = (formElement) => {
      if (!formElement) {
        return;
      }
      formElement.resetFields();
      ctx.emit('cancelMountAndUnmountPartition');
    };
    
    const confirmUnmountPartition = async () => {
      try {
        await unmountPartitions({ PartitionID: props.mountPartitionData.partitionId });
        ctx.emit('mountPartitionSuccessfully', 'unmount');
      } catch (err) {
        ctx.emit('cancelMountAndUnmountPartition');
        handleOperationResponseError(err);
      }
    };
    
    const cancelUnmountPartition = () => {
      ctx.emit('cancelMountAndUnmountPartition');
    };
    
    const isEmmc = () => props.mountPartitionData?.partitionId?.indexOf('mmcblk0') !== -1;
    
    return {
      props,
      partitionForm,
      partitionFormRef,
      partitionFormRule,
      confirmMountPartition,
      cancelMountPartition,
      confirmUnmountPartition,
      cancelUnmountPartition,
      isEmmc,
    };
  },
});
</script>
