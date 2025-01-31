<template>
	<nav class="fixed top-0 left-0 right-0 text-black h-24 z-30 bg-M1">
		<!-- Top bar -->
		<div class="relative bg-M1 h-full py-5 z-20 flex items-center">
			<div class="website-content w-full flex justify-between items-center">
				<div class="h-full *:h-full">
					<!-- nav bar icon -->
					<ui-generic v-if="props.ui.l" :ui="props.ui.l" />
				</div>
				<div class="flex justify-end items-center gap-x-6 h-full">
					<!-- Top level menu entries -->
					<div
						v-for="(menuEntry, index) in ui.m"
						:key="index"
						ref="menuEntryElements"
						class="h-full"
						:data-index="index"
					>
						<TopLevelMenuEntry
							:ui="menuEntry"
							:menu-entry-index="index"
							:mode="'navigationBar'"
							@focus-first-linked-sub-menu-entry="focusFirstLinkedSubMenuEntry"
							@expand="expandMenuEntry"
						/>
					</div>
					<!--					<ThemeToggle/>-->
				</div>
			</div>
		</div>

		<div class="relative z-10">
			<!-- Navigation bar border -->
			<div ref="navigationBarBorder" class="absolute top-0 left-0 right-0 border-b border-b-M5 z-0"></div>
			<!-- Sub menu triangle -->
			<div
				v-show="subMenuEntries.length > 0"
				ref="subMenuTriangle"
				class="sub-menu-triangle absolute -top-2 left-0 rotate-45 border border-disabled-background bg-primary-98 darkmode:bg-primary-10 size-4 z-10"
				:style="`--sub-menu-triangle-left-offset: ${subMenuTriangleLeftOffset}px`"
			></div>
		</div>

		<!-- Sub menu -->
		<Transition name="slide">
			<div
				v-if="subMenuEntries.length > 0"
				ref="subMenu"
				class="relative bg-M4 rounded-b-2xl shadow-md pt-8 pb-10 z-0"
			>
				<div class="website-content flex justify-center items-start gap-x-8">
					<!-- Sub menu entries -->
					<div v-for="(subMenuEntry, subMenuEntryIndex) in subMenuEntries" :key="subMenuEntryIndex">
						<p
							ref="subMenuEntryElements"
							class="font-medium rounded-full px-2"
							:class="{
								'mb-4': subMenuEntry.m?.length > 0,
								'cursor-pointer hover:underline focus-visible:underline': subMenuEntry.a,
								'bg-M7 bg-opacity-35': isActiveMenuEntry(subMenuEntry),
							}"
							:tabindex="subMenuEntry.a ? '0' : '-1'"
							@click="menuEntryClicked(subMenuEntry)"
							@keydown.enter="menuEntryClicked(subMenuEntry)"
						>
							{{ subMenuEntry.t }}
						</p>
						<!-- Sub sub menu entries -->
						<p
							v-for="(subSubMenuEntry, subSubMenuEntryIndex) in subMenuEntry.m"
							:key="subSubMenuEntryIndex"
							ref="subSubMenuEntryElements"
							class="sub-sub-menu-entry rounded-full px-2"
							:class="{
								'cursor-pointer hover:underline focus-visible:underline': subSubMenuEntry.a,
								'bg-M7 bg-opacity-35': isActiveMenuEntry(subSubMenuEntry),
							}"
							:tabindex="subSubMenuEntry.a ? '0' : '-1'"
							@click="menuEntryClicked(subSubMenuEntry)"
							@keydown.enter="menuEntryClicked(subSubMenuEntry)"
						>
							{{ subSubMenuEntry.t }}
						</p>
					</div>
				</div>
			</div>
		</Transition>
	</nav>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';
import TopLevelMenuEntry from '@/components/scaffold/TopLevelMenuEntry.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { Scaffold } from '@/shared/protocol/ora/scaffold';
import { ScaffoldMenuEntry } from '@/shared/protocol/ora/scaffoldMenuEntry';

const props = defineProps<{
	ui: Scaffold;
}>();

