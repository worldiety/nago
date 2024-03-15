export interface PageConfiguration {
	id: string;
	link: string;
	anchor: string;
	authenticated: boolean;
}

export interface LivePageConfiguration {
	id: string;
	link: string;
	anchor: string;
	authenticated: boolean;
}

export interface PagesConfiguration {
	name: string;
	pages: PageConfiguration[];
	index: string;
	oidc: OIDCProvider[]
	livePages: LivePageConfiguration[]
}

export interface OIDCProvider {
	name: string
	authority: string
	clientID: string
	clientSecret: string
	redirectURL: string
	postLogoutRedirectUri: string
}

export interface UiDescription {
	renderTree: UiElement;
	viewModel: any;
	redirect: Redirection | null;
}

export interface Redirection {
	type: 'Redirect';
	url: string;
	direction: 'forward' | 'backward';
	redirect: boolean;
}

export type UiElement =
	TextElement
	| ButtonElement
	| GridElement
	| Scaffold
	| ListView
	| FormField
	| CardView
	| LiveComponent
	| SVGElement;

export interface CardView {
	type: 'CardView'
	cards: Card[]
}

export interface Card {
	type: 'Card'
	title: string
	subtitle: string
	content: any
	prependIcon: FontIcon
	appendIcon: FontIcon
	actions: Button[]
	primaryAction: Action

}

export interface Button {
	type: 'Button'
	caption: string
	action: Action
}

export type Action = Redirect

export interface Redirect {
	type: 'Redirect'
	target: string
}

export interface CardMetricText {
	type: 'CardMetricText';
	value: string;
	icon: Image | null
}

export interface TextElement {
	type: 'AttributedText' | 'Text';
	value: string;
}

export interface SVGElement {
	type: 'SVG';
	svg: string;
	maxWidth: string;
}


export interface TimelineElement {
	type: 'Timeline';
	items: TimelineItem[];
}

export interface TimelineItem {
	type: 'TimelineItem'
	icon: Image
	color: string | null
	title: string
	alternateDotText: string | null
	target: string
}


export interface Scaffold {
	type: 'Scaffold';
	children: URL[];
	title: string;
	navigation: NavItem[];
	breadcrumbs: Breadcrumb[]
}

export interface Breadcrumb {
	title: string
	href: string
	disabled: boolean
}

export interface NavItem {
	title: string;
	link: NavAction;
	anchor: string,
	icon: Image;
}

export type Image = FontIcon;

export interface FontIcon {
	type: 'FontIcon';
	name: string;
	color: string;
}

export type NavAction = Navigation;

export interface Navigation {
	type: 'Navigation';
	target: string;
	payload: any;
}

export type Persona = ListView;

export interface ListView {
	type: 'ListView';
	links: LVLinks;
}

export interface LVLinks {
	list: URL | null
	delete: URL | null
}

export interface FormField {
	type: 'TextField' | 'FileUploadField' | 'SelectField'
	label: string
	id: string
	value: string | null
	hint: string
	error: string
	disabled: boolean
	fileMultiple: boolean | null
	fileAccept: string | null
	selectMultiple: boolean | null
	selectItems: SelectItem[]
	selectValues: string[]
}

export interface SelectItem {
	type: 'SelectItem'
	id: string
	caption: string
}

export interface Form {
	type: 'Form'
	submitText: string
	deleteText: string
	links: FormLinks
}

export interface FormLinks {
	load: URL | null
	submit: URL | null
	delete: URL | null
}

export interface ListViewList {
	data: ListItemModel[];
}

export interface ListItemModel {
	type: 'ListItem';
	id: string;
	title: string;
	action: NavAction;
}

export interface ButtonElement {
	type: 'Button';
	title: TextElement;
	onClick: UiEvent;
}

export interface CardElement {
	type: 'Card';
	onClick: UiEvent;
	views: UiElement[];
}

export interface UiEvent {
	trigger: string;
	eventType: string;
	data: any;
}

export interface GridCellElement {
	type: 'GridCell';
	colSpan: number;
	rowSpan: number;
	views: UiElement[];
}

export interface GridElement {
	type: 'Grid';
	columns: number;
	rows: number;
	gap: number;
	padding: string;
	cells: GridCellElement[];
}

export interface NavbarElement {
	type: 'Navbar';
	caption: UiElement;
	menuItems: UiElement[];
}

export interface TableElement {
	type: 'Table'
	links: TableLinks;
}

export interface TableLinks {
	list: URL | null
	delete: URL | null
}

export interface TableListResponse {
	rows: any[];
	headers: TableHeader[];
}

export interface TableHeader {
	title: string
	align: string
	key: string
}


export interface TableColumnHeader {
	type: 'TableColumnHeader';
	views: UiElement[];
}

export interface TableCell {
	type: 'TableCell';
	views: UiElement[];
}

export interface InputTextElement {
	type: 'InputText';
	name: string;
	value: string;
	label: string;
}


export interface InputFileElement {
	type: 'InputFile';
	name: string;
	multiple: boolean;
	accept: string;
}

export type LiveMessage = Invalidation

export interface Invalidation {
	type: 'Invalidation' | 'HistoryPushState' | 'HistoryBack'
	root: LiveComponent
	modals: ComponentList<LiveComponent>
	token: string
}


export type LiveComponent =
	LiveTextField
	| VBox
	| LiveTable
	| LiveDropdown
	| LiveDropdownItem
	| LiveButton
	| LiveTableCell
	| LiveTableRow
	| LiveGridCell
	| LiveGrid
	| LiveDialog
	| LiveToggle
	| LiveStepper
	| LiveStepInfo
	| LiveTextArea
	| LiveChip
	| LivePage
	| LiveImage


