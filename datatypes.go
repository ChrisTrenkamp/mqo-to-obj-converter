package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Scene struct {
	Materials []Material
	Objects   []Object
}

type Material struct {
	Name               string
	TextureMappingFile string
}

type Object struct {
	Name     string
	Vertices []Vertex
	Faces    []Face
}

type Vertex struct {
	X float32
	Y float32
	Z float32
}

type Face struct {
	VertexIndices      [4]int
	MaterialIndex      int
	TextureCoordinates []float32
}

type mqoState struct {
	pos  int
	cur  int
	data []byte
}

func (s *Scene) Parse(mqoFile []byte) error {
	state := mqoState{0, 0, mqoFile}
	return s.parseMain(&state)
}

func (s *Scene) parseMain(state *mqoState) error {
	state.skipSpace()
	line := state.readLine()

	if line != "Metasequoia Document" {
		return fmt.Errorf("missing 'Metasequoia Document' declaration")
	}

	state.skipSpace()
	line = state.readLine()

	if line != "Format Text Ver 1.0" {
		return fmt.Errorf("unsupported text version")
	}

	for {
		word := state.readWord()

		if word == "" {
			return nil
		}

		switch word {
		case "":
			return nil
		case "Object":
			object := Object{}
			object.parseObject(state)
			s.Objects = append(s.Objects, object)
		case "Material":
			s.parseMaterials(state)
		case "{":
			state.skipToEndBracket(1)
		}
	}
}

func (s *Scene) parseMaterials(state *mqoState) {
	state.skipSpace()

	numMaterials, err := state.readPositiveInteger()
	if err != nil {
		//TODO: Report error
		return
	}

	word := state.readWord()

	if word != "{" {
		//TODO: Report error
		return
	}

	s.Materials = make([]Material, numMaterials)

	for i := 0; i < numMaterials; i++ {
		state.skipSpace()

		line := state.readLine()
		materialState := mqoState{0, 0, []byte(line)}
		material := Material{}
		material.parseMaterial(&materialState)
		s.Materials[i] = material
	}
}

func (m *Material) parseMaterial(state *mqoState) {
	name, err := state.readQuote()
	if err != nil {
		//TODO: Report error
		return
	}

	m.Name = name

	for {
		state.skipSpace()

		name, content, err := state.readParentheticArg()
		if err != nil {
			//TODO: Report error
			return
		}

		if name == "" && err == nil {
			return
		}

		materialFunc := materialArgumentFuncs[name]

		if materialFunc != nil {
			err = materialFunc(m, content)

			if err != nil {
				//TODO: Report error
				return
			}
		}
	}
}

var materialArgumentFuncs = map[string]func(m *Material, content string) error{
	"tex": parseMaterialTex,
}

func parseMaterialTex(m *Material, content string) error {
	texState := mqoState{0, 0, []byte(content)}
	file, err := texState.readQuote()
	if err != nil {
		return err
	}

	m.TextureMappingFile = file
	return nil
}

func (o *Object) parseObject(state *mqoState) {
	state.skipSpace()

	word, err := state.readQuote()
	if err != nil {
		//TODO: Report error
		return
	}

	o.Name = word
	word = state.readWord()

	if word != "{" {
		//TODO: Report error
		return
	}

	for {
		word = state.readWord()

		switch word {
		case "}":
			return
		case "vertex":
			o.parseVertices(state)
		case "face":
			o.parseFaces(state)
		}
	}
}

func (o *Object) parseVertices(state *mqoState) {
	state.skipSpace()

	numVertices, err := state.readPositiveInteger()

	if err != nil {
		//TODO: Report error
		return
	}

	word := state.readWord()

	if word != "{" {
		//TODO: Report error
		return
	}

	o.Vertices = make([]Vertex, numVertices)

	for i := 0; i < numVertices; i++ {
		state.skipSpace()
		x, err := state.readFloat()

		if err != nil {
			//TODO: Report error
			return
		}

		state.skipSpace()
		y, err := state.readFloat()

		if err != nil {
			//TODO: Report error
			return
		}

		state.skipSpace()
		z, err := state.readFloat()

		if err != nil {
			//TODO: Report error
			return
		}

		o.Vertices[i] = Vertex{x, y, z}
	}

	word = state.readWord()

	if word != "}" {
		//TODO: Report error
		return
	}
}

var whiteSpaceRegexp = regexp.MustCompile(`\s`)

func (o *Object) parseFaces(state *mqoState) {
	state.skipSpace()

	numFaces, err := state.readPositiveInteger()

	if err != nil {
		//TODO: Report error
		return
	}

	word := state.readWord()

	if word != "{" {
		//TODO: Report error
		return
	}

	o.Faces = make([]Face, numFaces)

	for i := 0; i < numFaces; i++ {
		state.skipSpace()
		line := state.readLine()

		if line == "" {
			return
		}

		faceState := mqoState{0, 0, []byte(line)}
		face := Face{
			VertexIndices:      [4]int{-1, -1, -1, -1},
			MaterialIndex:      -1,
			TextureCoordinates: nil,
		}
		face.parseFaceLine(&faceState)
		o.Faces[i] = face
	}
}

