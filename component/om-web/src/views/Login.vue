<template>
  <div class="main">
    <div class="top-info">
      <img src="/WhiteboxConfig/img/logo.png" alt=""/>
      <span class="bold-font">{{ systemName }}</span>
    </div>
    <span class='nav-block'>
      <app-language-switch />
    </span>
    <div class="login-form">
      <div class="welcome">{{ $t('common.welcome') }}</div>
      <div class="system-name bold-font">{{ systemName }}</div>
      <el-form
        :model="loginForm"
        ref="loginFormRef"
        style="margin-top: 6vh; margin-bottom: 10vh;"
      >
        <el-form-item prop="UserName">
          <el-input
            v-model="loginForm.UserName"
            class="login-inp"
            :class="{errorInput: !checkUsername.isPass}"
            :placeholder="$t('common.usernamePlaceholder')"
            style="height: 48px;"
            autocomplete="off"
          />
          <div class="error-text" v-if="!checkUsername.isPass">{{ checkUsername.errorText }}</div>
        </el-form-item>
        <el-form-item prop="Password">
          <el-input
            class="login-inp"
            :class="{errorInput: !checkPassword.isPass}"
            v-model="loginForm.Password"
            :placeholder="$t('common.passwordPlaceholder')"
            style="height: 48px; margin-top: 10px;"
            type="password"
            @keyup.enter="submitLogin(loginFormRef)"
            autocomplete="off"
            @copy.native.capture.prevent="()=>{}"
            @paste.native.capture.prevent="()=>{}"
            @cut.native.capture.prevent="()=>{}"
          />
          <div class="error-text" v-if="!checkPassword.isPass">{{ checkPassword.errorText }}</div>
        </el-form-item>
      </el-form>
      <el-button style="width: 100%; height: 48px;" class="medium-font" type="primary" @click="submitLogin(loginFormRef)">{{ $t('common.login') }}</el-button>
    </div>
    <div class="copy-right">{{ copyRight }}</div>
  </div>
</template>

