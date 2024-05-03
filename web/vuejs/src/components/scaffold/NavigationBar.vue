<template>
	<nav class="fixed top-0 left-0 right-0 text-black dark:text-white z-30">
		<!-- Top bar -->
		<div class="relative bg-white dark:bg-darkmode-gray h-20 py-2 z-20">
			<div class="website-content flex justify-between items-center h-full">
				<div class="h-full *:h-full" v-html="ui.logo.v"></div>
        <div class="flex justify-end items-center gap-x-8">
					<div v-for="(menuEntry, index) in ui.menu.v" :key="index" ref="menuEntryElements" :data-index="index">
						<MenuEntryComponent
							:ui="menuEntry"
							:menu-entry-index="index"
							:expanded="menuEntry.id === activeMenuEntry?.id"
							@expand-menu-entry="expandMenuEntry"
							@collapse-menu-entry="collapseMenuEntry"
						/>
					</div>
          <ThemeToggle />
        </div>
			</div>
		</div>

		<div class="relative z-10">
			<!-- Navigation bar border -->
			<div ref="navigationBarBorder" class="absolute top-0 left-0 right-0 border-b border-b-disabled-background dark:border-b-disabled-text z-0"></div>
			<!-- Sub menu triangle -->
			<div
				v-show="subMenuEntries.length > 0"
				ref="subMenuTriangle"
				class="sub-menu-triangle absolute -top-2 left-0 rotate-45 border border-disabled-background bg-white dark:bg-darkmode-gray dark:border-disabled-text size-4 z-10"
				:style="`--sub-menu-triangle-left-offset: ${subMenuTriangleLeftOffset}px`"
			></div>
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
						<a v-if="subMenuEntry.url.v" :href="subMenuEntry.url.v" class="font-medium">{{ subMenuEntry.title.v }}</a>
						<p v-else class="font-medium" :class="{'mb-4': subMenuEntry.menu.v?.length > 0}">{{ subMenuEntry.title.v }}</p>
						<p
							v-for="(subSubMenuEntry, subSubMenuEntryIndex) in subMenuEntry.menu.v"
							:key="subSubMenuEntryIndex"
						>
							<a v-if="subSubMenuEntry.url.v" :href="subSubMenuEntry.url.v">{{ subSubMenuEntry.title.v }}</a>
							<span v-else>{{ subSubMenuEntry.title.v }}</span>
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
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue';
import type { MenuEntry } from '@/shared/protocol/gen/menuEntry';

defineProps<{
	ui: NavigationComponent;
}>();

const navigationBarBorder = ref<HTMLElement|undefined>();
const subMenu = ref<HTMLElement|undefined>();
const activeMenuEntry = ref<MenuEntry|null>(null);
const activeMenuEntryIndex = ref<number|null>(null);
const menuEntryElements = ref<HTMLElement[]>([]);
const subMenuTriangle = ref<HTMLElement|undefined>();
const subMenuTriangleLeftOffset = ref<number>(0);

onMounted(() => {
	document.addEventListener('mousemove', handleMouseMove);
	window.addEventListener('resize', updateSubMenuTriangleLeftOffset, { passive: true });
});

onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
	window.removeEventListener('resize', updateSubMenuTriangleLeftOffset);
});

const subMenuEntries = computed((): MenuEntry[] => activeMenuEntry.value?.menu.v ?? []);

function handleMouseMove(event: MouseEvent): void {
	const threshold = subMenu.value?.getBoundingClientRect().bottom
		?? navigationBarBorder.value?.getBoundingClientRect().bottom
		?? 0;
	if (event.y > threshold) {
		activeMenuEntry.value = null;
	}
}

function expandMenuEntry(menuEntry: MenuEntry, menuEntryIndex: number): void {
	setActiveMenuEntry(menuEntry, menuEntryIndex);
	nextTick(updateSubMenuTriangleLeftOffset);
}

function collapseMenuEntry(): void {
	setActiveMenuEntry(null, null);
}

function setActiveMenuEntry(menuEntry: MenuEntry|null, menuEntryIndex: number|null): void {
	activeMenuEntry.value = menuEntry;
	activeMenuEntryIndex.value = menuEntryIndex;
}

function updateSubMenuTriangleLeftOffset(): void {
	if (!subMenuTriangle.value || activeMenuEntryIndex.value === null) {
		return;
	}
	const activeMenuEntryElement = menuEntryElements.value.find((element) => {
		return element.getAttribute('data-index') === activeMenuEntryIndex.value + '';
	});
	if (!activeMenuEntryElement) {
		return;
	}
	subMenuTriangleLeftOffset.value = activeMenuEntryElement.getBoundingClientRect().x + activeMenuEntryElement.offsetWidth / 2 - subMenuTriangle.value.offsetWidth / 2;
}
</script>

<style scoped>
.sub-menu-triangle {
	left: var(--sub-menu-triangle-left-offset);
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
