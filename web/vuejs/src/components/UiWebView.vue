<script lang="ts" setup>
import { useNetworkStore } from '@/stores/networkStore';
import type { LivePage } from '@/shared/model/livePage';
import {WebView} from "@/shared/model/webView";

const props = defineProps<{
	ui: WebView;
	page: LivePage;
}>();

</script>

<template>
<!--	<div v-html='props.ui.value.value'></div> TODO not sure about the semantics we want-->
	<iframe style="width: 100%;height:100vh" :srcdoc='props.ui.value.value'></iframe>

</template>
