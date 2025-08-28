package bench_test

// op tags used for per-tool filtering (if needed later)
const (
	OpDefault     = "default"     // ${A:-a}
	OpReplace1    = "repl1"       // ${X/foo/bar}
	OpReplaceAll  = "replAll"     // ${X//foo/bar}
	OpTrimPref1   = "trimPre1"    // ${Y#pre}
	OpTrimPrefAll = "trimPreAll"  // ${Y##pre}
	OpTrimSuf1    = "trimSuf1"    // ${Z%ing}
	OpTrimSufAll  = "trimSufAll"  // ${Z%%ing}
	OpCaseFirstUp = "caseFirstUp" // ${c^}
	OpCaseAllUp   = "caseAllUp"   // ${c^^}
	OpCaseFirstLo = "caseFirstLo" // ${C,}
	OpCaseAllLo   = "caseAllLo"   // ${C,,}
	OpSubstr      = "substr"      // ${S:off[:len]}
	OpNested      = "nested"      // ${N:-${FALLBACK}}
)

type benchCase struct {
	name    string
	pattern string
	env     []string // complete env for the child; anything not listed stays unset
	ops     []string // which features this case exercises
}

var extendedCases = []benchCase{
	{"DefaultA_Unset_B_Set", "A=${A:-a} B=${B} LITERAL-0123456789\n", []string{"B=bee"}, []string{OpDefault}},
	{"DefaultA_Set_B_Set", "A=${A:-a} B=${B} LITERAL-0123456789\n", []string{"A=Alpha", "B=bee"}, []string{OpDefault}},
	{"ReplaceFirst", "X=${X/foo/bar}\n", []string{"X=foo foo"}, []string{OpReplace1}},
	{"ReplaceAll", "X=${X//foo/bar}\n", []string{"X=foo foo foo"}, []string{OpReplaceAll}},
	{"TrimPrefixOnce", "Y=${Y#pre}\n", []string{"Y=preprefix"}, []string{OpTrimPref1}},
	{"TrimPrefixAll", "Y=${Y##pre}\n", []string{"Y=prepreX"}, []string{OpTrimPrefAll}},
	{"TrimSuffixOnce", "Z=${Z%ing}\n", []string{"Z=ending"}, []string{OpTrimSuf1}},
	{"TrimSuffixAll", "Z=${Z%%ing}\n", []string{"Z=endinging"}, []string{OpTrimSufAll}},
	{"Case_FirstUpper", "c=${c^}\n", []string{"c=äbc"}, []string{OpCaseFirstUp}},
	{"Case_AllUpper", "c=${c^^}\n", []string{"c=ÄbÇd"}, []string{OpCaseAllUp}},
	{"Case_FirstLower", "C=${C,}\n", []string{"C=ÄBC"}, []string{OpCaseFirstLo}},
	{"Case_AllLower", "C=${C,,}\n", []string{"C=ÄbÇd"}, []string{OpCaseAllLo}},
	{"Substr_PosLen", "S=${S:2:3}\n", []string{"S=abcdef"}, []string{OpSubstr}},
	{"Substr_NegFromEnd", "S=${S:-3:2}\n", []string{"S=abcdef"}, []string{OpSubstr}},
	{"Nested_Default_FB", "N=${N:-${FALLBACK}}\n", []string{"FALLBACK=fb"}, []string{OpNested, OpDefault}},
	{"ReplaceToEmpty", "E=${E/a/}\n", []string{"E=a"}, []string{OpReplace1}},
	{
		"Mixed_Realistic", "A=${A:-default} X=${X//foo/bar} Y=${Y##pre} Z=${Z%ing} C=${C^^} S=${S:-abc}\n",
		[]string{"X=foo foo foo", "Y=prepreY", "Z=ending", "C=Go-lang"},
		[]string{OpDefault, OpReplaceAll, OpTrimPrefAll, OpTrimSuf1, OpCaseAllUp},
	},
}
