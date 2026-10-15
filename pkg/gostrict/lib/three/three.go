// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package three binds a part of three.js to the gostrict subset. Its handles are the objects of three.js itself, so
// a call costs a JavaScript call and nothing else, and the functions, which work on several kinds of objects, are
// generic over a union of handles, which the compiler checks. The native implementation is a small scene graph,
// which runs a program on a server or in a test, see [Dump].
//
// The program hands its scene to the host with [Show]. In a browser, the host is the object, which the bundle is
// called with, and implements show(scene, camera), e.g. by rendering the scene in each animation frame.
//
//gostrict:header
package three

import "go.wdy.de/nago/pkg/gostrict/std/js"

type Scene struct{ ref js.Ref }
type Camera struct{ ref js.Ref }
type Group struct{ ref js.Ref }
type Mesh struct{ ref js.Ref }
type Light struct{ ref js.Ref }
type Geometry struct{ ref js.Ref }
type Material struct{ ref js.Ref }

// Object is any handle, which has a position in a scene.
type Object interface {
	Mesh | Group | Light | Camera
}

// Container is any handle, which holds objects.
type Container interface {
	Scene | Group
}

//gostrict:js () => new THREE.Scene()
func NewScene() Scene {
	return Scene{newNode("Scene")}
}

//gostrict:js (s, color) => { s.background = new THREE.Color(color); }
func SetBackground(s Scene, color uint32) {
	nodeOf(s.ref).color = color
}

//gostrict:js (fov, aspect, near, far) => new THREE.PerspectiveCamera(fov, aspect, near, far)
func NewCamera(fov, aspect, near, far float64) Camera {
	n := newNode("Camera")
	nodeOf(n).aspect = aspect
	return Camera{n}
}

//gostrict:js (c, aspect) => { c.aspect = aspect; c.updateProjectionMatrix(); }
func SetAspect(c Camera, aspect float64) {
	nodeOf(c.ref).aspect = aspect
}

//gostrict:js (w, h, d) => new THREE.BoxGeometry(w, h, d)
func NewBox(w, h, d float64) Geometry {
	return Geometry{newNode("Box")}
}

//gostrict:js (radius, segments) => new THREE.SphereGeometry(radius, segments, segments)
func NewSphere(radius float64, segments int32) Geometry {
	return Geometry{newNode("Sphere")}
}

//gostrict:js (radius, tube) => new THREE.TorusKnotGeometry(radius, tube, 128, 16)
func NewTorusKnot(radius, tube float64) Geometry {
	return Geometry{newNode("TorusKnot")}
}

//gostrict:js (color, metalness, roughness) => new THREE.MeshStandardMaterial({ color, metalness, roughness })
func NewMaterial(color uint32, metalness, roughness float64) Material {
	n := newNode("Material")
	nodeOf(n).color = color
	return Material{n}
}

//gostrict:js (m, color) => { m.color.setHex(color); }
func SetColor(m Material, color uint32) {
	nodeOf(m.ref).color = color
}

//gostrict:js (g, m) => new THREE.Mesh(g, m)
func NewMesh(g Geometry, m Material) Mesh {
	n := newNode("Mesh")
	nodeOf(n).geometry = nodeOf(g.ref)
	nodeOf(n).material = nodeOf(m.ref)
	return Mesh{n}
}

//gostrict:js () => new THREE.Group()
func NewGroup() Group {
	return Group{newNode("Group")}
}

//gostrict:js (color, intensity) => new THREE.DirectionalLight(color, intensity)
func NewDirectionalLight(color uint32, intensity float64) Light {
	return Light{newNode("DirectionalLight")}
}

//gostrict:js (color, intensity) => new THREE.AmbientLight(color, intensity)
func NewAmbientLight(color uint32, intensity float64) Light {
	return Light{newNode("AmbientLight")}
}

//gostrict:js (parent, child) => { parent.add(child); }
func Add[P Container, T Object](parent P, child T) {
	p := nodeOf(refOf(parent))
	p.children = append(p.children, nodeOf(refOf(child)))
}

//gostrict:js (o, x, y, z) => { o.position.set(x, y, z); }
func SetPosition[T Object](o T, x, y, z float64) {
	nodeOf(refOf(o)).pos = [3]float64{x, y, z}
}

//gostrict:js (o, x, y, z) => { o.rotation.set(x, y, z); }
func SetRotation[T Object](o T, x, y, z float64) {
	nodeOf(refOf(o)).rot = [3]float64{x, y, z}
}

//gostrict:js (o, s) => { o.scale.setScalar(s); }
func SetScale[T Object](o T, s float64) {
	nodeOf(refOf(o)).scale = s
}

//gostrict:js (o, x, y, z) => { o.lookAt(x, y, z); }
func LookAt[T Object](o T, x, y, z float64) {
	nodeOf(refOf(o)).target = [3]float64{x, y, z}
}

//gostrict:js (s, c) => { $host.show(s, c); }
func Show(s Scene, c Camera) {
	shown = nodeOf(s.ref)
}
