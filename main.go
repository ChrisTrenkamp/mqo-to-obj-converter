package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var validFileNameChars = regexp.MustCompile("[^a-zA-Z0-9]+")

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Printf("%s usage: [path/to/file.mqo] [path/to/converted.obj]\n", os.Args[0])
		return
	}

	mqoFile, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	scene := Scene{}
	if err = scene.Parse(mqoFile); err != nil {
		fmt.Println(err)
		return
	}

	objFile := flag.Arg(1)

	if err = os.MkdirAll(filepath.Dir(objFile), 0755); err != nil {
		fmt.Println(err)
		return
	}

	mtlFile := strings.TrimSuffix(objFile, ".obj") + ".mtl"

	file, err := os.Create(mtlFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ConvertMqoObjectsToMtl(file, &scene)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = file.Close(); err != nil {
		fmt.Println(err)
		return
	}

	file, err = os.Create(objFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ConvertMqoObjectsToObj(file, &scene, filepath.Base(mtlFile))
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = file.Close(); err != nil {
		fmt.Println(err)
		return
	}

	err = CopyMaterialTextureFiles(filepath.Dir(flag.Arg(0)), filepath.Dir(objFile), &scene)

	if err != nil {
		fmt.Println(err)
		return
	}
}
