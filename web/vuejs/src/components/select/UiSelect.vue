<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<div class="ui-select" :style="frameStyles">
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

				<SelectORA
					v-if="ui.oRADropdown"
					v-model="selectedValue"
					:class="styleClass"
					:ui="ui"
					:style="inputStyle"
				/>
				<SelectDefault v-else v-model="selectedValue" :class="styleClass" :ui="ui" :style="inputStyle" />
			</div>
		</InputWrapper>
	</div>
</template>
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
import SelectDefault from '@/components/select/SelectDefault.vue';
import SelectORA from '@/components/select/SelectORA.vue';

const props = defineProps<{
	ui: Select;
}>();

const serviceAdapter = useServiceAdapter();
const leadingElement = useTemplateRef('leadingElement');
const selectedValue = ref(props.ui.value);

const styleClass = computed<string>(() => {
	switch (props.ui.style) {
		case TextFieldStyleValues.TextFieldBasic:
			return 'basic';
		case TextFieldStyleValues.TextFieldOutlined:
			return 'outlined';
		case TextFieldStyleValues.TextFieldReduced:
			return 'reduced';
		default:
			return 'outlined';
	}
});

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
