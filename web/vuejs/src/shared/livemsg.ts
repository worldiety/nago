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

export interface ComponentList{
    type: 'componentList'
    id: bigint
    value: LiveComponent[]
}

export interface LiveButton{
    type: 'Button'
    id: bigint
    caption: PropertyString
    preIcon: PropertyString
    postIcon: PropertyString
    color: PropertyString
    action: PropertyFunc
    disabled: PropertyBool
}

export interface LiveTextField {
    type: 'TextField'
    id: bigint
    label: PropertyString
    hint: PropertyString
    error: PropertyString
    value: PropertyString
    disabled: PropertyBool
    onTextChanged: PropertyFunc
}

export interface LiveScaffold {
    type: 'Scaffold'
    id: bigint

}

export type Property = PropertyString | PropertyBool

export interface PropertyString {
    type: 'string'
    id: bigint
    name: string
    value: string
}

export interface PropertyBool{
    type: 'bool'
    id: bigint
    name: string
    value: boolean
}

export interface PropertyFunc{
    type: 'func'
    id: bigint
    name: string
    value: bigint
}


export interface CallServerFunc{
    type: 'callFn'
    id: bigint
}

export interface SetServerProperty{
    type: 'setProp'
    id: bigint
    value: any
}