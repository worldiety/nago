<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import type { Component as OraComponent } from '@/shared/proto/nprotoc_gen';
import { vueComponentFor } from '@/shared/uiComponents';

const props = defineProps<{
	ui: OraComponent; // if we don't rename the import, Vue confuses everything up
}>();
</script>

<template>
	<component :is="vueComponentFor(props.ui)" :ui="props.ui" />
</template>
