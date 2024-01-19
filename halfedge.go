package lscm

type HalfEdge struct {
	edge         *Edge
	face         *Face
	vertex       *Vertex
	prev         *HalfEdge
	next         *HalfEdge
	angle        float64
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
