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

//const (
//	validRequired  = "required" //+
//	validParamName = "paramname"//+
//	validEnum      = "enum"		//+
//	validDefault   = "default" 	//+
//	validMin       = "min"		//+
//	validMax       = "max"		//+
//)
//
//const (
//	responseWriteFuncName = "responseWrite"
//	handlerPostfix        = "handler"
//)

type ApiMethod struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type ApiFunc struct {
	ApiMethod
	Name     string
	Receiver string
	InArg    string
	OutArg   string
}

type TagParams []string

type Field struct {
	FieldType        string
	FieldName        string
	ApiValidatorTags map[string]TagParams
}

type StructInfo struct {
	Fields []Field
}
type GenApiInfo struct {
	PackageName  string
	Funcs        map[string][]ApiFunc
	ValidStructs map[string]StructInfo
}

var (
	TempResponseWrite = template.Must(template.New("TempResponseWrite").Parse(`
func responseWrite(w http.ResponseWriter, r *http.Request, obj interface{},statusCode int) {
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
		api.handler{{ .Name }}(w,r)
	{{end}} default:
		responseWrite(w,r,ErrorAns{"unknown method"}, http.StatusNotFound)
	}
}
{{end}}`))
	TempHandleFunc = template.Must(template.New("TempServeHTTP").Funcs(template.FuncMap{
		"toLower": func(in string) string {
			return strings.ToLower(in)
		},
		"joinComma": func(req []string) string { return strings.Join(req, ", ") },
	}).Parse(`{{- /*gotype: github.com/japersik/go_hw/hw5_codegen/handlers_gen.GenApiInfo*/ -}}
{{ range $key, $val := .Funcs }}{{range $v := $val }} 
func (api *{{ $key }}) handler{{ $v.Name }}(w http.ResponseWriter, r *http.Request) {
    {{if $v.ApiMethod.Auth}}//checkAuth
    if r.Header.Get("X-Auth") != "100500" {
        responseWrite(w, r, ErrorAns{"unauthorized"}, http.StatusForbidden)
        return
    }
    {{end}}r.ParseForm()
	{{$v.InArg | toLower}} :={{$v.InArg}}{}
{{with (index $.ValidStructs ($v.InArg))}}
    {{- /*gotype: github.com/japersik/go_hw/hw5_codegen/handlers_gen.StructInfo*/ -}}
    {{$name := ""}}{{ range $field := .Fields}}{{if .ApiValidatorTags.paramname}} {{ $name = index .ApiValidatorTags.paramname 0}}{{else}}{{$name = .FieldName | toLower}}{{end}}
    //{{.FieldName}} --> {{$name}}
    {{if (eq .FieldType  "int")}}{{.FieldName | toLower}},err := strconv.Atoi(r.Form.Get("{{$name}}"))
    if err!=nil{
        responseWrite(w,r,ErrorAns{"{{$name}} must be int"}, http.StatusBadRequest)
        return
    }{{else}}{{.FieldName | toLower}} := r.Form.Get("{{$name}}"){{end}}
	{{$v.InArg | toLower}}.{{.FieldName}} = {{.FieldName | toLower}}
{{range $keyTag, $values:= .ApiValidatorTags }}{{if eq $keyTag "default"}}{{if (eq $field.FieldType "int")}}if {{$field.FieldName | toLower}} == 0{ 
    {{$field.FieldName | toLower}} = {{index $values 0 }}{{else}}if {{$field.FieldName | toLower}} == "" { 
        {{$field.FieldName | toLower}} = "{{index $values 0 }}"{{end}} 
    }{{end}}
    {{if eq $keyTag "required"}}{{if (eq $field.FieldType "int")}}if {{$field.FieldName | toLower}} == 0 { {{else}}if {{$field.FieldName | toLower}} == "" { {{end}}
        responseWrite(w,r,ErrorAns{"{{$name}} must be not empty"}, http.StatusBadRequest)
        return
    }{{end}}{{end}}
    {{range $keyTag, $values:= .ApiValidatorTags }}
    {{if eq $keyTag "min"}}if{{ $addtext := ""}}{{if (eq $field.FieldType "int")}} {{$field.FieldName | toLower}} {{else}} len({{$field.FieldName | toLower}}){{ $addtext = "len "}} {{end}}< {{index $values 0 }}{
        responseWrite(w,r,ErrorAns{"{{$name}} {{$addtext}}must be >= {{index $values 0 }}"}, http.StatusBadRequest)
        return
    }{{end}}{{if eq $keyTag "max"}}if{{ $addtext := ""}}{{if (eq $field.FieldType "int")}} {{$field.FieldName | toLower}} {{else}} len({{$field.FieldName | toLower}}){{ $addtext = "len "}} {{end}}> {{index $values 0 }}{
        responseWrite(w,r,ErrorAns{"{{$name}} {{$addtext}}must be <= {{index $values 0 }}"}, http.StatusBadRequest)
        return
    }{{end}}{{if eq $keyTag "enum"}}if !({{ $addtext := ""}}{{if (eq $field.FieldType "int")}} ({{range $enumval:= $values}}{{$field.FieldName | toLower}}  == {{$enumval}}{{end}}) || {{else}}{{range $enumval:= $values}}{{$field.FieldName | toLower}}  == "{{$enumval}}" || {{end}} {{end}}false){
        responseWrite(w,r,ErrorAns{"{{$name}} must be one of [{{$values | joinComma}}]"}, http.StatusBadRequest)
        return
    }{{end}}{{end}}{{end}}
{{end}}
    responseWrite(w,r,SomeAns{Ans:{{$v.InArg | toLower}}}, http.StatusOK)
}
{{end}}
{{end}}`))
	TempResponseStruct = template.Must(template.New("TempResponseStruct").Parse("type ErrorAns struct {\n\tErr string `json:\"error\"`\n}\n\ntype SomeAns struct {\n\tErrorAns\n\tAns interface{} `json:\"response\"`\n}"))
)

