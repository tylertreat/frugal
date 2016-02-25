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
		f.Modifier = Optional
	}
	return st
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 79, col: 1, offset: 2189},
			expr: &actionExpr{
				pos: position{line: 79, col: 11, offset: 2201},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 79, col: 11, offset: 2201},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 79, col: 11, offset: 2201},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 79, col: 14, offset: 2204},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 79, col: 25, offset: 2215},
								expr: &seqExpr{
									pos: position{line: 79, col: 27, offset: 2217},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 79, col: 27, offset: 2217},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 79, col: 37, offset: 2227},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 79, col: 44, offset: 2234},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 79, col: 44, offset: 2234},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 79, col: 50, offset: 2240},
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
			pos:  position{line: 152, col: 1, offset: 4849},
			expr: &actionExpr{
				pos: position{line: 152, col: 15, offset: 4865},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 152, col: 15, offset: 4865,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 156, col: 1, offset: 4923},
			expr: &actionExpr{
				pos: position{line: 156, col: 13, offset: 4937},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 156, col: 13, offset: 4937},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 156, col: 13, offset: 4937},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 156, col: 20, offset: 4944},
								expr: &seqExpr{
									pos: position{line: 156, col: 21, offset: 4945},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 156, col: 21, offset: 4945},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 156, col: 31, offset: 4955},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 156, col: 36, offset: 4960},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 156, col: 47, offset: 4971},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 156, col: 47, offset: 4971},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 156, col: 65, offset: 4989},
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
			pos:  position{line: 169, col: 1, offset: 5460},
			expr: &choiceExpr{
				pos: position{line: 169, col: 19, offset: 5480},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 169, col: 19, offset: 5480},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 29, offset: 5490},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 41, offset: 5502},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 49, offset: 5510},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 56, offset: 5517},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 66, offset: 5527},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 75, offset: 5536},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 87, offset: 5548},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 95, offset: 5556},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 171, col: 1, offset: 5565},
			expr: &actionExpr{
				pos: position{line: 171, col: 11, offset: 5577},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 171, col: 11, offset: 5577},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 171, col: 11, offset: 5577},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 21, offset: 5587},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 171, col: 23, offset: 5589},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 171, col: 28, offset: 5594},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 36, offset: 5602},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 175, col: 1, offset: 5650},
			expr: &actionExpr{
				pos: position{line: 175, col: 13, offset: 5664},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 175, col: 13, offset: 5664},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 13, offset: 5664},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 25, offset: 5676},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 27, offset: 5678},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 175, col: 33, offset: 5684},
								expr: &charClassMatcher{
									pos:        position{line: 175, col: 33, offset: 5684},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 43, offset: 5694},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 45, offset: 5696},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 48, offset: 5699},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 59, offset: 5710},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 182, col: 1, offset: 5835},
			expr: &actionExpr{
				pos: position{line: 182, col: 9, offset: 5845},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 182, col: 9, offset: 5845},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 182, col: 9, offset: 5845},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 17, offset: 5853},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 19, offset: 5855},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 23, offset: 5859},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 33, offset: 5869},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 35, offset: 5871},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 40, offset: 5876},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 51, offset: 5887},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 182, col: 53, offset: 5889},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 57, offset: 5893},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 59, offset: 5895},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 65, offset: 5901},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 76, offset: 5912},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 190, col: 1, offset: 6044},
			expr: &actionExpr{
				pos: position{line: 190, col: 8, offset: 6053},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 190, col: 8, offset: 6053},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 190, col: 8, offset: 6053},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 15, offset: 6060},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 17, offset: 6062},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 22, offset: 6067},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 33, offset: 6078},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 190, col: 36, offset: 6081},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 40, offset: 6085},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 43, offset: 6088},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 190, col: 50, offset: 6095},
								expr: &seqExpr{
									pos: position{line: 190, col: 51, offset: 6096},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 190, col: 51, offset: 6096},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 190, col: 61, offset: 6106},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 190, col: 66, offset: 6111},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 70, offset: 6115},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 213, col: 1, offset: 6727},
			expr: &actionExpr{
				pos: position{line: 213, col: 13, offset: 6741},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 213, col: 13, offset: 6741},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 213, col: 13, offset: 6741},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 20, offset: 6748},
								expr: &seqExpr{
									pos: position{line: 213, col: 21, offset: 6749},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 213, col: 21, offset: 6749},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 31, offset: 6759},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 213, col: 36, offset: 6764},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 41, offset: 6769},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 213, col: 52, offset: 6780},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 213, col: 54, offset: 6782},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 60, offset: 6788},
								expr: &seqExpr{
									pos: position{line: 213, col: 61, offset: 6789},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 213, col: 61, offset: 6789},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 65, offset: 6793},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 67, offset: 6795},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 213, col: 81, offset: 6809},
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 81, offset: 6809},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 228, col: 1, offset: 7145},
			expr: &actionExpr{
				pos: position{line: 228, col: 11, offset: 7157},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 228, col: 11, offset: 7157},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 11, offset: 7157},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 21, offset: 7167},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 23, offset: 7169},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 27, offset: 7173},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 37, offset: 7183},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 39, offset: 7185},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 44, offset: 7190},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 55, offset: 7201},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 235, col: 1, offset: 7310},
			expr: &actionExpr{
				pos: position{line: 235, col: 10, offset: 7321},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 235, col: 10, offset: 7321},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 10, offset: 7321},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 19, offset: 7330},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 21, offset: 7332},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 24, offset: 7335},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 236, col: 1, offset: 7375},
			expr: &actionExpr{
				pos: position{line: 236, col: 13, offset: 7389},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 236, col: 13, offset: 7389},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 236, col: 13, offset: 7389},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 236, col: 25, offset: 7401},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 236, col: 27, offset: 7403},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 236, col: 30, offset: 7406},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 237, col: 1, offset: 7457},
			expr: &actionExpr{
				pos: position{line: 237, col: 9, offset: 7467},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 237, col: 9, offset: 7467},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 237, col: 9, offset: 7467},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 17, offset: 7475},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 19, offset: 7477},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 22, offset: 7480},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 238, col: 1, offset: 7527},
			expr: &actionExpr{
				pos: position{line: 238, col: 14, offset: 7542},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 238, col: 14, offset: 7542},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 238, col: 14, offset: 7542},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 19, offset: 7547},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 30, offset: 7558},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 238, col: 33, offset: 7561},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 37, offset: 7565},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7568},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 47, offset: 7575},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 238, col: 57, offset: 7585},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 61, offset: 7589},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 248, col: 1, offset: 7750},
			expr: &actionExpr{
				pos: position{line: 248, col: 13, offset: 7764},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 248, col: 13, offset: 7764},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 248, col: 20, offset: 7771},
						expr: &seqExpr{
							pos: position{line: 248, col: 21, offset: 7772},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 248, col: 21, offset: 7772},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 248, col: 27, offset: 7778},
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
			pos:  position{line: 257, col: 1, offset: 7959},
			expr: &actionExpr{
				pos: position{line: 257, col: 9, offset: 7969},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 257, col: 9, offset: 7969},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 257, col: 9, offset: 7969},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 16, offset: 7976},
								expr: &seqExpr{
									pos: position{line: 257, col: 17, offset: 7977},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 257, col: 17, offset: 7977},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 27, offset: 7987},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 257, col: 32, offset: 7992},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 35, offset: 7995},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 47, offset: 8007},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 257, col: 49, offset: 8009},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 53, offset: 8013},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 55, offset: 8015},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 59, offset: 8019},
								expr: &ruleRefExpr{
									pos:  position{line: 257, col: 59, offset: 8019},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 69, offset: 8029},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 71, offset: 8031},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 75, offset: 8035},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 85, offset: 8045},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 87, offset: 8047},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 92, offset: 8052},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 103, offset: 8063},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 106, offset: 8066},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 110, offset: 8070},
								expr: &seqExpr{
									pos: position{line: 257, col: 111, offset: 8071},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 257, col: 111, offset: 8071},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 115, offset: 8075},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 117, offset: 8077},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 257, col: 130, offset: 8090},
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 130, offset: 8090},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 279, col: 1, offset: 8550},
			expr: &actionExpr{
				pos: position{line: 279, col: 12, offset: 8563},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 279, col: 13, offset: 8564},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 279, col: 13, offset: 8564},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 279, col: 26, offset: 8577},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 289, col: 1, offset: 8788},
			expr: &actionExpr{
				pos: position{line: 289, col: 11, offset: 8800},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 289, col: 11, offset: 8800},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 289, col: 11, offset: 8800},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 289, col: 21, offset: 8810},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 289, col: 23, offset: 8812},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 289, col: 28, offset: 8817},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 289, col: 39, offset: 8828},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 289, col: 41, offset: 8830},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 289, col: 49, offset: 8838},
								expr: &seqExpr{
									pos: position{line: 289, col: 50, offset: 8839},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 289, col: 50, offset: 8839},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 289, col: 60, offset: 8849},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 289, col: 63, offset: 8852},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 289, col: 74, offset: 8863},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 289, col: 79, offset: 8868},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 289, col: 82, offset: 8871},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 289, col: 86, offset: 8875},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 289, col: 89, offset: 8878},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 289, col: 97, offset: 8886},
								expr: &seqExpr{
									pos: position{line: 289, col: 98, offset: 8887},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 289, col: 98, offset: 8887},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 289, col: 107, offset: 8896},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 289, col: 113, offset: 8902},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 289, col: 113, offset: 8902},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 289, col: 119, offset: 8908},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 289, col: 138, offset: 8927},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 305, col: 1, offset: 9311},
			expr: &actionExpr{
				pos: position{line: 305, col: 21, offset: 9333},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 305, col: 21, offset: 9333,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 309, col: 1, offset: 9402},
			expr: &actionExpr{
				pos: position{line: 309, col: 12, offset: 9415},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 309, col: 12, offset: 9415},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 309, col: 12, offset: 9415},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 309, col: 19, offset: 9422},
								expr: &seqExpr{
									pos: position{line: 309, col: 20, offset: 9423},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 309, col: 20, offset: 9423},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 309, col: 30, offset: 9433},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 309, col: 35, offset: 9438},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 309, col: 42, offset: 9445},
								expr: &seqExpr{
									pos: position{line: 309, col: 43, offset: 9446},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 309, col: 43, offset: 9446},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 309, col: 52, offset: 9455},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 309, col: 57, offset: 9460},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 309, col: 61, offset: 9464},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 309, col: 74, offset: 9477},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 309, col: 77, offset: 9480},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 309, col: 82, offset: 9485},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 309, col: 93, offset: 9496},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 309, col: 95, offset: 9498},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 309, col: 99, offset: 9502},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 309, col: 102, offset: 9505},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 309, col: 112, offset: 9515},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 309, col: 122, offset: 9525},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 309, col: 126, offset: 9529},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 309, col: 129, offset: 9532},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 309, col: 140, offset: 9543},
								expr: &ruleRefExpr{
									pos:  position{line: 309, col: 140, offset: 9543},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 309, col: 148, offset: 9551},
							expr: &ruleRefExpr{
								pos:  position{line: 309, col: 148, offset: 9551},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 336, col: 1, offset: 10146},
			expr: &actionExpr{
				pos: position{line: 336, col: 16, offset: 10163},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 336, col: 16, offset: 10163},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 336, col: 21, offset: 10168},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 336, col: 21, offset: 10168},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 336, col: 30, offset: 10177},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 343, col: 1, offset: 10299},
			expr: &actionExpr{
				pos: position{line: 343, col: 10, offset: 10310},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 343, col: 10, offset: 10310},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 343, col: 10, offset: 10310},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 343, col: 19, offset: 10319},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 343, col: 22, offset: 10322},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 343, col: 26, offset: 10326},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 343, col: 29, offset: 10329},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 343, col: 40, offset: 10340},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 343, col: 50, offset: 10350},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 347, col: 1, offset: 10386},
			expr: &actionExpr{
				pos: position{line: 347, col: 13, offset: 10400},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 347, col: 13, offset: 10400},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 347, col: 18, offset: 10405},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 347, col: 18, offset: 10405},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 347, col: 29, offset: 10416},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 347, col: 45, offset: 10432},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 354, col: 1, offset: 10557},
			expr: &actionExpr{
				pos: position{line: 354, col: 12, offset: 10570},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 354, col: 13, offset: 10571},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 354, col: 13, offset: 10571},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 354, col: 22, offset: 10580},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 354, col: 31, offset: 10589},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 354, col: 39, offset: 10597},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 354, col: 47, offset: 10605},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 354, col: 55, offset: 10613},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 354, col: 66, offset: 10624},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 354, col: 77, offset: 10635},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 358, col: 1, offset: 10695},
			expr: &actionExpr{
				pos: position{line: 358, col: 17, offset: 10713},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 358, col: 17, offset: 10713},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 358, col: 22, offset: 10718},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 358, col: 22, offset: 10718},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 358, col: 32, offset: 10728},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 358, col: 42, offset: 10738},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 362, col: 1, offset: 10773},
			expr: &actionExpr{
				pos: position{line: 362, col: 11, offset: 10785},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 362, col: 11, offset: 10785},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 362, col: 11, offset: 10785},
							expr: &ruleRefExpr{
								pos:  position{line: 362, col: 11, offset: 10785},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 362, col: 20, offset: 10794},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 362, col: 27, offset: 10801},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 362, col: 30, offset: 10804},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 362, col: 34, offset: 10808},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 362, col: 44, offset: 10818},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 362, col: 47, offset: 10821},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 362, col: 51, offset: 10825},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 362, col: 54, offset: 10828},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 362, col: 60, offset: 10834},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 362, col: 70, offset: 10844},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 362, col: 73, offset: 10847},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 370, col: 1, offset: 10970},
			expr: &actionExpr{
				pos: position{line: 370, col: 11, offset: 10982},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 370, col: 11, offset: 10982},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 370, col: 11, offset: 10982},
							expr: &ruleRefExpr{
								pos:  position{line: 370, col: 11, offset: 10982},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 370, col: 20, offset: 10991},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 27, offset: 10998},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 370, col: 30, offset: 11001},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 370, col: 34, offset: 11005},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 44, offset: 11015},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 370, col: 47, offset: 11018},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 377, col: 1, offset: 11109},
			expr: &actionExpr{
				pos: position{line: 377, col: 12, offset: 11122},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 377, col: 12, offset: 11122},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 377, col: 12, offset: 11122},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 377, col: 20, offset: 11130},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 377, col: 23, offset: 11133},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 377, col: 27, offset: 11137},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 377, col: 37, offset: 11147},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 377, col: 40, offset: 11150},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 384, col: 1, offset: 11242},
			expr: &actionExpr{
				pos: position{line: 384, col: 11, offset: 11254},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 384, col: 11, offset: 11254},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 384, col: 11, offset: 11254},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 384, col: 22, offset: 11265},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 384, col: 30, offset: 11273},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 388, col: 1, offset: 11310},
			expr: &choiceExpr{
				pos: position{line: 388, col: 14, offset: 11325},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 388, col: 14, offset: 11325},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 388, col: 24, offset: 11335},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 388, col: 41, offset: 11352},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 388, col: 55, offset: 11366},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 388, col: 66, offset: 11377},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 388, col: 78, offset: 11389},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 390, col: 1, offset: 11401},
			expr: &actionExpr{
				pos: position{line: 390, col: 15, offset: 11417},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 390, col: 15, offset: 11417},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 390, col: 15, offset: 11417},
							expr: &charClassMatcher{
								pos:        position{line: 390, col: 15, offset: 11417},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 390, col: 21, offset: 11423},
							expr: &ruleRefExpr{
								pos:  position{line: 390, col: 21, offset: 11423},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 394, col: 1, offset: 11487},
			expr: &actionExpr{
				pos: position{line: 394, col: 18, offset: 11506},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 394, col: 18, offset: 11506},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 394, col: 18, offset: 11506},
							expr: &charClassMatcher{
								pos:        position{line: 394, col: 18, offset: 11506},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 394, col: 24, offset: 11512},
							expr: &ruleRefExpr{
								pos:  position{line: 394, col: 24, offset: 11512},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 394, col: 31, offset: 11519},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 394, col: 35, offset: 11523},
							expr: &ruleRefExpr{
								pos:  position{line: 394, col: 35, offset: 11523},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 394, col: 42, offset: 11530},
							expr: &seqExpr{
								pos: position{line: 394, col: 44, offset: 11532},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 394, col: 44, offset: 11532},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 394, col: 51, offset: 11539},
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
			pos:  position{line: 398, col: 1, offset: 11609},
			expr: &actionExpr{
				pos: position{line: 398, col: 13, offset: 11623},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 398, col: 13, offset: 11623},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 398, col: 13, offset: 11623},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 17, offset: 11627},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 398, col: 20, offset: 11630},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 398, col: 27, offset: 11637},
								expr: &seqExpr{
									pos: position{line: 398, col: 28, offset: 11638},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 398, col: 28, offset: 11638},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 39, offset: 11649},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 398, col: 42, offset: 11652},
											expr: &ruleRefExpr{
												pos:  position{line: 398, col: 42, offset: 11652},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 57, offset: 11667},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 62, offset: 11672},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 398, col: 65, offset: 11675},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 407, col: 1, offset: 11869},
			expr: &actionExpr{
				pos: position{line: 407, col: 12, offset: 11882},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 407, col: 12, offset: 11882},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 407, col: 12, offset: 11882},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 407, col: 16, offset: 11886},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 407, col: 19, offset: 11889},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 407, col: 26, offset: 11896},
								expr: &seqExpr{
									pos: position{line: 407, col: 27, offset: 11897},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 407, col: 27, offset: 11897},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 407, col: 38, offset: 11908},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 407, col: 41, offset: 11911},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 407, col: 45, offset: 11915},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 407, col: 48, offset: 11918},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 407, col: 59, offset: 11929},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 407, col: 63, offset: 11933},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 407, col: 63, offset: 11933},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 407, col: 69, offset: 11939},
													expr: &litMatcher{
														pos:        position{line: 407, col: 70, offset: 11940},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 407, col: 75, offset: 11945},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 407, col: 80, offset: 11950},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 427, col: 1, offset: 12500},
			expr: &choiceExpr{
				pos: position{line: 427, col: 19, offset: 12520},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 427, col: 19, offset: 12520},
						name: "Scope",
					},
					&ruleRefExpr{
						pos:  position{line: 427, col: 27, offset: 12528},
						name: "Async",
					},
				},
			},
		},
		{
			name: "Scope",
			pos:  position{line: 429, col: 1, offset: 12535},
			expr: &actionExpr{
				pos: position{line: 429, col: 9, offset: 12545},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 429, col: 9, offset: 12545},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 429, col: 9, offset: 12545},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 429, col: 16, offset: 12552},
								expr: &seqExpr{
									pos: position{line: 429, col: 17, offset: 12553},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 429, col: 17, offset: 12553},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 429, col: 27, offset: 12563},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 429, col: 32, offset: 12568},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 429, col: 40, offset: 12576},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 429, col: 43, offset: 12579},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 429, col: 48, offset: 12584},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 429, col: 59, offset: 12595},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 429, col: 62, offset: 12598},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 429, col: 69, offset: 12605},
								expr: &ruleRefExpr{
									pos:  position{line: 429, col: 69, offset: 12605},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 429, col: 77, offset: 12613},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 429, col: 80, offset: 12616},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 429, col: 84, offset: 12620},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 429, col: 87, offset: 12623},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 429, col: 98, offset: 12634},
								expr: &seqExpr{
									pos: position{line: 429, col: 99, offset: 12635},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 429, col: 99, offset: 12635},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 429, col: 109, offset: 12645},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 429, col: 115, offset: 12651},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 429, col: 115, offset: 12651},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 429, col: 121, offset: 12657},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 429, col: 138, offset: 12674},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 450, col: 1, offset: 13219},
			expr: &actionExpr{
				pos: position{line: 450, col: 19, offset: 13239},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 450, col: 19, offset: 13239,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 454, col: 1, offset: 13306},
			expr: &actionExpr{
				pos: position{line: 454, col: 10, offset: 13317},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 454, col: 10, offset: 13317},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 454, col: 10, offset: 13317},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 454, col: 19, offset: 13326},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 454, col: 22, offset: 13329},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 454, col: 34, offset: 13341},
							expr: &seqExpr{
								pos: position{line: 454, col: 35, offset: 13342},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 454, col: 35, offset: 13342},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 454, col: 39, offset: 13346},
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
			pos:  position{line: 459, col: 1, offset: 13477},
			expr: &choiceExpr{
				pos: position{line: 459, col: 15, offset: 13493},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 459, col: 16, offset: 13494},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 459, col: 16, offset: 13494},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 459, col: 20, offset: 13498},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 459, col: 31, offset: 13509},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 459, col: 38, offset: 13516},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 461, col: 1, offset: 13528},
			expr: &oneOrMoreExpr{
				pos: position{line: 461, col: 14, offset: 13543},
				expr: &charClassMatcher{
					pos:        position{line: 461, col: 14, offset: 13543},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 463, col: 1, offset: 13561},
			expr: &actionExpr{
				pos: position{line: 463, col: 13, offset: 13575},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 463, col: 13, offset: 13575},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 463, col: 13, offset: 13575},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 463, col: 20, offset: 13582},
								expr: &seqExpr{
									pos: position{line: 463, col: 21, offset: 13583},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 463, col: 21, offset: 13583},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 463, col: 31, offset: 13593},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 463, col: 36, offset: 13598},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 463, col: 41, offset: 13603},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 463, col: 52, offset: 13614},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 463, col: 54, offset: 13616},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 463, col: 58, offset: 13620},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 463, col: 61, offset: 13623},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 463, col: 65, offset: 13627},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 463, col: 76, offset: 13638},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Async",
			pos:  position{line: 475, col: 1, offset: 13908},
			expr: &actionExpr{
				pos: position{line: 475, col: 9, offset: 13918},
				run: (*parser).callonAsync1,
				expr: &seqExpr{
					pos: position{line: 475, col: 9, offset: 13918},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 475, col: 9, offset: 13918},
							val:        "async",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 17, offset: 13926},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 475, col: 19, offset: 13928},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 475, col: 24, offset: 13933},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 35, offset: 13944},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 475, col: 37, offset: 13946},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 475, col: 45, offset: 13954},
								expr: &seqExpr{
									pos: position{line: 475, col: 46, offset: 13955},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 475, col: 46, offset: 13955},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 475, col: 56, offset: 13965},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 475, col: 59, offset: 13968},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 475, col: 70, offset: 13979},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 75, offset: 13984},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 475, col: 78, offset: 13987},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 82, offset: 13991},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 475, col: 85, offset: 13994},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 475, col: 93, offset: 14002},
								expr: &seqExpr{
									pos: position{line: 475, col: 94, offset: 14003},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 475, col: 94, offset: 14003},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 475, col: 103, offset: 14012},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 475, col: 109, offset: 14018},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 475, col: 109, offset: 14018},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 475, col: 115, offset: 14024},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 134, offset: 14043},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 495, col: 1, offset: 14674},
			expr: &actionExpr{
				pos: position{line: 495, col: 11, offset: 14686},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 495, col: 12, offset: 14687},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 495, col: 13, offset: 14688},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 495, col: 13, offset: 14688},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 495, col: 17, offset: 14692},
									expr: &choiceExpr{
										pos: position{line: 495, col: 18, offset: 14693},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 495, col: 18, offset: 14693},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 495, col: 25, offset: 14700},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 495, col: 32, offset: 14707},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 495, col: 40, offset: 14715},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 495, col: 40, offset: 14715},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 495, col: 45, offset: 14720},
									expr: &choiceExpr{
										pos: position{line: 495, col: 46, offset: 14721},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 495, col: 46, offset: 14721},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 495, col: 53, offset: 14728},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 495, col: 60, offset: 14735},
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
			pos:  position{line: 502, col: 1, offset: 14951},
			expr: &actionExpr{
				pos: position{line: 502, col: 14, offset: 14966},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 502, col: 14, offset: 14966},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 502, col: 14, offset: 14966},
							expr: &choiceExpr{
								pos: position{line: 502, col: 15, offset: 14967},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 502, col: 15, offset: 14967},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 502, col: 24, offset: 14976},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 502, col: 30, offset: 14982},
							expr: &choiceExpr{
								pos: position{line: 502, col: 31, offset: 14983},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 502, col: 31, offset: 14983},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 502, col: 40, offset: 14992},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 502, col: 48, offset: 15000},
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
			pos:  position{line: 506, col: 1, offset: 15055},
			expr: &charClassMatcher{
				pos:        position{line: 506, col: 17, offset: 15073},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 507, col: 1, offset: 15078},
			expr: &charClassMatcher{
				pos:        position{line: 507, col: 10, offset: 15089},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 508, col: 1, offset: 15098},
			expr: &charClassMatcher{
				pos:        position{line: 508, col: 9, offset: 15108},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 510, col: 1, offset: 15115},
			expr: &anyMatcher{
				line: 510, col: 14, offset: 15130,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 511, col: 1, offset: 15132},
			expr: &actionExpr{
				pos: position{line: 511, col: 13, offset: 15146},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 511, col: 13, offset: 15146},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 511, col: 13, offset: 15146},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 511, col: 20, offset: 15153},
							expr: &seqExpr{
								pos: position{line: 511, col: 22, offset: 15155},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 511, col: 22, offset: 15155},
										expr: &litMatcher{
											pos:        position{line: 511, col: 23, offset: 15156},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 511, col: 28, offset: 15161},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 511, col: 42, offset: 15175},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 517, col: 1, offset: 15355},
			expr: &choiceExpr{
				pos: position{line: 517, col: 11, offset: 15367},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 517, col: 11, offset: 15367},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 517, col: 30, offset: 15386},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 518, col: 1, offset: 15404},
			expr: &seqExpr{
				pos: position{line: 518, col: 20, offset: 15425},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 518, col: 20, offset: 15425},
						expr: &ruleRefExpr{
							pos:  position{line: 518, col: 21, offset: 15426},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 518, col: 31, offset: 15436},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 518, col: 36, offset: 15441},
						expr: &seqExpr{
							pos: position{line: 518, col: 38, offset: 15443},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 518, col: 38, offset: 15443},
									expr: &litMatcher{
										pos:        position{line: 518, col: 39, offset: 15444},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 518, col: 44, offset: 15449},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 518, col: 58, offset: 15463},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 519, col: 1, offset: 15468},
			expr: &seqExpr{
				pos: position{line: 519, col: 36, offset: 15505},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 519, col: 36, offset: 15505},
						expr: &ruleRefExpr{
							pos:  position{line: 519, col: 37, offset: 15506},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 519, col: 47, offset: 15516},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 519, col: 52, offset: 15521},
						expr: &seqExpr{
							pos: position{line: 519, col: 54, offset: 15523},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 519, col: 54, offset: 15523},
									expr: &choiceExpr{
										pos: position{line: 519, col: 57, offset: 15526},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 519, col: 57, offset: 15526},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 519, col: 64, offset: 15533},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 519, col: 70, offset: 15539},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 519, col: 84, offset: 15553},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 520, col: 1, offset: 15558},
			expr: &choiceExpr{
				pos: position{line: 520, col: 21, offset: 15580},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 520, col: 22, offset: 15581},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 520, col: 22, offset: 15581},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 520, col: 27, offset: 15586},
								expr: &seqExpr{
									pos: position{line: 520, col: 29, offset: 15588},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 520, col: 29, offset: 15588},
											expr: &ruleRefExpr{
												pos:  position{line: 520, col: 30, offset: 15589},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 520, col: 34, offset: 15593},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 520, col: 52, offset: 15611},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 520, col: 52, offset: 15611},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 520, col: 56, offset: 15615},
								expr: &seqExpr{
									pos: position{line: 520, col: 58, offset: 15617},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 520, col: 58, offset: 15617},
											expr: &ruleRefExpr{
												pos:  position{line: 520, col: 59, offset: 15618},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 520, col: 63, offset: 15622},
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
			pos:  position{line: 522, col: 1, offset: 15638},
			expr: &zeroOrMoreExpr{
				pos: position{line: 522, col: 6, offset: 15645},
				expr: &choiceExpr{
					pos: position{line: 522, col: 8, offset: 15647},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 522, col: 8, offset: 15647},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 522, col: 21, offset: 15660},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 522, col: 27, offset: 15666},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 523, col: 1, offset: 15677},
			expr: &zeroOrMoreExpr{
				pos: position{line: 523, col: 5, offset: 15683},
				expr: &choiceExpr{
					pos: position{line: 523, col: 7, offset: 15685},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 523, col: 7, offset: 15685},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 523, col: 20, offset: 15698},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 524, col: 1, offset: 15734},
			expr: &zeroOrMoreExpr{
				pos: position{line: 524, col: 6, offset: 15741},
				expr: &ruleRefExpr{
					pos:  position{line: 524, col: 6, offset: 15741},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 526, col: 1, offset: 15754},
			expr: &charClassMatcher{
				pos:        position{line: 526, col: 14, offset: 15769},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 527, col: 1, offset: 15777},
			expr: &litMatcher{
				pos:        position{line: 527, col: 7, offset: 15785},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 528, col: 1, offset: 15790},
			expr: &choiceExpr{
				pos: position{line: 528, col: 7, offset: 15798},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 528, col: 7, offset: 15798},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 528, col: 7, offset: 15798},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 528, col: 10, offset: 15801},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 528, col: 16, offset: 15807},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 528, col: 16, offset: 15807},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 528, col: 18, offset: 15809},
								expr: &ruleRefExpr{
									pos:  position{line: 528, col: 18, offset: 15809},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 528, col: 37, offset: 15828},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 528, col: 43, offset: 15834},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 528, col: 43, offset: 15834},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 528, col: 46, offset: 15837},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 530, col: 1, offset: 15842},
			expr: &notExpr{
				pos: position{line: 530, col: 7, offset: 15850},
				expr: &anyMatcher{
					line: 530, col: 8, offset: 15851,
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
	if req != nil {
		f.Modifier = req.(FieldModifier)
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
	return p.cur.onField1(stack["docstr"], stack["id"], stack["req"], stack["typ"], stack["name"], stack["def"])
}

func (c *current) onFieldReq1() (interface{}, error) {
	if bytes.Equal(c.text, []byte("required")) {
		return Required, nil
	} else if bytes.Equal(c.text, []byte("optional")) {
		return Optional, nil
	}

	return Default, nil
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
