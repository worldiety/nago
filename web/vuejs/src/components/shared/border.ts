import { colorValue } from '@/components/shared/colors';
import { cssLengthValue } from '@/components/shared/length';
import { Border } from '@/shared/proto/nprotoc_gen';

export function borderCSS(border?: Border): string[] {
	const css: string[] = [];

	if (!border) {
		return css;
	}

	// border radius
	if (!border.topLeftRadius.isZero()) {
		css.push(`border-top-left-radius: ${cssLengthValue(border.topLeftRadius.value)}`);
	}

	if (!border.topRightRadius.isZero()) {
		css.push(`border-top-right-radius: ${cssLengthValue(border.topRightRadius.value)}`);
	}

	if (!border.bottomLeftRadius.isZero()) {
		css.push(`border-bottom-left-radius: ${cssLengthValue(border.bottomLeftRadius.value)}`);
	}

	if (!border.bottomRightRadius.isZero()) {
		css.push(`border-bottom-right-radius: ${cssLengthValue(border.bottomRightRadius.value)}`);
	}

	// border color
	if (!border.topColor.isZero()) {
		css.push(`border-top-color: ${colorValue(border.topColor.value)}`);
	}

	if (!border.bottomColor.isZero()) {
		css.push(`border-bottom-color: ${colorValue(border.bottomColor.value)}`);
	}

	if (!border.leftColor.isZero()) {
		css.push(`border-left-color: ${colorValue(border.leftColor.value)}`);
	}

	if (!border.rightColor.isZero()) {
		css.push(`border-right-color: ${colorValue(border.rightColor.value)}`);
	}

	// border width
	if (!border.topWidth.isZero()) {
		css.push(`border-top-width: ${cssLengthValue(border.topWidth.value)}`);
	}

	if (!border.bottomWidth.isZero()) {
		css.push(`border-bottom-width: ${cssLengthValue(border.bottomWidth.value)}`);
	}

	if (!border.leftWidth.isZero()) {
		css.push(`border-left-width: ${cssLengthValue(border.leftWidth.value)}`);
	}

	if (!border.rightWidth.isZero()) {
		css.push(`border-right-width: ${cssLengthValue(border.rightWidth.value)}`);
	}

	// shadow
	if (!border.boxShadow.isZero()) {
		if (border.boxShadow.radius.isZero()) {
			border.boxShadow.radius.value = '10px';
		}

		if (border.boxShadow.color.isZero()) {
			border.boxShadow.color.value = '#00000020';
		}

		if (border.boxShadow.x.isZero()) {
			border.boxShadow.x.value = '0dp';
		}

		if (border.boxShadow.y.isZero()) {
			border.boxShadow.y.value = '0dp';
		}

		css.push(
			`box-shadow: ${cssLengthValue(border.boxShadow.x.value)} ${cssLengthValue(border.boxShadow.y.value)} ${cssLengthValue(border.boxShadow.radius.value)} 0 ${border.boxShadow.color.value}`
		);
	}

	return css;
}
