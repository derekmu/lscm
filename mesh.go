package lscm

type Mesh struct {
	vertices []*Vertex
	faces    []*Face
	edges    []*Edge
	edgeMap  map[EdgeKey]*Edge
}

func NewMesh() *Mesh {
	return &Mesh{
		edgeMap: make(map[EdgeKey]*Edge, 128),
	}
}

func (m *Mesh) AddVertex(p Point3D, fixed bool, uv Point2D) *Vertex {
	id := len(m.vertices)
	v := &Vertex{
		id:    id,
		point: p,
		uv:    uv,
		fixed: fixed,
	}
	m.vertices = append(m.vertices, v)
	return v
}

func (m *Mesh) AddFace(vis [3]int) *Face {
	face := &Face{}
	m.faces = append(m.faces, face)
	// create halfedges
	halfedges := [3]*HalfEdge{}
	for i, vi := range vis {
		v := m.vertices[vi]
		he := &HalfEdge{vertex: v}
		v.halfedge = he
		halfedges[i] = he
	}
	// link to each other
	for i := 0; i < 3; i++ {
		halfedges[i].next = halfedges[(i+1)%3]
		halfedges[i].prev = halfedges[(i+2)%3]
	}
	// link to face
	face.halfedge = halfedges[0]
	// link to edges
	for i, vi := range vis {
		edge := m.addEdge(vi, vis[(i+2)%3])
		if edge.halfedges[0] == nil {
			edge.halfedges[0] = halfedges[i]
		} else {
			edge.halfedges[1] = halfedges[i]
		}
		halfedges[i].edge = edge
	}
	return face
}

func (m *Mesh) addEdge(vi1, vi2 int) *Edge {
	key := EdgeKey{min(vi1, vi2), max(vi1, vi2)}
	if edge, ok := m.edgeMap[key]; ok {
		return edge
	}
	edge := &Edge{
		halfedges: [2]*HalfEdge{},
	}
	m.edges = append(m.edges, edge)
	m.edgeMap[key] = edge
	return edge
}

func (m *Mesh) RemoveDanglingVertices() {
	for i := 0; i < len(m.vertices); i++ {
		v := m.vertices[i]
		if v.halfedge == nil {
			m.vertices[i] = m.vertices[len(m.vertices)-1]
			m.vertices = m.vertices[:len(m.vertices)-1]
			i--
		}
		m.vertices[i].id = i
	}
}

func (m *Mesh) UpdateBoundary() {
	for _, edge := range m.edges {
		if edge.halfedges[1] == nil {
			// make boundary vertex halfedges to be the most counter-clockwise
			edge.halfedges[0].vertex.rotateCcwAboutTarget()
			edge.halfedges[0].prev.vertex.rotateCcwAboutTarget()
		}
	}
}
