<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import { borderCSS } from '@/components/shared/border';
import { frameCSS } from '@/components/shared/frame';
import { marginCSS } from '@/components/shared/padding';
import { Divider } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Divider;
}>();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.border);
	styles.push(...frameCSS(props.ui.frame));
	styles.push(...marginCSS(props.ui.padding));

	return styles.join(';');
});
</script>

<template>
	<div :style="styles" />
</template>
