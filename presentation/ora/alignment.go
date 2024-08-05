package ora

// Alignment is specified as follows:
//
//	┌─TopLeading───────────Top─────────TopTrailing─┐
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│ Leading            Center            Trailing│
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	│                                              │
//	└BottomLeading───────Bottom──────BottomTrailing┘
//
// An empty Alignment must be interpreted as Center (="c").
//
// #[go.TypeScript "path":"web/vuejs/src/shared/protocol/ora"]
type Alignment string
