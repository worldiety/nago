export interface LiveTable {
	type: 'Table'
	id: number
	headers: ComponentList<LiveTableCell>
	rows: ComponentList<LiveTableRow>
}
