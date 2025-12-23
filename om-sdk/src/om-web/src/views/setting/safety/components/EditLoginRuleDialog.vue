<template>
  <el-dialog
    :title='title'
    :modelValue="isShow"
    :width="640"
    style="height: 670px;"
    @close="onDialogCancel"
    :close-on-click-modal="false"
  >
    <app-tip :tip-text='$t("safety.loginRules.createTip")' />
    <el-form
      ref='ruleFormRef'
      label-position='left'
      label-width="auto"
      :model='ruleForm'
      :rules='ruleFormRules'
      class='add-ip-form'
      :hide-required-asterisk="true"
      style="word-break: break-word;"
    >
      <el-form-item :label='$t("safety.loginRules.time")' id='input-time' class='not-required time-picker' prop='time'>
        <div style="display: flex;">
          <el-time-picker
            style="width: 100%"
            format='HH:mm'
            :default-value="parseTime('00:00')"
            v-model="ruleForm.startTime"
            :placeholder='$t("safety.loginRules.startTimePlaceholder")'
          />
          <span style="margin-left: 6px; margin-right: 6px;">-</span>
          <el-time-picker
            style="width: 100%"
            format='HH:mm'
            :default-value="parseTime('00:00')"
            v-model="ruleForm.endTime"
            :placeholder='$t("safety.loginRules.endTimePlaceholder")'
          />
        </div>
      </el-form-item>
      <el-form-item :label='$t("safety.loginRules.ipSegmentFormat")' id='input-ip-with-seg' class='not-required' prop='ipSegmentFormat'>
        <span>{{ $t("safety.loginRules.ipTip") }}</span>
      </el-form-item>
      <el-form-item :label='$t("safety.loginRules.networkSegment")' id='input-ip' class='not-required' prop='ipAddress'>
        <el-input v-model='ruleForm.ipAddress'/>
      </el-form-item>
      <el-form-item :label='$t("safety.loginRules.macFormat")' id='input-mac1' class='not-required' prop='macFormat'>
        <span>{{ $t("safety.loginRules.macTip") }}</span>
      </el-form-item>
      <el-form-item :label='$t("safety.loginRules.macAddress")' id='input-mac2' class='not-required' prop='macAddress'>
        <el-input v-if="isWholeFormatMacAddress()" v-model='ruleForm.macAddress'/>
        <el-input v-else v-model='ruleForm.macAddress'/>
      </el-form-item>
      <el-form-item :label='$t("safety.loginRules.status")' id='input-status' class='not-required' prop='status'>
        <el-switch v-model='ruleForm.status' :active-value="'true'" :inactive-value="'false'"/>
      </el-form-item>
      <el-form-item :label='$t("safety.loginRules.password")' id='input-password' class='required' prop='password'>
        <el-input v-model='ruleForm.password'
                  type="password"
                  @keyup.enter='onDialogConfirm(ruleFormRef)'
                  autocomplete="off"
                  @copy.native.capture.prevent="()=>{}"
                  @paste.native.capture.prevent="()=>{}"
                  @cut.native.capture.prevent="()=>{}"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button type="primary" @click="onDialogConfirm(ruleFormRef)">
          {{ $t("common.confirm") }}
        </el-button>
        <el-button @click="onDialogCancel(ruleFormRef)">
          {{ $t("common.cancel") }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import AppTip from '@/components/AppTip.vue';
import AppPopover from '@/components/AppPopover.vue';
import { onMounted, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { modifyLoginRules } from '@/api/safety';
import { AppMessage, deepCopyDictList, handleOperationResponseError } from '@/utils/commonMethods';
import { ipWithMaskChecker, macChecker } from '@/utils/validator';

export default {
  name: 'EditLoginRuleDialog',
  components: {
    AppPopover,
    AppTip,
  },
  props: {
    isShow: Boolean,
    title: String,
    rules: Array,
    editRuleIndex: Number,
  },
  setup(props, ctx) {
    const events = {
      refresh: 'refresh',
      cancel: 'cancel',
    };
    const { t } = useI18n();
    const ruleFormRef = ref();
    const ruleForm = reactive({
      ruleName: '',
      password: null,
      startTime: '',
      endTime: '',
      ipSegmentFormat: 'ip',
      networkSegment: '',
      networkSegmentMaskRange: '',
      macFormat: 'part',
      macAddress: '',
      ipAddress: '',
      status: false,
    });
    const timeChecker = (rule, value, callback) => {
      let hasStart = Boolean(ruleForm.startTime);
      let hasEnd = Boolean(ruleForm.endTime);
      if ((hasStart && hasEnd) || (!hasStart && !hasEnd)) {
        callback();
      } else {
        callback(new Error(t('safety.loginRules.timeWrongTip')));
      }
    };
    const ruleFormRules = ref({
      ruleName: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
      time: [
        {
          validator: timeChecker,
          trigger: 'blur',
        },
      ],
      ipAddress: [
        {
          validator: ipWithMaskChecker,
          trigger: 'blur',
        },
      ],
      macAddress: [
        {
          validator: macChecker,
          trigger: 'blur',
        },
      ],
      password: [
        {
          required: true,
          message: t('common.notEmpty'),
          trigger: 'blur',
        },
      ],
    });

    const isWholeFormatMacAddress = () => ruleForm.macFormat === 'whole'

    const isCreateNew = () => props.editRuleIndex === undefined;

    const onDialogConfirm = async (formRef) => {
      if (!formRef) {
        return
      }
      await formRef.validate(async (valid) => {
        if (!valid) {
          return
        }
        try {
          let loadCfg = deepCopyDictList(props.rules);
          let rule = {
            'enable': String(ruleForm.status),
            'start_time': formatTime(ruleForm.startTime),
            'end_time': formatTime(ruleForm.endTime),
            'ip_addr': ruleForm.ipAddress || null,
            'mac_addr': ruleForm.macAddress || null,
          }
          if (isCreateNew()) {
            loadCfg.splice(0, 0, rule);
          } else {
            loadCfg[props.editRuleIndex] = rule;
          }
          await modifyLoginRules({
            'Password': ruleForm.password,
            'load_cfg': loadCfg,
          });
          AppMessage.success(t('common.saveSuccessfully'));
          ctx.emit(events.refresh);
        } catch (err) {
          handleOperationResponseError(err);
          ctx.emit(events.cancel);
        }
      });
    }

    const onDialogCancel = (formElement) => {
      ctx.emit(events.cancel);
    }

    const parseTime = (hhmm) => {
      if (!hhmm) {
        return '';
      }
      return new Date(`1970 ${hhmm}`);
    }

    const formatTime = (date) => {
      if (!date) {
        return null;
      }
      let hour = (date.getHours() + '').padStart(2, '0');
      let min = (date.getMinutes() + '').padStart(2, '0');
      return `${hour}:${min}`;
    }

    const initRule = () => {
      if (isCreateNew()) {
        return;
      }
      let currRule = props.rules[props.editRuleIndex];
      ruleForm.ruleName = currRule.ruleName;
      ruleForm.startTime = parseTime(currRule?.start_time ?? '');
      ruleForm.endTime = parseTime(currRule?.end_time ?? '');
      ruleForm.ipAddress = currRule?.ip_addr ?? '';
      ruleForm.macAddress = currRule?.mac_addr ?? '';
      ruleForm.status = currRule.enable;
    }

    onMounted(async () => {
      initRule();
    })

    return {
      parseTime,
      ruleForm,
      ruleFormRules,
      ruleFormRef,
      isWholeFormatMacAddress,
      onDialogConfirm,
      onDialogCancel,
    };
  },
};
</script>

<style lang="scss" scoped>
.rule-tip {
  font-size: 12px;
  font-family: HarmonyOS_Sans_SC,sans-serif;
  color: var(--el-customize-tip-text-color);
  line-height: 16px;
  font-weight: 400;
  margin-left: 10px;
}

#input-time {
:deep(.el-form-item__content) {
  display: block;
}
}

</style>