package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	identifier     = regexp.MustCompile("^[A-Za-z]+[A-Za-z0-9]")
	prefixVariable = regexp.MustCompile("{\\w*}")
	defaultPrefix  = &ScopePrefix{String: "", Variables: make([]string, 0)}
)

type statementWrapper struct {
	comment   []string
	statement interface{}
}

type exception *Struct

type union *Struct

type include string

func newScopePrefix(prefix string) (*ScopePrefix, error) {
	variables := []string{}
	for _, variable := range prefixVariable.FindAllString(prefix, -1) {
		variable = variable[1 : len(variable)-1]
		if len(variable) == 0 || !identifier.MatchString(variable) {
			return nil, fmt.Errorf("parser: invalid prefix variable '%s'", variable)
		}
		variables = append(variables, variable)
	}
	return &ScopePrefix{String: prefix, Variables: variables}, nil
}

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

func ifaceSliceToString(v interface{}) string {
	ifs := toIfaceSlice(v)
	b := make([]byte, len(ifs))
	for i, v := range ifs {
		b[i] = v.([]uint8)[0]
	}
	return string(b)
}

func rawCommentToDocStr(raw string) []string {
	rawLines := strings.Split(raw, "\n")
	comment := make([]string, len(rawLines))
	for i, line := range rawLines {
		comment[i] = strings.TrimLeft(line, "* ")
	}
	return comment
}

