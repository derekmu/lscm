package lscm

import (
	"errors"
	"gonum.org/v1/gonum/mat"
	"math"
)

func RunLSCM(mesh *Mesh) error {
	mesh.updateBoundary()

	// divide vertices into fixed and unfixed
	vertices := make([]*vertex, 0, len(mesh.vertices))
	fixedVertices := make([]*vertex, 0, 2)
	for _, v := range mesh.vertices {
		if v.halfedge == nil {
			// ignore dangling vertices
		} else if v.fixed {
			fixedVertices = append(fixedVertices, v)
		} else {
			vertices = append(vertices, v)
		}
	}
	for i, v := range vertices {
		v.index = i
	}
	for i, v := range fixedVertices {
		v.index = i
	}
	if len(fixedVertices) < 2 {
		return errors.New("at least two fixed vertices are required")
	}

	// prepare matrices of coefficients
	fn := len(mesh.faces)
	vfn := len(fixedVertices)
	vn := len(vertices)
	amat := mat.NewDense(2*fn, 2*vn, nil)
	bmat := mat.NewDense(2*fn, 2*vfn, nil)
	fmat := mat.NewVecDense(2*vfn, nil)
	for _, e := range mesh.edges {
		p1 := mesh.getPoint(e.halfedges[0].source().id)
		p2 := mesh.getPoint(e.halfedges[0].target().id)
		vd := p1.sub(&p2)
		e.length = vd.norm()
	}
	for fid, f := range mesh.faces {
		hel := [3]float32{}
		he := f.halfedge
		for i := 0; i < 3; i++ {
			hel[i] = he.edge.length
			he = he.next
		}
		// law of cosines
		a := math.Acos(float64((hel[0]*hel[0] + hel[2]*hel[2] - hel[1]*hel[1]) / (2 * hel[0] * hel[2])))
		p := [3]point3D{
			{0, 0, 0},
			{hel[0], 0, 0},
			{hel[2] * float32(math.Cos(a)), hel[2] * float32(math.Sin(a)), 0},
		}
		n0 := p[1].sub(&p[0])
		n1 := p[2].sub(&p[0])
		n := n0.cross(&n1)
		area := n.norm() / 2.0
		n.divide(area)
		he = f.halfedge
		for i := 0; i < 3; i++ {
			np := p[(i+1)%3].sub(&p[i])
			c := n.cross(&np)
			c.divide(float32(math.Sqrt(float64(area))))
			v := he.next.target()
			vid := v.index
			if !v.fixed {
				amat.Set(fid, vid, float64(c.x))
				amat.Set(fn+fid, vn+vid, float64(c.x))
				amat.Set(fid, vn+vid, float64(-c.y))
				amat.Set(fn+fid, vid, float64(c.y))
			} else {
				bmat.Set(fid, vid, float64(c.x))
				bmat.Set(fn+fid, vfn+vid, float64(c.x))
				bmat.Set(fid, vfn+vid, float64(-c.y))
				bmat.Set(fn+fid, vid, float64(c.y))
				uv := mesh.getUV(v.id)
				fmat.SetVec(vid, float64(uv.x))
				fmat.SetVec(vfn+vid, float64(uv.y))
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
	uvMin := point2D{
		x: math.MaxFloat32,
		y: math.MaxFloat32,
	}
	uvMax := point2D{
		x: -math.MaxFloat32,
		y: -math.MaxFloat32,
	}
	for i, v := range vertices {
		uv := point2D{
			x: float32(smat.At(i, 0)),
			y: float32(smat.At(i+vn, 0)),
		}
		mesh.setUV(v.id, uv)
		uvMin.x = min(uvMin.x, uv.x)
		uvMin.y = min(uvMin.y, uv.y)
		uvMax.x = max(uvMax.x, uv.x)
		uvMax.y = max(uvMax.y, uv.y)
	}
	// scale UVs to be within the range [0:1]
	for _, v := range mesh.vertices {
		if v.halfedge == nil {
			// ignore dangling vertices
			continue
		}
		uv := mesh.getUV(v.id)
		uv = point2D{
			x: (uv.x - uvMin.x) / (uvMax.x - uvMin.x),
			y: (uv.y - uvMin.y) / (uvMax.y - uvMin.y),
		}
		mesh.setUV(v.id, uv)
	}

	return nil
}
