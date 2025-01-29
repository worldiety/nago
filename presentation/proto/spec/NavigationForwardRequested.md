NavigationForwardToRequested is an Event triggered by the backend which requests a forward navigation action within the frontend.
A frontend must put the new component to create by the factory on top of the current component within the scope.
The frontend is free keep multiple components alive at the same time, however it must ensure that the UX is sane.