<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, ref, useTemplateRef, watch } from 'vue';
import ArrowDownIcon from '@/assets/svg/arrowDown.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { inputWrapperStyleFrom } from '@/components/shared/inputWrapperStyle';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { Select } from '@/shared/proto/nprotoc_gen';
import { TextFieldStyleValues, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Select;
}>();

const serviceAdapter = useServiceAdapter();
const leadingElement = useTemplateRef('leadingElement');
const selectedValue = ref(props.ui.value);

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

const id = computed<string>(() => {
	if (props.ui.id) {
		return props.ui.id;
	}

	return 'tf-' + props.ui.inputValue;
});

const inputStyle = computed<string>(() => {
	const styles: string[] = [];

	if (props.ui.style == TextFieldStyleValues.TextFieldBasic) {
		styles.push('display:inline', 'background:unset');

		return styles.join(';');
	}

	const leadingElementWidth = leadingElement.value?.offsetWidth;
	const paddingLeft = leadingElementWidth ? `${leadingElementWidth}px` : 'auto';

	styles.push('padding-left:' + paddingLeft);
	return styles.join(';');
});

watch(
	() => props.ui.value,
	(newValue) => {
		if (document.getElementById(id.value) !== document.activeElement && selectedValue.value !== newValue) {
			selectedValue.value = newValue;
		}
	}
);

watch(selectedValue, () => {
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), selectedValue.value));
});
</script>

<template>
	<div :style="frameStyles">
		<InputWrapper
			:wrapper-style="inputWrapperStyleFrom(props.ui.style)"
			:label="props.ui.label"
			:error="props.ui.errorText"
			:help="props.ui.supportingText"
			:disabled="props.ui.disabled"
		>
			<div class="relative">
				<!-- Leading view -->
				<div
					v-if="props.ui.leading"
					ref="leadingElement"
					class="absolute inset-y-0 left-0 pl-2 pr-1 flex items-center pointer-events-none"
				>
					<UiGeneric :ui="props.ui.leading" />
				</div>

				<select
					:id="id"
					v-model="selectedValue"
					class="input-field !pr-8 cursor-pointer"
					:style="inputStyle"
					:disabled="props.ui.disabled"
				>
					<template v-if="props.ui.options">
						<option
							v-for="option in props.ui.options.value"
							:key="`select_${id}_${option.value}`"
							:value="option.value"
							:disabled="option.disabled"
						>
							{{ option.label }}
						</option>
					</template>
				</select>

				<!-- chevron -->
				<div class="absolute inset-y-0 right-0 pr-2 pl-1 flex items-center pointer-events-none">
					<ArrowDownIcon class="size-4" />
				</div>
			</div>
		</InputWrapper>
	</div>
</template>