// toStruct converts a union to a struct with all fields optional.
func unionToStruct(u union) *Struct {
	st := (*Struct)(u)
	for _, f := range st.Fields {
		f.Optional = true
	}
	return st
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 79, col: 1, offset: 2185},
			expr: &actionExpr{
				pos: position{line: 79, col: 11, offset: 2197},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 79, col: 11, offset: 2197},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 79, col: 11, offset: 2197},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 79, col: 14, offset: 2200},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 79, col: 25, offset: 2211},
								expr: &seqExpr{
									pos: position{line: 79, col: 27, offset: 2213},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 79, col: 27, offset: 2213},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 79, col: 37, offset: 2223},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 79, col: 44, offset: 2230},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 79, col: 44, offset: 2230},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 79, col: 50, offset: 2236},
									name: "SyntaxError",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SyntaxError",
			pos:  position{line: 152, col: 1, offset: 4845},
			expr: &actionExpr{
				pos: position{line: 152, col: 15, offset: 4861},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 152, col: 15, offset: 4861,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 156, col: 1, offset: 4919},
			expr: &actionExpr{
				pos: position{line: 156, col: 13, offset: 4933},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 156, col: 13, offset: 4933},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 156, col: 13, offset: 4933},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 156, col: 20, offset: 4940},
								expr: &seqExpr{
									pos: position{line: 156, col: 21, offset: 4941},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 156, col: 21, offset: 4941},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 156, col: 31, offset: 4951},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 156, col: 36, offset: 4956},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 156, col: 47, offset: 4967},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 156, col: 47, offset: 4967},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 156, col: 65, offset: 4985},
										name: "FrugalStatement",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ThriftStatement",
			pos:  position{line: 169, col: 1, offset: 5456},
			expr: &choiceExpr{
				pos: position{line: 169, col: 19, offset: 5476},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 169, col: 19, offset: 5476},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 29, offset: 5486},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 41, offset: 5498},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 49, offset: 5506},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 56, offset: 5513},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 66, offset: 5523},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 75, offset: 5532},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 87, offset: 5544},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 95, offset: 5552},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 171, col: 1, offset: 5561},
			expr: &actionExpr{
				pos: position{line: 171, col: 11, offset: 5573},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 171, col: 11, offset: 5573},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 171, col: 11, offset: 5573},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 21, offset: 5583},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 171, col: 23, offset: 5585},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 171, col: 28, offset: 5590},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 36, offset: 5598},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 175, col: 1, offset: 5646},
			expr: &actionExpr{
				pos: position{line: 175, col: 13, offset: 5660},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 175, col: 13, offset: 5660},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 13, offset: 5660},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 25, offset: 5672},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 27, offset: 5674},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 175, col: 33, offset: 5680},
								expr: &charClassMatcher{
									pos:        position{line: 175, col: 33, offset: 5680},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 43, offset: 5690},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 45, offset: 5692},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 48, offset: 5695},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 59, offset: 5706},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 182, col: 1, offset: 5831},
			expr: &actionExpr{
				pos: position{line: 182, col: 9, offset: 5841},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 182, col: 9, offset: 5841},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 182, col: 9, offset: 5841},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 17, offset: 5849},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 19, offset: 5851},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 23, offset: 5855},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 33, offset: 5865},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 35, offset: 5867},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 40, offset: 5872},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 51, offset: 5883},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 182, col: 53, offset: 5885},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 57, offset: 5889},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 59, offset: 5891},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 65, offset: 5897},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 76, offset: 5908},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 190, col: 1, offset: 6040},
			expr: &actionExpr{
				pos: position{line: 190, col: 8, offset: 6049},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 190, col: 8, offset: 6049},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 190, col: 8, offset: 6049},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 15, offset: 6056},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 17, offset: 6058},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 22, offset: 6063},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 33, offset: 6074},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 190, col: 36, offset: 6077},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 40, offset: 6081},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 43, offset: 6084},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 190, col: 50, offset: 6091},
								expr: &seqExpr{
									pos: position{line: 190, col: 51, offset: 6092},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 190, col: 51, offset: 6092},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 190, col: 61, offset: 6102},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 190, col: 66, offset: 6107},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 70, offset: 6111},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 213, col: 1, offset: 6723},
			expr: &actionExpr{
				pos: position{line: 213, col: 13, offset: 6737},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 213, col: 13, offset: 6737},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 213, col: 13, offset: 6737},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 20, offset: 6744},
								expr: &seqExpr{
									pos: position{line: 213, col: 21, offset: 6745},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 213, col: 21, offset: 6745},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 31, offset: 6755},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 213, col: 36, offset: 6760},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 41, offset: 6765},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 213, col: 52, offset: 6776},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 213, col: 54, offset: 6778},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 60, offset: 6784},
								expr: &seqExpr{
									pos: position{line: 213, col: 61, offset: 6785},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 213, col: 61, offset: 6785},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 65, offset: 6789},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 67, offset: 6791},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 213, col: 81, offset: 6805},
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 81, offset: 6805},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 228, col: 1, offset: 7141},
			expr: &actionExpr{
				pos: position{line: 228, col: 11, offset: 7153},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 228, col: 11, offset: 7153},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 11, offset: 7153},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 21, offset: 7163},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 23, offset: 7165},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 27, offset: 7169},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 37, offset: 7179},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 39, offset: 7181},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 44, offset: 7186},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 55, offset: 7197},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 235, col: 1, offset: 7306},
			expr: &actionExpr{
				pos: position{line: 235, col: 10, offset: 7317},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 235, col: 10, offset: 7317},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 10, offset: 7317},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 19, offset: 7326},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 21, offset: 7328},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 24, offset: 7331},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 236, col: 1, offset: 7371},
			expr: &actionExpr{
				pos: position{line: 236, col: 13, offset: 7385},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 236, col: 13, offset: 7385},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 236, col: 13, offset: 7385},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 236, col: 25, offset: 7397},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 236, col: 27, offset: 7399},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 236, col: 30, offset: 7402},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 237, col: 1, offset: 7453},
			expr: &actionExpr{
				pos: position{line: 237, col: 9, offset: 7463},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 237, col: 9, offset: 7463},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 237, col: 9, offset: 7463},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 17, offset: 7471},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 19, offset: 7473},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 22, offset: 7476},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 238, col: 1, offset: 7523},
			expr: &actionExpr{
				pos: position{line: 238, col: 14, offset: 7538},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 238, col: 14, offset: 7538},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 238, col: 14, offset: 7538},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 19, offset: 7543},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 30, offset: 7554},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 238, col: 33, offset: 7557},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 37, offset: 7561},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7564},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 47, offset: 7571},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 238, col: 57, offset: 7581},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 61, offset: 7585},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 248, col: 1, offset: 7746},
			expr: &actionExpr{
				pos: position{line: 248, col: 13, offset: 7760},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 248, col: 13, offset: 7760},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 248, col: 20, offset: 7767},
						expr: &seqExpr{
							pos: position{line: 248, col: 21, offset: 7768},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 248, col: 21, offset: 7768},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 248, col: 27, offset: 7774},
									name: "__",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Field",
			pos:  position{line: 257, col: 1, offset: 7955},
			expr: &actionExpr{
				pos: position{line: 257, col: 9, offset: 7965},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 257, col: 9, offset: 7965},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 257, col: 9, offset: 7965},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 16, offset: 7972},
								expr: &seqExpr{
									pos: position{line: 257, col: 17, offset: 7973},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 257, col: 17, offset: 7973},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 27, offset: 7983},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 257, col: 32, offset: 7988},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 35, offset: 7991},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 47, offset: 8003},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 257, col: 49, offset: 8005},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 53, offset: 8009},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 55, offset: 8011},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 59, offset: 8015},
								expr: &ruleRefExpr{
									pos:  position{line: 257, col: 59, offset: 8015},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 69, offset: 8025},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 71, offset: 8027},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 75, offset: 8031},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 85, offset: 8041},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 87, offset: 8043},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 92, offset: 8048},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 103, offset: 8059},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 106, offset: 8062},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 110, offset: 8066},
								expr: &seqExpr{
									pos: position{line: 257, col: 111, offset: 8067},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 257, col: 111, offset: 8067},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 115, offset: 8071},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 117, offset: 8073},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 257, col: 130, offset: 8086},
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 130, offset: 8086},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 276, col: 1, offset: 8505},
			expr: &actionExpr{
				pos: position{line: 276, col: 12, offset: 8518},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 276, col: 13, offset: 8519},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 276, col: 13, offset: 8519},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 276, col: 26, offset: 8532},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 280, col: 1, offset: 8606},
			expr: &actionExpr{
				pos: position{line: 280, col: 11, offset: 8618},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 280, col: 11, offset: 8618},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 280, col: 11, offset: 8618},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 21, offset: 8628},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 280, col: 23, offset: 8630},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 280, col: 28, offset: 8635},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 39, offset: 8646},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 280, col: 41, offset: 8648},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 280, col: 49, offset: 8656},
								expr: &seqExpr{
									pos: position{line: 280, col: 50, offset: 8657},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 280, col: 50, offset: 8657},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 60, offset: 8667},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 63, offset: 8670},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 74, offset: 8681},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 79, offset: 8686},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 280, col: 82, offset: 8689},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 86, offset: 8693},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 280, col: 89, offset: 8696},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 280, col: 97, offset: 8704},
								expr: &seqExpr{
									pos: position{line: 280, col: 98, offset: 8705},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 280, col: 98, offset: 8705},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 107, offset: 8714},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 280, col: 113, offset: 8720},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 280, col: 113, offset: 8720},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 280, col: 119, offset: 8726},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 138, offset: 8745},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 296, col: 1, offset: 9129},
			expr: &actionExpr{
				pos: position{line: 296, col: 21, offset: 9151},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 296, col: 21, offset: 9151,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 300, col: 1, offset: 9220},
			expr: &actionExpr{
				pos: position{line: 300, col: 12, offset: 9233},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 300, col: 12, offset: 9233},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 300, col: 12, offset: 9233},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 19, offset: 9240},
								expr: &seqExpr{
									pos: position{line: 300, col: 20, offset: 9241},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 300, col: 20, offset: 9241},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 30, offset: 9251},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 300, col: 35, offset: 9256},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 42, offset: 9263},
								expr: &seqExpr{
									pos: position{line: 300, col: 43, offset: 9264},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 300, col: 43, offset: 9264},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 52, offset: 9273},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 300, col: 57, offset: 9278},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 61, offset: 9282},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 74, offset: 9295},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 77, offset: 9298},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 82, offset: 9303},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 93, offset: 9314},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 300, col: 95, offset: 9316},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 99, offset: 9320},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 102, offset: 9323},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 112, offset: 9333},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 300, col: 122, offset: 9343},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 126, offset: 9347},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 129, offset: 9350},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 140, offset: 9361},
								expr: &ruleRefExpr{
									pos:  position{line: 300, col: 140, offset: 9361},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 300, col: 148, offset: 9369},
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 148, offset: 9369},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 327, col: 1, offset: 9960},
			expr: &actionExpr{
				pos: position{line: 327, col: 16, offset: 9977},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 327, col: 16, offset: 9977},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 327, col: 21, offset: 9982},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 327, col: 21, offset: 9982},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 327, col: 30, offset: 9991},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 334, col: 1, offset: 10113},
			expr: &actionExpr{
				pos: position{line: 334, col: 10, offset: 10124},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 334, col: 10, offset: 10124},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 334, col: 10, offset: 10124},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 334, col: 19, offset: 10133},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 334, col: 22, offset: 10136},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 334, col: 26, offset: 10140},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 334, col: 29, offset: 10143},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 334, col: 40, offset: 10154},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 334, col: 50, offset: 10164},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 338, col: 1, offset: 10200},
			expr: &actionExpr{
				pos: position{line: 338, col: 13, offset: 10214},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 338, col: 13, offset: 10214},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 338, col: 18, offset: 10219},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 338, col: 18, offset: 10219},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 338, col: 29, offset: 10230},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 338, col: 45, offset: 10246},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 345, col: 1, offset: 10371},
			expr: &actionExpr{
				pos: position{line: 345, col: 12, offset: 10384},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 345, col: 13, offset: 10385},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 345, col: 13, offset: 10385},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 22, offset: 10394},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 31, offset: 10403},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 39, offset: 10411},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 47, offset: 10419},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 55, offset: 10427},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 66, offset: 10438},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 77, offset: 10449},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 349, col: 1, offset: 10509},
			expr: &actionExpr{
				pos: position{line: 349, col: 17, offset: 10527},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 349, col: 17, offset: 10527},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 349, col: 22, offset: 10532},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 349, col: 22, offset: 10532},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 32, offset: 10542},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 42, offset: 10552},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 353, col: 1, offset: 10587},
			expr: &actionExpr{
				pos: position{line: 353, col: 11, offset: 10599},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 353, col: 11, offset: 10599},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 353, col: 11, offset: 10599},
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 11, offset: 10599},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 353, col: 20, offset: 10608},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 27, offset: 10615},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 353, col: 30, offset: 10618},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 34, offset: 10622},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 44, offset: 10632},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 353, col: 47, offset: 10635},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 51, offset: 10639},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 353, col: 54, offset: 10642},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 60, offset: 10648},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 70, offset: 10658},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 353, col: 73, offset: 10661},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 361, col: 1, offset: 10784},
			expr: &actionExpr{
				pos: position{line: 361, col: 11, offset: 10796},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 361, col: 11, offset: 10796},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 361, col: 11, offset: 10796},
							expr: &ruleRefExpr{
								pos:  position{line: 361, col: 11, offset: 10796},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 361, col: 20, offset: 10805},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 361, col: 27, offset: 10812},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 361, col: 30, offset: 10815},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 361, col: 34, offset: 10819},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 361, col: 44, offset: 10829},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 361, col: 47, offset: 10832},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 368, col: 1, offset: 10923},
			expr: &actionExpr{
				pos: position{line: 368, col: 12, offset: 10936},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 368, col: 12, offset: 10936},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 368, col: 12, offset: 10936},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 20, offset: 10944},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 368, col: 23, offset: 10947},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 27, offset: 10951},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 37, offset: 10961},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 368, col: 40, offset: 10964},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 375, col: 1, offset: 11056},
			expr: &actionExpr{
				pos: position{line: 375, col: 11, offset: 11068},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 375, col: 11, offset: 11068},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 375, col: 11, offset: 11068},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 375, col: 22, offset: 11079},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 30, offset: 11087},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 379, col: 1, offset: 11124},
			expr: &choiceExpr{
				pos: position{line: 379, col: 14, offset: 11139},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 379, col: 14, offset: 11139},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 24, offset: 11149},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 41, offset: 11166},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 55, offset: 11180},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 66, offset: 11191},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 78, offset: 11203},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 381, col: 1, offset: 11215},
			expr: &actionExpr{
				pos: position{line: 381, col: 15, offset: 11231},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 381, col: 15, offset: 11231},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 381, col: 15, offset: 11231},
							expr: &charClassMatcher{
								pos:        position{line: 381, col: 15, offset: 11231},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 381, col: 21, offset: 11237},
							expr: &ruleRefExpr{
								pos:  position{line: 381, col: 21, offset: 11237},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 385, col: 1, offset: 11301},
			expr: &actionExpr{
				pos: position{line: 385, col: 18, offset: 11320},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 385, col: 18, offset: 11320},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 385, col: 18, offset: 11320},
							expr: &charClassMatcher{
								pos:        position{line: 385, col: 18, offset: 11320},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 385, col: 24, offset: 11326},
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 24, offset: 11326},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 385, col: 31, offset: 11333},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 385, col: 35, offset: 11337},
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 35, offset: 11337},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 385, col: 42, offset: 11344},
							expr: &seqExpr{
								pos: position{line: 385, col: 44, offset: 11346},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 385, col: 44, offset: 11346},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 385, col: 51, offset: 11353},
										name: "IntConstant",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ConstList",
			pos:  position{line: 389, col: 1, offset: 11423},
			expr: &actionExpr{
				pos: position{line: 389, col: 13, offset: 11437},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 389, col: 13, offset: 11437},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 389, col: 13, offset: 11437},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 17, offset: 11441},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 20, offset: 11444},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 389, col: 27, offset: 11451},
								expr: &seqExpr{
									pos: position{line: 389, col: 28, offset: 11452},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 389, col: 28, offset: 11452},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 389, col: 39, offset: 11463},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 389, col: 42, offset: 11466},
											expr: &ruleRefExpr{
												pos:  position{line: 389, col: 42, offset: 11466},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 389, col: 57, offset: 11481},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 62, offset: 11486},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 389, col: 65, offset: 11489},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 398, col: 1, offset: 11683},
			expr: &actionExpr{
				pos: position{line: 398, col: 12, offset: 11696},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 398, col: 12, offset: 11696},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 398, col: 12, offset: 11696},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 16, offset: 11700},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 398, col: 19, offset: 11703},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 398, col: 26, offset: 11710},
								expr: &seqExpr{
									pos: position{line: 398, col: 27, offset: 11711},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 398, col: 27, offset: 11711},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 38, offset: 11722},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 398, col: 41, offset: 11725},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 45, offset: 11729},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 48, offset: 11732},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 59, offset: 11743},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 398, col: 63, offset: 11747},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 398, col: 63, offset: 11747},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 398, col: 69, offset: 11753},
													expr: &litMatcher{
														pos:        position{line: 398, col: 70, offset: 11754},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 75, offset: 11759},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 398, col: 80, offset: 11764},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 418, col: 1, offset: 12314},
			expr: &choiceExpr{
				pos: position{line: 418, col: 19, offset: 12334},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 418, col: 19, offset: 12334},
						name: "Scope",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 27, offset: 12342},
						name: "Async",
					},
				},
			},
		},
		{
			name: "Scope",
			pos:  position{line: 420, col: 1, offset: 12349},
			expr: &actionExpr{
				pos: position{line: 420, col: 9, offset: 12359},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 420, col: 9, offset: 12359},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 420, col: 9, offset: 12359},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 420, col: 16, offset: 12366},
								expr: &seqExpr{
									pos: position{line: 420, col: 17, offset: 12367},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 420, col: 17, offset: 12367},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 420, col: 27, offset: 12377},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 420, col: 32, offset: 12382},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 40, offset: 12390},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 43, offset: 12393},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 420, col: 48, offset: 12398},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 59, offset: 12409},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 62, offset: 12412},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 420, col: 69, offset: 12419},
								expr: &ruleRefExpr{
									pos:  position{line: 420, col: 69, offset: 12419},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 77, offset: 12427},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 420, col: 80, offset: 12430},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 84, offset: 12434},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 87, offset: 12437},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 420, col: 98, offset: 12448},
								expr: &seqExpr{
									pos: position{line: 420, col: 99, offset: 12449},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 420, col: 99, offset: 12449},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 420, col: 109, offset: 12459},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 420, col: 115, offset: 12465},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 420, col: 115, offset: 12465},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 420, col: 121, offset: 12471},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 138, offset: 12488},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 441, col: 1, offset: 13033},
			expr: &actionExpr{
				pos: position{line: 441, col: 19, offset: 13053},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 441, col: 19, offset: 13053,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 445, col: 1, offset: 13120},
			expr: &actionExpr{
				pos: position{line: 445, col: 10, offset: 13131},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 445, col: 10, offset: 13131},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 445, col: 10, offset: 13131},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 19, offset: 13140},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 445, col: 22, offset: 13143},
							label: "name",
							expr: &zeroOrMoreExpr{
								pos: position{line: 445, col: 27, offset: 13148},
								expr: &charClassMatcher{
									pos:        position{line: 445, col: 27, offset: 13148},
									val:        "[^\\r\\n\\t\\f ]",
									chars:      []rune{'\r', '\n', '\t', '\f', ' '},
									ignoreCase: false,
									inverted:   true,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 449, col: 1, offset: 13220},
			expr: &actionExpr{
				pos: position{line: 449, col: 13, offset: 13234},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 449, col: 13, offset: 13234},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 449, col: 13, offset: 13234},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 449, col: 20, offset: 13241},
								expr: &seqExpr{
									pos: position{line: 449, col: 21, offset: 13242},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 449, col: 21, offset: 13242},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 449, col: 31, offset: 13252},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 449, col: 36, offset: 13257},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 449, col: 41, offset: 13262},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 52, offset: 13273},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 449, col: 54, offset: 13275},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 58, offset: 13279},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 449, col: 61, offset: 13282},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 449, col: 67, offset: 13288},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 78, offset: 13299},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Async",
			pos:  position{line: 461, col: 1, offset: 13560},
			expr: &actionExpr{
				pos: position{line: 461, col: 9, offset: 13570},
				run: (*parser).callonAsync1,
				expr: &seqExpr{
					pos: position{line: 461, col: 9, offset: 13570},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 461, col: 9, offset: 13570},
							val:        "async",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 17, offset: 13578},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 19, offset: 13580},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 461, col: 24, offset: 13585},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 35, offset: 13596},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 37, offset: 13598},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 461, col: 45, offset: 13606},
								expr: &seqExpr{
									pos: position{line: 461, col: 46, offset: 13607},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 461, col: 46, offset: 13607},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 56, offset: 13617},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 59, offset: 13620},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 70, offset: 13631},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 75, offset: 13636},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 461, col: 78, offset: 13639},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 82, offset: 13643},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 85, offset: 13646},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 461, col: 93, offset: 13654},
								expr: &seqExpr{
									pos: position{line: 461, col: 94, offset: 13655},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 461, col: 94, offset: 13655},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 103, offset: 13664},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 461, col: 109, offset: 13670},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 461, col: 109, offset: 13670},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 461, col: 115, offset: 13676},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 134, offset: 13695},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 481, col: 1, offset: 14326},
			expr: &actionExpr{
				pos: position{line: 481, col: 11, offset: 14338},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 481, col: 12, offset: 14339},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 481, col: 13, offset: 14340},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 481, col: 13, offset: 14340},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 481, col: 17, offset: 14344},
									expr: &choiceExpr{
										pos: position{line: 481, col: 18, offset: 14345},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 481, col: 18, offset: 14345},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 481, col: 25, offset: 14352},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 481, col: 32, offset: 14359},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 481, col: 40, offset: 14367},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 481, col: 40, offset: 14367},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 481, col: 45, offset: 14372},
									expr: &choiceExpr{
										pos: position{line: 481, col: 46, offset: 14373},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 481, col: 46, offset: 14373},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 481, col: 53, offset: 14380},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 481, col: 60, offset: 14387},
									val:        "'",
									ignoreCase: false,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Identifier",
			pos:  position{line: 488, col: 1, offset: 14603},
			expr: &actionExpr{
				pos: position{line: 488, col: 14, offset: 14618},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 488, col: 14, offset: 14618},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 488, col: 14, offset: 14618},
							expr: &choiceExpr{
								pos: position{line: 488, col: 15, offset: 14619},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 488, col: 15, offset: 14619},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 488, col: 24, offset: 14628},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 488, col: 30, offset: 14634},
							expr: &choiceExpr{
								pos: position{line: 488, col: 31, offset: 14635},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 488, col: 31, offset: 14635},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 488, col: 40, offset: 14644},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 488, col: 48, offset: 14652},
										val:        "[._]",
										chars:      []rune{'.', '_'},
										ignoreCase: false,
										inverted:   false,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ListSeparator",
			pos:  position{line: 492, col: 1, offset: 14707},
			expr: &charClassMatcher{
				pos:        position{line: 492, col: 17, offset: 14725},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 493, col: 1, offset: 14730},
			expr: &charClassMatcher{
				pos:        position{line: 493, col: 10, offset: 14741},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 494, col: 1, offset: 14750},
			expr: &charClassMatcher{
				pos:        position{line: 494, col: 9, offset: 14760},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 496, col: 1, offset: 14767},
			expr: &anyMatcher{
				line: 496, col: 14, offset: 14782,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 497, col: 1, offset: 14784},
			expr: &actionExpr{
				pos: position{line: 497, col: 13, offset: 14798},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 497, col: 13, offset: 14798},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 497, col: 13, offset: 14798},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 497, col: 20, offset: 14805},
							expr: &seqExpr{
								pos: position{line: 497, col: 22, offset: 14807},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 497, col: 22, offset: 14807},
										expr: &litMatcher{
											pos:        position{line: 497, col: 23, offset: 14808},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 497, col: 28, offset: 14813},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 497, col: 42, offset: 14827},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 503, col: 1, offset: 15007},
			expr: &choiceExpr{
				pos: position{line: 503, col: 11, offset: 15019},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 503, col: 11, offset: 15019},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 503, col: 30, offset: 15038},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 504, col: 1, offset: 15056},
			expr: &seqExpr{
				pos: position{line: 504, col: 20, offset: 15077},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 504, col: 20, offset: 15077},
						expr: &ruleRefExpr{
							pos:  position{line: 504, col: 21, offset: 15078},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 504, col: 31, offset: 15088},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 504, col: 36, offset: 15093},
						expr: &seqExpr{
							pos: position{line: 504, col: 38, offset: 15095},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 504, col: 38, offset: 15095},
									expr: &litMatcher{
										pos:        position{line: 504, col: 39, offset: 15096},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 504, col: 44, offset: 15101},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 504, col: 58, offset: 15115},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 505, col: 1, offset: 15120},
			expr: &seqExpr{
				pos: position{line: 505, col: 36, offset: 15157},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 505, col: 36, offset: 15157},
						expr: &ruleRefExpr{
							pos:  position{line: 505, col: 37, offset: 15158},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 505, col: 47, offset: 15168},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 505, col: 52, offset: 15173},
						expr: &seqExpr{
							pos: position{line: 505, col: 54, offset: 15175},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 505, col: 54, offset: 15175},
									expr: &choiceExpr{
										pos: position{line: 505, col: 57, offset: 15178},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 505, col: 57, offset: 15178},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 505, col: 64, offset: 15185},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 505, col: 70, offset: 15191},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 505, col: 84, offset: 15205},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 506, col: 1, offset: 15210},
			expr: &choiceExpr{
				pos: position{line: 506, col: 21, offset: 15232},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 506, col: 22, offset: 15233},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 506, col: 22, offset: 15233},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 506, col: 27, offset: 15238},
								expr: &seqExpr{
									pos: position{line: 506, col: 29, offset: 15240},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 506, col: 29, offset: 15240},
											expr: &ruleRefExpr{
												pos:  position{line: 506, col: 30, offset: 15241},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 506, col: 34, offset: 15245},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 506, col: 52, offset: 15263},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 506, col: 52, offset: 15263},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 506, col: 56, offset: 15267},
								expr: &seqExpr{
									pos: position{line: 506, col: 58, offset: 15269},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 506, col: 58, offset: 15269},
											expr: &ruleRefExpr{
												pos:  position{line: 506, col: 59, offset: 15270},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 506, col: 63, offset: 15274},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "__",
			pos:  position{line: 508, col: 1, offset: 15290},
			expr: &zeroOrMoreExpr{
				pos: position{line: 508, col: 6, offset: 15297},
				expr: &choiceExpr{
					pos: position{line: 508, col: 8, offset: 15299},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 508, col: 8, offset: 15299},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 21, offset: 15312},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 27, offset: 15318},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 509, col: 1, offset: 15329},
			expr: &zeroOrMoreExpr{
				pos: position{line: 509, col: 5, offset: 15335},
				expr: &choiceExpr{
					pos: position{line: 509, col: 7, offset: 15337},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 509, col: 7, offset: 15337},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 509, col: 20, offset: 15350},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 510, col: 1, offset: 15386},
			expr: &zeroOrMoreExpr{
				pos: position{line: 510, col: 6, offset: 15393},
				expr: &ruleRefExpr{
					pos:  position{line: 510, col: 6, offset: 15393},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 512, col: 1, offset: 15406},
			expr: &charClassMatcher{
				pos:        position{line: 512, col: 14, offset: 15421},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 513, col: 1, offset: 15429},
			expr: &litMatcher{
				pos:        position{line: 513, col: 7, offset: 15437},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 514, col: 1, offset: 15442},
			expr: &choiceExpr{
				pos: position{line: 514, col: 7, offset: 15450},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 514, col: 7, offset: 15450},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 7, offset: 15450},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 514, col: 10, offset: 15453},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 514, col: 16, offset: 15459},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 16, offset: 15459},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 514, col: 18, offset: 15461},
								expr: &ruleRefExpr{
									pos:  position{line: 514, col: 18, offset: 15461},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 514, col: 37, offset: 15480},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 514, col: 43, offset: 15486},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 43, offset: 15486},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 514, col: 46, offset: 15489},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 516, col: 1, offset: 15494},
			expr: &notExpr{
				pos: position{line: 516, col: 7, offset: 15502},
				expr: &anyMatcher{
					line: 516, col: 8, offset: 15503,
				},
			},
		},
	},
}

