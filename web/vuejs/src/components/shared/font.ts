import {Font} from "@/shared/protocol/ora/font";

// see also https://developer.mozilla.org/en-US/docs/Web/CSS/font
export function fontCSS(font?: Font): string[] {
	const styles: string[] = [];
	if (!font) {
		return styles
	}


	// style and weight must precede size
	switch (font.t) {
		case "n":
			styles.push("font-style: normal")
			break
		case "i":
			styles.push("font-style: italic")
			break
	}

	if (font.w) {
		styles.push(`font-weight: ${font.w}`)
	}


	if (font.n) {
		styles.push(`font-family: ${font.n}`)
	}

	if (font.s) {
		styles.push(`font-size: ${font.s}`)
	}

	return styles
}
