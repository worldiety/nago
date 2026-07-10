/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { cssLengthValue } from '@/components/shared/length';
import { Transformation } from '@/shared/proto/nprotoc_gen';

export function transformationCSS(transformation?: Transformation): string[] {
	if (!transformation) {
		return [];
	}

	const transformations: string[] = [];

	if (transformation.rotateZ) {
		transformations.push('rotateZ(' + transformation.rotateZ + 'deg)');
	}

	if (transformation.scaleX) {
		transformations.push('scaleX(' + transformation.scaleX + ')');
	}

	if (transformation.scaleY) {
		transformations.push('scaleY(' + transformation.scaleY + ')');
	}

	if (transformation.scaleZ) {
		transformations.push('scaleZ(' + transformation.scaleZ + ')');
	}

	if (transformation.translateX) {
		transformations.push('translateX(' + cssLengthValue(transformation.translateX) + ')');
	}

	if (transformation.translateY) {
		transformations.push('translateY(' + cssLengthValue(transformation.translateY) + ')');
	}

	if (transformation.translateZ) {
		transformations.push('translateZ(' + cssLengthValue(transformation.translateZ) + ')');
	}

	return [`transform: ${transformations.join(' ')}`];
}
