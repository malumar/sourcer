package model

import "fmt"

// Pochodzi z daogen (Jako Field) - ale warto zaimplementować globalnie
type ExtField struct {
	FullInfo         *Field
	CommentLines     []string
	CommentRightSide []string
	Annotations      []string
}

type StructInfo struct {
	ExtFromStruct    string
	Table            string
	Model            string
	Struct           *Struct
	CommentLines     []string
	Annotations      []string
	CommentRightSide []string
	Fields           []ExtField
}

// Moze sluzyc do powiazanie struktur e sobą
type LinkedStructs struct {
	items []*StructInfo
	// do komunikacji między modułami i twojego wew. użytku
	Values map[string]interface{}
}

func NewLinkedStructs() *LinkedStructs {
	return &LinkedStructs{
		items: make([]*StructInfo, 0),
		Values: make(map[string]interface{},),
	}
}
func (c *LinkedStructs) AddValue(key string, value interface{}) {
	if c.Values == nil {
		c.Values = make(map[string]interface{})
	}
	c.Values[key] = value
}

func (c *LinkedStructs) GetValue(key string) string {
	if val,found := c.Values[key]; found {
		return fmt.Sprintf("%v", val)
	}

	return ""
}

// AddStrValue("true", []string{"klucz1","klucz2","klucz3"...})
func (c *LinkedStructs) AddMassValue(value interface{}, keys []string) {
	if c.Values == nil {
		c.Values = make(map[string]interface{})
	}
	for _,key:=range keys {
		c.Values[key] = value
	}
}


func (this *LinkedStructs) HaveValue(name string) bool {
	if _,found := this.Values[name]; found {
		return true
	}

	return false
}

func (this *LinkedStructs) GetItems() []*StructInfo {
	return this.items
}

func (this *LinkedStructs) Append(s *StructInfo) {
	this.items = append(this.items, s)
}
func (this *LinkedStructs) FindStructByTable(name string) *StructInfo {
	for i, _ := range this.items {
		if this.items[i].Table == name {
			return this.items[i]
		}
	}
	return nil
}

func (this *LinkedStructs) FindStructByModel(name string) *StructInfo {
	for i, _ := range this.items {
		if this.items[i].Model == name {
			return this.items[i]
		}
	}
	return nil
}
