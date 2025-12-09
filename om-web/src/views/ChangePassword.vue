<template>
  <div class="main">
    <div class="top-info">
      <img src="/WhiteboxConfig/img/logo.png" alt=""/>
      <span class="bold-font">{{ systemName }}</span>
    </div>
    <span class='nav-block'>
      <app-language-switch />
    </span>
    <div class="form">
      <div class="form-title bold-font">{{ $t('common.changePwd') }}</div>
      <app-tip :tip-list="formTipLists" :tip-theme="'light'" />
      <el-form
        :model='changeUserInfoForm'
        :rules='changeUserInfoFormRule'
        :hide-required-asterisk="true"
        ref='changeUserInfoFormRef'
        style='margin: 20px auto;'
      >
        <el-form-item prop='old_password'>
          <el-input
            class='login-inp'
            v-model='changeUserInfoForm.old_password'
            :placeholder='$t("userInfo.oldPwdTip")'
            type='password'
            style="height: 48px;"
            autocomplete="off"
            @copy.native.capture.prevent="()=>{}"
            @paste.native.capture.prevent="()=>{}"
            @cut.native.capture.prevent="()=>{}"
          />
        </el-form-item>
        <el-form-item prop='Password'>
          <el-input
            class='login-inp'
            v-model='changeUserInfoForm.Password'
            :placeholder='$t("userInfo.newPwdTip")'
            type='password'
            style="height: 48px;"
            autocomplete="off"
            @copy.native.capture.prevent="()=>{}"
            @paste.native.capture.prevent="()=>{}"
            @cut.native.capture.prevent="()=>{}"
          />
        </el-form-item>
        <div style="margin-top: 30px; margin-bottom: 30px; width: 100%; display: flex; align-items: center;" >
          <div
            class="password-strength"
            :class="{
              'low': isLowPasswordStrengthLevel(),
              'medium': isMediumPasswordStrengthLevel(),
              'high': isHighPasswordStrengthLevel(),
            }" />
          <div
            class="password-strength"
            :class="{
              'medium': isMediumPasswordStrengthLevel(),
              'high': isHighPasswordStrengthLevel(),
            }"
          />
          <div class="password-strength" :class="{ 'high': isHighPasswordStrengthLevel() }" />
          <div class="password-strength-text">{{ strengthText }}</div>
        </div>
        <el-form-item prop='new_password_second' style="clear:both;">
          <el-input
            class='login-inp'
            v-model='changeUserInfoForm.new_password_second'
            :placeholder='$t("userInfo.confirmNewPwdTip")'
            type='password'
            style="height: 48px;"
            @keyup.enter='submitChange(changeUserInfoFormRef)'
            autocomplete="off"
            @copy.native.capture.prevent="()=>{}"
            @paste.native.capture.prevent="()=>{}"
            @cut.native.capture.prevent="()=>{}"
          />
        </el-form-item>
      </el-form>
      <el-button
        style="background-color: var(--el-customize-deep-blue-btn-bg-color);width: 100%; margin-top: 10px; height: 48px;"
        class="medium-font"
        @click='submitChange(changeUserInfoFormRef)'
      >{{ $t('common.confirm') }}</el-button>
      <el-button
        style="width: 100%; margin-left: 0; margin-top: 10px; height: 48px;"
        class="button-cancel"
        @click="cancelChange"
      >{{ $t("common.cancel") }}</el-button>
    </div>
  </div>
</template>

