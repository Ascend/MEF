<template>
  <el-tabs v-model="activeName">
    <el-tab-pane lazy :label="$t('safety.webTimeout.tabName')" name="webTimeout">
      <web-timeout />
    </el-tab-pane>
    <el-tab-pane lazy :label="$t('safety.passwordValidity.tabName')" name="passwordValidity">
      <password-validity />
    </el-tab-pane>
    <el-tab-pane lazy :label="$t('safety.loginRules.tabName')" name="loginRules">
      <login-rule />
    </el-tab-pane>
    <el-tab-pane lazy :label="$t('safety.weakPassword.tabName')" name="weakPassword">
      <weak-password />
    </el-tab-pane>
    <el-tab-pane lazy :label="$t('safety.cert.tabName')" name="certInfo">
      <cert-info />
    </el-tab-pane>
    <el-tab-pane lazy :label="$t('safety.certificateExpiryAlarm.tabName')" name="certificateExpiryAlarm">
      <cert-expiry-alarm />
    </el-tab-pane>
  </el-tabs>
</template>

<script>
import { defineComponent, ref } from 'vue';
import WebTimeout from '@/views/setting/safety/WebTimeout.vue';
import PasswordValidity from '@/views/setting/safety/PasswordValidity.vue';
import LoginRule from '@/views/setting/safety/LoginRule.vue';
import WeakPassword from '@/views/setting/safety/WeakPassword.vue';
import CertInfo from '@/views/setting/safety/CertInfo.vue';
import CertExpiryAlarm from '@/views/setting/safety/CertExpiryAlarm.vue';

export default defineComponent({
  name: 'SettingSafety',
  components: {
    WebTimeout,
    PasswordValidity,
    LoginRule,
    WeakPassword,
    CertInfo,
    CertExpiryAlarm,
  },
  setup() {
    const activeName = ref('webTimeout')
    return {
      activeName,
    }
  },
});
</script>

<style lang="scss" scoped>
</style>