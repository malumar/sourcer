package generator

import (
	"bytes"
	"fmt"
	"log"
	"malumar/sourcer/annotations"
	"malumar/sourcer/model"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	// Do wew. użytku dla generatorów aby mogły się ze sobą komunikować
	Values map[string]interface{}

	importLines []string

	annotationRegistry annotations.AnnotationRegister
	outputDir          string
	files              map[string]*bytes.Buffer

	packageName   string
	parsedSources *model.ParsedSources
}

func (c *Config) AddValue(key string, value interface{}) {
	if c.Values == nil {
		c.Values = make(map[string]interface{})
	}
	c.Values[key] = value
}

// AddStrValue("true", []string{"klucz1","klucz2","klucz3"...})
func (c *Config) AddMassValue(value interface{}, keys []string) {
	if c.Values == nil {
		c.Values = make(map[string]interface{})
	}
	for _, key := range keys {
		c.Values[key] = value
	}
}

func (c Config) IsSetValue(name string) bool {
	if _, found := c.Values[name]; found {
		return true
	}
	return false
}

func NewConfig(packageName, outputdir string, parsedSources *model.ParsedSources) *Config {
	return NewConfigUsingRegistry(packageName, outputdir, parsedSources, annotations.NewGlobalRegistry())
}

func findPackageName(parsedSources *model.ParsedSources) string {

	for _, t := range parsedSources.Structs {
		if t.PackageName != "" {
			return t.PackageName
		}
	}

	for _, t := range parsedSources.Operations {
		if t.PackageName != "" {
			return t.PackageName
		}
	}

	for _, t := range parsedSources.Interfaces {
		if t.PackageName != "" {
			return t.PackageName
		}
	}

	for _, t := range parsedSources.Enums {
		if t.PackageName != "" {
			return t.PackageName
		}
	}

	return ""
}
func NewConfigUsingRegistry(packageName, outputdir string, parsedSources *model.ParsedSources, register annotations.AnnotationRegister) *Config {

	if packageName == "" {
		packageName = findPackageName(parsedSources)
	}

	return &Config{
		Values:             make(map[string]interface{}),
		files:              make(map[string]*bytes.Buffer),
		packageName:        packageName,
		outputDir:          outputdir,
		parsedSources:      parsedSources,
		annotationRegistry: register,
	}
}

type Generator interface {
	Name() string
	GetAnnotations() []annotations.AnnotationDescriptor
	Generate(config *Config) error
}

func (f *Config) Registry() annotations.AnnotationRegister {
	return f.annotationRegistry
}

// aby skrócić kod
func (f *Config) Operations() []model.Operation {
	return f.ParsedSources().Operations
}

func (f *Config) Interfaces() []model.Interface {
	return f.ParsedSources().Interfaces
}
func (f *Config) AddImport(name ...string) {
	f.importLines = model.AddUnique(f.importLines, name...)
}

func (f Config) GetImports() []string {
	return f.importLines
}

func (f *Config) Enums() []model.Enum {
	return f.ParsedSources().Enums
}

func (f *Config) Structs() []model.Struct {
	return f.ParsedSources().Structs
}

func (f *Config) Typedefs() []model.Typedef {
	return f.ParsedSources().Typedefs
}

// aby skrócić kod

func (f *Config) ParsedSources() *model.ParsedSources {
	return f.parsedSources
}

func (f *Config) OutputDir() string {
	return f.outputDir
}

func (f *Config) Get(name string) *bytes.Buffer {
	if val, found := f.files[name]; !found {
		f.files[name] = &bytes.Buffer{}
		return f.files[name]
	} else {
		return val
	}
}

func (f *Config) FileNames() []string {
	fns := make([]string, 0)
	for v, _ := range f.files {
		fns = append(fns, v)
	}
	return fns
}

func (f *Config) Save(evenIfError bool, filenames ...string) error {
	for i, fn := range filenames {
		log.Printf("Zapisuje %d z %d: %s//%s \n", i, len(fn), f.outputDir, fn)
		err := SaveFileAsGo(evenIfError, f.outputDir, fn, f.Get(fn))
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func (f *Config) SaveAll(evenIfError bool) error {
	return f.Save(evenIfError, f.FileNames()...)
}
func (f *Config) PackageName() string {
	return f.packageName
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

func SaveFileAsGo(evenIfError bool, directory string, filename string, buffer *bytes.Buffer) error {
	if !(strings.HasSuffix(filename, ".go") || strings.HasSuffix(filename, ".go.")) {
		return fmt.Errorf("Plik `%s` nie jest plikiem go - brak sufixu", filename)
	}
	formated, formatError := FormatByGoFmt(buffer)
	if formatError != nil {
		formatError = fmt.Errorf("Błąd formatowania pliku `%s`: %v\n\n---\n%s\n---\n", filename, formatError, buffer.String())
		if !evenIfError {
			return formatError
		}
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(directory), 0751)
		if err != nil {
			return fmt.Errorf("Podczas zapisu pliku jako go - nie udało się utworzyć folderu `%s`:%v", directory, err)
		}
	}

	if !isDirectory(directory) {
		return fmt.Errorf("Wskazana ścieżka do zapisu nie jest katalogiem `%s`", directory)

	}

	w, err := os.Create(filepath.Join(directory, filename))
	if err != nil {
		return err
	}
	defer w.Close()

	if formatError == nil {
		w.Write(formated.Bytes())
	} else {
		w.Write(buffer.Bytes())
	}
	return formatError
}

func FormatByGoFmt(b *bytes.Buffer) (*bytes.Buffer, error) {

	var out bytes.Buffer
	cmd := exec.Command("/bin/which", "gofmt")
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {

		fmt.Println("Nie udało się wywołać which")
		return nil, err
	}
	gofmtFile := strings.TrimSuffix(out.String(), "\n")
	fmt.Println("używam gofmt: " + gofmtFile)
	//cmd = exec.Command("/usr/local/opt/go/bin/gofmt", "/dev/stdin")
	cmd = exec.Command(gofmtFile, "/dev/stdin")
	cmd.Stdin = strings.NewReader(b.String())
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	// b.Reset()
	// b.Write(out.Bytes())
	output := bytes.Buffer{}
	output.Write(out.Bytes())
	return &output, nil
}

func FormatByGoFmtInto(b *bytes.Buffer) error {
	o, err := FormatByGoFmt(b)
	if err != nil {
		return err
	}
	b.Reset()
	b.Write(o.Bytes())
	return nil
}

var generators []Generator = make([]Generator, 0)

func RegisterGenerator(gen Generator) {
	for _, g := range generators {
		if g.Name() == gen.Name() {
			log.Printf("Generator %s jest już zarejestrowany\n", gen.Name())
			return
		}
	}
	generators = append(generators, gen)
}

func GenerateAll(config *Config, generators []Generator) error {
	for i, gen := range generators {
		log.Printf("Generowanie %d z %d %s\n", i, len(generators), gen.Name)
		if err := gen.Generate(config); err != nil {
			log.Printf("Błąd generowania przez generator %s: %v\n", gen.Name, err)
			return err
		}
	}

	return nil
}

func GenerateUsingRegistered(config *Config) error {
	return GenerateAll(config, generators)
}
