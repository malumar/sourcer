package model

import (
	"fmt"
	"strings"
)

//go:generate golangAnnotations -input-dir .
func AddUnique(arr []string, values ...string) []string {
	for _, value := range values {

		if strings.TrimSpace(value) == "" {
			continue
		}
		for _, t := range arr {
			if t == value {
				continue
			}
		}
		if arr == nil {
			arr = make([]string, 0)
		}

		arr = append(arr, value)
	}
	return arr
}

type ParsedSourcesLink struct {
	Structs    []*Struct    `json:"structs,omitempty"`
	Operations []*Operation `json:"operations,omitempty"`
	Interfaces []*Interface `json:"interfaces,omitempty"`
	Typedefs   []*Typedef   `json:"typedefs,omitempty"`
	Enums      []*Enum      `json:"enums,omitempty"`
	Imports    []*string    `json:"imports"`
}

// @JsonStruct()
type ParsedSources struct {
	Structs    []Struct    `json:"structs,omitempty"`
	Operations []Operation `json:"operations,omitempty"`
	Interfaces []Interface `json:"interfaces,omitempty"`
	Typedefs   []Typedef   `json:"typedefs,omitempty"`
	Enums      []Enum      `json:"enums,omitempty"`
	Imports    []string    `json:"imports"`

	FileName map[string]*ParsedSourcesLink
}

func inStrSlice(value string, slice []string) bool {
	for _, n := range slice {
		if n == value {
			return true
		}
	}
	return false
}

// Nazwy plików, w których znaleźliśmy definicje
func (p ParsedSources) FileNames() []string {
	ret := make([]string, 0)
	for k, _ := range p.FileName {
		ret = append(ret, k)
	}
	return ret
}

// @JsonStruct()
type Operation struct {
	Imports          []string `json:"imports"`
	PackageName      string   `json:"packageName,omitempty"`
	Filename         string   `json:"filename,omitempty"`
	DocLines         []string `json:"docLines,omitempty"`
	RelatedStruct    *Field   `json:"relatedStruct,omitempty"` // optional
	StructDefinition *Struct
	Name             string   `json:"name"`
	InputArgs        []Field  `json:"inputArgs,omitempty"`
	OutputArgs       []Field  `json:"outputArgs,omitempty"`
	CommentLines     []string `json:"commentLines,omitempty"`
	OnlyComments     []string `json:"onlyComments,omitempty"`
}

func (o Operation) FindOutputFieldByName(name string) *Field {

	for i, _ := range o.OutputArgs {
		field := &o.OutputArgs[i]
		if field.Name == name {
			return field
		}
	}

	return nil

}
func (o Operation) FindInputFieldByName(name string) *Field {
	for i, _ := range o.InputArgs {
		field := &o.InputArgs[i]
		if field.Name == name {
			return field
		}
	}

	return nil

}
func (o Operation) HaveReturn() bool {
	return len(o.OutputArgs) > 0
}

func (o Operation) HaveParams() bool {
	return len(o.InputArgs) > 0
}

// @JsonStruct()
type Struct struct {
	// możesz sobie coś dopisać (markery to jakieś własne flagi)
	Markers map[string]string
	// w data możesz zbindować co chcesz
	Data interface{}

	PackageName  string       `json:"packageName"`
	Imports      []string     `json:"imports"`
	Filename     string       `json:"filename"`
	DocLines     []string     `json:"docLines,omitempty"`
	Name         string       `json:"name"`
	Fields       []Field      `json:"fields,omitempty"`
	Operations   []*Operation `json:"operations,omitempty"`
	CommentLines []string     `json:"commentLines,omitempty"`
	OnlyComments []string     `json:"onlyComments,omitempty"`
}

func (s Struct) HaveData() bool {
	return s.Data != nil
}

func (s *Struct) SetMarker(name string, value string) {
	if s.Markers == nil {
		s.Markers = make(map[string]string)
	}
	s.Markers[name] = value
}

func (s *Struct) HaveMarkerEq(name string, value string) bool {
	if s.Markers == nil {
		return false
	}
	if val, found := s.Markers[name]; found {
		return val == value
	}
	return false
}

func (s Struct) HaveMarker(name string) bool {
	if s.Markers == nil {
		return false
	}
	if _, found := s.Markers[name]; found {
		return true
	}

	return false
}

func (s Struct) FindFieldByName(name string) *Field {
	for i, f := range s.Fields {
		if name == f.Name {
			return &s.Fields[i]
		}
	}
	return nil
}

func (s Struct) IsFieldByNameExists(name string) bool {
	return s.FindFieldByName(name) != nil
}

func (s Struct) FindOperationByName(name string) *Operation {
	for i, f := range s.Operations {
		if name == f.Name {
			return s.Operations[i]
		}
	}
	return nil
}

func (s Struct) IsOperationByNameExists(name string) bool {
	return s.FindOperationByName(name) != nil
}

