<template>
  <el-tabs v-model="activeName">
    <el-tab-pane :label="$t('disk.localDiskManagement.tabName')" name="localDisk">
      <local-disk />
    </el-tab-pane>
    <el-tab-pane lazy :label="$t('disk.nfsManagement.tabName')" name="nfs">
      <nfs />
    </el-tab-pane>
  </el-tabs>
</template>

<script>
import { defineComponent, ref } from 'vue';
import LocalDisk from '@/views/manager/disk/LocalDisk.vue';
import Nfs from '@/views/manager/disk/Nfs.vue';

export default defineComponent({
  name: 'ManagerDisk',
  components: {
    LocalDisk,
    Nfs,
  },
  setup() {
    const activeName = ref('localDisk')

    return {
      activeName,
    }
  },
});
</script>

<style lang="scss" scoped>
</style>