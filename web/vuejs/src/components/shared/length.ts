import {Length} from "@/shared/protocol/ora/length";

export function cssLengthValue(l?: Length): string {
	if (!l || l === "") {
		return ""
	}

	// px is just wrong in CSS, they always mean dp
	l = l.replaceAll("dp", "px")

	if (l.charAt(0)==='-' || l.charAt(0) >= '0' && l.charAt(0) <= '9') {
		return l
	}

	if (l.startsWith("calc")){
		return l
	}

	return `var(--${l})`
}

export function cssLengthValue0Px(l?: Length): string {
	if (!l) {
		return "0px"
	}

	l = l.replaceAll("dp", "px")
	return l
}
