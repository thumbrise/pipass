package generator

import (
	"errors"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"
)

var ErrRender = errors.New("render error")

type Field struct {
	Name         string
	Type         reflect.Type
	ElemTypeName string
	IsNode       bool
}

type Entity struct {
	Name   string
	Fields []Field
}

type Generator struct {
	file *jen.File
}

func NewGenerator(file *jen.File) *Generator {
	if file == nil {
		panic("jen.File must not be nil")
	}

	return &Generator{file: file}
}

func (g *Generator) GenerateFields(root *Entity) {
	interfaceName := root.Name + "Pass"
	structName := root.Name + "PipePass"

	generateInterface(g.file, interfaceName, root)
	generateStruct(g.file, structName, root)
	generateConstructor(g.file, structName)
	generateAccessors(g.file, structName, root)
}

func privateName(name string) string {
	if len(name) == 0 {
		return ""
	}

	return strings.ToLower(name[:1]) + name[1:]
}

//nolint:cyclop,exhaustive,funlen
func typeToStatement(t reflect.Type) *jen.Statement {
	// Check named types BEFORE Kind switch to preserve type identity
	// (e.g. Score -> testdata.Score, not float64)
	if pkg := t.PkgPath(); pkg != "" {
		return jen.Qual(pkg, t.Name())
	}

	if name := t.Name(); name != "" {
		return jen.Id(name)
	}

	switch t.Kind() {
	case reflect.Pointer:
		return jen.Op("*").Add(typeToStatement(t.Elem()))
	case reflect.Slice:
		return jen.Index().Add(typeToStatement(t.Elem()))
	case reflect.Map:
		return jen.Map(typeToStatement(t.Key())).Add(typeToStatement(t.Elem()))
	case reflect.String:
		return jen.String()
	case reflect.Bool:
		return jen.Bool()
	case reflect.Int:
		return jen.Int()
	case reflect.Int8:
		return jen.Int8()
	case reflect.Int16:
		return jen.Int16()
	case reflect.Int32:
		return jen.Int32()
	case reflect.Int64:
		return jen.Int64()
	case reflect.Uint:
		return jen.Uint()
	case reflect.Uint8:
		return jen.Uint8()
	case reflect.Uint16:
		return jen.Uint16()
	case reflect.Uint32:
		return jen.Uint32()
	case reflect.Uint64:
		return jen.Uint64()
	case reflect.Float32:
		return jen.Float32()
	case reflect.Float64:
		return jen.Float64()
	case reflect.Interface:
		return jen.Interface()
	default:
		return jen.Interface()
	}
}

func fieldStatement(field Field, isInterface bool) *jen.Statement {
	if field.IsNode {
		if field.Type.Kind() == reflect.Slice {
			if isInterface {
				return jen.Index().Id(field.ElemTypeName + "Pass")
			}

			return jen.Index().Op("*").Id(field.ElemTypeName + "PipePass")
		}

		if isInterface {
			return jen.Id(field.ElemTypeName + "Pass")
		}

		return jen.Op("*").Id(field.ElemTypeName + "PipePass")
	}

	return typeToStatement(field.Type)
}

func generateInterface(f *jen.File, interfaceName string, root *Entity) {
	f.Type().Id(interfaceName).InterfaceFunc(func(grp *jen.Group) {
		for _, field := range root.Fields {
			switch {
			case field.IsNode && field.Type.Kind() == reflect.Slice:
				grp.Id(field.Name).Params().Add(fieldStatement(field, true))
				grp.Id("Map" + field.Name).Params(jen.Func().Params(jen.Id("child").Id(field.ElemTypeName + "Pass")).Error()).Error()
				grp.Id("Append"+field.Name).Params(jen.Id("value").Id(field.ElemTypeName+"Pass"), jen.Id("reason").String())
			case field.IsNode:
				grp.Id(field.Name).Params().Add(fieldStatement(field, true))
				grp.Id("Set"+field.Name).Params(jen.Id("value").Add(fieldStatement(field, true)), jen.Id("reason").String())
			case field.Type.Kind() == reflect.Map:
				grp.Id(field.Name + "Key").Params(jen.Id("key").Add(typeToStatement(field.Type.Key()))).Add(typeToStatement(field.Type.Elem()))
				grp.Id("Set"+field.Name+"Key").Params(jen.Id("key").Add(typeToStatement(field.Type.Key())), jen.Id("value").Add(typeToStatement(field.Type.Elem())), jen.Id("reason").String())
			default:
				grp.Id(field.Name).Params().Add(fieldStatement(field, true))
				grp.Id("Set"+field.Name).Params(jen.Id("value").Add(fieldStatement(field, true)), jen.Id("reason").String())
			}
		}
	})
}

