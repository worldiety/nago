<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<div class="select-default">
		<div v-if="props.ui.leading" ref="leading" class="leading">
			<UiGeneric :ui="props.ui.leading" />
		</div>

		<select
			:id="id"
			v-model="selectedValue"
			:autocomplete="ui.autocomplete"
			class="input-field !pr-8 cursor-pointer"
			:disabled="props.ui.disabled"
			:style="{ paddingLeft: paddingLeft ? `${paddingLeft}px` : undefined }"
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
		<div class="chevron">
			<ArrowDownIcon class="size-3" />
		</div>
	</div>
</template>
<script lang="ts" setup>
import { computed, ref, useTemplateRef, watch } from 'vue';
import type { Select } from '@/shared/proto/nprotoc_gen';
import ArrowDownIcon from '@/assets/svg/arrowDown.svg';
import UiGeneric from '@/components/UiGeneric.vue';

interface Props {
	ui: Select;
}

interface Emits {
	(e: 'update:modelValue', value: string | undefined): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const leading = useTemplateRef('leading');
const selectedValue = ref(props.ui.value);

const paddingLeft = computed<number | undefined>(() => {
	return leading.value?.offsetWidth;
});

const id = computed<string>(() => {
	if (props.ui.id) {
		return props.ui.id;
	}

	return 'tf-' + props.ui.inputValue;
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
	emit('update:modelValue', selectedValue.value);
});
</script>
<style scoped>
.select-default {
	@apply relative;

	.leading {
		@apply absolute inset-y-0 left-0 pl-2 pr-1 flex items-center pointer-events-none;
	}

	.chevron {
		@apply absolute inset-y-0 right-0 pr-3 pl-1 flex items-center pointer-events-none;
	}

	&:hover .chevron {
		@apply text-I0;
	}
}
</style>
