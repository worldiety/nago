<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { frameCSSObject } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { go } from '@codemirror/lang-go';
import { html } from '@codemirror/lang-html';
import { javascript } from '@codemirror/lang-javascript';
import { json } from '@codemirror/lang-json';
import { Extension } from '@codemirror/state';
import { oneDark } from '@codemirror/theme-one-dark';
import { Codemirror } from 'vue-codemirror';
import { CodeEditor, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';
import { ThemeKey, useThemeManager } from '@/shared/themeManager';
import { css } from '@codemirror/lang-css';

const props = defineProps<{
	ui: CodeEditor;
}>();

const serviceAdapter = useServiceAdapter();

const value = ref(props.ui.value);

const frameStyles = computed<any | undefined>(() => {
	return frameCSSObject(props.ui.frame);
});

const themeManager = useThemeManager();

function onTextChange(): any {
	return undefined;
}

let lastSent: string | undefined;

function onBlur(): any {
	if (props.ui.inputValue && lastSent !== value.value) {
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, undefined, nextRID(), value.value));

		lastSent = value.value;
	}
	return undefined;
}

function onFocus(): any {
	return undefined;
}

function extensions(): Extension[] {
	let tmp: Extension[] = [];
	switch (props.ui.language) {
		case 'markdown':
			tmp.push(javascript());
			break;
		case 'go':
			tmp.push(go());
			break;
		case 'html':
			tmp.push(html());
			break;
		case 'css':
			tmp.push(css());
			break;
		case 'json':
			tmp.push(json());
			break;
	}
	if (themeManager.getActiveThemeKey() == ThemeKey.DARK) {
		tmp.push(oneDark);
	}

	return tmp;
}

watch(
	() => props.ui.value,
	() => (value.value = props.ui.value)
);
</script>

<template v-if="props.ui.iv">
	<codemirror
		v-model="value"
		:style="frameStyles"
		placeholder=""
		:autofocus="true"
		:indent-with-tab="true"
		:tab-size="props.ui.tabSize ? props.ui.tabSize : 2"
		:disabled="props.ui.disabled"
		:extensions="extensions()"
		@change="onTextChange"
		@focus="onFocus"
		@blur="onBlur"
	/>
</template>
