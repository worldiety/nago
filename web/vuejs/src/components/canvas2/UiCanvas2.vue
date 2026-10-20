<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { AssetLoader } from '@/components/gostrict/assets';
import { type ProgramHandle, startProgram } from '@/components/gostrict/run';
import type { Session } from '@/components/gostrict/session';
import { frameCSS } from '@/components/shared/frame';
import { Canvas2 } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Canvas2;
}>();

const rootRef = ref<HTMLDivElement | null>(null);

const frameStyles = computed<string>(() => frameCSS(props.ui.frame).join(';'));

/**
 * Host is the object, which the bindings of package lib/canvas call. The program draws in CSS pixels, the host
 * scales the canvas to the pixels of the device.
 */
interface Host {
	context: CanvasRenderingContext2D;
	dpr: number;
	width: number;
	height: number;
	loadImage(url: string): unknown;
}

let handle: ProgramHandle | null = null;
let canvas: HTMLCanvasElement | null = null;
let assets: AssetLoader | null = null;
let starts = 0;

async function start() {
	// a new version of the program takes over the state of the running one, if it has the same structure
	const previous = handle?.session.takeOver() ?? null;
	stop();
	const root = rootRef.value;
	if (!root || !props.ui.program?.bundle) return;
	const run = ++starts;

	const c = document.createElement('canvas');
	c.tabIndex = 0;
	c.className = 'block w-full h-full outline-hidden';
	c.style.touchAction = 'none';
	root.appendChild(c);
	canvas = c;

	let session: Session | null = null;
	const loader = new AssetLoader((type, values) => session?.emit(type, values));
	assets = loader;
	const host: Host = {
		context: c.getContext('2d')!,
		dpr: 1,
		width: 0,
		height: 0,
		loadImage: (url: string) => loader.image(url),
	};

	const h = await startProgram({
		id: props.ui.id ?? '',
		program: props.ui.program,
		root,
		input: c,
		libs: {},
		host,
		resize(width, height, dpr) {
			host.width = width;
			host.height = height;
			host.dpr = dpr;
			c.width = Math.max(1, Math.round(width * dpr));
			c.height = Math.max(1, Math.round(height * dpr));
			host.context.setTransform(dpr, 0, 0, dpr, 0, 0);
		},
		// each frame starts with the scale of the device and no other transformation
		before() {
			host.context.setTransform(host.dpr, 0, 0, host.dpr, 0, 0);
		},
		render() {},
		attached(s) {
			session = s;
		},
		previous,
	});
	if (run !== starts) {
		h?.stop();
		return;
	}
	handle = h;
}

function stop() {
	starts++;
	handle?.stop();
	handle = null;
	assets?.dispose();
	assets = null;
	canvas?.remove();
	canvas = null;
}

onMounted(start);
onUnmounted(stop);

// a new version of the program, e.g. after an edit, replaces the running one, while a render of nago, which
// changes nothing of it, keeps the program and its state
watch(
	() => `${props.ui.program?.bundle}|${props.ui.program?.init}`,
	() => start()
);
</script>

<template>
	<div :id="ui.id" ref="rootRef" :style="frameStyles" class="relative overflow-hidden"></div>
</template>
