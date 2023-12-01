export type LiveMessage = Invalidation

export interface Invalidation {
    type: 'Invalidation'
    root: 'LiveComponent'
}

export type LiveComponent = LiveTextField | VBox | LiveTable | LiveButton | LiveTableCell | LiveTableRow

export interface VBox {
    type: 'VBox'
    children: ComponentList
}

export interface Divider {
    type: 'Divider'
}

export interface HBox {
    type: 'HBox'
    children: ComponentList<LiveComponent>
}

export interface ComponentList<T extends LiveComponent> {
    type: 'componentList'
    id: number
    value: T[]
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

export interface LiveScaffold {
    type: 'Scaffold'
    id: number
    title: PropertyString
    breadcrumbs: ComponentList<LiveButton> // currently ever of LiveButton
    menu: ComponentList<LiveComponent> // currently always of LiveButton
    body: LiveComponent
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
    if (action && action.id != 0) {
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