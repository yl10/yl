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
	//GenColumns 固定的匹配字符串
	GenColumns = "@tablename"
)

var (
	colTagName = cmdGenColumns.Flag.String("tag", "xorm", "tagname")
	colPath    = cmdGenColumns.Flag.String("path", ".", "指定路径")
	colMapRole = cmdGenColumns.Flag.Int("m", colGonicMapper, "名称映射模式，0:SnakeMapper ;1:SameMapper ;2:GonicMapper ;3:没有tag的字段忽略.默认：2")
	colPrefix  = cmdGenColumns.Flag.String("prefix", "", "表名的前缀")
	coldebug   = cmdGenColumns.Flag.Bool("debug", false, "show generated data")
)

var cmdGenColumns = &Command{

	UsageLine: "gencolumns [-tag] [-path] [-m ] [-prefix] [-debug]",
	Short:     "根据struct的tag定义，生成_columns.go文件，避免拼SQL的时候出错。",
	Long: `
		tag说明
		-tag	定义字段时用的tag标记，默认是xorm。
		-path	需要解析的目录，默认是当前目录。注意：目录会一直往下递归。
		-m		名称映射模式。0:SnakeMapper ;1:SameMapper ;2:GonicMapper ;3:没有tag的字段忽略.默认：2
		-prefix	生成表名的时候希望增加的前缀，如果注释中制定了表名，前缀不生效。
		-debug	在终端显示生成的内容。
		
		功能说明：

		在使用ORM或者直接拼SQL语句的时候，字段名都是直接写出来的，写错了或被修改了，编译的时候也不会报错，从而导致最终在数据库里执行的SQL语句有问题。
		genclomns会自动创建_columns.go文件，生成和相应struct完全一样的字段。比如：

		User.go文件
		package main
		//User @tablename user
		type User struct{
			ID int
			Name string
		}

		执行 yl gencolumns 

		自动创建 User_columns.go
		package main

		var(
			UserColumns=_UserColumns{
				ID:"id",
				Name:"name",
				TableName:"user",
			}
		)
		type(
			_UserColumns struct{
				ID string
				Name string
			}
		)

		拼SQL的时候，直接使用UserColumns.Name 来替代 "name",UserColumns.TableName 来表示表名。

		===========QA:==============
		具备哪些条件的sturct定义才会生成?
		1.需要加注释，注释中包含 @tablename，不区分大小写。
		2.可导出的类型，即首字母要大写。

		

    `,
}

var temp = `
package {{.PackageName}}

var(
	{{range .Structs}}
 // {{.StructName}}Columns {{lower .StructName}} columns name and table name.
{{.StructName}}Columns=  _{{.StructName}}Column{
	 {{range $key, $value := .Columns}} {{$key}} : "{{$value}}",
        {{end}}
}
	{{end}}

)




{{range .Structs}}
{{$structname:=.StructName}}

    type _{{$structname}}Column struct {
        {{range $key, $value := .Columns}} {{ $key }} string
        {{end}}
    }
   
{{end}}
   
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
	fmt.Println("	开始解析文件....")
	err := filepath.Walk(*colPath, func(filename string, f os.FileInfo, _ error) error {
		if filepath.Ext(filename) == ".go" {
			if strings.Contains(filename, "_column.go") || strings.HasSuffix(filename, "_test.go") {
				return nil
			}

			return handleFile(filename)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("异常:%v", err)
	}
}

func handleFile(filename string) error {
	fmt.Printf("开始解析文件：%s", filename)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("解析文件%s出现异常：%v", filename, err)
		}
	}()
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

	getTableName := func(structname string, cg *ast.CommentGroup) string {
		//*colPrefix
		//解析制定的名字
		specName := ""
		for _, v := range cg.List {
			line := v.Text

			if strings.Contains(strings.ToLower(line), GenColumns) {

				i := strings.Index(line, "@")
				line = line[i:]

				s := strings.Fields(line)

				for i, ok, l := 0, false, len(s); i < l && !ok; i++ {
					if strings.ToLower(s[i]) == GenColumns && i != l-1 {
						specName = s[i+1]
						ok = true
					}
				}

			}

		}
		if specName == "" {
			switch *colMapRole {
			case colSnakeMapper:
				return *colPrefix + SnakeCasedName(structname)
			case colGonicMapper:
				return *colPrefix + GonicCasedName(structname)
			default:
				return *colPrefix + structname
			}
		}
		return specName
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				for _, s := range x.Specs {
					vSpec := s.(*ast.TypeSpec)
					if specStruct, ok := vSpec.Type.(*ast.StructType); ok {

						//不是大写
						if !isASCIIUpper(rune(vSpec.Name.Name[0])) {
							continue
						}

						doc := strings.ToLower(x.Doc.Text())
						if !strings.Contains(doc, GenColumns) {
							continue
						}

						var tempData ColStruct
						tempData.StructName = vSpec.Name.Name

						tempData.Columns = make(map[string]string, 0)
						for _, specField := range specStruct.Fields.List {
							if fieldname := getFieldName(specField); fieldname != "" {
								tempData.Columns[specField.Names[0].Name] = fieldname
							}
						}

						if len(tempData.Columns) > 0 {
							tempData.Columns["TableName"] = getTableName(tempData.StructName, x.Doc)
							colfile.Structs = append(colfile.Structs, tempData)
						}

					}
				}
			}
		}
		return true
	})
	if *coldebug {
		err = colfile.writeTo(os.Stdout)
	}
	err = colfile.WriteToFile()
	if err == nil {
		colfile.writeSuccess()
	}
	return err
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

func (d *ColFile) writeSuccess() error {

	str := `
	成功解析文件：{{.FileName}}
	包括以下struct：
		{{range .Structs}}{{.StructName}}
		{{end}}
	`

	return template.Must(template.New("success").Parse(str)).Execute(os.Stdout, d)
}

// WriteToFile 将生成好的模块文件写到本地
func (d *ColFile) WriteToFile() error {

	fname := strings.Replace(d.FileName, ".go", "_columns.go", -1)
	if len(d.Structs) == 0 {
		err := os.Remove(fname)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	file, err := os.Create(fname)
	if err != nil {
		return err
	}

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

	return err
}
