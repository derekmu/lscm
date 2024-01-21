package lscm

type Vertex struct {
	point    Point3D
	normal   Point3D
	uv       Point2D
	halfedge *HalfEdge

	id    int // mesh index of the vertex
	index int // unfixed/fixed index for LSCM
	fixed bool
}

func (v *Vertex) rotateCcwAboutTarget() {
	for v.halfedge.other() != nil {
		v.halfedge = v.halfedge.other().prev
	}
}
