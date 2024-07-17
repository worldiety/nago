import {cssLengthValue} from "@/components/shared/length";
import {Padding} from "@/shared/protocol/ora/padding";

// paddingCSS applies the padding length values. Note, that negative paddings are interpreted as negative margins,
// because negative padding values are not allowed and it seems practical to move views around for some nice effects.
export function paddingCSS(frame?: Padding): string[] {

	const styles: string[] = [];
	if (frame?.b) {
		if (frame.b.startsWith("-")) {
			styles.push(`margin-bottom:${cssLengthValue(frame.b)}`)
		} else {
			styles.push(`padding-bottom:${cssLengthValue(frame.b)}`)
		}

	}

	if (frame?.t) {
		if (frame.t.startsWith("-")) {
			styles.push(`margin-top: ${cssLengthValue(frame.t)}`)
		} else {
			styles.push(`padding-top:${cssLengthValue(frame.t)}`)
		}

	}

	if (frame?.r) {
		if (frame.r.startsWith("-")) {
			styles.push(`margin-right:${cssLengthValue(frame.r)}`)
		} else {
			styles.push(`padding-right:${cssLengthValue(frame.r)}`)
		}

	}

	if (frame?.l) {
		if (frame.l.startsWith("-")) {
			styles.push(`margin-left:${cssLengthValue(frame.l)}`)
		} else {
			styles.push(`padding-left:${cssLengthValue(frame.l)}`)
		}

	}


	return styles
}


// marginCSS is like padding but interprets all padding lengths as margin length
export function marginCSS(frame?: Padding): string[] {

	const styles: string[] = [];
	if (frame?.b) {
		styles.push(`margin-bottom:${cssLengthValue(frame.b)}`)
	}

	if (frame?.t) {
		styles.push(`margin-top: ${cssLengthValue(frame.t)}`)
	}

	if (frame?.r) {
		styles.push(`margin-right:${cssLengthValue(frame.r)}`)
	}

	if (frame?.l) {
		styles.push(`margin-left:${cssLengthValue(frame.l)}`)
	}

	return styles
}
