<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import ConnectionHandler from '@/shared/network/connectionHandler';
import {
	CallRequested,
	Canvas,
	CanvasArc,
	CanvasArcTo,
	CanvasBeginPath,
	CanvasBezierCurveTo,
	CanvasCallList,
	CanvasClear,
	CanvasClearRect,
	CanvasClip,
	CanvasClosePath,
	CanvasDrawImage,
	CanvasEndList,
	CanvasFill,
	CanvasFillRect,
	CanvasFillStyle,
	CanvasFillText,
	CanvasFont,
	CanvasLineCap,
	CanvasLineJoin,
	CanvasLineTo,
	CanvasLineWidth,
	CanvasLoadImage,
	CanvasMiterLimit,
	CanvasMoveTo,
	CanvasNewList,
	CanvasQuadraticCurveTo,
	CanvasRect,
	CanvasRestore,
	CanvasRotate,
	CanvasSave,
	CanvasScale,
	CanvasSetTransform,
	CanvasShadowBlur,
	CanvasShadowColor,
	CanvasShadowOffsetX,
	CanvasShadowOffsetY,
	CanvasStrokeRect,
	CanvasStrokeStyle,
	CanvasStrokeText,
	CanvasTextAlign,
	CanvasTextBaseline,
	CanvasTranslate,
	type NagoEvent,
} from '@/shared/proto/nprotoc_gen';

const serviceAdapter = useServiceAdapter();

// Local aliases for the DOM canvas literal union types, because the imported proto
// classes shadow the global lib.dom.d.ts type names of the same name.
type CanvasLineCap_ = 'butt' | 'round' | 'square';
type CanvasLineJoin_ = 'bevel' | 'miter' | 'round';
type CanvasTextAlign_ = 'center' | 'end' | 'left' | 'right' | 'start';
type CanvasTextBaseline_ = 'alphabetic' | 'bottom' | 'hanging' | 'ideographic' | 'middle' | 'top';

const props = defineProps<{
	ui: Canvas;
}>();

const canvasRef = ref<HTMLCanvasElement | null>(null);

const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

function or0(v: number | undefined): number {
	if (v == undefined) return 0;
	return v;
}

let images = new Map<number, HTMLImageElement>();
let displayLists = new Map<number, CallRequested[]>();

/**
 * Extrahiert den Pixelwert aus einem CSS-String (z. B. "100px", "2rem").
 * Der canvas-width/height-Attribut muss einheitenlos (px) sein.
 */
function extractPixelValue(value: string | undefined): number | undefined {
	if (value == undefined) return undefined;
	if (value.endsWith('rem')) {
		const rem = parseFloat(value);
		const baseFontSize = parseFloat(getComputedStyle(document.documentElement).fontSize);
		return rem * baseFontSize;
	}
	if (value.endsWith('px')) {
		return parseFloat(value);
	}
	const n = parseFloat(value);
	return isNaN(n) ? undefined : n;
}

let eventCallback: ((evt: NagoEvent) => void) | null = null;

onMounted(() => {
	const ctx = canvasRef.value?.getContext('2d');
	if (!ctx) return;

	const id = props.ui.id;
	let activeList: CallRequested[] = [];
	let activeListHnd: number = 0;

	eventCallback = (invoke) => {
		if (invoke instanceof CallRequested) {
			// this assert is valid, all canvas calls have this field
			if (!invoke.call || (invoke.call as { id?: string }).id !== id) {
				return;
			}

			if (invoke.call instanceof CanvasNewList) {
				activeListHnd = invoke.call.handle!;
				activeList = [];
				return;
			}

			if (invoke.call instanceof CanvasEndList) {
				displayLists.set(activeListHnd, activeList);
				activeListHnd = 0;
				activeList = [];
				return;
			}

			if (invoke.call instanceof CanvasLoadImage && invoke.call.hnd !== undefined) {
				if (!images.has(invoke.call.hnd)) {
					let call = invoke.call;
					let img = new Image();
					img.onload = () => {
						console.log('image loaded', call.url, call.hnd);
						canvasRef.value?.dispatchEvent(new CustomEvent('invalidated', {}));
					};
					img.src = invoke.call.url!;
					images.set(invoke.call.hnd, img);
					console.log('loading image', invoke.call.url, 'with handle', invoke.call.hnd);
				}
			}

			if (activeListHnd !== 0) {
				activeList.push(invoke);
			} else {
				apply(ctx, invoke);
			}
		}
	};

	ConnectionHandler.addEventListener(eventCallback);
});

const MAX_CALL_DEPTH = 16;

