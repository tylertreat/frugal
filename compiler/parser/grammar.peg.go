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

type namespace struct {
	scope     string
	namespace string
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
			pos:  position{line: 85, col: 1, offset: 2275},
			expr: &actionExpr{
				pos: position{line: 85, col: 11, offset: 2287},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 85, col: 11, offset: 2287},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 85, col: 11, offset: 2287},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 14, offset: 2290},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 85, col: 25, offset: 2301},
								expr: &seqExpr{
									pos: position{line: 85, col: 27, offset: 2303},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 85, col: 27, offset: 2303},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 85, col: 37, offset: 2313},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 85, col: 44, offset: 2320},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 85, col: 44, offset: 2320},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 85, col: 50, offset: 2326},
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
			pos:  position{line: 149, col: 1, offset: 4463},
			expr: &actionExpr{
				pos: position{line: 149, col: 15, offset: 4479},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 149, col: 15, offset: 4479,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 153, col: 1, offset: 4537},
			expr: &actionExpr{
				pos: position{line: 153, col: 13, offset: 4551},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 153, col: 13, offset: 4551},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 153, col: 13, offset: 4551},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 153, col: 20, offset: 4558},
								expr: &seqExpr{
									pos: position{line: 153, col: 21, offset: 4559},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 153, col: 21, offset: 4559},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 153, col: 31, offset: 4569},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 153, col: 36, offset: 4574},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 153, col: 47, offset: 4585},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 153, col: 47, offset: 4585},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 153, col: 65, offset: 4603},
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
			pos:  position{line: 166, col: 1, offset: 5074},
			expr: &choiceExpr{
				pos: position{line: 166, col: 19, offset: 5094},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 166, col: 19, offset: 5094},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 29, offset: 5104},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 41, offset: 5116},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 49, offset: 5124},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 56, offset: 5131},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 66, offset: 5141},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 75, offset: 5150},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 87, offset: 5162},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 95, offset: 5170},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 168, col: 1, offset: 5179},
			expr: &actionExpr{
				pos: position{line: 168, col: 11, offset: 5191},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 168, col: 11, offset: 5191},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 168, col: 11, offset: 5191},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 168, col: 21, offset: 5201},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 168, col: 23, offset: 5203},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 168, col: 28, offset: 5208},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 168, col: 36, offset: 5216},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 172, col: 1, offset: 5264},
			expr: &actionExpr{
				pos: position{line: 172, col: 13, offset: 5278},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 172, col: 13, offset: 5278},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 172, col: 13, offset: 5278},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 25, offset: 5290},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 172, col: 27, offset: 5292},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 172, col: 33, offset: 5298},
								expr: &charClassMatcher{
									pos:        position{line: 172, col: 33, offset: 5298},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 42, offset: 5307},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 172, col: 44, offset: 5309},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 172, col: 47, offset: 5312},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 58, offset: 5323},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 179, col: 1, offset: 5456},
			expr: &actionExpr{
				pos: position{line: 179, col: 9, offset: 5466},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 179, col: 9, offset: 5466},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 179, col: 9, offset: 5466},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 17, offset: 5474},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 179, col: 19, offset: 5476},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 179, col: 23, offset: 5480},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 33, offset: 5490},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 179, col: 35, offset: 5492},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 179, col: 40, offset: 5497},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 51, offset: 5508},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 179, col: 53, offset: 5510},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 57, offset: 5514},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 179, col: 59, offset: 5516},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 179, col: 65, offset: 5522},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 76, offset: 5533},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 187, col: 1, offset: 5665},
			expr: &actionExpr{
				pos: position{line: 187, col: 8, offset: 5674},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 187, col: 8, offset: 5674},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 187, col: 8, offset: 5674},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 187, col: 15, offset: 5681},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 187, col: 17, offset: 5683},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 187, col: 22, offset: 5688},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 187, col: 33, offset: 5699},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 187, col: 36, offset: 5702},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 187, col: 40, offset: 5706},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 187, col: 43, offset: 5709},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 187, col: 50, offset: 5716},
								expr: &seqExpr{
									pos: position{line: 187, col: 51, offset: 5717},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 187, col: 51, offset: 5717},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 187, col: 61, offset: 5727},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 187, col: 66, offset: 5732},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 187, col: 70, offset: 5736},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 210, col: 1, offset: 6348},
			expr: &actionExpr{
				pos: position{line: 210, col: 13, offset: 6362},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 210, col: 13, offset: 6362},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 210, col: 13, offset: 6362},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 210, col: 20, offset: 6369},
								expr: &seqExpr{
									pos: position{line: 210, col: 21, offset: 6370},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 210, col: 21, offset: 6370},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 210, col: 31, offset: 6380},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 210, col: 36, offset: 6385},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 210, col: 41, offset: 6390},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 210, col: 52, offset: 6401},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 210, col: 54, offset: 6403},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 210, col: 60, offset: 6409},
								expr: &seqExpr{
									pos: position{line: 210, col: 61, offset: 6410},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 210, col: 61, offset: 6410},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 210, col: 65, offset: 6414},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 210, col: 67, offset: 6416},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 210, col: 81, offset: 6430},
							expr: &ruleRefExpr{
								pos:  position{line: 210, col: 81, offset: 6430},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 225, col: 1, offset: 6766},
			expr: &actionExpr{
				pos: position{line: 225, col: 11, offset: 6778},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 225, col: 11, offset: 6778},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 225, col: 11, offset: 6778},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 21, offset: 6788},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 225, col: 23, offset: 6790},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 225, col: 27, offset: 6794},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 37, offset: 6804},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 225, col: 39, offset: 6806},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 225, col: 44, offset: 6811},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 55, offset: 6822},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 232, col: 1, offset: 6931},
			expr: &actionExpr{
				pos: position{line: 232, col: 10, offset: 6942},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 232, col: 10, offset: 6942},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 232, col: 10, offset: 6942},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 232, col: 19, offset: 6951},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 232, col: 21, offset: 6953},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 232, col: 24, offset: 6956},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 233, col: 1, offset: 6996},
			expr: &actionExpr{
				pos: position{line: 233, col: 13, offset: 7010},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 233, col: 13, offset: 7010},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 233, col: 13, offset: 7010},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 25, offset: 7022},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 233, col: 27, offset: 7024},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 233, col: 30, offset: 7027},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 234, col: 1, offset: 7078},
			expr: &actionExpr{
				pos: position{line: 234, col: 9, offset: 7088},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 234, col: 9, offset: 7088},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 234, col: 9, offset: 7088},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 234, col: 17, offset: 7096},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 234, col: 19, offset: 7098},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 234, col: 22, offset: 7101},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 235, col: 1, offset: 7148},
			expr: &actionExpr{
				pos: position{line: 235, col: 14, offset: 7163},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 235, col: 14, offset: 7163},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 235, col: 14, offset: 7163},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 19, offset: 7168},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 30, offset: 7179},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 235, col: 33, offset: 7182},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 37, offset: 7186},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 40, offset: 7189},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 47, offset: 7196},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 235, col: 57, offset: 7206},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 61, offset: 7210},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 245, col: 1, offset: 7371},
			expr: &actionExpr{
				pos: position{line: 245, col: 13, offset: 7385},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 245, col: 13, offset: 7385},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 245, col: 20, offset: 7392},
						expr: &seqExpr{
							pos: position{line: 245, col: 21, offset: 7393},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 245, col: 21, offset: 7393},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 245, col: 27, offset: 7399},
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
			pos:  position{line: 254, col: 1, offset: 7580},
			expr: &actionExpr{
				pos: position{line: 254, col: 9, offset: 7590},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 254, col: 9, offset: 7590},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 254, col: 9, offset: 7590},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 254, col: 16, offset: 7597},
								expr: &seqExpr{
									pos: position{line: 254, col: 17, offset: 7598},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 254, col: 17, offset: 7598},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 254, col: 27, offset: 7608},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 254, col: 32, offset: 7613},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 254, col: 35, offset: 7616},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 254, col: 47, offset: 7628},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 254, col: 49, offset: 7630},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 254, col: 53, offset: 7634},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 254, col: 55, offset: 7636},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 254, col: 59, offset: 7640},
								expr: &ruleRefExpr{
									pos:  position{line: 254, col: 59, offset: 7640},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 254, col: 69, offset: 7650},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 254, col: 71, offset: 7652},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 254, col: 75, offset: 7656},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 254, col: 85, offset: 7666},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 254, col: 87, offset: 7668},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 254, col: 92, offset: 7673},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 254, col: 103, offset: 7684},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 254, col: 106, offset: 7687},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 254, col: 110, offset: 7691},
								expr: &seqExpr{
									pos: position{line: 254, col: 111, offset: 7692},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 254, col: 111, offset: 7692},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 254, col: 115, offset: 7696},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 254, col: 117, offset: 7698},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 254, col: 130, offset: 7711},
							expr: &ruleRefExpr{
								pos:  position{line: 254, col: 130, offset: 7711},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 273, col: 1, offset: 8130},
			expr: &actionExpr{
				pos: position{line: 273, col: 12, offset: 8143},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 273, col: 13, offset: 8144},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 273, col: 13, offset: 8144},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 273, col: 26, offset: 8157},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 277, col: 1, offset: 8231},
			expr: &actionExpr{
				pos: position{line: 277, col: 11, offset: 8243},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 277, col: 11, offset: 8243},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 277, col: 11, offset: 8243},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 277, col: 21, offset: 8253},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 277, col: 23, offset: 8255},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 277, col: 28, offset: 8260},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 277, col: 39, offset: 8271},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 277, col: 41, offset: 8273},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 277, col: 49, offset: 8281},
								expr: &seqExpr{
									pos: position{line: 277, col: 50, offset: 8282},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 277, col: 50, offset: 8282},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 277, col: 60, offset: 8292},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 277, col: 63, offset: 8295},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 277, col: 74, offset: 8306},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 277, col: 79, offset: 8311},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 277, col: 82, offset: 8314},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 277, col: 86, offset: 8318},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 277, col: 89, offset: 8321},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 277, col: 97, offset: 8329},
								expr: &seqExpr{
									pos: position{line: 277, col: 98, offset: 8330},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 277, col: 98, offset: 8330},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 277, col: 107, offset: 8339},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 277, col: 113, offset: 8345},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 277, col: 113, offset: 8345},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 277, col: 119, offset: 8351},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 277, col: 138, offset: 8370},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 292, col: 1, offset: 8765},
			expr: &actionExpr{
				pos: position{line: 292, col: 21, offset: 8787},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 292, col: 21, offset: 8787,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 296, col: 1, offset: 8856},
			expr: &actionExpr{
				pos: position{line: 296, col: 12, offset: 8869},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 296, col: 12, offset: 8869},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 296, col: 12, offset: 8869},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 296, col: 19, offset: 8876},
								expr: &seqExpr{
									pos: position{line: 296, col: 20, offset: 8877},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 296, col: 20, offset: 8877},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 296, col: 30, offset: 8887},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 296, col: 35, offset: 8892},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 296, col: 42, offset: 8899},
								expr: &seqExpr{
									pos: position{line: 296, col: 43, offset: 8900},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 296, col: 43, offset: 8900},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 296, col: 52, offset: 8909},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 296, col: 57, offset: 8914},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 296, col: 61, offset: 8918},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 296, col: 74, offset: 8931},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 296, col: 77, offset: 8934},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 296, col: 82, offset: 8939},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 296, col: 93, offset: 8950},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 296, col: 95, offset: 8952},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 296, col: 99, offset: 8956},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 296, col: 102, offset: 8959},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 296, col: 112, offset: 8969},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 296, col: 122, offset: 8979},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 296, col: 126, offset: 8983},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 296, col: 129, offset: 8986},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 296, col: 140, offset: 8997},
								expr: &ruleRefExpr{
									pos:  position{line: 296, col: 140, offset: 8997},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 296, col: 148, offset: 9005},
							expr: &ruleRefExpr{
								pos:  position{line: 296, col: 148, offset: 9005},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 323, col: 1, offset: 9596},
			expr: &actionExpr{
				pos: position{line: 323, col: 16, offset: 9613},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 323, col: 16, offset: 9613},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 323, col: 21, offset: 9618},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 323, col: 21, offset: 9618},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 323, col: 30, offset: 9627},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 330, col: 1, offset: 9749},
			expr: &actionExpr{
				pos: position{line: 330, col: 10, offset: 9760},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 330, col: 10, offset: 9760},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 330, col: 10, offset: 9760},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 330, col: 19, offset: 9769},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 330, col: 22, offset: 9772},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 330, col: 26, offset: 9776},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 330, col: 29, offset: 9779},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 330, col: 40, offset: 9790},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 330, col: 50, offset: 9800},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 334, col: 1, offset: 9836},
			expr: &actionExpr{
				pos: position{line: 334, col: 13, offset: 9850},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 334, col: 13, offset: 9850},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 334, col: 18, offset: 9855},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 334, col: 18, offset: 9855},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 334, col: 29, offset: 9866},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 334, col: 45, offset: 9882},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "DefinitionType",
			pos:  position{line: 341, col: 1, offset: 10007},
			expr: &actionExpr{
				pos: position{line: 341, col: 18, offset: 10026},
				run: (*parser).callonDefinitionType1,
				expr: &labeledExpr{
					pos:   position{line: 341, col: 18, offset: 10026},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 341, col: 23, offset: 10031},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 341, col: 23, offset: 10031},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 341, col: 34, offset: 10042},
								name: "ContainerType",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 345, col: 1, offset: 10082},
			expr: &actionExpr{
				pos: position{line: 345, col: 12, offset: 10095},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 345, col: 13, offset: 10096},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 345, col: 13, offset: 10096},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 22, offset: 10105},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 31, offset: 10114},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 39, offset: 10122},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 47, offset: 10130},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 55, offset: 10138},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 66, offset: 10149},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 77, offset: 10160},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 349, col: 1, offset: 10220},
			expr: &actionExpr{
				pos: position{line: 349, col: 17, offset: 10238},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 349, col: 17, offset: 10238},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 349, col: 22, offset: 10243},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 349, col: 22, offset: 10243},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 32, offset: 10253},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 42, offset: 10263},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 353, col: 1, offset: 10298},
			expr: &actionExpr{
				pos: position{line: 353, col: 11, offset: 10310},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 353, col: 11, offset: 10310},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 353, col: 11, offset: 10310},
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 11, offset: 10310},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 353, col: 20, offset: 10319},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 27, offset: 10326},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 353, col: 30, offset: 10329},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 34, offset: 10333},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 44, offset: 10343},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 353, col: 47, offset: 10346},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 51, offset: 10350},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 353, col: 54, offset: 10353},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 60, offset: 10359},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 70, offset: 10369},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 353, col: 73, offset: 10372},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 361, col: 1, offset: 10495},
			expr: &actionExpr{
				pos: position{line: 361, col: 11, offset: 10507},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 361, col: 11, offset: 10507},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 361, col: 11, offset: 10507},
							expr: &ruleRefExpr{
								pos:  position{line: 361, col: 11, offset: 10507},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 361, col: 20, offset: 10516},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 361, col: 27, offset: 10523},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 361, col: 30, offset: 10526},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 361, col: 34, offset: 10530},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 361, col: 44, offset: 10540},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 361, col: 47, offset: 10543},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 368, col: 1, offset: 10634},
			expr: &actionExpr{
				pos: position{line: 368, col: 12, offset: 10647},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 368, col: 12, offset: 10647},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 368, col: 12, offset: 10647},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 20, offset: 10655},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 368, col: 23, offset: 10658},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 27, offset: 10662},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 37, offset: 10672},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 368, col: 40, offset: 10675},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 375, col: 1, offset: 10767},
			expr: &actionExpr{
				pos: position{line: 375, col: 11, offset: 10779},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 375, col: 11, offset: 10779},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 375, col: 11, offset: 10779},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 375, col: 22, offset: 10790},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 30, offset: 10798},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 379, col: 1, offset: 10835},
			expr: &choiceExpr{
				pos: position{line: 379, col: 14, offset: 10850},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 379, col: 14, offset: 10850},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 24, offset: 10860},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 41, offset: 10877},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 55, offset: 10891},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 66, offset: 10902},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 78, offset: 10914},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 381, col: 1, offset: 10926},
			expr: &actionExpr{
				pos: position{line: 381, col: 15, offset: 10942},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 381, col: 15, offset: 10942},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 381, col: 15, offset: 10942},
							expr: &charClassMatcher{
								pos:        position{line: 381, col: 15, offset: 10942},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 381, col: 21, offset: 10948},
							expr: &ruleRefExpr{
								pos:  position{line: 381, col: 21, offset: 10948},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 385, col: 1, offset: 11012},
			expr: &actionExpr{
				pos: position{line: 385, col: 18, offset: 11031},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 385, col: 18, offset: 11031},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 385, col: 18, offset: 11031},
							expr: &charClassMatcher{
								pos:        position{line: 385, col: 18, offset: 11031},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 385, col: 24, offset: 11037},
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 24, offset: 11037},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 385, col: 31, offset: 11044},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 385, col: 35, offset: 11048},
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 35, offset: 11048},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 385, col: 42, offset: 11055},
							expr: &seqExpr{
								pos: position{line: 385, col: 44, offset: 11057},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 385, col: 44, offset: 11057},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 385, col: 51, offset: 11064},
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
			pos:  position{line: 389, col: 1, offset: 11134},
			expr: &actionExpr{
				pos: position{line: 389, col: 13, offset: 11148},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 389, col: 13, offset: 11148},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 389, col: 13, offset: 11148},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 17, offset: 11152},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 20, offset: 11155},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 389, col: 27, offset: 11162},
								expr: &seqExpr{
									pos: position{line: 389, col: 28, offset: 11163},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 389, col: 28, offset: 11163},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 389, col: 39, offset: 11174},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 389, col: 42, offset: 11177},
											expr: &ruleRefExpr{
												pos:  position{line: 389, col: 42, offset: 11177},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 389, col: 57, offset: 11192},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 62, offset: 11197},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 389, col: 65, offset: 11200},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 398, col: 1, offset: 11394},
			expr: &actionExpr{
				pos: position{line: 398, col: 12, offset: 11407},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 398, col: 12, offset: 11407},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 398, col: 12, offset: 11407},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 16, offset: 11411},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 398, col: 19, offset: 11414},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 398, col: 26, offset: 11421},
								expr: &seqExpr{
									pos: position{line: 398, col: 27, offset: 11422},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 398, col: 27, offset: 11422},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 38, offset: 11433},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 398, col: 41, offset: 11436},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 45, offset: 11440},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 48, offset: 11443},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 59, offset: 11454},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 398, col: 63, offset: 11458},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 398, col: 63, offset: 11458},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 398, col: 69, offset: 11464},
													expr: &litMatcher{
														pos:        position{line: 398, col: 70, offset: 11465},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 75, offset: 11470},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 398, col: 80, offset: 11475},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 418, col: 1, offset: 12025},
			expr: &ruleRefExpr{
				pos:  position{line: 418, col: 19, offset: 12045},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 420, col: 1, offset: 12052},
			expr: &actionExpr{
				pos: position{line: 420, col: 9, offset: 12062},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 420, col: 9, offset: 12062},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 420, col: 9, offset: 12062},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 420, col: 16, offset: 12069},
								expr: &seqExpr{
									pos: position{line: 420, col: 17, offset: 12070},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 420, col: 17, offset: 12070},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 420, col: 27, offset: 12080},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 420, col: 32, offset: 12085},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 40, offset: 12093},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 43, offset: 12096},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 420, col: 48, offset: 12101},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 59, offset: 12112},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 420, col: 62, offset: 12115},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 66, offset: 12119},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 69, offset: 12122},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 420, col: 76, offset: 12129},
								expr: &ruleRefExpr{
									pos:  position{line: 420, col: 76, offset: 12129},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 84, offset: 12137},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 87, offset: 12140},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 420, col: 98, offset: 12151},
								expr: &seqExpr{
									pos: position{line: 420, col: 99, offset: 12152},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 420, col: 99, offset: 12152},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 420, col: 109, offset: 12162},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 420, col: 115, offset: 12168},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 420, col: 115, offset: 12168},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 420, col: 121, offset: 12174},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 138, offset: 12191},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 441, col: 1, offset: 12736},
			expr: &actionExpr{
				pos: position{line: 441, col: 19, offset: 12756},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 441, col: 19, offset: 12756,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 445, col: 1, offset: 12823},
			expr: &actionExpr{
				pos: position{line: 445, col: 10, offset: 12834},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 445, col: 10, offset: 12834},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 445, col: 10, offset: 12834},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 19, offset: 12843},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 445, col: 21, offset: 12845},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 445, col: 26, offset: 12850},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 34, offset: 12858},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 449, col: 1, offset: 12908},
			expr: &actionExpr{
				pos: position{line: 449, col: 13, offset: 12922},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 449, col: 13, offset: 12922},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 449, col: 13, offset: 12922},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 449, col: 20, offset: 12929},
								expr: &seqExpr{
									pos: position{line: 449, col: 21, offset: 12930},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 449, col: 21, offset: 12930},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 449, col: 31, offset: 12940},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 449, col: 36, offset: 12945},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 449, col: 41, offset: 12950},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 52, offset: 12961},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 449, col: 54, offset: 12963},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 58, offset: 12967},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 449, col: 61, offset: 12970},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 449, col: 67, offset: 12976},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 78, offset: 12987},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 465, col: 1, offset: 13489},
			expr: &actionExpr{
				pos: position{line: 465, col: 11, offset: 13501},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 465, col: 12, offset: 13502},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 465, col: 13, offset: 13503},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 465, col: 13, offset: 13503},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 465, col: 17, offset: 13507},
									expr: &choiceExpr{
										pos: position{line: 465, col: 18, offset: 13508},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 465, col: 18, offset: 13508},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 465, col: 25, offset: 13515},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 465, col: 32, offset: 13522},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 465, col: 40, offset: 13530},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 465, col: 40, offset: 13530},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 465, col: 45, offset: 13535},
									expr: &choiceExpr{
										pos: position{line: 465, col: 46, offset: 13536},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 465, col: 46, offset: 13536},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 465, col: 53, offset: 13543},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 465, col: 60, offset: 13550},
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
			pos:  position{line: 472, col: 1, offset: 13766},
			expr: &actionExpr{
				pos: position{line: 472, col: 14, offset: 13781},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 472, col: 14, offset: 13781},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 472, col: 14, offset: 13781},
							expr: &choiceExpr{
								pos: position{line: 472, col: 15, offset: 13782},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 472, col: 15, offset: 13782},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 472, col: 24, offset: 13791},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 472, col: 30, offset: 13797},
							expr: &choiceExpr{
								pos: position{line: 472, col: 31, offset: 13798},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 472, col: 31, offset: 13798},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 472, col: 40, offset: 13807},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 472, col: 48, offset: 13815},
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
			pos:  position{line: 476, col: 1, offset: 13870},
			expr: &charClassMatcher{
				pos:        position{line: 476, col: 17, offset: 13888},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 477, col: 1, offset: 13893},
			expr: &charClassMatcher{
				pos:        position{line: 477, col: 10, offset: 13904},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 478, col: 1, offset: 13913},
			expr: &charClassMatcher{
				pos:        position{line: 478, col: 9, offset: 13923},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 480, col: 1, offset: 13930},
			expr: &anyMatcher{
				line: 480, col: 14, offset: 13945,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 481, col: 1, offset: 13947},
			expr: &actionExpr{
				pos: position{line: 481, col: 13, offset: 13961},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 481, col: 13, offset: 13961},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 481, col: 13, offset: 13961},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 481, col: 20, offset: 13968},
							expr: &seqExpr{
								pos: position{line: 481, col: 22, offset: 13970},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 481, col: 22, offset: 13970},
										expr: &litMatcher{
											pos:        position{line: 481, col: 23, offset: 13971},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 481, col: 28, offset: 13976},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 481, col: 42, offset: 13990},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 487, col: 1, offset: 14170},
			expr: &choiceExpr{
				pos: position{line: 487, col: 11, offset: 14182},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 487, col: 11, offset: 14182},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 487, col: 30, offset: 14201},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 488, col: 1, offset: 14219},
			expr: &seqExpr{
				pos: position{line: 488, col: 20, offset: 14240},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 488, col: 20, offset: 14240},
						expr: &ruleRefExpr{
							pos:  position{line: 488, col: 21, offset: 14241},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 488, col: 31, offset: 14251},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 488, col: 36, offset: 14256},
						expr: &seqExpr{
							pos: position{line: 488, col: 38, offset: 14258},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 488, col: 38, offset: 14258},
									expr: &litMatcher{
										pos:        position{line: 488, col: 39, offset: 14259},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 488, col: 44, offset: 14264},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 488, col: 58, offset: 14278},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 489, col: 1, offset: 14283},
			expr: &seqExpr{
				pos: position{line: 489, col: 36, offset: 14320},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 489, col: 36, offset: 14320},
						expr: &ruleRefExpr{
							pos:  position{line: 489, col: 37, offset: 14321},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 489, col: 47, offset: 14331},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 489, col: 52, offset: 14336},
						expr: &seqExpr{
							pos: position{line: 489, col: 54, offset: 14338},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 489, col: 54, offset: 14338},
									expr: &choiceExpr{
										pos: position{line: 489, col: 57, offset: 14341},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 489, col: 57, offset: 14341},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 489, col: 64, offset: 14348},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 489, col: 70, offset: 14354},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 489, col: 84, offset: 14368},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 490, col: 1, offset: 14373},
			expr: &choiceExpr{
				pos: position{line: 490, col: 21, offset: 14395},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 490, col: 22, offset: 14396},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 490, col: 22, offset: 14396},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 490, col: 27, offset: 14401},
								expr: &seqExpr{
									pos: position{line: 490, col: 29, offset: 14403},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 490, col: 29, offset: 14403},
											expr: &ruleRefExpr{
												pos:  position{line: 490, col: 30, offset: 14404},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 490, col: 34, offset: 14408},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 490, col: 52, offset: 14426},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 490, col: 52, offset: 14426},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 490, col: 56, offset: 14430},
								expr: &seqExpr{
									pos: position{line: 490, col: 58, offset: 14432},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 490, col: 58, offset: 14432},
											expr: &ruleRefExpr{
												pos:  position{line: 490, col: 59, offset: 14433},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 490, col: 63, offset: 14437},
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
			pos:  position{line: 492, col: 1, offset: 14453},
			expr: &zeroOrMoreExpr{
				pos: position{line: 492, col: 6, offset: 14460},
				expr: &choiceExpr{
					pos: position{line: 492, col: 8, offset: 14462},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 492, col: 8, offset: 14462},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 492, col: 21, offset: 14475},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 492, col: 27, offset: 14481},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 493, col: 1, offset: 14492},
			expr: &zeroOrMoreExpr{
				pos: position{line: 493, col: 5, offset: 14498},
				expr: &choiceExpr{
					pos: position{line: 493, col: 7, offset: 14500},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 493, col: 7, offset: 14500},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 493, col: 20, offset: 14513},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 494, col: 1, offset: 14549},
			expr: &zeroOrMoreExpr{
				pos: position{line: 494, col: 6, offset: 14556},
				expr: &ruleRefExpr{
					pos:  position{line: 494, col: 6, offset: 14556},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 496, col: 1, offset: 14569},
			expr: &charClassMatcher{
				pos:        position{line: 496, col: 14, offset: 14584},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 497, col: 1, offset: 14592},
			expr: &litMatcher{
				pos:        position{line: 497, col: 7, offset: 14600},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 498, col: 1, offset: 14605},
			expr: &choiceExpr{
				pos: position{line: 498, col: 7, offset: 14613},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 498, col: 7, offset: 14613},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 498, col: 7, offset: 14613},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 498, col: 10, offset: 14616},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 498, col: 16, offset: 14622},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 498, col: 16, offset: 14622},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 498, col: 18, offset: 14624},
								expr: &ruleRefExpr{
									pos:  position{line: 498, col: 18, offset: 14624},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 498, col: 37, offset: 14643},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 498, col: 43, offset: 14649},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 498, col: 43, offset: 14649},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 498, col: 46, offset: 14652},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 500, col: 1, offset: 14657},
			expr: &notExpr{
				pos: position{line: 500, col: 7, offset: 14665},
				expr: &anyMatcher{
					line: 500, col: 8, offset: 14666,
				},
			},
		},
	},
}