const serviceAdapter = useServiceAdapter();
const subMenuEntryElements = ref<HTMLElement[]>([]);
const subSubMenuEntryElements = ref<HTMLElement[]>([]);
const navigationBarBorder = ref<HTMLElement | undefined>();
const subMenu = ref<HTMLElement | undefined>();
const menuEntryElements = ref<HTMLElement[]>([]);
const subMenuTriangle = ref<HTMLElement | undefined>();
const subMenuTriangleLeftOffset = ref<number>(0);

onMounted(() => {
	document.addEventListener('mousemove', handleMouseMove);
	window.addEventListener('resize', updateSubMenuTriangleLeftOffset, { passive: true });
});

onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
	window.removeEventListener('resize', updateSubMenuTriangleLeftOffset);
});

watch(
	() => props.ui,
	() => {
		nextTick(updateSubMenuTriangleLeftOffset);
	}
);

const expandedMenuEntry = computed((): ScaffoldMenuEntry | undefined => {
	return props.ui.m?.find((menuEntry) => menuEntry.x);
});

const subMenuEntries = computed((): ScaffoldMenuEntry[] => {
	const entries: ScaffoldMenuEntry[] = props.ui.m
		?.filter((menuEntry) => menuEntry.x)
		.flatMap((menuEntry) => menuEntry.m ?? []);
	// Add the expanded menu entry without its sub menu entries, if it has an action
	if (entries.length > 0 && expandedMenuEntry.value?.a) {
		entries.unshift({
			...expandedMenuEntry.value,
			m: {
				...expandedMenuEntry.value.m,
				v: [],
			},
		});
	}
	return entries;
});

function isActiveMenuEntry(menuEntry: ScaffoldMenuEntry): boolean {
	// Active, if its component factory ID matches the current page's path name
	if (menuEntry.f == '.' && (window.location.pathname == '' || window.location.pathname == '/')) {
		return true;
	}

	return `/${menuEntry.f}` === window.location.pathname;
}

function handleMouseMove(event: MouseEvent): void {
	const threshold =
		subMenu.value?.getBoundingClientRect().bottom ?? navigationBarBorder.value?.getBoundingClientRect().bottom ?? 0;
	if (event.y > threshold) {
		// Collapse the sub menu when threshold is passed
		// const updatedExpandedProperties = props.ui.m
		// 	?.filter((menuEntry) => menuEntry.x)
		// 	.map((menuEntry) => ({
		// 		...menuEntry.x,
		// 		v: false,
		// 	}));
		// if (updatedExpandedProperties.length > 0) {
		// 	serviceAdapter.setProperties(...updatedExpandedProperties);
		// }

		props.ui.m?.forEach((value) => (value.x = false));
	}
}

function updateSubMenuTriangleLeftOffset(): void {
	const activeMenuEntryIndex: number | undefined = props.ui.m?.findIndex((menuEntry) => menuEntry.x);
	if (!subMenuTriangle.value || activeMenuEntryIndex === undefined) {
		return;
	}
	const activeMenuEntryElement = menuEntryElements.value.find((element) => {
		return element.getAttribute('data-index') === activeMenuEntryIndex + '';
	});
	if (!activeMenuEntryElement) {
		return;
	}
	subMenuTriangleLeftOffset.value =
		activeMenuEntryElement.getBoundingClientRect().x +
		activeMenuEntryElement.offsetWidth / 2 -
		subMenuTriangle.value.offsetWidth / 2;
}

function menuEntryClicked(menuEntry: ScaffoldMenuEntry): void {
	if (menuEntry.a) {
		serviceAdapter.executeFunctions(menuEntry.a);
	}
}

function focusFirstLinkedSubMenuEntry(): void {
	const elementToFocus =
		subMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0) ??
		subSubMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0);
	elementToFocus?.focus();
}

function expandMenuEntry(menuEntry: ScaffoldMenuEntry): void {
	// const propertiesToSet: Property<boolean>[] = props.ui.m.map((entry) => {
	// 	return {
	// 		...entry.x,
	// 		v: entry.id === menuEntry.id,
	// 	};
	// });
	//
	// serviceAdapter.setPropertiesAndCallFunctions(propertiesToSet, [menuEntry.onFocus]);

	if (!props.ui.m) {
		return;
	}

	for (let i = 0; i < props.ui.m.length; i++) {
		let m = props.ui.m.at(i)!;
		m.x = false;
	}

	menuEntry.x = true;

	updateSubMenuTriangleLeftOffset();
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
