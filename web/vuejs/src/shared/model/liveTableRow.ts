export interface LiveTableRow {
	type: 'TableRow'
	id: number
	cells: ComponentList<LiveTableCell>
}
