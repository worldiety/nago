<template>
	<div class="signature-overlay">
		<div class="dialog">
			<div class="flex justify-between items-center gap-x-2">
				<p class="truncate">{{ ui.label ?? '' }}</p>
				<button
					ref="closeButton"
					class="close-button button-tertiary square"
					tabindex="0"
					@click="$emit('close')"
					@keydown.enter="$emit('close')"
				>
					<Close class="h-3.5" aria-hidden="true" />
				</button>
			</div>

			<hr class="border-I0" />

			<canvas ref="canvas" />

			<hr class="border-ST0" />

			<div class="footer">
				<button class="button-tertiary" :disabled="!clearButtonVisible" @click="clearInputValue">
					<eraser-icon class="h-4 fill-current" aria-hidden="true" />
					{{ $t('signature-field.overlay.reset') }}
				</button>
				<button class="button-confirm button-primary" @click="onSubmit">
					{{ $t('signature-field.overlay.submit') }}
				</button>
			</div>
		</div>

		<!-- Blurred Background -->
		<div class="bg" @click="$emit('close')"></div>
	</div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue';
import SignaturePad from 'signature_pad';
import Close from '@/assets/svg/closeBold.svg';
import { Signature, SignatureField } from '@/shared/proto/nprotoc_gen';
import EraserIcon from '@/assets/svg/eraser.svg';

defineProps<{
	ui: SignatureField;
}>();

const emit = defineEmits<{
	(e: 'close'): void;
	(e: 'submit', signature: Signature): void;
}>();

const canvas = ref<HTMLCanvasElement>();
const signaturePad = ref<SignaturePad>();

const clearButtonVisible = ref(false);

function onSubmit(): void {
	if (!signaturePad.value) return;
	emit('submit', new Signature(signaturePad.value.toSVG()));
}

function clearInputValue(): void {
	if (!signaturePad.value) return;
	signaturePad.value.clear();
	clearButtonVisible.value = false;
}

onMounted(() => {
	if (!canvas.value) return;
	canvas.value.width = canvas.value.clientWidth;
	canvas.value.height = canvas.value.clientHeight;
	signaturePad.value = new SignaturePad(canvas.value, { penColor: 'currentColor' }); // TODO: Change color
	signaturePad.value.addEventListener('endStroke', () => {
		clearButtonVisible.value = !signaturePad.value?.isEmpty();
	});
});
</script>
<style scoped>
.signature-overlay {
	@apply fixed top-0 left-0 bottom-0 right-0 flex justify-center items-center z-30;

	hr {
		@apply my-3;
	}

	.dialog {
		@apply relative bg-M1 rounded-xl shadow-lg p-6 z-10;

		.close-button {
			@apply size-auto -mr-2.5 !p-2.5 text-ST0 focus:text-I0 hover:opacity-90;
		}

		canvas {
			@apply size-full min-w-96 min-h-64 bg-M2 rounded-lg shadow;
		}

		.footer {
			@apply flex items-center justify-between w-full gap-2;
		}
	}

	.bg {
		@apply absolute top-0 left-0 bottom-0 right-0 bg-opacity-60 bg-black z-0;
	}
}
</style>
