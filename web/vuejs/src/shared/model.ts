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

export interface InputTextElement {
    type: "InputText",
    name: string,
    value: string,
}