func generateStruct(f *jen.File, structName string, root *Entity) {
	f.Type().Id(structName).StructFunc(func(grp *jen.Group) {
		grp.Id("_path").String()
		grp.Id("_ledger").Qual(LedgerPkgPath(), "Ledger")

		for _, field := range root.Fields {
			grp.Id(privateName(field.Name)).Add(fieldStatement(field, false))
		}
	})
}

func generateConstructor(f *jen.File, structName string) {
	f.Func().Id("New"+structName).
		Params(
			jen.Id("path").String(),
			jen.Id("ledger").Qual(LedgerPkgPath(), "Ledger"),
		).
		Op("*").Id(structName).
		Block(
			jen.Return(jen.Op("&").Id(structName).Values(jen.Dict{
				jen.Id("_path"):   jen.Id("path"),
				jen.Id("_ledger"): jen.Id("ledger"),
			})),
		)
}

func generateAccessors(f *jen.File, structName string, root *Entity) {
	for _, field := range root.Fields {
		switch {
		case field.IsNode && field.Type.Kind() == reflect.Slice:
			generateNodeMethods(f, structName, field)
		case field.IsNode:
			generateSingularNodeMethods(f, structName, field)
		case field.Type.Kind() == reflect.Map:
			generateMapMethods(f, structName, field)
		default:
			generateScalarMethods(f, structName, field)
		}
	}
}

func generateNodeMethods(f *jen.File, structName string, field Field) {
	pName := privateName(field.Name)

	// Getter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id(field.Name).Params().Add(fieldStatement(field, true)).
		Block(
			jen.If(jen.Id("p").Dot(pName).Op("==").Nil()).Block(jen.Return(jen.Nil())),
			jen.Id("res").Op(":=").Make(fieldStatement(field, true), jen.Len(jen.Id("p").Dot(pName))),
			jen.For(jen.Id("i").Op(":=").Lit(0), jen.Id("i").Op("<").Len(jen.Id("p").Dot(pName)), jen.Id("i").Op("++")).Block(
				jen.Id("res").Index(jen.Id("i")).Op("=").Id("p").Dot(pName).Index(jen.Id("i")),
			),
			jen.Return(jen.Id("res")),
		)

	// Map
	f.Func().
		Params(jen.Id("p").Op("*").Id(structName)).
		Id("Map"+field.Name).
		Params(jen.Id("fn").Func().Params(jen.Id("child").Id(field.ElemTypeName+"Pass")).Error()).
		Error().
		Block(
			jen.If(jen.Id("p").Dot(pName).Op("==").Nil()).Block(
				jen.Id("p").Dot(pName).Op("=").Index().Op("*").Id(field.ElemTypeName+"PipePass").Values(),
			),
			jen.For(jen.Id("i").Op(":=").Lit(0), jen.Id("i").Op("<").Len(jen.Id("p").Dot(pName)), jen.Id("i").Op("++")).Block(
				jen.Id("childPath").Op(":=").Id("p").Dot("_path").Op("+").Lit("."+field.Name+"[").Op("+").Qual("strconv", "Itoa").Params(jen.Id("i")).Op("+").Lit("]"),
				jen.Id("childPass").Op(":=").Id("p").Dot(pName).Index(jen.Id("i")),
				jen.Id("childPass").Dot("_path").Op("=").Id("childPath"),
				jen.Id("childPass").Dot("_ledger").Op("=").Id("p").Dot("_ledger"),
				jen.If(jen.Id("err").Op(":=").Id("fn").Params(jen.Id("childPass")), jen.Id("err").Op("!=").Nil()).Block(
					jen.Return(jen.Id("err")),
				),
			),
			jen.Return(jen.Nil()),
		)

	// Append
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id("Append"+field.Name).
		Params(jen.Id("value").Id(field.ElemTypeName+"Pass"), jen.Id("reason").String()).
		Block(
			jen.If(jen.Id("p").Dot(pName).Op("==").Nil()).Block(
				jen.Id("p").Dot(pName).Op("=").Index().Op("*").Id(field.ElemTypeName+"PipePass").Values(),
			),
			jen.List(jen.Id("concreteChild"), jen.Id("ok")).Op(":=").Id("value").Assert(jen.Op("*").Id(field.ElemTypeName+"PipePass")),
			jen.If(jen.Id("ok").Op("&&").Id("concreteChild").Op("!=").Nil()).Block(
				jen.Id("p").Dot(pName).Op("=").Id("append").Params(jen.Id("p").Dot(pName), jen.Id("concreteChild")),
				jen.Id("idx").Op(":=").Len(jen.Id("p").Dot(pName)).Op("-").Lit(1),
				jen.Id("childPath").Op(":=").Id("p").Dot("_path").Op("+").Lit("."+field.Name+"[").Op("+").Qual("strconv", "Itoa").Params(jen.Id("idx")).Op("+").Lit("]"),
				jen.Id("concreteChild").Dot("_path").Op("=").Id("childPath"),
				jen.Id("concreteChild").Dot("_ledger").Op("=").Id("p").Dot("_ledger"),
				jen.If(jen.Id("p").Dot("_ledger").Op("!=").Nil()).Block(
					jen.Id("p").Dot("_ledger").Dot("Log").Params(jen.Id("childPath"), jen.Id("reason"), jen.Nil(), jen.Id("concreteChild")),
				),
			),
		)
}

