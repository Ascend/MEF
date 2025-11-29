<template>
  <el-button @click="clickOnlineHelp" style="border: none; padding: 0;">
    <img src='@/assets/img/common/help.svg'>
  </el-button>
</template>

<script>
import { defineComponent } from 'vue';

import { getLanguageDefaultChinese, isA500 } from '@/utils/commonMethods';

export default defineComponent({
  name: 'OnlineHelpIcon',
  setup() {
    const pathOnlineHelpMapper = {
      home: 'help_006.html',
      network: 'help_009.html',
      time: 'help_012.html',
      registration: isA500() ? 'help_014.html' : 'help_013.html',
      disk: isA500() ? 'help_015.html' : 'help_014.html',
      alarm: isA500() ? 'help_018.html' : 'help_017.html',
      journal: isA500() ? 'help_022.html' : 'help_021.html',
      update: isA500() ? 'help_023.html' : 'help_022.html',
      reload: isA500() ? 'help_026.html' : 'help_025.html',
      information: isA500() ? 'help_027.html' : 'help_026.html',
      safety: isA500() ? 'help_029.html' : 'help_028.html',
      reset: 'help_036.html',
      restore: 'help_037.html',
    }

    const openUrl = (url) => {
      let newTab = window.open();
      newTab.opener = null;
      newTab.location = url;
    }

    const isEnglish = () => Boolean(getLanguageDefaultChinese().indexOf('zh') === -1);

    const clickOnlineHelp = () => {
      let sutNum = window.location.href.lastIndexOf('/') + 1;
      let currPage = window.location.href.substring(sutNum) || 'home';
      let lang = getLanguageDefaultChinese();
      if (Object.prototype.hasOwnProperty.call(pathOnlineHelpMapper, currPage)) {
        openUrl(`/onlineHelp/${lang}/${pathOnlineHelpMapper[currPage]}`);
      }
    }

    return {
      clickOnlineHelp,
      isEnglish,
    }
  },
})
</script>

<style scoped>

</style>