import {Frame} from "@/shared/protocol/ora/frame";
import {Border} from "@/shared/protocol/ora/border";
import {cssLengthValue} from "@/components/shared/length";

export function borderCSS(border?: Border): string[] {
	const css: string[] = [];

	if (!border){
		return css
	}

	// border radius
	if (border.tlr){
		css.push(`border-top-left-radius: ${cssLengthValue(border.tlr)}`)
	}

	if (border.trr){
		css.push(`border-top-right-radius: ${cssLengthValue(border.trr)}`)
	}

	if (border.blr){
		css.push(`border-bottom-left-radius: ${cssLengthValue(border.blr)}`)
	}

	if (border.brr){
		css.push(`border-bottom-right-radius: ${cssLengthValue(border.brr)}`)
	}

	// border color
	if (border.tc){
		css.push(`border-top-color: ${border.tc}`)
	}

	if (border.bc){
		css.push(`border-bottom-color: ${border.bc}`)
	}

	if (border.lc){
		css.push(`border-left-color: ${border.lc}`)
	}

	if (border.rc){
		css.push(`border-right-color: ${border.rc}`)
	}

	// border width
	if (border.tw){
		css.push(`border-top-width: ${cssLengthValue(border.tw)}`)
	}

	if (border.bw){
		css.push(`border-bottom-width: ${cssLengthValue(border.bw)}`)
	}

	if (border.lw){
		css.push(`border-left-width: ${cssLengthValue(border.lw)}`)
	}

	if (border.rw){
		css.push(`border-right-width: ${cssLengthValue(border.rw)}`)
	}

	// shadow
	if (border.s){
		if (!border.s.r){
			border.s.r="10px"
		}

		if (!border.s.c){
			border.s.c="#00000020"
		}

		if (!border.s.x){
			border.s.x="0dp"
		}

		if (!border.s.y){
			border.s.y="0dp"
		}


		css.push(`box-shadow: ${cssLengthValue(border.s.x)} ${cssLengthValue(border.s.y)} ${cssLengthValue(border.s.r)} 0 ${border.s.c}`)
	}

	console.log(css)

	return css
}
