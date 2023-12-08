<script lang="ts" setup>
import {invokeFunc, LivePage, LiveText, textColor2Tailwind, textSize2Tailwind} from "@/shared/livemsg";
import {computed} from "vue";

const props = defineProps<{
  ui: LiveText;
  ws: WebSocket;
  page: LivePage
}>();

const clazz = computed<string>(() => {
  let tmp = ""
  if (props.ui.color.value) {
    tmp += textColor2Tailwind(props.ui.color.value)
  }else{
    tmp+="text-gray-900"
  }

  if (props.ui.colorDark.value) {
    tmp += " dark:" + textColor2Tailwind(props.ui.color.value)
  }else{
    tmp+=" dark:text-white"
  }

  if (props.ui.size.value) {
    tmp += " " + textSize2Tailwind(props.ui.size.value)
  }

  return tmp
})

function onClick() {
  invokeFunc(props.ws, props.ui.onClick)
}

function onMouseEnter() {
  invokeFunc(props.ws, props.ui.onHoverStart)
}

function onMouseLeave() {
  invokeFunc(props.ws, props.ui.onHoverEnd)
}

</script>

<template>
  <span :class="clazz" @click="onClick" @mouseenter="onMouseEnter" @mouseleave="onMouseLeave">{{ props.ui.value.value }}</span>
</template>
