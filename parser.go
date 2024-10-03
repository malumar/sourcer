package sourcer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"malumar/sourcer/generator"
	"malumar/sourcer/model"
)

var (
	debugAstOfSources = false
)

//const GO_FILES = "^[^\\$]+\\.go$"
const GO_FILES = "^(.*).go$"

// const SKIP_FILES =  "^(zz)?[^\\$][^_]+\\$\\.go$"
const SKIP_FILES = "(^(zz)?[^\\$][^_]+\\$)|(_test)\\.go$"

// Prasuj pliki, których nazwa pasują do wyrażenia w folder dirName
func ParseByRegExp(packageName, dirName string, filenameRegex string) (*generator.Config, error) {
	return ParseSourceDir(packageName, dirName, filenameRegex, nil)
}

// Prasuj wszyskie  pliki *.go  w folder dirName
func ParseGoFilesInDir(packageName, dirname string) (*generator.Config, error) {
	return ParseByRegExp(packageName, dirname, GO_FILES)
}

// Prasuj wszyskie  pliki *.go  z listy files w folder dirName
func ParseSelectedGoFiles(packageName, dirName string, files []string) (*generator.Config, error) {
	return ParseSourceDir(packageName, dirName, GO_FILES, func(name string) bool {
		for _, fn := range files {
			if name == fn {
				return true
			}
		}
		return false
	})
}

func ParseSourceDir(packageName, dirName string, filenameRegex string, matcher func(name string) bool) (*generator.Config, error) {

	if debugAstOfSources {
		dumpFilesInDir(dirName)
	}
	packages, err := parseDir(dirName, filenameRegex, matcher)
	if err != nil {
		log.Printf("error parsing dir %s: %s", dirName, err.Error())
		return nil, err
	}

	v := &astVisitor{
		Imports: map[string]string{},
	}
	for _, aPackage := range packages {
		for _, fileEntry := range sortedFileEntries(aPackage.Files) {
			v.CurrentFilename = fileEntry.key

			appEngineOnly := true
			for _, commentGroup := range fileEntry.file.Comments {
				if commentGroup != nil {
					for _, comment := range commentGroup.List {
						if comment != nil && comment.Text == "// +build !appengine" {
							appEngineOnly = false
						}
					}
				}
			}
			if appEngineOnly {
				ast.Walk(v, &fileEntry.file)
			}
		}
	}

	embedOperationsInStructs(v)

	embedTypedefDocLinesInEnum(v)

	result := model.ParsedSources{
		Structs:    v.Structs,
		Operations: v.Operations,
		Interfaces: v.Interfaces,
		Typedefs:   v.Typedefs,
		Enums:      v.Enums,
		FileName:   make(map[string]*model.ParsedSourcesLink),
	}

	for i, item := range result.Structs {

		if dest, found := result.FileName[item.Filename]; !found {
			result.FileName[item.Filename] = &model.ParsedSourcesLink{
				Structs:    make([]*model.Struct, 0),
				Operations: make([]*model.Operation, 0),
				Interfaces: make([]*model.Interface, 0),
				Typedefs:   make([]*model.Typedef, 0),
				Enums:      make([]*model.Enum, 0),
			}
		} else {
			dest.Structs = append(dest.Structs, &result.Structs[i])
		}
	}

	for i, item := range result.Operations {
		if dest, found := result.FileName[item.Filename]; !found {
			result.FileName[item.Filename] = &model.ParsedSourcesLink{
				Structs:    make([]*model.Struct, 0),
				Operations: make([]*model.Operation, 0),
				Interfaces: make([]*model.Interface, 0),
				Typedefs:   make([]*model.Typedef, 0),
				Enums:      make([]*model.Enum, 0),
			}
		} else {
			dest.Operations = append(dest.Operations, &result.Operations[i])
		}
	}

	for i, item := range result.Interfaces {
		if dest, found := result.FileName[item.Filename]; !found {
			result.FileName[item.Filename] = &model.ParsedSourcesLink{
				Structs:    make([]*model.Struct, 0),
				Operations: make([]*model.Operation, 0),
				Interfaces: make([]*model.Interface, 0),
				Typedefs:   make([]*model.Typedef, 0),
				Enums:      make([]*model.Enum, 0),
			}
		} else {
			dest.Interfaces = append(dest.Interfaces, &result.Interfaces[i])
		}
	}

	for i, item := range result.Typedefs {
		if dest, found := result.FileName[item.Filename]; !found {
			result.FileName[item.Filename] = &model.ParsedSourcesLink{
				Structs:    make([]*model.Struct, 0),
				Operations: make([]*model.Operation, 0),
				Interfaces: make([]*model.Interface, 0),
				Typedefs:   make([]*model.Typedef, 0),
				Enums:      make([]*model.Enum, 0),
			}
		} else {
			dest.Typedefs = append(dest.Typedefs, &result.Typedefs[i])
		}
	}

	for i, item := range result.Enums {
		if dest, found := result.FileName[item.Filename]; !found {
			result.FileName[item.Filename] = &model.ParsedSourcesLink{
				Structs:    make([]*model.Struct, 0),
				Operations: make([]*model.Operation, 0),
				Interfaces: make([]*model.Interface, 0),
				Typedefs:   make([]*model.Typedef, 0),
				Enums:      make([]*model.Enum, 0),
			}
		} else {
			dest.Enums = append(dest.Enums, &result.Enums[i])
		}
	}

	for i := range result.Operations {
		if result.Operations[i].RelatedStruct != nil {
			for j := range result.Structs {
				if result.Structs[j].Name == result.Operations[i].RelatedStruct.TypeName {
					result.Operations[i].StructDefinition = &result.Structs[j]
				}
			}

			//	fmt.Println("sourcer.typename",result.Operations[i].RelatedStruct.PackageName,"#", result.Operations[i].RelatedStruct.TypeName)
		}
	}

	return generator.NewConfig(packageName, dirName, &result), nil
}

