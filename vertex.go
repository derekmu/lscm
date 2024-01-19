package lscm

type Vertex struct {
	id       int
	point    Point3D
	normal   Point3D
	uv       Point2D
	halfedge *HalfEdge
	boundary bool
	fixed    bool
	index    int
	valence  int
	father   *Vertex
	touched  bool
}
