package lscm

type Edge struct {
	halfedges [2]*HalfEdge
	length    float64
}

func (e *Edge) other(he *HalfEdge) *HalfEdge {
	if e.halfedges[0] == he {
		return e.halfedges[1]
	} else {
		return e.halfedges[0]
	}
}

func (e *Edge) updateLength() {
	v1 := e.halfedges[0].source()
	v2 := e.halfedges[0].target()
	vd := v1.point.sub(&v2.point)
	e.length = vd.norm()
}

type EdgeKey struct {
	v1, v2 int
}
