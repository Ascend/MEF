<template>
  <el-menu
    :default-active='activeIndex'
    mode='horizontal'
    background-color='#1f2329'
    text-color='#FFFFFE'
    active-text-color='#0077FF'
    @select='handleSelect'
    style='height: 48px; padding-left: calc(50vw - 180px); position: fixed; width: 100vw; z-index: 100;'
  >
    <span style='margin-top: 12px; position: fixed; left: 24px; ' class="bold-font">{{ homeTitle }}</span>
    <el-menu-item class='menu-item' :index='navList.home.index'>{{ $t('nav.home') }}</el-menu-item>
    <el-menu-item class='menu-item' :index='navList.manager.index'>{{ $t('nav.manager') }}</el-menu-item>
    <el-menu-item class='menu-item' :index='navList.setting.index'>{{ $t('nav.setting') }}</el-menu-item>
    <span class='nav-block'>
      <app-language-switch />
      <el-dropdown @command='handleClickUserIcon'>
        <img style='margin-right: 0.8vw;' src='@/assets/img/common/user.svg' />
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command='changePassword'>{{ $t('common.changePassword') }}</el-dropdown-item>
            <el-dropdown-item command='logout'>{{ $t('common.logout') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <online-help-icon />
    </span>
  </el-menu>
</template>

<script>
import { useRoute, useRouter } from 'vue-router';
import { ref, watch, defineComponent, onMounted, onUnmounted } from 'vue';

import AppLanguageSwitch from '@/components/AppLanguageSwitch.vue';
import OnlineHelpIcon from '@/components/OnlineHelpIcon.vue';
import { querySessionService } from '@/api/account';
import constants from '@/utils/constants';
import { deleteSession } from '@/api/account';
import { fetchJson } from '@/api/common';
import { getLanguageDefaultChinese, clearSessionStorage } from '@/utils/commonMethods';

export default defineComponent({
  components: {
    AppLanguageSwitch,
    OnlineHelpIcon,
  },
  setup() {
    const router = useRouter();
    const route = useRoute();

    const navList = {
      home: {
        index: 'home',
        href: '/',
      },
      manager: {
        index: 'manager',
        href: '/manager',
      },
      setting: {
        index: 'setting',
        href: '/setting',
      },
    };

    const getActiveIndex = (currPath) => {
      if (currPath.startsWith(navList.manager.href)) {
        return navList.manager.index;
      } else if (currPath.startsWith(navList.setting.href)) {
        return navList.setting.index;
      } else {
        return navList.home.index;
      }
    }

    const activeIndex = ref();

    watch(route, (o) => {
      activeIndex.value = getActiveIndex(o.path);
    });

    const handleSelect = (key, keyPath) => {
      if (route.path.indexOf(key) === -1) {
        let index = keyPath[0];
        router.push(navList[index].href);
      }
    }

    const handleClickUserIcon = async (command) => {
      if (command === 'changePassword') {
        await router.push('/changePassword');
      } else if (command === 'logout') {
        try {
          await deleteSession(sessionStorage.getItem('sessionId'));
        } finally {
          clearSessionStorage();
          await router.push('/login');
        }
      }
    }

    let configJson;
    let homeTitle = ref('');

    const init = async () => {
      configJson = await fetchJson('/WhiteboxConfig/json/config.json');
      homeTitle.value = configJson.websiteTitle[getLanguageDefaultChinese()];
    }

    onMounted(async () => {
      await init();
      startSessionTimer();
    })

    let autoSessionTimer;
    const startSessionTimer = () => {
      autoSessionTimer = setInterval(async () => {
        await querySessionService(false, true)
      }
      , constants.SESSION_INTERVAL)
    }

    const stopSessionTimer = () => {
      clearInterval(autoSessionTimer);
      autoSessionTimer = null;
    }

    onUnmounted(() => {
      stopSessionTimer();
    })

    return {
      navList,
      activeIndex,
      homeTitle,
      handleSelect,
      handleClickUserIcon,
    }
  },
})
</script>

<style lang='scss'>
.menu-item {
  width: 120px;
}

img {
  height: 22px;
  width: 22px;
}

.nav-block {
  height: 48px;
  position: fixed;
  right: 1.3vw;
  display: flex;
  align-items: center;
  cursor: pointer;
}
</style>