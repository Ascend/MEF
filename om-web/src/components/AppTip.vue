<template>
  <div class="app-tip" :class="[props.tipType, props.tipTheme]">
    <img v-if="props.tipType === 'info'" src="@/assets/img/common/info.svg" alt=""/>
    <img v-if="props.tipType === 'warn'" src="@/assets/img/alarm.svg" alt=""/>
    <div v-if="tipText">
      {{ tipText }}
    </div>
    <div v-if="tipList">
      <div v-for="(text, index) in props.tipList" :key="index">
        {{ text }}
      </div>
    </div>

  </div>
</template>

<script>
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'AppTip',
  props: {
    tipType: {
      type: String,
      default: 'info',
    },
    tipText: {
      type: String,
    },
    tipList: Array,
    tipTheme: {
      type: String,
      default: 'dark',
    },
  },
  setup(props) {
    return {
      props,
    }
  },
})
</script>

<style scoped>
.dark.info {
  background-color: var(--el-bg-color-tip);
}

.dark.warn {
  background-color: var(--el-bg-warn-tip);
}

.light.info {
  background-color: var(--el-text-customize-blue-bg-color);
  border: 1px solid var(--el-text-customize-color-primary);
  border-radius: 2px;
}

.app-tip {
  padding: 10px 20px;
  margin-bottom: 10px;
  display: flex;
  align-items: flex-start;
  border-radius: 4px;
}

.app-tip img {
  margin-right: 10px;
  height: 16px;
  width: 16px;
}

.app-tip div {
  word-break: break-word;
  font-size: 12px;
  letter-spacing: 0;
  line-height: 16px;
  font-weight: 400;
}
</style>