func (c *current) onGrammar1(statements interface{}) (interface{}, error) {
	thrift := &Thrift{
		Includes:       []*Include{},
		Namespaces:     []*Namespace{},
		Typedefs:       []*TypeDef{},
		Constants:      []*Constant{},
		Enums:          []*Enum{},
		Structs:        []*Struct{},
		Exceptions:     []*Struct{},
		Unions:         []*Struct{},
		Services:       []*Service{},
		typedefIndex:   make(map[string]*TypeDef),
		namespaceIndex: make(map[string]*Namespace),
	}
	frugal := &Frugal{
		Thrift:         thrift,
		Scopes:         []*Scope{},
		Asyncs:         []*Async{},
		ParsedIncludes: make(map[string]*Frugal),
	}

	stmts := toIfaceSlice(statements)
	for _, st := range stmts {
		wrapper := st.([]interface{})[0].(*statementWrapper)
		switch v := wrapper.statement.(type) {
		case *Namespace:
			thrift.Namespaces = append(thrift.Namespaces, v)
			thrift.namespaceIndex[v.Scope] = v
		case *Constant:
			v.Comment = wrapper.comment
			thrift.Constants = append(thrift.Constants, v)
		case *Enum:
			v.Comment = wrapper.comment
			thrift.Enums = append(thrift.Enums, v)
		case *TypeDef:
			v.Comment = wrapper.comment
			thrift.Typedefs = append(thrift.Typedefs, v)
			thrift.typedefIndex[v.Name] = v
		case *Struct:
			v.Comment = wrapper.comment
			thrift.Structs = append(thrift.Structs, v)
		case exception:
			strct := (*Struct)(v)
			strct.Comment = wrapper.comment
			thrift.Exceptions = append(thrift.Exceptions, strct)
		case union:
			strct := unionToStruct(v)
			strct.Comment = wrapper.comment
			thrift.Unions = append(thrift.Unions, strct)
		case *Service:
			v.Comment = wrapper.comment
			thrift.Services = append(thrift.Services, v)
		case include:
			name := string(v)
			if ix := strings.LastIndex(name, "."); ix > 0 {
				name = name[:ix]
			}
			thrift.Includes = append(thrift.Includes, &Include{Name: name, Value: string(v)})
		case *Scope:
			v.Comment = wrapper.comment
			v.Frugal = frugal
			frugal.Scopes = append(frugal.Scopes, v)
		case *Async:
			v.Comment = wrapper.comment
			v.Frugal = frugal
			frugal.Asyncs = append(frugal.Asyncs, v)
		default:
			return nil, fmt.Errorf("parser: unknown value %#v", v)
		}
	}
	return frugal, nil
}

