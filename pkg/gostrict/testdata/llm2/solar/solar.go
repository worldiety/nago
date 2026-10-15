// Package solar shows a small solar system with three.js: a sun, planets on elliptic orbits and their moons.
// Clicking a planet selects it, the wheel zooms the camera and the backend changes the speed or adds planets.
package solar

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/lib/asset"
	"go.wdy.de/nago/pkg/gostrict/lib/input"
	"go.wdy.de/nago/pkg/gostrict/lib/three"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
	"go.wdy.de/nago/pkg/gostrict/std/text"
	"go.wdy.de/nago/pkg/gostrict/testdata/llm2/solar/proto"
)

const (
	sunURL       = "/sun.jpg"
	sunRadius    = 2.0
	maxPlanets   = 12
	orbitSteps   = 128
	minDistance  = 8.0
	maxDistance  = 120.0
	elevation    = 0.5 // angle of the camera above the plane of the orbits in radians
	highlight    = 0xffff66
	paletteCount = 8
)

var palette = list.Freeze(list.Of[uint32](
	0x9e9e9e, 0xe0b060, 0x3a7bd5, 0xc1440e, 0xd8ca9d, 0xe3c16f, 0x7fd3e0, 0x4b70dd,
))

// Moon circles its planet.
type Moon struct {
	Mesh  three.Mesh
	Dist  float64
	Speed float64
	Phase float64
}

// Planet moves on an ellipse around the sun, which is in a focus of the ellipse.
type Planet struct {
	Name  string
	Group three.Group // at the position of the planet, holds the planet and its moons
	Mesh  three.Mesh
	Mat   three.Material
	Orbit three.Line
	A     float64 // semi-major axis
	B     float64 // semi-minor axis
	Speed float64 // angular speed in radians per second
	Phase float64
	Spin  float64
	Moons list.List[Moon]
}

// System is the state of the program.
type System struct {
	Scene    three.Scene
	Camera   three.Camera
	Sun      three.Mesh
	SunMat   three.Material
	SunTex   three.Texture
	Planets  list.List[*Planet]
	Selected string
	Speed    float64
	SimTime  float64
	LastT    float64
	Started  bool
	Distance float64
}

// Init creates the scene with the sun and the given number of planets.
func Init(planets int32) *System {
	scene := three.NewScene()
	three.SetBackground(scene, 0x05060a)
	camera := three.NewCamera(50, 1, 0.1, 1000)

	sunMat := three.NewMaterial(0xffcc33, 0, 1)
	three.SetEmissive(sunMat, 0xffaa00, 1)
	sun := three.NewMesh(three.NewSphere(sunRadius, 48), sunMat)
	three.Add(scene, sun)
	three.Add(scene, three.NewPointLight(0xffffff, 3, 0))
	three.Add(scene, three.NewAmbientLight(0x404050, 0.4))

	s := &System{
		Scene:    scene,
		Camera:   camera,
		Sun:      sun,
		SunMat:   sunMat,
		SunTex:   three.LoadTexture(sunURL),
		Planets:  list.New[*Planet](),
		Selected: "",
		Speed:    1,
		SimTime:  0,
		LastT:    0,
		Started:  false,
		Distance: 40,
	}

	n := planets
	if n < 0 {
		n = 0
	}
	if n > maxPlanets {
		n = maxPlanets
	}
	for i := range n {
		fi := float64(i)
		dist := 5 + fi*3.2
		radius := 0.3 + 0.15*float64(i%3)
		ecc := 0.05 + 0.04*float64(i%4)
		name := "Planet " + text.FromInt(i+1)
		moons := i % 3
		addPlanet(s, name, radius, dist, ecc, colorOr(palette.Get(i%paletteCount)), moons, fi*1.3)
	}

	placeCamera(s)
	three.Show(scene, camera)
	return s
}

func colorOr(c uint32, ok bool) uint32 {
	if !ok {
		return 0xffffff
	}
	return c
}

