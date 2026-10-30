<template>
	<InputWrapper
		:wrapper-style="inputWrapperStyleFrom(ui.style)"
		:label="ui.label"
		:error="ui.errorText"
		:help="ui.supportingText"
		:disabled="ui.disabled"
		:input-id="id"
		:optional="ui.optional"
	>
		<div class="relative">
			<input
				:id="id"
				class="input-field"
				role="button"
				:aria-invalid="!!ui.errorText"
				:aria-description="ui.errorText || ui.supportingText"
				:aria-disabled="ui.disabled"
				:aria-label="`${ui.label} (${ui.value && ui.value.sVG ? $t('signature-field.input.has-value') : $t('signature-field.input.no-value')})`"
				readonly
				@click="onFocus"
				@keydown.enter="onFocus"
			/>
			<VHtml v-if="ui.value && ui.value.sVG" :html="ui.value.sVG" class="signature" aria-hidden="true" />
		</div>
	</InputWrapper>
</template>
<script lang="ts" setup>
import { inputWrapperStyleFrom } from '@/components/shared/inputWrapperStyle';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { SignatureField } from '@/shared/proto/nprotoc_gen';
import { randomStr } from '@/components/shared/util';
import VHtml from '@/components/VHtml.vue';

const props = defineProps<{
	ui: SignatureField;
}>();

const emit = defineEmits<{
	(e: 'expand'): void;
}>();

function onFocus(): void {
	if (props.ui.disabled) return;
	emit('expand');
}

const id = randomStr(16);
</script>
<style scoped>
@reference '@/assets/css/tailwind.css';
.input-field {
	@apply h-20 cursor-pointer;
}

.signature {
	@apply absolute top-0 left-0 size-full flex justify-center items-center pointer-events-none;
}
</style>
<style>
@reference '@/assets/css/tailwind.css';
.signature {
	& > svg {
		@apply max-h-full max-w-full;
	}
}
</style>
