package lscm

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type parseData struct {
	points       []float32
	uvs          []float32
	normals      []float32
	indices      []uint32
	fixedIndices []uint32
}

// ParseObj parses a Wavefront .obj file content into a Mesh.
//
// At least two of the vertex lines need "fix 0.0, 0.0" post-fixed to the line to indicate fixed vertices for LSCM
// The coordinates are fixed texture coordinates for those vertexes.
//
// This does not support all obj properties.
// Only parses vertices, normals, and faces while ignoring other lines.
// Assumes the corresponding normal indices are the same as the vertex indices.
func ParseObj(obj string) (*Mesh, error) {
	pd := &parseData{
		points:       make([]float32, 0, 5101*3),
		uvs:          make([]float32, 0, 5101*2),
		normals:      make([]float32, 0, 5101*3),
		indices:      make([]uint32, 0, 10000*3),
		fixedIndices: make([]uint32, 0, 2),
	}
	var line, t string
	var ok bool
	var err error
	for {
		line, obj, ok = strings.Cut(obj, "\n")
		if !ok {
			break
		}
		t, line, ok = strings.Cut(line, " ")
		if !ok {
			return nil, errors.New("expected space after line type token")
		}
		switch t {
		case "v":
			if err = pd.parseVertex(line); err != nil {
				return nil, err
			}
		case "vn":
			if err = pd.parseNormal(line); err != nil {
				return nil, err
			}
		case "f":
			if err = pd.parseFace(line); err != nil {
				return nil, err
			}
		default:
			// ignore everything else
		}
	}
	return NewMesh(pd.points, pd.uvs, pd.normals, pd.indices, pd.fixedIndices)
}

func (d *parseData) parseFace(line string) error {
	var as, bs, cs string
	var a, b, c int
	var ok bool
	var err error
	as, line, ok = strings.Cut(line, " ")
	if !ok {
		return errors.New("expected space after face index a")
	}
	as, _, _ = strings.Cut(as, "/")
	bs, line, ok = strings.Cut(line, " ")
	if !ok {
		return errors.New("expected space after face index b")
	}
	bs, _, _ = strings.Cut(bs, "/")
	cs, line, _ = strings.Cut(line, " ")
	cs, _, _ = strings.Cut(cs, "/")
	a, err = strconv.Atoi(as)
	if err != nil {
		return err
	}
	b, err = strconv.Atoi(bs)
	if err != nil {
		return err
	}
	c, err = strconv.Atoi(cs)
	if err != nil {
		return err
	}
	d.indices = append(d.indices, uint32(a-1), uint32(b-1), uint32(c-1))
	return nil
}

func (d *parseData) parseNormal(line string) error {
	var xs, ys, zs string
	var x, y, z float64
	var ok bool
	var err error
	xs, line, ok = strings.Cut(line, " ")
	if !ok {
		return errors.New("space not found separating coordinates for vertex normal")
	}
	ys, line, ok = strings.Cut(line, " ")
	if !ok {
		return errors.New("space not found separating coordinates for vertex normal")
	}
	zs, line, _ = strings.Cut(line, " ")
	x, err = strconv.ParseFloat(xs, 32)
	if err != nil {
		return err
	}
	y, err = strconv.ParseFloat(ys, 32)
	if err != nil {
		return err
	}
	z, err = strconv.ParseFloat(zs, 32)
	if err != nil {
		return err
	}
	d.normals = append(d.normals, float32(x), float32(y), float32(z))
	return nil
}

func (d *parseData) parseVertex(line string) error {
	var xs, ys, zs string
	var x, y, z, u, v float64
	var ok bool
	var err error
	if xs, line, ok = strings.Cut(line, " "); !ok {
		return errors.New("expected space after vertex x coordinate")
	}
	if ys, line, ok = strings.Cut(line, " "); !ok {
		return errors.New("expected space after vertex y coordinate")
	}
	zs, line, _ = strings.Cut(line, " ")
	if x, err = strconv.ParseFloat(xs, 32); err != nil {
		return err
	}
	if y, err = strconv.ParseFloat(ys, 32); err != nil {
		return err
	}
	if z, err = strconv.ParseFloat(zs, 32); err != nil {
		return err
	}
	fixLine, fixed := strings.CutPrefix(line, "fix ")
	if fixed {
		xs, ys, ok = strings.Cut(fixLine, " ")
		if !ok {
			return errors.New("expected space after fixed vertex x coordinate")
		}
		u, err = strconv.ParseFloat(xs, 32)
		if err != nil {
			return err
		}
		v, err = strconv.ParseFloat(ys, 32)
		if err != nil {
			return err
		}
		d.fixedIndices = append(d.fixedIndices, uint32(len(d.points)/3))
	}
	d.points = append(d.points, float32(x), float32(y), float32(z))
	d.uvs = append(d.uvs, float32(u), float32(v))
	return nil
}

// WriteObj writes a Wavefront .obj file content to the writer.
//
// This only supports vertices, texture coordinates, normals, and faces.
func WriteObj(w io.Writer, m *Mesh) error {
	// vertices
	for _, vertex := range m.vertices {
		p := m.getPoint(vertex.id)
		if _, err := fmt.Fprintf(w, "v %f %f %f\n", p.x, p.y, p.z); err != nil {
			return err
		}
	}
	// texture coordinates
	for _, vertex := range m.vertices {
		uv := m.getUV(vertex.id)
		if _, err := fmt.Fprintf(w, "vt %f %f\n", uv.x, uv.y); err != nil {
			return err
		}
	}
	// vertex normals
	for _, vertex := range m.vertices {
		n := m.getNormal(vertex.id)
		if _, err := fmt.Fprintf(w, "vn %f %f %f\n", n.x, n.y, n.z); err != nil {
			return err
		}
	}
	// faces
	for _, face := range m.faces {
		if _, err := fmt.Fprint(w, "f"); err != nil {
			return err
		}
		halfedge := face.halfedge
		for {
			vertex := halfedge.vertex
			if _, err := fmt.Fprintf(w, " %d/%d/%d", vertex.id+1, vertex.id+1, vertex.id+1); err != nil {
				return err
			}
			halfedge = halfedge.next
			if halfedge == face.halfedge {
				break
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return nil
}
