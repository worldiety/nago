export type LiveMessage = Invalidation

export interface Invalidation {
    type: 'Invalidation' | 'HistoryPushState' | 'HistoryBack'
    root: LiveComponent
    modals: ComponentList<LiveComponent>
}

export type LiveComponent =
    LiveTextField
    | VBox
    | LiveTable
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


export type Property = PropertyString | PropertyBool

export interface PropertyString {
    type: 'string'
    id: number
    name: string
    value: string
}

export interface PropertyBool {
    type: 'bool'
    id: number
    name: string
    value: boolean
}

export interface PropertyInt {
    type: 'int'
    id: number
    name: string
    value: number
}

export interface PropertyFunc {
    type: 'func'
    id: number
    name: string
    value: number
}

export interface PropertyComponent<T extends LiveComponent> {
    type: string
    id: number
    name: string
    value: T
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

export function invokeFunc(ws: WebSocket, action: PropertyFunc) {
    if (action && action.id != 0 && action.value != 0) {
        const callSrvFun: CallServerFunc = {
            type: "callFn",
            id: action.value
        }
        ws.send(JSON.stringify(callSrvFun))

    }
}

export function textColor2Tailwind(s: string): string {
    if (s == null || s == "") {
        return ""
    }

    if (s.startsWith('#')) {
        return "text-[" + s + "]"
    }

    return s
}

export function textSize2Tailwind(s: string): string {
    if (s == null || s == "") {
        return ""
    }

    if (s.endsWith('px') || s.endsWith('pt') || s.endsWith('rem')) {
        return "text-[" + s + "]"
    }

    return s
}

export function gapSize2Tailwind(s: string): string {
    if (s == null || s == "") {
        return ""
    }

    if (s.endsWith('px') || s.endsWith('pt') || s.endsWith('rem')) {
        return "gap-[" + s + "]"
    }

    return s
}