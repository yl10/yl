package main

// import (
// 	"fmt"
// 	"go/ast"
// 	"go/parser"
// 	"go/token"
// 	"os"
// 	"strconv"
// 	"time"

// 	"strings"

// 	"path/filepath"

// 	"github.com/pborman/uuid"
// )

// const (
// 	spaceUUIDString   = "92f9e6c9-e944-11e6-894d-00155d1f9a01"
// 	controllerDirName = "controllers"
// )
// const (
// 	//B 前端
// 	B bs = "B"
// 	//S 服务端
// 	S bs = "S"
// )

// var (
// 	spaceUUID = uuid.Parse(spaceUUIDString)
// )
// var (
// 	paramType = _paramtype{
// 		Body:     "body",
// 		Header:   "header",
// 		FormData: "formdata",
// 		Query:    "query",
// 		Path:     "path",
// 	}
// )

// type _paramtype struct {
// 	Body     string
// 	Header   string
// 	FormData string
// 	Query    string
// 	Path     string
// }

// //RouterParam RouterParam
// type RouterParam struct {
// 	RouterGUID string
// 	Name       string
// 	Type       string
// 	ValueType  string
// 	Must       bool
// 	Remark     string
// }

// //Router Router
// type Router struct {
// 	GUID       string
// 	Name       string
// 	URL        string
// 	Controller string
// 	Method     string
// 	HTTPMethod string
// 	BS         bs
// 	Remark     string
// 	Params     []RouterParam
// }

// var cmdRouter = &Command{

// 	UsageLine: "genrouter [-s]",
// 	Short:     "print Go version",
// 	Long: `
//     `,
// }

// var (
// 	routerS      = cmdRouter.Flag.Bool("s", false, "打印注释书写说明。")
// 	routerDNS    = cmdRouter.Flag.String("dns", "", "数据库连接字符串")
// 	routerDBType = cmdRouter.Flag.String("db", "mysql", "数据库类型，默认是mysql")

// 	routerPath = cmdRouter.Flag.String("p", "", "controllers目录")

// 	routerSQLFile = cmdRouter.Flag.Bool("f", false, "是否生成sql文件")
// )
// func init() {
// 	cmdGenRouter.Run = runGenRouter
// }
// func runGenRouter(cmd *Command, args []string) {

// 	if *routerS {
// 		fmt.Fprintf(os.Stdout, "注释格式如下，可设为代码段。：\r\n%s", routerSimple)
// 	}

// 	if *routerSQLFile {
// 		if fname, err := generateSQLFile(); err != nil {
// 			fmt.Fprintf(os.Stdout, "生成文件失败：%v", err)
// 			os.Exit(2)
// 		} else {
// 			fmt.Fprintf(os.Stdout, "生成文件成功，文件名为:%s", fname)
// 		}

// 	}

// }
