<template>
	<NavigationBar v-if="navigationBarVisible" :ui="props.ui.navigationComponent.v" />
	<Sidebar v-if="sidebarVisible" :ui="props.ui.navigationComponent.v" />
	<BurgerMenu v-if="burgerMenuVisible" :ui="props.ui.navigationComponent.v" />

  <div class="min-h-full" :class="bodyWrapperClass">
    <div class="website-content min-h-full">
      <ui-generic :ui="props.ui.body.v"  />
    </div>
  </div>
</template>

<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { Scaffold } from "@/shared/protocol/ora/scaffold";
import Sidebar from '@/components/scaffold/Sidebar.vue';
import NavigationBar from '@/components/scaffold/NavigationBar.vue';

import { computed, onMounted, onUnmounted, ref } from 'vue';
import BurgerMenu from '@/components/scaffold/burgermenu/BurgerMenu.vue';
import {Alignment} from "@/components/shared/alignments";

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
	return windowWidth.value >= 768 && props.ui.navigationComponent.v.alignment.v === Alignment.Top;
});

const sidebarVisible = computed((): boolean => {
	return windowWidth.value >= 768 && props.ui.navigationComponent.v.alignment.v === Alignment.Leading;
});

const burgerMenuVisible = computed((): boolean => {
	return !navigationBarVisible.value && !sidebarVisible.value;
});

const bodyWrapperClass = computed((): string|undefined => {
	if (burgerMenuVisible.value || navigationBarVisible.value) {
		return 'pt-28';
	}
  if (sidebarVisible.value) {
    return 'py-8 pl-32';
  }
	return undefined;
});

function updateWindowWidth(): void {
	windowWidth.value = window.innerWidth;
}
</script>
