package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/yl10/kit/guid"
)

const (
	//GenTableNamecn 固定的匹配字符串
	GenTableNamecn = "@tablecnname"
)

var ()

var cmdGenDict = &Command{

	UsageLine: "gendict [-tag] [-path] [-m ] [-prefix] [-debug] [-file]",
	Short:     "根据struct的tag定义，生成数据字典",
	Long: `
		tag说明
		-tag		定义字段时用的tag标记，默认是xorm。
		-path		需要解析的目录，默认是当前目录。注意：目录会一直往下递归。
		-m			名称映射模式。0:SnakeMapper ;1:SameMapper ;2:GonicMapper ;3:没有tag的字段忽略.默认：2
		-prefix		生成表名的时候希望增加的前缀，如果注释中制定了表名，前缀不生效。
		-mysql		是否连接mysql。
		-dsn		数据连接字符串

		-savefile	是否保存到文件
		-file		保存文件名称
		
		功能说明：

		根据注释和tag生成数据字典
		表名生成规则：
		表的注释中包含 @tablename
		格式：
		@tablename 表名[可以不指定] 
		@tablecnname 表中文名  表说明

		字段生成规则：
		按照tag格式   dict:"中文名  说明"

    `,
}

//FileDict FileDict
type FileDict struct {
	Name   string `json:"filename"`
	Tables []TableDict
}

//TableDict 表
type TableDict struct {
	GUID    guid.GUID     `xorm:"varchar(36) pk"`
	Name    string        `xorm:"varchar(100)"`
	CName   string        `xorm:"'cname varchar(100)"`
	Remark  string        `xorm:"varchar(255)"`
	Columns []ColumnsDict `xorm:"-"`
}

//ColumnsDict 字段表
type ColumnsDict struct {
	GUID      guid.GUID `xorm:"varchar(36) pk"`
	TableGUID guid.GUID `xorm:"varchar(36) index"`
	Name      string    `xorm:"varchar(100)"`
	CName     string    `xorm:"'cname varchar(100)"`
	Remark    string    `xorm:"varchar(255)"`
}

func init() {

	cmdGenDict.Run = runGenDict
	cmdGenDict.Flag.StringVar(flagTagName, "tag", "xorm", "")
	cmdGenDict.Flag.StringVar(flagPath, "path", ".", "")
	cmdGenDict.Flag.BoolVar(flagExpectFile, "savefile", false, "")
	cmdGenDict.Flag.StringVar(flagFilePath, "file", "", "")
	cmdGenDict.Flag.IntVar(flagMapRole, "m", colGonicMapper, "")
	cmdGenDict.Flag.StringVar(flagPrefix, "prefix", "", "")
	cmdGenDict.Flag.BoolVar(flagdebug, "debug", false, "")
	cmdGenDict.Flag.StringVar(flagDsn, "dsn", "", "")
	cmdGenDict.Flag.BoolVar(flagExpectMysql, "mysql", false, "")

}

