<template>
	<nav class="fixed top-0 left-0 right-0 text-black dark:text-white z-30">
		<!-- Top bar -->
		<div class="relative bg-white dark:bg-darkmode-gray h-24 py-5 z-20">
			<div class="website-content flex justify-between items-center h-full">
				<div class="h-full *:h-full" v-html="ui.logo.v"></div>
        <div class="flex justify-end items-center gap-x-6 h-full">
					<!-- Top level menu entries -->
					<div v-for="(menuEntry, index) in ui.menu.v" :key="index" ref="menuEntryElements" class="h-full" :data-index="index">
						<MenuEntryComponent
							:ui="menuEntry"
							:menu-entry-index="index"
							:mode="'navigationBar'"
							@focus-first-linked-sub-menu-entry="focusFirstLinkedSubMenuEntry"
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
				<div class="website-content flex justify-center items-start gap-x-8">
					<!-- Sub menu entries -->
					<div v-for="(subMenuEntry, subMenuEntryIndex) in subMenuEntries" :key="subMenuEntryIndex">
						<p
							ref="subMenuEntryElements"
							class="font-medium"
							:class="{
							'mb-4': subMenuEntry.menu.v?.length > 0,
							'cursor-pointer hover:underline focus-visible:underline': subMenuEntry.action.v,
						}"
							:tabindex="subMenuEntry.action.v ? '0' : '-1'"
							@click="menuEntryClicked(subMenuEntry)"
							@keydown.enter="menuEntryClicked(subMenuEntry)"
						>
							{{ subMenuEntry.title.v }}
						</p>
						<!-- Sub sub menu entries -->
						<p
							v-for="(subSubMenuEntry, subSubMenuEntryIndex) in subMenuEntry.menu.v"
							:key="subSubMenuEntryIndex"
							ref="subSubMenuEntryElements"
							class="sub-sub-menu-entry"
							:class="{'cursor-pointer hover:underline focus-visible:underline': subSubMenuEntry.action.v}"
							:tabindex="subSubMenuEntry.action.v ? '0' : '-1'"
							@click="menuEntryClicked(subSubMenuEntry)"
							@keydown.enter="menuEntryClicked(subSubMenuEntry)"
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
import type { NavigationComponent } from '@/shared/protocol/ora/navigationComponent';
import MenuEntryComponent from '@/components/scaffold/TopLevelMenuEntry.vue';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: NavigationComponent;
}>();

const serviceAdapter = useServiceAdapter();
const subMenuEntryElements = ref<HTMLElement[]>([]);
const subSubMenuEntryElements = ref<HTMLElement[]>([]);
const navigationBarBorder = ref<HTMLElement|undefined>();
const subMenu = ref<HTMLElement|undefined>();
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

watch(() => props.ui, () => {
	nextTick(updateSubMenuTriangleLeftOffset);
});

const expandedMenuEntry = computed((): MenuEntry|undefined => {
	return props.ui.menu.v?.find((menuEntry) => menuEntry.expanded.v);
})

const subMenuEntries = computed((): MenuEntry[] => {
	const entries: MenuEntry[] = props.ui.menu.v
		?.filter((menuEntry) => menuEntry.expanded.v)
		.flatMap((menuEntry) => menuEntry.menu.v ?? []);
	if (entries.length > 0 && expandedMenuEntry.value?.action.v) {
		entries.unshift({
			...expandedMenuEntry.value,
			menu: {
				...expandedMenuEntry.value.menu,
				v: [],
			}
		});
	}
	return entries;
});

function handleMouseMove(event: MouseEvent): void {
	const threshold = subMenu.value?.getBoundingClientRect().bottom
		?? navigationBarBorder.value?.getBoundingClientRect().bottom
		?? 0;
	if (event.y > threshold) {
		// Collapse the sub menu when threshold is passed
		const updatedExpandedProperties = props.ui.menu.v
			?.filter((menuEntry) => menuEntry.expanded.v)
			.map((menuEntry) => ({
				...menuEntry.expanded,
				v: false,
			}));
		if (updatedExpandedProperties.length > 0) {
			serviceAdapter.setProperties(...updatedExpandedProperties);
		}
	}
}

function updateSubMenuTriangleLeftOffset(): void {
	const activeMenuEntryIndex: number|undefined = props.ui.menu.v?.findIndex((menuEntry) => menuEntry.expanded.v);
	if (!subMenuTriangle.value || activeMenuEntryIndex === undefined) {
		return;
	}
	const activeMenuEntryElement = menuEntryElements.value.find((element) => {
		return element.getAttribute('data-index') === activeMenuEntryIndex + '';
	});
	if (!activeMenuEntryElement) {
		return;
	}
	subMenuTriangleLeftOffset.value = activeMenuEntryElement.getBoundingClientRect().x + activeMenuEntryElement.offsetWidth / 2 - subMenuTriangle.value.offsetWidth / 2;
}

function menuEntryClicked(menuEntry: MenuEntry): void {
	if (menuEntry.action.v) {
		serviceAdapter.executeFunctions(menuEntry.action);
	}
}

function focusFirstLinkedSubMenuEntry(): void {
	const elementToFocus =
		subMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0)
		?? subSubMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0);
	elementToFocus?.focus();
}
</script>

<style scoped>
.sub-menu-triangle {
	left: var(--sub-menu-triangle-left-offset);
}

.sub-sub-menu-entry:not(:last-child) {
	@apply mb-2;
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
