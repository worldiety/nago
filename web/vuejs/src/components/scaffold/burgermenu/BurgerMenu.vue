<template>
	<div class="fixed top-0 left-0 right-0 text-black bg-background border-b border-b-disabled-background h-24 py-4 px-8 z-30">
		<!-- Top bar -->
		<div class="relative flex justify-start items-center h-full">
			<MenuIcon class="relative cursor-pointer h-6 z-10" tabindex="0" @click="menuOpen = true" @keydown.enter="menuOpen = true" />
			<div class="absolute top-0 left-0 bottom-0 right-0 flex justify-center items-center h-full z-0">
				<div class="h-full *:h-full" v-html="ui.logo.v"></div>
			</div>
		</div>

		<!-- Menu -->
		<Transition name="slide">
			<div
				v-if="menuOpen"
				class="fixed top-0 left-0 bottom-0 flex flex-col justify-start items-start w-full xs:w-80 bg-background shadow-md z-20"
			>
				<div class="flex justify-start items-center h-24 p-8">
					<CloseIcon tabindex="0" class="cursor-pointer h-6" @click="menuOpen = false" @keydown.enter="menuOpen = false" />
				</div>
				<div class="flex flex-col justify-start items-start gap-y-4 overflow-y-auto basis-full w-full p-4">
					<template v-if="!subMenuVisible">
						<!-- Top level menu entries -->
						<BurgerMenuEntry
							v-for="(menuEntry, index) in ui.menu.v"
							:key="index"
							:ui="menuEntry"
							:top-level="true"
							@clicked="menuOpen = false"
						/>
					</template>
					<div v-else class="flex flex-col justify-start items-start gap-y-4 w-full pl-4">
						<!-- Back to top level menu button -->
						<div
							tabindex="0"
							class="relative flex justify-start items-center gap-x-2 cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35 rounded-full w-full py-4 pl-8 pr-4 -ml-4"
							@click="returnToTopLevelMenu"
							@keydown.enter="returnToTopLevelMenu"
						>
							<div class="h-6">
								<TriangleDown class="absolute top-0 left-4 bottom-0 rotate-90 h-2 my-auto" />
							</div>
							<p class="leading-tight font-semibold">{{ $t('scaffold.toMenu') }}</p>
						</div>
						<!-- Top level menu entry title button -->
						<div
							:tabindex="expandedTopLevelMenuEntryLinked ? '0' : '-1'"
							class="flex justify-start items-center gap-x-2 rounded-full w-full p-4"
							:class="{
								'cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35': expandedTopLevelMenuEntryLinked,
								'bg-disabled-background bg-opacity-35': expandedTopLevelMenuEntryActive,
							}"
							@click="navigateToExpandedTopLevelMenuEntry"
							@keydown.enter="navigateToExpandedTopLevelMenuEntry"
						>
							<div class="flex justify-start items-center h-6">
								<p class="leading-tight font-semibold">{{ expandedTopLevelMenuEntry?.title.v }}</p>
							</div>
						</div>
						<!-- Sub menu entries -->
						<BurgerMenuEntry
							v-for="(menuEntry, index) in subMenuEntries"
							:key="index"
							:ui="menuEntry"
							:top-level="false"
							@clicked="menuOpen = false"
						/>
					</div>
				</div>
				<div class="flex justify-center items-center w-full p-4">
					<ThemeToggle />
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
import { computed, ref, watch } from 'vue';
import BurgerMenuEntry from '@/components/scaffold/burgermenu/BurgerMenuEntry.vue';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: NavigationComponent;
}>();

const serviceAdapter = useServiceAdapter();
const menuOpen = ref<boolean>(false);

const expandedTopLevelMenuEntry = computed((): MenuEntry|null => {
	return props.ui.menu.v?.find((menuEntry: MenuEntry) => menuEntry.expanded.v) ?? null;
});

const expandedTopLevelMenuEntryLinked = computed((): boolean => {
	return !!expandedTopLevelMenuEntry.value && !!expandedTopLevelMenuEntry.value.action.v;
});

const expandedTopLevelMenuEntryActive = computed((): boolean => {
	return !!expandedTopLevelMenuEntry.value && `/${expandedTopLevelMenuEntry.value.componentFactoryId.v}` === window.location.pathname;
});

const subMenuVisible = computed((): boolean => {
	const expandedTopLevelMenuEntry = props.ui.menu.v?.find((menuEntry: MenuEntry) => menuEntry.expanded.v);
	return !!expandedTopLevelMenuEntry?.menu.v;
});

const subMenuEntries = computed((): MenuEntry[] => {
	if (!props.ui.menu.v) {
		return [];
	}
	const expandedTopLevelMenuEntry = props.ui.menu.v?.find((menuEntry: MenuEntry) => menuEntry.expanded.v);
	if (!expandedTopLevelMenuEntry) {
		return props.ui.menu.v;
	}
	return expandedTopLevelMenuEntry.menu.v ?? props.ui.menu.v;
});

function navigateToExpandedTopLevelMenuEntry(): void {
	if (!expandedTopLevelMenuEntry.value?.action.v) {
		return;
	}
	menuEntryClicked(expandedTopLevelMenuEntry.value);
}

function menuEntryClicked(menuEntry: MenuEntry): void {
	if (!menuEntry.action.v) {
		return;
	}
	serviceAdapter.executeFunctions(menuEntry.action);
}

function returnToTopLevelMenu(): void {
	if (!expandedTopLevelMenuEntry.value) {
		return;
	}
	serviceAdapter.setPropertiesAndCallFunctions([{
		...expandedTopLevelMenuEntry.value.expanded,
		v: false,
	}], [expandedTopLevelMenuEntry.value.onFocus])
}
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
