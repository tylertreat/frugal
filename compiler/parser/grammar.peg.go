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
			pos:  position{line: 147, col: 1, offset: 4665},
			expr: &actionExpr{
				pos: position{line: 147, col: 15, offset: 4681},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 147, col: 15, offset: 4681,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 151, col: 1, offset: 4739},
			expr: &actionExpr{
				pos: position{line: 151, col: 13, offset: 4753},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 151, col: 13, offset: 4753},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 151, col: 13, offset: 4753},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 151, col: 20, offset: 4760},
								expr: &seqExpr{
									pos: position{line: 151, col: 21, offset: 4761},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 151, col: 21, offset: 4761},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 151, col: 31, offset: 4771},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 151, col: 36, offset: 4776},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 151, col: 47, offset: 4787},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 151, col: 47, offset: 4787},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 151, col: 65, offset: 4805},
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
			pos:  position{line: 164, col: 1, offset: 5276},
			expr: &choiceExpr{
				pos: position{line: 164, col: 19, offset: 5296},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 164, col: 19, offset: 5296},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 29, offset: 5306},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 41, offset: 5318},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 49, offset: 5326},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 56, offset: 5333},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 66, offset: 5343},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 75, offset: 5352},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 87, offset: 5364},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 95, offset: 5372},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 166, col: 1, offset: 5381},
			expr: &actionExpr{
				pos: position{line: 166, col: 11, offset: 5393},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 166, col: 11, offset: 5393},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 166, col: 11, offset: 5393},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 166, col: 21, offset: 5403},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 166, col: 23, offset: 5405},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 166, col: 28, offset: 5410},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 166, col: 36, offset: 5418},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 170, col: 1, offset: 5466},
			expr: &actionExpr{
				pos: position{line: 170, col: 13, offset: 5480},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 170, col: 13, offset: 5480},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 170, col: 13, offset: 5480},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 25, offset: 5492},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 27, offset: 5494},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 170, col: 33, offset: 5500},
								expr: &charClassMatcher{
									pos:        position{line: 170, col: 33, offset: 5500},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 42, offset: 5509},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 44, offset: 5511},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 170, col: 47, offset: 5514},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 58, offset: 5525},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 177, col: 1, offset: 5650},
			expr: &actionExpr{
				pos: position{line: 177, col: 9, offset: 5660},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 177, col: 9, offset: 5660},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 177, col: 9, offset: 5660},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 17, offset: 5668},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 19, offset: 5670},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 23, offset: 5674},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 33, offset: 5684},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 35, offset: 5686},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 40, offset: 5691},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 51, offset: 5702},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 177, col: 53, offset: 5704},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 57, offset: 5708},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 59, offset: 5710},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 65, offset: 5716},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 76, offset: 5727},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 185, col: 1, offset: 5859},
			expr: &actionExpr{
				pos: position{line: 185, col: 8, offset: 5868},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 185, col: 8, offset: 5868},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 185, col: 8, offset: 5868},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 15, offset: 5875},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 185, col: 17, offset: 5877},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 185, col: 22, offset: 5882},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 33, offset: 5893},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 185, col: 36, offset: 5896},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 40, offset: 5900},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 185, col: 43, offset: 5903},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 185, col: 50, offset: 5910},
								expr: &seqExpr{
									pos: position{line: 185, col: 51, offset: 5911},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 185, col: 51, offset: 5911},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 185, col: 61, offset: 5921},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 185, col: 66, offset: 5926},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 70, offset: 5930},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 208, col: 1, offset: 6542},
			expr: &actionExpr{
				pos: position{line: 208, col: 13, offset: 6556},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 208, col: 13, offset: 6556},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 208, col: 13, offset: 6556},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 208, col: 20, offset: 6563},
								expr: &seqExpr{
									pos: position{line: 208, col: 21, offset: 6564},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 208, col: 21, offset: 6564},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 31, offset: 6574},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 208, col: 36, offset: 6579},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 208, col: 41, offset: 6584},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 208, col: 52, offset: 6595},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 208, col: 54, offset: 6597},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 208, col: 60, offset: 6603},
								expr: &seqExpr{
									pos: position{line: 208, col: 61, offset: 6604},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 208, col: 61, offset: 6604},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 65, offset: 6608},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 67, offset: 6610},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 208, col: 81, offset: 6624},
							expr: &ruleRefExpr{
								pos:  position{line: 208, col: 81, offset: 6624},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 223, col: 1, offset: 6960},
			expr: &actionExpr{
				pos: position{line: 223, col: 11, offset: 6972},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 223, col: 11, offset: 6972},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 223, col: 11, offset: 6972},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 223, col: 21, offset: 6982},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 223, col: 23, offset: 6984},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 223, col: 27, offset: 6988},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 223, col: 37, offset: 6998},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 223, col: 39, offset: 7000},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 223, col: 44, offset: 7005},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 223, col: 55, offset: 7016},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 230, col: 1, offset: 7125},
			expr: &actionExpr{
				pos: position{line: 230, col: 10, offset: 7136},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 230, col: 10, offset: 7136},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 230, col: 10, offset: 7136},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 230, col: 19, offset: 7145},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 230, col: 21, offset: 7147},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 230, col: 24, offset: 7150},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 231, col: 1, offset: 7190},
			expr: &actionExpr{
				pos: position{line: 231, col: 13, offset: 7204},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 231, col: 13, offset: 7204},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 231, col: 13, offset: 7204},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 231, col: 25, offset: 7216},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 231, col: 27, offset: 7218},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 231, col: 30, offset: 7221},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 232, col: 1, offset: 7272},
			expr: &actionExpr{
				pos: position{line: 232, col: 9, offset: 7282},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 232, col: 9, offset: 7282},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 232, col: 9, offset: 7282},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 232, col: 17, offset: 7290},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 232, col: 19, offset: 7292},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 232, col: 22, offset: 7295},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 233, col: 1, offset: 7342},
			expr: &actionExpr{
				pos: position{line: 233, col: 14, offset: 7357},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 233, col: 14, offset: 7357},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 233, col: 14, offset: 7357},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 233, col: 19, offset: 7362},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 30, offset: 7373},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 233, col: 33, offset: 7376},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 37, offset: 7380},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 233, col: 40, offset: 7383},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 233, col: 47, offset: 7390},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 233, col: 57, offset: 7400},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 61, offset: 7404},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 243, col: 1, offset: 7565},
			expr: &actionExpr{
				pos: position{line: 243, col: 13, offset: 7579},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 243, col: 13, offset: 7579},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 243, col: 20, offset: 7586},
						expr: &seqExpr{
							pos: position{line: 243, col: 21, offset: 7587},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 243, col: 21, offset: 7587},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 243, col: 27, offset: 7593},
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
			pos:  position{line: 252, col: 1, offset: 7774},
			expr: &actionExpr{
				pos: position{line: 252, col: 9, offset: 7784},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 252, col: 9, offset: 7784},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 252, col: 9, offset: 7784},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 16, offset: 7791},
								expr: &seqExpr{
									pos: position{line: 252, col: 17, offset: 7792},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 252, col: 17, offset: 7792},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 252, col: 27, offset: 7802},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 252, col: 32, offset: 7807},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 35, offset: 7810},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 47, offset: 7822},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 252, col: 49, offset: 7824},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 53, offset: 7828},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 55, offset: 7830},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 59, offset: 7834},
								expr: &ruleRefExpr{
									pos:  position{line: 252, col: 59, offset: 7834},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 69, offset: 7844},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 71, offset: 7846},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 75, offset: 7850},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 85, offset: 7860},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 87, offset: 7862},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 92, offset: 7867},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 103, offset: 7878},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 106, offset: 7881},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 110, offset: 7885},
								expr: &seqExpr{
									pos: position{line: 252, col: 111, offset: 7886},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 252, col: 111, offset: 7886},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 252, col: 115, offset: 7890},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 252, col: 117, offset: 7892},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 252, col: 130, offset: 7905},
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 130, offset: 7905},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 271, col: 1, offset: 8324},
			expr: &actionExpr{
				pos: position{line: 271, col: 12, offset: 8337},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 271, col: 13, offset: 8338},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 271, col: 13, offset: 8338},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 271, col: 26, offset: 8351},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 275, col: 1, offset: 8425},
			expr: &actionExpr{
				pos: position{line: 275, col: 11, offset: 8437},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 275, col: 11, offset: 8437},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 275, col: 11, offset: 8437},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 275, col: 21, offset: 8447},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 275, col: 23, offset: 8449},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 275, col: 28, offset: 8454},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 275, col: 39, offset: 8465},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 275, col: 41, offset: 8467},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 275, col: 49, offset: 8475},
								expr: &seqExpr{
									pos: position{line: 275, col: 50, offset: 8476},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 275, col: 50, offset: 8476},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 275, col: 60, offset: 8486},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 275, col: 63, offset: 8489},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 275, col: 74, offset: 8500},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 275, col: 79, offset: 8505},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 275, col: 82, offset: 8508},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 275, col: 86, offset: 8512},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 275, col: 89, offset: 8515},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 275, col: 97, offset: 8523},
								expr: &seqExpr{
									pos: position{line: 275, col: 98, offset: 8524},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 275, col: 98, offset: 8524},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 275, col: 107, offset: 8533},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 275, col: 113, offset: 8539},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 275, col: 113, offset: 8539},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 275, col: 119, offset: 8545},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 275, col: 138, offset: 8564},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 290, col: 1, offset: 8959},
			expr: &actionExpr{
				pos: position{line: 290, col: 21, offset: 8981},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 290, col: 21, offset: 8981,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 294, col: 1, offset: 9050},
			expr: &actionExpr{
				pos: position{line: 294, col: 12, offset: 9063},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 294, col: 12, offset: 9063},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 294, col: 12, offset: 9063},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 294, col: 19, offset: 9070},
								expr: &seqExpr{
									pos: position{line: 294, col: 20, offset: 9071},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 294, col: 20, offset: 9071},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 294, col: 30, offset: 9081},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 294, col: 35, offset: 9086},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 294, col: 42, offset: 9093},
								expr: &seqExpr{
									pos: position{line: 294, col: 43, offset: 9094},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 294, col: 43, offset: 9094},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 294, col: 52, offset: 9103},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 294, col: 57, offset: 9108},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 294, col: 61, offset: 9112},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 294, col: 74, offset: 9125},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 294, col: 77, offset: 9128},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 294, col: 82, offset: 9133},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 294, col: 93, offset: 9144},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 294, col: 95, offset: 9146},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 294, col: 99, offset: 9150},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 294, col: 102, offset: 9153},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 294, col: 112, offset: 9163},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 294, col: 122, offset: 9173},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 294, col: 126, offset: 9177},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 294, col: 129, offset: 9180},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 294, col: 140, offset: 9191},
								expr: &ruleRefExpr{
									pos:  position{line: 294, col: 140, offset: 9191},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 294, col: 148, offset: 9199},
							expr: &ruleRefExpr{
								pos:  position{line: 294, col: 148, offset: 9199},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 321, col: 1, offset: 9790},
			expr: &actionExpr{
				pos: position{line: 321, col: 16, offset: 9807},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 321, col: 16, offset: 9807},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 321, col: 21, offset: 9812},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 321, col: 21, offset: 9812},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 321, col: 30, offset: 9821},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 328, col: 1, offset: 9943},
			expr: &actionExpr{
				pos: position{line: 328, col: 10, offset: 9954},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 328, col: 10, offset: 9954},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 328, col: 10, offset: 9954},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 19, offset: 9963},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 328, col: 22, offset: 9966},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 328, col: 26, offset: 9970},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 328, col: 29, offset: 9973},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 328, col: 40, offset: 9984},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 328, col: 50, offset: 9994},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 332, col: 1, offset: 10030},
			expr: &actionExpr{
				pos: position{line: 332, col: 13, offset: 10044},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 332, col: 13, offset: 10044},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 332, col: 18, offset: 10049},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 332, col: 18, offset: 10049},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 332, col: 29, offset: 10060},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 332, col: 45, offset: 10076},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "DefinitionType",
			pos:  position{line: 339, col: 1, offset: 10201},
			expr: &actionExpr{
				pos: position{line: 339, col: 18, offset: 10220},
				run: (*parser).callonDefinitionType1,
				expr: &labeledExpr{
					pos:   position{line: 339, col: 18, offset: 10220},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 339, col: 23, offset: 10225},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 339, col: 23, offset: 10225},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 339, col: 34, offset: 10236},
								name: "ContainerType",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 343, col: 1, offset: 10276},
			expr: &actionExpr{
				pos: position{line: 343, col: 12, offset: 10289},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 343, col: 13, offset: 10290},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 343, col: 13, offset: 10290},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 343, col: 22, offset: 10299},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 343, col: 31, offset: 10308},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 343, col: 39, offset: 10316},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 343, col: 47, offset: 10324},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 343, col: 55, offset: 10332},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 343, col: 66, offset: 10343},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 343, col: 77, offset: 10354},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 347, col: 1, offset: 10414},
			expr: &actionExpr{
				pos: position{line: 347, col: 17, offset: 10432},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 347, col: 17, offset: 10432},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 347, col: 22, offset: 10437},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 347, col: 22, offset: 10437},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 347, col: 32, offset: 10447},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 347, col: 42, offset: 10457},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 351, col: 1, offset: 10492},
			expr: &actionExpr{
				pos: position{line: 351, col: 11, offset: 10504},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 351, col: 11, offset: 10504},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 351, col: 11, offset: 10504},
							expr: &ruleRefExpr{
								pos:  position{line: 351, col: 11, offset: 10504},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 351, col: 20, offset: 10513},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 351, col: 27, offset: 10520},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 351, col: 30, offset: 10523},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 351, col: 34, offset: 10527},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 351, col: 44, offset: 10537},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 351, col: 47, offset: 10540},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 351, col: 51, offset: 10544},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 351, col: 54, offset: 10547},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 351, col: 60, offset: 10553},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 351, col: 70, offset: 10563},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 351, col: 73, offset: 10566},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 359, col: 1, offset: 10689},
			expr: &actionExpr{
				pos: position{line: 359, col: 11, offset: 10701},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 359, col: 11, offset: 10701},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 359, col: 11, offset: 10701},
							expr: &ruleRefExpr{
								pos:  position{line: 359, col: 11, offset: 10701},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 359, col: 20, offset: 10710},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 359, col: 27, offset: 10717},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 359, col: 30, offset: 10720},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 359, col: 34, offset: 10724},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 359, col: 44, offset: 10734},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 359, col: 47, offset: 10737},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 366, col: 1, offset: 10828},
			expr: &actionExpr{
				pos: position{line: 366, col: 12, offset: 10841},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 366, col: 12, offset: 10841},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 366, col: 12, offset: 10841},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 366, col: 20, offset: 10849},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 366, col: 23, offset: 10852},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 366, col: 27, offset: 10856},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 366, col: 37, offset: 10866},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 366, col: 40, offset: 10869},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 373, col: 1, offset: 10961},
			expr: &actionExpr{
				pos: position{line: 373, col: 11, offset: 10973},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 373, col: 11, offset: 10973},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 373, col: 11, offset: 10973},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 373, col: 22, offset: 10984},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 373, col: 30, offset: 10992},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 377, col: 1, offset: 11029},
			expr: &choiceExpr{
				pos: position{line: 377, col: 14, offset: 11044},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 377, col: 14, offset: 11044},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 377, col: 24, offset: 11054},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 377, col: 41, offset: 11071},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 377, col: 55, offset: 11085},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 377, col: 66, offset: 11096},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 377, col: 78, offset: 11108},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 379, col: 1, offset: 11120},
			expr: &actionExpr{
				pos: position{line: 379, col: 15, offset: 11136},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 379, col: 15, offset: 11136},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 379, col: 15, offset: 11136},
							expr: &charClassMatcher{
								pos:        position{line: 379, col: 15, offset: 11136},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 379, col: 21, offset: 11142},
							expr: &ruleRefExpr{
								pos:  position{line: 379, col: 21, offset: 11142},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 383, col: 1, offset: 11206},
			expr: &actionExpr{
				pos: position{line: 383, col: 18, offset: 11225},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 383, col: 18, offset: 11225},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 383, col: 18, offset: 11225},
							expr: &charClassMatcher{
								pos:        position{line: 383, col: 18, offset: 11225},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 383, col: 24, offset: 11231},
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 24, offset: 11231},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 383, col: 31, offset: 11238},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 383, col: 35, offset: 11242},
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 35, offset: 11242},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 383, col: 42, offset: 11249},
							expr: &seqExpr{
								pos: position{line: 383, col: 44, offset: 11251},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 383, col: 44, offset: 11251},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 383, col: 51, offset: 11258},
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
			pos:  position{line: 387, col: 1, offset: 11328},
			expr: &actionExpr{
				pos: position{line: 387, col: 13, offset: 11342},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 387, col: 13, offset: 11342},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 387, col: 13, offset: 11342},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 387, col: 17, offset: 11346},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 387, col: 20, offset: 11349},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 387, col: 27, offset: 11356},
								expr: &seqExpr{
									pos: position{line: 387, col: 28, offset: 11357},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 387, col: 28, offset: 11357},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 387, col: 39, offset: 11368},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 387, col: 42, offset: 11371},
											expr: &ruleRefExpr{
												pos:  position{line: 387, col: 42, offset: 11371},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 387, col: 57, offset: 11386},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 387, col: 62, offset: 11391},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 387, col: 65, offset: 11394},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 396, col: 1, offset: 11588},
			expr: &actionExpr{
				pos: position{line: 396, col: 12, offset: 11601},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 396, col: 12, offset: 11601},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 396, col: 12, offset: 11601},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 396, col: 16, offset: 11605},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 396, col: 19, offset: 11608},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 396, col: 26, offset: 11615},
								expr: &seqExpr{
									pos: position{line: 396, col: 27, offset: 11616},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 396, col: 27, offset: 11616},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 38, offset: 11627},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 396, col: 41, offset: 11630},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 45, offset: 11634},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 48, offset: 11637},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 59, offset: 11648},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 396, col: 63, offset: 11652},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 396, col: 63, offset: 11652},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 396, col: 69, offset: 11658},
													expr: &litMatcher{
														pos:        position{line: 396, col: 70, offset: 11659},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 396, col: 75, offset: 11664},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 396, col: 80, offset: 11669},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 416, col: 1, offset: 12219},
			expr: &ruleRefExpr{
				pos:  position{line: 416, col: 19, offset: 12239},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 418, col: 1, offset: 12246},
			expr: &actionExpr{
				pos: position{line: 418, col: 9, offset: 12256},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 418, col: 9, offset: 12256},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 418, col: 9, offset: 12256},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 418, col: 16, offset: 12263},
								expr: &seqExpr{
									pos: position{line: 418, col: 17, offset: 12264},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 418, col: 17, offset: 12264},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 418, col: 27, offset: 12274},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 418, col: 32, offset: 12279},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 418, col: 40, offset: 12287},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 418, col: 43, offset: 12290},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 418, col: 48, offset: 12295},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 418, col: 59, offset: 12306},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 418, col: 62, offset: 12309},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 418, col: 66, offset: 12313},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 418, col: 69, offset: 12316},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 418, col: 76, offset: 12323},
								expr: &ruleRefExpr{
									pos:  position{line: 418, col: 76, offset: 12323},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 418, col: 84, offset: 12331},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 418, col: 87, offset: 12334},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 418, col: 98, offset: 12345},
								expr: &seqExpr{
									pos: position{line: 418, col: 99, offset: 12346},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 418, col: 99, offset: 12346},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 418, col: 109, offset: 12356},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 418, col: 115, offset: 12362},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 418, col: 115, offset: 12362},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 418, col: 121, offset: 12368},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 418, col: 138, offset: 12385},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 439, col: 1, offset: 12930},
			expr: &actionExpr{
				pos: position{line: 439, col: 19, offset: 12950},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 439, col: 19, offset: 12950,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 443, col: 1, offset: 13017},
			expr: &actionExpr{
				pos: position{line: 443, col: 10, offset: 13028},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 443, col: 10, offset: 13028},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 443, col: 10, offset: 13028},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 443, col: 19, offset: 13037},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 443, col: 21, offset: 13039},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 26, offset: 13044},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 443, col: 34, offset: 13052},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 447, col: 1, offset: 13102},
			expr: &actionExpr{
				pos: position{line: 447, col: 13, offset: 13116},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 447, col: 13, offset: 13116},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 447, col: 13, offset: 13116},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 447, col: 20, offset: 13123},
								expr: &seqExpr{
									pos: position{line: 447, col: 21, offset: 13124},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 447, col: 21, offset: 13124},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 447, col: 31, offset: 13134},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 447, col: 36, offset: 13139},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 447, col: 41, offset: 13144},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 52, offset: 13155},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 447, col: 54, offset: 13157},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 58, offset: 13161},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 447, col: 61, offset: 13164},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 447, col: 67, offset: 13170},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 78, offset: 13181},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 463, col: 1, offset: 13683},
			expr: &actionExpr{
				pos: position{line: 463, col: 11, offset: 13695},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 463, col: 12, offset: 13696},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 463, col: 13, offset: 13697},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 463, col: 13, offset: 13697},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 463, col: 17, offset: 13701},
									expr: &choiceExpr{
										pos: position{line: 463, col: 18, offset: 13702},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 463, col: 18, offset: 13702},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 463, col: 25, offset: 13709},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 463, col: 32, offset: 13716},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 463, col: 40, offset: 13724},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 463, col: 40, offset: 13724},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 463, col: 45, offset: 13729},
									expr: &choiceExpr{
										pos: position{line: 463, col: 46, offset: 13730},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 463, col: 46, offset: 13730},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 463, col: 53, offset: 13737},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 463, col: 60, offset: 13744},
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
			pos:  position{line: 470, col: 1, offset: 13960},
			expr: &actionExpr{
				pos: position{line: 470, col: 14, offset: 13975},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 470, col: 14, offset: 13975},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 470, col: 14, offset: 13975},
							expr: &choiceExpr{
								pos: position{line: 470, col: 15, offset: 13976},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 470, col: 15, offset: 13976},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 470, col: 24, offset: 13985},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 470, col: 30, offset: 13991},
							expr: &choiceExpr{
								pos: position{line: 470, col: 31, offset: 13992},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 470, col: 31, offset: 13992},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 470, col: 40, offset: 14001},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 470, col: 48, offset: 14009},
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
			pos:  position{line: 474, col: 1, offset: 14064},
			expr: &charClassMatcher{
				pos:        position{line: 474, col: 17, offset: 14082},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 475, col: 1, offset: 14087},
			expr: &charClassMatcher{
				pos:        position{line: 475, col: 10, offset: 14098},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 476, col: 1, offset: 14107},
			expr: &charClassMatcher{
				pos:        position{line: 476, col: 9, offset: 14117},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 478, col: 1, offset: 14124},
			expr: &anyMatcher{
				line: 478, col: 14, offset: 14139,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 479, col: 1, offset: 14141},
			expr: &actionExpr{
				pos: position{line: 479, col: 13, offset: 14155},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 479, col: 13, offset: 14155},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 479, col: 13, offset: 14155},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 479, col: 20, offset: 14162},
							expr: &seqExpr{
								pos: position{line: 479, col: 22, offset: 14164},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 479, col: 22, offset: 14164},
										expr: &litMatcher{
											pos:        position{line: 479, col: 23, offset: 14165},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 479, col: 28, offset: 14170},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 479, col: 42, offset: 14184},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 485, col: 1, offset: 14364},
			expr: &choiceExpr{
				pos: position{line: 485, col: 11, offset: 14376},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 485, col: 11, offset: 14376},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 485, col: 30, offset: 14395},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 486, col: 1, offset: 14413},
			expr: &seqExpr{
				pos: position{line: 486, col: 20, offset: 14434},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 486, col: 20, offset: 14434},
						expr: &ruleRefExpr{
							pos:  position{line: 486, col: 21, offset: 14435},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 486, col: 31, offset: 14445},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 486, col: 36, offset: 14450},
						expr: &seqExpr{
							pos: position{line: 486, col: 38, offset: 14452},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 486, col: 38, offset: 14452},
									expr: &litMatcher{
										pos:        position{line: 486, col: 39, offset: 14453},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 486, col: 44, offset: 14458},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 486, col: 58, offset: 14472},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 487, col: 1, offset: 14477},
			expr: &seqExpr{
				pos: position{line: 487, col: 36, offset: 14514},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 487, col: 36, offset: 14514},
						expr: &ruleRefExpr{
							pos:  position{line: 487, col: 37, offset: 14515},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 487, col: 47, offset: 14525},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 487, col: 52, offset: 14530},
						expr: &seqExpr{
							pos: position{line: 487, col: 54, offset: 14532},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 487, col: 54, offset: 14532},
									expr: &choiceExpr{
										pos: position{line: 487, col: 57, offset: 14535},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 487, col: 57, offset: 14535},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 487, col: 64, offset: 14542},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 487, col: 70, offset: 14548},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 487, col: 84, offset: 14562},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 488, col: 1, offset: 14567},
			expr: &choiceExpr{
				pos: position{line: 488, col: 21, offset: 14589},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 488, col: 22, offset: 14590},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 488, col: 22, offset: 14590},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 488, col: 27, offset: 14595},
								expr: &seqExpr{
									pos: position{line: 488, col: 29, offset: 14597},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 488, col: 29, offset: 14597},
											expr: &ruleRefExpr{
												pos:  position{line: 488, col: 30, offset: 14598},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 488, col: 34, offset: 14602},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 488, col: 52, offset: 14620},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 488, col: 52, offset: 14620},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 488, col: 56, offset: 14624},
								expr: &seqExpr{
									pos: position{line: 488, col: 58, offset: 14626},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 488, col: 58, offset: 14626},
											expr: &ruleRefExpr{
												pos:  position{line: 488, col: 59, offset: 14627},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 488, col: 63, offset: 14631},
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
			pos:  position{line: 490, col: 1, offset: 14647},
			expr: &zeroOrMoreExpr{
				pos: position{line: 490, col: 6, offset: 14654},
				expr: &choiceExpr{
					pos: position{line: 490, col: 8, offset: 14656},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 490, col: 8, offset: 14656},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 490, col: 21, offset: 14669},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 490, col: 27, offset: 14675},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 491, col: 1, offset: 14686},
			expr: &zeroOrMoreExpr{
				pos: position{line: 491, col: 5, offset: 14692},
				expr: &choiceExpr{
					pos: position{line: 491, col: 7, offset: 14694},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 491, col: 7, offset: 14694},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 491, col: 20, offset: 14707},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 492, col: 1, offset: 14743},
			expr: &zeroOrMoreExpr{
				pos: position{line: 492, col: 6, offset: 14750},
				expr: &ruleRefExpr{
					pos:  position{line: 492, col: 6, offset: 14750},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 494, col: 1, offset: 14763},
			expr: &charClassMatcher{
				pos:        position{line: 494, col: 14, offset: 14778},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 495, col: 1, offset: 14786},
			expr: &litMatcher{
				pos:        position{line: 495, col: 7, offset: 14794},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 496, col: 1, offset: 14799},
			expr: &choiceExpr{
				pos: position{line: 496, col: 7, offset: 14807},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 496, col: 7, offset: 14807},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 496, col: 7, offset: 14807},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 496, col: 10, offset: 14810},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 496, col: 16, offset: 14816},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 496, col: 16, offset: 14816},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 496, col: 18, offset: 14818},
								expr: &ruleRefExpr{
									pos:  position{line: 496, col: 18, offset: 14818},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 496, col: 37, offset: 14837},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 496, col: 43, offset: 14843},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 496, col: 43, offset: 14843},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 496, col: 46, offset: 14846},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 498, col: 1, offset: 14851},
			expr: &notExpr{
				pos: position{line: 498, col: 7, offset: 14859},
				expr: &anyMatcher{
					line: 498, col: 8, offset: 14860,
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