func parseSourceFile(srcFilename string) (model.ParsedSources, error) {
	if debugAstOfSources {
		dumpFile(srcFilename)
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, srcFilename, nil, parser.ParseComments)
	if err != nil {
		log.Printf("error parsing src %s: %s", srcFilename, err.Error())
		return model.ParsedSources{}, err
	}
	v := &astVisitor{
		Imports: map[string]string{},
	}
	v.CurrentFilename = srcFilename
	ast.Walk(v, file)

	embedOperationsInStructs(v)

	embedTypedefDocLinesInEnum(v)

	var impts []string
	for _, n := range v.Imports {
		impts = model.AddUnique(impts, n)
	}
	result := model.ParsedSources{
		Imports:    impts,
		Structs:    v.Structs,
		Operations: v.Operations,
		Interfaces: v.Interfaces,
		Typedefs:   v.Typedefs,
		Enums:      v.Enums,
	}
	return result, nil
}

type fileEntry struct {
	key  string
	file ast.File
}

type fileEntries []fileEntry

func (list fileEntries) Len() int {
	return len(list)
}

func (list fileEntries) Less(i, j int) bool {
	return list[i].key < list[j].key
}

func (list fileEntries) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func sortedFileEntries(fileMap map[string]*ast.File) fileEntries {
	var fileEntries fileEntries = make([]fileEntry, 0, len(fileMap))
	for key, file := range fileMap {
		if file != nil {
			fileEntries = append(fileEntries, fileEntry{
				key:  key,
				file: *file,
			})
		}
	}
	sort.Sort(fileEntries)
	return fileEntries
}

func embedOperationsInStructs(visitor *astVisitor) {
	mStructMap := make(map[string]*model.Struct)
	for idx := range visitor.Structs {
		mStructMap[(&visitor.Structs[idx]).Name] = &visitor.Structs[idx]
	}
	for idx := range visitor.Operations {
		mOperation := visitor.Operations[idx]
		if mOperation.RelatedStruct != nil {
			mStruct, ok := mStructMap[(*mOperation.RelatedStruct).TypeName]
			if ok {

				mStruct.Operations = append(mStruct.Operations, &mOperation)
			}
		}
	}

}

func embedTypedefDocLinesInEnum(visitor *astVisitor) {
	for idx, mEnum := range visitor.Enums {
		for _, typedef := range visitor.Typedefs {
			if typedef.Name == mEnum.Name {
				visitor.Enums[idx].DocLines = typedef.DocLines
				break
			}
		}
	}
}

func parseDir(dirName string, filenameRegex string, matcher func(filename string) bool) (map[string]*ast.Package, error) {
	var pattern = regexp.MustCompile(filenameRegex)
	var skipPattern = regexp.MustCompile(SKIP_FILES)
	var err error

	fileSet := token.NewFileSet()
	packageMap, err := parser.ParseDir(
		fileSet,
		dirName,
		func(fi os.FileInfo) bool {

			ok := pattern.MatchString(fi.Name()) && !skipPattern.MatchString(fi.Name())
			if ok && matcher != nil {
				if !matcher(fi.Name()) {
					ok = false
				}
			}

			return ok
		},
		parser.ParseComments)
	if err != nil {
		log.Printf("error parsing dir %s: %s", dirName, err.Error())
		return packageMap, err
	}

	return packageMap, nil
}

