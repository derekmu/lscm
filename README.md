# Least Square Conformal Mapping

Least Squares Conformal Mapping (LSCM) is a mathematical technique that helps flatten and map irregular surfaces onto a
2D plane while preserving the angles between neighboring points as much as possible.

This is particularly useful for tasks like texture mapping in 3D graphics.

## Usage

See [demo/main.go](./demo/main.go) for an example of how to read a Wavefront .obj file into a mesh and
generate texture coordinates using LSCM. An example face model is provided without vertex coordinates. The demo takes
several minutes to run.

```shell
go run demo/main.go
```

### Example

Below is an animation of the face model after adding LSCM-generated vertex coordinates. The texture is generated using
the vertex coordinates to draw red-black triangles for each face.
Rendering was done with [g3n](https://github.com/g3n/engine).

![human-face](./demo/human.gif)

## Credits

See the paper at https://dl.acm.org/doi/10.1145/566654.566590.

Code is based on https://github.com/icemiliang/lscm.