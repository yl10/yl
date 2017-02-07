package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"time"

	"strings"

	"path/filepath"

	"github.com/pborman/uuid"
)

const (
	spaceUUIDString   = "92f9e6c9-e944-11e6-894d-00155d1f9a01"
	controllerDirName = "controllers"
)
const (
	//B 前端
	B bs = "B"
	//S 服务端
	S bs = "S"
)

var (
	spaceUUID = uuid.Parse(spaceUUIDString)
)
var (
	paramType = _paramtype{
		Body:     "body",
		Header:   "header",
		FormData: "formdata",
		Query:    "query",
		Path:     "path",
	}
)

var (
	controllerDir string
	sqlFileName   string
)
var (
	createTpl = `
			CREATE TABLE {{.tablename}} (
				GUID  varchar(36) NOT NULL ,
				name  varchar(30) NULL ,
				url  varchar(100) NULL ,
				controller  varchar(30) NULL ,
				method  varchar(30) NULL ,
				httpmethod  varchar(30) NULL ,
				bs  varchar(10) NULL ,
				headerParam  varchar(255) NULL ,
				formDataParam  varchar(255) NULL ,
				bodyParam  varchar(255) NULL ,
				queryParam  varchar(255) NULL ,
				pathParam  varchar(255) NULL ,
				remark  varchar(255) NULL ,
				PRIMARY KEY (GUID)
				);`

	insertTpl = `
	{{$tablename:=.tablename}}
	
	{{range .routers}}
	insert {{$tablename}} (GUID,name,url,controller,method,httpmethod,bs,headerParam,formDataParam,bodyParam,queryParam,pathParam,remark)
	values("{{.GUID}}","{{.Name}}","{{.URL}}","{{.Controller}}","{{.Method}}","{{.HTTPMethod}}","{{.BS}}","{{.Remark }}";)
	{{.Params | mustJson}}
	{{range .Params}}
	insert into {{$tablename}}_params(RouterGUID,Name,Type,ValueType,Must,Remark)
	values("{{.RouterGUID}}","{{.Name}}","{{.Type}}","{{.ValueType}}","{{.Must}}","{{.Remark}}";)
	{{end}}
	{{end}}
	`
)

type _paramtype struct {
	Body     string
	Header   string
	FormData string
	Query    string
	Path     string
}

//RouterParam RouterParam
type RouterParam struct {
	RouterGUID string
	Name       string
	Type       string
	ValueType  string
	Must       bool
	Remark     string
}
type bs string

//Router Router
type Router struct {
	GUID       string
	Name       string
	URL        string
	Controller string
	Method     string
	HTTPMethod string
	BS         bs
	Remark     string
	Params     []RouterParam
}

var cmdRouter = &Command{

	UsageLine: "router [-s]",
	Short:     "print Go version",
	Long: `
    
    -s 打印注释书写说明
    -dns 数据库连接字符串
    -db 数据库类型，默认是mysql
    -p controller目录，不设置就为./controller
    -f 是否生成sql文件
    
    `,
}

var (
	routerS      = cmdRouter.Flag.Bool("s", false, "打印注释书写说明。")
	routerDNS    = cmdRouter.Flag.String("dns", "", "数据库连接字符串")
	routerDBType = cmdRouter.Flag.String("db", "mysql", "数据库类型，默认是mysql")

	routerPath = cmdRouter.Flag.String("p", "", "controllers目录")

	routerSQLFile = cmdRouter.Flag.Bool("f", false, "是否生成sql文件")
)
var ()
var routerSimple = `
// @Title 空格后面都是title
// @Description 空格后面都是描述
// @Success 200 {object} models.ZDTProduct.ProductList
// @Param   brand_id    query   int false       "brand id"
// @Param   query   query   string  false       "query of search"
// @Param   segment formData   string  false       "segment"
// @Param   sort    path   string  false       "sort option"
// @Param   dir     body   string  false       "direction asc or desc"
// @Param   offset  header   int     false       "offset"
// @Failure 400 no enough input
// @Failure 500 get products common error
// @router /products [get]
//格式为：@Param   参数名     [参数类型[formData、query、path、body、header，formData]]   [参数值类型] [是否必须]       [参数描述]
`

