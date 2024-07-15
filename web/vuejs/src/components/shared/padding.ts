import {cssLengthValue} from "@/components/shared/length";
import {Padding} from "@/shared/protocol/ora/padding";


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
