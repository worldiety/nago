// Package proto declares the messages between the solar system program and its backend.
package proto

// Selected tells the backend, that the user has clicked a planet.
type Selected struct {
	Name string
}

// Speed sets the factor of the simulation speed, 1 is normal.
type Speed struct {
	Factor float64
}

// AddPlanet adds a planet at runtime.
type AddPlanet struct {
	Name     string
	Radius   float64
	Distance float64
	Color    uint32
}
