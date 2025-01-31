// see also https://developer.mozilla.org/en-US/docs/Web/CSS/font
import {Font, FontStyleValues} from "@/shared/proto/nprotoc_gen";

export function fontCSS(font?: Font): string[] {
	const styles: string[] = [];
	if (!font) {
		return styles;
	}

	// style and weight must precede size
	switch (font.style.value) {
		case FontStyleValues.Normal:
			styles.push('font-style: normal');
			break;
		case FontStyleValues.Italic:
			styles.push('font-style: italic');
			break;
	}

	if (!font.weight.isZero()) {
		styles.push(`font-weight: ${font.weight.value}`);
	}

	if (!font.name.isZero()) {
		styles.push(`font-family: ${font.name.value}`);
	}

	if (!font.size.isZero()) {
		styles.push(`font-size: ${font.size.value}`);
	}

	return styles;
}
