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
			pos:  position{line: 184, col: 1, offset: 5914},
			expr: &actionExpr{
				pos: position{line: 184, col: 14, offset: 5927},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 184, col: 14, offset: 5927},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 184, col: 14, offset: 5927},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 184, col: 26, offset: 5939},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 184, col: 28, offset: 5941},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 184, col: 34, offset: 5947},
								expr: &charClassMatcher{
									pos:        position{line: 184, col: 34, offset: 5947},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 184, col: 44, offset: 5957},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 184, col: 46, offset: 5959},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 184, col: 49, offset: 5962},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 184, col: 60, offset: 5973},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 184, col: 62, offset: 5975},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 184, col: 74, offset: 5987},
								expr: &ruleRefExpr{
									pos:  position{line: 184, col: 74, offset: 5987},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 184, col: 91, offset: 6004},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 192, col: 1, offset: 6190},
			expr: &actionExpr{
				pos: position{line: 192, col: 10, offset: 6199},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 192, col: 10, offset: 6199},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 192, col: 10, offset: 6199},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 18, offset: 6207},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 192, col: 20, offset: 6209},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 192, col: 24, offset: 6213},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 34, offset: 6223},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 192, col: 36, offset: 6225},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 192, col: 41, offset: 6230},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 52, offset: 6241},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 192, col: 54, offset: 6243},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 58, offset: 6247},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 192, col: 60, offset: 6249},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 192, col: 66, offset: 6255},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 77, offset: 6266},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 192, col: 79, offset: 6268},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 192, col: 91, offset: 6280},
								expr: &ruleRefExpr{
									pos:  position{line: 192, col: 91, offset: 6280},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 192, col: 108, offset: 6297},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 201, col: 1, offset: 6491},
			expr: &actionExpr{
				pos: position{line: 201, col: 9, offset: 6499},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 201, col: 9, offset: 6499},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 201, col: 9, offset: 6499},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 16, offset: 6506},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 201, col: 18, offset: 6508},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 201, col: 23, offset: 6513},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 34, offset: 6524},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 201, col: 37, offset: 6527},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 41, offset: 6531},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 201, col: 44, offset: 6534},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 201, col: 51, offset: 6541},
								expr: &seqExpr{
									pos: position{line: 201, col: 52, offset: 6542},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 201, col: 52, offset: 6542},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 201, col: 62, offset: 6552},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 201, col: 67, offset: 6557},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 71, offset: 6561},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 201, col: 73, offset: 6563},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 201, col: 85, offset: 6575},
								expr: &ruleRefExpr{
									pos:  position{line: 201, col: 85, offset: 6575},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 102, offset: 6592},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 225, col: 1, offset: 7254},
			expr: &actionExpr{
				pos: position{line: 225, col: 14, offset: 7267},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 225, col: 14, offset: 7267},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 225, col: 14, offset: 7267},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 225, col: 21, offset: 7274},
								expr: &seqExpr{
									pos: position{line: 225, col: 22, offset: 7275},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 225, col: 22, offset: 7275},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 225, col: 32, offset: 7285},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 225, col: 37, offset: 7290},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 225, col: 42, offset: 7295},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 53, offset: 7306},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 225, col: 55, offset: 7308},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 225, col: 61, offset: 7314},
								expr: &seqExpr{
									pos: position{line: 225, col: 62, offset: 7315},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 225, col: 62, offset: 7315},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 225, col: 66, offset: 7319},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 225, col: 68, offset: 7321},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 82, offset: 7335},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 225, col: 84, offset: 7337},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 225, col: 96, offset: 7349},
								expr: &ruleRefExpr{
									pos:  position{line: 225, col: 96, offset: 7349},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 225, col: 113, offset: 7366},
							expr: &ruleRefExpr{
								pos:  position{line: 225, col: 113, offset: 7366},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 241, col: 1, offset: 7764},
			expr: &actionExpr{
				pos: position{line: 241, col: 12, offset: 7775},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 241, col: 12, offset: 7775},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 241, col: 12, offset: 7775},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 241, col: 22, offset: 7785},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 241, col: 24, offset: 7787},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 241, col: 28, offset: 7791},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 241, col: 38, offset: 7801},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 241, col: 40, offset: 7803},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 241, col: 45, offset: 7808},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 241, col: 56, offset: 7819},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 241, col: 58, offset: 7821},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 241, col: 70, offset: 7833},
								expr: &ruleRefExpr{
									pos:  position{line: 241, col: 70, offset: 7833},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 241, col: 87, offset: 7850},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 249, col: 1, offset: 8022},
			expr: &actionExpr{
				pos: position{line: 249, col: 11, offset: 8032},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 249, col: 11, offset: 8032},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 249, col: 11, offset: 8032},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 20, offset: 8041},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 22, offset: 8043},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 25, offset: 8046},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 250, col: 1, offset: 8086},
			expr: &actionExpr{
				pos: position{line: 250, col: 14, offset: 8099},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 250, col: 14, offset: 8099},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 250, col: 14, offset: 8099},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 26, offset: 8111},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 28, offset: 8113},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 31, offset: 8116},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 251, col: 1, offset: 8167},
			expr: &actionExpr{
				pos: position{line: 251, col: 10, offset: 8176},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 251, col: 10, offset: 8176},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 251, col: 10, offset: 8176},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 251, col: 18, offset: 8184},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 251, col: 20, offset: 8186},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 251, col: 23, offset: 8189},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 252, col: 1, offset: 8236},
			expr: &actionExpr{
				pos: position{line: 252, col: 15, offset: 8250},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 252, col: 15, offset: 8250},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 252, col: 15, offset: 8250},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 20, offset: 8255},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 31, offset: 8266},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 252, col: 34, offset: 8269},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 38, offset: 8273},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 41, offset: 8276},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 48, offset: 8283},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 252, col: 58, offset: 8293},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 62, offset: 8297},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 64, offset: 8299},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 76, offset: 8311},
								expr: &ruleRefExpr{
									pos:  position{line: 252, col: 76, offset: 8311},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 93, offset: 8328},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 263, col: 1, offset: 8545},
			expr: &actionExpr{
				pos: position{line: 263, col: 14, offset: 8558},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 263, col: 14, offset: 8558},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 263, col: 21, offset: 8565},
						expr: &seqExpr{
							pos: position{line: 263, col: 22, offset: 8566},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 263, col: 22, offset: 8566},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 263, col: 28, offset: 8572},
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
			pos:  position{line: 272, col: 1, offset: 8753},
			expr: &actionExpr{
				pos: position{line: 272, col: 10, offset: 8762},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 272, col: 10, offset: 8762},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 272, col: 10, offset: 8762},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 272, col: 17, offset: 8769},
								expr: &seqExpr{
									pos: position{line: 272, col: 18, offset: 8770},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 272, col: 18, offset: 8770},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 272, col: 28, offset: 8780},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 272, col: 33, offset: 8785},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 272, col: 36, offset: 8788},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 48, offset: 8800},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 272, col: 50, offset: 8802},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 54, offset: 8806},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 56, offset: 8808},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 272, col: 60, offset: 8812},
								expr: &ruleRefExpr{
									pos:  position{line: 272, col: 60, offset: 8812},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 75, offset: 8827},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 77, offset: 8829},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 272, col: 81, offset: 8833},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 91, offset: 8843},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 93, offset: 8845},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 272, col: 98, offset: 8850},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 109, offset: 8861},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 112, offset: 8864},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 272, col: 116, offset: 8868},
								expr: &seqExpr{
									pos: position{line: 272, col: 117, offset: 8869},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 272, col: 117, offset: 8869},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 272, col: 121, offset: 8873},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 272, col: 123, offset: 8875},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 272, col: 136, offset: 8888},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 272, col: 138, offset: 8890},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 272, col: 150, offset: 8902},
								expr: &ruleRefExpr{
									pos:  position{line: 272, col: 150, offset: 8902},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 272, col: 167, offset: 8919},
							expr: &ruleRefExpr{
								pos:  position{line: 272, col: 167, offset: 8919},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 295, col: 1, offset: 9451},
			expr: &actionExpr{
				pos: position{line: 295, col: 18, offset: 9468},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 295, col: 19, offset: 9469},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 295, col: 19, offset: 9469},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 295, col: 32, offset: 9482},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 303, col: 1, offset: 9625},
			expr: &actionExpr{
				pos: position{line: 303, col: 12, offset: 9636},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 303, col: 12, offset: 9636},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 303, col: 12, offset: 9636},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 22, offset: 9646},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 24, offset: 9648},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 303, col: 29, offset: 9653},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 40, offset: 9664},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 42, offset: 9666},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 303, col: 50, offset: 9674},
								expr: &seqExpr{
									pos: position{line: 303, col: 51, offset: 9675},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 303, col: 51, offset: 9675},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 61, offset: 9685},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 64, offset: 9688},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 75, offset: 9699},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 80, offset: 9704},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 303, col: 83, offset: 9707},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 87, offset: 9711},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 90, offset: 9714},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 303, col: 98, offset: 9722},
								expr: &seqExpr{
									pos: position{line: 303, col: 99, offset: 9723},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 303, col: 99, offset: 9723},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 108, offset: 9732},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 303, col: 114, offset: 9738},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 303, col: 114, offset: 9738},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 303, col: 120, offset: 9744},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 139, offset: 9763},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 141, offset: 9765},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 303, col: 153, offset: 9777},
								expr: &ruleRefExpr{
									pos:  position{line: 303, col: 153, offset: 9777},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 170, offset: 9794},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 320, col: 1, offset: 10235},
			expr: &actionExpr{
				pos: position{line: 320, col: 22, offset: 10256},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 320, col: 22, offset: 10256,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 324, col: 1, offset: 10325},
			expr: &actionExpr{
				pos: position{line: 324, col: 13, offset: 10337},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 324, col: 13, offset: 10337},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 324, col: 13, offset: 10337},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 324, col: 20, offset: 10344},
								expr: &seqExpr{
									pos: position{line: 324, col: 21, offset: 10345},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 324, col: 21, offset: 10345},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 324, col: 31, offset: 10355},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 324, col: 36, offset: 10360},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 324, col: 43, offset: 10367},
								expr: &seqExpr{
									pos: position{line: 324, col: 44, offset: 10368},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 324, col: 44, offset: 10368},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 324, col: 53, offset: 10377},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 324, col: 58, offset: 10382},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 62, offset: 10386},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 75, offset: 10399},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 78, offset: 10402},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 83, offset: 10407},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 94, offset: 10418},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 324, col: 96, offset: 10420},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 100, offset: 10424},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 103, offset: 10427},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 113, offset: 10437},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 324, col: 123, offset: 10447},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 127, offset: 10451},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 130, offset: 10454},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 324, col: 141, offset: 10465},
								expr: &ruleRefExpr{
									pos:  position{line: 324, col: 141, offset: 10465},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 149, offset: 10473},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 151, offset: 10475},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 324, col: 163, offset: 10487},
								expr: &ruleRefExpr{
									pos:  position{line: 324, col: 163, offset: 10487},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 324, col: 180, offset: 10504},
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 180, offset: 10504},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 352, col: 1, offset: 11155},
			expr: &actionExpr{
				pos: position{line: 352, col: 17, offset: 11171},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 352, col: 17, offset: 11171},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 352, col: 22, offset: 11176},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 352, col: 22, offset: 11176},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 352, col: 31, offset: 11185},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 359, col: 1, offset: 11307},
			expr: &actionExpr{
				pos: position{line: 359, col: 11, offset: 11317},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 359, col: 11, offset: 11317},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 359, col: 11, offset: 11317},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 359, col: 20, offset: 11326},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 359, col: 23, offset: 11329},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 359, col: 27, offset: 11333},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 359, col: 30, offset: 11336},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 359, col: 41, offset: 11347},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 359, col: 51, offset: 11357},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 363, col: 1, offset: 11393},
			expr: &actionExpr{
				pos: position{line: 363, col: 14, offset: 11406},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 363, col: 14, offset: 11406},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 363, col: 19, offset: 11411},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 363, col: 19, offset: 11411},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 363, col: 30, offset: 11422},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 363, col: 46, offset: 11438},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 370, col: 1, offset: 11563},
			expr: &actionExpr{
				pos: position{line: 370, col: 13, offset: 11575},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 370, col: 13, offset: 11575},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 370, col: 13, offset: 11575},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 370, col: 18, offset: 11580},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 31, offset: 11593},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 370, col: 33, offset: 11595},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 370, col: 45, offset: 11607},
								expr: &ruleRefExpr{
									pos:  position{line: 370, col: 45, offset: 11607},
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
			pos:  position{line: 377, col: 1, offset: 11743},
			expr: &actionExpr{
				pos: position{line: 377, col: 17, offset: 11759},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 377, col: 18, offset: 11760},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 377, col: 18, offset: 11760},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 27, offset: 11769},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 36, offset: 11778},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 44, offset: 11786},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 52, offset: 11794},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 60, offset: 11802},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 71, offset: 11813},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 377, col: 82, offset: 11824},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 381, col: 1, offset: 11871},
			expr: &actionExpr{
				pos: position{line: 381, col: 18, offset: 11888},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 381, col: 18, offset: 11888},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 381, col: 23, offset: 11893},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 381, col: 23, offset: 11893},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 381, col: 33, offset: 11903},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 381, col: 43, offset: 11913},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 385, col: 1, offset: 11948},
			expr: &actionExpr{
				pos: position{line: 385, col: 12, offset: 11959},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 385, col: 12, offset: 11959},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 385, col: 12, offset: 11959},
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 12, offset: 11959},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 385, col: 21, offset: 11968},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 28, offset: 11975},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 385, col: 31, offset: 11978},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 35, offset: 11982},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 45, offset: 11992},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 385, col: 48, offset: 11995},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 52, offset: 11999},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 385, col: 55, offset: 12002},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 61, offset: 12008},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 71, offset: 12018},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 385, col: 74, offset: 12021},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 78, offset: 12025},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 385, col: 80, offset: 12027},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 385, col: 92, offset: 12039},
								expr: &ruleRefExpr{
									pos:  position{line: 385, col: 92, offset: 12039},
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
			pos:  position{line: 394, col: 1, offset: 12237},
			expr: &actionExpr{
				pos: position{line: 394, col: 12, offset: 12248},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 394, col: 12, offset: 12248},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 394, col: 12, offset: 12248},
							expr: &ruleRefExpr{
								pos:  position{line: 394, col: 12, offset: 12248},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 394, col: 21, offset: 12257},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 394, col: 28, offset: 12264},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 394, col: 31, offset: 12267},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 394, col: 35, offset: 12271},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 394, col: 45, offset: 12281},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 394, col: 48, offset: 12284},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 394, col: 52, offset: 12288},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 394, col: 54, offset: 12290},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 394, col: 66, offset: 12302},
								expr: &ruleRefExpr{
									pos:  position{line: 394, col: 66, offset: 12302},
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
			pos:  position{line: 402, col: 1, offset: 12464},
			expr: &actionExpr{
				pos: position{line: 402, col: 13, offset: 12476},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 402, col: 13, offset: 12476},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 402, col: 13, offset: 12476},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 402, col: 21, offset: 12484},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 402, col: 24, offset: 12487},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 402, col: 28, offset: 12491},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 402, col: 38, offset: 12501},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 402, col: 41, offset: 12504},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 402, col: 45, offset: 12508},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 402, col: 47, offset: 12510},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 402, col: 59, offset: 12522},
								expr: &ruleRefExpr{
									pos:  position{line: 402, col: 59, offset: 12522},
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
			pos:  position{line: 410, col: 1, offset: 12685},
			expr: &actionExpr{
				pos: position{line: 410, col: 12, offset: 12696},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 410, col: 12, offset: 12696},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 410, col: 12, offset: 12696},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 410, col: 23, offset: 12707},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 410, col: 31, offset: 12715},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 414, col: 1, offset: 12752},
			expr: &choiceExpr{
				pos: position{line: 414, col: 15, offset: 12766},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 414, col: 15, offset: 12766},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 25, offset: 12776},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 40, offset: 12791},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 57, offset: 12808},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 71, offset: 12822},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 82, offset: 12833},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 414, col: 94, offset: 12845},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 416, col: 1, offset: 12857},
			expr: &actionExpr{
				pos: position{line: 416, col: 20, offset: 12876},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 416, col: 20, offset: 12876},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 416, col: 20, offset: 12876},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 416, col: 24, offset: 12880},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 416, col: 27, offset: 12883},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 416, col: 39, offset: 12895},
								expr: &ruleRefExpr{
									pos:  position{line: 416, col: 39, offset: 12895},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 416, col: 55, offset: 12911},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 424, col: 1, offset: 13075},
			expr: &actionExpr{
				pos: position{line: 424, col: 19, offset: 13093},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 424, col: 19, offset: 13093},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 424, col: 19, offset: 13093},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 424, col: 24, offset: 13098},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 424, col: 35, offset: 13109},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 424, col: 37, offset: 13111},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 424, col: 43, offset: 13117},
								expr: &actionExpr{
									pos: position{line: 424, col: 44, offset: 13118},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 424, col: 44, offset: 13118},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 424, col: 44, offset: 13118},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 424, col: 48, offset: 13122},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 424, col: 51, offset: 13125},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 424, col: 57, offset: 13131},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 424, col: 89, offset: 13163},
							expr: &ruleRefExpr{
								pos:  position{line: 424, col: 89, offset: 13163},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 424, col: 104, offset: 13178},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 435, col: 1, offset: 13374},
			expr: &actionExpr{
				pos: position{line: 435, col: 17, offset: 13390},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 435, col: 18, offset: 13391},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 435, col: 18, offset: 13391},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 435, col: 27, offset: 13400},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 439, col: 1, offset: 13455},
			expr: &actionExpr{
				pos: position{line: 439, col: 16, offset: 13470},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 439, col: 16, offset: 13470},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 439, col: 16, offset: 13470},
							expr: &charClassMatcher{
								pos:        position{line: 439, col: 16, offset: 13470},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 439, col: 22, offset: 13476},
							expr: &ruleRefExpr{
								pos:  position{line: 439, col: 22, offset: 13476},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 443, col: 1, offset: 13540},
			expr: &actionExpr{
				pos: position{line: 443, col: 19, offset: 13558},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 443, col: 19, offset: 13558},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 443, col: 19, offset: 13558},
							expr: &charClassMatcher{
								pos:        position{line: 443, col: 19, offset: 13558},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 443, col: 25, offset: 13564},
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 25, offset: 13564},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 443, col: 32, offset: 13571},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 443, col: 36, offset: 13575},
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 36, offset: 13575},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 443, col: 43, offset: 13582},
							expr: &seqExpr{
								pos: position{line: 443, col: 45, offset: 13584},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 443, col: 45, offset: 13584},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 443, col: 52, offset: 13591},
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
			pos:  position{line: 447, col: 1, offset: 13661},
			expr: &actionExpr{
				pos: position{line: 447, col: 14, offset: 13674},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 447, col: 14, offset: 13674},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 447, col: 14, offset: 13674},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 18, offset: 13678},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 447, col: 21, offset: 13681},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 447, col: 28, offset: 13688},
								expr: &seqExpr{
									pos: position{line: 447, col: 29, offset: 13689},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 447, col: 29, offset: 13689},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 447, col: 40, offset: 13700},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 447, col: 43, offset: 13703},
											expr: &ruleRefExpr{
												pos:  position{line: 447, col: 43, offset: 13703},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 447, col: 58, offset: 13718},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 63, offset: 13723},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 447, col: 66, offset: 13726},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 456, col: 1, offset: 13920},
			expr: &actionExpr{
				pos: position{line: 456, col: 13, offset: 13932},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 456, col: 13, offset: 13932},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 456, col: 13, offset: 13932},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 17, offset: 13936},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 456, col: 20, offset: 13939},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 456, col: 27, offset: 13946},
								expr: &seqExpr{
									pos: position{line: 456, col: 28, offset: 13947},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 456, col: 28, offset: 13947},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 39, offset: 13958},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 456, col: 42, offset: 13961},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 46, offset: 13965},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 49, offset: 13968},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 60, offset: 13979},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 456, col: 64, offset: 13983},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 456, col: 64, offset: 13983},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 456, col: 70, offset: 13989},
													expr: &litMatcher{
														pos:        position{line: 456, col: 71, offset: 13990},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 76, offset: 13995},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 456, col: 81, offset: 14000},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 476, col: 1, offset: 14550},
			expr: &ruleRefExpr{
				pos:  position{line: 476, col: 20, offset: 14569},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 478, col: 1, offset: 14576},
			expr: &actionExpr{
				pos: position{line: 478, col: 10, offset: 14585},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 478, col: 10, offset: 14585},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 478, col: 10, offset: 14585},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 478, col: 17, offset: 14592},
								expr: &seqExpr{
									pos: position{line: 478, col: 18, offset: 14593},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 478, col: 18, offset: 14593},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 478, col: 28, offset: 14603},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 478, col: 33, offset: 14608},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 41, offset: 14616},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 478, col: 44, offset: 14619},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 478, col: 49, offset: 14624},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 60, offset: 14635},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 478, col: 63, offset: 14638},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 478, col: 70, offset: 14645},
								expr: &ruleRefExpr{
									pos:  position{line: 478, col: 70, offset: 14645},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 78, offset: 14653},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 478, col: 81, offset: 14656},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 85, offset: 14660},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 478, col: 88, offset: 14663},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 478, col: 99, offset: 14674},
								expr: &seqExpr{
									pos: position{line: 478, col: 100, offset: 14675},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 478, col: 100, offset: 14675},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 478, col: 110, offset: 14685},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 478, col: 116, offset: 14691},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 478, col: 116, offset: 14691},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 478, col: 122, offset: 14697},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 139, offset: 14714},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 478, col: 141, offset: 14716},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 478, col: 153, offset: 14728},
								expr: &ruleRefExpr{
									pos:  position{line: 478, col: 153, offset: 14728},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 478, col: 170, offset: 14745},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 500, col: 1, offset: 15342},
			expr: &actionExpr{
				pos: position{line: 500, col: 20, offset: 15361},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 500, col: 20, offset: 15361,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 504, col: 1, offset: 15428},
			expr: &actionExpr{
				pos: position{line: 504, col: 11, offset: 15438},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 504, col: 11, offset: 15438},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 504, col: 11, offset: 15438},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 504, col: 20, offset: 15447},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 504, col: 23, offset: 15450},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 504, col: 35, offset: 15462},
							expr: &seqExpr{
								pos: position{line: 504, col: 36, offset: 15463},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 504, col: 36, offset: 15463},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 504, col: 40, offset: 15467},
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
			pos:  position{line: 509, col: 1, offset: 15598},
			expr: &choiceExpr{
				pos: position{line: 509, col: 16, offset: 15613},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 509, col: 17, offset: 15614},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 509, col: 17, offset: 15614},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 509, col: 21, offset: 15618},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 509, col: 32, offset: 15629},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 509, col: 39, offset: 15636},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 511, col: 1, offset: 15648},
			expr: &oneOrMoreExpr{
				pos: position{line: 511, col: 15, offset: 15662},
				expr: &charClassMatcher{
					pos:        position{line: 511, col: 15, offset: 15662},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 513, col: 1, offset: 15680},
			expr: &actionExpr{
				pos: position{line: 513, col: 14, offset: 15693},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 513, col: 14, offset: 15693},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 513, col: 14, offset: 15693},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 513, col: 21, offset: 15700},
								expr: &seqExpr{
									pos: position{line: 513, col: 22, offset: 15701},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 513, col: 22, offset: 15701},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 513, col: 32, offset: 15711},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 513, col: 37, offset: 15716},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 513, col: 42, offset: 15721},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 513, col: 53, offset: 15732},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 513, col: 55, offset: 15734},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 513, col: 59, offset: 15738},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 513, col: 62, offset: 15741},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 513, col: 66, offset: 15745},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 513, col: 77, offset: 15756},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 513, col: 79, offset: 15758},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 513, col: 91, offset: 15770},
								expr: &ruleRefExpr{
									pos:  position{line: 513, col: 91, offset: 15770},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 513, col: 108, offset: 15787},
							expr: &ruleRefExpr{
								pos:  position{line: 513, col: 108, offset: 15787},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 530, col: 1, offset: 16373},
			expr: &actionExpr{
				pos: position{line: 530, col: 12, offset: 16384},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 530, col: 13, offset: 16385},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 530, col: 14, offset: 16386},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 530, col: 14, offset: 16386},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 530, col: 18, offset: 16390},
									expr: &choiceExpr{
										pos: position{line: 530, col: 19, offset: 16391},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 530, col: 19, offset: 16391},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 530, col: 26, offset: 16398},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 530, col: 33, offset: 16405},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 530, col: 41, offset: 16413},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 530, col: 41, offset: 16413},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 530, col: 46, offset: 16418},
									expr: &choiceExpr{
										pos: position{line: 530, col: 47, offset: 16419},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 530, col: 47, offset: 16419},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 530, col: 54, offset: 16426},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 530, col: 61, offset: 16433},
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
			pos:  position{line: 539, col: 1, offset: 16719},
			expr: &actionExpr{
				pos: position{line: 539, col: 15, offset: 16733},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 539, col: 15, offset: 16733},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 539, col: 15, offset: 16733},
							expr: &choiceExpr{
								pos: position{line: 539, col: 16, offset: 16734},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 539, col: 16, offset: 16734},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 539, col: 25, offset: 16743},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 539, col: 31, offset: 16749},
							expr: &choiceExpr{
								pos: position{line: 539, col: 32, offset: 16750},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 539, col: 32, offset: 16750},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 539, col: 41, offset: 16759},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 539, col: 49, offset: 16767},
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
			pos:  position{line: 543, col: 1, offset: 16822},
			expr: &charClassMatcher{
				pos:        position{line: 543, col: 18, offset: 16839},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 544, col: 1, offset: 16844},
			expr: &charClassMatcher{
				pos:        position{line: 544, col: 11, offset: 16854},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 545, col: 1, offset: 16863},
			expr: &charClassMatcher{
				pos:        position{line: 545, col: 10, offset: 16872},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 547, col: 1, offset: 16879},
			expr: &anyMatcher{
				line: 547, col: 15, offset: 16893,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 548, col: 1, offset: 16895},
			expr: &actionExpr{
				pos: position{line: 548, col: 14, offset: 16908},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 548, col: 14, offset: 16908},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 548, col: 14, offset: 16908},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 548, col: 21, offset: 16915},
							expr: &seqExpr{
								pos: position{line: 548, col: 23, offset: 16917},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 548, col: 23, offset: 16917},
										expr: &litMatcher{
											pos:        position{line: 548, col: 24, offset: 16918},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 548, col: 29, offset: 16923},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 548, col: 43, offset: 16937},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 554, col: 1, offset: 17117},
			expr: &choiceExpr{
				pos: position{line: 554, col: 12, offset: 17128},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 554, col: 12, offset: 17128},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 554, col: 31, offset: 17147},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 555, col: 1, offset: 17165},
			expr: &seqExpr{
				pos: position{line: 555, col: 21, offset: 17185},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 555, col: 21, offset: 17185},
						expr: &ruleRefExpr{
							pos:  position{line: 555, col: 22, offset: 17186},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 555, col: 32, offset: 17196},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 555, col: 37, offset: 17201},
						expr: &seqExpr{
							pos: position{line: 555, col: 39, offset: 17203},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 555, col: 39, offset: 17203},
									expr: &litMatcher{
										pos:        position{line: 555, col: 40, offset: 17204},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 555, col: 45, offset: 17209},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 555, col: 59, offset: 17223},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 556, col: 1, offset: 17228},
			expr: &seqExpr{
				pos: position{line: 556, col: 37, offset: 17264},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 556, col: 37, offset: 17264},
						expr: &ruleRefExpr{
							pos:  position{line: 556, col: 38, offset: 17265},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 556, col: 48, offset: 17275},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 556, col: 53, offset: 17280},
						expr: &seqExpr{
							pos: position{line: 556, col: 55, offset: 17282},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 556, col: 55, offset: 17282},
									expr: &choiceExpr{
										pos: position{line: 556, col: 58, offset: 17285},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 556, col: 58, offset: 17285},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 556, col: 65, offset: 17292},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 556, col: 71, offset: 17298},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 556, col: 85, offset: 17312},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 557, col: 1, offset: 17317},
			expr: &choiceExpr{
				pos: position{line: 557, col: 22, offset: 17338},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 557, col: 23, offset: 17339},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 557, col: 23, offset: 17339},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 557, col: 28, offset: 17344},
								expr: &seqExpr{
									pos: position{line: 557, col: 30, offset: 17346},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 557, col: 30, offset: 17346},
											expr: &ruleRefExpr{
												pos:  position{line: 557, col: 31, offset: 17347},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 557, col: 35, offset: 17351},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 557, col: 53, offset: 17369},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 557, col: 53, offset: 17369},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 557, col: 57, offset: 17373},
								expr: &seqExpr{
									pos: position{line: 557, col: 59, offset: 17375},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 557, col: 59, offset: 17375},
											expr: &ruleRefExpr{
												pos:  position{line: 557, col: 60, offset: 17376},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 557, col: 64, offset: 17380},
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
			pos:  position{line: 559, col: 1, offset: 17396},
			expr: &zeroOrMoreExpr{
				pos: position{line: 559, col: 7, offset: 17402},
				expr: &choiceExpr{
					pos: position{line: 559, col: 9, offset: 17404},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 559, col: 9, offset: 17404},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 559, col: 22, offset: 17417},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 559, col: 28, offset: 17423},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 560, col: 1, offset: 17434},
			expr: &zeroOrMoreExpr{
				pos: position{line: 560, col: 6, offset: 17439},
				expr: &choiceExpr{
					pos: position{line: 560, col: 8, offset: 17441},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 560, col: 8, offset: 17441},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 560, col: 21, offset: 17454},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 561, col: 1, offset: 17490},
			expr: &zeroOrMoreExpr{
				pos: position{line: 561, col: 7, offset: 17496},
				expr: &ruleRefExpr{
					pos:  position{line: 561, col: 7, offset: 17496},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 563, col: 1, offset: 17509},
			expr: &charClassMatcher{
				pos:        position{line: 563, col: 15, offset: 17523},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 564, col: 1, offset: 17531},
			expr: &litMatcher{
				pos:        position{line: 564, col: 8, offset: 17538},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 565, col: 1, offset: 17543},
			expr: &choiceExpr{
				pos: position{line: 565, col: 8, offset: 17550},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 565, col: 8, offset: 17550},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 565, col: 8, offset: 17550},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 565, col: 11, offset: 17553},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 565, col: 17, offset: 17559},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 565, col: 17, offset: 17559},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 565, col: 19, offset: 17561},
								expr: &ruleRefExpr{
									pos:  position{line: 565, col: 19, offset: 17561},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 565, col: 38, offset: 17580},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 565, col: 44, offset: 17586},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 565, col: 44, offset: 17586},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 565, col: 47, offset: 17589},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 567, col: 1, offset: 17594},
			expr: &notExpr{
				pos: position{line: 567, col: 8, offset: 17601},
				expr: &anyMatcher{
					line: 567, col: 9, offset: 17602,
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
	name := file.(string)
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
		Type:        &Type{Name: string(typ.(Identifier))},
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
