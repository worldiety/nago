<template>
	<NavigationBar v-if="navigationBarVisible" :ui="props.ui" />
	<Sidebar v-if="sidebarVisible" :ui="props.ui" />
	<BurgerMenu v-if="burgerMenuVisible" :ui="props.ui" />

	<div class="min-h-full" :class="bodyWrapperClass">
		<div class="website-content min-h-full">
			<ui-generic :ui="props.ui.b" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import NavigationBar from '@/components/scaffold/NavigationBar.vue';
import Sidebar from '@/components/scaffold/Sidebar.vue';
import BurgerMenu from '@/components/scaffold/burgermenu/BurgerMenu.vue';
import { Alignment } from '@/components/shared/alignments';
import type { Scaffold } from '@/shared/protocol/ora/scaffold';

const props = defineProps<{
	ui: Scaffold;
}>();

const windowWidth = ref<number>(window.innerWidth);

onMounted(() => {
	window.addEventListener('resize', updateWindowWidth);
});

onUnmounted(() => {
	window.removeEventListener('resize', updateWindowWidth);
});

const navigationBarVisible = computed((): boolean => {
	return windowWidth.value >= 768 && props.ui.a === Alignment.Top;
});

const sidebarVisible = computed((): boolean => {
	return windowWidth.value >= 768 && props.ui.a === Alignment.Leading;
});

const burgerMenuVisible = computed((): boolean => {
	return !navigationBarVisible.value && !sidebarVisible.value;
});

const bodyWrapperClass = computed((): string | undefined => {
	if (burgerMenuVisible.value || navigationBarVisible.value) {
		return 'pt-28';
	}
	if (sidebarVisible.value) {
		return 'pl-32'; // py-8 would cause to introduce scrollbar with 100dvh
	}
	return undefined;
});

function updateWindowWidth(): void {
	windowWidth.value = window.innerWidth;
}
</script>