func (p *parser) callonGrammar1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onGrammar1(stack["statements"])
}

func (c *current) onSyntaxError1() (interface{}, error) {
	return nil, errors.New("parser: syntax error")
}

func (p *parser) callonSyntaxError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSyntaxError1()
}

func (c *current) onStatement1(docstr, statement interface{}) (interface{}, error) {
	wrapper := &statementWrapper{statement: statement}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		wrapper.comment = rawCommentToDocStr(raw)
	}
	return wrapper, nil
}

func (p *parser) callonStatement1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStatement1(stack["docstr"], stack["statement"])
}

func (c *current) onInclude1(file interface{}) (interface{}, error) {
	return include(file.(string)), nil
}

func (p *parser) callonInclude1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInclude1(stack["file"])
}

func (c *current) onNamespace1(scope, ns interface{}) (interface{}, error) {
	return &Namespace{
		Scope: ifaceSliceToString(scope),
		Value: string(ns.(Identifier)),
	}, nil
}

func (p *parser) callonNamespace1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNamespace1(stack["scope"], stack["ns"])
}

func (c *current) onConst1(typ, name, value interface{}) (interface{}, error) {
	return &Constant{
		Name:  string(name.(Identifier)),
		Type:  typ.(*Type),
		Value: value,
	}, nil
}

