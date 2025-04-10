<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<NavigationBar v-if="navigationBarVisible" :ui="props.ui" />
	<Sidebar v-if="sidebarVisible" :ui="props.ui" />
	<BurgerMenu v-if="burgerMenuVisible" :ui="props.ui" />

	<div class="min-h-full flex flex-col min-h-screen" :class="bodyWrapperClass">
		<div class="website-content min-h-full flex-grow w-full">
			<ui-generic v-if="props.ui.body" :ui="props.ui.body" />
		</div>

		<!-- Footer -->
		<footer v-if="props.ui.footer" class="">
			<ui-generic :ui="props.ui.footer" />
		</footer>
	</div>
</template>

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import NavigationBar from '@/components/scaffold/NavigationBar.vue';
import Sidebar from '@/components/scaffold/Sidebar.vue';
import BurgerMenu from '@/components/scaffold/burgermenu/BurgerMenu.vue';
import { Scaffold, ScaffoldAlignmentValues } from '@/shared/proto/nprotoc_gen';

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
	return (
		(windowWidth.value >= (props.ui.breakpoint ?? 768) && props.ui.alignment === undefined) ||
		props.ui.alignment === ScaffoldAlignmentValues.ScaffoldAlignmentTop
	);
});

const sidebarVisible = computed((): boolean => {
	return (
		windowWidth.value >= (props.ui.breakpoint ?? 768) &&
		props.ui.alignment === ScaffoldAlignmentValues.ScaffoldAlignmentLeading
	);
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

console.log(props.ui);

function updateWindowWidth(): void {
	windowWidth.value = window.innerWidth;
}
</script>
