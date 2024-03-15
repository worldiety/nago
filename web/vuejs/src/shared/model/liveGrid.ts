export interface LiveGrid {
	type: 'Grid'
	id: number
	cells: ComponentList<LiveGridCell>
	rows: PropertyInt
	columns: PropertyInt
	smColumns: PropertyInt
	mdColumns: PropertyInt
	lgColumns: PropertyInt
	gap: PropertyString
}
