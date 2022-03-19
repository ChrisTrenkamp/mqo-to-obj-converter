package main

import "testing"

func TestVertices(t *testing.T) {
	in := `
Metasequoia Document
Format Text Ver 1.0

unknown {
	Object {
		vertex 1 {
			-1 -1 -1
		}
	}
}

Material 1 {
	"texture1" power(5.00) tex("somefile.bmp")
}

Object "bone:Bone" {
	vertex 5 {
		0.0000 1 02
		3.14 4.0 5.0
		6 7 8
		9 10 11
		12 13 14
	}
	face 2 {
		2 V(0 1)
		3 V(2 3 4) M(1) UV(0.0 1.0 0.5 0.5 0.25 0.25)
	}
}
`

	scene := Scene{}
	if err := scene.Parse([]byte(in)); err != nil {
		t.Error(err)
		return
	}

	v := scene.Objects[0].Vertices[0]
	if v.X != 0 || v.Y != 1 || v.Z != 2 {
		t.Error("Got vertex", v)
		t.Error("Expected 0, 1, 2")
		return
	}

	v = scene.Objects[0].Vertices[1]
	if v.X != 3.14 || v.Y != 4 || v.Z != 5 {
		t.Error("Got vertex", v)
		t.Error("Expected 3.14, 4, 5")
		return
	}

	f := scene.Objects[0].Faces[0]
	if f.VertexIndices[0] != 0 || f.VertexIndices[1] != 1 || f.VertexIndices[2] >= 0 || f.VertexIndices[3] >= 0 {
		t.Error("Got face", f)
		t.Error("Expected 0, 1")
		return
	}

	f = scene.Objects[0].Faces[1]
	if f.VertexIndices[0] != 2 || f.VertexIndices[1] != 3 || f.VertexIndices[2] != 4 || f.VertexIndices[3] >= 0 {
		t.Error("Got face", f)
		t.Error("Expected 2, 3, 4")
		return
	}

	if f.TextureCoordinates[0] != 0.0 ||
		f.TextureCoordinates[1] != 1.0 ||
		f.TextureCoordinates[2] != 0.5 ||
		f.TextureCoordinates[3] != 0.5 ||
		f.TextureCoordinates[4] != 0.25 ||
		f.TextureCoordinates[5] != 0.25 {
		t.Error("Bad texture coordinates,", f.TextureCoordinates)
		t.Error("Expected 0.0 1.0 0.5 0.5 0.25 0.25")
		return
	}

	m := scene.Materials[0]
	if m.Name != "texture1" {
		t.Error("Texture name is", m.Name)
		t.Error("Expected texture1")
		return
	}

	if m.TextureMappingFile != "somefile.bmp" {
		t.Error("TextureMappingFile is", m.TextureMappingFile)
		t.Error("Expected somefile.bmp")
		return
	}
}
