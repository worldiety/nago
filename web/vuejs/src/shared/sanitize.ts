/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import DOMPurify from 'dompurify';

// Links which open a new tab must not get access to this window.
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
	if (node.tagName === 'A' && node.getAttribute('target') === '_blank') {
		node.setAttribute('rel', 'noopener noreferrer');
	}
});

/**
 * Removes everything active from HTML or SVG markup, which may come from users, e.g. a signature, an uploaded
 * image or rich text: event handlers, scripts, foreignObject and javascript: URLs. Shapes, gradients, filters,
 * styles and currentColor are kept. DOMPurify parses within an inert document, so nothing runs while sanitizing.
 */
export function sanitizeMarkup(markup: string): string {
	return DOMPurify.sanitize(markup ?? '', {
		USE_PROFILES: { html: true, svg: true, svgFilters: true },
		ADD_ATTR: ['target'],
	});
}