func (p *parser) callonConst1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConst1(stack["typ"], stack["name"], stack["value"])
}

func (c *current) onEnum1(name, values interface{}) (interface{}, error) {
	vs := toIfaceSlice(values)
	en := &Enum{
		Name:   string(name.(Identifier)),
		Values: make(map[string]*EnumValue, len(vs)),
	}
	// Assigns numbers in order. This will behave badly if some values are
	// defined and other are not, but I think that's ok since that's a silly
	// thing to do.
	next := 0
	for _, v := range vs {
		ev := v.([]interface{})[0].(*EnumValue)
		if ev.Value < 0 {
			ev.Value = next
		}
		if ev.Value >= next {
			next = ev.Value + 1
		}
		en.Values[ev.Name] = ev
	}
	return en, nil
}

func (p *parser) callonEnum1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnum1(stack["name"], stack["values"])
}

func (c *current) onEnumValue1(docstr, name, value interface{}) (interface{}, error) {
	ev := &EnumValue{
		Name:  string(name.(Identifier)),
		Value: -1,
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		ev.Comment = rawCommentToDocStr(raw)
	}
	if value != nil {
		ev.Value = int(value.([]interface{})[2].(int64))
	}
	return ev, nil
}

func (p *parser) callonEnumValue1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnumValue1(stack["docstr"], stack["name"], stack["value"])
}

