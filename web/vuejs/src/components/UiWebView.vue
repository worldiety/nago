<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { WebView } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: WebView;
}>();

const styles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);

	return styles.join(';');
});
</script>

<template>
	<!--	<div v-html='props.ui.value.value'></div> TODO not sure about the semantics we want-->
	<iframe
		:title="props.ui.title"
		:allow="props.ui.allow"
		:referrerpolicy="props.ui.referrerPolicy"
		:style="styles"
		:srcdoc="props.ui.raw"
		:src="props.ui.uRI"
	></iframe>
</template>
