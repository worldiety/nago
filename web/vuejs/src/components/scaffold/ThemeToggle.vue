<template>
	<button
		type="button"
		class="border rounded-full size-10 p-2"
		@click="toggleDarkMode"
	>
		<MoonIcon v-if="darkModeActive" class="h-full" />
		<SunIcon v-else class="h-full" />
	</button>
</template>

<script setup lang="ts">
import SunIcon from '@/assets/svg/sun.svg';
import MoonIcon from '@/assets/svg/moon.svg';
import {onMounted, ref} from 'vue';
import {ThemeKey, useThemeManager} from '@/shared/themeManager';
import {useEventBus} from "@/composables/eventBus";
import {EventType} from "@/shared/eventbus/eventType";

const themeManager = useThemeManager();
const darkModeActive = ref<boolean>(false);
const eventBus = useEventBus();
onMounted(() => {
	darkModeActive.value = themeManager.getActiveThemeKey() === ThemeKey.DARK;
});

function toggleDarkMode() {
	themeManager.toggleDarkMode();
	darkModeActive.value = themeManager.getActiveThemeKey() === ThemeKey.DARK;

	eventBus.publish(EventType.WindowInfoChanged, {})
}
</script>
