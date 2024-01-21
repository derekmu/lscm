package lscm

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseObj parses a Wavefront .obj file content into a Mesh.
//
// At least two of the vertex lines need "fix 0.0, 0.0" post-fixed to the line to indicate fixed vertices for LSCM
// The coordinates are fixed texture coordinates for those vertexes.
//
// This does not support all obj properties.
// Only parses vertices, normals, and faces while ignoring other lines.
// Assumes the corresponding normal indices are the same as the vertex indices.
func ParseObj(obj string) (*Mesh, error) {
	m := NewMesh()
	ni := 0
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
			if err = parseVertex(m, line); err != nil {
				return nil, err
			}
		case "vn":
			if err = parseNormal(m, line, ni); err != nil {
				return nil, err
			}
			ni++
		case "f":
			if err = parseFace(m, line); err != nil {
				return nil, err
			}
		default:
			// ignore everything else
		}
	}
	return m, nil
}

func parseFace(m *Mesh, line string) error {
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
	m.AddFace([3]int{a - 1, b - 1, c - 1})
	return nil
}

func parseNormal(m *Mesh, line string, ni int) error {
	var xs, ys, zs string
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
	v := m.vertices[ni]
	v.normal.X, err = strconv.ParseFloat(xs, 64)
	if err != nil {
		return err
	}
	v.normal.Y, err = strconv.ParseFloat(ys, 64)
	if err != nil {
		return err
	}
	v.normal.Z, err = strconv.ParseFloat(zs, 64)
	if err != nil {
		return err
	}
	return nil
}

func parseVertex(m *Mesh, line string) error {
	var xs, ys, zs string
	var ok bool
	var err error
	if xs, line, ok = strings.Cut(line, " "); !ok {
		return errors.New("expected space after vertex x coordinate")
	}
	if ys, line, ok = strings.Cut(line, " "); !ok {
		return errors.New("expected space after vertex y coordinate")
	}
	zs, line, _ = strings.Cut(line, " ")
	p := Point3D{}
	if p.X, err = strconv.ParseFloat(xs, 64); err != nil {
		return err
	}
	if p.Y, err = strconv.ParseFloat(ys, 64); err != nil {
		return err
	}
	if p.Z, err = strconv.ParseFloat(zs, 64); err != nil {
		return err
	}
	fixLine, fixed := strings.CutPrefix(line, "fix ")
	if fixed {
		xs, ys, ok = strings.Cut(fixLine, " ")
		if !ok {
			return errors.New("expected space after fixed vertex x coordinate")
		}
		uv := Point2D{}
		uv.X, err = strconv.ParseFloat(xs, 64)
		if err != nil {
			return err
		}
		uv.Y, err = strconv.ParseFloat(ys, 64)
		if err != nil {
			return err
		}
		m.AddVertex(p, true, uv)
	} else {
		m.AddVertex(p, false, Point2D{})
	}
	return nil
}

// WriteObj writes a Wavefront .obj file content to the writer.
//
// This only supports vertices, texture coordinates, normals, and faces.
func WriteObj(w io.Writer, m *Mesh) error {
	// vertices
	for _, vertex := range m.vertices {
		if _, err := fmt.Fprintf(w, "v %f %f %f\n", vertex.point.X, vertex.point.Y, vertex.point.Z); err != nil {
			return err
		}
	}
	// texture coordinates
	for _, vertex := range m.vertices {
		if _, err := fmt.Fprintf(w, "vt %f %f\n", vertex.uv.X, vertex.uv.Y); err != nil {
			return err
		}
	}
	// vertex normals
	for _, vertex := range m.vertices {
		if _, err := fmt.Fprintf(w, "vn %f %f %f\n", vertex.normal.X, vertex.normal.Y, vertex.normal.Z); err != nil {
			return err
		}
	}
	// faces
	for _, face := range m.faces {
		if _, err := fmt.Fprint(w, "f "); err != nil {
			return err
		}
		halfedge := face.halfedge
		for {
			vertex := halfedge.vertex
			if _, err := fmt.Fprintf(w, "%d/%d/%d ", vertex.id+1, vertex.id+1, vertex.id+1); err != nil {
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
