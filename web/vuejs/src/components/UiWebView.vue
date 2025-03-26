<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { WebView } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: WebView;
}>();
</script>

<template>
	<!--	<div v-html='props.ui.value.value'></div> TODO not sure about the semantics we want-->
	<iframe style="width: 100%; height: 100vh" :srcdoc="props.ui.uRI"></iframe>
</template>
