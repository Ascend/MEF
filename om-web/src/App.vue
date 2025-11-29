<template>
  <el-config-provider :locale="locale">
    <el-header v-if="isShowHeader(route.name)" style="padding: 0;">
      <app-header></app-header>
    </el-header>
    <el-container>
      <router-view></router-view>
    </el-container>
    <el-footer class="footer"></el-footer>
  </el-config-provider>
</template>

<script>
import { defineComponent, onMounted } from 'vue';
import { useRoute } from 'vue-router';

import zh from 'element-plus/dist/locale/zh-cn.mjs';
import en from 'element-plus/dist/locale/en.mjs';
import { getLanguageDefaultChinese } from '@/utils/commonMethods';
import AppHeader from '@/components/AppHeader.vue';
import { fetchJson } from '@/api/common';

export default defineComponent({
  components: {
    AppHeader,
  },
  setup() {
    const locale = {
      zh,
      en,
    }[getLanguageDefaultChinese()];
    const route = useRoute();
    let configJson;

    const setWebsiteTitle = () => {
      document.title = configJson.websiteTitle[getLanguageDefaultChinese()];
    }

    const isShowHeader = (path) => path !== 'login' && path !== 'changePassword'

    onMounted(async () => {
      configJson = await fetchJson('/WhiteboxConfig/json/config.json');
      setWebsiteTitle();
    })

    return {
      locale,
      route,
      isShowHeader,
    }
  },
})
</script>

<style lang='scss' scoped>
.footer {
  position: fixed;
  z-index: 100;
  bottom: 0;
  width: 100vw;
  height: 40px;
  margin-top: 20px;
}
</style>