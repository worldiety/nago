<template>
	<button
		type="button"
		class="button-secondary !p-2 !size-10"
		@click="toggleDarkMode"
	>
		<MoonIcon v-if="darkModeActive" class="h-full" />
		<SunIcon v-else class="h-full" />
	</button>
</template>

<script setup lang="ts">
import SunIcon from '@/assets/svg/sun.svg';
import MoonIcon from '@/assets/svg/moon.svg';
import { onMounted, ref } from 'vue';

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
