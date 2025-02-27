<script lang="ts" setup>
import { WindowTitle } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: WindowTitle;
}>();

if (props.ui.value) {
	document.title = props.ui.value;
}
</script>

<template></template>