func dumpFile(srcFilename string) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, srcFilename, nil, parser.ParseComments)
	if err != nil {
		log.Printf("error parsing src %s: %s", srcFilename, err.Error())
		return
	}
	ast.Print(fileSet, file)
}

func dumpFilesInDir(dirName string) {
	fileSet := token.NewFileSet()
	packageMap, err := parser.ParseDir(
		fileSet,
		dirName,
		nil,
		parser.ParseComments)
	if err != nil {
		log.Printf("error parsing dir %s: %s", dirName, err.Error())
	}
	for _, aPackage := range packageMap {
		for _, file := range aPackage.Files {
			ast.Print(fileSet, file)
		}
	}
}

type astVisitor struct {
	CurrentFilename string
	PackageName     string
	Filename        string
	Imports         map[string]string
	Structs         []model.Struct
	Operations      []model.Operation
	Interfaces      []model.Interface
	Typedefs        []model.Typedef
	Enums           []model.Enum
}

func (v *astVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {

		// package-name is in isolated node
		packageName, ok := extractPackageName(node)
		if ok {
			v.PackageName = packageName
		}

		// extract all imports into a map
		v.extractGenDeclImports(node)

		v.parseAsStruct(node)
		v.parseAsTypedef(node)
		v.parseAsEnum(node)
		v.parseAsInterFace(node)
		v.parseAsOperation(node)

	}
	return v
}

func (v *astVisitor) extractGenDeclImports(node ast.Node) {
	genDecl, ok := node.(*ast.GenDecl)
	if ok {
		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if ok {
				quotedImport := importSpec.Path.Value
				unquotedImport := strings.Trim(quotedImport, "\"")
				init, last := filepath.Split(unquotedImport)
				if init == "" {
					last = init
				}
				v.Imports[last] = unquotedImport
			}
		}
	}
}

func (v *astVisitor) extractGenDeclImportsToSlice(node ast.Node) []string {
	genDecl, ok := node.(*ast.GenDecl)
	var ret []string
	if ok {
		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if ok {
				quotedImport := importSpec.Path.Value
				unquotedImport := strings.Trim(quotedImport, "\"")
				model.AddUnique(ret, unquotedImport)
			}
		}
	}
	return ret
}

func (v *astVisitor) parseAsStruct(node ast.Node) {
	mStruct := extractGenDeclForStruct(node, v.Imports)
	if mStruct != nil {
		mStruct.PackageName = v.PackageName
		mStruct.Filename = v.CurrentFilename
		mStruct.Imports = v.extractGenDeclImportsToSlice(node)
		v.Structs = append(v.Structs, *mStruct)
	}
}

func (v *astVisitor) parseAsTypedef(node ast.Node) {
	mTypedef := extractGenDeclForTypedef(node)
	if mTypedef != nil {
		mTypedef.PackageName = v.PackageName
		mTypedef.Filename = v.CurrentFilename
		mTypedef.Imports = v.extractGenDeclImportsToSlice(node)
		v.Typedefs = append(v.Typedefs, *mTypedef)
	}
}

func (v *astVisitor) parseAsEnum(node ast.Node) {
	mEnum := extractGenDeclForEnum(node)
	if mEnum != nil {
		mEnum.PackageName = v.PackageName
		mEnum.Filename = v.CurrentFilename
		mEnum.Imports = v.extractGenDeclImportsToSlice(node)
		v.Enums = append(v.Enums, *mEnum)
	}
}

func (v *astVisitor) parseAsInterFace(node ast.Node) {
	// if interfaces, get its methods
	mInterface := extractGenDecForInterface(node, v.Imports)
	if mInterface != nil {
		mInterface.PackageName = v.PackageName
		mInterface.Filename = v.CurrentFilename
		mInterface.Imports = v.extractGenDeclImportsToSlice(node)
		v.Interfaces = append(v.Interfaces, *mInterface)
	}
}

func (v *astVisitor) parseAsOperation(node ast.Node) {
	// if mOperation, get its signature
	mOperation := extractOperation(node, v.Imports)
	if mOperation != nil {
		mOperation.PackageName = v.PackageName
		mOperation.Filename = v.CurrentFilename
		mOperation.Imports = v.extractGenDeclImportsToSlice(node)
		v.Operations = append(v.Operations, *mOperation)
	}
}

