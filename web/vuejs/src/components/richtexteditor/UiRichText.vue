<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import sanitizeHtml from 'sanitize-html';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { routeAnchorClick } from '@/eventhandling';
import { RichText } from '@/shared/proto/nprotoc_gen';
import VHtml from '@/components/VHtml.vue';

const props = defineProps<{
	ui: RichText;
}>();

const serviceAdapter = useServiceAdapter();

// security note: an attacker may inject malicious html into our model and the server or backend developer
// forgot to check or remove it.
// Thus, let us sanitize the input.

const sanitizedHtml = computed<string>(() => {
	if (!props.ui.value) {
		return '';
	}

	return sanitizeHtml(props.ui.value, {
		transformTags: {
			// an external link opens in a new tab, which gets no access to this one, see sanitizeMarkup
			a: (tagName, attribs) => {
				if (externalURL(attribs.href) && !attribs.target) {
					return { tagName, attribs: { ...attribs, target: '_blank' } };
				}

				return { tagName, attribs };
			},
		},
	});
});

const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);
	if (!styles) {
		return '';
	}

	return styles.join(';');
});

/**
 * externalURL returns the address of a link to another origin, which the browser would load, otherwise undefined.
 */
function externalURL(href: string | undefined): URL | undefined {
	if (!href) {
		return undefined;
	}

	try {
		const url = new URL(href, window.location.href);
		if ((url.protocol === 'http:' || url.protocol === 'https:') && url.origin !== window.location.origin) {
			return url;
		}
	} catch {
		// not an address at all
	}

	return undefined;
}

/**
 * onLinkClick handles the clicks on the links of the text. A link of the own origin navigates within the
 * application instead of loading the page anew. An external link opens in a new tab, after the user confirmed its
 * full address, if the backend asks for it, whatever button or modifier has been used.
 */
function onLinkClick(event: MouseEvent): void {
	const a = (event.target as Element | null)?.closest('a');
	if (!a || !(event.currentTarget as Element).contains(a)) {
		return;
	}

	const url = externalURL(a.getAttribute('href') ?? undefined);
	if (!url) {
		if (event.type === 'click') {
			routeAnchorClick(serviceAdapter, event, a);
		}

		return;
	}

	const question = props.ui.externalLinkConfirmation;
	if (!question || event.defaultPrevented || (event.type === 'auxclick' && event.button !== 1)) {
		return;
	}

	event.preventDefault();
	if (window.confirm(`${question}\n\n${url.href}`)) {
		window.open(url.href, '_blank', 'noopener,noreferrer');
	}
}
</script>

<template>
	<VHtml
		tag="div"
		:style="frameStyles"
		class="prose-custom"
		:html="sanitizedHtml"
		@click="onLinkClick"
		@auxclick="onLinkClick"
	/>
</template>
