/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Acknowledged } from '@/shared/protocol/ora/acknowledged';
import type { ComponentDestructionRequested } from '@/shared/protocol/ora/componentDestructionRequested';
import type { ComponentInvalidated } from '@/shared/protocol/ora/componentInvalidated';
import type { ComponentInvalidationRequested } from '@/shared/protocol/ora/componentInvalidationRequested';
import type { ConfigurationDefined } from '@/shared/protocol/ora/configurationDefined';
import type { ConfigurationRequested } from '@/shared/protocol/ora/configurationRequested';
import type { ErrorOccurred } from '@/shared/protocol/ora/errorOccurred';
import type { EventsAggregated } from '@/shared/protocol/ora/eventsAggregated';
import type { FunctionCallRequested } from '@/shared/protocol/ora/functionCallRequested';
import type { NavigationBackRequested } from '@/shared/protocol/ora/navigationBackRequested';
import type { NavigationForwardToRequested } from '@/shared/protocol/ora/navigationForwardToRequested';
import type { NavigationReloadRequested } from '@/shared/protocol/ora/navigationReloadRequested';
import type { NavigationResetRequested } from '@/shared/protocol/ora/navigationResetRequested';
import type { NewComponentRequested } from '@/shared/protocol/ora/newComponentRequested';
import type { Ping } from '@/shared/protocol/ora/ping';
import type { ScopeDestructionRequested } from '@/shared/protocol/ora/scopeDestructionRequested';
import type { SendMultipleRequested } from '@/shared/protocol/ora/sendMultipleRequested';
import type { SessionAssigned } from '@/shared/protocol/ora/sessionAssigned';
import type { SetPropertyValueRequested } from '@/shared/protocol/ora/setPropertyValueRequested';

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
| NavigationForwardToRequested
| NavigationReloadRequested
| NavigationResetRequested
| NavigationBackRequested
| SessionAssigned
| Ping
| SendMultipleRequested
