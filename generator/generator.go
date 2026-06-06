package generator

import (
	"errors"
	"strings"

	"github.com/dave/jennifer/jen"
)

var ErrRender = errors.New("render error")

type Field struct {
	Name         string
	Kind         string
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

func fieldStatement(field Field, isInterface bool) *jen.Statement {
	if field.IsNode {
		if isInterface {
			return jen.Index().Id(field.ElemTypeName + "Pass")
		}

		return jen.Index().Op("*").Id(field.ElemTypeName + "PipePass")
	}

	switch field.Kind {
	case "string":
		return jen.String()
	case "bool":
		return jen.Bool()
	case "int":
		return jen.Int()
	case "int64":
		return jen.Int64()
	case "map": //nolint:goconst
		return jen.Map(jen.String()).Interface()
	default:
		return jen.Interface()
	}
}

func generateInterface(f *jen.File, interfaceName string, root *Entity) {
	f.Type().Id(interfaceName).InterfaceFunc(func(grp *jen.Group) {
		for _, field := range root.Fields {
			switch {
			case field.IsNode:
				grp.Id(field.Name).Params().Add(fieldStatement(field, true))
				grp.Id("Map" + field.Name).Params(jen.Func().Params(jen.Id("child").Id(field.ElemTypeName + "Pass")).Error()).Error()
				grp.Id("Append"+field.Name).Params(jen.Id("value").Id(field.ElemTypeName+"Pass"), jen.Id("reason").String())
			case field.Kind == "map":
				grp.Id(field.Name + "Key").Params(jen.Id("key").String()).Interface()
				grp.Id("Set"+field.Name+"Key").Params(jen.Id("key").String(), jen.Id("value").Interface(), jen.Id("reason").String())
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
		case field.IsNode:
			generateNodeMethods(f, structName, field)
		case field.Kind == "map":
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
			jen.Id("concreteChild").Op(":=").Id("value").Assert(jen.Op("*").Id(field.ElemTypeName+"PipePass")),
			jen.Id("p").Dot(pName).Op("=").Id("append").Params(jen.Id("p").Dot(pName), jen.Id("concreteChild")),
			jen.Id("idx").Op(":=").Len(jen.Id("p").Dot(pName)).Op("-").Lit(1),
			jen.Id("childPath").Op(":=").Id("p").Dot("_path").Op("+").Lit("."+field.Name+"[").Op("+").Qual("strconv", "Itoa").Params(jen.Id("idx")).Op("+").Lit("]"),
			jen.Id("concreteChild").Dot("_path").Op("=").Id("childPath"),
			jen.Id("concreteChild").Dot("_ledger").Op("=").Id("p").Dot("_ledger"),
			jen.If(jen.Id("p").Dot("_ledger").Op("!=").Nil()).Block(
				jen.Id("p").Dot("_ledger").Dot("Log").Params(jen.Id("childPath"), jen.Id("reason"), jen.Nil(), jen.Id("concreteChild")),
			),
		)
}

func generateMapMethods(f *jen.File, structName string, field Field) {
	pName := privateName(field.Name)

	// Key Getter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id(field.Name+"Key").
		Params(jen.Id("key").String()).Interface().
		Block(
			jen.If(jen.Id("p").Dot(pName).Op("==").Nil()).Block(jen.Return(jen.Nil())),
			jen.Return(jen.Id("p").Dot(pName).Index(jen.Id("key"))),
		)

	// Key Setter
	f.Func().Params(jen.Id("p").Op("*").Id(structName)).Id("Set"+field.Name+"Key").
		Params(jen.Id("key").String(), jen.Id("value").Interface(), jen.Id("reason").String()).
		Block(
			jen.If(jen.Id("p").Dot(pName).Op("==").Nil()).Block(
				jen.Id("p").Dot(pName).Op("=").Make(fieldStatement(field, false)),
			),
			jen.If(jen.Id("p").Dot(pName).Index(jen.Id("key")).Op("==").Id("value")).Block(jen.Return()),
			jen.Id("prev").Op(":=").Id("p").Dot(pName).Index(jen.Id("key")),
			jen.Id("p").Dot(pName).Index(jen.Id("key")).Op("=").Id("value"),
			jen.If(jen.Id("p").Dot("_ledger").Op("!=").Nil()).Block(
				jen.Id("p").Dot("_ledger").Dot("Log").Params(
					jen.Id("p").Dot("_path").Op("+").Lit("."+field.Name+"[\"").Op("+").Id("key").Op("+").Lit("\"]"),
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
			jen.If(jen.Id("p").Dot(pName).Op("==").Id("value")).Block(jen.Return()),
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
