# yl gencolumns #
## 功能说明 ##
		在使用ORM或者直接拼SQL语句的时候，字段名都是直接写出来的，写错了或被修改了，编译的时候也不会报错，从而导致最终在数据库里执行的SQL语句有问题。
		genclomns会自动创建_columns.go文件，生成和相应struct完全一样的字段。比如：
## tag说明 ##

		-tag	定义字段时用的tag标记，默认是xorm。
		-path	需要解析的目录，默认是当前目录。注意：目录会一直往下递归。
		-m		名称映射模式。0:SnakeMapper ;1:SameMapper ;2:GonicMapper ;3:没有tag的字段忽略.默认：2
		-prefix	生成表名的时候希望增加的前缀，如果注释中制定了表名，前缀不生效。
		-debug	在终端显示生成的内容。
		
## 例子 ##

		User.go文件

```go
		package main
		//User @tablename user
		type User struct{
			ID int
			Name string
		}
```


		执行 yl gencolumns 

		自动创建 User_columns.go
```go
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

```	

		拼SQL的时候，直接使用UserColumns.Name 来替代 "name",UserColumns.TableName 来表示表名。

## QA ##
		具备哪些条件的sturct定义才会生成?
		1.需要加注释，注释中包含 @tablename，不区分大小写。
		2.可导出的类型，即首字母要大写。

		