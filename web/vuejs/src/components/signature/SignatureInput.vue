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
			<input class="input-field" readonly @click="onFocus" @keydown.enter="onFocus" />
			<VHtml v-if="ui.value && ui.value.sVG" :html="ui.value.sVG" class="signature" />
		</div>
	</InputWrapper>
</template>
<script lang="ts" setup>
import { inputWrapperStyleFrom } from '@/components/shared/inputWrapperStyle';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { SignatureField } from '@/shared/proto/nprotoc_gen';
import { randomStr } from '@/components/shared/util';
import VHtml from '@/components/VHtml.vue';

defineProps<{
	ui: SignatureField;
}>();

const emit = defineEmits<{
	(e: 'expand'): void;
}>();

function onFocus(): void {
	emit('expand');
}

const id = randomStr(16);
</script>
<style scoped>
.input-field {
	@apply h-20 cursor-pointer;
}

.signature {
	@apply absolute top-0 left-0 size-full flex justify-center items-center pointer-events-none;
}
</style>
<style>
.signature {
	& > svg {
		@apply max-h-full max-w-full;
	}
}
</style>
