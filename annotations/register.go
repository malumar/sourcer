package annotations


var registeredGlobalAnnotations []AnnotationDescriptor = []AnnotationDescriptor{}


func Ok(name string) ValidationFunc {
	return func(annot Annotation) bool {
		return annot.Name == annot.Name
	}
}

func ClearRegisteredAnnotations() {
	registeredGlobalAnnotations = []AnnotationDescriptor{}
}

func RegisterAnnotation(name string, paramNames []string, validator ValidationFunc) {
	RegisterAnnotationExt(false, name, paramNames, validator)
}

// Odnotuj wspierająca anotacja do wlasnego parsowania (bez nawiasów) wymagane fullline=true
func RegisterAnnotationLine(name string, paramNames []string, validator ValidationFunc) {
	RegisterAnnotationExt(true, name, paramNames, validator)
}

// Odnotuj wspierająca anotacja do wlasnego parsowania (bez nawiasów) wymagane fullline=true
// @Anotacja
func RegisterAnnotationExt(fullLine bool, name string, paramNames []string, validator ValidationFunc) {

	if validator == nil {
		validator = Ok(name)
	}
	if paramNames == nil {
		paramNames = []string{}
	}

	registeredGlobalAnnotations = append(
		registeredGlobalAnnotations,
		AnnotationDescriptor{
			Name:       name,
			ParamNames: paramNames,
			Validator:  validator,
			FullLine:   fullLine,
		},
	)
}

func RegisterAnnotations(ann ...AnnotationDescriptor) {
	registeredGlobalAnnotations = append(registeredGlobalAnnotations, ann...)
}