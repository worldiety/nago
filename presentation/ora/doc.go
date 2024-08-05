// Package ora contains the (implementation first) JSON based protocol specification.
// This is the truth for both sides, backend (e.g. Go) and frontend (e.g. VueJS) and any other future client (e.g. iOS).
// TODO refactor to github.com/worldiety/macro for sumtypes and ts generation
// This package is by definition unstable and will change over time due to protocol optimization.
// For example, it is very likely to be replaced by a generated protobuf version, so never rely in your
// application code on this package and instead always use the ui package which is kept stable.
package ora
