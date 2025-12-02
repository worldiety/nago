<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div class="flex flex-col-reverse">
		<!-- Input -->
		<div class="peer relative">
			<div class="peer input-field-wrapper" :class="inputFieldWrapperClasses">
				<slot/>
			</div>
		</div>

		<!-- Label with optional hint -->
		<div class="flex justify-between items-end text-sm" :class="{ 'peer-focus-within:font-semibold': !disabled }">
			<div v-if="label" class="flex justify-start items-center gap-x-1 pb-1" :class="labelClass">
				<LockIcon v-if="disabled" class="h-4"/>
				<ErrorIcon v-else-if="error" class="h-2.5"/>
				<span>{{ label }}</span>
			</div>
			<div v-if="!disabled && (error || hint)" class="font-normal">
				<span v-if="error" class="text-error">{{ t('inputWrapper.error') }}</span>
				<span v-else-if="hint" class="text-disabled-text">{{ hint }}</span>
			</div>
		</div>
	</div>

	<!-- Error message -->
	<div v-if="error || help" class="mt-1 text-sm">
		<p v-if="error" class="text-error">{{ error }}</p>
		<p v-else-if="help" class="text-disabled-text">{{ help }}</p>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue';
import {useI18n} from 'vue-i18n';
import ErrorIcon from '@/assets/svg/closeBold.svg';
import LockIcon from '@/assets/svg/lock.svg';
import {InputWrapperStyle} from '@/components/shared/inputWrapperStyle';

const props = defineProps<{
	wrapperStyle?: InputWrapperStyle;
	label?: string;
	error?: string;
	hint?: string;
	help?: string;
	disabled?: boolean;
	noHoverEffect?: boolean;
}>();

const {t} = useI18n();

const labelClass = computed((): string | null => {
	if (props.disabled) {
		return 'text-ST0';
	}
	if (props.error) {
		return 'text-SE0';
	}
	return null;
});

const inputFieldWrapperClasses = computed((): string | null => {
	const classes: string[] = [];

	if (props.wrapperStyle === InputWrapperStyle.REDUCED) {
		classes.push('input-field-wrapper-reduced');
	} else if (props.wrapperStyle === InputWrapperStyle.BASIC) {
		classes.push('input-field-wrapper-basic');
	} else {
		classes.push('input-field-wrapper-detailed');
	}

	if (props.disabled) {
		classes.push('input-field-wrapper-disabled');
	} else if (props.error) {
		classes.push('input-field-wrapper-error');
	}

	if (props.noHoverEffect) {
		classes.push('input-field-wrapper-no-hover');
	}

	return classes.join(' ') || null;
});
</script>

<style>
.input-field-wrapper .input-field {
	@apply relative bg-transparent border-M8 text-M8 w-full py-2;
}

.input-field-wrapper textarea.input-field {
	@apply resize-none max-h-80;
}

.input-field-wrapper.input-field-wrapper-detailed .input-field {
	@apply border rounded-md px-3;
}

.input-field-wrapper.input-field-wrapper-reduced .input-field,
.input-field-wrapper.input-field-wrapper-basic .input-field {
	@apply border-b;
}

.input-field-wrapper.input-field-wrapper-basic .input-field {
	@apply border-transparent;
}

.input-field-wrapper.input-field-wrapper-error .input-field {
	@apply border-SE0;
}

.text-error {
	@apply text-SE0;
}

.input-field-wrapper input::placeholder {
	@apply text-ST0;
}

.input-field-wrapper:not(.input-field-wrapper-no-hover):hover .input-field,
.input-field-wrapper .input-field:focus {
	@apply border-I0 border-opacity-75 text-M8;
}

.input-field-wrapper .input-field:focus {
	@apply border-I0 border-opacity-75 text-M8;
}

.input-field-wrapper .input-field:focus {
	@apply border-I0 border-opacity-75 text-M8;
}

.input-field-wrapper .input-field:focus {
	@apply outline-none ring-0;
}

.input-field-wrapper.input-field-wrapper-disabled .input-field {
	@apply border-b-ST0 text-ST0;
}

.input-field-wrapper.input-field-wrapper-detailed.input-field-wrapper-disabled .input-field,
.input-field-wrapper.input-field-wrapper-detailed.input-field-wrapper-disabled .input-field::placeholder {
	@apply border-ST0;
}

.input-field-wrapper {
	@apply text-M8;
}

.input-field-wrapper.input-field-wrapper-disabled {
	@apply text-ST0 pointer-events-none;
}
</style>
