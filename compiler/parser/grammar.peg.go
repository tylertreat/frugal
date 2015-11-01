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

type namespace struct {
	scope     string
	namespace string
}

type typeDef struct {
	name string
	typ  *Type
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
			pos:  position{line: 75, col: 1, offset: 2362},
			expr: &actionExpr{
				pos: position{line: 75, col: 11, offset: 2374},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 75, col: 11, offset: 2374},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 75, col: 11, offset: 2374},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 75, col: 14, offset: 2377},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 75, col: 25, offset: 2388},
								expr: &seqExpr{
									pos: position{line: 75, col: 27, offset: 2390},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 75, col: 27, offset: 2390},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 75, col: 37, offset: 2400},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 75, col: 44, offset: 2407},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 75, col: 44, offset: 2407},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 75, col: 50, offset: 2413},
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
			pos:  position{line: 126, col: 1, offset: 4420},
			expr: &actionExpr{
				pos: position{line: 126, col: 15, offset: 4436},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 126, col: 15, offset: 4436,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 130, col: 1, offset: 4498},
			expr: &choiceExpr{
				pos: position{line: 130, col: 13, offset: 4512},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 130, col: 13, offset: 4512},
						name: "ThriftStatement",
					},
					&ruleRefExpr{
						pos:  position{line: 130, col: 31, offset: 4530},
						name: "FrugalStatement",
					},
				},
			},
		},
		{
			name: "ThriftStatement",
			pos:  position{line: 136, col: 1, offset: 4788},
			expr: &choiceExpr{
				pos: position{line: 136, col: 19, offset: 4808},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 136, col: 19, offset: 4808},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 29, offset: 4818},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 41, offset: 4830},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 49, offset: 4838},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 56, offset: 4845},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 66, offset: 4855},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 75, offset: 4864},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 87, offset: 4876},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 136, col: 95, offset: 4884},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 138, col: 1, offset: 4893},
			expr: &actionExpr{
				pos: position{line: 138, col: 11, offset: 4905},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 138, col: 11, offset: 4905},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 138, col: 11, offset: 4905},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 138, col: 21, offset: 4915},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 138, col: 23, offset: 4917},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 138, col: 28, offset: 4922},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 138, col: 36, offset: 4930},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 142, col: 1, offset: 4982},
			expr: &actionExpr{
				pos: position{line: 142, col: 13, offset: 4996},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 142, col: 13, offset: 4996},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 142, col: 13, offset: 4996},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 142, col: 25, offset: 5008},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 142, col: 27, offset: 5010},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 142, col: 33, offset: 5016},
								expr: &charClassMatcher{
									pos:        position{line: 142, col: 33, offset: 5016},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 142, col: 42, offset: 5025},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 142, col: 44, offset: 5027},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 142, col: 47, offset: 5030},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 142, col: 58, offset: 5041},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 149, col: 1, offset: 5198},
			expr: &actionExpr{
				pos: position{line: 149, col: 9, offset: 5208},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 149, col: 9, offset: 5208},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 149, col: 9, offset: 5208},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 149, col: 17, offset: 5216},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 149, col: 19, offset: 5218},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 149, col: 23, offset: 5222},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 149, col: 33, offset: 5232},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 149, col: 35, offset: 5234},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 149, col: 40, offset: 5239},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 149, col: 51, offset: 5250},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 149, col: 53, offset: 5252},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 149, col: 57, offset: 5256},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 149, col: 59, offset: 5258},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 149, col: 65, offset: 5264},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 149, col: 76, offset: 5275},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 157, col: 1, offset: 5407},
			expr: &actionExpr{
				pos: position{line: 157, col: 8, offset: 5416},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 157, col: 8, offset: 5416},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 157, col: 8, offset: 5416},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 157, col: 15, offset: 5423},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 157, col: 17, offset: 5425},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 157, col: 22, offset: 5430},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 157, col: 33, offset: 5441},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 157, col: 36, offset: 5444},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 157, col: 40, offset: 5448},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 157, col: 43, offset: 5451},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 157, col: 50, offset: 5458},
								expr: &seqExpr{
									pos: position{line: 157, col: 51, offset: 5459},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 157, col: 51, offset: 5459},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 157, col: 61, offset: 5469},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 157, col: 66, offset: 5474},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 157, col: 70, offset: 5478},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 180, col: 1, offset: 6090},
			expr: &actionExpr{
				pos: position{line: 180, col: 13, offset: 6104},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 180, col: 13, offset: 6104},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 180, col: 13, offset: 6104},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 180, col: 18, offset: 6109},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 180, col: 29, offset: 6120},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 180, col: 31, offset: 6122},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 180, col: 37, offset: 6128},
								expr: &seqExpr{
									pos: position{line: 180, col: 38, offset: 6129},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 180, col: 38, offset: 6129},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 180, col: 42, offset: 6133},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 180, col: 44, offset: 6135},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 180, col: 58, offset: 6149},
							expr: &ruleRefExpr{
								pos:  position{line: 180, col: 58, offset: 6149},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 191, col: 1, offset: 6361},
			expr: &actionExpr{
				pos: position{line: 191, col: 11, offset: 6373},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 191, col: 11, offset: 6373},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 191, col: 11, offset: 6373},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 21, offset: 6383},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 191, col: 23, offset: 6385},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 191, col: 27, offset: 6389},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 37, offset: 6399},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 191, col: 39, offset: 6401},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 191, col: 44, offset: 6406},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 55, offset: 6417},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 198, col: 1, offset: 6525},
			expr: &actionExpr{
				pos: position{line: 198, col: 10, offset: 6536},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 198, col: 10, offset: 6536},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 198, col: 10, offset: 6536},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 19, offset: 6545},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 21, offset: 6547},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 198, col: 24, offset: 6550},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 199, col: 1, offset: 6590},
			expr: &actionExpr{
				pos: position{line: 199, col: 13, offset: 6604},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 199, col: 13, offset: 6604},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 199, col: 13, offset: 6604},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 25, offset: 6616},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 27, offset: 6618},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 199, col: 30, offset: 6621},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 200, col: 1, offset: 6672},
			expr: &actionExpr{
				pos: position{line: 200, col: 9, offset: 6682},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 200, col: 9, offset: 6682},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 200, col: 9, offset: 6682},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 200, col: 17, offset: 6690},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 200, col: 19, offset: 6692},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 200, col: 22, offset: 6695},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 201, col: 1, offset: 6742},
			expr: &actionExpr{
				pos: position{line: 201, col: 14, offset: 6757},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 201, col: 14, offset: 6757},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 201, col: 14, offset: 6757},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 201, col: 19, offset: 6762},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 30, offset: 6773},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 201, col: 33, offset: 6776},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 37, offset: 6780},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 201, col: 40, offset: 6783},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 201, col: 47, offset: 6790},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 201, col: 57, offset: 6800},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 61, offset: 6804},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 211, col: 1, offset: 6965},
			expr: &actionExpr{
				pos: position{line: 211, col: 13, offset: 6979},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 211, col: 13, offset: 6979},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 211, col: 20, offset: 6986},
						expr: &seqExpr{
							pos: position{line: 211, col: 21, offset: 6987},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 211, col: 21, offset: 6987},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 211, col: 27, offset: 6993},
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
			pos:  position{line: 220, col: 1, offset: 7174},
			expr: &actionExpr{
				pos: position{line: 220, col: 9, offset: 7184},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 220, col: 9, offset: 7184},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 220, col: 9, offset: 7184},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 220, col: 12, offset: 7187},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 24, offset: 7199},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 220, col: 26, offset: 7201},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 30, offset: 7205},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 220, col: 32, offset: 7207},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 220, col: 36, offset: 7211},
								expr: &ruleRefExpr{
									pos:  position{line: 220, col: 36, offset: 7211},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 46, offset: 7221},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 220, col: 48, offset: 7223},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 220, col: 52, offset: 7227},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 62, offset: 7237},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 220, col: 64, offset: 7239},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 220, col: 69, offset: 7244},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 220, col: 80, offset: 7255},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 220, col: 83, offset: 7258},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 220, col: 87, offset: 7262},
								expr: &seqExpr{
									pos: position{line: 220, col: 88, offset: 7263},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 220, col: 88, offset: 7263},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 220, col: 92, offset: 7267},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 220, col: 94, offset: 7269},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 220, col: 107, offset: 7282},
							expr: &ruleRefExpr{
								pos:  position{line: 220, col: 107, offset: 7282},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 235, col: 1, offset: 7593},
			expr: &actionExpr{
				pos: position{line: 235, col: 12, offset: 7606},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 235, col: 13, offset: 7607},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 13, offset: 7607},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 235, col: 26, offset: 7620},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 239, col: 1, offset: 7694},
			expr: &actionExpr{
				pos: position{line: 239, col: 11, offset: 7706},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 239, col: 11, offset: 7706},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 239, col: 11, offset: 7706},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 21, offset: 7716},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 239, col: 23, offset: 7718},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 239, col: 28, offset: 7723},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 39, offset: 7734},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 239, col: 41, offset: 7736},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 239, col: 49, offset: 7744},
								expr: &seqExpr{
									pos: position{line: 239, col: 50, offset: 7745},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 239, col: 50, offset: 7745},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 239, col: 60, offset: 7755},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 239, col: 63, offset: 7758},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 239, col: 74, offset: 7769},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 79, offset: 7774},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 239, col: 82, offset: 7777},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 86, offset: 7781},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 239, col: 89, offset: 7784},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 239, col: 97, offset: 7792},
								expr: &seqExpr{
									pos: position{line: 239, col: 98, offset: 7793},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 239, col: 98, offset: 7793},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 239, col: 107, offset: 7802},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 239, col: 113, offset: 7808},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 239, col: 113, offset: 7808},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 239, col: 119, offset: 7814},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 138, offset: 7833},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 254, col: 1, offset: 8228},
			expr: &actionExpr{
				pos: position{line: 254, col: 21, offset: 8250},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 254, col: 21, offset: 8250,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 258, col: 1, offset: 8319},
			expr: &actionExpr{
				pos: position{line: 258, col: 12, offset: 8332},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 258, col: 12, offset: 8332},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 258, col: 12, offset: 8332},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 258, col: 19, offset: 8339},
								expr: &seqExpr{
									pos: position{line: 258, col: 20, offset: 8340},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 258, col: 20, offset: 8340},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 258, col: 29, offset: 8349},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 258, col: 34, offset: 8354},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 258, col: 38, offset: 8358},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 258, col: 51, offset: 8371},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 258, col: 54, offset: 8374},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 258, col: 59, offset: 8379},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 258, col: 70, offset: 8390},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 258, col: 72, offset: 8392},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 258, col: 76, offset: 8396},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 258, col: 79, offset: 8399},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 258, col: 89, offset: 8409},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 258, col: 99, offset: 8419},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 258, col: 103, offset: 8423},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 258, col: 106, offset: 8426},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 258, col: 117, offset: 8437},
								expr: &ruleRefExpr{
									pos:  position{line: 258, col: 117, offset: 8437},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 258, col: 125, offset: 8445},
							expr: &ruleRefExpr{
								pos:  position{line: 258, col: 125, offset: 8445},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 281, col: 1, offset: 8913},
			expr: &actionExpr{
				pos: position{line: 281, col: 16, offset: 8930},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 281, col: 16, offset: 8930},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 281, col: 21, offset: 8935},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 281, col: 21, offset: 8935},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 281, col: 30, offset: 8944},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 288, col: 1, offset: 9066},
			expr: &actionExpr{
				pos: position{line: 288, col: 10, offset: 9077},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 288, col: 10, offset: 9077},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 288, col: 10, offset: 9077},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 288, col: 19, offset: 9086},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 288, col: 22, offset: 9089},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 288, col: 26, offset: 9093},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 288, col: 29, offset: 9096},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 288, col: 40, offset: 9107},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 288, col: 50, offset: 9117},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 292, col: 1, offset: 9153},
			expr: &actionExpr{
				pos: position{line: 292, col: 13, offset: 9167},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 292, col: 13, offset: 9167},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 292, col: 18, offset: 9172},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 292, col: 18, offset: 9172},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 292, col: 29, offset: 9183},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 292, col: 45, offset: 9199},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "DefinitionType",
			pos:  position{line: 299, col: 1, offset: 9324},
			expr: &actionExpr{
				pos: position{line: 299, col: 18, offset: 9343},
				run: (*parser).callonDefinitionType1,
				expr: &labeledExpr{
					pos:   position{line: 299, col: 18, offset: 9343},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 299, col: 23, offset: 9348},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 299, col: 23, offset: 9348},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 299, col: 34, offset: 9359},
								name: "ContainerType",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 303, col: 1, offset: 9399},
			expr: &actionExpr{
				pos: position{line: 303, col: 12, offset: 9412},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 303, col: 13, offset: 9413},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 303, col: 13, offset: 9413},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 303, col: 22, offset: 9422},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 303, col: 31, offset: 9431},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 303, col: 39, offset: 9439},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 303, col: 47, offset: 9447},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 303, col: 55, offset: 9455},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 303, col: 66, offset: 9466},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 303, col: 77, offset: 9477},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 307, col: 1, offset: 9537},
			expr: &actionExpr{
				pos: position{line: 307, col: 17, offset: 9555},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 307, col: 17, offset: 9555},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 307, col: 22, offset: 9560},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 307, col: 22, offset: 9560},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 307, col: 32, offset: 9570},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 307, col: 42, offset: 9580},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 311, col: 1, offset: 9615},
			expr: &actionExpr{
				pos: position{line: 311, col: 11, offset: 9627},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 311, col: 11, offset: 9627},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 311, col: 11, offset: 9627},
							expr: &ruleRefExpr{
								pos:  position{line: 311, col: 11, offset: 9627},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 311, col: 20, offset: 9636},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 311, col: 27, offset: 9643},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 311, col: 30, offset: 9646},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 311, col: 34, offset: 9650},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 311, col: 44, offset: 9660},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 311, col: 47, offset: 9663},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 311, col: 51, offset: 9667},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 311, col: 54, offset: 9670},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 311, col: 60, offset: 9676},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 311, col: 70, offset: 9686},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 311, col: 73, offset: 9689},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 319, col: 1, offset: 9812},
			expr: &actionExpr{
				pos: position{line: 319, col: 11, offset: 9824},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 319, col: 11, offset: 9824},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 319, col: 11, offset: 9824},
							expr: &ruleRefExpr{
								pos:  position{line: 319, col: 11, offset: 9824},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 319, col: 20, offset: 9833},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 319, col: 27, offset: 9840},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 319, col: 30, offset: 9843},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 319, col: 34, offset: 9847},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 319, col: 44, offset: 9857},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 319, col: 47, offset: 9860},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 326, col: 1, offset: 9951},
			expr: &actionExpr{
				pos: position{line: 326, col: 12, offset: 9964},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 326, col: 12, offset: 9964},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 326, col: 12, offset: 9964},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 326, col: 20, offset: 9972},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 326, col: 23, offset: 9975},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 326, col: 27, offset: 9979},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 326, col: 37, offset: 9989},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 326, col: 40, offset: 9992},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 333, col: 1, offset: 10084},
			expr: &actionExpr{
				pos: position{line: 333, col: 11, offset: 10096},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 333, col: 11, offset: 10096},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 333, col: 11, offset: 10096},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 333, col: 22, offset: 10107},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 333, col: 30, offset: 10115},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 337, col: 1, offset: 10152},
			expr: &choiceExpr{
				pos: position{line: 337, col: 14, offset: 10167},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 337, col: 14, offset: 10167},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 337, col: 24, offset: 10177},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 337, col: 41, offset: 10194},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 337, col: 55, offset: 10208},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 337, col: 66, offset: 10219},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 337, col: 78, offset: 10231},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 339, col: 1, offset: 10243},
			expr: &actionExpr{
				pos: position{line: 339, col: 15, offset: 10259},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 339, col: 15, offset: 10259},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 339, col: 15, offset: 10259},
							expr: &charClassMatcher{
								pos:        position{line: 339, col: 15, offset: 10259},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 339, col: 21, offset: 10265},
							expr: &ruleRefExpr{
								pos:  position{line: 339, col: 21, offset: 10265},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 343, col: 1, offset: 10329},
			expr: &actionExpr{
				pos: position{line: 343, col: 18, offset: 10348},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 343, col: 18, offset: 10348},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 343, col: 18, offset: 10348},
							expr: &charClassMatcher{
								pos:        position{line: 343, col: 18, offset: 10348},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 343, col: 24, offset: 10354},
							expr: &ruleRefExpr{
								pos:  position{line: 343, col: 24, offset: 10354},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 343, col: 31, offset: 10361},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 343, col: 35, offset: 10365},
							expr: &ruleRefExpr{
								pos:  position{line: 343, col: 35, offset: 10365},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 343, col: 42, offset: 10372},
							expr: &seqExpr{
								pos: position{line: 343, col: 44, offset: 10374},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 343, col: 44, offset: 10374},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 343, col: 51, offset: 10381},
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
			pos:  position{line: 347, col: 1, offset: 10451},
			expr: &actionExpr{
				pos: position{line: 347, col: 13, offset: 10465},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 347, col: 13, offset: 10465},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 347, col: 13, offset: 10465},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 17, offset: 10469},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 347, col: 20, offset: 10472},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 347, col: 27, offset: 10479},
								expr: &seqExpr{
									pos: position{line: 347, col: 28, offset: 10480},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 347, col: 28, offset: 10480},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 347, col: 39, offset: 10491},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 347, col: 42, offset: 10494},
											expr: &ruleRefExpr{
												pos:  position{line: 347, col: 42, offset: 10494},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 347, col: 57, offset: 10509},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 62, offset: 10514},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 347, col: 65, offset: 10517},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 356, col: 1, offset: 10711},
			expr: &actionExpr{
				pos: position{line: 356, col: 12, offset: 10724},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 356, col: 12, offset: 10724},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 356, col: 12, offset: 10724},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 16, offset: 10728},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 356, col: 19, offset: 10731},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 356, col: 26, offset: 10738},
								expr: &seqExpr{
									pos: position{line: 356, col: 27, offset: 10739},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 356, col: 27, offset: 10739},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 356, col: 38, offset: 10750},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 356, col: 41, offset: 10753},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 356, col: 45, offset: 10757},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 356, col: 48, offset: 10760},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 356, col: 59, offset: 10771},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 356, col: 63, offset: 10775},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 356, col: 63, offset: 10775},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 356, col: 69, offset: 10781},
													expr: &litMatcher{
														pos:        position{line: 356, col: 70, offset: 10782},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 356, col: 75, offset: 10787},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 356, col: 80, offset: 10792},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 376, col: 1, offset: 11342},
			expr: &ruleRefExpr{
				pos:  position{line: 376, col: 19, offset: 11362},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 378, col: 1, offset: 11369},
			expr: &actionExpr{
				pos: position{line: 378, col: 9, offset: 11379},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 378, col: 9, offset: 11379},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 378, col: 9, offset: 11379},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 378, col: 17, offset: 11387},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 378, col: 20, offset: 11390},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 378, col: 25, offset: 11395},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 378, col: 36, offset: 11406},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 378, col: 39, offset: 11409},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 378, col: 43, offset: 11413},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 378, col: 46, offset: 11416},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 378, col: 53, offset: 11423},
								expr: &ruleRefExpr{
									pos:  position{line: 378, col: 53, offset: 11423},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 378, col: 61, offset: 11431},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 378, col: 64, offset: 11434},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 378, col: 75, offset: 11445},
								expr: &seqExpr{
									pos: position{line: 378, col: 76, offset: 11446},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 378, col: 76, offset: 11446},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 378, col: 86, offset: 11456},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 378, col: 92, offset: 11462},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 378, col: 92, offset: 11462},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 378, col: 98, offset: 11468},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 378, col: 115, offset: 11485},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 395, col: 1, offset: 11983},
			expr: &actionExpr{
				pos: position{line: 395, col: 19, offset: 12003},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 395, col: 19, offset: 12003,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 399, col: 1, offset: 12074},
			expr: &actionExpr{
				pos: position{line: 399, col: 10, offset: 12085},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 399, col: 10, offset: 12085},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 399, col: 10, offset: 12085},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 19, offset: 12094},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 21, offset: 12096},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 399, col: 26, offset: 12101},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 34, offset: 12109},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 403, col: 1, offset: 12163},
			expr: &actionExpr{
				pos: position{line: 403, col: 13, offset: 12177},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 403, col: 13, offset: 12177},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 403, col: 13, offset: 12177},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 403, col: 18, offset: 12182},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 403, col: 29, offset: 12193},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 403, col: 31, offset: 12195},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 403, col: 35, offset: 12199},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 403, col: 38, offset: 12202},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 403, col: 44, offset: 12208},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 403, col: 55, offset: 12219},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 415, col: 1, offset: 12626},
			expr: &actionExpr{
				pos: position{line: 415, col: 11, offset: 12638},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 415, col: 12, offset: 12639},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 415, col: 13, offset: 12640},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 415, col: 13, offset: 12640},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 415, col: 17, offset: 12644},
									expr: &choiceExpr{
										pos: position{line: 415, col: 18, offset: 12645},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 415, col: 18, offset: 12645},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 415, col: 25, offset: 12652},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 415, col: 32, offset: 12659},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 415, col: 40, offset: 12667},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 415, col: 40, offset: 12667},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 415, col: 45, offset: 12672},
									expr: &choiceExpr{
										pos: position{line: 415, col: 46, offset: 12673},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 415, col: 46, offset: 12673},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 415, col: 53, offset: 12680},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 415, col: 60, offset: 12687},
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
			pos:  position{line: 422, col: 1, offset: 12923},
			expr: &actionExpr{
				pos: position{line: 422, col: 14, offset: 12938},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 422, col: 14, offset: 12938},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 422, col: 14, offset: 12938},
							expr: &choiceExpr{
								pos: position{line: 422, col: 15, offset: 12939},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 422, col: 15, offset: 12939},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 422, col: 24, offset: 12948},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 422, col: 30, offset: 12954},
							expr: &choiceExpr{
								pos: position{line: 422, col: 31, offset: 12955},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 422, col: 31, offset: 12955},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 422, col: 40, offset: 12964},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 422, col: 48, offset: 12972},
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
			pos:  position{line: 426, col: 1, offset: 13031},
			expr: &charClassMatcher{
				pos:        position{line: 426, col: 17, offset: 13049},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 427, col: 1, offset: 13054},
			expr: &charClassMatcher{
				pos:        position{line: 427, col: 10, offset: 13065},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 428, col: 1, offset: 13074},
			expr: &charClassMatcher{
				pos:        position{line: 428, col: 9, offset: 13084},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 430, col: 1, offset: 13091},
			expr: &anyMatcher{
				line: 430, col: 14, offset: 13106,
			},
		},
		{
			name: "Comment",
			pos:  position{line: 431, col: 1, offset: 13108},
			expr: &choiceExpr{
				pos: position{line: 431, col: 11, offset: 13120},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 431, col: 11, offset: 13120},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 431, col: 30, offset: 13139},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 432, col: 1, offset: 13157},
			expr: &seqExpr{
				pos: position{line: 432, col: 20, offset: 13178},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 432, col: 20, offset: 13178},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 432, col: 25, offset: 13183},
						expr: &seqExpr{
							pos: position{line: 432, col: 27, offset: 13185},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 432, col: 27, offset: 13185},
									expr: &litMatcher{
										pos:        position{line: 432, col: 28, offset: 13186},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 432, col: 33, offset: 13191},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 432, col: 47, offset: 13205},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 433, col: 1, offset: 13210},
			expr: &seqExpr{
				pos: position{line: 433, col: 36, offset: 13247},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 433, col: 36, offset: 13247},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 433, col: 41, offset: 13252},
						expr: &seqExpr{
							pos: position{line: 433, col: 43, offset: 13254},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 433, col: 43, offset: 13254},
									expr: &choiceExpr{
										pos: position{line: 433, col: 46, offset: 13257},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 433, col: 46, offset: 13257},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 433, col: 53, offset: 13264},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 433, col: 59, offset: 13270},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 433, col: 73, offset: 13284},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 434, col: 1, offset: 13289},
			expr: &choiceExpr{
				pos: position{line: 434, col: 21, offset: 13311},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 434, col: 22, offset: 13312},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 434, col: 22, offset: 13312},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 434, col: 27, offset: 13317},
								expr: &seqExpr{
									pos: position{line: 434, col: 29, offset: 13319},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 434, col: 29, offset: 13319},
											expr: &ruleRefExpr{
												pos:  position{line: 434, col: 30, offset: 13320},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 434, col: 34, offset: 13324},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 434, col: 52, offset: 13342},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 434, col: 52, offset: 13342},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 434, col: 56, offset: 13346},
								expr: &seqExpr{
									pos: position{line: 434, col: 58, offset: 13348},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 434, col: 58, offset: 13348},
											expr: &ruleRefExpr{
												pos:  position{line: 434, col: 59, offset: 13349},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 434, col: 63, offset: 13353},
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
			pos:  position{line: 436, col: 1, offset: 13369},
			expr: &zeroOrMoreExpr{
				pos: position{line: 436, col: 6, offset: 13376},
				expr: &choiceExpr{
					pos: position{line: 436, col: 8, offset: 13378},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 436, col: 8, offset: 13378},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 436, col: 21, offset: 13391},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 436, col: 27, offset: 13397},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 437, col: 1, offset: 13408},
			expr: &zeroOrMoreExpr{
				pos: position{line: 437, col: 5, offset: 13414},
				expr: &choiceExpr{
					pos: position{line: 437, col: 7, offset: 13416},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 437, col: 7, offset: 13416},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 437, col: 20, offset: 13429},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 438, col: 1, offset: 13465},
			expr: &zeroOrMoreExpr{
				pos: position{line: 438, col: 6, offset: 13472},
				expr: &ruleRefExpr{
					pos:  position{line: 438, col: 6, offset: 13472},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 440, col: 1, offset: 13485},
			expr: &charClassMatcher{
				pos:        position{line: 440, col: 14, offset: 13500},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 441, col: 1, offset: 13508},
			expr: &litMatcher{
				pos:        position{line: 441, col: 7, offset: 13516},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 442, col: 1, offset: 13521},
			expr: &choiceExpr{
				pos: position{line: 442, col: 7, offset: 13529},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 442, col: 7, offset: 13529},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 442, col: 7, offset: 13529},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 442, col: 10, offset: 13532},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 442, col: 16, offset: 13538},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 442, col: 16, offset: 13538},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 442, col: 18, offset: 13540},
								expr: &ruleRefExpr{
									pos:  position{line: 442, col: 18, offset: 13540},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 442, col: 37, offset: 13559},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 442, col: 43, offset: 13565},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 442, col: 43, offset: 13565},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 442, col: 46, offset: 13568},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 444, col: 1, offset: 13573},
			expr: &notExpr{
				pos: position{line: 444, col: 7, offset: 13581},
				expr: &anyMatcher{
					line: 444, col: 8, offset: 13582,
				},
			},
		},
	},
}

