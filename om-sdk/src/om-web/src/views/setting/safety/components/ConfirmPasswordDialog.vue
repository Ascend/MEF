<template>
  <el-dialog
    :title='title'
    :modelValue='isShow'
    :width='640'
    @close="onCancel"
    align-center
  >
    <app-tip :tip-text='tip'/>
    <br>
    <el-form
      ref='formRef'
      label-position='left'
      label-width="auto"
      :model='form'
      :rules='formRules'
      :hide-required-asterisk="true"
      @submit.prevent
    >
      <el-form-item :label='$t("safety.loginRules.password")'
                    class='required' prop='password'>
        <el-input v-model='form.password'
                  type="password"
                  @keyup.enter='onConfirm(formRef)'
                  autocomplete="off"
                  @copy.native.capture.prevent="()=>{}"
                  @paste.native.capture.prevent="()=>{}"
                  @cut.native.capture.prevent="()=>{}"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="onConfirm(formRef)">
          {{ $t("common.confirm") }}
        </el-button>
        <el-button @click="onCancel(formRef)">
          {{ $t("common.cancel") }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import AppTip from '@/components/AppTip.vue';
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

export default {
  name: 'ConfirmPasswordDialog',
  components: { AppTip },
  props: {
    isShow: Boolean,
    title: String,
    tip: String,
  },
  setup(props, ctx) {
    const events = {
      submit: 'submitConfirm',
      cancel: 'cancel',
    };

    const { t } = useI18n();
    const formRef = ref();
    const form = reactive({
      password: null,
    });
    const formRules = ref({
      password: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
    });

    const onConfirm = async (formRef_) => {
      if (!formRef_) {
        return
      }
      await formRef_.validate(async (valid) => {
        if (!valid) {
          return
        }
        ctx.emit(events.submit, form.password);
      });
    }

    const onCancel = async () => {
      ctx.emit(events.cancel);
    }

    return {
      form,
      formRules,
      formRef,
      onConfirm,
      onCancel,
    };
  },
};
</script>

<style scoped>

</style>