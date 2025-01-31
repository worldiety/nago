<template>
	<div
		class="fixed top-0 left-0 w-screen h-screen z-50 bg-stone-900/50 backdrop-blur flex items-center justify-center"
	>
		<div class="flex flex-col items-center gap-4">
			<div class="">
				{{ $t('connectingChannelOverlay.content') }}
			</div>
			<LoadingAnimation class="w-5 h-5" />
		</div>
	</div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import LoadingAnimation from '@/components/shared/LoadingAnimation.vue';

const dots = ref(3);

function blurActiveElement(): void {
	const active = document.activeElement;
	if (active) {
		(active as HTMLElement).blur();
	}
}

function updateDots(): void {
	setInterval(() => {
		dots.value = (dots.value % 3) + 1;
	}, 1000);
}

onMounted(() => {
	blurActiveElement();
	updateDots();
});
</script>
