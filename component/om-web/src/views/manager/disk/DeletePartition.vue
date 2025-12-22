<template>
  <el-dialog
    :title="$t('common.deleteTip')"
    :modelValue="props.deletePartitionData.isShow"
    style="width: 500px;"
    @close="cancelDeletePartition"
  >
    <el-row style="margin-bottom: 40px; word-break: break-word;">
      <el-col :span="2">
        <img class="alarm-img" alt="" src="@/assets/img/alarm.svg" />
      </el-col>
      <el-col :span="20">
        {{ $t('disk.localDiskManagement.deletePartitionTip.tip', { 'partition': props.deletePartitionData?.partitionId }) }}
      </el-col>
    </el-row>
    <div style="margin-left: calc(50% - 86px);">
      <el-button type="primary" @click="confirmDeletePartition">
        {{ $t('common.confirm') }}
      </el-button>
      <el-button @click="cancelDeletePartition">
        {{ $t('common.cancel') }}
      </el-button>
    </div>
  </el-dialog>
</template>

<script>
import { defineComponent } from 'vue';
import { deletePartition } from '@/api/drive';
import { handleOperationResponseError } from '@/utils/commonMethods';

export default defineComponent({
  name: 'DeletePartition',
  props: {
    deletePartitionData: Object,
  },
  setup(props, ctx) {
    const confirmDeletePartition = async () => {
      try {
        await deletePartition(props.deletePartitionData.partitionId);
        ctx.emit('deletePartitionSuccessfully');
      } catch (err) {
        handleOperationResponseError(err);
      } finally {
        cancelDeletePartition();
      }
    };
    
    const cancelDeletePartition = () => {
      ctx.emit('cancelDeletePartition');
    };
    
    return {
      props,
      confirmDeletePartition,
      cancelDeletePartition,
    };
  },
});
</script>

<style scoped>
.alarm-img {
  width: 24px;
  height: 24px;
}
</style>