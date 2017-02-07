package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	colSnakeMapper = iota
	colSameMapper
	colGonicMapper
	colMustTagMapper
)
const (
	//GenColumns GenColumns
	GenColumns = "GenColumns"
)

var (
	colTagName = cmdGenColumns.Flag.String("tag", "xorm", "tagname")
	colPath    = cmdGenColumns.Flag.String("path", ".", "指定路径")
	colMapRole = cmdGenColumns.Flag.Int("m", 0, "名称映射模式，0:SnakeMapper ;1:SameMapper ;2:GonicMapper ;3:没有tag的字段忽略")

	debug = cmdGenColumns.Flag.Bool("debug", false, "show generated data")
)

var cmdGenColumns = &Command{

	UsageLine: "gencol [-tag]",
	Short:     "Generate ",
	Long: `

    `,
}

var temp = `
package {{.PackageName}}
{{range .Structs}}
{{$structname:=.StructName}}
    type _{{$structname}}Column struct {
        {{range $key, $value := .Columns}} {{ $key }} string
        {{end}}
    }
    // {{$structname}}Columns {{lower $structname}} columns name
    var {{$structname}}Columns  _{{$structname}}Column


{{end}}
    func init() {
        {{range .Structs}}
{{$structname:=.StructName}}
   {{range $key, $value := .Columns}} {{ $structname}}Columns.{{$key}} = "{{$value}}"
        {{end}}
{{end}}
}
`

//ColFile 文件
type ColFile struct {
	FileName    string
	PackageName string
	Structs     []ColStruct
}

// ColStruct 表示生成template所需要的数据结构
type ColStruct struct {
	StructName string
	Columns    map[string]string
}

func init() {
	cmdGenColumns.Run = runGenColumns
}

func runGenColumns(cmd *Command, args []string) {
	fmt.Println("开始解析文件....")
	filepath.Walk(*colPath, func(filename string, f os.FileInfo, _ error) error {
		if filepath.Ext(filename) == ".go" {
			if strings.Contains(filename, "_column.go") || strings.HasSuffix(filename, "_test.go") {
				return nil
			}

			return handleFile(filename)
		}
		return nil
	})
}

func handleFile(filename string) error {
	fmt.Printf("解析%v", filename)
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	colfile := ColFile{
		FileName:    filename,
		PackageName: f.Name.Name,
		Structs:     make([]ColStruct, 0),
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				for _, s := range x.Specs {
					vSpec := s.(*ast.TypeSpec)
					if specStruct, ok := vSpec.Type.(*ast.StructType); ok {
						fmt.Printf("%v\r\n", vSpec.Name.Name)
						//不是大写
						if !isASCIIUpper(rune(vSpec.Name.Name[0])) {
							continue
						}

						doc := x.Doc.Text()
						if !strings.Contains(doc, GenColumns) {
							continue
						}
						fmt.Println("找到需要处理的sturct:", vSpec.Name.Name)
						var tempData ColStruct
						tempData.StructName = vSpec.Name.Name
						tempData.Columns = make(map[string]string, 0)
						for _, specField := range specStruct.Fields.List {
							if fieldname := getFieldName(specField); fieldname != "" {
								tempData.Columns[specField.Names[0].Name] = fieldname
							}
						}
						fmt.Println("tempdata", tempData)
						//if len(tempData.Columns) > 0 {
						colfile.Structs = append(colfile.Structs, tempData)
						//}

					}
				}
			}
		}
		return true
	})
	fmt.Println("colfile:", colfile)
	if *debug {
		colfile.writeTo(os.Stdout)
	}
	colfile.WriteToFile()
	return nil
}

//getDefaultFiledName 根据字段名默认生成
func getDefaultFiledName(fname string) string {
	switch *colMapRole {
	case colSnakeMapper:
		return SnakeCasedName(fname)
	case colGonicMapper:
		return GonicCasedName(fname)
	case colMustTagMapper:
		return ""
	default:
		return fname
	}

}
func getFieldName(f *ast.Field) string {
	//字段小写
	if !isASCIIUpper(rune(f.Names[0].Name[0])) {
		return ""
	}
	//没有标记，解析模式却是必须按照tag
	if f.Tag == nil && *colMapRole == colMustTagMapper {
		return ""
	}

	//先按tag来获取，获取到返回，获取不到继续
	name := getFieldNameByTag(f)
	if name != "" {
		return name
	}
	//根据字段名和转换模式返回
	return getDefaultFiledName(f.Names[0].Name)

}
func getFieldNameByTag(f *ast.Field) string {

	if f.Tag == nil {
		return ""
	}
	tag := reflect.StructTag(f.Tag.Value)
	switch *colTagName {
	case "xorm":
		value, ok := tag.Lookup("xorm")
		if !ok {
			return ""
		}
		tmps := strings.Fields(value)
		for _, v := range tmps {
			if strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'") {
				return strings.Trim(v, "'")
			}
		}
		return ""
	case "orm":
		return ""
	default:

		return tag.Get(*colTagName)
	}

}

func (d *ColFile) writeTo(w io.Writer) error {
	funcMap := template.FuncMap{
		"lower": strings.ToLower,
	}
	return template.Must(template.New("temp").Funcs(funcMap).Parse(temp)).Execute(w, d)

}

// WriteToFile 将生成好的模块文件写到本地
func (d *ColFile) WriteToFile() error {

	fname := strings.Replace(d.FileName, ".go", "_colunms.go", -1)
	if len(d.Structs) == 0 {
		return os.Remove(fname)

	}

	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	fmt.Println("d:", d)
	defer file.Close()
	var buf bytes.Buffer
	err = d.writeTo(&buf)

	if err != nil {
		fmt.Println(err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(err)
	}
	_, err = file.Write(formatted)

	fmt.Println("err:", err)
	return err
}