function apply(ctx: CanvasRenderingContext2D, invoke: CallRequested, depth: number = 0) {
	// --- Rekursiver CallList-Aufruf ---
	if (invoke.call instanceof CanvasCallList) {
		if (depth >= MAX_CALL_DEPTH) {
			console.warn('CanvasCallList: max recursion depth reached', depth);
			return;
		}
		let list = displayLists.get(invoke.call.handle!);
		if (list == undefined) {
			console.log('display is undefined', invoke.call.handle);
			return;
		}
		list.forEach((call) => apply(ctx, call, depth + 1));
		return;
	}

	// --- Stil & Schrift ---
	if (invoke.call instanceof CanvasFillStyle) {
		if (invoke.call.style) {
			ctx.fillStyle = invoke.call.style;
		}
		return;
	}

	if (invoke.call instanceof CanvasFont) {
		if (invoke.call.font) {
			ctx.font = invoke.call.font;
		}
		return;
	}

	if (invoke.call instanceof CanvasStrokeStyle) {
		if (invoke.call.style) {
			ctx.strokeStyle = invoke.call.style;
		}
		return;
	}

	if (invoke.call instanceof CanvasLineWidth) {
		ctx.lineWidth = or0(invoke.call.width);
		return;
	}

	if (invoke.call instanceof CanvasLineCap) {
		if (invoke.call.cap) {
			ctx.lineCap = invoke.call.cap as CanvasLineCap_;
		}
		return;
	}

	if (invoke.call instanceof CanvasLineJoin) {
		if (invoke.call.join) {
			ctx.lineJoin = invoke.call.join as CanvasLineJoin_;
		}
		return;
	}

	if (invoke.call instanceof CanvasMiterLimit) {
		ctx.miterLimit = or0(invoke.call.limit);
		return;
	}

	if (invoke.call instanceof CanvasTextAlign) {
		if (invoke.call.textAlign) {
			ctx.textAlign = invoke.call.textAlign as CanvasTextAlign_;
		}
		return;
	}

	if (invoke.call instanceof CanvasTextBaseline) {
		if (invoke.call.baseline) {
			ctx.textBaseline = invoke.call.baseline as CanvasTextBaseline_;
		}
		return;
	}

	// --- Transformationen & State ---
	if (invoke.call instanceof CanvasSave) {
		ctx.save();
		return;
	}

	if (invoke.call instanceof CanvasRestore) {
		ctx.restore();
		return;
	}

	if (invoke.call instanceof CanvasTranslate) {
		ctx.translate(or0(invoke.call.x), or0(invoke.call.y));
		return;
	}

	if (invoke.call instanceof CanvasRotate) {
		ctx.rotate(or0(invoke.call.angle));
		return;
	}

	if (invoke.call instanceof CanvasScale) {
		ctx.scale(or0(invoke.call.x), or0(invoke.call.y));
		return;
	}

	if (invoke.call instanceof CanvasSetTransform) {
		ctx.setTransform(
			or0(invoke.call.a),
			or0(invoke.call.b),
			or0(invoke.call.c),
			or0(invoke.call.d),
			or0(invoke.call.e),
			or0(invoke.call.f)
		);
		return;
	}

	// --- Shadow ---
	if (invoke.call instanceof CanvasShadowOffsetX) {
		ctx.shadowOffsetX = or0(invoke.call!.offsetX);
		return;
	}

	if (invoke.call instanceof CanvasShadowOffsetY) {
		ctx.shadowOffsetY = or0(invoke.call!.offsetY);
		return;
	}

	if (invoke.call instanceof CanvasShadowColor) {
		if (invoke.call!.color) {
			ctx.shadowColor = invoke.call!.color;
		}
		return;
	}

	if (invoke.call instanceof CanvasShadowBlur) {
		ctx.shadowBlur = or0(invoke.call!.blur);
		return;
	}

	// --- Pfad-Operationen ---
	if (invoke.call instanceof CanvasBeginPath) {
		ctx.beginPath();
		return;
	}

	if (invoke.call instanceof CanvasClosePath) {
		ctx.closePath();
		return;
	}

	if (invoke.call instanceof CanvasMoveTo) {
		ctx.moveTo(or0(invoke.call.x), or0(invoke.call.y));
		return;
	}

	if (invoke.call instanceof CanvasLineTo) {
		ctx.lineTo(or0(invoke.call.x), or0(invoke.call.y));
		return;
	}

	if (invoke.call instanceof CanvasArc) {
		ctx.arc(
			or0(invoke.call.x),
			or0(invoke.call.y),
			or0(invoke.call.r),
			or0(invoke.call.start),
			or0(invoke.call.end),
			invoke.call.antiClockwise === true
		);
		return;
	}

	if (invoke.call instanceof CanvasArcTo) {
		ctx.arcTo(
			or0(invoke.call.x1),
			or0(invoke.call.y1),
			or0(invoke.call.x2),
			or0(invoke.call.y2),
			or0(invoke.call.radius)
		);
		return;
	}

	if (invoke.call instanceof CanvasBezierCurveTo) {
		ctx.bezierCurveTo(
			or0(invoke.call.cp1x),
			or0(invoke.call.cp1y),
			or0(invoke.call.cp2x),
			or0(invoke.call.cp2y),
			or0(invoke.call.x),
			or0(invoke.call.y)
		);
		return;
	}

	if (invoke.call instanceof CanvasQuadraticCurveTo) {
		ctx.quadraticCurveTo(or0(invoke.call.cpx), or0(invoke.call.cpy), or0(invoke.call.x), or0(invoke.call.y));
		return;
	}

	if (invoke.call instanceof CanvasRect) {
		ctx.rect(or0(invoke.call.x), or0(invoke.call.y), or0(invoke.call.w), or0(invoke.call.h));
		return;
	}

	// --- Füll- & Clip-Operationen ---
	if (invoke.call instanceof CanvasFill) {
		ctx.fill();
		return;
	}

	if (invoke.call instanceof CanvasClip) {
		ctx.clip();
		return;
	}

	if (invoke.call instanceof CanvasFillRect) {
		ctx.fillRect(or0(invoke.call.x), or0(invoke.call.y), or0(invoke.call.w), or0(invoke.call.h));
		return;
	}

	if (invoke.call instanceof CanvasFillText) {
		const maxWidth = invoke.call.maxWidth;
		if (maxWidth !== undefined && maxWidth > 0) {
			ctx.fillText(invoke.call.text ?? '', or0(invoke.call.x), or0(invoke.call.y), maxWidth);
		} else {
			ctx.fillText(invoke.call.text ?? '', or0(invoke.call.x), or0(invoke.call.y));
		}
		return;
	}

	// --- Stroke-Operationen ---
	if (invoke.call instanceof CanvasStrokeRect) {
		ctx.strokeRect(or0(invoke.call.x), or0(invoke.call.y), or0(invoke.call.w), or0(invoke.call.h));
		return;
	}

	if (invoke.call instanceof CanvasStrokeText) {
		const maxWidth = invoke.call.maxWidth;
		if (maxWidth !== undefined && maxWidth > 0) {
			ctx.strokeText(invoke.call.text ?? '', or0(invoke.call.x), or0(invoke.call.y), maxWidth);
		} else {
			ctx.strokeText(invoke.call.text ?? '', or0(invoke.call.x), or0(invoke.call.y));
		}
		return;
	}

	// --- Rect löschen ---
	if (invoke.call instanceof CanvasClearRect) {
		ctx.clearRect(or0(invoke.call.x), or0(invoke.call.y), or0(invoke.call.w), or0(invoke.call.h));
		return;
	}

	if (invoke.call instanceof CanvasClear) {
		ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
		return;
	}

	// --- Bild zeichnen ---
	if (invoke.call instanceof CanvasDrawImage) {
		const img = images.get(or0(invoke.call.hnd));
		if (img == undefined) {
			console.warn('CanvasDrawImage: image with handle', invoke.call.hnd, 'not found');
			return;
		}
		const sx = invoke.call.sx;
		if (sx !== undefined) {
			// 9-Argument-Form: drawImage(img, sx, sy, sw, sh, dx, dy, dw, dh)
			ctx.drawImage(
				img,
				or0(invoke.call.sx),
				or0(invoke.call.sy),
				or0(invoke.call.sw),
				or0(invoke.call.sh),
				or0(invoke.call.dx),
				or0(invoke.call.dy),
				or0(invoke.call.dw),
				or0(invoke.call.dh)
			);
		} else {
			if (invoke.call.dw == undefined) {
				ctx.drawImage(img, or0(invoke.call.dx), or0(invoke.call.dy));
			} else {
				// 5-Argument-Form: drawImage(img, dx, dy, dw, dh)
				ctx.drawImage(img, or0(invoke.call.dx), or0(invoke.call.dy), or0(invoke.call.dw), or0(invoke.call.dh));
			}
		}
		return;
	}
}

onUnmounted(() => {
	if (eventCallback) {
		ConnectionHandler.removeEventListener(eventCallback);
		eventCallback = null;
	}
});
</script>

<template>
	<!-- canvas -->
	<canvas
		:id="props.ui.id"
		ref="canvasRef"
		:width="extractPixelValue(props.ui.frame?.width)"
		:height="extractPixelValue(props.ui.frame?.height)"
		:style="frameStyles"
	></canvas>
</template>
