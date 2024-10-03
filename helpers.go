package sourcer

import (
	"fmt"
	"log"
	"malumar/sourcer/annotations"
	"malumar/sourcer/model"
	"os"
	"path/filepath"
	"strings"
)

// isDirectory reports whether the named file is a directory.
func IsDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

func ErrorInAnno(fileName string, ann annotations.Annotation, format string, args ...interface{}) error {
	return fmt.Errorf(fmt.Sprintf("Błąd w pliku %s  -> Anotacja %v [%+v]: %s", fileName, ann.Name, ann, format), args...)
}

func ErrorInOperation(op *model.Operation, annotName string, format string, args ...interface{}) error {
	return fmt.Errorf(fmt.Sprintf("Błąd w pliku %s  -> operacja %v @%s: %s", op.Filename, op.Name, annotName, format), args...)
}

func IfErrorPrint(err error, format string, args ...interface{}) {
	if err != nil {
		fmt.Printf("\nZakończenie aplikacji z powodu błędu %v\n", err)
		fmt.Printf(format, args...)
		fmt.Println("")
		os.Exit(2)
	}
}

func InStrSlice(value string, slice []string) bool {
	for _, n := range slice {
		if n == value {
			return true
		}
	}
	return false
}

// Czy w tablicy $slice wystepują zmienne $values?
// przed porównaniem jeśli $beforeCompare jest != nil wykonaj na elemencie $beforeCompare($slice[x])
// $allowNotListed czy jest ok jeśli znajdziemy elementy nie pasujące do $values
// $min min wymagana liczba odnaleziony elementów
// $max min wymagana liczba elementów
func InStrSliceRequire(slice []string, beforeCompare func(s string) string, allowNotListed bool, min int, max int, values ...string) (err error, foundedItems []string) {

	for _, n := range slice {

		if beforeCompare != nil {
			n = beforeCompare(n)
		}

		found := InStrSlice(n, values)

		if !found {
			if !allowNotListed {
				return fmt.Errorf("Nie rozpoznana wartość %s (dozwolone to %v) ", n, values), nil
			}
		} else {
			foundedItems = AddUnique(foundedItems, n)
		}

	}

	retCount := len(foundedItems)
	if (min > 0 && retCount < min) || (max > 0 && retCount > max) {
		return fmt.Errorf("Wymagaliśmy od %d do %d wartości, mamy %v z dozwolonych %v", min, max, foundedItems, values), nil
	}

	return nil, foundedItems
}

func GetOnlyFileNames(filepaths []string) []string {
	ret := make([]string, 0)
	for _, n := range filepaths {
		fn := filepath.Base(n)
		if strings.TrimSpace(fn) != "" && !InStrSlice(fn, ret) {
			ret = append(ret, fn)
		}
	}
	return ret
}

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
