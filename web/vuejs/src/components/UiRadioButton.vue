<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { ref, watch } from 'vue';
import { bool2Str, randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Radiobutton, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Radiobutton;
}>();

const id = props.ui.id || randomStr(16);

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

function radioButtonClicked(event: MouseEvent): void {
	if (!props.ui.disabled) {
		// Like the checkbox, the radio button handles its click itself. Otherwise, the click bubbles up to a clickable
		// stack around it (e.g. a selectable card), which sends its action for the same click as well. The server then
		// renders twice, because the action of the stack refers to the tree before the state update and is stale.
		// This also covers the click, which the browser forwards to the input for a click on an associated label.
		event.stopPropagation();
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), bool2Str(!checked.value))
		);
	}
}
</script>

<template>
	<div v-if="!ui.invisible" class="radio-button-container">
		<div class="input-radio" :class="{ 'input-radio-disabled': ui.disabled }" @click="radioButtonClicked">
			<input :id="id" :name="ui.name" :checked="checked" type="radio" :disabled="ui.disabled" />
		</div>
		<label v-if="ui.label" :for="id" :class="{ disabled: ui.disabled }">
			{{ ui.label }}
		</label>
	</div>
</template>

<style scoped>
@reference '@/assets/css/tailwind.css';

.radio-button-container {
	@apply flex items-center text-[0px];

	.input-radio {
		@apply relative cursor-pointer p-2.5;

		input {
			@apply appearance-none cursor-pointer overflow-hidden;
			@apply relative size-4 rounded-full outline-solid outline-1 -outline-offset-1 outline-current;

			&:checked {
				&:before {
					content: '';
					@apply block absolute size-3 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-I0 rounded-full;
				}
			}
		}

		&:before {
			content: '';
			@apply block absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 size-full rounded-full bg-I0/0 scale-75 duration-100;
		}

		&.input-radio-disabled {
			@apply cursor-default saturate-50 opacity-60;

			input {
				@apply cursor-default;
			}
		}

		&:not(.input-radio-disabled) {
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
					@apply bg-I0/20;
				}
			}

			&:focus-within {
				&:before {
					@apply bg-I0/30;
				}
			}

			&:active {
				&:before {
					@apply bg-I0/40;
				}
			}
		}
	}

	label {
		@apply text-base cursor-pointer pl-1;

		&.disabled {
			@apply opacity-60 cursor-default;
		}
	}
}
</style>
