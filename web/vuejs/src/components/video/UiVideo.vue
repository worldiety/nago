<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import { Video } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Video;
}>();

const styles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);

	return styles.join(';');
});
</script>

<template>
	<video
		:controls="ui.controls || undefined"
		:autoplay="ui.autoplay || undefined"
		:loop="ui.loop || undefined"
		:poster="ui.poster || undefined"
		:muted="ui.muted || undefined"
		:playsinline="ui.playsInline || undefined"
		:style="styles"
	>
		<source :src="ui.src" />
	</video>
</template>
