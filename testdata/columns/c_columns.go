package columns

var (

	// AColumns a columns name and table name.
	AColumns = _AColumn{
		AAA:       "a_a_a",
		BBB:       "b_b_b",
		ID:        "id",
		TableName: "tableA",
	}

	// DColumns d columns name and table name.
	DColumns = _DColumn{
		AAA:        "a_a_a",
		C:          "c",
		GUID:       "guid",
		ParentGUID: "parent_guid",
		TableName:  "@tablenamecn",
	}

	// EColumns e columns name and table name.
	EColumns = _EColumn{
		AAA:       "a_a_a",
		TableName: "tableE",
	}
)

type _AColumn struct {
	AAA       string
	BBB       string
	ID        string
	TableName string
}

type _DColumn struct {
	AAA        string
	C          string
	GUID       string
	ParentGUID string
	TableName  string
}

type _EColumn struct {
	AAA       string
	TableName string
}
