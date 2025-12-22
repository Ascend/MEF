<template>
  <el-popover placement="right" trigger="hover" v-if="text.length > 20 " :width="400">
    <template #reference>
      <img src="@/assets/img/common/question.svg" alt=""/>
    </template>
    <div style="word-break: break-word;">
      {{ text }}
    </div>
  </el-popover>
  <el-popover placement="right" trigger="hover" v-else>
    <template #reference>
      <img src="@/assets/img/common/question.svg" alt=""/>
    </template>
    <div style="word-break: break-word;">
      {{ text }}
    </div>
  </el-popover>
</template>

<script>
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'AppPopover',
  props: {
    text: String,
  },
})
</script>

<style scoped>
</style>