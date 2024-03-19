<script lang="ts" setup>
import { onMounted } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveScaffold } from '@/shared/model/liveScaffold';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: LiveScaffold;
	page: LivePage;
}>();

const networkStore = useNetworkStore();

function initDarkModeToggle() {
	var themeToggleDarkIcon = document.getElementById('theme-toggle-dark-icon');
	var themeToggleLightIcon = document.getElementById('theme-toggle-light-icon');

	// Change the icons inside the button based on previous settings
	if (
		localStorage.getItem('color-theme') === 'dark' ||
		(!('color-theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)
	) {
		themeToggleLightIcon.classList.remove('hidden');
	} else {
		themeToggleDarkIcon.classList.remove('hidden');
	}

	var themeToggleBtn = document.getElementById('theme-toggle');

	themeToggleBtn.addEventListener('click', function () {
		// toggle icons inside button
		themeToggleDarkIcon.classList.toggle('hidden');
		themeToggleLightIcon.classList.toggle('hidden');

		// if set via local storage previously
		if (localStorage.getItem('color-theme')) {
			if (localStorage.getItem('color-theme') === 'light') {
				document.documentElement.classList.add('dark');
				localStorage.setItem('color-theme', 'dark');
			} else {
				document.documentElement.classList.remove('dark');
				localStorage.setItem('color-theme', 'light');
			}

			// if NOT set via local storage previously
		} else {
			if (document.documentElement.classList.contains('dark')) {
				document.documentElement.classList.remove('dark');
				localStorage.setItem('color-theme', 'light');
			} else {
				document.documentElement.classList.add('dark');
				localStorage.setItem('color-theme', 'dark');
			}
		}
	});
}

onMounted(() => {
	initDarkModeToggle();
});
</script>

<template>
	<div class="fixed z-30 flex w-full flex-1 flex-col dark:text-white">
		<nav class="flex h-16 justify-between bg-white px-4 shadow dark:bg-gray-700">
			<div class="flex items-center pl-4">
				<ui-generic v-if="props.ui.topbarLeft.value" :ui="props.ui.topbarLeft.value" :page="page" />
			</div>

			<div class="flex items-center">
				<ui-generic v-if="props.ui.topbarMid.value" :ui="props.ui.topbarMid.value" :page="page" />
			</div>

			<ul class="flex items-center pr-4">
				<ui-generic v-if="props.ui.topbarRight.value" :ui="props.ui.topbarRight.value" :page="page" />
			</ul>
		</nav>
	</div>

	<aside
		id="default-sidebar"
		class="fixed left-0 z-20 h-screen w-64 -translate-x-full pt-16 transition-transform sm:translate-x-0"
		aria-label="Sidebar"
	>
		<div class="h-full overflow-y-auto bg-gray-50 px-3 py-4 dark:bg-gray-800">
			<ul class="space-y-2 font-medium">
				<li v-for="btn in props.ui.menu.value">
					<a
						@click="networkStore.invokeFunc(btn.action)"
						class="group flex cursor-pointer items-center rounded-lg p-2 text-gray-900 hover:bg-gray-100 dark:text-white dark:hover:bg-gray-700"
					>
						<svg
							v-inline
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
							v-if="btn.preIcon.value"
							v-html="btn.preIcon.value"
						></svg>
						<span class="ms-3">{{ btn.caption.value }}</span>
						<svg
							v-inline
							class="h-5 w-5 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
							v-if="btn.postIcon.value"
							v-html="btn.postIcon.value"
						></svg>
					</a>
				</li>
			</ul>

			<div class="flex flex-col-reverse">
				<div>
					<button
						id="theme-toggle"
						type="button"
						class="rounded-lg p-2.5 text-sm text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-4 focus:ring-gray-200 dark:text-gray-400 dark:hover:bg-gray-700 dark:focus:ring-gray-700"
					>
						<svg
							id="theme-toggle-dark-icon"
							class="hidden h-5 w-5"
							fill="currentColor"
							viewBox="0 0 20 20"
							xmlns="http://www.w3.org/2000/svg"
						>
							<path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z"></path>
						</svg>
						<svg
							id="theme-toggle-light-icon"
							class="hidden h-5 w-5"
							fill="currentColor"
							viewBox="0 0 20 20"
							xmlns="http://www.w3.org/2000/svg"
						>
							<path
								d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z"
								fill-rule="evenodd"
								clip-rule="evenodd"
							></path>
						</svg>
					</button>
				</div>
			</div>
		</div>
	</aside>

	<div class="p-4 pb-16 pt-16 sm:ml-64 sm:pb-0">
		<div class="p-4">
			<nav v-if="props.ui.breadcrumbs.value" class="flex pb-4" aria-label="Breadcrumb">
				<ol class="inline-flex items-center space-x-1 rtl:space-x-reverse md:space-x-2">
					<li v-for="btn in props.ui.breadcrumbs.value" class="inline-flex items-center">
						<a
							@click="networkStore.invokeFunc(btn.action)"
							class="inline-flex cursor-pointer items-center text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-400 dark:hover:text-white"
						>
							<svg
								v-inline
								class="me-2.5 h-3 w-3"
								v-if="btn.preIcon.value"
								v-html="btn.preIcon.value"
							></svg>
							{{ btn.caption.value }}
						</a>
					</li>
				</ol>
			</nav>

			<ui-generic :ui="props.ui.body.value" :page="page" />
		</div>
	</div>

	<div
		class="fixed bottom-0 left-0 z-20 h-16 w-full border-t border-gray-200 bg-white dark:border-gray-600 dark:bg-gray-700 sm:hidden"
	>
		<div class="mx-auto grid h-full max-w-lg auto-cols-auto grid-flow-col font-medium">
			<button
				type="button"
				class="group inline-flex cursor-pointer flex-col items-center justify-center px-5 hover:bg-gray-50 dark:hover:bg-gray-800"
				v-for="btn in props.ui.menu.value"
				@click="networkStore.invokeFunc(btn.action)"
			>
				<svg
					v-inline
					class="mb-2 h-5 w-5 text-gray-500 group-hover:text-blue-600 dark:text-gray-400 dark:group-hover:text-blue-500"
					v-if="btn.preIcon.value"
					v-html="btn.preIcon.value"
				></svg>
				<span
					class="text-sm text-gray-500 group-hover:text-blue-600 dark:text-gray-400 dark:group-hover:text-blue-500"
					>{{ btn.caption.value }}</span
				>
				<svg
					v-inline
					class="mb-2 h-5 w-5 text-gray-500 group-hover:text-blue-600 dark:text-gray-400 dark:group-hover:text-blue-500"
					v-if="btn.postIcon.value"
					v-html="btn.postIcon.value"
				></svg>
			</button>
		</div>
	</div>
</template>
