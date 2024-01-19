package lscm

type HalfEdge struct {
	edge         *Edge
	vertex       *Vertex
	prev         *HalfEdge
	next         *HalfEdge
	coefficients Point3D
}

func (e *HalfEdge) other() *HalfEdge {
	return e.edge.other(e)
}

func (e *HalfEdge) source() *Vertex {
	return e.prev.vertex
}

func (e *HalfEdge) target() *Vertex {
	return e.vertex
}