// @JsonStruct()
type Interface struct {
	Imports []string `json:"imports"`

	PackageName  string      `json:"packageName"`
	Filename     string      `json:"filename"`
	DocLines     []string    `json:"docLines,omitempty"`
	Name         string      `json:"name"`
	Methods      []Operation `json:"methods,omitempty"`
	CommentLines []string    `json:"commentLines,omitempty"`
	OnlyComments []string    `json:"onlyComments,omitempty"`
}

func (i Interface) FindMethodByName(name string) *Operation {
	for x, f := range i.Methods {
		if name == f.Name {
			return &i.Methods[x]
		}
	}
	return nil
}

func (s Interface) IsMethodByNameExists(name string) bool {
	return s.FindMethodByName(name) != nil
}

// @JsonStruct()
type Field struct {
	PackageName  string   `json:"packageName,omitempty"`
	DocLines     []string `json:"docLines,omitempty"`
	Name         string   `json:"name,omitempty"`
	TypeName     string   `json:"typeName,omitempty"`
	IsSlice      bool     `json:"isSlice,omitempty"`
	IsInterface  bool     `json:"isSlice,omitempty"`
	IsPointer    bool     `json:"isPointer,omitempty"`
	Tag          string   `json:"tag,omitempty"`
	CommentLines []string `json:"commentLines,omitempty"`
	OnlyComments []string `json:"onlyComments,omitempty"`
	// wlasne dane uzywane przez dany moduł
	StrValues map[string]string
}

// jest currentPackageName != nazwy pakietu z ktorego pochodzi pole to:
// - user []*package.User
// inaczej:
// - user []*User
func (f Field) AsParamDefinition(currentPackageName string) string {
	var sb strings.Builder

	sb.WriteString(f.Name)
	sb.WriteRune(' ')
	if f.IsSlice {
		sb.WriteString("[]")
	}
	if f.IsPointer {
		sb.WriteString("*")
	}

	sb.WriteString(f.TypeNameWithPackageNameIfNotSame(currentPackageName))

	return sb.String()
}

func (f Field) AsParamDefinitionCutPointer(currentPackageName string) string {
	return strings.Replace(f.AsParamDefinition(currentPackageName), "*", "", -1)
}
func (f Field) TypeNameWithPackageNameIfNotSame(packageName string) string {
	if packageName != f.PackageName {
		switch f.TypeName {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "complex64", "complex128",
			"string", "bytes", "interface{}", "rune", "byte":
			return f.TypeName
		}
		//		fmt.Println("#####", f.TypeName)
		if f.PackageName == "" {
			return f.TypeName
		}
		return fmt.Sprintf("%s.%s", f.PackageName, f.TypeName)
	} else {
		return f.TypeName
	}
}

func (f Field) TypeNameAsDefinition() string {

	return fmt.Sprintf("%s.%s", f.PackageName, f.TypeName)
}

func (f *Field) HasValueEq(key, find string) bool {
	if f.StrValues == nil {
		return false
	}
	if val, found := f.StrValues[key]; found {
		return find == val
	} else {
		return false
	}
}
func (f *Field) HasValue(key string) bool {
	if f.StrValues == nil {
		return false
	}
	if _, found := f.StrValues[key]; found {
		return true
	} else {
		return false
	}
}
func (f *Field) AddStrValue(key, value string) {
	if f.StrValues == nil {
		f.StrValues = make(map[string]string)
	}
	f.StrValues[key] = value
}

// AddStrValue("true", []string{"klucz1","klucz2","klucz3"...})
func (f *Field) AddMassStrValue(value string, keys []string) {
	if f.StrValues == nil {
		f.StrValues = make(map[string]string)
	}
	for _, key := range keys {
		f.StrValues[key] = value
	}
}

// @JsonStruct()
type Typedef struct {
	Imports []string `json:"imports"`

	PackageName  string   `json:"packageName"`
	Filename     string   `json:"filename"`
	DocLines     []string `json:"docLines,omitempty"`
	Name         string   `json:"name"`
	Type         string   `json:"type,omitempty"`
	OnlyComments []string `json:"onlyComments,omitempty"`
}

// @JsonStruct()
type Enum struct {
	Imports []string `json:"imports"`

	PackageName  string        `json:"packageName"`
	Filename     string        `json:"filename"`
	DocLines     []string      `json:"docLines,omitempty"`
	Name         string        `json:"name,omitempty"`
	EnumLiterals []EnumLiteral `json:"enumLiterals,omitempty"`
	CommentLines []string      `json:"commentLines,omitempty"`
	OnlyComments []string      `json:"onlyComments,omitempty"`
}

func (e Enum) FindEnumLiteralByName(name string) *EnumLiteral {
	for i, l := range e.EnumLiterals {
		if name == l.Name {
			return &e.EnumLiterals[i]
		}
	}
	return nil
}

func (e Enum) FindEnumLiteralByValue(value string) *EnumLiteral {
	for i, l := range e.EnumLiterals {
		if value == l.Value {
			return &e.EnumLiterals[i]
		}
	}
	return nil
}

func (s Enum) IsEnumLiteralByValueExists(value string) bool {
	return s.FindEnumLiteralByValue(value) != nil
}

// @JsonStruct()
type EnumLiteral struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}