func (c *current) onTypeDef1(typ, name interface{}) (interface{}, error) {
	return &TypeDef{
		Name: string(name.(Identifier)),
		Type: typ.(*Type),
	}, nil
}

func (p *parser) callonTypeDef1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeDef1(stack["typ"], stack["name"])
}

func (c *current) onStruct1(st interface{}) (interface{}, error) {
	return st.(*Struct), nil
}

func (p *parser) callonStruct1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStruct1(stack["st"])
}

func (c *current) onException1(st interface{}) (interface{}, error) {
	return exception(st.(*Struct)), nil
}

func (p *parser) callonException1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onException1(stack["st"])
}

func (c *current) onUnion1(st interface{}) (interface{}, error) {
	return union(st.(*Struct)), nil
}

func (p *parser) callonUnion1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnion1(stack["st"])
}

func (c *current) onStructLike1(name, fields interface{}) (interface{}, error) {
	st := &Struct{
		Name: string(name.(Identifier)),
	}
	if fields != nil {
		st.Fields = fields.([]*Field)
	}
	return st, nil
}

func (p *parser) callonStructLike1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStructLike1(stack["name"], stack["fields"])
}

func (c *current) onFieldList1(fields interface{}) (interface{}, error) {
	fs := fields.([]interface{})
	flds := make([]*Field, len(fs))
	for i, f := range fs {
		flds[i] = f.([]interface{})[0].(*Field)
	}
	return flds, nil
}

