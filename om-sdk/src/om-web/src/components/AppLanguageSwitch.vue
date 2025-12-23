<template>
  <span style='display:flex; align-items: center; cursor: pointer;'>
    <img style='margin-right: 0.2vw;' @click='switchLanguage' src='@/assets/img/common/language.svg' />
    <span class="regular-font" @click='switchLanguage' style='margin-right: 0.8vw;'>{{ currLanguage }}</span>
  </span>
</template>

<script>
import { ref, defineComponent, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';

import { getLanguageDefaultChinese } from '@/utils/commonMethods';

export default defineComponent({
  name: 'AppLanguageSwitch',
  setup() {
    const currLanguage = ref('')
    const { locale } = useI18n();

    const switchLanguage = () => {
      sessionStorage.setItem('locale', {
        '简体中文': 'en',
        'English': 'zh',
      }[currLanguage.value])
      updateSetting();
      window.location.reload();
    }

    const updateSetting = () => {
      currLanguage.value = {
        zh: '简体中文',
        en: 'English',
      }[getLanguageDefaultChinese()]
      locale.value = getLanguageDefaultChinese()
    }

    onMounted(() => {
      updateSetting();
    })

    return {
      currLanguage,
      switchLanguage,
    }
  },
})
</script>

<style scoped>

</style>