export interface LivePage {
	type: 'Page'
	root: LiveComponent
	modals: ComponentList<LiveComponent>
	token: string
}

export interface LiveChip {
	type: 'Chip'
	caption: PropertyString
	action: PropertyFunc
	onClose: PropertyFunc
	color: PropertyString
}

export interface VBox {
	type: 'VBox'
	children: ComponentList<LiveComponent>
}

export interface LiveCard {
	type: 'Card'
	children: ComponentList<LiveComponent>
	action: PropertyFunc
}


export interface LiveImage {
	type: 'Image'
	url: PropertyString
	downloadToken: PropertyString
	caption: PropertyString
}


export interface Divider {
	type: 'Divider'
}

export interface HBox {
	type: 'HBox'
	children: ComponentList<LiveComponent>
	alignment: PropertyString
}

export interface ComponentList<T extends LiveComponent> {
	type: 'componentList'
	id: number
	value: T[]
}

export interface LiveStepper {
	type: 'Stepper'
	id: number
	steps: ComponentList<LiveStepInfo>
	selectedIndex: PropertyInt
}

export interface LiveStepInfo {
	type: 'StepInfo'
	id: number
	number: PropertyString
	caption: PropertyString
	details: PropertyString
}

export interface LiveDropdown {
	type: 'Dropdown',
	id: number,
	items: ComponentList<LiveDropdownItem>,
	selectedIndex: PropertyInt,
	expanded: PropertyBool,
	disabled: PropertyBool,
	onToggleExpanded: PropertyFunc,
}

export interface LiveDropdownItem {
	type: 'DropdownItem',
	itemIndex: PropertyInt,
	content: PropertyString,
	onSelected: PropertyFunc,
}

export interface LiveButton {
	type: 'Button'
	id: number
	caption: PropertyString
	preIcon: PropertyString
	postIcon: PropertyString
	color: PropertyString
	action: PropertyFunc
	disabled: PropertyBool
}

export interface LiveDialog {
	type: 'Dialog'
	id: number
	title: PropertyString
	body: PropertyComponent<LiveComponent>
	icon: PropertyString
	actions: ComponentList<LiveButton>
}

export interface LiveToggle {
	type: 'Toggle'
	id: number
	label: PropertyString
	checked: PropertyBool
	disabled: PropertyBool
	onCheckedChanged: PropertyFunc
}

export interface LiveScaffold {
	type: 'Scaffold'
	id: number
	title: PropertyString
	breadcrumbs: ComponentList<LiveButton> // currently ever of LiveButton
	menu: ComponentList<LiveButton> // currently always of LiveButton
	body: PropertyComponent<LiveComponent>
	topbarLeft: PropertyComponent<LiveComponent>
	topbarMid: PropertyComponent<LiveComponent>
	topbarRight: PropertyComponent<LiveComponent>
}

export interface LiveTextField {
	type: 'TextField'
	id: number
	label: PropertyString
	hint: PropertyString
	error: PropertyString
	value: PropertyString
	disabled: PropertyBool
	onTextChanged: PropertyFunc
}

export interface LiveUploadField {
	type: 'FileField'
	id: number
	label: PropertyString
	hint: PropertyString
	error: PropertyString
	disabled: PropertyBool
	filter: PropertyString
	multiple: PropertyBool
	uploadToken: PropertyString
}

export interface LiveTextArea {
	type: 'TextArea'
	id: number
	label: PropertyString
	hint: PropertyString
	error: PropertyString
	value: PropertyString
	rows: PropertyInt
	disabled: PropertyBool
	onTextChanged: PropertyFunc
}

export interface LiveScaffold {
	type: 'Scaffold'
	id: number

}

export interface LiveText {
	type: 'Text'
	id: number
	value: PropertyString
	color: PropertyString
	colorDark: PropertyString
	size: PropertyString
	onClick: PropertyFunc
	onHoverStart: PropertyFunc
	onHoverEnd: PropertyFunc
}

export interface LiveTable {
	type: 'Table'
	id: number
	headers: ComponentList<LiveTableCell>
	rows: ComponentList<LiveTableRow>
}

export interface LiveTableCell {
	type: 'TableCell'
	id: number
	body: PropertyComponent<LiveComponent>
}

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

export interface LiveTableRow {
	type: 'TableRow'
	id: number
	cells: ComponentList<LiveTableCell>
}


export type Property = PropertyString | PropertyBool | PropertyInt

export interface PropertyString {
	id: number
	name: string
	value: string
}

export interface PropertyBool {
	id: number
	name: string
	value: boolean
}

export interface PropertyInt {
	id: number
	name: string
	value: number
}

export interface PropertyFunc {
	id: number
	name: string
	value: number
}

export interface PropertyComponent<T extends LiveComponent> {
	id: number
	name: string
	value: T
}


export interface CallBatch {
	tx: (CallServerFunc | SetServerProperty | UpdateJWT | ClientHello) []
}

export interface CallServerFunc {
	type: 'callFn'
	id: number
}

export interface SetServerProperty {
	type: 'setProp'
	id: number
	value: any
}

export interface UpdateJWT {
	type: 'updateJWT'
	token: string
	OIDCName: 'Keycloak'
}

export interface ClientHello {
	type: 'hello'
	auth: ClientHelloAuth
}

export interface ClientHelloAuth {
	keycloak: string
}
