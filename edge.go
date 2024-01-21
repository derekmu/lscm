package lscm

type edge struct {
	halfedges [2]*halfEdge
	length    float32
}

func (e *edge) other(he *halfEdge) *halfEdge {
	if e.halfedges[0] == he {
		return e.halfedges[1]
	} else {
		return e.halfedges[0]
	}
}

type edgeKey struct {
	v1, v2 int
}
