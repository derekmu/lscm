package lscm

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ParseObj(obj string) (*Mesh, error) {
	m := NewMesh()
	vid := 1
	nid := 1
	var ok bool
	var line, t, xs, ys, zs, as, bs, cs string
	var a, b, c int
	var av, bv, cv *Vertex
	var err error
	for {
		line, obj, ok = strings.Cut(obj, "\n")
		if !ok {
			break
		}
		t, line, ok = strings.Cut(line, " ")
		if !ok {
			return nil, errors.New("space not found separating line type token")
		}
		switch t {
		case "v":
			p := Point3D{}
			xs, line, ok = strings.Cut(line, " ")
			if !ok {
				return nil, errors.New("space not found separating coordinates for vertex")
			}
			ys, line, ok = strings.Cut(line, " ")
			if !ok {
				return nil, errors.New("space not found separating coordinates for vertex")
			}
			zs, line, _ = strings.Cut(line, " ")

			p.X, err = strconv.ParseFloat(xs, 64)
			if err != nil {
				return nil, err
			}
			p.Y, err = strconv.ParseFloat(ys, 64)
			if err != nil {
				return nil, err
			}
			p.Z, err = strconv.ParseFloat(zs, 64)
			if err != nil {
				return nil, err
			}

			v := m.createVertex(vid, p)
			line, v.fixed = strings.CutPrefix(line, "fix ")
			if v.fixed {
				xs, ys, ok = strings.Cut(line, " ")
				if !ok {
					return nil, errors.New("space not found separating fixed coordinates for vertex")
				}
				v.uv.X, err = strconv.ParseFloat(xs, 64)
				if err != nil {
					return nil, err
				}
				v.uv.Y, err = strconv.ParseFloat(ys, 64)
				if err != nil {
					return nil, err
				}
			}
			vid++
		case "vn":
			v := m.vertexMap[nid]
			xs, line, ok = strings.Cut(line, " ")
			if !ok {
				return nil, errors.New("space not found separating coordinates for vertex normal")
			}
			ys, line, ok = strings.Cut(line, " ")
			if !ok {
				return nil, errors.New("space not found separating coordinates for vertex normal")
			}
			zs, line, _ = strings.Cut(line, " ")

			v.normal.X, err = strconv.ParseFloat(xs, 64)
			if err != nil {
				return nil, err
			}
			v.normal.Y, err = strconv.ParseFloat(ys, 64)
			if err != nil {
				return nil, err
			}
			v.normal.Z, err = strconv.ParseFloat(zs, 64)
			if err != nil {
				return nil, err
			}
			nid++
		case "f":
			as, line, ok = strings.Cut(line, " ")
			if !ok {
				return nil, errors.New("space not found separating vertex index for face")
			}
			as, _, _ = strings.Cut(as, "/")
			bs, line, ok = strings.Cut(line, " ")
			if !ok {
				return nil, errors.New("space not found separating vertex index for face")
			}
			bs, _, _ = strings.Cut(bs, "/")
			cs, line, _ = strings.Cut(line, " ")
			cs, _, _ = strings.Cut(cs, "/")

			a, err = strconv.Atoi(as)
			if err != nil {
				return nil, err
			}
			b, err = strconv.Atoi(bs)
			if err != nil {
				return nil, err
			}
			c, err = strconv.Atoi(cs)
			if err != nil {
				return nil, err
			}

			if av, ok = m.vertexMap[a]; !ok {
				return nil, errors.New("vertex not found for face")
			}
			if bv, ok = m.vertexMap[b]; !ok {
				return nil, errors.New("vertex not found for face")
			}
			if cv, ok = m.vertexMap[c]; !ok {
				return nil, errors.New("vertex not found for face")
			}

			m.createFace([3]*Vertex{av, bv, cv})
		default:
			// we don't need anything except vertices and faces
		}
	}
	// label boundary edges
	for _, edge := range m.edges {
		if edge.halfedges[1] != nil {
			if edge.halfedges[0].target().id < edge.halfedges[0].source().id {
				edge.halfedges[0], edge.halfedges[1] = edge.halfedges[1], edge.halfedges[0]
			}
		} else {
			edge.halfedges[0].vertex.boundary = true
			edge.halfedges[0].prev.vertex.boundary = true
		}
	}
	// remove dangling vertices
	for i := 0; i < len(m.vertices); i++ {
		v := m.vertices[i]
		if v.halfedge == nil {
			m.vertices[i] = m.vertices[len(m.vertices)-1]
			m.vertices = m.vertices[:len(m.vertices)-1]
			i--
			delete(m.vertexMap, v.id)
		}
	}
	// arrange the half edge of boundary vertices to make it's halfedge the most counterclockwise
	for _, v := range m.vertices {
		if v.boundary {
			for v.halfedge.other() != nil {
				v.halfedge = v.halfedge.other().prev
			}
		}
	}
	return m, nil
}

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
			if _, err := fmt.Fprintf(w, "%d/%d/%d ", vertex.id, vertex.id, vertex.id); err != nil {
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