func init() {
	cmdRouter.Run = runRouter
}
func runRouter(cmd *Command, args []string) {

	if *routerS {
		fmt.Fprintf(os.Stdout, "注释格式如下，可设为代码段。：\r\n%s", routerSimple)
	}

	if *routerSQLFile {
		if fname, err := generateSQLFile(); err != nil {
			fmt.Fprintf(os.Stdout, "生成文件失败：%v", err)
			os.Exit(2)
		} else {
			fmt.Fprintf(os.Stdout, "生成文件成功，文件名为:%s", fname)
		}

	}

}

func generateSQLFile() (string, error) {
	rs, err := parseRouterAll()
	return fmt.Sprintf("%v", rs), err
}

func parseRouter() ([]Router, error) {
	//默认检查controller目录和当前目录
	_, err := os.Open("./controllers")
	if err == nil {
		return parseRouterDir("./controllers")

	}
	return parseRouterDir("./")
}

func parseRouterAll() ([]Router, error) {
	rootName := *routerPath
	os.Open(controllerDir)

	if rootName == "" {
		rootName = controllerDirName
	}

	_, err := os.Open(rootName)
	if err != nil {
		return nil, err
	}
	var rs []Router
	err = filepath.Walk(rootName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			innerrs, err := parseRouterDir(info.Name())
			if err != nil {

				return err
			}
			rs = append(rs, innerrs...)

			return nil
		}
		return nil
	})
	return rs, err
}

func parseRouterDir(pathname string) ([]Router, error) {
	fs := token.NewFileSet()

	pkgs, err := parser.ParseDir(fs, pathname, func(finfo os.FileInfo) bool {
		fname := finfo.Name()
		return !finfo.IsDir() && !strings.HasSuffix(fname, "_test.go") && !strings.HasPrefix(fname, ".") && strings.HasSuffix(fname, ".go")
	}, parser.ParseComments)

	if err != nil {
		return nil, err
	}

	var rs []Router

	for _, pkg := range pkgs {
		for _, fl := range pkg.Files {
			for _, d := range fl.Decls {
				switch specDecl := d.(type) {
				case *ast.FuncDecl:

					comments := specDecl.Doc

					if specDecl.Recv != nil && strings.Contains(comments.Text(), "@router") {

						exp, ok := specDecl.Recv.List[0].Type.(*ast.StarExpr)
						if ok {
							method := specDecl.Name.String()
							controllername := fmt.Sprintf("%v", exp.X)
							rs = append(rs, parseComment(controllername, method, comments))
						}
					}

				}
			}
		}
	}
	return rs, nil

}

func parseComment(cname, method string, comment *ast.CommentGroup) Router {
	r := Router{}

	r.Controller = cname
	r.Method = method

	for _, v := range comment.List {

		s := strings.Fields(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(v.Text, "//"), "/*"), "*/")))
		if len(s) > 1 {
			switch s[0] {
			case "@Title":
				r.Name = strings.Join(s[1:], " ")
			case "@Param":
				if len(s) > 5 {
					must, _ := strconv.ParseBool(s[4])
					remark := strings.Join(s[5:], " ")
					pm := RouterParam{Name: s[1],
						Type:      strings.ToLower(s[2]),
						ValueType: s[3],
						Must:      must,
						Remark:    remark,
					}
					r.Params = append(r.Params, pm)

				}
			case "@Description":
				r.Remark = strings.Join(s[1:], " ")
			case "@router":
				switch len(s) {
				case 2:
					r.URL = s[1]
				case 3:
					r.URL = s[1]
					r.HTTPMethod = s[2][1 : len(s[2])-2]
				}

			}
		}

	}
	r.GUID = getRouterUUID(cname, method, r.URL)
	r.BS = S

	for _, v := range r.Params {
		v.RouterGUID = r.GUID
	}

	return r
}

//getRouterUUID 根据controller名称+方法+路由，生成唯一的uuid，相同参数重复生成的值是一样的。
func getRouterUUID(controller, method, url string) string {
	return uuid.NewMD5(spaceUUID, []byte(controller+method+url)).String()
}

func gettablename() string {
	return time.Now().Format("router_200601021504")
}
func getfilename() string {
	return time.Now().Format("sql_200601021504.sql")
}
