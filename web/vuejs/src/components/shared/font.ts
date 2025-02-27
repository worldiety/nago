// see also https://developer.mozilla.org/en-US/docs/Web/CSS/font
import { Font, FontStyleValues } from '@/shared/proto/nprotoc_gen';

export function fontCSS(font?: Font): string[] {
	const styles: string[] = [];
	if (!font) {
		return styles;
	}

	// style and weight must precede size
	switch (font.style) {
		case FontStyleValues.Normal:
			styles.push('font-style: normal');
			break;
		case FontStyleValues.Italic:
			styles.push('font-style: italic');
			break;
	}

	if (font.weight) {
		styles.push(`font-weight: ${font.weight}`);
	}

	if (font.name) {
		styles.push(`font-family: ${font.name}`);
	}

	if (font.size) {
		styles.push(`font-size: ${font.size}`);
	}

	return styles;
}
