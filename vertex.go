package lscm

type vertex struct {
	id       int // mesh index of the vertex
	index    int // unfixed/fixed index for LSCM
	fixed    bool
	halfedge *halfEdge
}

func (v *vertex) rotateCcwAboutTarget() {
	for v.halfedge.other() != nil {
		v.halfedge = v.halfedge.other().prev
	}
}