func (p *parser) callonFieldList1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFieldList1(stack["fields"])
}

func (c *current) onField1(docstr, id, req, typ, name, def interface{}) (interface{}, error) {
	f := &Field{
		ID:   int(id.(int64)),
		Name: string(name.(Identifier)),
		Type: typ.(*Type),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		f.Comment = rawCommentToDocStr(raw)
	}
	if req != nil && !req.(bool) {
		f.Optional = true
	}
	if def != nil {
		f.Default = def.([]interface{})[2]
	}
	return f, nil
}

func (p *parser) callonField1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onField1(stack["docstr"], stack["id"], stack["req"], stack["typ"], stack["name"], stack["def"])
}

func (c *current) onFieldReq1() (interface{}, error) {
	return !bytes.Equal(c.text, []byte("optional")), nil
}

func (p *parser) callonFieldReq1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFieldReq1()
}

func (c *current) onService1(name, extends, methods interface{}) (interface{}, error) {
	ms := methods.([]interface{})
	svc := &Service{
		Name:    string(name.(Identifier)),
		Methods: make([]*Method, len(ms)),
	}
	if extends != nil {
		svc.Extends = string(extends.([]interface{})[2].(Identifier))
	}
	for i, m := range ms {
		mt := m.([]interface{})[0].(*Method)
		svc.Methods[i] = mt
	}
	return svc, nil
}

func (p *parser) callonService1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onService1(stack["name"], stack["extends"], stack["methods"])
}

func (c *current) onEndOfServiceError1() (interface{}, error) {
	return nil, errors.New("parser: expected end of service")
}

func (p *parser) callonEndOfServiceError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEndOfServiceError1()
}

func (c *current) onFunction1(docstr, oneway, typ, name, arguments, exceptions interface{}) (interface{}, error) {
	m := &Method{
		Name: string(name.(Identifier)),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		m.Comment = rawCommentToDocStr(raw)
	}
	t := typ.(*Type)
	if t.Name != "void" {
		m.ReturnType = t
	}
	if oneway != nil {
		m.Oneway = true
	}
	if arguments != nil {
		m.Arguments = arguments.([]*Field)
	}
	if exceptions != nil {
		m.Exceptions = exceptions.([]*Field)
		for _, e := range m.Exceptions {
			e.Optional = true
		}
	}
	return m, nil
}

func (p *parser) callonFunction1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunction1(stack["docstr"], stack["oneway"], stack["typ"], stack["name"], stack["arguments"], stack["exceptions"])
}

func (c *current) onFunctionType1(typ interface{}) (interface{}, error) {
	if t, ok := typ.(*Type); ok {
		return t, nil
	}
	return &Type{Name: string(c.text)}, nil
}

func (p *parser) callonFunctionType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunctionType1(stack["typ"])
}

func (c *current) onThrows1(exceptions interface{}) (interface{}, error) {
	return exceptions, nil
}

func (p *parser) callonThrows1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onThrows1(stack["exceptions"])
}

func (c *current) onFieldType1(typ interface{}) (interface{}, error) {
	if t, ok := typ.(Identifier); ok {
		return &Type{Name: string(t)}, nil
	}
	return typ, nil
}

func (p *parser) callonFieldType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFieldType1(stack["typ"])
}

func (c *current) onBaseType1() (interface{}, error) {
	return &Type{Name: string(c.text)}, nil
}

func (p *parser) callonBaseType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBaseType1()
}

func (c *current) onContainerType1(typ interface{}) (interface{}, error) {
	return typ, nil
}

func (p *parser) callonContainerType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onContainerType1(stack["typ"])
}

func (c *current) onMapType1(key, value interface{}) (interface{}, error) {
	return &Type{
		Name:      "map",
		KeyType:   key.(*Type),
		ValueType: value.(*Type),
	}, nil
}

func (p *parser) callonMapType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMapType1(stack["key"], stack["value"])
}

func (c *current) onSetType1(typ interface{}) (interface{}, error) {
	return &Type{
		Name:      "set",
		ValueType: typ.(*Type),
	}, nil
}

func (p *parser) callonSetType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSetType1(stack["typ"])
}

func (c *current) onListType1(typ interface{}) (interface{}, error) {
	return &Type{
		Name:      "list",
		ValueType: typ.(*Type),
	}, nil
}

func (p *parser) callonListType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onListType1(stack["typ"])
}

func (c *current) onCppType1(cppType interface{}) (interface{}, error) {
	return cppType, nil
}

func (p *parser) callonCppType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCppType1(stack["cppType"])
}

