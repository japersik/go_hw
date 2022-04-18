package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type apiMethod struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}
type apiFunc struct {
	apiMethod
	receiver string
}
type genApiInfo struct {
}

func parseConcreteStruct(currType *ast.TypeSpec, structType *ast.StructType, apiInfo *genApiInfo) {
	fmt.Println(currType.Name, structType.Fields)
}

func parseStruct(decl *ast.GenDecl, apiInfo *genApiInfo) {
	for _, spec := range decl.Specs {
		currType, ok := spec.(*ast.TypeSpec)
		if ok {
			currStruct, ok := currType.Type.(*ast.StructType)
			if ok {
				parseConcreteStruct(currType, currStruct, apiInfo)
			}
		}
	}
}
func parseFunc(decl *ast.FuncDecl, apiInfo *genApiInfo) {
	var funcInfo = apiFunc{}
	if decl.Doc == nil {
		return
	}
	for _, elem := range decl.Doc.List {
		if !strings.HasPrefix(elem.Text, "// apigen:api ") {
			continue
		}

		json.Unmarshal([]byte(strings.TrimPrefix(elem.Text, "// apigen:api ")), &funcInfo.apiMethod)

		var recName string
		if decl.Recv != nil {
			if dl, ok := decl.Recv.List[0].Type.(*ast.StarExpr); ok {
				recName = dl.X.(*ast.Ident).Obj.Name
			} else if dl, ok := decl.Recv.List[0].Type.(*ast.Ident); ok {
				recName = dl.Obj.Name
			} else {
				log.Println("Parsing error: '" + decl.Name.Name + "' function receiver type error")
			}
		} else {
			log.Println("Parsing error: '" + decl.Name.Name + "' function receiver not exist")
		}
		funcInfo.receiver = recName
	}
	fmt.Println(funcInfo)
}
func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	genApi := new(genApiInfo)
	// parsing
	for _, f := range file.Decls {
		switch g := (f).(type) {
		case *ast.GenDecl:
			//parseStruct(g, genApi)
		case *ast.FuncDecl:
			parseFunc(g, genApi)
		default:
			continue
		}

		//	for _, spec := range g.Specs {
		//		fmt.Println("_______________")
		//		cType, ok := spec.(*ast.TypeSpec)
		//		if ok {
		//			//fmt.Println(cType.Name)
		//			//fmt.Println(cType)
		//			switch stType := (cType.Type).(type) {
		//			case *ast.StructType:
		//				fmt.Println("Fields")
		//				for _, field := range stType.Fields.List {
		//					fmt.Printf("%s: ", field.Names[0].Name)
		//					if field.Tag == nil {
		//						continue
		//					}
		//					tag := reflect.StructTag(field.Tag.Value)
		//					fmt.Println(tag)
		//				}
		//			case *ast.FuncType:
		//				fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!")
		//				fmt.Println(cType.Name)
		//				fmt.Println(cType)
		//			}
		//
		//		}
		//	}
		//case *ast.FuncDecl:
		//	fmt.Println("!!!!!!!!!!!!!")
		//	fmt.Println(g.Name.Name)
		//	fmt.Println(g.Doc.Text())
		//	fmt.Println("!!!!!!!!!!!!!")

	}
	// Generating
}
