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

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 77, col: 1, offset: 2164},
			expr: &actionExpr{
				pos: position{line: 77, col: 11, offset: 2176},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 77, col: 11, offset: 2176},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 77, col: 11, offset: 2176},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 77, col: 14, offset: 2179},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 77, col: 25, offset: 2190},
								expr: &seqExpr{
									pos: position{line: 77, col: 27, offset: 2192},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 77, col: 27, offset: 2192},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 37, offset: 2202},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 77, col: 44, offset: 2209},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 77, col: 44, offset: 2209},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 77, col: 50, offset: 2215},
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
			pos:  position{line: 144, col: 1, offset: 4607},
			expr: &actionExpr{
				pos: position{line: 144, col: 15, offset: 4623},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 144, col: 15, offset: 4623,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 148, col: 1, offset: 4681},
			expr: &actionExpr{
				pos: position{line: 148, col: 13, offset: 4695},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 148, col: 13, offset: 4695},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 148, col: 13, offset: 4695},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 148, col: 20, offset: 4702},
								expr: &seqExpr{
									pos: position{line: 148, col: 21, offset: 4703},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 148, col: 21, offset: 4703},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 148, col: 31, offset: 4713},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 148, col: 36, offset: 4718},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 148, col: 47, offset: 4729},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 148, col: 47, offset: 4729},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 148, col: 65, offset: 4747},
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
			pos:  position{line: 161, col: 1, offset: 5218},
			expr: &choiceExpr{
				pos: position{line: 161, col: 19, offset: 5238},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 161, col: 19, offset: 5238},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 161, col: 29, offset: 5248},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 161, col: 41, offset: 5260},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 161, col: 49, offset: 5268},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 161, col: 56, offset: 5275},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 161, col: 66, offset: 5285},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 161, col: 75, offset: 5294},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 161, col: 87, offset: 5306},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 161, col: 95, offset: 5314},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 163, col: 1, offset: 5323},
			expr: &actionExpr{
				pos: position{line: 163, col: 11, offset: 5335},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 163, col: 11, offset: 5335},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 163, col: 11, offset: 5335},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 163, col: 21, offset: 5345},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 163, col: 23, offset: 5347},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 163, col: 28, offset: 5352},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 163, col: 36, offset: 5360},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 171, col: 1, offset: 5537},
			expr: &actionExpr{
				pos: position{line: 171, col: 13, offset: 5551},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 171, col: 13, offset: 5551},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 171, col: 13, offset: 5551},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 25, offset: 5563},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 171, col: 27, offset: 5565},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 171, col: 33, offset: 5571},
								expr: &charClassMatcher{
									pos:        position{line: 171, col: 33, offset: 5571},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 43, offset: 5581},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 171, col: 45, offset: 5583},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 171, col: 48, offset: 5586},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 59, offset: 5597},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 178, col: 1, offset: 5722},
			expr: &actionExpr{
				pos: position{line: 178, col: 9, offset: 5732},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 178, col: 9, offset: 5732},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 178, col: 9, offset: 5732},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 178, col: 17, offset: 5740},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 178, col: 19, offset: 5742},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 178, col: 23, offset: 5746},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 178, col: 33, offset: 5756},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 178, col: 35, offset: 5758},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 178, col: 40, offset: 5763},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 178, col: 51, offset: 5774},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 178, col: 53, offset: 5776},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 178, col: 57, offset: 5780},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 178, col: 59, offset: 5782},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 178, col: 65, offset: 5788},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 178, col: 76, offset: 5799},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 186, col: 1, offset: 5931},
			expr: &actionExpr{
				pos: position{line: 186, col: 8, offset: 5940},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 186, col: 8, offset: 5940},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 186, col: 8, offset: 5940},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 186, col: 15, offset: 5947},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 186, col: 17, offset: 5949},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 186, col: 22, offset: 5954},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 186, col: 33, offset: 5965},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 186, col: 36, offset: 5968},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 186, col: 40, offset: 5972},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 186, col: 43, offset: 5975},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 186, col: 50, offset: 5982},
								expr: &seqExpr{
									pos: position{line: 186, col: 51, offset: 5983},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 186, col: 51, offset: 5983},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 186, col: 61, offset: 5993},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 186, col: 66, offset: 5998},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 186, col: 70, offset: 6002},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 209, col: 1, offset: 6603},
			expr: &actionExpr{
				pos: position{line: 209, col: 13, offset: 6617},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 209, col: 13, offset: 6617},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 209, col: 13, offset: 6617},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 209, col: 20, offset: 6624},
								expr: &seqExpr{
									pos: position{line: 209, col: 21, offset: 6625},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 209, col: 21, offset: 6625},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 209, col: 31, offset: 6635},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 209, col: 36, offset: 6640},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 209, col: 41, offset: 6645},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 209, col: 52, offset: 6656},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 209, col: 54, offset: 6658},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 209, col: 60, offset: 6664},
								expr: &seqExpr{
									pos: position{line: 209, col: 61, offset: 6665},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 209, col: 61, offset: 6665},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 209, col: 65, offset: 6669},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 209, col: 67, offset: 6671},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 209, col: 81, offset: 6685},
							expr: &ruleRefExpr{
								pos:  position{line: 209, col: 81, offset: 6685},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 224, col: 1, offset: 7021},
			expr: &actionExpr{
				pos: position{line: 224, col: 11, offset: 7033},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 224, col: 11, offset: 7033},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 224, col: 11, offset: 7033},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 224, col: 21, offset: 7043},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 224, col: 23, offset: 7045},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 224, col: 27, offset: 7049},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 224, col: 37, offset: 7059},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 224, col: 39, offset: 7061},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 224, col: 44, offset: 7066},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 224, col: 55, offset: 7077},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 231, col: 1, offset: 7186},
			expr: &actionExpr{
				pos: position{line: 231, col: 10, offset: 7197},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 231, col: 10, offset: 7197},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 231, col: 10, offset: 7197},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 231, col: 19, offset: 7206},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 231, col: 21, offset: 7208},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 231, col: 24, offset: 7211},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 232, col: 1, offset: 7251},
			expr: &actionExpr{
				pos: position{line: 232, col: 13, offset: 7265},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 232, col: 13, offset: 7265},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 232, col: 13, offset: 7265},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 232, col: 25, offset: 7277},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 232, col: 27, offset: 7279},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 232, col: 30, offset: 7282},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 233, col: 1, offset: 7333},
			expr: &actionExpr{
				pos: position{line: 233, col: 9, offset: 7343},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 233, col: 9, offset: 7343},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 233, col: 9, offset: 7343},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 17, offset: 7351},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 233, col: 19, offset: 7353},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 233, col: 22, offset: 7356},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 234, col: 1, offset: 7403},
			expr: &actionExpr{
				pos: position{line: 234, col: 14, offset: 7418},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 234, col: 14, offset: 7418},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 234, col: 14, offset: 7418},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 234, col: 19, offset: 7423},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 234, col: 30, offset: 7434},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 234, col: 33, offset: 7437},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 234, col: 37, offset: 7441},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 234, col: 40, offset: 7444},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 234, col: 47, offset: 7451},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 234, col: 57, offset: 7461},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 234, col: 61, offset: 7465},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 244, col: 1, offset: 7626},
			expr: &actionExpr{
				pos: position{line: 244, col: 13, offset: 7640},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 244, col: 13, offset: 7640},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 244, col: 20, offset: 7647},
						expr: &seqExpr{
							pos: position{line: 244, col: 21, offset: 7648},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 244, col: 21, offset: 7648},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 244, col: 27, offset: 7654},
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
			pos:  position{line: 253, col: 1, offset: 7835},
			expr: &actionExpr{
				pos: position{line: 253, col: 9, offset: 7845},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 253, col: 9, offset: 7845},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 253, col: 9, offset: 7845},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 253, col: 16, offset: 7852},
								expr: &seqExpr{
									pos: position{line: 253, col: 17, offset: 7853},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 253, col: 17, offset: 7853},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 253, col: 27, offset: 7863},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 253, col: 32, offset: 7868},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 253, col: 35, offset: 7871},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 253, col: 47, offset: 7883},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 253, col: 49, offset: 7885},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 253, col: 53, offset: 7889},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 253, col: 55, offset: 7891},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 253, col: 59, offset: 7895},
								expr: &ruleRefExpr{
									pos:  position{line: 253, col: 59, offset: 7895},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 253, col: 74, offset: 7910},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 253, col: 76, offset: 7912},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 253, col: 80, offset: 7916},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 253, col: 90, offset: 7926},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 253, col: 92, offset: 7928},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 253, col: 97, offset: 7933},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 253, col: 108, offset: 7944},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 253, col: 111, offset: 7947},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 253, col: 115, offset: 7951},
								expr: &seqExpr{
									pos: position{line: 253, col: 116, offset: 7952},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 253, col: 116, offset: 7952},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 253, col: 120, offset: 7956},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 253, col: 122, offset: 7958},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 253, col: 135, offset: 7971},
							expr: &ruleRefExpr{
								pos:  position{line: 253, col: 135, offset: 7971},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 275, col: 1, offset: 8433},
			expr: &actionExpr{
				pos: position{line: 275, col: 17, offset: 8451},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 275, col: 18, offset: 8452},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 275, col: 18, offset: 8452},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 275, col: 31, offset: 8465},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 283, col: 1, offset: 8608},
			expr: &actionExpr{
				pos: position{line: 283, col: 11, offset: 8620},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 283, col: 11, offset: 8620},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 283, col: 11, offset: 8620},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 283, col: 21, offset: 8630},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 283, col: 23, offset: 8632},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 283, col: 28, offset: 8637},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 283, col: 39, offset: 8648},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 283, col: 41, offset: 8650},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 283, col: 49, offset: 8658},
								expr: &seqExpr{
									pos: position{line: 283, col: 50, offset: 8659},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 283, col: 50, offset: 8659},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 283, col: 60, offset: 8669},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 283, col: 63, offset: 8672},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 283, col: 74, offset: 8683},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 283, col: 79, offset: 8688},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 283, col: 82, offset: 8691},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 283, col: 86, offset: 8695},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 283, col: 89, offset: 8698},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 283, col: 97, offset: 8706},
								expr: &seqExpr{
									pos: position{line: 283, col: 98, offset: 8707},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 283, col: 98, offset: 8707},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 283, col: 107, offset: 8716},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 283, col: 113, offset: 8722},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 283, col: 113, offset: 8722},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 283, col: 119, offset: 8728},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 283, col: 138, offset: 8747},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 299, col: 1, offset: 9131},
			expr: &actionExpr{
				pos: position{line: 299, col: 21, offset: 9153},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 299, col: 21, offset: 9153,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 303, col: 1, offset: 9222},
			expr: &actionExpr{
				pos: position{line: 303, col: 12, offset: 9235},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 303, col: 12, offset: 9235},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 303, col: 12, offset: 9235},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 303, col: 19, offset: 9242},
								expr: &seqExpr{
									pos: position{line: 303, col: 20, offset: 9243},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 303, col: 20, offset: 9243},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 30, offset: 9253},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 303, col: 35, offset: 9258},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 303, col: 42, offset: 9265},
								expr: &seqExpr{
									pos: position{line: 303, col: 43, offset: 9266},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 303, col: 43, offset: 9266},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 303, col: 52, offset: 9275},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 303, col: 57, offset: 9280},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 303, col: 61, offset: 9284},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 74, offset: 9297},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 77, offset: 9300},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 303, col: 82, offset: 9305},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 93, offset: 9316},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 303, col: 95, offset: 9318},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 99, offset: 9322},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 102, offset: 9325},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 303, col: 112, offset: 9335},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 303, col: 122, offset: 9345},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 303, col: 126, offset: 9349},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 303, col: 129, offset: 9352},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 303, col: 140, offset: 9363},
								expr: &ruleRefExpr{
									pos:  position{line: 303, col: 140, offset: 9363},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 303, col: 148, offset: 9371},
							expr: &ruleRefExpr{
								pos:  position{line: 303, col: 148, offset: 9371},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 330, col: 1, offset: 9966},
			expr: &actionExpr{
				pos: position{line: 330, col: 16, offset: 9983},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 330, col: 16, offset: 9983},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 330, col: 21, offset: 9988},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 330, col: 21, offset: 9988},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 330, col: 30, offset: 9997},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 337, col: 1, offset: 10119},
			expr: &actionExpr{
				pos: position{line: 337, col: 10, offset: 10130},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 337, col: 10, offset: 10130},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 337, col: 10, offset: 10130},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 337, col: 19, offset: 10139},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 337, col: 22, offset: 10142},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 337, col: 26, offset: 10146},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 337, col: 29, offset: 10149},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 337, col: 40, offset: 10160},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 337, col: 50, offset: 10170},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 341, col: 1, offset: 10206},
			expr: &actionExpr{
				pos: position{line: 341, col: 13, offset: 10220},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 341, col: 13, offset: 10220},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 341, col: 18, offset: 10225},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 341, col: 18, offset: 10225},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 341, col: 29, offset: 10236},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 341, col: 45, offset: 10252},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 348, col: 1, offset: 10377},
			expr: &actionExpr{
				pos: position{line: 348, col: 12, offset: 10390},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 348, col: 13, offset: 10391},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 348, col: 13, offset: 10391},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 348, col: 22, offset: 10400},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 348, col: 31, offset: 10409},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 348, col: 39, offset: 10417},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 348, col: 47, offset: 10425},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 348, col: 55, offset: 10433},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 348, col: 66, offset: 10444},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 348, col: 77, offset: 10455},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 352, col: 1, offset: 10515},
			expr: &actionExpr{
				pos: position{line: 352, col: 17, offset: 10533},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 352, col: 17, offset: 10533},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 352, col: 22, offset: 10538},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 352, col: 22, offset: 10538},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 352, col: 32, offset: 10548},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 352, col: 42, offset: 10558},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 356, col: 1, offset: 10593},
			expr: &actionExpr{
				pos: position{line: 356, col: 11, offset: 10605},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 356, col: 11, offset: 10605},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 356, col: 11, offset: 10605},
							expr: &ruleRefExpr{
								pos:  position{line: 356, col: 11, offset: 10605},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 356, col: 20, offset: 10614},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 27, offset: 10621},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 356, col: 30, offset: 10624},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 356, col: 34, offset: 10628},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 44, offset: 10638},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 356, col: 47, offset: 10641},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 51, offset: 10645},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 356, col: 54, offset: 10648},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 356, col: 60, offset: 10654},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 70, offset: 10664},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 356, col: 73, offset: 10667},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 364, col: 1, offset: 10790},
			expr: &actionExpr{
				pos: position{line: 364, col: 11, offset: 10802},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 364, col: 11, offset: 10802},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 364, col: 11, offset: 10802},
							expr: &ruleRefExpr{
								pos:  position{line: 364, col: 11, offset: 10802},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 364, col: 20, offset: 10811},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 364, col: 27, offset: 10818},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 364, col: 30, offset: 10821},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 364, col: 34, offset: 10825},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 364, col: 44, offset: 10835},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 364, col: 47, offset: 10838},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 371, col: 1, offset: 10929},
			expr: &actionExpr{
				pos: position{line: 371, col: 12, offset: 10942},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 371, col: 12, offset: 10942},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 371, col: 12, offset: 10942},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 371, col: 20, offset: 10950},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 371, col: 23, offset: 10953},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 371, col: 27, offset: 10957},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 371, col: 37, offset: 10967},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 371, col: 40, offset: 10970},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 378, col: 1, offset: 11062},
			expr: &actionExpr{
				pos: position{line: 378, col: 11, offset: 11074},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 378, col: 11, offset: 11074},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 378, col: 11, offset: 11074},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 378, col: 22, offset: 11085},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 378, col: 30, offset: 11093},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 382, col: 1, offset: 11130},
			expr: &choiceExpr{
				pos: position{line: 382, col: 14, offset: 11145},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 382, col: 14, offset: 11145},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 382, col: 24, offset: 11155},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 382, col: 39, offset: 11170},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 382, col: 56, offset: 11187},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 382, col: 70, offset: 11201},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 382, col: 81, offset: 11212},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 382, col: 93, offset: 11224},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 384, col: 1, offset: 11236},
			expr: &actionExpr{
				pos: position{line: 384, col: 16, offset: 11253},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 384, col: 17, offset: 11254},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 384, col: 17, offset: 11254},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 384, col: 26, offset: 11263},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 388, col: 1, offset: 11318},
			expr: &actionExpr{
				pos: position{line: 388, col: 15, offset: 11334},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 388, col: 15, offset: 11334},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 388, col: 15, offset: 11334},
							expr: &charClassMatcher{
								pos:        position{line: 388, col: 15, offset: 11334},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 388, col: 21, offset: 11340},
							expr: &ruleRefExpr{
								pos:  position{line: 388, col: 21, offset: 11340},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 392, col: 1, offset: 11404},
			expr: &actionExpr{
				pos: position{line: 392, col: 18, offset: 11423},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 392, col: 18, offset: 11423},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 392, col: 18, offset: 11423},
							expr: &charClassMatcher{
								pos:        position{line: 392, col: 18, offset: 11423},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 392, col: 24, offset: 11429},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 24, offset: 11429},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 392, col: 31, offset: 11436},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 392, col: 35, offset: 11440},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 35, offset: 11440},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 392, col: 42, offset: 11447},
							expr: &seqExpr{
								pos: position{line: 392, col: 44, offset: 11449},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 392, col: 44, offset: 11449},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 392, col: 51, offset: 11456},
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
			pos:  position{line: 396, col: 1, offset: 11526},
			expr: &actionExpr{
				pos: position{line: 396, col: 13, offset: 11540},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 396, col: 13, offset: 11540},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 396, col: 13, offset: 11540},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 396, col: 17, offset: 11544},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 396, col: 20, offset: 11547},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 396, col: 27, offset: 11554},
								expr: &seqExpr{
									pos: position{line: 396, col: 28, offset: 11555},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 396, col: 28, offset: 11555},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 39, offset: 11566},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 396, col: 42, offset: 11569},
											expr: &ruleRefExpr{
												pos:  position{line: 396, col: 42, offset: 11569},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 57, offset: 11584},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 396, col: 62, offset: 11589},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 396, col: 65, offset: 11592},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 405, col: 1, offset: 11786},
			expr: &actionExpr{
				pos: position{line: 405, col: 12, offset: 11799},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 405, col: 12, offset: 11799},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 405, col: 12, offset: 11799},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 405, col: 16, offset: 11803},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 405, col: 19, offset: 11806},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 405, col: 26, offset: 11813},
								expr: &seqExpr{
									pos: position{line: 405, col: 27, offset: 11814},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 405, col: 27, offset: 11814},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 38, offset: 11825},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 405, col: 41, offset: 11828},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 45, offset: 11832},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 48, offset: 11835},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 59, offset: 11846},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 405, col: 63, offset: 11850},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 405, col: 63, offset: 11850},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 405, col: 69, offset: 11856},
													expr: &litMatcher{
														pos:        position{line: 405, col: 70, offset: 11857},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 405, col: 75, offset: 11862},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 405, col: 80, offset: 11867},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 425, col: 1, offset: 12417},
			expr: &ruleRefExpr{
				pos:  position{line: 425, col: 19, offset: 12437},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 427, col: 1, offset: 12444},
			expr: &actionExpr{
				pos: position{line: 427, col: 9, offset: 12454},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 427, col: 9, offset: 12454},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 427, col: 9, offset: 12454},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 427, col: 16, offset: 12461},
								expr: &seqExpr{
									pos: position{line: 427, col: 17, offset: 12462},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 427, col: 17, offset: 12462},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 427, col: 27, offset: 12472},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 427, col: 32, offset: 12477},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 40, offset: 12485},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 43, offset: 12488},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 427, col: 48, offset: 12493},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 59, offset: 12504},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 62, offset: 12507},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 427, col: 69, offset: 12514},
								expr: &ruleRefExpr{
									pos:  position{line: 427, col: 69, offset: 12514},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 77, offset: 12522},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 427, col: 80, offset: 12525},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 84, offset: 12529},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 427, col: 87, offset: 12532},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 427, col: 98, offset: 12543},
								expr: &seqExpr{
									pos: position{line: 427, col: 99, offset: 12544},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 427, col: 99, offset: 12544},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 427, col: 109, offset: 12554},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 427, col: 115, offset: 12560},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 427, col: 115, offset: 12560},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 427, col: 121, offset: 12566},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 427, col: 138, offset: 12583},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 448, col: 1, offset: 13128},
			expr: &actionExpr{
				pos: position{line: 448, col: 19, offset: 13148},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 448, col: 19, offset: 13148,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 452, col: 1, offset: 13215},
			expr: &actionExpr{
				pos: position{line: 452, col: 10, offset: 13226},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 452, col: 10, offset: 13226},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 452, col: 10, offset: 13226},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 452, col: 19, offset: 13235},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 452, col: 22, offset: 13238},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 452, col: 34, offset: 13250},
							expr: &seqExpr{
								pos: position{line: 452, col: 35, offset: 13251},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 452, col: 35, offset: 13251},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 452, col: 39, offset: 13255},
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
			pos:  position{line: 457, col: 1, offset: 13386},
			expr: &choiceExpr{
				pos: position{line: 457, col: 15, offset: 13402},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 457, col: 16, offset: 13403},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 457, col: 16, offset: 13403},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 457, col: 20, offset: 13407},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 457, col: 31, offset: 13418},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 457, col: 38, offset: 13425},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 459, col: 1, offset: 13437},
			expr: &oneOrMoreExpr{
				pos: position{line: 459, col: 14, offset: 13452},
				expr: &charClassMatcher{
					pos:        position{line: 459, col: 14, offset: 13452},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 461, col: 1, offset: 13470},
			expr: &actionExpr{
				pos: position{line: 461, col: 13, offset: 13484},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 461, col: 13, offset: 13484},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 461, col: 13, offset: 13484},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 461, col: 20, offset: 13491},
								expr: &seqExpr{
									pos: position{line: 461, col: 21, offset: 13492},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 461, col: 21, offset: 13492},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 31, offset: 13502},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 461, col: 36, offset: 13507},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 461, col: 41, offset: 13512},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 52, offset: 13523},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 461, col: 54, offset: 13525},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 58, offset: 13529},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 61, offset: 13532},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 461, col: 65, offset: 13536},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 76, offset: 13547},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 477, col: 1, offset: 14058},
			expr: &actionExpr{
				pos: position{line: 477, col: 11, offset: 14070},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 477, col: 12, offset: 14071},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 477, col: 13, offset: 14072},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 477, col: 13, offset: 14072},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 477, col: 17, offset: 14076},
									expr: &choiceExpr{
										pos: position{line: 477, col: 18, offset: 14077},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 477, col: 18, offset: 14077},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 477, col: 25, offset: 14084},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 477, col: 32, offset: 14091},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 477, col: 40, offset: 14099},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 477, col: 40, offset: 14099},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 477, col: 45, offset: 14104},
									expr: &choiceExpr{
										pos: position{line: 477, col: 46, offset: 14105},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 477, col: 46, offset: 14105},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 477, col: 53, offset: 14112},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 477, col: 60, offset: 14119},
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
			pos:  position{line: 486, col: 1, offset: 14405},
			expr: &actionExpr{
				pos: position{line: 486, col: 14, offset: 14420},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 486, col: 14, offset: 14420},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 486, col: 14, offset: 14420},
							expr: &choiceExpr{
								pos: position{line: 486, col: 15, offset: 14421},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 486, col: 15, offset: 14421},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 486, col: 24, offset: 14430},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 486, col: 30, offset: 14436},
							expr: &choiceExpr{
								pos: position{line: 486, col: 31, offset: 14437},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 486, col: 31, offset: 14437},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 486, col: 40, offset: 14446},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 486, col: 48, offset: 14454},
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
			pos:  position{line: 490, col: 1, offset: 14509},
			expr: &charClassMatcher{
				pos:        position{line: 490, col: 17, offset: 14527},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 491, col: 1, offset: 14532},
			expr: &charClassMatcher{
				pos:        position{line: 491, col: 10, offset: 14543},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 492, col: 1, offset: 14552},
			expr: &charClassMatcher{
				pos:        position{line: 492, col: 9, offset: 14562},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 494, col: 1, offset: 14569},
			expr: &anyMatcher{
				line: 494, col: 14, offset: 14584,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 495, col: 1, offset: 14586},
			expr: &actionExpr{
				pos: position{line: 495, col: 13, offset: 14600},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 495, col: 13, offset: 14600},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 495, col: 13, offset: 14600},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 495, col: 20, offset: 14607},
							expr: &seqExpr{
								pos: position{line: 495, col: 22, offset: 14609},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 495, col: 22, offset: 14609},
										expr: &litMatcher{
											pos:        position{line: 495, col: 23, offset: 14610},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 495, col: 28, offset: 14615},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 495, col: 42, offset: 14629},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 501, col: 1, offset: 14809},
			expr: &choiceExpr{
				pos: position{line: 501, col: 11, offset: 14821},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 501, col: 11, offset: 14821},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 501, col: 30, offset: 14840},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 502, col: 1, offset: 14858},
			expr: &seqExpr{
				pos: position{line: 502, col: 20, offset: 14879},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 502, col: 20, offset: 14879},
						expr: &ruleRefExpr{
							pos:  position{line: 502, col: 21, offset: 14880},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 502, col: 31, offset: 14890},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 502, col: 36, offset: 14895},
						expr: &seqExpr{
							pos: position{line: 502, col: 38, offset: 14897},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 502, col: 38, offset: 14897},
									expr: &litMatcher{
										pos:        position{line: 502, col: 39, offset: 14898},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 502, col: 44, offset: 14903},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 502, col: 58, offset: 14917},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 503, col: 1, offset: 14922},
			expr: &seqExpr{
				pos: position{line: 503, col: 36, offset: 14959},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 503, col: 36, offset: 14959},
						expr: &ruleRefExpr{
							pos:  position{line: 503, col: 37, offset: 14960},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 503, col: 47, offset: 14970},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 503, col: 52, offset: 14975},
						expr: &seqExpr{
							pos: position{line: 503, col: 54, offset: 14977},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 503, col: 54, offset: 14977},
									expr: &choiceExpr{
										pos: position{line: 503, col: 57, offset: 14980},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 503, col: 57, offset: 14980},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 503, col: 64, offset: 14987},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 503, col: 70, offset: 14993},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 503, col: 84, offset: 15007},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 504, col: 1, offset: 15012},
			expr: &choiceExpr{
				pos: position{line: 504, col: 21, offset: 15034},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 504, col: 22, offset: 15035},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 504, col: 22, offset: 15035},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 504, col: 27, offset: 15040},
								expr: &seqExpr{
									pos: position{line: 504, col: 29, offset: 15042},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 504, col: 29, offset: 15042},
											expr: &ruleRefExpr{
												pos:  position{line: 504, col: 30, offset: 15043},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 504, col: 34, offset: 15047},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 504, col: 52, offset: 15065},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 504, col: 52, offset: 15065},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 504, col: 56, offset: 15069},
								expr: &seqExpr{
									pos: position{line: 504, col: 58, offset: 15071},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 504, col: 58, offset: 15071},
											expr: &ruleRefExpr{
												pos:  position{line: 504, col: 59, offset: 15072},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 504, col: 63, offset: 15076},
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
			pos:  position{line: 506, col: 1, offset: 15092},
			expr: &zeroOrMoreExpr{
				pos: position{line: 506, col: 6, offset: 15099},
				expr: &choiceExpr{
					pos: position{line: 506, col: 8, offset: 15101},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 506, col: 8, offset: 15101},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 506, col: 21, offset: 15114},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 506, col: 27, offset: 15120},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 507, col: 1, offset: 15131},
			expr: &zeroOrMoreExpr{
				pos: position{line: 507, col: 5, offset: 15137},
				expr: &choiceExpr{
					pos: position{line: 507, col: 7, offset: 15139},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 507, col: 7, offset: 15139},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 507, col: 20, offset: 15152},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 508, col: 1, offset: 15188},
			expr: &zeroOrMoreExpr{
				pos: position{line: 508, col: 6, offset: 15195},
				expr: &ruleRefExpr{
					pos:  position{line: 508, col: 6, offset: 15195},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 510, col: 1, offset: 15208},
			expr: &charClassMatcher{
				pos:        position{line: 510, col: 14, offset: 15223},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 511, col: 1, offset: 15231},
			expr: &litMatcher{
				pos:        position{line: 511, col: 7, offset: 15239},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 512, col: 1, offset: 15244},
			expr: &choiceExpr{
				pos: position{line: 512, col: 7, offset: 15252},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 512, col: 7, offset: 15252},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 512, col: 7, offset: 15252},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 512, col: 10, offset: 15255},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 512, col: 16, offset: 15261},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 512, col: 16, offset: 15261},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 512, col: 18, offset: 15263},
								expr: &ruleRefExpr{
									pos:  position{line: 512, col: 18, offset: 15263},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 512, col: 37, offset: 15282},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 512, col: 43, offset: 15288},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 512, col: 43, offset: 15288},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 512, col: 46, offset: 15291},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 514, col: 1, offset: 15296},
			expr: &notExpr{
				pos: position{line: 514, col: 7, offset: 15304},
				expr: &anyMatcher{
					line: 514, col: 8, offset: 15305,
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
		Values: make([]*EnumValue, len(vs)),
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

func (c *current) onField1(docstr, id, mod, typ, name, def interface{}) (interface{}, error) {
	f := &Field{
		ID:   int(id.(int64)),
		Name: string(name.(Identifier)),
		Type: typ.(*Type),
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
	return p.cur.onField1(stack["docstr"], stack["id"], stack["mod"], stack["typ"], stack["name"], stack["def"])
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
			e.Modifier = Optional
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

func (c *current) onPrefix1() (interface{}, error) {
	prefix := strings.TrimSpace(strings.TrimPrefix(string(c.text), "prefix"))
	return newScopePrefix(prefix)
}

func (p *parser) callonPrefix1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPrefix1()
}

func (c *current) onOperation1(docstr, name, typ interface{}) (interface{}, error) {
	o := &Operation{
		Name: string(name.(Identifier)),
		Type: &Type{Name: string(typ.(Identifier))},
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
	return p.cur.onOperation1(stack["docstr"], stack["name"], stack["typ"])
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
