<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { bool2Str } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { Toggle } from '@/shared/proto/nprotoc_gen';
import { UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Toggle;
}>();

const serviceAdapter = useServiceAdapter();

function toggle() {
	if (props.ui.disabled) {
		return;
	}

	serviceAdapter.sendEvent(
		new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), bool2Str(!props.ui.value))
	);
}
</script>

<template>
	<input
		:id="ui.id"
		class="toggle-switch"
		type="checkbox"
		:checked="!!props.ui.value"
		:disabled="props.ui.disabled"
		@change="toggle"
	/>
</template>

<style scoped>
.toggle-switch {
	@apply relative inline-block appearance-none rounded-full outline outline-1 -outline-offset-1 w-12 h-6 duration-100 cursor-pointer;
	@apply focus:outline-I0 focus-visible:outline-I0 focus:outline-2 focus:-outline-offset-2;

	&:before {
		content: '';
		@apply block size-[1.125rem] absolute top-[0.1875rem] left-[0.1875rem] rounded-full border border-current duration-100 origin-center;
	}

	&:checked {
		@apply bg-I0/50;

		&:before {
			@apply left-[1.6875rem] bg-current;
		}
	}

	&:hover {
		@apply bg-I0/25;
	}

	&:active {
		&:before {
			@apply scale-75;
		}
	}

	&:disabled {
		@apply opacity-50 pointer-events-none;
	}
}
</style>
