package lscm

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Mesh struct {
	edges     []*Edge
	faces     []*Face
	vertices  []*Vertex
	vertexMap map[int]*Vertex
	faceMap   map[int]*Face
	edgeMap   map[EdgeKey]*Edge
}

func NewMesh() *Mesh {
	return &Mesh{
		vertexMap: make(map[int]*Vertex, 128),
		faceMap:   make(map[int]*Face, 128),
		edgeMap:   make(map[EdgeKey]*Edge, 128),
	}
}

func (m *Mesh) Parse(mesh string) error {
	vid := 1
	fid := 1
	nid := 1
	var ok bool
	var line, t, xs, ys, zs, as, bs, cs string
	var a, b, c int
	var av, bv, cv *Vertex
	var err error
	for {
		line, mesh, ok = strings.Cut(mesh, "\n")
		if !ok {
			break
		}
		t, line, ok = strings.Cut(line, " ")
		if !ok {
			return errors.New("space not found separating line type token")
		}
		switch t {
		case "v":
			p := Point3D{}
			xs, line, ok = strings.Cut(line, " ")
			if !ok {
				return errors.New("space not found separating coordinates for vertex")
			}
			ys, line, ok = strings.Cut(line, " ")
			if !ok {
				return errors.New("space not found separating coordinates for vertex")
			}
			zs, line, _ = strings.Cut(line, " ")

			p.X, err = strconv.ParseFloat(xs, 64)
			if err != nil {
				return err
			}
			p.Y, err = strconv.ParseFloat(ys, 64)
			if err != nil {
				return err
			}
			p.Z, err = strconv.ParseFloat(zs, 64)
			if err != nil {
				return err
			}

			v := m.createVertex(vid, p)
			line, v.fixed = strings.CutPrefix(line, "fix ")
			if v.fixed {
				xs, ys, ok = strings.Cut(line, " ")
				if !ok {
					return errors.New("space not found separating fixed coordinates for vertex")
				}
				v.uv.X, err = strconv.ParseFloat(xs, 64)
				if err != nil {
					return err
				}
				v.uv.Y, err = strconv.ParseFloat(ys, 64)
				if err != nil {
					return err
				}
			}
			vid++
		case "vn":
			v := m.vertexMap[nid]
			xs, line, ok = strings.Cut(line, " ")
			if !ok {
				return errors.New("space not found separating coordinates for vertex normal")
			}
			ys, line, ok = strings.Cut(line, " ")
			if !ok {
				return errors.New("space not found separating coordinates for vertex normal")
			}
			zs, line, _ = strings.Cut(line, " ")

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
			nid++
		case "f":
			as, line, ok = strings.Cut(line, " ")
			if !ok {
				return errors.New("space not found separating vertex index for face")
			}
			as, _, _ = strings.Cut(as, "/")
			bs, line, ok = strings.Cut(line, " ")
			if !ok {
				return errors.New("space not found separating vertex index for face")
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

			if av, ok = m.vertexMap[a]; !ok {
				return errors.New("vertex not found for face")
			}
			if bv, ok = m.vertexMap[b]; !ok {
				return errors.New("vertex not found for face")
			}
			if cv, ok = m.vertexMap[c]; !ok {
				return errors.New("vertex not found for face")
			}

			m.createFace(fid, [3]*Vertex{av, bv, cv})
			fid++
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
	return nil
}

func (m *Mesh) Write(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Panic(err)
		}
	}(file)
	writer := bufio.NewWriter(file)
	// vertices
	for _, vertex := range m.vertices {
		_, err = fmt.Fprintf(writer, "v %f %f %f\n", vertex.point.X, vertex.point.Y, vertex.point.Z)
		if err != nil {
			return err
		}
	}
	// texture coordinates
	for _, vertex := range m.vertices {
		_, err = fmt.Fprintf(writer, "vt %f %f\n", vertex.uv.X, vertex.uv.Y)
		if err != nil {
			return err
		}
	}
	// vertex normals
	for _, vertex := range m.vertices {
		_, err = fmt.Fprintf(writer, "vn %f %f %f\n", vertex.normal.X, vertex.normal.Y, vertex.normal.Z)
		if err != nil {
			return err
		}
	}
	// faces
	for _, face := range m.faces {
		_, err := writer.WriteString("f ")
		if err != nil {
			return err
		}
		halfedge := face.halfedge
		for {
			vertex := halfedge.vertex
			_, err = fmt.Fprintf(writer, "%d/%d/%d ", vertex.id, vertex.id, vertex.id)
			if err != nil {
				return err
			}
			halfedge = halfedge.next
			if halfedge == face.halfedge {
				break
			}
		}
		_, err = fmt.Fprintln(writer)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (m *Mesh) createVertex(id int, p Point3D) *Vertex {
	v := &Vertex{
		id:    id,
		point: p,
	}
	m.vertices = append(m.vertices, v)
	m.vertexMap[id] = v
	return v
}

func (m *Mesh) createFace(id int, vertices [3]*Vertex) *Face {
	face := &Face{
		id: id,
	}
	m.faces = append(m.faces, face)
	m.faceMap[id] = face

	// create halfedges
	halfedges := [3]*HalfEdge{}
	for i := 0; i < 3; i++ {
		vertex := vertices[i]
		he := &HalfEdge{
			face:   face,
			vertex: vertex,
		}
		vertex.halfedge = he
		halfedges[i] = he
	}
	// link to each other
	for i := 0; i < 3; i++ {
		halfedges[i].next = halfedges[(i+1)%3]
		halfedges[i].prev = halfedges[(i+2)%3]
	}
	// link to face
	face.halfedge = halfedges[0]
	// link to edges
	for i := 0; i < 3; i++ {
		edge := m.createEdge(vertices[i], vertices[(i+2)%3])
		if edge.halfedges[0] == nil {
			edge.halfedges[0] = halfedges[i]
		} else {
			edge.halfedges[1] = halfedges[i]
		}
		halfedges[i].edge = edge
	}
	return face
}

func (m *Mesh) createEdge(vertex1 *Vertex, vertex2 *Vertex) *Edge {
	key := EdgeKey{min(vertex1.id, vertex2.id), max(vertex1.id, vertex2.id)}
	if edge, ok := m.edgeMap[key]; ok {
		return edge
	}
	edge := &Edge{
		halfedges: [2]*HalfEdge{},
	}
	m.edges = append(m.edges, edge)
	m.edgeMap[key] = edge
	return edge
}
