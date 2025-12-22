<template>
  <el-aside width="12vw" style="position: fixed;">
    <app-aside :route-name="routeName"></app-aside>
  </el-aside>
  <el-main style="margin-left: 12vw">
    <router-view></router-view>
  </el-main>
</template>

<script setup>

import AppAside from '@/components/AppAside.vue';

const routeName = 'manager';

</script>
<style scoped>

</style>