<script setup lang="ts">
import LockIcon from '@/assets/svg/lock.svg';
import ErrorIcon from '@/assets/svg/closeBold.svg';
import { computed } from 'vue';

const props = defineProps<{
	label?: string;
	error?: string;
	hint?: string;
	disabled?: boolean;
}>();

const labelClass = computed((): string|null => {
	if (props.disabled) {
		return 'text-disabled-text';
	}
	if (props.error) {
		return 'text-error';
	}
	return null;
});

const inputFieldWrapperClass = computed((): string|null => {
	if (props.disabled) {
		return 'input-field-wrapper-disabled';
	}
	if (props.error) {
		return 'input-field-wrapper-error';
	}
	return null;
});
</script>

<template>
	<!-- Label -->
	<div class="flex justify-between items-end text-sm">
		<div v-if="label" class="flex justify-start items-center gap-x-1 mb-1" :class="labelClass">
			<LockIcon v-if="disabled" class="h-4" />
			<ErrorIcon v-else-if="error" class="h-2.5" />
			<span>{{ label }}</span>
		</div>
		<span v-if="hint" class="text-disabled-text">{{ hint }}</span>
	</div>

  <div class="relative">
    <div class="peer input-field-wrapper" :class="inputFieldWrapperClass">
      <slot />
    </div>
    <div class="input-field-outline"></div>
  </div>

	<p v-if="!disabled && error" class="mt-1 text-sm text-end text-error">{{ error }}</p>
</template>

<style>
.input-field-wrapper input, .input-field-wrapper .input-field {
  @apply relative bg-transparent border-0 border-b border-b-black text-black cursor-default w-full px-0;
  @apply dark:border-b-white dark:text-white;
}

.input-field-wrapper.input-field-wrapper-error input, .input-field-wrapper.input-field-wrapper-error .input-field {
	@apply border-b-error;
}

.input-field-wrapper input::placeholder {
  @apply text-disabled-text;
}

.input-field-wrapper:hover input, .input-field-wrapper:hover .input-field {
  @apply border-b-ora-orange border-opacity-75;
}

.input-field-wrapper input:active, .input-field-wrapper .input-field:active {
  @apply border-opacity-65;
}

.input-field-wrapper input:focus-visible, .input-field-wrapper .input-field:focus {
  @apply outline-none ring-0 border-b-transparent;
}

.input-field-wrapper input:disabled, .input-field-wrapper.input-field-wrapper-disabled .input-field {
  @apply border-b-disabled-text text-disabled-text;
}

.input-field-outline {
  @apply absolute top-0 -left-2 bottom-0 -right-2 rounded-sm pointer-events-none;
  @apply peer-focus-within:outline-none peer-focus-within:outline-2 peer-focus-within:outline-black peer-focus-within:ring-2 peer-focus-within:ring-white;
}

.input-field-wrapper {
	@apply text-black;
	@apply dark:text-white;
}

.input-field-wrapper.input-field-wrapper-disabled {
	@apply text-disabled-text pointer-events-none;
}
</style>
