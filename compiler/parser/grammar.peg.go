package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
		f.Modifier = Optional
	}
	return st
}

// toAnnotations converts an interface{} to an Annotation slice.
func toAnnotations(v interface{}) Annotations {
	if v == nil {
		return nil
	}
	return Annotations(v.([]*Annotation))
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 85, col: 1, offset: 2393},
			expr: &actionExpr{
				pos: position{line: 85, col: 12, offset: 2404},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 85, col: 12, offset: 2404},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 85, col: 12, offset: 2404},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 15, offset: 2407},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 85, col: 26, offset: 2418},
								expr: &seqExpr{
									pos: position{line: 85, col: 28, offset: 2420},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 85, col: 28, offset: 2420},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 85, col: 38, offset: 2430},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 85, col: 45, offset: 2437},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 85, col: 45, offset: 2437},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 85, col: 51, offset: 2443},
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
			pos:  position{line: 153, col: 1, offset: 4872},
			expr: &actionExpr{
				pos: position{line: 153, col: 16, offset: 4887},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 153, col: 16, offset: 4887,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 157, col: 1, offset: 4945},
			expr: &actionExpr{
				pos: position{line: 157, col: 14, offset: 4958},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 157, col: 14, offset: 4958},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 157, col: 14, offset: 4958},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 157, col: 21, offset: 4965},
								expr: &seqExpr{
									pos: position{line: 157, col: 22, offset: 4966},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 157, col: 22, offset: 4966},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 157, col: 32, offset: 4976},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 157, col: 37, offset: 4981},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 157, col: 48, offset: 4992},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 157, col: 48, offset: 4992},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 157, col: 66, offset: 5010},
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
			pos:  position{line: 170, col: 1, offset: 5481},
			expr: &choiceExpr{
				pos: position{line: 170, col: 20, offset: 5500},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 170, col: 20, offset: 5500},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 30, offset: 5510},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 42, offset: 5522},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 50, offset: 5530},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 57, offset: 5537},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 67, offset: 5547},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 76, offset: 5556},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 88, offset: 5568},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 170, col: 96, offset: 5576},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 172, col: 1, offset: 5585},
			expr: &actionExpr{
				pos: position{line: 172, col: 12, offset: 5596},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 172, col: 12, offset: 5596},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 172, col: 12, offset: 5596},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 22, offset: 5606},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 172, col: 24, offset: 5608},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 172, col: 29, offset: 5613},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 37, offset: 5621},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 172, col: 39, offset: 5623},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 172, col: 51, offset: 5635},
								expr: &ruleRefExpr{
									pos:  position{line: 172, col: 51, offset: 5635},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 68, offset: 5652},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 184, col: 1, offset: 5929},
			expr: &actionExpr{
				pos: position{line: 184, col: 14, offset: 5942},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 184, col: 14, offset: 5942},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 184, col: 14, offset: 5942},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 184, col: 26, offset: 5954},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 184, col: 28, offset: 5956},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 184, col: 34, offset: 5962},
								expr: &charClassMatcher{
									pos:        position{line: 184, col: 34, offset: 5962},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 184, col: 44, offset: 5972},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 184, col: 46, offset: 5974},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 184, col: 49, offset: 5977},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 184, col: 60, offset: 5988},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 184, col: 62, offset: 5990},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 184, col: 74, offset: 6002},
								expr: &ruleRefExpr{
									pos:  position{line: 184, col: 74, offset: 6002},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 184, col: 91, offset: 6019},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 192, col: 1, offset: 6205},
			expr: &actionExpr{
				pos: position{line: 192, col: 10, offset: 6214},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 192, col: 10, offset: 6214},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 192, col: 10, offset: 6214},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 18, offset: 6222},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 192, col: 20, offset: 6224},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 192, col: 24, offset: 6228},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 34, offset: 6238},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 192, col: 36, offset: 6240},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 192, col: 41, offset: 6245},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 52, offset: 6256},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 192, col: 54, offset: 6258},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 58, offset: 6262},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 192, col: 60, offset: 6264},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 192, col: 66, offset: 6270},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 77, offset: 6281},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 192, col: 79, offset: 6283},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 192, col: 91, offset: 6295},
								expr: &ruleRefExpr{
									pos:  position{line: 192, col: 91, offset: 6295},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 108, offset: 6312},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 201, col: 1, offset: 6506},
			expr: &actionExpr{
				pos: position{line: 201, col: 9, offset: 6514},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 201, col: 9, offset: 6514},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 201, col: 9, offset: 6514},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 16, offset: 6521},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 201, col: 18, offset: 6523},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 201, col: 23, offset: 6528},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 34, offset: 6539},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 201, col: 37, offset: 6542},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 41, offset: 6546},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 201, col: 44, offset: 6549},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 201, col: 51, offset: 6556},
								expr: &seqExpr{
									pos: position{line: 201, col: 52, offset: 6557},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 201, col: 52, offset: 6557},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 201, col: 62, offset: 6567},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 201, col: 67, offset: 6572},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 71, offset: 6576},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 201, col: 73, offset: 6578},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 201, col: 85, offset: 6590},
								expr: &ruleRefExpr{
									pos:  position{line: 201, col: 85, offset: 6590},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 102, offset: 6607},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 225, col: 1, offset: 7269},
			expr: &actionExpr{
				pos: position{line: 225, col: 14, offset: 7282},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 225, col: 14, offset: 7282},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 225, col: 14, offset: 7282},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 225, col: 21, offset: 7289},
								expr: &seqExpr{
									pos: position{line: 225, col: 22, offset: 7290},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 225, col: 22, offset: 7290},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 225, col: 32, offset: 7300},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 225, col: 37, offset: 7305},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 225, col: 42, offset: 7310},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 53, offset: 7321},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 225, col: 55, offset: 7323},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 225, col: 61, offset: 7329},
								expr: &seqExpr{
									pos: position{line: 225, col: 62, offset: 7330},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 225, col: 62, offset: 7330},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 225, col: 66, offset: 7334},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 225, col: 68, offset: 7336},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 82, offset: 7350},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 225, col: 84, offset: 7352},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 225, col: 96, offset: 7364},
								expr: &ruleRefExpr{
									pos:  position{line: 225, col: 96, offset: 7364},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 225, col: 113, offset: 7381},
							expr: &ruleRefExpr{
								pos:  position{line: 225, col: 113, offset: 7381},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 241, col: 1, offset: 7779},
			expr: &actionExpr{
				pos: position{line: 241, col: 12, offset: 7790},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 241, col: 12, offset: 7790},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 241, col: 12, offset: 7790},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 241, col: 22, offset: 7800},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 241, col: 24, offset: 7802},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 241, col: 28, offset: 7806},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 241, col: 38, offset: 7816},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 241, col: 40, offset: 7818},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 241, col: 45, offset: 7823},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 241, col: 56, offset: 7834},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 241, col: 58, offset: 7836},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 241, col: 70, offset: 7848},
								expr: &ruleRefExpr{
									pos:  position{line: 241, col: 70, offset: 7848},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 241, col: 87, offset: 7865},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 249, col: 1, offset: 8037},
			expr: &actionExpr{
				pos: position{line: 249, col: 11, offset: 8047},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 249, col: 11, offset: 8047},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 249, col: 11, offset: 8047},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 20, offset: 8056},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 22, offset: 8058},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 25, offset: 8061},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 250, col: 1, offset: 8101},
			expr: &actionExpr{
				pos: position{line: 250, col: 14, offset: 8114},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 250, col: 14, offset: 8114},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 250, col: 14, offset: 8114},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 26, offset: 8126},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 28, offset: 8128},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 31, offset: 8131},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 251, col: 1, offset: 8182},
			expr: &actionExpr{
				pos: position{line: 251, col: 10, offset: 8191},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 251, col: 10, offset: 8191},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 251, col: 10, offset: 8191},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 251, col: 18, offset: 8199},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 251, col: 20, offset: 8201},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 251, col: 23, offset: 8204},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 252, col: 1, offset: 8251},
			expr: &actionExpr{
				pos: position{line: 252, col: 15, offset: 8265},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 252, col: 15, offset: 8265},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 252, col: 15, offset: 8265},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 20, offset: 8270},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 31, offset: 8281},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 252, col: 34, offset: 8284},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 38, offset: 8288},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 41, offset: 8291},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 48, offset: 8298},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 252, col: 58, offset: 8308},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 62, offset: 8312},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 64, offset: 8314},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 76, offset: 8326},
								expr: &ruleRefExpr{
									pos:  position{line: 252, col: 76, offset: 8326},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 93, offset: 8343},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 263, col: 1, offset: 8560},
			expr: &actionExpr{
				pos: position{line: 263, col: 14, offset: 8573},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 263, col: 14, offset: 8573},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 263, col: 21, offset: 8580},
						expr: &seqExpr{
							pos: position{line: 263, col: 22, offset: 8581},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 263, col: 22, offset: 8581},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 263, col: 28, offset: 8587},
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
			pos:  position{line: 272, col: 1, offset: 8768},
			expr: &actionExpr{
				pos: position{line: 272, col: 10, offset: 8777},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 272, col: 10, offset: 8777},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 272, col: 10, offset: 8777},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 272, col: 17, offset: 8784},
								expr: &seqExpr{
									pos: position{line: 272, col: 18, offset: 8785},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 272, col: 18, offset: 8785},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 272, col: 28, offset: 8795},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 272, col: 33, offset: 8800},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 272, col: 36, offset: 8803},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 48, offset: 8815},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 272, col: 50, offset: 8817},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 54, offset: 8821},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 56, offset: 8823},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 272, col: 60, offset: 8827},
								expr: &ruleRefExpr{
									pos:  position{line: 272, col: 60, offset: 8827},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 75, offset: 8842},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 77, offset: 8844},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 272, col: 81, offset: 8848},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 91, offset: 8858},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 93, offset: 8860},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 272, col: 98, offset: 8865},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 109, offset: 8876},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 112, offset: 8879},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 272, col: 116, offset: 8883},
								expr: &seqExpr{
									pos: position{line: 272, col: 117, offset: 8884},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 272, col: 117, offset: 8884},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 272, col: 121, offset: 8888},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 272, col: 123, offset: 8890},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 136, offset: 8903},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 138, offset: 8905},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 272, col: 150, offset: 8917},
								expr: &ruleRefExpr{
									pos:  position{line: 272, col: 150, offset: 8917},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 272, col: 167, offset: 8934},
							expr: &ruleRefExpr{
								pos:  position{line: 272, col: 167, offset: 8934},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 295, col: 1, offset: 9466},
			expr: &actionExpr{
				pos: position{line: 295, col: 18, offset: 9483},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 295, col: 19, offset: 9484},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 295, col: 19, offset: 9484},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 295, col: 32, offset: 9497},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 303, col: 1, offset: 9640},
			expr: &actionExpr{
				pos: position{line: 303, col: 12, offset: 9651},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 303, col: 12, offset: 9651},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 303, col: 12, offset: 9651},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 22, offset: 9661},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 24, offset: 9663},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 303, col: 29, offset: 9668},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 40, offset: 9679},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 42, offset: 9681},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 303, col: 50, offset: 9689},
								expr: &seqExpr{
									pos: position{line: 303, col: 51, offset: 9690},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 303, col: 51, offset: 9690},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 61, offset: 9700},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 64, offset: 9703},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 75, offset: 9714},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 80, offset: 9719},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 303, col: 83, offset: 9722},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 87, offset: 9726},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 90, offset: 9729},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 303, col: 98, offset: 9737},
								expr: &seqExpr{
									pos: position{line: 303, col: 99, offset: 9738},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 303, col: 99, offset: 9738},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 108, offset: 9747},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 303, col: 114, offset: 9753},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 303, col: 114, offset: 9753},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 303, col: 120, offset: 9759},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 139, offset: 9778},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 141, offset: 9780},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 303, col: 153, offset: 9792},
								expr: &ruleRefExpr{
									pos:  position{line: 303, col: 153, offset: 9792},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 170, offset: 9809},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 320, col: 1, offset: 10250},
			expr: &actionExpr{
				pos: position{line: 320, col: 22, offset: 10271},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 320, col: 22, offset: 10271,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 324, col: 1, offset: 10340},
			expr: &actionExpr{
				pos: position{line: 324, col: 13, offset: 10352},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 324, col: 13, offset: 10352},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 324, col: 13, offset: 10352},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 324, col: 20, offset: 10359},
								expr: &seqExpr{
									pos: position{line: 324, col: 21, offset: 10360},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 324, col: 21, offset: 10360},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 324, col: 31, offset: 10370},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 324, col: 36, offset: 10375},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 324, col: 43, offset: 10382},
								expr: &seqExpr{
									pos: position{line: 324, col: 44, offset: 10383},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 324, col: 44, offset: 10383},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 324, col: 53, offset: 10392},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 324, col: 58, offset: 10397},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 62, offset: 10401},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 75, offset: 10414},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 78, offset: 10417},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 83, offset: 10422},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 94, offset: 10433},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 324, col: 96, offset: 10435},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 100, offset: 10439},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 103, offset: 10442},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 113, offset: 10452},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 324, col: 123, offset: 10462},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 127, offset: 10466},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 130, offset: 10469},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 324, col: 141, offset: 10480},
								expr: &ruleRefExpr{
									pos:  position{line: 324, col: 141, offset: 10480},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 149, offset: 10488},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 151, offset: 10490},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 324, col: 163, offset: 10502},
								expr: &ruleRefExpr{
									pos:  position{line: 324, col: 163, offset: 10502},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 324, col: 180, offset: 10519},
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 180, offset: 10519},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 352, col: 1, offset: 11170},
			expr: &actionExpr{
				pos: position{line: 352, col: 17, offset: 11186},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 352, col: 17, offset: 11186},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 352, col: 22, offset: 11191},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 352, col: 22, offset: 11191},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 352, col: 31, offset: 11200},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 359, col: 1, offset: 11322},
			expr: &actionExpr{
				pos: position{line: 359, col: 11, offset: 11332},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 359, col: 11, offset: 11332},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 359, col: 11, offset: 11332},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 359, col: 20, offset: 11341},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 359, col: 23, offset: 11344},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 359, col: 27, offset: 11348},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 359, col: 30, offset: 11351},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 359, col: 41, offset: 11362},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 359, col: 51, offset: 11372},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 363, col: 1, offset: 11408},
			expr: &actionExpr{
				pos: position{line: 363, col: 14, offset: 11421},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 363, col: 14, offset: 11421},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 363, col: 19, offset: 11426},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 363, col: 19, offset: 11426},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 363, col: 30, offset: 11437},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 363, col: 46, offset: 11453},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 370, col: 1, offset: 11578},
			expr: &actionExpr{
				pos: position{line: 370, col: 13, offset: 11590},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 370, col: 13, offset: 11590},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 370, col: 13, offset: 11590},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 370, col: 18, offset: 11595},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 31, offset: 11608},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 370, col: 33, offset: 11610},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 370, col: 45, offset: 11622},
								expr: &ruleRefExpr{
									pos:  position{line: 370, col: 45, offset: 11622},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "BaseTypeName",
			pos:  position{line: 377, col: 1, offset: 11758},
			expr: &actionExpr{
				pos: position{line: 377, col: 17, offset: 11774},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 377, col: 18, offset: 11775},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 377, col: 18, offset: 11775},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 27, offset: 11784},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 36, offset: 11793},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 44, offset: 11801},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 52, offset: 11809},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 60, offset: 11817},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 71, offset: 11828},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 82, offset: 11839},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 381, col: 1, offset: 11886},
			expr: &actionExpr{
				pos: position{line: 381, col: 18, offset: 11903},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 381, col: 18, offset: 11903},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 381, col: 23, offset: 11908},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 381, col: 23, offset: 11908},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 381, col: 33, offset: 11918},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 381, col: 43, offset: 11928},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 385, col: 1, offset: 11963},
			expr: &actionExpr{
				pos: position{line: 385, col: 12, offset: 11974},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 385, col: 12, offset: 11974},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 385, col: 12, offset: 11974},
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 12, offset: 11974},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 385, col: 21, offset: 11983},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 28, offset: 11990},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 385, col: 31, offset: 11993},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 35, offset: 11997},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 45, offset: 12007},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 385, col: 48, offset: 12010},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 52, offset: 12014},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 385, col: 55, offset: 12017},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 61, offset: 12023},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 71, offset: 12033},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 385, col: 74, offset: 12036},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 78, offset: 12040},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 385, col: 80, offset: 12042},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 385, col: 92, offset: 12054},
								expr: &ruleRefExpr{
									pos:  position{line: 385, col: 92, offset: 12054},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 394, col: 1, offset: 12252},
			expr: &actionExpr{
				pos: position{line: 394, col: 12, offset: 12263},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 394, col: 12, offset: 12263},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 394, col: 12, offset: 12263},
							expr: &ruleRefExpr{
								pos:  position{line: 394, col: 12, offset: 12263},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 394, col: 21, offset: 12272},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 394, col: 28, offset: 12279},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 394, col: 31, offset: 12282},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 394, col: 35, offset: 12286},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 394, col: 45, offset: 12296},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 394, col: 48, offset: 12299},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 394, col: 52, offset: 12303},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 394, col: 54, offset: 12305},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 394, col: 66, offset: 12317},
								expr: &ruleRefExpr{
									pos:  position{line: 394, col: 66, offset: 12317},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 402, col: 1, offset: 12479},
			expr: &actionExpr{
				pos: position{line: 402, col: 13, offset: 12491},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 402, col: 13, offset: 12491},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 402, col: 13, offset: 12491},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 402, col: 21, offset: 12499},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 402, col: 24, offset: 12502},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 402, col: 28, offset: 12506},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 402, col: 38, offset: 12516},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 402, col: 41, offset: 12519},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 402, col: 45, offset: 12523},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 402, col: 47, offset: 12525},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 402, col: 59, offset: 12537},
								expr: &ruleRefExpr{
									pos:  position{line: 402, col: 59, offset: 12537},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 410, col: 1, offset: 12700},
			expr: &actionExpr{
				pos: position{line: 410, col: 12, offset: 12711},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 410, col: 12, offset: 12711},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 410, col: 12, offset: 12711},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 410, col: 23, offset: 12722},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 410, col: 31, offset: 12730},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 414, col: 1, offset: 12767},
			expr: &choiceExpr{
				pos: position{line: 414, col: 15, offset: 12781},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 414, col: 15, offset: 12781},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 25, offset: 12791},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 40, offset: 12806},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 57, offset: 12823},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 71, offset: 12837},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 82, offset: 12848},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 94, offset: 12860},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 416, col: 1, offset: 12872},
			expr: &actionExpr{
				pos: position{line: 416, col: 20, offset: 12891},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 416, col: 20, offset: 12891},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 416, col: 20, offset: 12891},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 416, col: 24, offset: 12895},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 416, col: 27, offset: 12898},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 416, col: 39, offset: 12910},
								expr: &ruleRefExpr{
									pos:  position{line: 416, col: 39, offset: 12910},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 416, col: 55, offset: 12926},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 424, col: 1, offset: 13090},
			expr: &actionExpr{
				pos: position{line: 424, col: 19, offset: 13108},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 424, col: 19, offset: 13108},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 424, col: 19, offset: 13108},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 424, col: 24, offset: 13113},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 424, col: 35, offset: 13124},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 424, col: 37, offset: 13126},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 424, col: 43, offset: 13132},
								expr: &actionExpr{
									pos: position{line: 424, col: 44, offset: 13133},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 424, col: 44, offset: 13133},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 424, col: 44, offset: 13133},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 424, col: 48, offset: 13137},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 424, col: 51, offset: 13140},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 424, col: 57, offset: 13146},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 424, col: 89, offset: 13178},
							expr: &ruleRefExpr{
								pos:  position{line: 424, col: 89, offset: 13178},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 424, col: 104, offset: 13193},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 435, col: 1, offset: 13389},
			expr: &actionExpr{
				pos: position{line: 435, col: 17, offset: 13405},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 435, col: 18, offset: 13406},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 435, col: 18, offset: 13406},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 435, col: 27, offset: 13415},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 439, col: 1, offset: 13470},
			expr: &actionExpr{
				pos: position{line: 439, col: 16, offset: 13485},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 439, col: 16, offset: 13485},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 439, col: 16, offset: 13485},
							expr: &charClassMatcher{
								pos:        position{line: 439, col: 16, offset: 13485},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 439, col: 22, offset: 13491},
							expr: &ruleRefExpr{
								pos:  position{line: 439, col: 22, offset: 13491},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 443, col: 1, offset: 13555},
			expr: &actionExpr{
				pos: position{line: 443, col: 19, offset: 13573},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 443, col: 19, offset: 13573},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 443, col: 19, offset: 13573},
							expr: &charClassMatcher{
								pos:        position{line: 443, col: 19, offset: 13573},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 443, col: 25, offset: 13579},
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 25, offset: 13579},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 443, col: 32, offset: 13586},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 443, col: 36, offset: 13590},
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 36, offset: 13590},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 443, col: 43, offset: 13597},
							expr: &seqExpr{
								pos: position{line: 443, col: 45, offset: 13599},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 443, col: 45, offset: 13599},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 443, col: 52, offset: 13606},
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
			pos:  position{line: 447, col: 1, offset: 13676},
			expr: &actionExpr{
				pos: position{line: 447, col: 14, offset: 13689},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 447, col: 14, offset: 13689},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 447, col: 14, offset: 13689},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 18, offset: 13693},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 447, col: 21, offset: 13696},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 447, col: 28, offset: 13703},
								expr: &seqExpr{
									pos: position{line: 447, col: 29, offset: 13704},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 447, col: 29, offset: 13704},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 447, col: 40, offset: 13715},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 447, col: 43, offset: 13718},
											expr: &ruleRefExpr{
												pos:  position{line: 447, col: 43, offset: 13718},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 447, col: 58, offset: 13733},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 63, offset: 13738},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 447, col: 66, offset: 13741},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 456, col: 1, offset: 13935},
			expr: &actionExpr{
				pos: position{line: 456, col: 13, offset: 13947},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 456, col: 13, offset: 13947},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 456, col: 13, offset: 13947},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 17, offset: 13951},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 456, col: 20, offset: 13954},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 456, col: 27, offset: 13961},
								expr: &seqExpr{
									pos: position{line: 456, col: 28, offset: 13962},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 456, col: 28, offset: 13962},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 39, offset: 13973},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 456, col: 42, offset: 13976},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 46, offset: 13980},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 49, offset: 13983},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 60, offset: 13994},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 456, col: 64, offset: 13998},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 456, col: 64, offset: 13998},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 456, col: 70, offset: 14004},
													expr: &litMatcher{
														pos:        position{line: 456, col: 71, offset: 14005},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 76, offset: 14010},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 456, col: 81, offset: 14015},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 476, col: 1, offset: 14565},
			expr: &ruleRefExpr{
				pos:  position{line: 476, col: 20, offset: 14584},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 478, col: 1, offset: 14591},
			expr: &actionExpr{
				pos: position{line: 478, col: 10, offset: 14600},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 478, col: 10, offset: 14600},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 478, col: 10, offset: 14600},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 478, col: 17, offset: 14607},
								expr: &seqExpr{
									pos: position{line: 478, col: 18, offset: 14608},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 478, col: 18, offset: 14608},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 478, col: 28, offset: 14618},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 478, col: 33, offset: 14623},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 41, offset: 14631},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 478, col: 44, offset: 14634},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 478, col: 49, offset: 14639},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 60, offset: 14650},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 478, col: 63, offset: 14653},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 478, col: 70, offset: 14660},
								expr: &ruleRefExpr{
									pos:  position{line: 478, col: 70, offset: 14660},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 78, offset: 14668},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 478, col: 81, offset: 14671},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 85, offset: 14675},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 478, col: 88, offset: 14678},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 478, col: 99, offset: 14689},
								expr: &seqExpr{
									pos: position{line: 478, col: 100, offset: 14690},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 478, col: 100, offset: 14690},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 478, col: 110, offset: 14700},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 478, col: 116, offset: 14706},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 478, col: 116, offset: 14706},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 478, col: 122, offset: 14712},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 139, offset: 14729},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 478, col: 141, offset: 14731},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 478, col: 153, offset: 14743},
								expr: &ruleRefExpr{
									pos:  position{line: 478, col: 153, offset: 14743},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 170, offset: 14760},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 500, col: 1, offset: 15357},
			expr: &actionExpr{
				pos: position{line: 500, col: 20, offset: 15376},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 500, col: 20, offset: 15376,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 504, col: 1, offset: 15443},
			expr: &actionExpr{
				pos: position{line: 504, col: 11, offset: 15453},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 504, col: 11, offset: 15453},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 504, col: 11, offset: 15453},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 504, col: 20, offset: 15462},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 504, col: 23, offset: 15465},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 504, col: 35, offset: 15477},
							expr: &seqExpr{
								pos: position{line: 504, col: 36, offset: 15478},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 504, col: 36, offset: 15478},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 504, col: 40, offset: 15482},
										name: "PrefixToken",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "PrefixToken",
			pos:  position{line: 509, col: 1, offset: 15613},
			expr: &choiceExpr{
				pos: position{line: 509, col: 16, offset: 15628},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 509, col: 17, offset: 15629},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 509, col: 17, offset: 15629},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 509, col: 21, offset: 15633},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 509, col: 32, offset: 15644},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 509, col: 39, offset: 15651},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 511, col: 1, offset: 15663},
			expr: &oneOrMoreExpr{
				pos: position{line: 511, col: 15, offset: 15677},
				expr: &charClassMatcher{
					pos:        position{line: 511, col: 15, offset: 15677},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 513, col: 1, offset: 15695},
			expr: &actionExpr{
				pos: position{line: 513, col: 14, offset: 15708},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 513, col: 14, offset: 15708},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 513, col: 14, offset: 15708},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 513, col: 21, offset: 15715},
								expr: &seqExpr{
									pos: position{line: 513, col: 22, offset: 15716},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 513, col: 22, offset: 15716},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 513, col: 32, offset: 15726},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 513, col: 37, offset: 15731},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 513, col: 42, offset: 15736},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 513, col: 53, offset: 15747},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 513, col: 55, offset: 15749},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 513, col: 59, offset: 15753},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 513, col: 62, offset: 15756},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 513, col: 66, offset: 15760},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 513, col: 76, offset: 15770},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 513, col: 78, offset: 15772},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 513, col: 90, offset: 15784},
								expr: &ruleRefExpr{
									pos:  position{line: 513, col: 90, offset: 15784},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 513, col: 107, offset: 15801},
							expr: &ruleRefExpr{
								pos:  position{line: 513, col: 107, offset: 15801},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 530, col: 1, offset: 16361},
			expr: &actionExpr{
				pos: position{line: 530, col: 12, offset: 16372},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 530, col: 13, offset: 16373},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 530, col: 14, offset: 16374},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 530, col: 14, offset: 16374},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 530, col: 18, offset: 16378},
									expr: &choiceExpr{
										pos: position{line: 530, col: 19, offset: 16379},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 530, col: 19, offset: 16379},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 530, col: 26, offset: 16386},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 530, col: 33, offset: 16393},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 530, col: 41, offset: 16401},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 530, col: 41, offset: 16401},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 530, col: 46, offset: 16406},
									expr: &choiceExpr{
										pos: position{line: 530, col: 47, offset: 16407},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 530, col: 47, offset: 16407},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 530, col: 54, offset: 16414},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 530, col: 61, offset: 16421},
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
			pos:  position{line: 539, col: 1, offset: 16707},
			expr: &actionExpr{
				pos: position{line: 539, col: 15, offset: 16721},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 539, col: 15, offset: 16721},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 539, col: 15, offset: 16721},
							expr: &choiceExpr{
								pos: position{line: 539, col: 16, offset: 16722},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 539, col: 16, offset: 16722},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 539, col: 25, offset: 16731},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 539, col: 31, offset: 16737},
							expr: &choiceExpr{
								pos: position{line: 539, col: 32, offset: 16738},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 539, col: 32, offset: 16738},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 539, col: 41, offset: 16747},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 539, col: 49, offset: 16755},
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
			pos:  position{line: 543, col: 1, offset: 16810},
			expr: &charClassMatcher{
				pos:        position{line: 543, col: 18, offset: 16827},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 544, col: 1, offset: 16832},
			expr: &charClassMatcher{
				pos:        position{line: 544, col: 11, offset: 16842},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 545, col: 1, offset: 16851},
			expr: &charClassMatcher{
				pos:        position{line: 545, col: 10, offset: 16860},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 547, col: 1, offset: 16867},
			expr: &anyMatcher{
				line: 547, col: 15, offset: 16881,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 548, col: 1, offset: 16883},
			expr: &actionExpr{
				pos: position{line: 548, col: 14, offset: 16896},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 548, col: 14, offset: 16896},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 548, col: 14, offset: 16896},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 548, col: 21, offset: 16903},
							expr: &seqExpr{
								pos: position{line: 548, col: 23, offset: 16905},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 548, col: 23, offset: 16905},
										expr: &litMatcher{
											pos:        position{line: 548, col: 24, offset: 16906},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 548, col: 29, offset: 16911},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 548, col: 43, offset: 16925},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 554, col: 1, offset: 17105},
			expr: &choiceExpr{
				pos: position{line: 554, col: 12, offset: 17116},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 554, col: 12, offset: 17116},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 554, col: 31, offset: 17135},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 555, col: 1, offset: 17153},
			expr: &seqExpr{
				pos: position{line: 555, col: 21, offset: 17173},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 555, col: 21, offset: 17173},
						expr: &ruleRefExpr{
							pos:  position{line: 555, col: 22, offset: 17174},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 555, col: 32, offset: 17184},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 555, col: 37, offset: 17189},
						expr: &seqExpr{
							pos: position{line: 555, col: 39, offset: 17191},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 555, col: 39, offset: 17191},
									expr: &litMatcher{
										pos:        position{line: 555, col: 40, offset: 17192},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 555, col: 45, offset: 17197},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 555, col: 59, offset: 17211},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 556, col: 1, offset: 17216},
			expr: &seqExpr{
				pos: position{line: 556, col: 37, offset: 17252},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 556, col: 37, offset: 17252},
						expr: &ruleRefExpr{
							pos:  position{line: 556, col: 38, offset: 17253},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 556, col: 48, offset: 17263},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 556, col: 53, offset: 17268},
						expr: &seqExpr{
							pos: position{line: 556, col: 55, offset: 17270},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 556, col: 55, offset: 17270},
									expr: &choiceExpr{
										pos: position{line: 556, col: 58, offset: 17273},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 556, col: 58, offset: 17273},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 556, col: 65, offset: 17280},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 556, col: 71, offset: 17286},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 556, col: 85, offset: 17300},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 557, col: 1, offset: 17305},
			expr: &choiceExpr{
				pos: position{line: 557, col: 22, offset: 17326},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 557, col: 23, offset: 17327},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 557, col: 23, offset: 17327},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 557, col: 28, offset: 17332},
								expr: &seqExpr{
									pos: position{line: 557, col: 30, offset: 17334},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 557, col: 30, offset: 17334},
											expr: &ruleRefExpr{
												pos:  position{line: 557, col: 31, offset: 17335},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 557, col: 35, offset: 17339},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 557, col: 53, offset: 17357},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 557, col: 53, offset: 17357},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 557, col: 57, offset: 17361},
								expr: &seqExpr{
									pos: position{line: 557, col: 59, offset: 17363},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 557, col: 59, offset: 17363},
											expr: &ruleRefExpr{
												pos:  position{line: 557, col: 60, offset: 17364},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 557, col: 64, offset: 17368},
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
			pos:  position{line: 559, col: 1, offset: 17384},
			expr: &zeroOrMoreExpr{
				pos: position{line: 559, col: 7, offset: 17390},
				expr: &choiceExpr{
					pos: position{line: 559, col: 9, offset: 17392},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 559, col: 9, offset: 17392},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 559, col: 22, offset: 17405},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 559, col: 28, offset: 17411},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 560, col: 1, offset: 17422},
			expr: &zeroOrMoreExpr{
				pos: position{line: 560, col: 6, offset: 17427},
				expr: &choiceExpr{
					pos: position{line: 560, col: 8, offset: 17429},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 560, col: 8, offset: 17429},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 560, col: 21, offset: 17442},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 561, col: 1, offset: 17478},
			expr: &zeroOrMoreExpr{
				pos: position{line: 561, col: 7, offset: 17484},
				expr: &ruleRefExpr{
					pos:  position{line: 561, col: 7, offset: 17484},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 563, col: 1, offset: 17497},
			expr: &charClassMatcher{
				pos:        position{line: 563, col: 15, offset: 17511},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 564, col: 1, offset: 17519},
			expr: &litMatcher{
				pos:        position{line: 564, col: 8, offset: 17526},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 565, col: 1, offset: 17531},
			expr: &choiceExpr{
				pos: position{line: 565, col: 8, offset: 17538},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 565, col: 8, offset: 17538},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 565, col: 8, offset: 17538},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 565, col: 11, offset: 17541},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 565, col: 17, offset: 17547},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 565, col: 17, offset: 17547},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 565, col: 19, offset: 17549},
								expr: &ruleRefExpr{
									pos:  position{line: 565, col: 19, offset: 17549},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 565, col: 38, offset: 17568},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 565, col: 44, offset: 17574},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 565, col: 44, offset: 17574},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 565, col: 47, offset: 17577},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 567, col: 1, offset: 17582},
			expr: &notExpr{
				pos: position{line: 567, col: 8, offset: 17589},
				expr: &anyMatcher{
					line: 567, col: 9, offset: 17590,
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
	stmts := toIfaceSlice(statements)
	frugal := &Frugal{
		Thrift:         thrift,
		Scopes:         []*Scope{},
		ParsedIncludes: make(map[string]*Frugal),
	}

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
			v.Type = StructTypeStruct
			v.Comment = wrapper.comment
			thrift.Structs = append(thrift.Structs, v)
		case exception:
			strct := (*Struct)(v)
			strct.Type = StructTypeException
			strct.Comment = wrapper.comment
			thrift.Exceptions = append(thrift.Exceptions, strct)
		case union:
			strct := unionToStruct(v)
			strct.Type = StructTypeUnion
			strct.Comment = wrapper.comment
			thrift.Unions = append(thrift.Unions, strct)
		case *Service:
			v.Comment = wrapper.comment
			v.Thrift = frugal.Thrift
			thrift.Services = append(thrift.Services, v)
		case *Include:
			thrift.Includes = append(thrift.Includes, v)
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

func (c *current) onInclude1(file, annotations interface{}) (interface{}, error) {
	name := filepath.Base(file.(string))
	if ix := strings.LastIndex(name, "."); ix > 0 {
		name = name[:ix]
	}
	return &Include{
		Name:        name,
		Value:       file.(string),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonInclude1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInclude1(stack["file"], stack["annotations"])
}

func (c *current) onNamespace1(scope, ns, annotations interface{}) (interface{}, error) {
	return &Namespace{
		Scope:       ifaceSliceToString(scope),
		Value:       string(ns.(Identifier)),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonNamespace1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onNamespace1(stack["scope"], stack["ns"], stack["annotations"])
}

func (c *current) onConst1(typ, name, value, annotations interface{}) (interface{}, error) {
	return &Constant{
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Value:       value,
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonConst1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConst1(stack["typ"], stack["name"], stack["value"], stack["annotations"])
}

func (c *current) onEnum1(name, values, annotations interface{}) (interface{}, error) {
	vs := toIfaceSlice(values)
	en := &Enum{
		Name:        string(name.(Identifier)),
		Values:      make([]*EnumValue, len(vs)),
		Annotations: toAnnotations(annotations),
	}
	// Assigns numbers in order. This will behave badly if some values are
	// defined and other are not, but I think that's ok since that's a silly
	// thing to do.
	next := 0
	for idx, v := range vs {
		ev := v.([]interface{})[0].(*EnumValue)
		if ev.Value < 0 {
			ev.Value = next
		}
		if ev.Value >= next {
			next = ev.Value + 1
		}
		en.Values[idx] = ev
	}
	return en, nil
}

func (p *parser) callonEnum1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEnum1(stack["name"], stack["values"], stack["annotations"])
}

func (c *current) onEnumValue1(docstr, name, value, annotations interface{}) (interface{}, error) {
	ev := &EnumValue{
		Name:        string(name.(Identifier)),
		Value:       -1,
		Annotations: toAnnotations(annotations),
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
	return p.cur.onEnumValue1(stack["docstr"], stack["name"], stack["value"], stack["annotations"])
}

func (c *current) onTypeDef1(typ, name, annotations interface{}) (interface{}, error) {
	return &TypeDef{
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonTypeDef1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeDef1(stack["typ"], stack["name"], stack["annotations"])
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

func (c *current) onStructLike1(name, fields, annotations interface{}) (interface{}, error) {
	st := &Struct{
		Name:        string(name.(Identifier)),
		Annotations: toAnnotations(annotations),
	}
	if fields != nil {
		st.Fields = fields.([]*Field)
	}
	return st, nil
}

func (p *parser) callonStructLike1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStructLike1(stack["name"], stack["fields"], stack["annotations"])
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

func (c *current) onField1(docstr, id, mod, typ, name, def, annotations interface{}) (interface{}, error) {
	f := &Field{
		ID:          int(id.(int64)),
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Annotations: toAnnotations(annotations),
	}
	if docstr != nil {
		raw := docstr.([]interface{})[0].(string)
		f.Comment = rawCommentToDocStr(raw)
	}
	if mod != nil {
		f.Modifier = mod.(FieldModifier)
	} else {
		f.Modifier = Default
	}

	if def != nil {
		f.Default = def.([]interface{})[2]
	}
	return f, nil
}

func (p *parser) callonField1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onField1(stack["docstr"], stack["id"], stack["mod"], stack["typ"], stack["name"], stack["def"], stack["annotations"])
}

func (c *current) onFieldModifier1() (interface{}, error) {
	if bytes.Equal(c.text, []byte("required")) {
		return Required, nil
	} else {
		return Optional, nil
	}
}

func (p *parser) callonFieldModifier1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFieldModifier1()
}

func (c *current) onService1(name, extends, methods, annotations interface{}) (interface{}, error) {
	ms := methods.([]interface{})
	svc := &Service{
		Name:        string(name.(Identifier)),
		Methods:     make([]*Method, len(ms)),
		Annotations: toAnnotations(annotations),
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
	return p.cur.onService1(stack["name"], stack["extends"], stack["methods"], stack["annotations"])
}

func (c *current) onEndOfServiceError1() (interface{}, error) {
	return nil, errors.New("parser: expected end of service")
}

func (p *parser) callonEndOfServiceError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEndOfServiceError1()
}

func (c *current) onFunction1(docstr, oneway, typ, name, arguments, exceptions, annotations interface{}) (interface{}, error) {
	m := &Method{
		Name:        string(name.(Identifier)),
		Annotations: toAnnotations(annotations),
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
			e.Modifier = Optional
		}
	}
	return m, nil
}

func (p *parser) callonFunction1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFunction1(stack["docstr"], stack["oneway"], stack["typ"], stack["name"], stack["arguments"], stack["exceptions"], stack["annotations"])
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

func (c *current) onBaseType1(name, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        name.(string),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonBaseType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBaseType1(stack["name"], stack["annotations"])
}

func (c *current) onBaseTypeName1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonBaseTypeName1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBaseTypeName1()
}

func (c *current) onContainerType1(typ interface{}) (interface{}, error) {
	return typ, nil
}

func (p *parser) callonContainerType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onContainerType1(stack["typ"])
}

func (c *current) onMapType1(key, value, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "map",
		KeyType:     key.(*Type),
		ValueType:   value.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonMapType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMapType1(stack["key"], stack["value"], stack["annotations"])
}

func (c *current) onSetType1(typ, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "set",
		ValueType:   typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonSetType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSetType1(stack["typ"], stack["annotations"])
}

func (c *current) onListType1(typ, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "list",
		ValueType:   typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonListType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onListType1(stack["typ"], stack["annotations"])
}

func (c *current) onCppType1(cppType interface{}) (interface{}, error) {
	return cppType, nil
}

func (p *parser) callonCppType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCppType1(stack["cppType"])
}

func (c *current) onTypeAnnotations1(annotations interface{}) (interface{}, error) {
	var anns []*Annotation
	for _, ann := range annotations.([]interface{}) {
		anns = append(anns, ann.(*Annotation))
	}
	return anns, nil
}

func (p *parser) callonTypeAnnotations1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotations1(stack["annotations"])
}