func parseConcreteStruct(currType *ast.TypeSpec, structType *ast.StructType, apiInfo *GenApiInfo) {
	log.Printf("Parsing struct on position %d : %s  ", currType.Pos(), currType.Name.Name)
	currentStructInfo := StructInfo{}
	flagNeedToAdd := false
	for _, f := range structType.Fields.List {
		if idend, ok := f.Type.(*ast.Ident); !ok {
			log.Println("Parsing struct breaking: '" + currType.Name.Name + "' -- parser  supports only string types")
			break
		} else {
			for _, name := range f.Names {
				fieldInfo := Field{
					FieldType:        idend.Name,
					FieldName:        name.Name,
					ApiValidatorTags: map[string]TagParams{},
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
							fieldInfo.ApiValidatorTags[splitTag[0]] = argsTag
						} else {
							fieldInfo.ApiValidatorTags[splitTag[0]] = nil
						}
					}
				}
				currentStructInfo.Fields = append(currentStructInfo.Fields, fieldInfo)
			}
		}
	}
	if flagNeedToAdd {
		log.Printf("ok\n")
		apiInfo.ValidStructs[currType.Name.Name] = currentStructInfo
	} else {
		log.Printf("validation params not found\n")
	}

}

func parseStruct(decl *ast.GenDecl, apiInfo *GenApiInfo) {
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
func parseFunc(decl *ast.FuncDecl, apiInfo *GenApiInfo) {
	var funcInfo = ApiFunc{}
	if decl.Doc == nil {
		return
	}
	for _, elem := range decl.Doc.List {
		if !strings.HasPrefix(elem.Text, "// apigen:api ") {
			continue
		}
		log.Printf("Parsing func on position %d : %s ", elem.Pos(), elem.Text)
		json.Unmarshal([]byte(strings.TrimPrefix(elem.Text, "// apigen:api ")), &funcInfo.ApiMethod)
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
		apiInfo.Funcs[funcInfo.Receiver] = make([]ApiFunc, 0)
	}
	apiInfo.Funcs[funcInfo.Receiver] = append(apiInfo.Funcs[funcInfo.Receiver], funcInfo)
}

func genWriteHead(pkgName string, writer io.Writer) {
	fmt.Fprintf(writer, "package %s\n\n", pkgName)
	fmt.Fprintf(writer, `import (
	"net/http"
	"strconv"
	"encoding/json"
     )

`)
}
func generate(apiInfo *GenApiInfo, writer io.Writer) {
	genWriteHead(apiInfo.PackageName, writer)
	TempResponseStruct.Execute(writer, nil)
	TempServeHTTP.Execute(writer, apiInfo)
	err := TempHandleFunc.Execute(writer, apiInfo)
	fmt.Println(err)
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
	genApi := new(GenApiInfo)
	genApi.PackageName = file.Name.Name
	genApi.Funcs = make(map[string][]ApiFunc, 0)
	genApi.ValidStructs = make(map[string]StructInfo, 0)
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

	//fmt.Println(genApi.Funcs)
	fmt.Println(genApi.ValidStructs["ProfileParams"])
	generate(genApi, outFile)
}
