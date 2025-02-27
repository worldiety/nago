import { colorValue } from '@/components/shared/colors';
import { cssLengthValue } from '@/components/shared/length';
import { Border } from '@/shared/proto/nprotoc_gen';

export function borderCSS(border?: Border): string[] {
	const css: string[] = [];

	if (!border) {
		return css;
	}

	// border radius
	if (border.topLeftRadius) {
		css.push(`border-top-left-radius: ${cssLengthValue(border.topLeftRadius)}`);
	}

	if (border.topRightRadius) {
		css.push(`border-top-right-radius: ${cssLengthValue(border.topRightRadius)}`);
	}

	if (border.bottomLeftRadius) {
		css.push(`border-bottom-left-radius: ${cssLengthValue(border.bottomLeftRadius)}`);
	}

	if (border.bottomRightRadius) {
		css.push(`border-bottom-right-radius: ${cssLengthValue(border.bottomRightRadius)}`);
	}

	// border color
	if (border.topColor) {
		css.push(`border-top-color: ${colorValue(border.topColor)}`);
	}

	if (border.bottomColor) {
		css.push(`border-bottom-color: ${colorValue(border.bottomColor)}`);
	}

	if (border.leftColor) {
		css.push(`border-left-color: ${colorValue(border.leftColor)}`);
	}

	if (border.rightColor) {
		css.push(`border-right-color: ${colorValue(border.rightColor)}`);
	}

	// border width
	if (border.topWidth) {
		css.push(`border-top-width: ${cssLengthValue(border.topWidth)}`);
	}

	if (border.bottomWidth) {
		css.push(`border-bottom-width: ${cssLengthValue(border.bottomWidth)}`);
	}

	if (border.leftWidth) {
		css.push(`border-left-width: ${cssLengthValue(border.leftWidth)}`);
	}

	if (border.rightWidth) {
		css.push(`border-right-width: ${cssLengthValue(border.rightWidth)}`);
	}

	// shadow
	if (!border.boxShadow.isZero()) {
		if (!border.boxShadow.radius) {
			border.boxShadow.radius = '10px';
		}

		if (!border.boxShadow.color) {
			border.boxShadow.color = '#00000020';
		}

		if (!border.boxShadow.x) {
			border.boxShadow.x = '0dp';
		}

		if (!border.boxShadow.y) {
			border.boxShadow.y = '0dp';
		}

		css.push(
			`box-shadow: ${cssLengthValue(border.boxShadow.x)} ${cssLengthValue(border.boxShadow.y)} ${cssLengthValue(border.boxShadow.radius)} 0 ${border.boxShadow.color}`
		);
	}

	return css;
}