func (c *current) onTypeAnnotation8(value interface{}) (interface{}, error) {
	return value, nil
}

func (p *parser) callonTypeAnnotation8() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotation8(stack["value"])
}

func (c *current) onTypeAnnotation1(name, value interface{}) (interface{}, error) {
	var optValue string
	if value != nil {
		optValue = value.(string)
	}
	return &Annotation{
		Name:  string(name.(Identifier)),
		Value: optValue,
	}, nil
}

func (p *parser) callonTypeAnnotation1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotation1(stack["name"], stack["value"])
}

func (c *current) onBoolConstant1() (interface{}, error) {
	return string(c.text) == "true", nil
}

func (p *parser) callonBoolConstant1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBoolConstant1()
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

func (c *current) onScope1(docstr, name, prefix, operations, annotations interface{}) (interface{}, error) {
	ops := operations.([]interface{})
	scope := &Scope{
		Name:        string(name.(Identifier)),
		Operations:  make([]*Operation, len(ops)),
		Prefix:      defaultPrefix,
		Annotations: toAnnotations(annotations),
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
	return p.cur.onScope1(stack["docstr"], stack["name"], stack["prefix"], stack["operations"], stack["annotations"])
}

func (c *current) onEndOfScopeError1() (interface{}, error) {
	return nil, errors.New("parser: expected end of scope")
}

func (p *parser) callonEndOfScopeError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEndOfScopeError1()
}

func (c *current) onPrefix1() (interface{}, error) {
	prefix := strings.TrimSpace(strings.TrimPrefix(string(c.text), "prefix"))
	return newScopePrefix(prefix)
}

func (p *parser) callonPrefix1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPrefix1()
}

func (c *current) onOperation1(docstr, name, typ, annotations interface{}) (interface{}, error) {
	o := &Operation{
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Annotations: toAnnotations(annotations),
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
	return p.cur.onOperation1(stack["docstr"], stack["name"], stack["typ"], stack["annotations"])
}

func (c *current) onLiteral1() (interface{}, error) {
	if len(c.text) != 0 && c.text[0] == '\'' {
		intermediate := strings.Replace(string(c.text[1:len(c.text)-1]), `\'`, `'`, -1)
		return strconv.Unquote(`"` + strings.Replace(intermediate, `"`, `\"`, -1) + `"`)
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
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String()}
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
