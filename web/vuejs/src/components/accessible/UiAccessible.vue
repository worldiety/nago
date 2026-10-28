<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { Accessible } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Accessible;
}>();

// attrs contains only the set attributes, because Vue omits attributes with an undefined value
const attrs = computed<Record<string, string | undefined>>(() => ({
	'role': props.ui.role || undefined,
	'aria-label': props.ui.label || undefined,
	'aria-labelledby': props.ui.labelledBy || undefined,
	'aria-hidden': props.ui.hidden ? 'true' : undefined,
	'aria-valuetext': props.ui.valueText || undefined,
	'aria-valuenow': props.ui.valueNow || undefined,
	'aria-valuemin': props.ui.valueMin || undefined,
	'aria-valuemax': props.ui.valueMax || undefined,
	'aria-expanded': props.ui.expanded || undefined,
	'aria-selected': props.ui.selected || undefined,
	'aria-disabled': props.ui.disabled || undefined,
}));
</script>

<template>
	<!-- UiAccessible: no own element, the attributes fall through to the root element of the content -->
	<UiGeneric v-if="props.ui.content" :ui="props.ui.content" v-bind="attrs" />
</template>
