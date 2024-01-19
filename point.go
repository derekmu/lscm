package lscm

import "math"

type Point3D struct {
	X float64
	Y float64
	Z float64
}

func (d *Point3D) sub(point *Point3D) Point3D {
	return Point3D{
		X: d.X - point.X,
		Y: d.Y - point.Y,
		Z: d.Z - point.Z,
	}
}

func (d *Point3D) norm() float64 {
	return math.Sqrt(d.X*d.X + d.Y*d.Y + d.Z*d.Z)
}

func (d *Point3D) cross(p *Point3D) Point3D {
	return Point3D{
		X: d.Y*p.Z - d.Z*p.Y,
		Y: d.Z*p.X - d.X*p.Z,
		Z: d.X*p.Y - d.Y*p.X,
	}
}

func (d *Point3D) divide(v float64) {
	d.X /= v
	d.Y /= v
	d.Z /= v
}

type Point2D struct {
	X float64
	Y float64
}
