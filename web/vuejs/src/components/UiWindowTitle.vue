<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { WindowTitle } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: WindowTitle;
}>();

if (props.ui.value) {
	document.title = props.ui.value;
}
</script>

<template></template>
