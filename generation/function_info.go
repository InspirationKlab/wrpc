package main

import (
	"errors"
	"fmt"
	ast2 "go/ast"
	"go/parser"
	"go/token"
	"os"
)

type TerminalCtrlSequence string

var builtInTypes = map[string]bool{
	"int":     true,
	"int32":   true,
	"float32": true,
	"int64":   true,
	"float64": true,
	"string":  true,
	"byte":    true,
}

const (
	SwitchToRed TerminalCtrlSequence = "\u001b[31m"
	ResetClr    TerminalCtrlSequence = "\u001b[0m"
	SetBold     TerminalCtrlSequence = "\u001b[1m"
)

type ParameterInfo struct {
	Name         string
	NeedsPackage bool
	Package      string
	Typename     string
}

func (pInfo *ParameterInfo) Render() string {
	if pInfo.NeedsPackage {
		return fmt.Sprintf("%s.%s", pInfo.Package, pInfo.Typename)
	}
	return pInfo.Typename
}

type ExtendedFunctionInfo struct {
	Name             string
	IsAsync          bool
	IsVoid           bool
	Package          string
	RequestParameter ParameterInfo //nil if no parameter
	Dependencies     []ParameterInfo
}

func ParseParameterInfo(f *ast2.Field) []ParameterInfo {
	result := make([]ParameterInfo, 0, 1)
	for _, name := range f.Names {
		info := ParameterInfo{
			Typename: fmt.Sprintf("%v", f.Type),
			Package:  os.Getenv("GOPACKAGE"),
			Name:     name.String(),
		}
		switch fieldType := f.Type.(type) {
		case *ast2.Ident:
			info.NeedsPackage = !builtInTypes[fieldType.Name]
			break
		case *ast2.SelectorExpr:
			// can be only dot expr
			info.Typename = fmt.Sprintf("%s.%s", fieldType.X, fieldType.Sel.Name)
		}
		result = append(result, info)
	}

	return result
}

func InitializeFromAst(ast *ast2.FuncDecl) (ExtendedFunctionInfo, error) {
	result := ExtendedFunctionInfo{
		Name:    ast.Name.Name,
		Package: os.Getenv("GOPACKAGE"),
	}

	if len(ast.Type.Params.List) != 0 {
		requestParam := ParseParameterInfo(ast.Type.Params.List[0])
		if len(requestParam) != 1 {
			return ExtendedFunctionInfo{}, errors.New("too much parameters for request (expected 1)")
		}
		result.RequestParameter = requestParam[0]
	}

	for _, field := range ast.Type.Params.List[1:] {
		for _, info := range ParseParameterInfo(field) {
			result.Dependencies = append(result.Dependencies, info)
		}
	}

	// async function returns only channels
	isAsync := true

	if ast.Type.Results == nil {
		result.IsAsync = false
		result.IsVoid = true
	}

	for _, resultT := range ast.Type.Results.List {
		switch resultT.Type.(type) {
		case *ast2.ChanType:
			break
		default:
			isAsync = false
		}
	}

	result.IsAsync = isAsync

	return result, nil
}

type TypePackageContext map[string]string

func ParseFile(filename string) ([]ExtendedFunctionInfo, error) {
	fset := token.NewFileSet()

	ast, err := parser.ParseFile(fset, filename, nil, parser.DeclarationErrors|parser.ParseComments|parser.AllErrors)
	result := make([]ExtendedFunctionInfo, 0, 10)

	if err != nil {
		fmt.Printf("%sError : %v%s\n", SwitchToRed, err, ResetClr)
		return nil, err
	}

	for _, decl := range ast.Decls {
		switch t := decl.(type) {
		case *ast2.FuncDecl:

			fmt.Printf("%sFunc found: %s%s\n", SetBold, t.Name.Name, ResetClr)
			einfo, err := InitializeFromAst(t)
			if err != nil {
				fmt.Printf("%sError : %v%s\n", SwitchToRed, err, ResetClr)
				os.Exit(1)
			}
			result = append(result, einfo)

		}
	}
	return result, nil
}
