<template>
  <el-tabs v-model='activeName'>
    <el-tab-pane :label='$t("update.systemUpgrade.tabName")' name='system'>
      <system-update />
    </el-tab-pane>
  </el-tabs>
</template>

<script>
import { defineComponent, ref } from 'vue';

import SystemUpdate from '@/views/manager/update/SystemUpdate.vue';

export default defineComponent({
  name: 'MaintenanceUpdate',
  components: {
    SystemUpdate,
  },
  setup() {
    const activeName = ref('system')

    return {
      activeName,
    }
  },
});
</script>

<style scoped>

</style>