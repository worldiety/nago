<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<SwitcherComponent :key="key" :ui="ui" />
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import { randomStr } from '@/components/shared/util';
import SwitcherComponent from '@/components/switcher/Switcher.vue';
import type { Switcher } from '@/shared/proto/nprotoc_gen';

defineProps<{
	ui: Switcher;
}>();

const id = randomStr(16);
const refreshKey = ref(0);
const key = computed<string>(() => `${id}_r${refreshKey.value}`);

addEventListener('resize', () => refreshKey.value++);
</script>