<script>
import { defineComponent, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router';
import { ElNotification } from 'element-plus'

import { createSession, getUserInfo } from '@/api/account';
import { querySystemsSourceInfo } from '@/api/information';
import errorCode from '@/api/errorCode';
import { getLanguage, getLanguageDefaultChinese, AppMessage, generateA200Routes } from '@/utils/commonMethods';
import AppLanguageSwitch from '@/components/AppLanguageSwitch.vue';
import { fetchJson } from '@/api/common';
import errorTipMapper from '@/api/errorTipMapper';

export default defineComponent({
  name: 'LoginView',
  components: {
    AppLanguageSwitch,
  },
  setup() {
    const loginForm = ref({
      UserName: null,
      Password: null,
    })
    let configJson;
    const systemName = ref();
    const copyRight = ref();

    const checkPassword = ref({
      isPass: true,
      errorText: '',
    });
    const checkUsername = ref({
      isPass: true,
      errorText: '',
    });

    const router = useRouter();
    const { t } = useI18n();
    const loginFormRef = ref();

    const getLoginTime = async () => {
      let { data } = await getUserInfo(sessionStorage.getItem('userId'));
      if (!data?.Oem) {
        return;
      }
      if (data?.Oem.LastLoginSuccessTime < data?.Oem.LastLoginFailureTime) {
        let text = t('session.lastLoginFailed', {
          date: data?.Oem.LastLoginFailureTime,
          ipAddress: data?.Oem.LastLoginIP,
        })

        if (data?.Oem.PwordWrongTimes > 0) {
          text += t('session.wrongTimes', { times: data?.Oem.PwordWrongTimes })
        }

        ElNotification({
          title: t('session.loginInfo'),
          message: text,
          type: 'warning',
        })
      } else {
        let text = t('session.lastLoginSuccess',
          {
            date: data?.Oem.LastLoginSuccessTime,
            ipAddress: data?.Oem.LastLoginIP,
          }
        )
        ElNotification({
          title: t('session.loginInfo'),
          description: text,
          message: text,
          type: 'success',
        })
      }
    }

    const submitLogin = async (formEl) => {
      if (!formEl) {
        return
      }
      if (!loginForm.value.UserName) {
        checkUsername.value.isPass = false
        checkUsername.value.errorText = t('common.notEmpty')
      } else {
        checkUsername.value.isPass = true
        checkUsername.value.errorText = ''
      }

      if (!loginForm.value.Password) {
        checkPassword.value.isPass = false
        checkPassword.value.errorText = t('common.notEmpty')
      } else {
        checkPassword.value.isPass = true
        checkPassword.value.errorText = ''
      }
      if (!loginForm.value.UserName || !loginForm.value.Password) {
        return;
      }

      try {
        let { headers, data } = await createSession(loginForm.value);
        if (headers['x-auth-token']) {
          let token = headers['x-auth-token'];
          sessionStorage.setItem('token', token)
          sessionStorage.setItem('userId', data?.Oem?.UserId)
          sessionStorage.setItem('sessionId', data?.Id)
          if (!getLanguage()) {
            sessionStorage.setItem('locale', 'zh')
          }
          if (data?.Oem?.AccountInsecurePrompt) {
            AppMessage.warning(t('common.needModifyPassword'))
            await router.push('/changePassword');
            return;
          }
          await getModel();
          await generateA200Routes(router);
          await router.push('/');
          await getLoginTime();
        }
      } catch (err) {
        if (!err?.response?.data?.error) {
          return;
        }
        let status = err.response.data.error['@Message.ExtendedInfo'][0]?.Oem?.status;
        if (status === errorCode.SESSION.ERROR_CHANGE_PASSWORD) {
          AppMessage.warning(t('common.needModifyPassword'))
          await router.push('/changePassword');
        } else {
          checkPassword.value.isPass = false;
          if (status === errorCode.SESSION.ERROR_USER_LOCK_STATE) {
            let msg = err.response.data.error['@Message.ExtendedInfo'][0]?.Message;
            checkPassword.value.errorText = t('session.lockTip', { seconds: msg.split('[')[1].split(']')[0] })
          } else {
            checkPassword.value.errorText = t(errorTipMapper[status]) ?? t('session.loginErrorTip');
          }
        }
      } finally {
        loginForm.value = {
          UserName: null,
          Password: null,
        }
      }
    }
    const getModel = async () => {
      let { data } = await querySystemsSourceInfo();
      sessionStorage.setItem('model', data?.SupportModel);
    }

    const init = () => {
      systemName.value = configJson.systemName[getLanguageDefaultChinese()];
      copyRight.value = configJson.copyRight[getLanguageDefaultChinese()];
    }

    onMounted(async () => {
      configJson = await fetchJson('/WhiteboxConfig/json/config.json');
      sessionStorage.removeItem('token');
      init();
    })

    return {
      systemName,
      copyRight,
      loginForm,
      loginFormRef,
      checkPassword,
      checkUsername,
      submitLogin,
    }
  },
})
</script>

<style lang="scss" scoped>
.login-inp {
  width: 100%;
}

.main {
  width: 100vw;
  height: 100vh;
  background: url("/WhiteboxConfig/img/login.png") no-repeat ;
  background-size:100% 100%;
}

.top-info {
  display: flex;
  align-items: center;
  padding: 24px;
  font-size: 24px;
  color: var(--el-text-customize-black-color);
  letter-spacing: 0;
  line-height: 32px;
  font-weight: 600;
  img {
    margin-right: 16px;
    height: 32px;
    width: 32px;
  }
}

.nav-block {
  height: 48px;
  position: fixed;
  right: 24px;
  top: 0;
  display: flex;
  align-items: center;
  cursor: pointer;
  color: var(--el-bg-color);
  font-size: 14px;
  letter-spacing: 0;
  text-align: center;
  line-height: 20px;
  font-weight: 400;
}

.welcome {
  font-size: 12px;
  margin-top: 16px;
  margin-bottom: 8px;
  color: var(--el-text-customize-gray-color);
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
}

.login-form-title {
  font-size: 40px;
  color: var(--el-bg-color);
  line-height: 48px;
  font-weight: 400;
}

.login-form {
  box-shadow: 0 4px 16px 8px rgba(0,0,0,0.10);
  border-radius: 4px;
  width: 18vw;
  color: var(--el-bg-color);
  background: var(--el-text-color-regular);
  padding: 48px;
  position: absolute;
  right: 140px;
  top: 180px;
  :deep(.el-input__wrapper) {
    border: 1px solid rgba(0,0,0,0.10) ;
    box-shadow: 0 0 0 1px rgba(0,0,0,0.10) ;
  }
  :deep(.el-input__inner){
    color: var(--el-text-customize-black-color);
  }
}

.system-name {
  font-size: 40px;
  color: var(--el-bg-color);
  line-height: 48px;
  font-weight: 700;
}

.copy-right {
  font-size: 12px;
  color: var(--el-text-customize-gray-color);
  letter-spacing: 0;
  text-align: center;
  line-height: 16px;
  font-weight: 400;
  width: 100vw;
  position: fixed;
  bottom: 20px;
}

.errorInput {
  :deep(.el-input__wrapper) {
    box-shadow: 0 0 0 1px var(--el-color-danger) inset;
  }
}
</style>