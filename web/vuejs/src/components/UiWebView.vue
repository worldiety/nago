<script lang="ts" setup>
import { useNetworkStore } from '@/stores/networkStore';
import {WebView} from "@/shared/protocol/gen/webView";

const props = defineProps<{
	ui: WebView;
}>();

</script>

<template>
<!--	<div v-html='props.ui.value.value'></div> TODO not sure about the semantics we want-->
	<iframe style="width: 100%;height:100vh" :srcdoc='props.ui.value.v'></iframe>

</template>