func (c *current) onGrammar1(statements interface{}) (interface{}, error) {
	thrift := &Thrift{
		Includes:   make(map[string]string),
		Namespaces: make(map[string]string),
		Typedefs:   make(map[string]*Type),
		Constants:  make(map[string]*Constant),
		Enums:      make(map[string]*Enum),
		Structs:    make(map[string]*Struct),
		Exceptions: make(map[string]*Struct),
		Unions:     make(map[string]*Struct),
		Services:   make(map[string]*Service),
	}
	frugal := &Frugal{
		Thrift: thrift,
		Scopes: []*Scope{},
	}

	stmts := toIfaceSlice(statements)
	for _, st := range stmts {
		switch v := st.([]interface{})[0].(type) {
		case *namespace:
			thrift.Namespaces[v.scope] = v.namespace
		case *Constant:
			thrift.Constants[v.Name] = v
		case *Enum:
			thrift.Enums[v.Name] = v
		case *typeDef:
			thrift.Typedefs[v.name] = v.typ
		case *Struct:
			thrift.Structs[v.Name] = v
		case exception:
			thrift.Exceptions[v.Name] = (*Struct)(v)
		case union:
			thrift.Unions[v.Name] = unionToStruct(v)
		case *Service:
			thrift.Services[v.Name] = v
		case include:
			name := string(v)
			if ix := strings.LastIndex(name, "."); ix > 0 {
				name = name[:ix]
			}
			thrift.Includes[name] = string(v)
		case *Scope:
			frugal.Scopes = append(frugal.Scopes, v)
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

func (c *current) onInclude1(file interface{}) (interface{}, error) {
	return include(file.(string)), nil
}

func (p *parser) callonInclude1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInclude1(stack["file"])
}

func (c *current) onNamespace1(scope, ns interface{}) (interface{}, error) {
	return &namespace{
		scope:     ifaceSliceToString(scope),
		namespace: string(ns.(Identifier)),
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

func (c *current) onEnumValue1(name, value interface{}) (interface{}, error) {
	ev := &EnumValue{
		Name:  string(name.(Identifier)),
		Value: -1,
	}
	if value != nil {
		ev.Value = int(value.([]interface{})[2].(int64))
	}
	return ev, nil
}

func (p *parser) callonEnumValue1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnumValue1(stack["name"], stack["value"])
}

func (c *current) onTypeDef1(typ, name interface{}) (interface{}, error) {
	return &typeDef{
		name: string(name.(Identifier)),
		typ:  typ.(*Type),
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

func (c *current) onField1(id, req, typ, name, def interface{}) (interface{}, error) {
	f := &Field{
		ID:   int(id.(int64)),
		Name: string(name.(Identifier)),
		Type: typ.(*Type),
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
	return p.cur.onField1(stack["id"], stack["req"], stack["typ"], stack["name"], stack["def"])
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
		Methods: make(map[string]*Method, len(ms)),
	}
	if extends != nil {
		svc.Extends = string(extends.([]interface{})[2].(Identifier))
	}
	for _, m := range ms {
		mt := m.([]interface{})[0].(*Method)
		svc.Methods[mt.Name] = mt
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

func (c *current) onFunction1(oneway, typ, name, arguments, exceptions interface{}) (interface{}, error) {
	m := &Method{
		Name: string(name.(Identifier)),
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
	return p.cur.onFunction1(stack["oneway"], stack["typ"], stack["name"], stack["arguments"], stack["exceptions"])
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

func (c *current) onDefinitionType1(typ interface{}) (interface{}, error) {
	return typ, nil
}

func (p *parser) callonDefinitionType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDefinitionType1(stack["typ"])
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

func (c *current) onScope1(name, prefix, operations interface{}) (interface{}, error) {
	ops := operations.([]interface{})
	scope := &Scope{
		Name:       string(name.(Identifier)),
		Operations: make([]*Operation, len(ops)),
		Prefix:     defaultPrefix,
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
	return p.cur.onScope1(stack["name"], stack["prefix"], stack["operations"])
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
	return newScopePrefix(name.(string))
}

func (p *parser) callonPrefix1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPrefix1(stack["name"])
}

func (c *current) onOperation1(name, param interface{}) (interface{}, error) {
	o := &Operation{
		Name:  string(name.(Identifier)),
		Param: string(param.(Identifier)),
	}
	return o, nil
}

func (p *parser) callonOperation1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onOperation1(stack["name"], stack["param"])
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
