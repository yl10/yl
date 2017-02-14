# yl 忆零团队工具包 #



## gencolumns ##
 根据struct的tag定义，生成_columns.go文件，避免拼SQL的时候出错。
[更多说明](./gencolumns.md)
```text
        -tag	定义字段时用的tag标记，默认是xorm。
		-path	需要解析的目录，默认是当前目录。注意：目录会一直往下递归。
		-m		名称映射模式。0:SnakeMapper ;1:SameMapper ;2:GonicMapper ;3:没有tag的字段忽略.默认：2
		-prefix	生成表名的时候希望增加的前缀，如果注释中制定了表名，前缀不生效。
		-debug	在终端显示生成的内容。
```

## gendict

根据tag中的dict:"中文名  说明" 生成数据字典
flag同上，另外增加：
		-mysql		是否连接mysql。
		-dsn		数据连接字符串

		-savefile	是否保存到文件
		-file		保存文件名称