func (c *current) onGrammar1(statements interface{}) (interface{}, error) {
	thrift := &Thrift{
		Includes:   make(map[string]string),
		Namespaces: make(map[string]string),
		Typedefs:   make(map[string]*TypeDef),
		Constants:  make(map[string]*Constant),
		Enums:      make(map[string]*Enum),
		Structs:    make(map[string]*Struct),
		Exceptions: make(map[string]*Struct),
		Unions:     make(map[string]*Struct),
		Services:   make(map[string]*Service),
	}
	frugal := &Frugal{
		Thrift:         thrift,
		Scopes:         []*Scope{},
		ParsedIncludes: make(map[string]*Frugal),
	}

	stmts := toIfaceSlice(statements)
	for _, st := range stmts {
		wrapper := st.([]interface{})[0].(*statementWrapper)
		switch v := wrapper.statement.(type) {
		case *namespace:
			thrift.Namespaces[v.scope] = v.namespace
		case *Constant:
			v.Comment = wrapper.comment
			thrift.Constants[v.Name] = v
		case *Enum:
			v.Comment = wrapper.comment
			thrift.Enums[v.Name] = v
		case *TypeDef:
			v.Comment = wrapper.comment
			thrift.Typedefs[v.Name] = v
		case *Struct:
			v.Comment = wrapper.comment
			thrift.Structs[v.Name] = v
		case exception:
			strct := (*Struct)(v)
			strct.Comment = wrapper.comment
			thrift.Exceptions[v.Name] = strct
		case union:
			strct := unionToStruct(v)
			strct.Comment = wrapper.comment
			thrift.Unions[v.Name] = strct
		case *Service:
			v.Comment = wrapper.comment
			thrift.Services[v.Name] = v
		case include:
			name := string(v)
			if ix := strings.LastIndex(name, "."); ix > 0 {
				name = name[:ix]
			}
			thrift.Includes[name] = string(v)
		case *Scope:
			v.Comment = wrapper.comment
			v.Frugal = frugal
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
	return newScopePrefix(name.(string))
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
