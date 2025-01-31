<script lang="ts" setup>
import { WindowTitle } from '@/shared/protocol/ora/windowTitle';

const props = defineProps<{
	ui: WindowTitle;
}>();

if (props.ui.v) {
	document.title = props.ui.v;
}
</script>

<template></template>