func extractGenDeclForStruct(node ast.Node, imports map[string]string) *model.Struct {
	genDecl, ok := node.(*ast.GenDecl)
	if ok {
		// Continue parsing to see if it a struct
		mStruct := extractSpecsForStruct(genDecl.Specs, imports)
		if mStruct != nil {
			// Docline of struct (that could contain annotations) appear far before the details of the struct
			mStruct.DocLines = extractComments(genDecl.Doc)
			mStruct.OnlyComments = extractSimpleComment(genDecl.Doc)

			return mStruct
		}
	}
	return nil
}

func extractGenDeclForTypedef(node ast.Node) *model.Typedef {
	genDecl, ok := node.(*ast.GenDecl)
	if ok {
		// Continue parsing to see if it a struct
		mTypedef := extractSpecsForTypedef(genDecl.Specs)
		if mTypedef != nil {
			mTypedef.DocLines = extractComments(genDecl.Doc)
			mTypedef.OnlyComments = extractSimpleComment(genDecl.Doc)

			return mTypedef
		}
	}
	return nil
}

func extractGenDeclForEnum(node ast.Node) *model.Enum {
	genDecl, ok := node.(*ast.GenDecl)
	if ok {
		// Continue parsing to see if it is an enum
		// Docs live in the related typedef
		return extractSpecsForEnum(genDecl.Specs)
	}
	return nil
}

func extractGenDecForInterface(node ast.Node, imports map[string]string) *model.Interface {
	genDecl, ok := node.(*ast.GenDecl)
	if ok {
		// Continue parsing to see if it an interface
		mInterface := extractSpecsForInterface(genDecl.Specs, imports)
		if mInterface != nil {
			// Docline of interface (that could contain annotations) appear far before the details of the struct
			mInterface.DocLines = extractComments(genDecl.Doc)
			mInterface.OnlyComments = extractSimpleComment(genDecl.Doc)
			return mInterface
		}
	}
	return nil
}

func extractSpecsForStruct(specs []ast.Spec, imports map[string]string) *model.Struct {
	if len(specs) >= 1 {
		typeSpec, ok := specs[0].(*ast.TypeSpec)
		if ok {
			structType, ok := typeSpec.Type.(*ast.StructType)
			if ok {
				return &model.Struct{
					Name:   typeSpec.Name.Name,
					Fields: extractFieldList(structType.Fields, imports),
				}
			}
		}
	}
	return nil
}

func extractSpecsForEnum(specs []ast.Spec) *model.Enum {
	typeName, ok := extractEnumTypeName(specs)
	if ok {
		mEnum := model.Enum{
			Name:         typeName,
			EnumLiterals: []model.EnumLiteral{},
		}
		for _, spec := range specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if ok {
				enumLiteral := model.EnumLiteral{
					Name: valueSpec.Names[0].Name,
				}
				for _, value := range valueSpec.Values {
					basicLit, ok := value.(*ast.BasicLit)
					if ok {
						enumLiteral.Value = strings.Trim(basicLit.Value, "\"")
						break
					}
				}
				mEnum.EnumLiterals = append(mEnum.EnumLiterals, enumLiteral)
			}
		}
		return &mEnum
	}
	return nil
}

func extractEnumTypeName(specs []ast.Spec) (string, bool) {
	for _, spec := range specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if ok {
			if valueSpec.Type != nil {
				for _, name := range valueSpec.Names {
					ident, ok := valueSpec.Type.(*ast.Ident)
					if ok {
						if name.Obj.Kind == ast.Con {
							return ident.Name, true
						}
					}
				}
			}
		}
	}
	return "", false
}

func extractSpecsForInterface(specs []ast.Spec, imports map[string]string) *model.Interface {
	if len(specs) >= 1 {
		typeSpec, ok := specs[0].(*ast.TypeSpec)
		if ok {
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if ok {
				return &model.Interface{
					Name:    typeSpec.Name.Name,
					Methods: extractInterfaceMethods(interfaceType.Methods, imports),
				}
			}
		}
	}
	return nil
}

func extractPackageName(node ast.Node) (string, bool) {
	file, ok := node.(*ast.File)
	if ok {
		if file.Name != nil {
			return file.Name.Name, true
		}
	}
	return "", ok
}