func generateSingularNodeMethods(f *jen.File, structName string, field Field) {
	pName := privateName(field.Name)

	// Getter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id(field.Name).Params().Add(fieldStatement(field, true)).
		Block(
			jen.If(jen.Id("p").Dot(pName).Op("==").Nil()).Block(jen.Return(jen.Nil())),
			jen.Return(jen.Id("p").Dot(pName)),
		)

	// Setter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id("Set"+field.Name).
		Params(jen.Id("value").Id(field.ElemTypeName+"Pass"), jen.Id("reason").String()).
		Block(
			jen.If(jen.Qual("reflect", "DeepEqual").Params(jen.Id("p").Dot(pName), jen.Id("value"))).Block(jen.Return()),
			jen.Id("prev").Op(":=").Id("p").Dot(pName),
			jen.Id("concrete").Op(",").Id("ok").Op(":=").Id("value").Assert(jen.Op("*").Id(field.ElemTypeName+"PipePass")),
			jen.If(jen.Id("ok").Op("&&").Id("concrete").Op("!=").Nil()).Block(
				jen.Id("childPath").Op(":=").Id("p").Dot("_path").Op("+").Lit("."+field.Name),
				jen.Id("concrete").Dot("_path").Op("=").Id("childPath"),
				jen.Id("concrete").Dot("_ledger").Op("=").Id("p").Dot("_ledger"),
				jen.Id("p").Dot(pName).Op("=").Id("concrete"),
				jen.If(jen.Id("p").Dot("_ledger").Op("!=").Nil()).Block(
					jen.Id("p").Dot("_ledger").Dot("Log").Params(jen.Id("childPath"), jen.Id("reason"), jen.Id("prev"), jen.Id("concrete")),
				),
			),
		)
}

func generateMapMethods(f *jen.File, structName string, field Field) {
	pName := privateName(field.Name)

	// Key Getter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id(field.Name + "Key").
		Params(jen.Id("key").Add(typeToStatement(field.Type.Key()))).Add(typeToStatement(field.Type.Elem())).
		Block(
			jen.Return(jen.Id("p").Dot(pName).Index(jen.Id("key"))),
		)

	// Key Setter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id("Set"+field.Name+"Key").
		Params(jen.Id("key").Add(typeToStatement(field.Type.Key())), jen.Id("value").Add(typeToStatement(field.Type.Elem())), jen.Id("reason").String()).
		Block(
			jen.If(jen.Id("p").Dot(pName).Op("==").Nil()).Block(
				jen.Id("p").Dot(pName).Op("=").Make(fieldStatement(field, false)),
			),
			jen.If(jen.Qual("reflect", "DeepEqual").Params(jen.Id("p").Dot(pName).Index(jen.Id("key")), jen.Id("value"))).Block(jen.Return()),
			jen.Id("prev").Op(":=").Id("p").Dot(pName).Index(jen.Id("key")),
			jen.Id("p").Dot(pName).Index(jen.Id("key")).Op("=").Id("value"),
			jen.If(jen.Id("p").Dot("_ledger").Op("!=").Nil()).Block(
				jen.Id("p").Dot("_ledger").Dot("Log").Params(
					jen.Id("p").Dot("_path").Op("+").Lit("."+field.Name+"[").Op("+").Qual("fmt", "Sprint").Params(jen.Id("key")).Op("+").Lit("]"),
					jen.Id("reason"),
					jen.Id("prev"),
					jen.Id("value"),
				),
			),
		)
}

func generateScalarMethods(f *jen.File, structName string, field Field) {
	pName := privateName(field.Name)

	// Getter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id(field.Name).Params().Add(fieldStatement(field, true)).
		Block(jen.Return(jen.Id("p").Dot(pName)))

	// Setter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id("Set"+field.Name).
		Params(jen.Id("value").Add(fieldStatement(field, true)), jen.Id("reason").String()).
		Block(
			jen.If(jen.Qual("reflect", "DeepEqual").Params(jen.Id("p").Dot(pName), jen.Id("value"))).Block(jen.Return()),
			jen.Id("prev").Op(":=").Id("p").Dot(pName),
			jen.Id("p").Dot(pName).Op("=").Id("value"),
			jen.If(jen.Id("p").Dot("_ledger").Op("!=").Nil()).Block(
				jen.Id("p").Dot("_ledger").Dot("Log").Params(
					jen.Id("p").Dot("_path").Op("+").Lit("."+field.Name),
					jen.Id("reason"),
					jen.Id("prev"),
					jen.Id("value"),
				),
			),
		)
}
