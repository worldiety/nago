<script setup lang="ts">
import { computed } from 'vue';
import { frameCSSObject } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { go } from '@codemirror/lang-go';
import { html } from '@codemirror/lang-html';
import { javascript } from '@codemirror/lang-javascript';
import { json } from '@codemirror/lang-json';
import { Extension } from '@codemirror/state';
import { oneDark } from '@codemirror/theme-one-dark';
import { ViewUpdate } from '@codemirror/view';
import { Codemirror } from 'vue-codemirror';
import { CodeEditor, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';
import { ThemeKey, useThemeManager } from '@/shared/themeManager';
import { css } from '@codemirror/lang-css';

const props = defineProps<{
	ui: CodeEditor;
}>();

const serviceAdapter = useServiceAdapter();

const frameStyles = computed<Object | undefined>(() => {
	let styles = frameCSSObject(props.ui.frame);

	return styles;
});

const themeManager = useThemeManager();

function onTextChange(value: string, event: ViewUpdate): any {
	//console.log('changed');
	return undefined;
}

let lastSent: string | undefined;

function onBlur(event: ViewUpdate): any {
	//console.log('blurred');
	if (props.ui.inputValue && lastSent !== props.ui.value) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, undefined, nextRID(), props.ui.value)
		);

		lastSent = props.ui.value;
	}
	return undefined;
}

function onFocus(event: ViewUpdate): any {
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
</script>

<template v-if="props.ui.iv">
	<codemirror
		:style="frameStyles"
		v-model="props.ui.value"
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
