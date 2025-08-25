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
	const styles: string[] = [];

	if (!transformation) {
		return styles;
	}

	if (transformation.rotateZ) {
		styles.push('transform: rotateZ(' + transformation.rotateZ + 'deg)');
	}

	if (transformation.scaleX) {
		styles.push('transform: scaleX(' + transformation.scaleX + ')');
	}

	if (transformation.scaleY) {
		styles.push('transform: scaleY(' + transformation.scaleY + ')');
	}

	if (transformation.scaleZ) {
		styles.push('transform: scaleZ(' + transformation.scaleZ + ')');
	}

	if (transformation.translateX) {
		styles.push('transform: translateX(' + cssLengthValue(transformation.translateX) + ')');
	}

	if (transformation.translateY) {
		styles.push('transform: translateY(' + cssLengthValue(transformation.translateY) + ')');
	}

	if (transformation.translateZ) {
		styles.push('transform: translateZ(' + cssLengthValue(transformation.translateZ) + ')');
	}

	return styles;
}
