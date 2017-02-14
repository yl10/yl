package main

import (
	"flag"
)

var (
	flagTagName     = flag.String("tag", "xorm", "tagname")
	flagPath        = flag.String("path", ".", "指定路径")
	flagExpectFile  = flag.Bool("savefile", false, "是否保存到文件")
	flagExpectMysql = flag.Bool("mysql", false, "是否连接mysql")
	flagFilePath    = flag.String("file", "_", "指定保存文件的路径")
	flagMapRole     = flag.Int("m", colGonicMapper, "名称映射模式，0:SnakeMapper ;1:SameMapper ;2:GonicMapper ;3:没有tag的字段忽略.默认：2")
	flagPrefix      = flag.String("prefix", "", "表名的前缀")
	flagdebug       = flag.Bool("debug", false, "show generated data")
	flagDsn         = flag.String("dsn", "_", "数据库连接字符，只支持mysql")
)

//Must 忽略错误，抛出异常，入口需做错误处理
func Must(result ...interface{}) {
	l := len(result)
	if l > 0 {
		err, ok := result[l-1].(error)
		if ok {
			panic(err)
		}
	}
}