func extractOperation(node ast.Node, imports map[string]string) *model.Operation {
	funcDecl, ok := node.(*ast.FuncDecl)
	if ok {
		mOperation := model.Operation{
			DocLines:     extractComments(funcDecl.Doc),
			OnlyComments: extractSimpleComment(funcDecl.Doc),
		}

		if funcDecl.Recv != nil {
			fields := extractFieldList(funcDecl.Recv, imports)
			if len(fields) >= 1 {
				mOperation.RelatedStruct = &(fields[0])
			}
		}

		if funcDecl.Name != nil {
			mOperation.Name = funcDecl.Name.Name
		}

		if funcDecl.Type.Params != nil {
			mOperation.InputArgs = extractFieldList(funcDecl.Type.Params, imports)
		}

		if funcDecl.Type.Results != nil {
			mOperation.OutputArgs = extractFieldList(funcDecl.Type.Results, imports)
		}
		return &mOperation
	}
	return nil
}

func extractSpecsForTypedef(specs []ast.Spec) *model.Typedef {
	if len(specs) >= 1 {
		typeSpec, ok := specs[0].(*ast.TypeSpec)
		if ok {
			mTypedef := model.Typedef{
				Name: typeSpec.Name.Name,
			}
			ident, ok := typeSpec.Type.(*ast.Ident)
			if ok {
				mTypedef.Type = ident.Name
			}
			return &mTypedef
		}
	}
	return nil
}

func isRealComment(s string) bool {
	slashCount := 0

	for _, c := range s {
		switch c {

		case ' ', '\t':
			break

		case '/':
			slashCount++
			if slashCount > 2 {
				return true
			}
			break

		case '@':
			return false
			break

		default:
			return true
		}
	}

	return true
}

// wysciagnięcie komentarzy które nie są anotacjami
func extractSimpleComment(commentGroup *ast.CommentGroup) []string {
	lines := []string{}
	if commentGroup != nil {
		for _, comment := range commentGroup.List {
			if isRealComment(comment.Text) {
				lines = append(lines, comment.Text)
			}
		}
	}
	return lines
}
func extractComments(commentGroup *ast.CommentGroup) []string {
	lines := []string{}
	if commentGroup != nil {
		for _, comment := range commentGroup.List {
			lines = append(lines, comment.Text)
		}
	}
	return lines
}

func extractTag(basicLit *ast.BasicLit) string {
	if basicLit != nil {
		return basicLit.Value
	}
	return ""
}

func extractFieldList(fieldList *ast.FieldList, imports map[string]string) []model.Field {
	mFields := []model.Field{}
	if fieldList != nil {
		for _, field := range fieldList.List {
			mFields = append(mFields, extractFields(field, imports)...)
		}
	}
	return mFields
}

func extractInterfaceMethods(fieldList *ast.FieldList, imports map[string]string) []model.Operation {

	methods := []model.Operation{}
	for _, field := range fieldList.List {
		if len(field.Names) > 0 {
			funcType, ok := field.Type.(*ast.FuncType)
			if ok {

				methods = append(methods, model.Operation{
					DocLines:     extractComments(field.Doc),
					OnlyComments: extractSimpleComment(field.Doc),
					Name:         field.Names[0].Name,
					InputArgs:    extractFieldList(funcType.Params, imports),
					OutputArgs:   extractFieldList(funcType.Results, imports),
				})
			}
		}
	}
	return methods
}

func extractFields(field *ast.Field, imports map[string]string) []model.Field {
	mFields := []model.Field{}
	if field != nil {
		//fmt.Println("sourcer 1:", field.Tag)
		if len(field.Names) == 0 {
			f, ok := extractField(field, imports)
			if ok {
				mFields = append(mFields, f)
			}
		} else {
			// A single field can refer to multiple: example: x,y int -> x int, y int
			for _, name := range field.Names {
				field, ok := extractField(field, imports)
				if ok {
					field.Name = name.Name
					mFields = append(mFields, field)
				}
			}
		}
	}
	return mFields
}

func extractField(field *ast.Field, imports map[string]string) (model.Field, bool) {
	mField := model.Field{
		DocLines:     extractComments(field.Doc),
		OnlyComments: extractSimpleComment(field.Doc),
		CommentLines: extractComments(field.Comment),
		Tag:          extractTag(field.Tag),
	}

	if extractSliceField(field, &mField, imports) {
		return mField, true
	}

	if extractMapField(field, &mField, imports) {
		return mField, true
	}

	if extractPointerField(field, &mField, imports) {
		return mField, true
	}

	if extractIdentField(field, &mField, imports) {
		return mField, true
	}

	if extractSelectorField(field, &mField, imports) {
		return mField, true
	}

	// Dodane 02.09.2024 - moze nie byc zgodne z corecrm
	if extractInterfaceField(field, &mField, imports) {
		return mField, true
	}

	log.Printf("*** Could not understand field '%+v' --> %v", field.Type, mField)

	return mField, false
}

