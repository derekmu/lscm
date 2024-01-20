package lscm

type Mesh struct {
	edges     []*Edge
	faces     []*Face
	vertices  []*Vertex
	vertexMap map[int]*Vertex
	edgeMap   map[EdgeKey]*Edge
}

func NewMesh() *Mesh {
	return &Mesh{
		vertexMap: make(map[int]*Vertex, 128),
		edgeMap:   make(map[EdgeKey]*Edge, 128),
	}
}

func (m *Mesh) createVertex(id int, p Point3D) *Vertex {
	v := &Vertex{
		id:    id,
		point: p,
	}
	m.vertices = append(m.vertices, v)
	m.vertexMap[id] = v
	return v
}

func (m *Mesh) createFace(vertices [3]*Vertex) *Face {
	face := &Face{}
	m.faces = append(m.faces, face)

	// create halfedges
	halfedges := [3]*HalfEdge{}
	for i := 0; i < 3; i++ {
		vertex := vertices[i]
		he := &HalfEdge{vertex: vertex}
		vertex.halfedge = he
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
	for i := 0; i < 3; i++ {
		edge := m.createEdge(vertices[i], vertices[(i+2)%3])
		if edge.halfedges[0] == nil {
			edge.halfedges[0] = halfedges[i]
		} else {
			edge.halfedges[1] = halfedges[i]
		}
		halfedges[i].edge = edge
	}
	return face
}

func (m *Mesh) createEdge(vertex1 *Vertex, vertex2 *Vertex) *Edge {
	key := EdgeKey{min(vertex1.id, vertex2.id), max(vertex1.id, vertex2.id)}
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
