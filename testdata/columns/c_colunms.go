package columns

type _AColumn struct {
	AAA string
}

// AColumns a columns name
var AColumns _AColumn

type _DColumn struct {
	AAA string
}

// DColumns d columns name
var DColumns _DColumn

type _EColumn struct {
	AAA string
}

// EColumns e columns name
var EColumns _EColumn

func init() {

	AColumns.AAA = "a_a_a"

	DColumns.AAA = "a_a_a"

	EColumns.AAA = "a_a_a"

}