func (c *current) onIntConstant1() (interface{}, error) {
	return strconv.ParseInt(string(c.text), 10, 64)
}

func (p *parser) callonIntConstant1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIntConstant1()
}

func (c *current) onDoubleConstant1() (interface{}, error) {
	return strconv.ParseFloat(string(c.text), 64)
}

func (p *parser) callonDoubleConstant1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDoubleConstant1()
}

func (c *current) onConstList1(values interface{}) (interface{}, error) {
	valueSlice := values.([]interface{})
	vs := make([]interface{}, len(valueSlice))
	for i, v := range valueSlice {
		vs[i] = v.([]interface{})[0]
	}
	return vs, nil
}

func (p *parser) callonConstList1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConstList1(stack["values"])
}

func (c *current) onConstMap1(values interface{}) (interface{}, error) {
	if values == nil {
		return nil, nil
	}
	vals := values.([]interface{})
	kvs := make([]KeyValue, len(vals))
	for i, kv := range vals {
		v := kv.([]interface{})
		kvs[i] = KeyValue{
			Key:   v[0],
			Value: v[4],
		}
	}
	return kvs, nil
}

func (p *parser) callonConstMap1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConstMap1(stack["values"])
}

func (c *current) onScope1(docstr, name, prefix, operations interface{}) (interface{}, error) {
	ops := operations.([]interface{})
	scope := &Scope{
		Name:       string(name.(Identifier)),
		Operations: make([]*Operation, len(ops)),
		Prefix:     defaultPrefix,
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		scope.Comment = rawCommentToDocStr(raw)
	}
	if prefix != nil {
		scope.Prefix = prefix.(*ScopePrefix)
	}
	for i, o := range ops {
		op := o.([]interface{})[0].(*Operation)
		scope.Operations[i] = op
	}
	return scope, nil
}

func (p *parser) callonScope1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onScope1(stack["docstr"], stack["name"], stack["prefix"], stack["operations"])
}

func (c *current) onEndOfScopeError1() (interface{}, error) {
	return nil, errors.New("parser: expected end of scope")
}

func (p *parser) callonEndOfScopeError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEndOfScopeError1()
}

func (c *current) onPrefix1(name interface{}) (interface{}, error) {
	return newScopePrefix(ifaceSliceToString(name))
}

func (p *parser) callonPrefix1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPrefix1(stack["name"])
}

func (c *current) onOperation1(docstr, name, param interface{}) (interface{}, error) {
	o := &Operation{
		Name:  string(name.(Identifier)),
		Param: string(param.(Identifier)),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		o.Comment = rawCommentToDocStr(raw)
	}
	return o, nil
}

func (p *parser) callonOperation1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onOperation1(stack["docstr"], stack["name"], stack["param"])
}

func (c *current) onAsync1(name, extends, methods interface{}) (interface{}, error) {
	ms := methods.([]interface{})
	async := &Async{
		Name:    string(name.(Identifier)),
		Methods: make([]*Method, len(ms)),
	}
	if extends != nil {
		async.Extends = string(extends.([]interface{})[2].(Identifier))
	}
	for i, m := range ms {
		mt := m.([]interface{})[0].(*Method)
		async.Methods[i] = mt
	}
	return async, nil
}

func (p *parser) callonAsync1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onAsync1(stack["name"], stack["extends"], stack["methods"])
}

func (c *current) onLiteral1() (interface{}, error) {
	if len(c.text) != 0 && c.text[0] == '\'' {
		return strconv.Unquote(`"` + strings.Replace(string(c.text[1:len(c.text)-1]), `\'`, `'`, -1) + `"`)
	}
	return strconv.Unquote(string(c.text))
}

func (p *parser) callonLiteral1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLiteral1()
}

func (c *current) onIdentifier1() (interface{}, error) {
	return Identifier(string(c.text)), nil
}

func (p *parser) callonIdentifier1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIdentifier1()
}

func (c *current) onDocString1() (interface{}, error) {
	comment := string(c.text)
	comment = strings.TrimPrefix(comment, "/**@")
	comment = strings.TrimSuffix(comment, "*/")
	return strings.TrimSpace(comment), nil
}

func (p *parser) callonDocString1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDocString1()
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errNoMatch is returned if no match could be found.
	errNoMatch = errors.New("no match found")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match
}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos  position
	expr interface{}
	run  func(*parser) (interface{}, error)
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos        position
	val        string
	chars      []rune
	ranges     []rune
	classes    []*unicode.RangeTable
	ignoreCase bool
	inverted   bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner  error
	pos    position
	prefix string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
	}
	p.setOptions(opts)
	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v   interface{}
	b   bool
	end savepoint
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	recover bool
	debug   bool
	depth   int

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// stats
	exprCnt int
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth)+">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth)+"<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position)
}

func (p *parser) addErrAt(err error, pos position) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, prefix: buf.String()}
	p.errs.add(pe)
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError {
		if n > 0 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	// start rule is rule [0]
	p.read() // advance to first rune
	val, ok := p.parseRule(g.rules[0])
	if !ok {
		if len(*p.errs) == 0 {
			// make sure this doesn't go out silently
			p.addErr(errNoMatch)
		}
		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint
	var ok bool

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.exprCnt++
	var val interface{}
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position)
		}
		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn != utf8.RuneError {
		start := p.pt
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	// can't match EOF
	if cur == utf8.RuneError {
		return nil, false
	}
	start := p.pt
	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for _, alt := range ch.alternatives {
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			return val, ok
		}
	}
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(not.expr)
	p.popV()
	p.restore(pt)
	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	var vals []interface{}

	pt := p.pt
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}

func rangeTable(class string) *unicode.RangeTable {
	if rt, ok := unicode.Categories[class]; ok {
		return rt
	}
	if rt, ok := unicode.Properties[class]; ok {
		return rt
	}
	if rt, ok := unicode.Scripts[class]; ok {
		return rt
	}

	// cannot happen
	panic(fmt.Sprintf("invalid Unicode class: %s", class))
}
