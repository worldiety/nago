<template>
	<button type="button" class="border rounded-full size-10 p-2" @click="toggleDarkMode">
		<MoonIcon v-if="darkModeActive" class="h-full" />
		<SunIcon v-else class="h-full" />
	</button>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import MoonIcon from '@/assets/svg/moon.svg';
import SunIcon from '@/assets/svg/sun.svg';
import { useEventBus } from '@/composables/eventBus';
import { EventType } from '@/shared/eventbus/eventType';
import { ThemeKey, useThemeManager } from '@/shared/themeManager';
import {windowInfoChanged} from "@/eventhandling";
import {useServiceAdapter} from "@/composables/serviceAdapter";

const themeManager = useThemeManager();
const darkModeActive = ref<boolean>(false);
const eventBus = useEventBus();
const service = useServiceAdapter();

onMounted(() => {
	darkModeActive.value = themeManager.getActiveThemeKey() === ThemeKey.DARK;
});

function toggleDarkMode() {
	themeManager.toggleDarkMode();
	darkModeActive.value = themeManager.getActiveThemeKey() === ThemeKey.DARK;

	//eventBus.publish(EventType.WindowInfoChanged, {});
	windowInfoChanged(service);
}
</script>
