/**
 * bool2Str converts the given bool into a Go backend-string-parseable event value representation.
 */
export function bool2Str(b: boolean): string {
	return b ? "true" : "false";
}
