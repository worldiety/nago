<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { watchEffect } from 'vue';
import { WindowTitle } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: WindowTitle;
}>();

// the component stays mounted while its value changes (a new render at the same place), so the title follows the
// value instead of being set only once when the component is created
watchEffect(() => {
	if (props.ui.value) {
		document.title = props.ui.value;
	}
});
</script>

<template>
	<teleport to="body"></teleport>
</template>
