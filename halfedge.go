package lscm

type halfEdge struct {
	edge         *edge
	vertex       *vertex
	prev         *halfEdge
	next         *halfEdge
	coefficients point3D
}

func (e *halfEdge) other() *halfEdge {
	return e.edge.other(e)
}

func (e *halfEdge) source() *vertex {
	return e.prev.vertex
}

func (e *halfEdge) target() *vertex {
	return e.vertex
}
