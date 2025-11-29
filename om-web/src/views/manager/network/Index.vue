<template>
  <el-tabs v-model="activeName">
    <el-tab-pane :label="$t('network.networkWired.tabName')" name="wired">
      <wired />
    </el-tab-pane>
    <el-tab-pane v-if="hasWireless" :label="$t('network.networkWireless.tabName')" name="wireless">
      <wireless />
    </el-tab-pane>
  </el-tabs>
</template>

<script>
import { ref, defineComponent, onMounted } from 'vue';
import { useI18n } from 'vue-i18n'

import Wired from '@/views/manager/network/Wired.vue';
import Wireless from '@/views/manager/network/Wireless.vue';
import { queryWirelessStatusInfo } from '@/api/network';
import { getLanguageDefaultChinese, showErrorAlert } from '@/utils/commonMethods';

export default defineComponent({
  name: 'ManagerNetwork',
  components: {
    Wired,
    Wireless,
  },
  setup() {
    const activeName = ref('wired')
    const hasWireless = ref(false);
    const { t } = useI18n();
    const currentLocale = getLanguageDefaultChinese();

    const fetchStatusInfo = async () => {
      let { data } = await queryWirelessStatusInfo(false);
      return data;
    }

    const init = async () => {
      try {
        let statusInfo = await fetchStatusInfo();
        hasWireless.value = currentLocale === 'zh' ? statusInfo?.lte_enable : false;
      } catch (e) {
        showErrorAlert(t('common.requestError'))
      }
    }
    onMounted(async () => {
      await init();
    })
    return {
      activeName,
      hasWireless,
    }
  },
});
</script>

<style lang="scss" scoped>

</style>