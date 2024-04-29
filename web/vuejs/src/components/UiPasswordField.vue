<script lang="ts" setup>
import { ref, watch } from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import RevealIcon from '@/assets/svg/reveal.svg';
import HideIcon from '@/assets/svg/hide.svg';
import type { PasswordField } from '@/shared/protocol/gen/passwordField';
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: PasswordField;
}>();

const serviceAdapter = useServiceAdapter();
const passwordInput = ref<HTMLElement|undefined>();
const inputValue = ref<string>(props.ui.value.v);
const idPrefix = 'password-field-';

watch(inputValue, (newValue) => {
	serviceAdapter.setPropertiesAndCallFunctions([{
		...props.ui.value,
		v: newValue,
	}], [props.ui.onPasswordChanged]);
});

function toggleRevealed(): void {
	serviceAdapter.setProperties({
		...props.ui.revealed,
		v: !props.ui.revealed.v,
	});
	passwordInput.value?.focus();
}
</script>

<template>
	<div>
		<InputWrapper
			:simple="props.ui.simple.v"
			:label="props.ui.label.v"
			:error="props.ui.error.v"
			:hint="props.ui.hint.v"
			:help="props.ui.help.v"
			:disabled="props.ui.disabled.v"
		>
			<div class="relative hover:text-ora-orange focus-within:text-ora-orange">
				<input
					:id="idPrefix + props.ui.id.toString()"
					ref="passwordInput"
					v-model="inputValue"
					class="input-field"
					:class="{'!pr-12': inputValue}"
					:placeholder="props.ui.placeholder.v"
					:disabled="props.ui.disabled.v"
					:type="props.ui.revealed.v ? 'text' : 'password'"
				/>
				<div class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<div :tabindex="props.ui.disabled.v ? '-1' : '0'" @click="toggleRevealed" @keydown.enter="toggleRevealed">
						<RevealIcon v-if="!props.ui.revealed.v" class="w-6" />
						<HideIcon v-else class="w-6" />
					</div>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>
