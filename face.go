package lscm

type Face struct {
	id       int
	halfedge *HalfEdge
	touched  bool
}
