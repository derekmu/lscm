package lscm

import (
	"errors"
	"gonum.org/v1/gonum/mat"
	"math"
)

func RunLSCM(mesh *Mesh) error {
	mesh.RemoveDanglingVertices()
	mesh.UpdateBoundary()

	// set coefficients
	for _, edge := range mesh.edges {
		edge.updateLength()
	}
	for _, face := range mesh.faces {
		hel := [3]float64{}
		he := face.halfedge
		for i := 0; i < 3; i++ {
			hel[i] = he.edge.length
			he = he.next
		}
		// law of cosines
		a := math.Acos((hel[0]*hel[0] + hel[2]*hel[2] - hel[1]*hel[1]) / (2 * hel[0] * hel[2]))
		p := [3]Point3D{
			{0, 0, 0},
			{hel[0], 0, 0},
			{hel[2] * math.Cos(a), hel[2] * math.Sin(a), 0},
		}
		n0 := p[1].sub(&p[0])
		n1 := p[2].sub(&p[0])
		n := n0.cross(&n1)
		area := n.norm() / 2.0
		n.divide(area)
		he = face.halfedge
		for i := 0; i < 3; i++ {
			np := p[(i+1)%3].sub(&p[i])
			s := n.cross(&np)
			s.divide(math.Sqrt(area))
			he.coefficients = s
			he = he.next
		}
	}

	// divide vertices into fixed and unfixed
	vertices := make([]*Vertex, 0, len(mesh.vertices))
	fixedVertices := make([]*Vertex, 0, 2)
	for _, vertex := range mesh.vertices {
		if vertex.fixed {
			fixedVertices = append(fixedVertices, vertex)
		} else {
			vertices = append(vertices, vertex)
		}
	}
	for i, vertex := range vertices {
		vertex.index = i
	}
	for i, vertex := range fixedVertices {
		vertex.index = i
	}
	if len(fixedVertices) < 2 {
		return errors.New("at least two fixed vertices are required")
	}

	// prepare matrices for least squares
	fn := len(mesh.faces)
	vfn := len(fixedVertices)
	vn := len(vertices)
	amat := mat.NewDense(2*fn, 2*vn, nil)
	bmat := mat.NewDense(2*fn, 2*vfn, nil)
	fmat := mat.NewVecDense(2*vfn, nil)
	for fid, face := range mesh.faces {
		he := face.halfedge
		for i := 0; i < 3; i++ {
			v := he.next.target()
			vid := v.index
			if !v.fixed {
				amat.Set(fid, vid, he.coefficients.X)
				amat.Set(fn+fid, vn+vid, he.coefficients.X)
				amat.Set(fid, vn+vid, -he.coefficients.Y)
				amat.Set(fn+fid, vid, he.coefficients.Y)
			} else {
				bmat.Set(fid, vid, he.coefficients.X)
				bmat.Set(fn+fid, vfn+vid, he.coefficients.X)
				bmat.Set(fid, vfn+vid, -he.coefficients.Y)
				bmat.Set(fn+fid, vid, he.coefficients.Y)
				fmat.SetVec(vid, v.uv.X)
				fmat.SetVec(vfn+vid, v.uv.Y)
			}
			he = he.next
		}
	}
	rmat := mat.NewVecDense(2*fn, nil)
	rmat.MulVec(bmat, fmat)
	rmat.ScaleVec(-1, rmat)

	// solve least squares
	smat := mat.NewDense(2*vn, 1, nil)
	err := smat.Solve(amat, rmat)
	if err != nil {
		return err
	}

	// read UVs out to vertices
	uvMin := Point2D{}
	uvMax := Point2D{}
	for i, v := range vertices {
		v.uv = Point2D{
			X: smat.At(i, 0),
			Y: smat.At(i+vn, 0),
		}
		uvMin.X = min(uvMin.X, v.uv.X)
		uvMin.Y = min(uvMin.Y, v.uv.Y)
		uvMax.X = max(uvMax.X, v.uv.X)
		uvMax.Y = max(uvMax.Y, v.uv.Y)
	}
	// scale UVs to be within the range [0:1]
	for _, v := range mesh.vertices {
		v.uv = Point2D{
			X: (v.uv.X - uvMin.X) / (uvMax.X - uvMin.X),
			Y: (v.uv.Y - uvMin.Y) / (uvMax.Y - uvMin.Y),
		}
	}

	return nil
}
