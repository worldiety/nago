export function isNil(v: any): boolean {
	if (v == undefined) {
		return true
	}

	return v === 0
}
