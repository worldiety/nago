export interface PageConfiguration {
    id: string,
    endpoint: string,
    anchor: string,
    authenticated: boolean,
}

export interface PagesConfiguration {
    pages: PageConfiguration[],
}

export interface UiDescription {
    renderTree: UiElement,
    viewModel: any,
}

export type UiElement = TextElement | ButtonElement | GridElement;

export interface TextElement {
    type: "AttributedText" | "Text",
    value: string,
}

export interface ButtonElement {
    type: "Button",
    title: TextElement,
    onClick: UiEvent,
}

export interface CardElement {
    type: "Card",
    onClick: UiEvent,
    views: UiElement[],
}

export interface UiEvent {
    trigger: string,
    eventType: string,
    data: any,
}

export interface GridCellElement {
    type: "GridCell",
    start: number,
    end: number,
    span: number,
    views: UiElement[],
}

export interface GridElement {
    type: "Grid",
    columns: number,
    gap: number,
    cells: GridCellElement[],
}

export interface TableElement {
    type: "Table",
    rows: TableRow[],
    columnHeaders: TableColumnHeader[],
}

export interface TableRow{
    type:"TableRow",
    columns: TableCell[]
}

export interface TableColumnHeader{
    type: "TableColumnHeader",
    views: UiElement[],
}

export interface TableCell{
    type: "TableCell",
    views: UiElement[],
}

export interface InputTextElement {
    type: "InputText",
    name: string,
    value: string,
}

export interface InputFileElement {
    type: "InputFile",
    name: string,
    multiple: boolean,
    accept: string,
}