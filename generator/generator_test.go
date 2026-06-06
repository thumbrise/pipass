package generator_test

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/thumbrise/pipass/generator"
)

//nolint:cyclop
func TestGenerateFields_ProducesValidGo(t *testing.T) {
	entity := &generator.Entity{
		Name: "TestStruct",
		Fields: []generator.Field{
			{
				Name: "Name",
				Type: reflect.TypeOf(""),
			},
			{
				Name: "Active",
				Type: reflect.TypeOf((*bool)(nil)).Elem(), // *bool
			},
			{
				Name:         "Items",
				Type:         reflect.TypeOf([]struct{}{}),
				ElemTypeName: "Item",
				IsNode:       true,
			},
			{
				Name: "Meta",
				Type: reflect.TypeOf(map[string]string{}),
			},
		},
	}

	f := jen.NewFile("testpkg")
	gen := generator.NewGenerator(f)
	gen.GenerateFields(entity)

	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	code := buf.String()
	t.Logf("Generated code:\n%s", code)

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, "generated_test.go", code, 0)
	if err != nil {
		t.Fatalf("Generated code is not valid Go: %v", err)
	}

	hasType := func(name string) bool {
		for _, decl := range astFile.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if ts.Name.Name == name {
					return true
				}
			}
		}

		return false
	}

	if !hasType("TestStructPass") {
		t.Error("missing interface TestStructPass")
	}

	if !hasType("TestStructPipePass") {
		t.Error("missing struct TestStructPipePass")
	}

	expectedMethods := []string{
		"Name", "SetName",
		"Active", "SetActive",
		"Items", "MapItems", "AppendItems",
		"MetaKey", "SetMetaKey",
	}

	iface := findInterface(astFile, "TestStructPass")
	if iface == nil {
		t.Fatal("interface TestStructPass not found in AST")
	}

	for _, m := range expectedMethods {
		if !interfaceHasMethod(iface, m) {
			t.Errorf("interface TestStructPass missing method %s", m)
		}
	}
}

func findInterface(f *ast.File, name string) *ast.InterfaceType {
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != name {
				continue
			}

			if iface, ok := ts.Type.(*ast.InterfaceType); ok {
				return iface
			}
		}
	}

	return nil
}

func interfaceHasMethod(iface *ast.InterfaceType, name string) bool {
	for _, m := range iface.Methods.List {
		if len(m.Names) > 0 && m.Names[0].Name == name {
			return true
		}
	}

	return false
}