func (f *Face) parseFaceLine(state *mqoState) {
	state.skipSpace()

	numVertices, err := state.readPositiveInteger()

	if err != nil {
		//TODO: Report error
		return
	}

	for {
		state.skipSpace()

		name, content, err := state.readParentheticArg()
		if err != nil {
			//TODO: Report error
			return
		}

		if name == "" && err == nil {
			return
		}

		faceFunc := faceArgumentFuncs[name]

		if faceFunc != nil {
			err = faceFunc(f, numVertices, content)

			if err != nil {
				//TODO: Report error
				return
			}
		}
	}
}

var faceArgumentFuncs = map[string]func(f *Face, numVertices int, content string) error{
	"V":  parseFaceVertices,
	"M":  parseFaceMaterial,
	"UV": parseFaceUvCoordinates,
}

func parseFaceVertices(f *Face, numVertices int, content string) error {
	spl := whiteSpaceRegexp.Split(content, 4)

	if len(spl) != numVertices {
		return fmt.Errorf("V parameter does not match specified number of vertices")
	}

	for i := 0; i < len(spl); i++ {
		num, err := strconv.Atoi(spl[i])

		if err != nil {
			//TODO: Report error
			continue
		}

		f.VertexIndices[i] = num
	}

	return nil
}

func parseFaceMaterial(f *Face, numVertices int, content string) error {
	faceState := mqoState{0, 0, []byte(content)}
	index, err := faceState.readPositiveInteger()
	if err != nil {
		return err
	}
	f.MaterialIndex = index
	return nil
}

func parseFaceUvCoordinates(f *Face, numVertices int, content string) error {
	faceState := mqoState{0, 0, []byte(content)}
	processedCoordinate := 0
	f.TextureCoordinates = make([]float32, numVertices*2)

	for processedCoordinate < numVertices*2 {
		faceState.skipSpace()
		coordinate, err := faceState.readFloat()
		if err != nil {
			return err
		}

		f.TextureCoordinates[processedCoordinate] = coordinate
		processedCoordinate++
	}

	return nil
}

func (state *mqoState) readParentheticArg() (string, string, error) {
	name := state.readWord()

	if name == "" {
		return "", "", nil
	}

	if state.readChars("(") == "" {
		return "", "", fmt.Errorf("Expected (")
	}

	content := state.readTo(rune(')'))

	if state.readWord() == "" {
		return "", "", fmt.Errorf("Expected )")
	}

	return name, content, nil
}

func (state *mqoState) readQuote() (string, error) {
	startQuote := state.readWord()

	if startQuote != "\"" && startQuote != "'" {
		return "", fmt.Errorf("Expected quote")
	}

	quoteContent := state.readTo(rune(startQuote[0])) // Read the name
	endQuote := state.readWord()

	if endQuote != startQuote {
		return "", fmt.Errorf("Expected end of quote")
	}

	return quoteContent, nil
}

func (state *mqoState) readPositiveInteger() (int, error) {
	word := state.readChars("0123456789")
	return strconv.Atoi(word)
}

func (state *mqoState) readFloat() (float32, error) {
	word := state.readChars("0123456789.-")
	ret, err := strconv.ParseFloat(word, 32)
	return float32(ret), err
}

func (state *mqoState) readWord() string {
	state.skipSpace()

	for {
		r, size := state.getRune()

		if unicode.IsSpace(r) || size == 0 {
			word := state.data[state.pos:state.cur]
			state.commit()
			return string(word)
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			if state.pos == state.cur {
				state.walk(size)
			}

			word := state.data[state.pos:state.cur]
			state.commit()
			return string(word)
		}

		state.walk(size)
	}
}

func (state *mqoState) readChars(validChars string) string {
	state.skipSpace()

	for {
		r, size := state.getRune()

		if !strings.ContainsAny(string(r), validChars) || size == 0 {
			word := state.data[state.pos:state.cur]
			state.commit()
			return string(word)
		}

		state.walk(size)
	}
}

func (state *mqoState) readTo(char rune) string {
	state.skipSpace()

	for {
		r, size := state.getRune()

		if r == char || size == 0 {
			word := state.data[state.pos:state.cur]
			state.commit()
			return string(word)
		}

		state.walk(size)
	}
}

func (state *mqoState) readLine() string {
	for {
		r, size := state.getRune()

		if string(r) == "\n" || size == 0 {
			word := state.data[state.pos:state.cur]

			if word[len(word)-1] == '\r' {
				word = state.data[state.pos : state.cur-1]
			}

			state.commit()
			return string(word)
		}

		state.walk(size)
	}
}

func (state *mqoState) skipSpace() {
	for {
		r, size := state.getRune()

		if !unicode.IsSpace(r) || size == 0 {
			state.commit()
			return
		}

		state.walk(size)
	}
}

func (state *mqoState) skipToEndBracket(bracketSize int) {
	for {
		r, size := state.getRune()

		if string(r) == "{" {
			bracketSize++
		}

		if string(r) == "}" {
			bracketSize--

			if bracketSize == 0 {
				state.walk(size)
				state.commit()
				return
			}
		}

		if size == 0 {
			state.commit()
			return
		}

		state.walk(size)
	}
}

func (state *mqoState) getRune() (rune, int) {
	return utf8.DecodeRune(state.data[state.cur:])
}

func (state *mqoState) walk(size int) {
	state.cur += size
}

func (state *mqoState) commit() {
	state.pos = state.cur
}