<script>
import { defineComponent, reactive, ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router';

import { modifyUserInfo } from '@/api/account';
import AppLanguageSwitch from '@/components/AppLanguageSwitch.vue';
import AppTip from '@/components/AppTip.vue';
import {
  getLanguageDefaultChinese, AppMessage,
  handleOperationResponseError, clearSessionStorage
} from '@/utils/commonMethods';
import { fetchJson } from '@/api/common';

export default defineComponent({
  name: 'ChangePassword',
  components: {
    AppLanguageSwitch,
    AppTip,
  },
  setup() {
    const router = useRouter();
    const { t } = useI18n();
    const strengthText = ref('');
    const changeUserInfoFormRef = ref();
    const passwordStrengthLevel = ref();
    const changeUserInfoForm = ref({
      UserName: 'admin',
      old_password: null,
      Password: null,
      new_password_second: null,
    })
    let configJson;
    const systemName = ref();
    const copyRight = ref();

    let formTipLists = [
      t('common.changePwdTip'),
      t('common.savePwdTip'),
    ]

    const validatePassword = (rule, value, callback) => {
      if (value === '') {
        callback(new Error(t('userInfo.newPwdTip')))
        return;
      }

      let pattern = /^(?![a-zA-Z]+$)(?![A-Z0-9]+$)(?![A-Z\\W_!@#$%^&*()_+\-=~`{[\]}|\\:;\'"<,>.?/\s]+$)(?![a-z0-9]+$)(?![a-z\\W_!@#$%^&*()_+\-=~`{[\]}|\\:;\'"<,>.?/\s]+$)(?![0-9\\W_!@#$%^&*()_+\-=~`{[\]}|\\:;\'"<,>.?/\s]+$)[a-zA-Z0-9\\W_!@#$%^&*()_+\-=~`{[\]}|\\:;\'"<,>.?/\s]{8,20}$/;
      if (!pattern.exec(value)) {
        callback(new Error(t('userInfo.passwordStrengthTip')));
        return;
      }

      if (changeUserInfoForm.value.new_password_second !== '') {
        if (!changeUserInfoFormRef.value) {
          return;
        }
        changeUserInfoFormRef.value.validateField('new_password_second', () => null)
      }
      callback()
    }

    const validateConfirmedPassword = (rule, value, callback) => {
      if (value === '') {
        callback(new Error(t('userInfo.confirmNewPwdTip')))
      } else if (value !== changeUserInfoForm.value.Password) {
        callback(new Error(t('userInfo.confirmPasswordNotMatchTip')))
      } else {
        callback()
      }
    }

    const changeUserInfoFormRule = reactive({
      UserName: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
      old_password: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
      Password: [
        {
          validator: validatePassword,
          trigger: 'blur',
        },
      ],
      new_password_second: [
        {
          validator: validateConfirmedPassword,
          trigger: 'blur',
        },
      ],
    })

    const submitChange = async (formElement) => {
      if (!formElement) {
        return;
      }
      await formElement.validate(async (valid) => {
        if (!valid) {
          return;
        }
        try {
          await modifyUserInfo(sessionStorage.getItem('userId'), changeUserInfoForm.value);
          AppMessage.success(t('common.modifySuccessfully'))
          clearSessionStorage();
          await router.push('/login');
        } catch (err) {
          handleOperationResponseError(err, t('session.loginErrorTip'))
        } finally {
          formElement.resetFields();
        }
      })
    }

    const passwordStrength = (pwd) => {
      try {
        let reg1 = /[A-Z]/g;
        let reg2 = /[a-z]/g;
        let reg3 = /[0-9]/g;
        let reg4 = /[\ \`\~\!\@\#\$\%\^\&\*\(\)\-\_\=\+\\\|\[\{\}\]\;\:\'\"\,\<\.\>\/\?]+/g;
        let len = pwd.length >= 8 ? pwd.length >= 12 ? 20 : 10 : 0;
        let upper = reg1.test(pwd) ? pwd.match(reg1).length > 1 ? 20 : 10 : 0;
        let down = reg2.test(pwd) ? pwd.match(reg2).length > 1 ? 20 : 10 : 0;
        let number = reg3.test(pwd) ? pwd.match(reg3).length > 1 ? 20 : 10 : 0;
        let spec = reg4.test(pwd) ? pwd.match(reg4).length > 1 ? 20 : 10 : 0;
        return upper + down + number + spec + len;
      } catch (e) {
        return 0;
      }
    }

    watch(
      () => changeUserInfoForm.value.Password,
      () => {
        if (changeUserInfoForm.value.Password === '') {
          strengthText.value = '';
          passwordStrengthLevel.value = '';
        } else {
          let level = passwordStrength(changeUserInfoForm.value.Password);
          if (level >= 100) {
            strengthText.value = t('common.highSecurity');
            passwordStrengthLevel.value = 'high';
          } else if (level >= 60) {
            strengthText.value = t('common.mediumSecurity');
            passwordStrengthLevel.value = 'medium';
          } else {
            strengthText.value = t('common.lowSecurity');
            passwordStrengthLevel.value = 'low';
          }
        }
      }
    )

    const isLowPasswordStrengthLevel = () => passwordStrengthLevel.value === 'low';
    const isMediumPasswordStrengthLevel = () => passwordStrengthLevel.value === 'medium';
    const isHighPasswordStrengthLevel = () => passwordStrengthLevel.value === 'high';
    const init = () => {
      systemName.value = configJson.systemName[getLanguageDefaultChinese()];
      copyRight.value = configJson.copyRight[getLanguageDefaultChinese()];
    }

    onMounted(async () => {
      configJson = await fetchJson('/WhiteboxConfig/json/config.json');
      init();
    })
    const cancelChange = () => {
      history.go(-1)
    }

    return {
      systemName,
      copyRight,
      passwordStrengthLevel,
      formTipLists,
      strengthText,
      changeUserInfoForm,
      changeUserInfoFormRef,
      changeUserInfoFormRule,
      isLowPasswordStrengthLevel,
      isMediumPasswordStrengthLevel,
      isHighPasswordStrengthLevel,
      submitChange,
      cancelChange,
    }
  },
})
</script>

<style lang="scss" scoped>
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

.form {
  box-shadow: 0 4px 16px 8px rgba(0,0,0,0.10);
  border-radius: 4px;
  width: 20vw;
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

.form-title {
  color: var(--el-bg-color);
  font-size: 40px;
  line-height: 48px;
  font-weight: 700;
  margin-bottom: 10px;
}

.password-strength {
  float: left;
  height: 10px;
  width: calc(23% - 10px);
  margin-right: 10px;
  background: var(--el-customize-pwd-bg-color);
  border-radius: 4px;
}

.password-strength.low {
  background: var(--el-color-danger);
}

.password-strength.medium {
  background: var(--el-customize-pwd-medium-bg-color);
}

.password-strength.high {
  background: var(--el-customize-pwd-high-bg-color);
}

.password-strength-text {
  float: right;
  right: 0;
  color: var(--el-text-color-secondary);
  font-size: 14px;
  letter-spacing: 0;
  line-height: 20px;
  font-weight: 400;
}

.button-cancel {
  font-family: 'HarmonyOS_Sans_SC_Medium', serif;
  background-color: var(--el-text-color-regular);
  color: var(--el-customize-deep-blue-btn-bg-color);
}

.button-cancel:hover {
  color: var(--el-color-primary);
  border-color: var(--el-color-primary);
}
</style>