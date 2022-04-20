package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

const (
	validRequired  = "required"
	validParamName = "paramname"
	validEnum      = "enum"
	validDefault   = "default"
	validMin       = "min"
	validMax       = "max"
)

const (
	responseWriteFuncName = "responseWrite"
	handlerPostfix        = "handler"
)

type apiMethod struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type apiFunc struct {
	apiMethod
	Name     string
	Receiver string
	InArg    string
	OutArg   string
}

type tagParams []string

type field struct {
	fieldType        string
	fieldName        string
	apiValidatorTags map[string]tagParams
}

type structInfo struct {
	name   string
	fields []field
}
type genApiInfo struct {
	PackageName  string
	Funcs        map[string][]apiFunc
	ValidStructs []structInfo
}

var (
	TempResponseWrite = template.Must(template.New("TempResponseWrite").Parse(`
func ` + responseWriteFuncName + `(w http.ResponseWriter, r *http.Request, obj interface{},statusCode int) {
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
	b,_:= json.Marshal(obj)
	w.Write(b)
}`))

	TempServeHTTP = template.Must(template.New("TempServeHTTP").Parse(`
{{ range $key, $val := .Funcs }}
func (api *{{ $key }} ) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	{{range $val }} case "{{ .Url }}":
		api.` + handlerPostfix + `{{ .Name }}(w,r)
	{{end}} default:
		` + responseWriteFuncName + `(w,r,"some error",http.StatusBadRequest)
	}
}
{{end}}`))
	TempHandleFunc = template.Must(template.New("TempServeHTTP").Parse(`
{{ range $key, $val := .Funcs }}
{{range $val }} 
func (api *{{ $key }} ) ` + handlerPostfix + `{{ .Name }}(w http.ResponseWriter, r *http.Request) {
	//write Body
}
{{end}}
{{end}}`))
)

