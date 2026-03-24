<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { ref, watch } from 'vue';
import { bool2Str } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { Checkbox } from '@/shared/proto/nprotoc_gen';
import { UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Checkbox;
}>();

const serviceAdapter = useServiceAdapter();

const checked = ref<boolean>(props.ui.value ? props.ui.value : false);

watch(
	() => props.ui.value,
	(newValue) => {
		if (newValue) {
			checked.value = newValue;
		} else {
			checked.value = false;
		}
	}
);

function checkboxSelectedClick(event: Event): void {
	if (!props.ui.disabled) {
		event.stopPropagation();
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), bool2Str(!checked.value))
		);
	}
}

function checkboxSelected(): void {
	if (!props.ui.disabled) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), bool2Str(!checked.value))
		);
	}
}
</script>

<template>
	<div
		v-if="!ui.invisible"
		class="input-checkbox"
		:class="{ 'input-checkbox-disabled': ui.disabled }"
		@click="checkboxSelectedClick"
		@keydown.enter="checkboxSelected"
	>
		<input :id="ui.id" :checked="checked" type="checkbox" :disabled="ui.disabled" />
	</div>
</template>

<style scoped>
.input-checkbox {
	@apply relative cursor-pointer p-2.5 text-[0];

	input {
		@apply appearance-none cursor-pointer overflow-hidden;
		@apply relative size-4 rounded-[0.175rem] outline outline-1 -outline-offset-1 outline-current;

		&:checked {
			&:before {
				content: '';
				@apply block absolute size-full top-0 left-0 bg-I0;
			}

			&:after {
				content: '';
				font-size: 1rem;
				@apply block absolute size-3 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2;
				filter: invert(1);
				box-shadow: inset 1em 1em currentColor;
				clip-path: polygon(21% 49%, 10% 62%, 45% 90%, 90% 32%, 75% 19%, 42% 65%);
			}
		}
	}

	&:before {
		content: '';
		@apply block absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 size-full rounded-full bg-I0 bg-opacity-0 scale-75 duration-100;
	}

	&.input-checkbox-disabled {
		@apply cursor-default saturate-50 opacity-60;

		input {
			@apply cursor-default;
		}
	}

	&:not(.input-checkbox-disabled) {
		&:hover,
		&:focus-within {
			input {
				@apply outline-I0;
			}

			&:before {
				@apply scale-100;
			}
		}

		&:hover {
			&:before {
				@apply bg-opacity-20;
			}
		}

		&:focus-within {
			&:before {
				@apply bg-opacity-30;
			}
		}

		&:active {
			&:before {
				@apply bg-opacity-40;
			}
		}
	}
}
</style>
