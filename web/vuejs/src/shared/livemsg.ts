export type LiveMessage =  Invalidation

export interface Invalidation {
    type: 'Invalidation'
    root: 'LiveComponent'
}

export type LiveComponent = LiveTextField |VBox

export interface VBox {
    type: 'VBox'
    children: ComponentList
}

export interface Divider {
    type: 'Divider'
}

export interface HBox {
    type: 'HBox'
    children: ComponentList
}

export interface ComponentList{
    type: 'componentList'
    id: number
    value: LiveComponent[]
}

export interface LiveButton{
    type: 'Button'
    id: number
    caption: PropertyString
    preIcon: PropertyString
    postIcon: PropertyString
    color: PropertyString
    action: PropertyFunc
    disabled: PropertyBool
}

export interface LiveScaffold{
    type: 'Scaffold'
    id: number
    title: PropertyString
    breadcrumbs: ComponentList // currently ever of LiveButton
    menu: ComponentList // currently always of LiveButton
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

export type Property = PropertyString | PropertyBool

export interface PropertyString {
    type: 'string'
    id: number
    name: string
    value: string
}

export interface PropertyBool{
    type: 'bool'
    id: number
    name: string
    value: boolean
}

export interface PropertyFunc{
    type: 'func'
    id: number
    name: string
    value: number
}


export interface CallServerFunc{
    type: 'callFn'
    id: number
}

export interface SetServerProperty{
    type: 'setProp'
    id: number
    value: any
}

export function invokeFunc(ws :WebSocket, action:PropertyFunc){
    if (action && action.id != 0) {
        const callSrvFun: CallServerFunc = {
            type: "callFn",
            id: action.value
        }
        ws.send(JSON.stringify(callSrvFun))
    }
}