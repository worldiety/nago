<script lang="ts" setup>
import { WebView } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: WebView;
}>();
</script>

<template>
	<!--	<div v-html='props.ui.value.value'></div> TODO not sure about the semantics we want-->
	<iframe style="width: 100%; height: 100vh" :srcdoc="props.ui.uRI.value"></iframe>
</template>
