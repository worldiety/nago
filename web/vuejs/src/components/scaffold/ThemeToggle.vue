<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

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
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { windowInfoChanged } from '@/eventhandling';
import { ThemeKey, useThemeManager } from '@/shared/themeManager';

const themeManager = useThemeManager();
const darkModeActive = ref<boolean>(false);
const service = useServiceAdapter();

onMounted(() => {
	darkModeActive.value = themeManager.getActiveThemeKey() === ThemeKey.DARK;
});

function toggleDarkMode() {
	themeManager.toggleDarkMode();
	darkModeActive.value = themeManager.getActiveThemeKey() === ThemeKey.DARK;

	//eventBus.publish(EventType.WindowInfoChanged, {});
	windowInfoChanged(service, themeManager);
}
</script>
