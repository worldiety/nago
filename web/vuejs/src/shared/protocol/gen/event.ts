// DO NOT EDIT. Generated by ora-gen-ts

import type { Acknowledged } from '@/shared/protocol/gen/acknowledged';
import type { EventsAggregated } from '@/shared/protocol/gen/eventsAggregated';
import type { NewComponentRequested } from '@/shared/protocol/gen/newComponentRequested';
import type { ComponentInvalidated } from '@/shared/protocol/gen/componentInvalidated';
import type { ComponentInvalidationRequested } from '@/shared/protocol/gen/componentInvalidationRequested';
import type { ErrorOccurred } from '@/shared/protocol/gen/errorOccurred';
import type { ComponentDestructionRequested } from '@/shared/protocol/gen/componentDestructionRequested';
import type { ScopeDestructionRequested } from '@/shared/protocol/gen/scopeDestructionRequested';
import type { ConfigurationRequested } from '@/shared/protocol/gen/configurationRequested';
import type { ConfigurationDefined } from '@/shared/protocol/gen/configurationDefined';
import type { SetPropertyValueRequested } from '@/shared/protocol/gen/setPropertyValueRequested';
import type { FunctionCallRequested } from '@/shared/protocol/gen/functionCallRequested';


export type Event =
    | Acknowledged
    | EventsAggregated
    | NewComponentRequested
    | ComponentInvalidated
    | ComponentInvalidationRequested
    | ErrorOccurred
    | ComponentDestructionRequested
    | ScopeDestructionRequested
    | ConfigurationRequested
    | ConfigurationDefined
    | SetPropertyValueRequested
    | FunctionCallRequested
    

