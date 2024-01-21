package lscm

import "math"

type point3D struct {
	x float32
	y float32
	z float32
}

func (d *point3D) sub(point *point3D) point3D {
	return point3D{
		x: d.x - point.x,
		y: d.y - point.y,
		z: d.z - point.z,
	}
}

func (d *point3D) norm() float32 {
	return float32(math.Sqrt(float64(d.x*d.x + d.y*d.y + d.z*d.z)))
}

func (d *point3D) cross(p *point3D) point3D {
	return point3D{
		x: d.y*p.z - d.z*p.y,
		y: d.z*p.x - d.x*p.z,
		z: d.x*p.y - d.y*p.x,
	}
}

func (d *point3D) divide(v float32) {
	d.x /= v
	d.y /= v
	d.z /= v
}

type point2D struct {
	x float32
	y float32
}