// addPlanet creates a planet with its orbit and moons and adds it to the scene.
func addPlanet(s *System, name string, radius, dist, ecc float64, color uint32, moons int32, phase float64) {
	if radius < 0.05 {
		radius = 0.05
	}
	if dist < sunRadius+radius+0.5 {
		dist = sunRadius + radius + 0.5
	}
	if ecc < 0 {
		ecc = 0
	}
	if ecc > 0.9 {
		ecc = 0.9
	}
	a := dist
	b := a * math.Sqrt(1-ecc*ecc)
	focus := a * ecc

	// the orbit as a closed line through the points of the ellipse
	geo := three.NewBufferGeometry()
	for k := range int32(orbitSteps + 1) {
		angle := 2 * math.Pi * float64(k) / orbitSteps
		three.AddVertex(geo, a*math.Cos(angle)-focus, 0, b*math.Sin(angle))
	}
	three.UpdateGeometry(geo)
	orbit := three.NewLine(geo, three.NewLineMaterial(0x334455))
	three.Add(s.Scene, orbit)

	mat := three.NewMaterial(color, 0.1, 0.8)
	mesh := three.NewMesh(three.NewSphere(radius, 32), mat)
	three.SetName(mesh, name)
	group := three.NewGroup()
	three.Add(group, mesh)
	three.Add(s.Scene, group)

	p := &Planet{
		Name:  name,
		Group: group,
		Mesh:  mesh,
		Mat:   mat,
		Orbit: orbit,
		A:     a,
		B:     b,
		Speed: 6 / (a * math.Sqrt(a)),
		Phase: phase,
		Spin:  1 + radius,
		Moons: list.New[Moon](),
	}
	for m := range moons {
		fm := float64(m)
		moonMesh := three.NewMesh(three.NewSphere(radius*0.3, 16), three.NewMaterial(0xcccccc, 0, 0.9))
		three.Add(group, moonMesh)
		p.Moons.Append(Moon{
			Mesh:  moonMesh,
			Dist:  radius*2 + 0.4 + fm*0.5,
			Speed: 2.5 - fm*0.7,
			Phase: fm * 2.1,
		})
	}
	s.Planets.Append(p)
	placePlanet(p, s.SimTime)
}

// placePlanet moves the planet and its moons to their positions at the simulation time.
func placePlanet(p *Planet, t float64) {
	angle := p.Phase + t*p.Speed
	focus := math.Sqrt(math.Max(0, p.A*p.A-p.B*p.B))
	three.SetPosition(p.Group, p.A*math.Cos(angle)-focus, 0, p.B*math.Sin(angle))
	three.SetRotation(p.Mesh, 0, t*p.Spin, 0)
	for m := range p.Moons.Values() {
		ma := m.Phase + t*m.Speed
		three.SetPosition(m.Mesh, m.Dist*math.Cos(ma), 0, m.Dist*math.Sin(ma))
	}
}

func placeCamera(s *System) {
	three.SetPosition(s.Camera, 0, s.Distance*math.Sin(elevation), s.Distance*math.Cos(elevation))
	three.LookAt(s.Camera, 0, 0, 0)
}

// Resize adapts the camera to the aspect ratio of the canvas.
func Resize(s *System, aspect float64) {
	if aspect <= 0 {
		return
	}
	three.SetAspect(s.Camera, aspect)
}

// Frame advances the simulation to the time t in seconds.
func Frame(s *System, t float64) {
	dt := float64(0)
	if s.Started {
		dt = t - s.LastT
	}
	s.Started = true
	s.LastT = t
	if dt < 0 {
		dt = 0
	}
	if dt > 0.1 {
		dt = 0.1
	}
	s.SimTime += dt * s.Speed
	three.SetRotation(s.Sun, 0, s.SimTime*0.1, 0)
	for p := range s.Planets.Values() {
		placePlanet(p, s.SimTime)
	}
}

// OnInput selects a planet with the pointer and zooms with the wheel.
func OnInput(s *System, e input.Event) msg.Out[proto.Selected] {
	out := msg.New[proto.Selected]()
	switch e.Type {
	case input.PointerDown:
		name, _, ok := three.Pick(s.Scene, s.Camera, e.X, e.Y)
		if !ok {
			return out
		}
		if selectPlanet(s, name) {
			out.Send(proto.Selected{Name: name})
		}
	case input.MouseWheel:
		s.Distance *= math.Exp(e.DeltaY * 0.001)
		if s.Distance < minDistance {
			s.Distance = minDistance
		}
		if s.Distance > maxDistance {
			s.Distance = maxDistance
		}
		placeCamera(s)
	default:
	}
	return out
}

// selectPlanet highlights the planet of the name and reports, whether there is one.
func selectPlanet(s *System, name string) bool {
	found := false
	for p := range s.Planets.Values() {
		if p.Name == name {
			three.SetEmissive(p.Mat, highlight, 0.6)
			found = true
		} else {
			three.SetEmissive(p.Mat, 0x000000, 0)
		}
	}
	if found {
		s.Selected = name
	}
	return found
}

// OnLoaded applies the texture of the sun, when it has been loaded.
func OnLoaded(s *System, e asset.Loaded) {
	if e.URL == sunURL {
		three.SetColor(s.SunMat, 0xffffff)
		three.SetMap(s.SunMat, s.SunTex)
	}
}

// OnSpeed changes the speed of the simulation.
func OnSpeed(s *System, m proto.Speed) {
	f := m.Factor
	if f < 0 {
		f = 0
	}
	if f > 100 {
		f = 100
	}
	s.Speed = f
}

// OnAddPlanet adds a planet at runtime.
func OnAddPlanet(s *System, m proto.AddPlanet) {
	moons := int32(0)
	if m.Radius > 0.5 {
		moons = 1
	}
	phase := float64(s.Planets.Len()) * 1.3
	addPlanet(s, m.Name, m.Radius, m.Distance, 0.1, m.Color, moons, phase)
}
