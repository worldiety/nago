<template>
	<div class="fixed top-0 left-0 right-0 text-black dark:text-white bg-white dark:bg-darkmode-gray border-b border-b-disabled-background dark:border-b-disabled-text h-24 py-4 px-8 z-30">
		<!-- Top bar -->
		<div class="relative flex justify-start items-center h-full">
			<MenuIcon class="relative cursor-pointer h-6 z-10" tabindex="0" @click="menuOpen = true" @keydown.enter="menuOpen = true" />
			<div class="absolute top-0 left-0 bottom-0 right-0 flex justify-center items-center h-full z-0">
				<div class="h-full *:h-full" v-html="ui.logo.v"></div>
			</div>
		</div>

		<!-- Menu -->
		<Transition name="slide">
			<div v-if="menuOpen" class="fixed top-0 left-0 bottom-0 w-80 bg-white dark:bg-darkmode-gray shadow-md z-20">
				<div class="flex justify-start items-center h-24 p-8 mb-8">
					<CloseIcon tabindex="0" class="cursor-pointer h-6" @click="menuOpen = false" @keydown.enter="menuOpen = false" />
				</div>
				<div class="flex flex-col justify-start items-start gap-y-4 p-4">
					<!-- Top level menu entries -->
					<MenuEntry v-for="(menuEntry, index) in ui.menu.v" :key="index" :ui="menuEntry" />
				</div>
			</div>
		</Transition>
	</div>
</template>

<script setup lang="ts">
import MenuIcon from '@/assets/svg/menu.svg';
import CloseIcon from '@/assets/svg/closeBold.svg';
import type { NavigationComponent } from '@/shared/protocol/ora/navigationComponent';
import { ref } from 'vue';
import MenuEntry from '@/components/scaffold/burgermenu/MenuEntry.vue';

defineProps<{
	ui: NavigationComponent;
}>();

const menuOpen = ref<boolean>(false);
</script>

<style scoped>
/* Vue transitions: https://vuejs.org/guide/built-ins/transition#css-transitions */
.slide-enter-active,
.slide-leave-active {
	@apply transform duration-200 ease-in-out;
}

.slide-enter-from,
.slide-leave-to {
	@apply translate-x-[-100%];
}
</style>
