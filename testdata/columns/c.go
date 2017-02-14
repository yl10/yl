package columns

//@TableName tableA
type A struct {
	ID  int
	AAA string
	BBB int
}

//@TableName @tablecnname D  表格dddd哦
type D struct {
	GUID       string `dict:"aaa   beizh  "`
	ParentGUID string `dict:"父GUID   父GUID 你看不懂啊  "`
	AAA        string
	C          bool
}

//@TableName tableE @tablecnname E 表格eeeed哦
type E struct {
	AAA string
}
