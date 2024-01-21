package lscm

import "errors"

type Mesh struct {
	points  []float32
	uvs     []float32
	normals []float32

	vertices []*vertex
	faces    []*face
	edges    []*edge
	edgeMap  map[edgeKey]*edge
}

/*
NewMesh creates a Mesh object from the buffers provided. The buffers provided may be modified by LSCM.

Parameters:
  - points: slice of vertex coordinates, every 3 elements representing a vertex
  - uvs: slice of texture coordinates, every 2 elements representing a vertex, the values will be overwritten by LSCM
  - normals: may be nil, it is only included for convenience and not used by LSCM
  - indices: slice of vertex indices for each face, every 3 elements representing a face
  - fixedIndices: slice of vertex indices for fixed uvs for LSCM, at least 2 are required

If the wrong number of points, uvs, indices, or fixedIndices is detected, an error is returned.
*/
func NewMesh(points []float32, uvs []float32, normals []float32, indices []uint32, fixedIndices []uint32) (*Mesh, error) {
	if len(points)%3 != 0 {
		return nil, errors.New("the number of points must be divisible by 3")
	}
	if len(uvs)%2 != 0 {
		return nil, errors.New("the number of uvs must be divisible by 2")
	}
	if len(points)/3 != len(uvs)/2 {
		return nil, errors.New("there must be 2 uv coordinates for every 3 vertex coordinates")
	}
	if len(indices)%3 != 0 {
		return nil, errors.New("the number of indices must be divisible by 3")
	}
	if len(fixedIndices) < 2 {
		return nil, errors.New("the number of fixed indices must be at least 2")
	}
	vertexCount := len(points) / 3
	faceCount := len(indices) / 3
	edgeCount := vertexCount + faceCount
	m := &Mesh{
		points:   points,
		uvs:      uvs,
		normals:  normals,
		vertices: make([]*vertex, 0, vertexCount),
		faces:    make([]*face, 0, faceCount),
		edges:    make([]*edge, 0, edgeCount),
		edgeMap:  make(map[edgeKey]*edge, edgeCount),
	}
	for i := 0; i < vertexCount; i++ {
		v := &vertex{id: i}
		m.vertices = append(m.vertices, v)
	}
	for i := 0; i < len(indices); i += 3 {
		m.addFace([3]int{int(indices[i]), int(indices[i+1]), int(indices[i+2])})
	}
	for _, i := range fixedIndices {
		m.vertices[i].fixed = true
	}
	return m, nil
}

func (m *Mesh) getPoint(vi int) point3D {
	return point3D{
		x: m.points[vi*3],
		y: m.points[vi*3+1],
		z: m.points[vi*3+2],
	}
}

func (m *Mesh) setPoint(vi int, p point3D) {
	m.points[vi*3] = p.x
	m.points[vi*3+1] = p.y
	m.points[vi*3+2] = p.z
}

func (m *Mesh) getNormal(vi int) point3D {
	return point3D{
		x: m.normals[vi*3],
		y: m.normals[vi*3+1],
		z: m.normals[vi*3+2],
	}
}

func (m *Mesh) setNormal(vi int, n point3D) {
	m.normals[vi*3] = n.x
	m.normals[vi*3+1] = n.y
	m.normals[vi*3+2] = n.z
}

func (m *Mesh) getUV(vi int) point2D {
	return point2D{
		x: m.uvs[vi*2],
		y: m.uvs[vi*2+1],
	}
}

func (m *Mesh) setUV(vi int, uv point2D) {
	m.uvs[vi*2] = uv.x
	m.uvs[vi*2+1] = uv.y
}

func (m *Mesh) GetPoints() []float32 {
	return m.points
}

func (m *Mesh) GetNormals() []float32 {
	return m.normals
}

func (m *Mesh) GetUVs() []float32 {
	return m.uvs
}

func (m *Mesh) addFace(vis [3]int) *face {
	f := &face{}
	m.faces = append(m.faces, f)
	// create halfedges
	halfedges := [3]*halfEdge{}
	for i, vi := range vis {
		v := m.vertices[vi]
		he := &halfEdge{vertex: v}
		v.halfedge = he
		halfedges[i] = he
	}
	// link to each other
	for i := 0; i < 3; i++ {
		halfedges[i].next = halfedges[(i+1)%3]
		halfedges[i].prev = halfedges[(i+2)%3]
	}
	// link to face
	f.halfedge = halfedges[0]
	// link to edges
	for i, vi := range vis {
		e := m.addEdge(vi, vis[(i+2)%3])
		if e.halfedges[0] == nil {
			e.halfedges[0] = halfedges[i]
		} else {
			e.halfedges[1] = halfedges[i]
		}
		halfedges[i].edge = e
	}
	return f
}

func (m *Mesh) addEdge(vi1, vi2 int) *edge {
	key := edgeKey{min(vi1, vi2), max(vi1, vi2)}
	if e, ok := m.edgeMap[key]; ok {
		return e
	}
	e := &edge{
		halfedges: [2]*halfEdge{},
	}
	m.edges = append(m.edges, e)
	m.edgeMap[key] = e
	return e
}

func (m *Mesh) removeDanglingVertices() {
	for i := 0; i < len(m.vertices); i++ {
		v := m.vertices[i]
		if v.halfedge == nil {
			endi := len(m.vertices) - 1
			m.vertices[i] = m.vertices[endi]
			m.vertices = m.vertices[:endi]
			m.setPoint(i, m.getPoint(endi))
			m.points = m.points[:endi*3]
			m.setNormal(i, m.getNormal(endi))
			m.normals = m.normals[:endi*3]
			m.setUV(i, m.getUV(endi))
			m.uvs = m.uvs[:endi*2]
			i--
		} else {
			m.vertices[i].id = i
		}
	}
}

func (m *Mesh) updateBoundary() {
	for _, e := range m.edges {
		if e.halfedges[1] == nil {
			// make boundary vertex halfedges to be the most counter-clockwise
			e.halfedges[0].vertex.rotateCcwAboutTarget()
			e.halfedges[0].prev.vertex.rotateCcwAboutTarget()
		}
	}
}
