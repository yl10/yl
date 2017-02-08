package columns

//@TableName tableA
type A struct {
	ID  int
	AAA string
	BBB int
}

//@TableName
type D struct {
	GUID       string
	ParentGUID string
	AAA        string
	C          bool
}

//@TableName tableE
type E struct {
	AAA string
}
