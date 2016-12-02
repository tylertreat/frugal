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
			pos:  position{line: 156, col: 1, offset: 5051},
			expr: &actionExpr{
				pos: position{line: 156, col: 16, offset: 5066},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 156, col: 16, offset: 5066,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 160, col: 1, offset: 5124},
			expr: &actionExpr{
				pos: position{line: 160, col: 14, offset: 5137},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 160, col: 14, offset: 5137},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 160, col: 14, offset: 5137},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 160, col: 21, offset: 5144},
								expr: &seqExpr{
									pos: position{line: 160, col: 22, offset: 5145},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 160, col: 22, offset: 5145},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 160, col: 32, offset: 5155},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 160, col: 37, offset: 5160},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 160, col: 48, offset: 5171},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 160, col: 48, offset: 5171},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 160, col: 66, offset: 5189},
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
			pos:  position{line: 173, col: 1, offset: 5660},
			expr: &choiceExpr{
				pos: position{line: 173, col: 20, offset: 5679},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 173, col: 20, offset: 5679},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 30, offset: 5689},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 42, offset: 5701},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 50, offset: 5709},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 57, offset: 5716},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 67, offset: 5726},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 76, offset: 5735},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 88, offset: 5747},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 96, offset: 5755},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 175, col: 1, offset: 5764},
			expr: &actionExpr{
				pos: position{line: 175, col: 12, offset: 5775},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 175, col: 12, offset: 5775},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 12, offset: 5775},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 22, offset: 5785},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 24, offset: 5787},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 29, offset: 5792},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 37, offset: 5800},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 183, col: 1, offset: 5977},
			expr: &actionExpr{
				pos: position{line: 183, col: 14, offset: 5990},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 183, col: 14, offset: 5990},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 183, col: 14, offset: 5990},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 26, offset: 6002},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 183, col: 28, offset: 6004},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 183, col: 34, offset: 6010},
								expr: &charClassMatcher{
									pos:        position{line: 183, col: 34, offset: 6010},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 44, offset: 6020},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 183, col: 46, offset: 6022},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 183, col: 49, offset: 6025},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 60, offset: 6036},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 183, col: 62, offset: 6038},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 183, col: 74, offset: 6050},
								expr: &ruleRefExpr{
									pos:  position{line: 183, col: 74, offset: 6050},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 91, offset: 6067},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 191, col: 1, offset: 6253},
			expr: &actionExpr{
				pos: position{line: 191, col: 10, offset: 6262},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 191, col: 10, offset: 6262},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 191, col: 10, offset: 6262},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 18, offset: 6270},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 191, col: 20, offset: 6272},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 191, col: 24, offset: 6276},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 34, offset: 6286},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 191, col: 36, offset: 6288},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 191, col: 41, offset: 6293},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 52, offset: 6304},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 191, col: 54, offset: 6306},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 58, offset: 6310},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 191, col: 60, offset: 6312},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 191, col: 66, offset: 6318},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 77, offset: 6329},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 191, col: 79, offset: 6331},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 191, col: 91, offset: 6343},
								expr: &ruleRefExpr{
									pos:  position{line: 191, col: 91, offset: 6343},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 191, col: 108, offset: 6360},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 200, col: 1, offset: 6554},
			expr: &actionExpr{
				pos: position{line: 200, col: 9, offset: 6562},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 200, col: 9, offset: 6562},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 200, col: 9, offset: 6562},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 200, col: 16, offset: 6569},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 200, col: 18, offset: 6571},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 200, col: 23, offset: 6576},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 200, col: 34, offset: 6587},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 200, col: 37, offset: 6590},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 200, col: 41, offset: 6594},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 200, col: 44, offset: 6597},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 200, col: 51, offset: 6604},
								expr: &seqExpr{
									pos: position{line: 200, col: 52, offset: 6605},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 200, col: 52, offset: 6605},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 200, col: 62, offset: 6615},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 200, col: 67, offset: 6620},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 200, col: 71, offset: 6624},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 200, col: 73, offset: 6626},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 200, col: 85, offset: 6638},
								expr: &ruleRefExpr{
									pos:  position{line: 200, col: 85, offset: 6638},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 200, col: 102, offset: 6655},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 224, col: 1, offset: 7317},
			expr: &actionExpr{
				pos: position{line: 224, col: 14, offset: 7330},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 224, col: 14, offset: 7330},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 224, col: 14, offset: 7330},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 224, col: 21, offset: 7337},
								expr: &seqExpr{
									pos: position{line: 224, col: 22, offset: 7338},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 224, col: 22, offset: 7338},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 224, col: 32, offset: 7348},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 224, col: 37, offset: 7353},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 224, col: 42, offset: 7358},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 224, col: 53, offset: 7369},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 224, col: 55, offset: 7371},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 224, col: 61, offset: 7377},
								expr: &seqExpr{
									pos: position{line: 224, col: 62, offset: 7378},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 224, col: 62, offset: 7378},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 224, col: 66, offset: 7382},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 224, col: 68, offset: 7384},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 224, col: 82, offset: 7398},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 224, col: 84, offset: 7400},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 224, col: 96, offset: 7412},
								expr: &ruleRefExpr{
									pos:  position{line: 224, col: 96, offset: 7412},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 224, col: 113, offset: 7429},
							expr: &ruleRefExpr{
								pos:  position{line: 224, col: 113, offset: 7429},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 240, col: 1, offset: 7827},
			expr: &actionExpr{
				pos: position{line: 240, col: 12, offset: 7838},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 240, col: 12, offset: 7838},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 240, col: 12, offset: 7838},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 240, col: 22, offset: 7848},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 240, col: 24, offset: 7850},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 240, col: 28, offset: 7854},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 240, col: 38, offset: 7864},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 240, col: 40, offset: 7866},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 240, col: 45, offset: 7871},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 240, col: 56, offset: 7882},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 240, col: 58, offset: 7884},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 240, col: 70, offset: 7896},
								expr: &ruleRefExpr{
									pos:  position{line: 240, col: 70, offset: 7896},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 240, col: 87, offset: 7913},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 248, col: 1, offset: 8085},
			expr: &actionExpr{
				pos: position{line: 248, col: 11, offset: 8095},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 248, col: 11, offset: 8095},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 248, col: 11, offset: 8095},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 20, offset: 8104},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 22, offset: 8106},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 25, offset: 8109},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 249, col: 1, offset: 8149},
			expr: &actionExpr{
				pos: position{line: 249, col: 14, offset: 8162},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 249, col: 14, offset: 8162},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 249, col: 14, offset: 8162},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 26, offset: 8174},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 28, offset: 8176},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 31, offset: 8179},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 250, col: 1, offset: 8230},
			expr: &actionExpr{
				pos: position{line: 250, col: 10, offset: 8239},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 250, col: 10, offset: 8239},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 250, col: 10, offset: 8239},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 18, offset: 8247},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 20, offset: 8249},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 23, offset: 8252},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 251, col: 1, offset: 8299},
			expr: &actionExpr{
				pos: position{line: 251, col: 15, offset: 8313},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 251, col: 15, offset: 8313},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 251, col: 15, offset: 8313},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 251, col: 20, offset: 8318},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 251, col: 31, offset: 8329},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 251, col: 34, offset: 8332},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 251, col: 38, offset: 8336},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 251, col: 41, offset: 8339},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 251, col: 48, offset: 8346},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 251, col: 58, offset: 8356},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 251, col: 62, offset: 8360},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 251, col: 64, offset: 8362},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 251, col: 76, offset: 8374},
								expr: &ruleRefExpr{
									pos:  position{line: 251, col: 76, offset: 8374},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 251, col: 93, offset: 8391},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 262, col: 1, offset: 8608},
			expr: &actionExpr{
				pos: position{line: 262, col: 14, offset: 8621},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 262, col: 14, offset: 8621},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 262, col: 21, offset: 8628},
						expr: &seqExpr{
							pos: position{line: 262, col: 22, offset: 8629},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 262, col: 22, offset: 8629},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 262, col: 28, offset: 8635},
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
			pos:  position{line: 271, col: 1, offset: 8816},
			expr: &actionExpr{
				pos: position{line: 271, col: 10, offset: 8825},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 271, col: 10, offset: 8825},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 271, col: 10, offset: 8825},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 271, col: 17, offset: 8832},
								expr: &seqExpr{
									pos: position{line: 271, col: 18, offset: 8833},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 271, col: 18, offset: 8833},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 28, offset: 8843},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 271, col: 33, offset: 8848},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 271, col: 36, offset: 8851},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 48, offset: 8863},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 271, col: 50, offset: 8865},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 54, offset: 8869},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 56, offset: 8871},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 271, col: 60, offset: 8875},
								expr: &ruleRefExpr{
									pos:  position{line: 271, col: 60, offset: 8875},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 75, offset: 8890},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 77, offset: 8892},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 271, col: 81, offset: 8896},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 91, offset: 8906},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 93, offset: 8908},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 271, col: 98, offset: 8913},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 109, offset: 8924},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 112, offset: 8927},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 271, col: 116, offset: 8931},
								expr: &seqExpr{
									pos: position{line: 271, col: 117, offset: 8932},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 271, col: 117, offset: 8932},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 121, offset: 8936},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 123, offset: 8938},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 136, offset: 8951},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 138, offset: 8953},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 271, col: 150, offset: 8965},
								expr: &ruleRefExpr{
									pos:  position{line: 271, col: 150, offset: 8965},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 271, col: 167, offset: 8982},
							expr: &ruleRefExpr{
								pos:  position{line: 271, col: 167, offset: 8982},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 294, col: 1, offset: 9514},
			expr: &actionExpr{
				pos: position{line: 294, col: 18, offset: 9531},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 294, col: 19, offset: 9532},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 294, col: 19, offset: 9532},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 294, col: 32, offset: 9545},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 302, col: 1, offset: 9688},
			expr: &actionExpr{
				pos: position{line: 302, col: 12, offset: 9699},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 302, col: 12, offset: 9699},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 302, col: 12, offset: 9699},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 22, offset: 9709},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 302, col: 24, offset: 9711},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 302, col: 29, offset: 9716},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 40, offset: 9727},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 302, col: 42, offset: 9729},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 302, col: 50, offset: 9737},
								expr: &seqExpr{
									pos: position{line: 302, col: 51, offset: 9738},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 302, col: 51, offset: 9738},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 302, col: 61, offset: 9748},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 302, col: 64, offset: 9751},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 302, col: 75, offset: 9762},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 80, offset: 9767},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 302, col: 83, offset: 9770},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 87, offset: 9774},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 302, col: 90, offset: 9777},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 302, col: 98, offset: 9785},
								expr: &seqExpr{
									pos: position{line: 302, col: 99, offset: 9786},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 302, col: 99, offset: 9786},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 302, col: 108, offset: 9795},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 302, col: 114, offset: 9801},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 302, col: 114, offset: 9801},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 302, col: 120, offset: 9807},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 139, offset: 9826},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 302, col: 141, offset: 9828},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 302, col: 153, offset: 9840},
								expr: &ruleRefExpr{
									pos:  position{line: 302, col: 153, offset: 9840},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 170, offset: 9857},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 319, col: 1, offset: 10298},
			expr: &actionExpr{
				pos: position{line: 319, col: 22, offset: 10319},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 319, col: 22, offset: 10319,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 323, col: 1, offset: 10388},
			expr: &actionExpr{
				pos: position{line: 323, col: 13, offset: 10400},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 323, col: 13, offset: 10400},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 323, col: 13, offset: 10400},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 323, col: 20, offset: 10407},
								expr: &seqExpr{
									pos: position{line: 323, col: 21, offset: 10408},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 323, col: 21, offset: 10408},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 323, col: 31, offset: 10418},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 323, col: 36, offset: 10423},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 323, col: 43, offset: 10430},
								expr: &seqExpr{
									pos: position{line: 323, col: 44, offset: 10431},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 323, col: 44, offset: 10431},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 323, col: 53, offset: 10440},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 323, col: 58, offset: 10445},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 323, col: 62, offset: 10449},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 323, col: 75, offset: 10462},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 323, col: 78, offset: 10465},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 323, col: 83, offset: 10470},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 323, col: 94, offset: 10481},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 323, col: 96, offset: 10483},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 323, col: 100, offset: 10487},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 323, col: 103, offset: 10490},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 323, col: 113, offset: 10500},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 323, col: 123, offset: 10510},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 323, col: 127, offset: 10514},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 323, col: 130, offset: 10517},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 323, col: 141, offset: 10528},
								expr: &ruleRefExpr{
									pos:  position{line: 323, col: 141, offset: 10528},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 323, col: 149, offset: 10536},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 323, col: 151, offset: 10538},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 323, col: 163, offset: 10550},
								expr: &ruleRefExpr{
									pos:  position{line: 323, col: 163, offset: 10550},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 323, col: 180, offset: 10567},
							expr: &ruleRefExpr{
								pos:  position{line: 323, col: 180, offset: 10567},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 351, col: 1, offset: 11218},
			expr: &actionExpr{
				pos: position{line: 351, col: 17, offset: 11234},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 351, col: 17, offset: 11234},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 351, col: 22, offset: 11239},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 351, col: 22, offset: 11239},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 351, col: 31, offset: 11248},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 358, col: 1, offset: 11370},
			expr: &actionExpr{
				pos: position{line: 358, col: 11, offset: 11380},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 358, col: 11, offset: 11380},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 358, col: 11, offset: 11380},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 358, col: 20, offset: 11389},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 358, col: 23, offset: 11392},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 358, col: 27, offset: 11396},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 358, col: 30, offset: 11399},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 358, col: 41, offset: 11410},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 358, col: 51, offset: 11420},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 362, col: 1, offset: 11456},
			expr: &actionExpr{
				pos: position{line: 362, col: 14, offset: 11469},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 362, col: 14, offset: 11469},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 362, col: 19, offset: 11474},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 362, col: 19, offset: 11474},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 362, col: 30, offset: 11485},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 362, col: 46, offset: 11501},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 369, col: 1, offset: 11626},
			expr: &actionExpr{
				pos: position{line: 369, col: 13, offset: 11638},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 369, col: 13, offset: 11638},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 369, col: 13, offset: 11638},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 369, col: 18, offset: 11643},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 369, col: 31, offset: 11656},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 369, col: 33, offset: 11658},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 369, col: 45, offset: 11670},
								expr: &ruleRefExpr{
									pos:  position{line: 369, col: 45, offset: 11670},
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
			pos:  position{line: 376, col: 1, offset: 11806},
			expr: &actionExpr{
				pos: position{line: 376, col: 17, offset: 11822},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 376, col: 18, offset: 11823},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 376, col: 18, offset: 11823},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 376, col: 27, offset: 11832},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 376, col: 36, offset: 11841},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 376, col: 44, offset: 11849},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 376, col: 52, offset: 11857},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 376, col: 60, offset: 11865},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 376, col: 71, offset: 11876},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 376, col: 82, offset: 11887},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 380, col: 1, offset: 11934},
			expr: &actionExpr{
				pos: position{line: 380, col: 18, offset: 11951},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 380, col: 18, offset: 11951},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 380, col: 23, offset: 11956},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 380, col: 23, offset: 11956},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 380, col: 33, offset: 11966},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 380, col: 43, offset: 11976},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 384, col: 1, offset: 12011},
			expr: &actionExpr{
				pos: position{line: 384, col: 12, offset: 12022},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 384, col: 12, offset: 12022},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 384, col: 12, offset: 12022},
							expr: &ruleRefExpr{
								pos:  position{line: 384, col: 12, offset: 12022},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 384, col: 21, offset: 12031},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 384, col: 28, offset: 12038},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 384, col: 31, offset: 12041},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 384, col: 35, offset: 12045},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 384, col: 45, offset: 12055},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 384, col: 48, offset: 12058},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 384, col: 52, offset: 12062},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 384, col: 55, offset: 12065},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 384, col: 61, offset: 12071},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 384, col: 71, offset: 12081},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 384, col: 74, offset: 12084},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 384, col: 78, offset: 12088},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 384, col: 80, offset: 12090},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 384, col: 92, offset: 12102},
								expr: &ruleRefExpr{
									pos:  position{line: 384, col: 92, offset: 12102},
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
			pos:  position{line: 393, col: 1, offset: 12300},
			expr: &actionExpr{
				pos: position{line: 393, col: 12, offset: 12311},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 393, col: 12, offset: 12311},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 393, col: 12, offset: 12311},
							expr: &ruleRefExpr{
								pos:  position{line: 393, col: 12, offset: 12311},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 393, col: 21, offset: 12320},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 393, col: 28, offset: 12327},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 393, col: 31, offset: 12330},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 393, col: 35, offset: 12334},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 393, col: 45, offset: 12344},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 393, col: 48, offset: 12347},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 393, col: 52, offset: 12351},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 393, col: 54, offset: 12353},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 393, col: 66, offset: 12365},
								expr: &ruleRefExpr{
									pos:  position{line: 393, col: 66, offset: 12365},
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
			pos:  position{line: 401, col: 1, offset: 12527},
			expr: &actionExpr{
				pos: position{line: 401, col: 13, offset: 12539},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 401, col: 13, offset: 12539},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 401, col: 13, offset: 12539},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 401, col: 21, offset: 12547},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 401, col: 24, offset: 12550},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 401, col: 28, offset: 12554},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 401, col: 38, offset: 12564},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 401, col: 41, offset: 12567},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 401, col: 45, offset: 12571},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 401, col: 47, offset: 12573},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 401, col: 59, offset: 12585},
								expr: &ruleRefExpr{
									pos:  position{line: 401, col: 59, offset: 12585},
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
			pos:  position{line: 409, col: 1, offset: 12748},
			expr: &actionExpr{
				pos: position{line: 409, col: 12, offset: 12759},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 409, col: 12, offset: 12759},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 409, col: 12, offset: 12759},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 409, col: 23, offset: 12770},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 409, col: 31, offset: 12778},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 413, col: 1, offset: 12815},
			expr: &choiceExpr{
				pos: position{line: 413, col: 15, offset: 12829},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 413, col: 15, offset: 12829},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 413, col: 25, offset: 12839},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 413, col: 40, offset: 12854},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 413, col: 57, offset: 12871},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 413, col: 71, offset: 12885},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 413, col: 82, offset: 12896},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 413, col: 94, offset: 12908},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 415, col: 1, offset: 12920},
			expr: &actionExpr{
				pos: position{line: 415, col: 20, offset: 12939},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 415, col: 20, offset: 12939},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 415, col: 20, offset: 12939},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 415, col: 24, offset: 12943},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 415, col: 27, offset: 12946},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 415, col: 39, offset: 12958},
								expr: &ruleRefExpr{
									pos:  position{line: 415, col: 39, offset: 12958},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 415, col: 55, offset: 12974},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 423, col: 1, offset: 13138},
			expr: &actionExpr{
				pos: position{line: 423, col: 19, offset: 13156},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 423, col: 19, offset: 13156},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 423, col: 19, offset: 13156},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 423, col: 24, offset: 13161},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 423, col: 35, offset: 13172},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 423, col: 37, offset: 13174},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 423, col: 43, offset: 13180},
								expr: &actionExpr{
									pos: position{line: 423, col: 44, offset: 13181},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 423, col: 44, offset: 13181},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 423, col: 44, offset: 13181},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 423, col: 48, offset: 13185},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 423, col: 51, offset: 13188},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 423, col: 57, offset: 13194},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 423, col: 89, offset: 13226},
							expr: &ruleRefExpr{
								pos:  position{line: 423, col: 89, offset: 13226},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 423, col: 104, offset: 13241},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 434, col: 1, offset: 13437},
			expr: &actionExpr{
				pos: position{line: 434, col: 17, offset: 13453},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 434, col: 18, offset: 13454},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 434, col: 18, offset: 13454},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 434, col: 27, offset: 13463},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 438, col: 1, offset: 13518},
			expr: &actionExpr{
				pos: position{line: 438, col: 16, offset: 13533},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 438, col: 16, offset: 13533},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 438, col: 16, offset: 13533},
							expr: &charClassMatcher{
								pos:        position{line: 438, col: 16, offset: 13533},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 438, col: 22, offset: 13539},
							expr: &ruleRefExpr{
								pos:  position{line: 438, col: 22, offset: 13539},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 442, col: 1, offset: 13603},
			expr: &actionExpr{
				pos: position{line: 442, col: 19, offset: 13621},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 442, col: 19, offset: 13621},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 442, col: 19, offset: 13621},
							expr: &charClassMatcher{
								pos:        position{line: 442, col: 19, offset: 13621},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 442, col: 25, offset: 13627},
							expr: &ruleRefExpr{
								pos:  position{line: 442, col: 25, offset: 13627},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 442, col: 32, offset: 13634},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 442, col: 36, offset: 13638},
							expr: &ruleRefExpr{
								pos:  position{line: 442, col: 36, offset: 13638},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 442, col: 43, offset: 13645},
							expr: &seqExpr{
								pos: position{line: 442, col: 45, offset: 13647},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 442, col: 45, offset: 13647},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 442, col: 52, offset: 13654},
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
			pos:  position{line: 446, col: 1, offset: 13724},
			expr: &actionExpr{
				pos: position{line: 446, col: 14, offset: 13737},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 446, col: 14, offset: 13737},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 446, col: 14, offset: 13737},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 446, col: 18, offset: 13741},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 446, col: 21, offset: 13744},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 446, col: 28, offset: 13751},
								expr: &seqExpr{
									pos: position{line: 446, col: 29, offset: 13752},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 446, col: 29, offset: 13752},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 446, col: 40, offset: 13763},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 446, col: 43, offset: 13766},
											expr: &ruleRefExpr{
												pos:  position{line: 446, col: 43, offset: 13766},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 446, col: 58, offset: 13781},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 446, col: 63, offset: 13786},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 446, col: 66, offset: 13789},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 455, col: 1, offset: 13983},
			expr: &actionExpr{
				pos: position{line: 455, col: 13, offset: 13995},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 455, col: 13, offset: 13995},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 455, col: 13, offset: 13995},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 455, col: 17, offset: 13999},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 455, col: 20, offset: 14002},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 455, col: 27, offset: 14009},
								expr: &seqExpr{
									pos: position{line: 455, col: 28, offset: 14010},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 455, col: 28, offset: 14010},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 455, col: 39, offset: 14021},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 455, col: 42, offset: 14024},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 455, col: 46, offset: 14028},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 455, col: 49, offset: 14031},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 455, col: 60, offset: 14042},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 455, col: 64, offset: 14046},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 455, col: 64, offset: 14046},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 455, col: 70, offset: 14052},
													expr: &litMatcher{
														pos:        position{line: 455, col: 71, offset: 14053},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 455, col: 76, offset: 14058},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 455, col: 81, offset: 14063},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 475, col: 1, offset: 14613},
			expr: &ruleRefExpr{
				pos:  position{line: 475, col: 20, offset: 14632},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 477, col: 1, offset: 14639},
			expr: &actionExpr{
				pos: position{line: 477, col: 10, offset: 14648},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 477, col: 10, offset: 14648},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 477, col: 10, offset: 14648},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 477, col: 17, offset: 14655},
								expr: &seqExpr{
									pos: position{line: 477, col: 18, offset: 14656},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 477, col: 18, offset: 14656},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 477, col: 28, offset: 14666},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 477, col: 33, offset: 14671},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 477, col: 41, offset: 14679},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 477, col: 44, offset: 14682},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 477, col: 49, offset: 14687},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 477, col: 60, offset: 14698},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 477, col: 63, offset: 14701},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 477, col: 70, offset: 14708},
								expr: &ruleRefExpr{
									pos:  position{line: 477, col: 70, offset: 14708},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 477, col: 78, offset: 14716},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 477, col: 81, offset: 14719},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 477, col: 85, offset: 14723},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 477, col: 88, offset: 14726},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 477, col: 99, offset: 14737},
								expr: &seqExpr{
									pos: position{line: 477, col: 100, offset: 14738},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 477, col: 100, offset: 14738},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 477, col: 110, offset: 14748},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 477, col: 116, offset: 14754},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 477, col: 116, offset: 14754},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 477, col: 122, offset: 14760},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 477, col: 139, offset: 14777},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 477, col: 141, offset: 14779},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 477, col: 153, offset: 14791},
								expr: &ruleRefExpr{
									pos:  position{line: 477, col: 153, offset: 14791},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 477, col: 170, offset: 14808},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 499, col: 1, offset: 15405},
			expr: &actionExpr{
				pos: position{line: 499, col: 20, offset: 15424},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 499, col: 20, offset: 15424,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 503, col: 1, offset: 15491},
			expr: &actionExpr{
				pos: position{line: 503, col: 11, offset: 15501},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 503, col: 11, offset: 15501},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 503, col: 11, offset: 15501},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 503, col: 20, offset: 15510},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 503, col: 23, offset: 15513},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 503, col: 35, offset: 15525},
							expr: &seqExpr{
								pos: position{line: 503, col: 36, offset: 15526},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 503, col: 36, offset: 15526},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 503, col: 40, offset: 15530},
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
			pos:  position{line: 508, col: 1, offset: 15661},
			expr: &choiceExpr{
				pos: position{line: 508, col: 16, offset: 15676},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 508, col: 17, offset: 15677},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 508, col: 17, offset: 15677},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 508, col: 21, offset: 15681},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 508, col: 32, offset: 15692},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 508, col: 39, offset: 15699},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 510, col: 1, offset: 15711},
			expr: &oneOrMoreExpr{
				pos: position{line: 510, col: 15, offset: 15725},
				expr: &charClassMatcher{
					pos:        position{line: 510, col: 15, offset: 15725},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 512, col: 1, offset: 15743},
			expr: &actionExpr{
				pos: position{line: 512, col: 14, offset: 15756},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 512, col: 14, offset: 15756},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 512, col: 14, offset: 15756},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 512, col: 21, offset: 15763},
								expr: &seqExpr{
									pos: position{line: 512, col: 22, offset: 15764},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 512, col: 22, offset: 15764},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 512, col: 32, offset: 15774},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 512, col: 37, offset: 15779},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 512, col: 42, offset: 15784},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 512, col: 53, offset: 15795},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 512, col: 55, offset: 15797},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 512, col: 59, offset: 15801},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 512, col: 62, offset: 15804},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 512, col: 66, offset: 15808},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 512, col: 77, offset: 15819},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 512, col: 79, offset: 15821},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 512, col: 91, offset: 15833},
								expr: &ruleRefExpr{
									pos:  position{line: 512, col: 91, offset: 15833},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 512, col: 108, offset: 15850},
							expr: &ruleRefExpr{
								pos:  position{line: 512, col: 108, offset: 15850},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 529, col: 1, offset: 16436},
			expr: &actionExpr{
				pos: position{line: 529, col: 12, offset: 16447},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 529, col: 13, offset: 16448},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 529, col: 14, offset: 16449},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 529, col: 14, offset: 16449},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 529, col: 18, offset: 16453},
									expr: &choiceExpr{
										pos: position{line: 529, col: 19, offset: 16454},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 529, col: 19, offset: 16454},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 529, col: 26, offset: 16461},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 529, col: 33, offset: 16468},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 529, col: 41, offset: 16476},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 529, col: 41, offset: 16476},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 529, col: 46, offset: 16481},
									expr: &choiceExpr{
										pos: position{line: 529, col: 47, offset: 16482},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 529, col: 47, offset: 16482},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 529, col: 54, offset: 16489},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 529, col: 61, offset: 16496},
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
			pos:  position{line: 538, col: 1, offset: 16782},
			expr: &actionExpr{
				pos: position{line: 538, col: 15, offset: 16796},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 538, col: 15, offset: 16796},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 538, col: 15, offset: 16796},
							expr: &choiceExpr{
								pos: position{line: 538, col: 16, offset: 16797},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 538, col: 16, offset: 16797},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 538, col: 25, offset: 16806},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 538, col: 31, offset: 16812},
							expr: &choiceExpr{
								pos: position{line: 538, col: 32, offset: 16813},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 538, col: 32, offset: 16813},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 538, col: 41, offset: 16822},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 538, col: 49, offset: 16830},
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
			pos:  position{line: 542, col: 1, offset: 16885},
			expr: &charClassMatcher{
				pos:        position{line: 542, col: 18, offset: 16902},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 543, col: 1, offset: 16907},
			expr: &charClassMatcher{
				pos:        position{line: 543, col: 11, offset: 16917},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 544, col: 1, offset: 16926},
			expr: &charClassMatcher{
				pos:        position{line: 544, col: 10, offset: 16935},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 546, col: 1, offset: 16942},
			expr: &anyMatcher{
				line: 546, col: 15, offset: 16956,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 547, col: 1, offset: 16958},
			expr: &actionExpr{
				pos: position{line: 547, col: 14, offset: 16971},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 547, col: 14, offset: 16971},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 547, col: 14, offset: 16971},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 547, col: 21, offset: 16978},
							expr: &seqExpr{
								pos: position{line: 547, col: 23, offset: 16980},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 547, col: 23, offset: 16980},
										expr: &litMatcher{
											pos:        position{line: 547, col: 24, offset: 16981},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 547, col: 29, offset: 16986},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 547, col: 43, offset: 17000},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 553, col: 1, offset: 17180},
			expr: &choiceExpr{
				pos: position{line: 553, col: 12, offset: 17191},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 553, col: 12, offset: 17191},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 553, col: 31, offset: 17210},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 554, col: 1, offset: 17228},
			expr: &seqExpr{
				pos: position{line: 554, col: 21, offset: 17248},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 554, col: 21, offset: 17248},
						expr: &ruleRefExpr{
							pos:  position{line: 554, col: 22, offset: 17249},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 554, col: 32, offset: 17259},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 554, col: 37, offset: 17264},
						expr: &seqExpr{
							pos: position{line: 554, col: 39, offset: 17266},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 554, col: 39, offset: 17266},
									expr: &litMatcher{
										pos:        position{line: 554, col: 40, offset: 17267},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 554, col: 45, offset: 17272},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 554, col: 59, offset: 17286},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 555, col: 1, offset: 17291},
			expr: &seqExpr{
				pos: position{line: 555, col: 37, offset: 17327},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 555, col: 37, offset: 17327},
						expr: &ruleRefExpr{
							pos:  position{line: 555, col: 38, offset: 17328},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 555, col: 48, offset: 17338},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 555, col: 53, offset: 17343},
						expr: &seqExpr{
							pos: position{line: 555, col: 55, offset: 17345},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 555, col: 55, offset: 17345},
									expr: &choiceExpr{
										pos: position{line: 555, col: 58, offset: 17348},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 555, col: 58, offset: 17348},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 555, col: 65, offset: 17355},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 555, col: 71, offset: 17361},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 555, col: 85, offset: 17375},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 556, col: 1, offset: 17380},
			expr: &choiceExpr{
				pos: position{line: 556, col: 22, offset: 17401},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 556, col: 23, offset: 17402},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 556, col: 23, offset: 17402},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 556, col: 28, offset: 17407},
								expr: &seqExpr{
									pos: position{line: 556, col: 30, offset: 17409},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 556, col: 30, offset: 17409},
											expr: &ruleRefExpr{
												pos:  position{line: 556, col: 31, offset: 17410},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 556, col: 35, offset: 17414},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 556, col: 53, offset: 17432},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 556, col: 53, offset: 17432},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 556, col: 57, offset: 17436},
								expr: &seqExpr{
									pos: position{line: 556, col: 59, offset: 17438},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 556, col: 59, offset: 17438},
											expr: &ruleRefExpr{
												pos:  position{line: 556, col: 60, offset: 17439},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 556, col: 64, offset: 17443},
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
			pos:  position{line: 558, col: 1, offset: 17459},
			expr: &zeroOrMoreExpr{
				pos: position{line: 558, col: 7, offset: 17465},
				expr: &choiceExpr{
					pos: position{line: 558, col: 9, offset: 17467},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 558, col: 9, offset: 17467},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 558, col: 22, offset: 17480},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 558, col: 28, offset: 17486},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 559, col: 1, offset: 17497},
			expr: &zeroOrMoreExpr{
				pos: position{line: 559, col: 6, offset: 17502},
				expr: &choiceExpr{
					pos: position{line: 559, col: 8, offset: 17504},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 559, col: 8, offset: 17504},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 559, col: 21, offset: 17517},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 560, col: 1, offset: 17553},
			expr: &zeroOrMoreExpr{
				pos: position{line: 560, col: 7, offset: 17559},
				expr: &ruleRefExpr{
					pos:  position{line: 560, col: 7, offset: 17559},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 562, col: 1, offset: 17572},
			expr: &charClassMatcher{
				pos:        position{line: 562, col: 15, offset: 17586},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 563, col: 1, offset: 17594},
			expr: &litMatcher{
				pos:        position{line: 563, col: 8, offset: 17601},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 564, col: 1, offset: 17606},
			expr: &choiceExpr{
				pos: position{line: 564, col: 8, offset: 17613},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 564, col: 8, offset: 17613},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 564, col: 8, offset: 17613},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 564, col: 11, offset: 17616},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 564, col: 17, offset: 17622},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 564, col: 17, offset: 17622},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 564, col: 19, offset: 17624},
								expr: &ruleRefExpr{
									pos:  position{line: 564, col: 19, offset: 17624},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 564, col: 38, offset: 17643},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 564, col: 44, offset: 17649},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 564, col: 44, offset: 17649},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 564, col: 47, offset: 17652},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 566, col: 1, offset: 17657},
			expr: &notExpr{
				pos: position{line: 566, col: 8, offset: 17664},
				expr: &anyMatcher{
					line: 566, col: 9, offset: 17665,
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
		Thrift:             thrift,
		Scopes:             []*Scope{},
		ParsedIncludes:     make(map[string]*Frugal),
		OrderedDefinitions: make([]interface{}, len(stmts)),
	}

	for i, st := range stmts {
		wrapper := st.([]interface{})[0].(*statementWrapper)
		frugal.OrderedDefinitions[i] = wrapper.statement
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
			frugal.OrderedDefinitions[i] = strct
		case union:
			strct := unionToStruct(v)
			strct.Type = StructTypeUnion
			strct.Comment = wrapper.comment
			thrift.Unions = append(thrift.Unions, strct)
			frugal.OrderedDefinitions[i] = strct
		case *Service:
			v.Comment = wrapper.comment
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

func (c *current) onInclude1(file interface{}) (interface{}, error) {
	name := file.(string)
	if ix := strings.LastIndex(name, "."); ix > 0 {
		name = name[:ix]
	}
	return &Include{Name: name, Value: file.(string)}, nil
}

func (p *parser) callonInclude1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onInclude1(stack["file"])
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
