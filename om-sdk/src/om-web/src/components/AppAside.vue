<template>
  <template v-for="nav in navList" v-bind:key="nav.name" >
    <div class="nav-item" @click="handleSelect(nav)" :class="{ 'active': nav.path === activeIndex }">
      <img :src="nav.img" alt="" class="nav-icon"/>
      <span class="nav-text">{{ $t(nav.name) }}</span>
    </div>
  </template>
</template>

<script>
import { useRoute, useRouter } from 'vue-router';
import { ref, watch, defineComponent } from 'vue';

export default defineComponent({
  name: 'AppAside',
  props: {
    routeName: String,
  },
  setup(props) {
    const router = useRouter();
    const route = useRoute();
    const asideRouter = router.options.routes.find(el => el.name === props.routeName);
    const navList = asideRouter.children.filter(el => el.name);
    const activeIndex = ref(route.name.split('.')[0]);
    watch(route, () => {
      activeIndex.value = route.name.split('.')[0]
    });

    const handleSelect = (item) => {
      if (activeIndex.value !== item.path) {
        router.push({ path: item.path })
      }
    }

    return {
      navList,
      activeIndex,
      handleSelect,
    }
  },
})

</script>

<style scoped>
.nav-item {
  line-height: 16px;
  margin-left: 24px;
  border-radius: 4px;
  background: var(--el-bg-color-overlay);
  margin-bottom: 8px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: 16px;
}

.nav-text {
  margin-left: 10px;
  font-size: 14px;
  line-height: 16px;
  font-weight: 400;
}

.active {
  color: var(--el-text-customize-color-primary);
}

.nav-icon {
  height: 16px;
  width: 16px;
}
</style>