func parseConcreteStruct(currType *ast.TypeSpec, structType *ast.StructType, apiInfo *genApiInfo) {
	log.Printf("Parsing struct on position %d : %s  ", currType.Pos(), currType.Name.Name)
	currentStructInfo := structInfo{}
	currentStructInfo.name = currType.Name.Name
	flagNeedToAdd := false
	for _, f := range structType.Fields.List {
		if idend, ok := f.Type.(*ast.Ident); !ok {
			log.Println("Parsing struct breaking: '" + currType.Name.Name + "' -- parser  supports only string types")
			break
		} else {
			for _, name := range f.Names {
				fieldInfo := field{
					fieldType:        idend.Name,
					fieldName:        name.Name,
					apiValidatorTags: map[string]tagParams{},
				}
				if f.Tag != nil {
					tt := (reflect.StructTag)(strings.ReplaceAll(f.Tag.Value, "`", ""))
					validTagString := tt.Get("apivalidator")
					if len(validTagString) == 0 {
						continue
					}
					flagNeedToAdd = true
					validTags := strings.Split(validTagString, ",")
					for _, tag := range validTags {
						splitTag := strings.Split(tag, "=")
						if len(splitTag) > 1 {
							argsTag := strings.Split(splitTag[1], "|")
							fieldInfo.apiValidatorTags[splitTag[0]] = argsTag
						} else {
							fieldInfo.apiValidatorTags[splitTag[0]] = nil
						}
					}
				}
				currentStructInfo.fields = append(currentStructInfo.fields, fieldInfo)
			}
		}
	}
	if flagNeedToAdd {
		log.Printf("ok\n")
		apiInfo.ValidStructs = append(apiInfo.ValidStructs, currentStructInfo)
	} else {
		log.Printf("validation params not found\n")
	}
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
		log.Printf("Parsing func on position %d : %s ", elem.Pos(), elem.Text)
		json.Unmarshal([]byte(strings.TrimPrefix(elem.Text, "// apigen:api ")), &funcInfo.apiMethod)
		funcInfo.Name = decl.Name.Name
		var recName string
		if decl.Recv != nil {
			if dl, ok := decl.Recv.List[0].Type.(*ast.StarExpr); ok {
				recName = dl.X.(*ast.Ident).Obj.Name
			} else if dl, ok := decl.Recv.List[0].Type.(*ast.Ident); ok {
				recName = dl.Obj.Name
			} else {
				log.Println("Parsing error: '" + decl.Name.Name + "' function Receiver type error")
			}
		} else {
			log.Println("Parsing error: '" + decl.Name.Name + "' function Receiver not exist")
		}
		funcInfo.Receiver = recName

		var argName string
		if decl.Type.Params != nil {
			if len(decl.Type.Params.List) == 2 {
				if selector, ok := decl.Type.Params.List[0].Type.(*ast.SelectorExpr); !ok || "Context" != selector.Sel.Name {
					log.Println("Parsing error: '" + decl.Name.Name + "' type of first arg must be 'context.Context' ")
				} else {
					if indent, ok := decl.Type.Params.List[1].Type.(*ast.Ident); ok && len(decl.Type.Params.List[1].Names) == 1 {
						argName = indent.Name
					} else {
						log.Println("Parsing error: '" + decl.Name.Name + "' type of second arg must struct ")
					}
				}
			} else {
				log.Println("Parsing error: '" + decl.Name.Name + "' function must have 2 arguments ")
			}
		} else {
			log.Println("Parsing error: '" + decl.Name.Name + "' function Receiver not exist")
		}
		funcInfo.InArg = argName

		var outName string
		if decl.Type.Results != nil {
			if len(decl.Type.Results.List) == 2 {
				if selector, ok := decl.Type.Results.List[1].Type.(*ast.Ident); !ok || "error" != selector.Name {
					log.Println("Parsing error: '" + decl.Name.Name + "' type of second out must be 'error' ")
				} else {
					if starExpr, ok := decl.Type.Results.List[0].Type.(*ast.StarExpr); ok && len(decl.Type.Params.List[1].Names) == 1 {
						outName = starExpr.X.(*ast.Ident).Obj.Name
					} else {
						log.Println("Parsing error: '" + decl.Name.Name + "' type of first arg must be starExpr of struct ")
					}
				}
			} else {
				log.Println("Parsing error: '" + decl.Name.Name + "' function must have 2 arguments ")
			}
		} else {
			log.Println("Parsing error: '" + decl.Name.Name + "' function Receiver not exist")
		}
		funcInfo.OutArg = outName
	}
	if apiInfo.Funcs[funcInfo.Receiver] == nil {
		apiInfo.Funcs[funcInfo.Receiver] = make([]apiFunc, 0)
	}
	apiInfo.Funcs[funcInfo.Receiver] = append(apiInfo.Funcs[funcInfo.Receiver], funcInfo)
}

func genWriteHead(pkgName string, writer io.Writer) {
	fmt.Fprintf(writer, "package %s\n\n", pkgName)
	fmt.Fprintf(writer, `import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"encoding/json"
     )

`)
}
func generate(apiInfo *genApiInfo, writer io.Writer) {
	genWriteHead(apiInfo.PackageName, writer)
	TempServeHTTP.Execute(writer, apiInfo)
	TempHandleFunc.Execute(writer, apiInfo)
	TempResponseWrite.Execute(writer, nil)
}
func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s [input_file] [output_file]", os.Args[0])
		return
	}

	outFile, err := os.Create(os.Args[2])
	if err != nil {
		log.Printf("Open output file error: %s", err)
		return
	}
	defer outFile.Close()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	genApi := new(genApiInfo)
	genApi.PackageName = file.Name.Name
	genApi.Funcs = make(map[string][]apiFunc, 0)
	// Parsing
	for _, f := range file.Decls {
		switch g := (f).(type) {
		case *ast.GenDecl:
			parseStruct(g, genApi)
		case *ast.FuncDecl:
			parseFunc(g, genApi)
		default:
			continue
		}
	}

	//Generating

	fmt.Println(genApi.Funcs)
	fmt.Println(genApi.ValidStructs)
	generate(genApi, outFile)
}