func runGenDict(cmd *Command, args []string) {

	getTableName := func(structname string, cg *ast.CommentGroup) *TableDict {
		fmt.Printf("===:%v,%v\r\n", structname, cg)
		if cg == nil {
			return nil
		}
		td := &TableDict{}
		for _, v := range cg.List {
			line := v.Text

			lowerLine := strings.ToLower(line)
			if hasname, hascnname := strings.Contains(lowerLine, GenColumns), strings.Contains(lowerLine, GenTableNamecn); hasname || hascnname {
				s := strings.Fields(line)

				for i, nameok, cnnameok, slen := 0, false, false, len(s); i < slen && !(nameok && cnnameok); i++ {
					if !hasname {
						nameok = true
					}
					if !hascnname {
						cnnameok = true
					}
					if hasname && !nameok {
						if strings.ToLower(s[i]) == GenColumns && i < slen-1 {
							if s[i+1] != GenTableNamecn {
								td.Name = s[i+1]
							}
							nameok = true
						}

					}
					if hascnname && !cnnameok {

						if strings.ToLower(s[i]) == GenTableNamecn && i < slen-1 {
							if s[i+1] != GenColumns {
								td.CName = s[i+1]
								if i < slen-2 {
									td.Remark = strings.Join(s[i+2:], " ")
								}

							}
							cnnameok = true
						}
					}

				}

			}

		}
		if td.Name == "" {
			switch *flagMapRole {
			case colSnakeMapper:
				td.Name = *flagPrefix + SnakeCasedName(structname)
			case colGonicMapper:
				td.Name = *flagPrefix + GonicCasedName(structname)
			default:
				td.Name = *flagPrefix + structname
			}
		}
		if td.Name == "" {
			return nil
		}
		td.GUID = guid.NewMD5GUID(structname + td.Name)
		return td
	}

	handleDictFile := func(filename string) *FileDict {
		fset := token.NewFileSet()

		f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("解析文件失败:%s     %v", filename, err)
		}
		dictFile := FileDict{Name: filename}
		dictFile.Tables = make([]TableDict, 0)

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

							td := getTableName(vSpec.Name.Name, x.Doc)

							if td == nil {
								continue
							}
							fmt.Printf("td:%v", td)
							td.Columns = make([]ColumnsDict, 0)
							for _, specField := range specStruct.Fields.List {

								if fieldname := getFieldName(specField); fieldname != "" {
									col := ColumnsDict{}
									col.GUID = guid.NewMD5GUID(td.Name + fieldname)
									col.TableGUID = td.GUID
									col.Name = fieldname
									col.CName, col.Remark = getDictByTag(specField)
									td.Columns = append(td.Columns, col)
								}
							}

							if len(td.Columns) > 0 {
								dictFile.Tables = append(dictFile.Tables, *td)

							}

						}
					}
				}
			}
			return true
		})
		if len(dictFile.Tables) > 0 {
			return &dictFile
		}
		return nil
	}

	fmt.Println("	开始解析文件....")
	var dictFiles []FileDict
	dictFiles = make([]FileDict, 0)
	filepath.Walk(*flagPath, func(filename string, f os.FileInfo, _ error) error {
		if filepath.Ext(filename) == ".go" {
			if strings.Contains(filename, "_column.go") || strings.HasSuffix(filename, "_test.go") {
				return nil
			}

			df := handleDictFile(filename)
			if df != nil {
				dictFiles = append(dictFiles, *df)
			}

		}
		return nil
	})
	//保存到文件
	if *flagExpectFile {

		jsonfilename := *flagFilePath
		if jsonfilename == "" {
			jsonfilename = "sql.json"
		}
		jsonfile, err := os.Create(jsonfilename)
		if err != nil {
			log.Fatalf("打开结果文件出错：%v", err)
		}
		data, err := json.Marshal(dictFiles)
		if err != nil {
			log.Fatalf("出错：%v", err)
		}
		_, err = jsonfile.Write(data)
		if err != nil {
			log.Fatalf("出错：%v", err)
		}
		fmt.Printf("解析结果成功保存到:%s", jsonfilename)
	}

	//保存到数据库

	fmt.Println("shifou baoc数据接口：", *flagExpectMysql)
	if *flagExpectMysql {
		dsn := *flagDsn
		if dsn == "" {
			dsn = "root:123456@/test?charset=utf8"
		}

		fmt.Println("shujuk:", dsn)
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("保存到数据库失败：", err)
			}

		}()
		fmt.Println("开始存入数据库...")
		db, err := xorm.NewEngine("mysql", dsn)
		if err != nil {
			panic(err)
		}
		db.SetMapper(core.GonicMapper{"GUID": true})
		Must(db.Sync2(new(TableDict), new(ColumnsDict)))
		for _, v := range dictFiles {
			for _, t := range v.Tables {
				Must(db.Id(t.GUID).Delete(new(TableDict)))
				Must(db.Where("table_guid=?", t.GUID).Delete(new(ColumnsDict)))
				Must(db.InsertOne(t))
				Must(db.Insert(t.Columns))
			}
		}
		fmt.Println("成功保存到数据库.")
	}

}

func getDictByTag(f *ast.Field) (string, string) {

	if f.Tag == nil {
		return "", ""
	}
	tag := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
	value, ok := tag.Lookup("dict")

	if !ok {
		return "", ""
	}

	tmps := strings.Fields(value)
	switch l := len(temp); l {
	case 1:
		return tmps[0], ""
	case 0:
		return "", ""
	default:

		return tmps[0], strings.Join(tmps[1:], " ")
	}

}
