export interface PageConfiguration {
    id: string;
    link: string;
    anchor: string;
    authenticated: boolean;
}

export interface PagesConfiguration {
    name: string;
    pages: PageConfiguration[];
    index: string;
    oidc:OIDCProvider[]

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
