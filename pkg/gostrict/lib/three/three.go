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
// called with, and implements show(scene, camera), e.g. by rendering the scene in each animation frame. The host
// also provides the size of the canvas for [Pick], turns the shadows of the renderer on, see [Shadows], and loads
// textures and models, see [LoadTexture] and [LoadModel], whose results it reports with the events of package
// lib/asset.
//
// A program calls [Show] once, e.g. in its Init, and changes the scene afterwards, which the host renders after
// each frame. It finds the objects of the scene by their names, e.g. the one, which the user has clicked, see
// [Pick] and [SetName], so that no handle is ever missing. Pick uses the camera, so the program keeps its aspect
// ratio up to date with [SetAspect] in its Resize function. The URL of the events of lib/asset is the string,
// which the program passed to [LoadTexture] or [LoadModel].
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
type Texture struct{ ref js.Ref }

// Line draws the segments between the vertices of a geometry.
type Line struct{ ref js.Ref }

// Object is any handle, which has a position in a scene.
type Object interface {
	Mesh | Group | Light | Camera | Line
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

//gostrict:js (c, aspect) => { if (c.isOrthographicCamera) { const h = (c.top - c.bottom) / 2; c.left = -h * aspect; c.right = h * aspect; } else { c.aspect = aspect; } c.updateProjectionMatrix(); }
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
	return newMaterial(color)
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

// NewOrthoCamera returns an orthographic camera, which shows the height in units of the scene at any distance,
// e.g. for a map or a 2D game. SetAspect adapts its width.
//
//gostrict:js (height, aspect, near, far) => new THREE.OrthographicCamera(-height / 2 * aspect, height / 2 * aspect, height / 2, -height / 2, near, far)
func NewOrthoCamera(height, aspect, near, far float64) Camera {
	n := newNode("OrthoCamera")
	nodeOf(n).aspect = aspect
	return Camera{n}
}

// Remove takes the child out of its parent.
//
//gostrict:js (parent, child) => { parent.remove(child); }
func Remove[P Container, T Object](parent P, child T) {
	p := nodeOf(refOf(parent))
	c := nodeOf(refOf(child))
	for i, n := range p.children {
		if n == c {
			p.children = append(p.children[:i], p.children[i+1:]...)
			return
		}
	}
}

// SetVisible shows or hides the object and its children.
//
//gostrict:js (o, v) => { o.visible = v; }
func SetVisible[T Object](o T, visible bool) {
	nodeOf(refOf(o)).hidden = !visible
}

// SetName names the object, e.g. to find it with [Pick].
//
//gostrict:js (o, name) => { o.name = name; }
func SetName[T Object](o T, name string) {
	nodeOf(refOf(o)).name = name
}

// NewPlane returns a rectangle in the xy plane.
//
//gostrict:js (w, h) => new THREE.PlaneGeometry(w, h)
func NewPlane(w, h float64) Geometry {
	return Geometry{newNode("Plane")}
}

// NewCylinder returns a cylinder along the y axis, whose radii at the top and the bottom may differ.
//
//gostrict:js (top, bottom, h, segments) => new THREE.CylinderGeometry(top, bottom, h, Math.max(segments, 3))
func NewCylinder(radiusTop, radiusBottom, height float64, segments int32) Geometry {
	return Geometry{newNode("Cylinder")}
}

// NewCone returns a cone along the y axis.
//
//gostrict:js (r, h, segments) => new THREE.ConeGeometry(r, h, Math.max(segments, 3))
func NewCone(radius, height float64, segments int32) Geometry {
	return Geometry{newNode("Cone")}
}

// NewTorus returns a ring with the radius and the radius of its tube.
//
//gostrict:js (r, tube) => new THREE.TorusGeometry(r, tube, 24, 64)
func NewTorus(radius, tube float64) Geometry {
	return Geometry{newNode("Torus")}
}

// NewBufferGeometry returns a geometry without vertices. [AddVertex] adds them and [UpdateGeometry] uploads them,
// e.g. the points of a line or the corners of triangles.
//
//gostrict:js () => { const g = new THREE.BufferGeometry(); g.userData.points = []; return g; }
func NewBufferGeometry() Geometry {
	return Geometry{newNode("BufferGeometry")}
}

// AddVertex adds a vertex to a geometry of [NewBufferGeometry].
//
//gostrict:js (g, x, y, z) => { if (g.userData.points) g.userData.points.push(x, y, z); }
func AddVertex(g Geometry, x, y, z float64) {
	n := nodeOf(g.ref)
	n.vertices = append(n.vertices, [3]float64{x, y, z})
}

// ClearVertices removes the vertices of a geometry of [NewBufferGeometry].
//
//gostrict:js (g) => { if (g.userData.points) g.userData.points.length = 0; }
func ClearVertices(g Geometry) {
	nodeOf(g.ref).vertices = nil
}

// UpdateGeometry uploads the vertices of a geometry of [NewBufferGeometry], which form triangles for a mesh or
// segments for a line.
//
//gostrict:js (g) => { if (!g.userData.points) return; g.setAttribute('position', new THREE.Float32BufferAttribute(g.userData.points, 3)); g.computeVertexNormals(); g.computeBoundingSphere(); }
func UpdateGeometry(g Geometry) {
}

// NewLine returns a line through the vertices of the geometry in their order, a strip like THREE.Line: a closed
// loop repeats its first vertex at the end.
//
//gostrict:js (g, m) => new THREE.Line(g, m)
func NewLine(g Geometry, m Material) Line {
	n := newNode("Line")
	nodeOf(n).geometry = nodeOf(g.ref)
	nodeOf(n).material = nodeOf(m.ref)
	return Line{n}
}

// NewBasicMaterial returns a material, which ignores the lights.
//
//gostrict:js (color) => new THREE.MeshBasicMaterial({ color })
func NewBasicMaterial(color uint32) Material {
	return newMaterial(color)
}

// NewPhongMaterial returns a shiny material.
//
//gostrict:js (color, shininess) => new THREE.MeshPhongMaterial({ color, shininess })
func NewPhongMaterial(color uint32, shininess float64) Material {
	return newMaterial(color)
}

// NewPhysicalMaterial returns a material like [NewMaterial] with a clear coat from 0 to 1, e.g. for lacquer.
//
//gostrict:js (color, metalness, roughness, clearcoat) => new THREE.MeshPhysicalMaterial({ color, metalness, roughness, clearcoat })
func NewPhysicalMaterial(color uint32, metalness, roughness, clearcoat float64) Material {
	return newMaterial(color)
}

// NewLineMaterial returns the material of a line.
//
//gostrict:js (color) => new THREE.LineBasicMaterial({ color })
func NewLineMaterial(color uint32) Material {
	return newMaterial(color)
}

// SetOpacity sets the opacity of the material from 0 to 1.
//
//gostrict:js (m, o) => { m.opacity = o; m.transparent = o < 1; }
func SetOpacity(m Material, opacity float64) {
	nodeOf(m.ref).opacity = opacity
}

// SetWireframe draws only the edges of the triangles.
//
//gostrict:js (m, on) => { m.wireframe = on; }
func SetWireframe(m Material, on bool) {
}

// SetEmissive lets the material glow in the color with the intensity, if it supports it. SetEmissive(m, 0, 0)
// removes the glow.
//
//gostrict:js (m, color, intensity) => { if (m.emissive) { m.emissive.setHex(color); m.emissiveIntensity = intensity; } }
func SetEmissive(m Material, color uint32, intensity float64) {
}

// SetMap covers the material with the texture.
//
//gostrict:js (m, t) => { m.map = t; m.needsUpdate = true; }
func SetMap(m Material, t Texture) {
}

// LoadTexture loads the image at the URL of the application, e.g. a resource or an upload, as a texture, which is
// empty until the host reports asset.Loaded.
//
//gostrict:js (url) => $host.loadTexture(url)
func LoadTexture(url string) Texture {
	return Texture{newNode("Texture " + url)}
}

// SetRepeat repeats the texture x times horizontally and y times vertically.
//
//gostrict:js (t, x, y) => { t.wrapS = THREE.RepeatWrapping; t.wrapT = THREE.RepeatWrapping; t.repeat.set(x, y); }
func SetRepeat(t Texture, x, y float64) {
}

// LoadModel loads the glTF model at the URL of the application into a group, which is empty until the host
// reports asset.Loaded.
//
//gostrict:js (url) => $host.loadModel(url)
func LoadModel(url string) Group {
	return Group{newNode("Model " + url)}
}

// DisposeGeometry releases the memory of the geometry on the GPU, after it was removed from the scene.
//
//gostrict:js (g) => { g.dispose(); }
func DisposeGeometry(g Geometry) {
}

// DisposeMaterial releases the memory of the material on the GPU, after it was removed from the scene.
//
//gostrict:js (m) => { m.dispose(); }
func DisposeMaterial(m Material) {
}

// DisposeTexture releases the memory of the texture on the GPU.
//
//gostrict:js (t) => { t.dispose(); }
func DisposeTexture(t Texture) {
}

// NewPointLight returns a light, which shines from a point in all directions up to the distance, or without
// limit, if it is 0.
//
//gostrict:js (color, intensity, distance) => new THREE.PointLight(color, intensity, distance)
func NewPointLight(color uint32, intensity, distance float64) Light {
	return Light{newNode("PointLight")}
}

// NewSpotLight returns a light, which shines in a cone with the angle in radians.
//
//gostrict:js (color, intensity, distance, angle) => new THREE.SpotLight(color, intensity, distance, angle)
func NewSpotLight(color uint32, intensity, distance, angle float64) Light {
	return Light{newNode("SpotLight")}
}

// Shadows turns the shadows of the renderer on or off, see [SetCastShadow] and [SetReceiveShadow].
//
//gostrict:js (on) => { $host.shadows(on); }
func Shadows(on bool) {
}

// SetCastShadow lets a mesh or a light cast shadows.
//
//gostrict:js (o, on) => { o.castShadow = on; }
func SetCastShadow[T Mesh | Light](o T, on bool) {
}

// SetReceiveShadow lets a mesh receive shadows.
//
//gostrict:js (m, on) => { m.receiveShadow = on; }
func SetReceiveShadow(m Mesh, on bool) {
}

// Pick returns the name and the distance of the nearest named object at the point (x, y) of the canvas in CSS
// pixels, e.g. of an input event, or false, if there is none. Objects without a name are skipped, but not their
// named parents, so a part of a model finds the model.
//
//gostrict:js (s, c, x, y) => { const rc = new THREE.Raycaster(); rc.setFromCamera(new THREE.Vector2(x / $host.width * 2 - 1, 1 - y / $host.height * 2), c); for (const h of rc.intersectObjects(s.children, true)) { for (let o = h.object; o; o = o.parent) { if (o.name !== '') return [o.name, h.distance, true]; } } return ['', 0, false]; }
func Pick(s Scene, c Camera, x, y float64) (string, float64, bool) {
	return "", 0, false
}