func extractSliceField(field *ast.Field, mField *model.Field, imports map[string]string) bool {
	arrayType, ok := field.Type.(*ast.ArrayType)
	if ok {
		mField.IsSlice = true
		if extractSliceSelectorField(arrayType, mField, imports) {
			return true
		}
		if extractSlicePointerField(arrayType, mField, imports) {
			return true
		}
	}
	return false
}

func extractInterfaceField(field *ast.Field, mField *model.Field, imports map[string]string) bool {

	if _, ok := field.Type.(*ast.InterfaceType); ok {
		mField.IsInterface = true
		mField.TypeName = "interface{}"
		return true
	}
	return false
}

func extractSliceSelectorField(arrayType *ast.ArrayType, mField *model.Field, imports map[string]string) bool {
	ident, ok := arrayType.Elt.(*ast.Ident)
	if ok {
		mField.TypeName = ident.Name
		return true
	}

	selectorExpr, ok := arrayType.Elt.(*ast.SelectorExpr)
	if ok {
		ident, ok = selectorExpr.X.(*ast.Ident)
		if ok {
			mField.TypeName = fmt.Sprintf("%s.%s", ident.Name, selectorExpr.Sel.Name)
			mField.PackageName = imports[ident.Name]
			return true
		}
	}
	return false
}

func extractSlicePointerField(arrayType *ast.ArrayType, mField *model.Field, imports map[string]string) bool {
	starExpr, ok := arrayType.Elt.(*ast.StarExpr)
	if ok {
		if ok {
			ident, ok := starExpr.X.(*ast.Ident)
			if ok {
				mField.TypeName = ident.Name
				mField.IsPointer = true
				return true
			}
		}

		selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
		if ok {
			ident, ok := selectorExpr.X.(*ast.Ident)
			if ok {
				mField.PackageName = imports[ident.Name]
				mField.IsPointer = true
				mField.TypeName = fmt.Sprintf("%s.%s", ident.Name, selectorExpr.Sel.Name)
				return true
			}
		}
	}
	return false
}

func extractMapField(field *ast.Field, mField *model.Field, imports map[string]string) bool {
	mapKey := ""
	mapValue := ""

	mapType, ok := field.Type.(*ast.MapType)
	if ok {
		{
			key, ok := mapType.Key.(*ast.Ident)
			if ok {
				mapKey = key.Name
			}
		}
		{
			value, ok := mapType.Value.(*ast.Ident)
			if ok {
				mapValue = value.Name
			}
		}
	}
	if mapKey != "" && mapValue != "" {
		mField.TypeName = fmt.Sprintf("map[%s]%s", mapKey, mapValue)
		return true
	}

	return false
}

func extractPointerField(field *ast.Field, mField *model.Field, imports map[string]string) bool {
	{
		starExpr, ok := field.Type.(*ast.StarExpr)
		if ok {
			ident, ok := starExpr.X.(*ast.Ident)
			if ok {
				mField.TypeName = ident.Name
				mField.IsPointer = true
				return true
			}

			selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
			if ok {
				ident, ok = selectorExpr.X.(*ast.Ident)
				if ok {
					mField.TypeName = fmt.Sprintf("%s.%s", ident.Name, selectorExpr.Sel.Name)
					mField.IsPointer = true
					mField.PackageName = imports[ident.Name]
					return true
				}
			}
		}
	}
	return false
}

func extractIdentField(field *ast.Field, mField *model.Field, imports map[string]string) bool {
	ident, ok := field.Type.(*ast.Ident)
	if ok {
		mField.TypeName = ident.Name
		return true
	}
	return false
}

func extractSelectorField(field *ast.Field, mField *model.Field, imports map[string]string) bool {
	selectorExpr, ok := field.Type.(*ast.SelectorExpr)
	if ok {
		ident, ok := selectorExpr.X.(*ast.Ident)
		if ok {
			mField.Name = ident.Name
			mField.TypeName = fmt.Sprintf("%s.%s", ident.Name, selectorExpr.Sel.Name)
			mField.PackageName = imports[ident.Name]
			return true
		}
	}
	return false
}
