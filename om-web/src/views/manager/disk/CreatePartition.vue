<template>
  <el-dialog
    :title="$t('disk.localDiskManagement.createPartition')"
    :modelValue="props.createPartitionData.isShow"
    :close-on-click-modal="false"
    style="width: 660px;"
    @close="cancelCreatePartition(partitionFormRef)"
  >
    <el-form
      :model="partitionForm"
      label-position="right"
      label-width="auto"
      :rules="partitionFormRule"
      ref="partitionFormRef"
      style="width: 600px; margin-top: 20px;"
    >
      <el-form-item :label="$t('disk.localDiskManagement.deviceType')" prop="deviceType">
        <el-input v-model="partitionForm.deviceType" disabled />
      </el-form-item>
      <el-form-item :label="$t('disk.common.capacityBytes')" prop="capacityBytes">
        <el-input v-model="partitionForm.capacityBytes" disabled />
      </el-form-item>
      <el-form-item :label="$t('disk.common.freeBytes')" prop="freeBytes">
        <el-row :gutter="4">
          <el-col :span="22">
            <el-input v-model="partitionForm.freeBytes" disabled style="width: 220px;" />
          </el-col>
          <el-col :span="2" style="padding-top: 6px;">
            <app-popover class="tooltip" :text="$t('disk.localDiskManagement.createPartitionTip.freeBytes')" />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item :label="$t('disk.localDiskManagement.partitionNum')" prop="number">
        <el-row :gutter="4">
          <el-col :span="22">
            <el-input-number style="width: 220px;" v-model="partitionForm.number" :min="1" :max="16" />
          </el-col>
          <el-col :span="2" style="padding-top: 6px;">
            <app-popover class="tooltip" :text="$t('disk.localDiskManagement.createPartitionTip.partitionNum')" />
          </el-col>
        </el-row>
      </el-form-item>
      <el-form-item :label="$t('disk.localDiskManagement.partition')" prop="capacity">
        <el-row :gutter="4">
          <el-col :span="14">
            <el-input
              v-model="partitionForm.capacity"
              style="width: 220px;"
            />
            GB
          </el-col>
          <el-col :span="9">
            <el-select v-model="partitionForm.fileSystem" disabled style="width: 150px;">
              <el-option
                v-for="item in fileSystemOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-col>
          <el-col :span="1" style="padding-top: 6px;">
            <app-popover class="tooltip" :text="$t('disk.localDiskManagement.createPartitionTip.fileSystem')" />
          </el-col>
        </el-row>
      </el-form-item>
      <div style="text-align: center;">
        <el-button type="primary" @click="confirmCreatePartition(partitionFormRef)">
          {{ $t('common.confirm') }}
        </el-button>
        <el-button @click="cancelCreatePartition(partitionFormRef)">
          {{ $t('common.cancel') }}
        </el-button>
      </div>
    </el-form>
  </el-dialog>
</template>

<script>
import { defineComponent, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { createPartitions } from '@/api/drive';
import AppPopover from '@/components/AppPopover.vue';
import { handleOperationResponseError } from '@/utils/commonMethods';
import AppTip from '@/components/AppTip.vue';

export default defineComponent({
  name: 'CreatePartition',
  components: {
    AppPopover,
    AppTip,
  },
  props: {
    createPartitionData: Object,
  },
  setup(props, ctx) {
    const { t } = useI18n();
    const fileSystemOptions = [
      {
        value: 'ext4',
        label: 'ext4',
      },
    ];
    const partitionForm = ref({
      deviceType: props.createPartitionData.currDisk?.deviceType,
      capacityBytes: props.createPartitionData.currDisk?.capacityBytes,
      freeBytes: props.createPartitionData.currDisk?.leftBytes,
      total: parseFloat(props.createPartitionData.currDisk?.leftBytes?.split(' ')[0]),
      number: 1,
      capacity: 0.5,
      fileSystem: 'ext4',
    });
    
    const partitionFormRef = ref();
    
    const validateNumber = (rule, value, callback) => {
      if (!value) {
        callback();
        return;
      }
      
      if (partitionForm.value.capacity && partitionForm.value.capacity * parseInt(value) > partitionForm.value.total) {
        callback(new Error(t('disk.localDiskManagement.createPartitionTip.numberExceedTip')));
        return;
      }

      if (partitionForm.value.capacity) {
        partitionFormRef.value.clearValidate(['capacity']);
      }
      callback();
    };
    
    const validateCapacity = (rule, value, callback) => {
      if (!value) {
        callback();
        return;
      }
      
      const pattern = /^((\d*(\.)?\d{1})|(\d+)(\.\d{0,1})?)$/;
      if (!pattern.test(value)) {
        callback(new Error(t('disk.localDiskManagement.createPartitionTip.capacityFormatErrorTip')));
        return;
      }
      if (parseFloat(value) < 0.5) {
        callback(new Error(t('disk.localDiskManagement.createPartitionTip.capacityTooLowTip')));
        return;
      }
      if (parseFloat(value) % 0.5 !== 0) {
        callback(new Error(t('disk.localDiskManagement.createPartitionTip.capacityNumErrorTip')));
        return;
      }
      if (partitionForm.value.number && partitionForm.value.number * parseFloat(value) > partitionForm.value.total) {
        callback(new Error(t('disk.localDiskManagement.createPartitionTip.capacityExceedTip')));
        return;
      }
      if (partitionForm.value.number) {
        partitionFormRef.value.clearValidate(['number']);
      }
      callback();
    };
    
    const partitionFormRule = reactive({
      number: [
        {
          required: true,
          message: t('common.wrongInput'),
          trigger: 'blur',
        },
        {
          validator: validateNumber,
          trigger: 'change',
        },
      ],
      capacity: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
        {
          validator: validateCapacity,
          trigger: 'blur',
        },
      ],
    });
    
    const cancelCreatePartition = (formElement) => {
      formElement.resetFields();
      ctx.emit('cancelCreatePartition');
    };
    
    const confirmCreatePartition = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        let params = {
          CapacityBytes: String(partitionForm.value.capacity),
          FileSystem: 'ext4',
          Number: partitionForm.value.number,
          Links: [{
            Device: {
              '@odata.id': props.createPartitionData.currDisk?.driveLetter,
            },
          }],
        };
        try {
          await createPartitions(params);
          ctx.emit('createPartitionSuccessfully');
        } catch (err) {
          handleOperationResponseError(err);
        } finally {
          cancelCreatePartition(formElement);
        }
      });
    };
    
    return {
      props,
      fileSystemOptions,
      partitionForm,
      partitionFormRef,
      partitionFormRule,
      cancelCreatePartition,
      confirmCreatePartition,
    };
  },
});
</script>
