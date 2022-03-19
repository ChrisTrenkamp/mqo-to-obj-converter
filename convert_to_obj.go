package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
)

func ConvertMqoObjectsToMtl(out io.Writer, scene *Scene) (err error) {
	for _, material := range scene.Materials {
		_, err = fmt.Fprintf(out, "newmtl %s\n", material.Name)
		if err != nil {
			return err
		}

		if material.TextureMappingFile != "" {
			_, err = fmt.Fprintf(out, "map_Kd %s\n", material.TextureMappingFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ConvertMqoObjectsToObj(out io.Writer, scene *Scene, mtlName string) (err error) {
	objects := scene.Objects
	vertexOffsets := make([]int, len(objects)+1)
	currentMaterialIndex := -1
	currentUvCoordinateIndex := 1

	_, err = fmt.Fprintf(out, "mtllib %s\n", mtlName)

	if err != nil {
		return
	}

	for offset, mqo := range objects {
		vertexOffsets[offset+1] = len(mqo.Vertices) + vertexOffsets[offset]

		for _, i := range mqo.Vertices {
			_, err = fmt.Fprintf(out, "v %f %f %f\n", i.X, i.Y, i.Z)

			if err != nil {
				return
			}
		}
	}

	for offset, mqo := range objects {
		vertexOffset := vertexOffsets[offset]

		_, err = fmt.Fprintf(out, "o %s\ng %s\n", mqo.Name, mqo.Name)
		if err != nil {
			return
		}

		for _, i := range mqo.Faces {
			end := 0

			for end < 4 {
				if i.VertexIndices[end] < 0 {
					break
				}

				end++
			}

			if end < 2 {
				continue
			}

			if i.MaterialIndex != currentMaterialIndex {
				currentMaterialIndex = i.MaterialIndex
				_, err = fmt.Fprintf(out, "usemtl %s\n", scene.Materials[currentMaterialIndex].Name)

				if err != nil {
					return
				}
			}

			if i.TextureCoordinates == nil {
				err = writeObjFaceWithoutUvCoordinates(out, &i, end, vertexOffset)
			} else {
				err = writeObjFaceWithUvCoordinates(out, &i, end, vertexOffset, currentUvCoordinateIndex)
				currentUvCoordinateIndex += len(i.TextureCoordinates) / 2
			}

			if err != nil {
				return
			}

			_, err = fmt.Fprintf(out, "\n")

			if err != nil {
				return
			}
		}
	}

	return
}

func writeObjFaceWithoutUvCoordinates(out io.Writer, i *Face, end, vertexOffset int) error {
	_, err := fmt.Fprintf(out, "f ")

	if err != nil {
		return err
	}

	for start := 0; start < end; start++ {
		_, err = fmt.Fprintf(out, "%d ", i.VertexIndices[start]+1+vertexOffset)

		if err != nil {
			return err
		}
	}

	return nil
}

func writeObjFaceWithUvCoordinates(out io.Writer, i *Face, end, vertexOffset, currentUvCoordinateIndex int) error {
	for start := 0; start < len(i.TextureCoordinates); start += 2 {
		_, err := fmt.Fprintf(out, "vt %f %f\n", i.TextureCoordinates[start], i.TextureCoordinates[start+1])

		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(out, "f ")

	if err != nil {
		return err
	}

	for start := 0; start < end; start++ {
		_, err = fmt.Fprintf(out, "%d/", i.VertexIndices[start]+1+vertexOffset)

		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(out, "%d ", start+currentUvCoordinateIndex)

		if err != nil {
			return err
		}
	}

	return nil
}

func CopyMaterialTextureFiles(mqoDir, objDir string, scene *Scene) error {
	for _, tex := range scene.Materials {
		if tex.TextureMappingFile != "" {
			texData, err := ioutil.ReadFile(filepath.Join(mqoDir, tex.TextureMappingFile))

			if err != nil {
				return err
			}

			err = ioutil.WriteFile(filepath.Join(objDir, tex.TextureMappingFile), texData, 0644)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
