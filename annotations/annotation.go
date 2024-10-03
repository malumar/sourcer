package annotations

import "strings"

type AnnotationRegister interface {
	ResolveAnnotations(annotationDocline []string) []Annotation
	ResolveAnnotationByName(annotationDocline []string, name string) (Annotation, bool)
	ResolveAnnotation(annotationDocline string) (Annotation, bool)
	ResolveAllAnnotationByName(annotationDocline []string, name string) ([]Annotation, bool)
	ResolveFullLineAllAnnotationByName(annotationDocline []string, name string) ([]string, bool)
}

type annotationRegistry struct {
	descriptors []AnnotationDescriptor
}

func NewRegistry(descriptors []AnnotationDescriptor) AnnotationRegister {
	return &annotationRegistry{
		descriptors: descriptors,
	}
}

func NewGlobalRegistry() AnnotationRegister {
	return &annotationRegistry{
		descriptors: registeredGlobalAnnotations,
	}
}

type Annotation struct {
	Name       string
	// dla wsparcia anotacij bez nawiasow, tj
	// @Anotacja
	LineValue  string
	IsFullLine bool
	Param      string
	Attributes map[string]string
}

func (a Annotation) IsFullLineWithContent() bool {
	return a.IsFullLine && strings.TrimSpace(a.LineValue) != ""
}

func (a Annotation) IsSetAttribute(name string) bool {
	if _, ok := a.Attributes[name]; ok {
		return true
	}
	return false
}

type ValidationFunc func(annot Annotation) bool

type AnnotationDescriptor struct {
	FullLine bool

	// jeżeli full line to możemy także miec pola
	IsField bool
	MinFields int
	MaxFields int


	Name       string
	ParamNames []string
	Validator  ValidationFunc
}

func (ar *annotationRegistry) ResolveAnnotations(annotationDocline []string) []Annotation {
	annotations := []Annotation{}
	for _, line := range annotationDocline {
		ann, ok := ar.ResolveAnnotation(strings.TrimSpace(line))
		if ok {
			annotations = append(annotations, ann)
		}
	}
	return annotations
}

func (ar *annotationRegistry) ResolveAnnotationByName(annotationDocline []string, name string) (Annotation, bool) {
	for _, line := range annotationDocline {
		ann, ok := ar.ResolveAnnotation(strings.TrimSpace(line))
		if ok && ann.Name == name {
			return ann, true
		}
	}
	return Annotation{}, false
}

func (ar *annotationRegistry) ResolveAnnotation(annotationDocline string) (Annotation, bool) {
	for _, descriptor := range ar.descriptors {
		ann, err := parseAnnotation(annotationDocline)
		if err != nil {
			continue
		}

		if ann.Name != descriptor.Name {
			continue
		}

		ok := descriptor.Validator(ann)
		if !ok {
			continue
		}

		return ann, true
	}
	return Annotation{}, false
}

func (ar *annotationRegistry) ResolveAllAnnotationByName(annotationDocline []string, name string) ([]Annotation, bool) {

	res := ar.ResolveAnnotations(annotationDocline)
	if len(res) == 0 {
		return nil, false
	}
	var af []Annotation
	for _,a:=range res {
		if a.Name == name {
			if af == nil {
				af = make([]Annotation,0)
			}
			af = append(af, a)
		}
	}

	if len(af) > 0 {
		return af, true
	}
	return nil, false
}

func (ar *annotationRegistry) ResolveFullLineAllAnnotationByName(annotationDocline []string, name string) ([]string, bool) {
	res := ar.ResolveAnnotations(annotationDocline)
	if len(res) == 0 {
		return nil, false
	}
	var af []string
	for _,a:=range res {
		if a.Name == name && a.IsFullLine && a.LineValue != "" {
			if af == nil {
				af = make([]string,0)
			}
			af = append(af, a.LineValue)
		}
	}

	if len(af) > 0 {
		return af, true
	}
	return nil, false
}