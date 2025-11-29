<template>
  <div class='container'>
    <div class='basic-info'>{{ $t('common.basicInfo') }}</div>
    <app-tip :tip-type='"info"' :tip-text='$t("time.time.saveTip")' />
    <slot></slot>
  </div>
</template>

<script>
import { defineComponent } from 'vue';

import AppTip from '@/components/AppTip.vue';

export default defineComponent({
  name: 'BasicInformationContainer',
  components: {
    AppTip,
  },
})
</script>

<style scoped>
.container {
  background: var(--el-bg-color-overlay);
  border-radius: 4px;
  padding: 24px 48px;
}

.basic-info {
  font-size: 16px;
  letter-spacing: 0;
  line-height: 24px;
  font-weight: 400;
  margin-bottom: 8px;
}
</style>