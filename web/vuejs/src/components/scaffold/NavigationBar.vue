<template>
	<nav class="fixed top-0 left-0 right-0 text-black dark:text-white z-30">
		<!-- Top bar -->
		<div class="relative bg-white dark:bg-darkmode-gray h-20 py-2 z-20">
			<div class="website-content flex justify-between items-center h-full">
				<div class="h-full *:h-full" v-html="ui.logo.v"></div>
        <div class="flex justify-end items-center gap-x-8">
          <MenuEntryComponent
						v-for="(menuEntry, index) in ui.menu.v"
						:key="index"
						:ui="menuEntry"
						@menu-entry-hovered="setActiveMenuEntry"
					/>
          <ThemeToggle />
        </div>
			</div>
		</div>

		<!-- Sub menu triangle -->
		<!-- TODO: Show triangle at the appropriate location -->
		<div class="relative z-10">
			<div class="absolute top-0 left-0 right-0 border-b border-b-disabled-background dark:border-b-disabled-text z-0"></div>
			<div class="absolute -top-2 left-64 rotate-45 border border-disabled-background bg-white size-4 z-10"></div>
		</div>

		<!-- Sub menu -->
		<Transition name="slide">
			<div
				v-if="subMenuEntries.length > 0"
				ref="subMenu"
				class="relative bg-white dark:bg-darkmode-gray rounded-b-2xl shadow-md pt-8 pb-10 z-0"
			>
				<div class="website-content flex justify-start items-start gap-x-8">
					<div v-for="(subMenuEntry, subMenuEntryIndex) in subMenuEntries" :key="subMenuEntryIndex">
						<p class="font-medium" :class="{'mb-4': subMenuEntry.menu.v?.length > 0}">{{ subMenuEntry.title.v }}</p>
						<p
							v-for="(subSubMenuEntry, subSubMenuEntryIndex) in subMenuEntry.menu.v"
							:key="subSubMenuEntryIndex"
						>
							{{ subSubMenuEntry.title.v }}
						</p>
					</div>
				</div>
			</div>
		</Transition>
	</nav>
</template>

<script setup lang="ts">
import type { NavigationComponent } from '@/shared/protocol/gen/navigationComponent';
import MenuEntryComponent from '@/components/scaffold/MenuEntryComponent.vue';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import type { MenuEntry } from '@/shared/protocol/gen/menuEntry';

defineProps<{
	ui: NavigationComponent;
}>();

const subMenu = ref<HTMLElement|undefined>();
const activeMenuEntry = ref<MenuEntry|null>(null);

onMounted(() => {
	document.addEventListener('mousemove', handleMouseMove);
});

onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
});

const subMenuEntries = computed((): MenuEntry[] => activeMenuEntry.value?.menu.v ?? []);

function handleMouseMove(event: MouseEvent): void {
	const threshold = subMenu.value?.getBoundingClientRect().bottom ?? 0;
	if (event.y > threshold) {
		activeMenuEntry.value = null;
	}
}

function setActiveMenuEntry(menuEntry: MenuEntry): void {
	activeMenuEntry.value = menuEntry;
}
</script>

<style scoped>
.triangle {
	@apply size-0 border-x-[20px] border-x-transparent border-t-[20px] border-t-white;
}

/* Vue transitions: https://vuejs.org/guide/built-ins/transition#css-transitions */
.slide-enter-active,
.slide-leave-active {
	@apply transform duration-200 ease-in-out;
}

.slide-enter-from,
.slide-leave-to {
	@apply translate-y-[-100%];
}
</style>
