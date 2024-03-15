export interface LiveGridCell {
	type: 'GridCell'
	id: number
	body: PropertyComponent<LiveComponent>
	colStart: PropertyInt
	colEnd: PropertyInt
	rowStart: PropertyInt
	rowEnd: PropertyInt
	colSpan: PropertyInt
	smColSpan: PropertyInt
	mdColSpan: PropertyInt
	lgColSpan: PropertyInt
}
