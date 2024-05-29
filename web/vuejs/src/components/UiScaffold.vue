<template>
	<div class="fixed z-30 flex w-full flex-1 flex-col">
		<nav class="flex h-16 justify-between bg-white px-4 shadow">
			<div class="flex items-center pl-4">
				<ui-generic v-if="props.ui.topbarLeft.v" :ui="props.ui.topbarLeft.v"  />
			</div>

			<div class="flex items-center">
				<ui-generic v-if="props.ui.topbarMid.v" :ui="props.ui.topbarMid.v"  />
			</div>

			<ul class="flex items-center pr-4">
				<ui-generic v-if="props.ui.topbarRight.v" :ui="props.ui.topbarRight.v"  />
			</ul>
		</nav>
	</div>

	<aside
		id="default-sidebar"
		class="fixed left-0 z-20 h-screen w-64 -translate-x-full pt-16 transition-transform sm:translate-x-0"
		aria-label="Sidebar"
	>
		<div class="h-full overflow-y-auto bg-gray-50 px-3 py-4">
			<ul class="space-y-2 font-medium">
				<li v-for="btn in props.ui.menu.v">
					<a
						@click="serviceAdapter.executeFunctions(btn.action)"
						class="group flex cursor-pointer items-center rounded-lg p-2 text-gray-900 hover:bg-gray-100"
					>
						<svg
							v-inline
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900"
							v-if="btn.preIcon.v"
							v-html="btn.preIcon.v"
						></svg>
						<span class="ms-3">{{ btn.caption.v }}</span>
						<svg
							v-inline
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900"
							v-if="btn.postIcon.v"
							v-html="btn.postIcon.v"
						></svg>
					</a>
				</li>
			</ul>

			<div class="flex flex-col-reverse">
				<div>
					<button
						type="button"
						class="rounded-lg p-2.5 text-sm text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-4 focus:ring-gray-200"
						@click="toggleDarkMode"
					>
						<MoonIcon v-if="darkModeActive" class="h-5 w-5" />
						<SunIcon v-else class="h-5 w-5" />
					</button>
				</div>
			</div>
		</div>
	</aside>

	<div class="p-4 pb-16 pt-16 sm:ml-64 sm:pb-0">
		<div class="p-4">
			<nav v-if="props.ui.breadcrumbs.v" class="flex pb-4" aria-label="Breadcrumb">
				<ol class="inline-flex items-center space-x-1 rtl:space-x-reverse md:space-x-2">
					<li v-for="btn in props.ui.breadcrumbs.v" class="inline-flex items-center">
						<a
							@click="serviceAdapter.executeFunctions(btn.action)"
							class="inline-flex cursor-pointer items-center text-sm font-medium text-gray-700 hover:text-blue-600"
						>
							<svg
								v-inline
								class="me-2.5 h-3 w-3"
								v-if="btn.preIcon.v"
								v-html="btn.preIcon.v"
							></svg>
							{{ btn.caption.v }}
						</a>
					</li>
				</ol>
			</nav>

			<ui-generic :ui="props.ui.body.v"  />
		</div>
	</div>

	<div
		class="fixed bottom-0 left-0 z-20 h-16 w-full border-t border-gray-200 bg-white sm:hidden"
	>
		<div class="mx-auto grid h-full max-w-lg auto-cols-auto grid-flow-col font-medium">
			<button
				v-for="(button, index) in props.ui.menu.v"
				:key="index"
				type="button"
				class="group inline-flex cursor-pointer flex-col items-center justify-center px-5 hover:bg-gray-50"
				@click="serviceAdapter.executeFunctions(button.action)"
			>
				<svg
					v-if="button.preIcon.v"
					v-inline
					class="mb-2 h-5 w-5 text-gray-500 group-hover:text-blue-600"
					v-html="button.preIcon.v"
				></svg>
				<span
					class="text-sm text-gray-500 group-hover:text-blue-600"
				>{{ button.caption.v }}</span
				>
				<svg
					v-inline
					class="mb-2 h-5 w-5 text-gray-500 group-hover:text-blue-600"
					v-if="button.postIcon.v"
					v-html="button.postIcon.v"
				></svg>
			</button>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import type {Scaffold} from "@/shared/protocol/ora/scaffold";
import { useServiceAdapter } from '@/composables/serviceAdapter';
import SunIcon from '@/assets/svg/sun.svg';
import MoonIcon from '@/assets/svg/moon.svg';

const props = defineProps<{
	ui: Scaffold;
}>();

const serviceAdapter = useServiceAdapter();
const darkModeActive = ref<boolean>(false);

onMounted(() => {
	darkModeActive.value = localStorage.getItem('color-theme') === 'dark' ||
	(!('color-theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)
});

function toggleDarkMode(): void {
	if ('color-theme' in localStorage) {
		// if set via local storage previously
		localStorage.getItem('color-theme') === 'light' ? activateDarkMode() : deactivateDarkMode();
	} else {
		// if NOT set via local storage previously
		document.documentElement.classList.contains('dark') ? deactivateDarkMode() : activateDarkMode();
	}
}

function activateDarkMode(): void {
	document.documentElement.classList.add('dark');
	localStorage.setItem('color-theme', 'dark');
	darkModeActive.value = true;
}

function deactivateDarkMode(): void {
	document.documentElement.classList.remove('dark');
	localStorage.setItem('color-theme', 'light');
	darkModeActive.value = false;
}
</script>
