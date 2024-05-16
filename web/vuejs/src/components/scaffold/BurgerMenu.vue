<template>
	<div class="fixed top-0 left-0 right-0 text-black dark:text-white bg-white dark:bg-darkmode-gray border-b border-b-disabled-background dark:border-b-disabled-text h-24 p-4 z-30">
		<!-- Top bar -->
		<div class="relative flex justify-start items-center h-full">
			<MenuIcon class="relative cursor-pointer h-6 z-10" tabindex="0" @click="menuOpen = true" @keydown.enter="menuOpen = true" />
			<div class="absolute top-0 left-0 bottom-0 right-0 flex justify-center items-center h-full z-0">
				<div class="h-full *:h-full" v-html="ui.logo.v"></div>
			</div>
		</div>

		<!-- Menu -->
		<Transition name="slide">
			<div v-if="menuOpen" class="fixed top-0 left-0 bottom-0 w-80 bg-white z-20">
				<div class="flex justify-start items-center h-24 p-4 mb-8">
					<CloseIcon class="cursor-pointer h-6" />
				</div>
				<div class="flex flex-col justify-start items-start gap-y-4 p-4">
					<!-- Top level menu entries -->
					<div
						v-for="(menuEntry, index) in ui.menu.v"
						:key="index"
						class="menu-entry flex justify-start items-center gap-x-2 cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35 rounded-full w-full p-2"
						tabindex="0"
						@click="expandMenuEntry(menuEntry)"
						@keydown.enter="expandMenuEntry(menuEntry)"
					>
						<div class="relative h-full">
							<div class="menu-entry-icon h-4 *:h-full" v-html="menuEntry.icon.v"></div>
							<div class="menu-entry-icon-active h-4 *:h-full" v-html="menuEntry.iconActive.v"></div>
							<!-- Optional red badge -->
							<div
								v-if="menuEntry.badge.v"
								class="absolute -top-1.5 -right-1.5 flex justify-center items-center h-3.5 px-1 rounded-full bg-error"
							>
								<p class="text-xs text-white">{{ menuEntry.badge.v }}</p>
							</div>
						</div>
						<p class="grow leading-tight select-none">{{ menuEntry.title.v }}</p>
						<TriangleDown class="shrink-0 basis-2" :class="{'rotate-180': menuEntry.expanded.v}" />
					</div>
				</div>
			</div>
		</Transition>
	</div>
</template>

<script setup lang="ts">
import MenuIcon from '@/assets/svg/menu.svg';
import CloseIcon from '@/assets/svg/closeBold.svg';
import TriangleDown from '@/assets/svg/triangleDown.svg';
import type { NavigationComponent } from '@/shared/protocol/ora/navigationComponent';
import { ref } from 'vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';

defineProps<{
	ui: NavigationComponent;
}>();

const serviceAdapter = useServiceAdapter();
const menuOpen = ref<boolean>(false);

function expandMenuEntry(menuEntry: MenuEntry): void {
	serviceAdapter.setPropertiesAndCallFunctions([
		{
			...menuEntry.expanded,
			v: true,
		},
	], [menuEntry.onFocus]);
}
</script>

<style scoped>
.menu-entry:hover .menu-entry-icon,
.menu-entry .menu-entry-icon-active {
	@apply hidden;
}

.menu-entry:hover .menu-entry-icon-active {
	@apply block;
}

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
