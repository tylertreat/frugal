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
			pos:  position{line: 147, col: 1, offset: 4669},
			expr: &actionExpr{
				pos: position{line: 147, col: 15, offset: 4685},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 147, col: 15, offset: 4685,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 151, col: 1, offset: 4743},
			expr: &actionExpr{
				pos: position{line: 151, col: 13, offset: 4757},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 151, col: 13, offset: 4757},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 151, col: 13, offset: 4757},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 151, col: 20, offset: 4764},
								expr: &seqExpr{
									pos: position{line: 151, col: 21, offset: 4765},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 151, col: 21, offset: 4765},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 151, col: 31, offset: 4775},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 151, col: 36, offset: 4780},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 151, col: 47, offset: 4791},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 151, col: 47, offset: 4791},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 151, col: 65, offset: 4809},
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
			pos:  position{line: 164, col: 1, offset: 5280},
			expr: &choiceExpr{
				pos: position{line: 164, col: 19, offset: 5300},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 164, col: 19, offset: 5300},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 29, offset: 5310},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 41, offset: 5322},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 49, offset: 5330},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 56, offset: 5337},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 66, offset: 5347},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 75, offset: 5356},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 87, offset: 5368},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 95, offset: 5376},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 166, col: 1, offset: 5385},
			expr: &actionExpr{
				pos: position{line: 166, col: 11, offset: 5397},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 166, col: 11, offset: 5397},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 166, col: 11, offset: 5397},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 166, col: 21, offset: 5407},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 166, col: 23, offset: 5409},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 166, col: 28, offset: 5414},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 166, col: 36, offset: 5422},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 170, col: 1, offset: 5470},
			expr: &actionExpr{
				pos: position{line: 170, col: 13, offset: 5484},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 170, col: 13, offset: 5484},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 170, col: 13, offset: 5484},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 25, offset: 5496},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 27, offset: 5498},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 170, col: 33, offset: 5504},
								expr: &charClassMatcher{
									pos:        position{line: 170, col: 33, offset: 5504},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 43, offset: 5514},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 45, offset: 5516},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 170, col: 48, offset: 5519},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 59, offset: 5530},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 177, col: 1, offset: 5655},
			expr: &actionExpr{
				pos: position{line: 177, col: 9, offset: 5665},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 177, col: 9, offset: 5665},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 177, col: 9, offset: 5665},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 17, offset: 5673},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 19, offset: 5675},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 23, offset: 5679},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 33, offset: 5689},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 35, offset: 5691},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 40, offset: 5696},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 51, offset: 5707},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 177, col: 53, offset: 5709},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 57, offset: 5713},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 59, offset: 5715},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 65, offset: 5721},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 76, offset: 5732},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 185, col: 1, offset: 5864},
			expr: &actionExpr{
				pos: position{line: 185, col: 8, offset: 5873},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 185, col: 8, offset: 5873},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 185, col: 8, offset: 5873},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 15, offset: 5880},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 185, col: 17, offset: 5882},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 185, col: 22, offset: 5887},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 33, offset: 5898},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 185, col: 36, offset: 5901},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 40, offset: 5905},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 185, col: 43, offset: 5908},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 185, col: 50, offset: 5915},
								expr: &seqExpr{
									pos: position{line: 185, col: 51, offset: 5916},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 185, col: 51, offset: 5916},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 185, col: 61, offset: 5926},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 185, col: 66, offset: 5931},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 70, offset: 5935},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 208, col: 1, offset: 6547},
			expr: &actionExpr{
				pos: position{line: 208, col: 13, offset: 6561},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 208, col: 13, offset: 6561},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 208, col: 13, offset: 6561},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 208, col: 20, offset: 6568},
								expr: &seqExpr{
									pos: position{line: 208, col: 21, offset: 6569},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 208, col: 21, offset: 6569},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 31, offset: 6579},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 208, col: 36, offset: 6584},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 208, col: 41, offset: 6589},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 208, col: 52, offset: 6600},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 208, col: 54, offset: 6602},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 208, col: 60, offset: 6608},
								expr: &seqExpr{
									pos: position{line: 208, col: 61, offset: 6609},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 208, col: 61, offset: 6609},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 65, offset: 6613},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 67, offset: 6615},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 208, col: 81, offset: 6629},
							expr: &ruleRefExpr{
								pos:  position{line: 208, col: 81, offset: 6629},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 223, col: 1, offset: 6965},
			expr: &actionExpr{
				pos: position{line: 223, col: 11, offset: 6977},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 223, col: 11, offset: 6977},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 223, col: 11, offset: 6977},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 223, col: 21, offset: 6987},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 223, col: 23, offset: 6989},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 223, col: 27, offset: 6993},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 223, col: 37, offset: 7003},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 223, col: 39, offset: 7005},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 223, col: 44, offset: 7010},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 223, col: 55, offset: 7021},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 230, col: 1, offset: 7130},
			expr: &actionExpr{
				pos: position{line: 230, col: 10, offset: 7141},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 230, col: 10, offset: 7141},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 230, col: 10, offset: 7141},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 230, col: 19, offset: 7150},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 230, col: 21, offset: 7152},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 230, col: 24, offset: 7155},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 231, col: 1, offset: 7195},
			expr: &actionExpr{
				pos: position{line: 231, col: 13, offset: 7209},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 231, col: 13, offset: 7209},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 231, col: 13, offset: 7209},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 231, col: 25, offset: 7221},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 231, col: 27, offset: 7223},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 231, col: 30, offset: 7226},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 232, col: 1, offset: 7277},
			expr: &actionExpr{
				pos: position{line: 232, col: 9, offset: 7287},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 232, col: 9, offset: 7287},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 232, col: 9, offset: 7287},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 232, col: 17, offset: 7295},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 232, col: 19, offset: 7297},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 232, col: 22, offset: 7300},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 233, col: 1, offset: 7347},
			expr: &actionExpr{
				pos: position{line: 233, col: 14, offset: 7362},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 233, col: 14, offset: 7362},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 233, col: 14, offset: 7362},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 233, col: 19, offset: 7367},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 30, offset: 7378},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 233, col: 33, offset: 7381},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 37, offset: 7385},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 233, col: 40, offset: 7388},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 233, col: 47, offset: 7395},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 233, col: 57, offset: 7405},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 233, col: 61, offset: 7409},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 243, col: 1, offset: 7570},
			expr: &actionExpr{
				pos: position{line: 243, col: 13, offset: 7584},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 243, col: 13, offset: 7584},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 243, col: 20, offset: 7591},
						expr: &seqExpr{
							pos: position{line: 243, col: 21, offset: 7592},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 243, col: 21, offset: 7592},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 243, col: 27, offset: 7598},
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
			pos:  position{line: 252, col: 1, offset: 7779},
			expr: &actionExpr{
				pos: position{line: 252, col: 9, offset: 7789},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 252, col: 9, offset: 7789},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 252, col: 9, offset: 7789},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 16, offset: 7796},
								expr: &seqExpr{
									pos: position{line: 252, col: 17, offset: 7797},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 252, col: 17, offset: 7797},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 252, col: 27, offset: 7807},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 252, col: 32, offset: 7812},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 35, offset: 7815},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 47, offset: 7827},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 252, col: 49, offset: 7829},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 53, offset: 7833},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 55, offset: 7835},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 59, offset: 7839},
								expr: &ruleRefExpr{
									pos:  position{line: 252, col: 59, offset: 7839},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 74, offset: 7854},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 76, offset: 7856},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 80, offset: 7860},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 90, offset: 7870},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 92, offset: 7872},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 97, offset: 7877},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 252, col: 108, offset: 7888},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 252, col: 111, offset: 7891},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 252, col: 115, offset: 7895},
								expr: &seqExpr{
									pos: position{line: 252, col: 116, offset: 7896},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 252, col: 116, offset: 7896},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 252, col: 120, offset: 7900},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 252, col: 122, offset: 7902},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 252, col: 135, offset: 7915},
							expr: &ruleRefExpr{
								pos:  position{line: 252, col: 135, offset: 7915},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 274, col: 1, offset: 8377},
			expr: &actionExpr{
				pos: position{line: 274, col: 17, offset: 8395},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 274, col: 18, offset: 8396},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 274, col: 18, offset: 8396},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 274, col: 31, offset: 8409},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 282, col: 1, offset: 8552},
			expr: &actionExpr{
				pos: position{line: 282, col: 11, offset: 8564},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 282, col: 11, offset: 8564},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 282, col: 11, offset: 8564},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 282, col: 21, offset: 8574},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 282, col: 23, offset: 8576},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 282, col: 28, offset: 8581},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 282, col: 39, offset: 8592},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 282, col: 41, offset: 8594},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 282, col: 49, offset: 8602},
								expr: &seqExpr{
									pos: position{line: 282, col: 50, offset: 8603},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 282, col: 50, offset: 8603},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 282, col: 60, offset: 8613},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 282, col: 63, offset: 8616},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 282, col: 74, offset: 8627},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 282, col: 79, offset: 8632},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 282, col: 82, offset: 8635},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 282, col: 86, offset: 8639},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 282, col: 89, offset: 8642},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 282, col: 97, offset: 8650},
								expr: &seqExpr{
									pos: position{line: 282, col: 98, offset: 8651},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 282, col: 98, offset: 8651},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 282, col: 107, offset: 8660},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 282, col: 113, offset: 8666},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 282, col: 113, offset: 8666},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 282, col: 119, offset: 8672},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 282, col: 138, offset: 8691},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 298, col: 1, offset: 9075},
			expr: &actionExpr{
				pos: position{line: 298, col: 21, offset: 9097},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 298, col: 21, offset: 9097,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 302, col: 1, offset: 9166},
			expr: &actionExpr{
				pos: position{line: 302, col: 12, offset: 9179},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 302, col: 12, offset: 9179},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 302, col: 12, offset: 9179},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 302, col: 19, offset: 9186},
								expr: &seqExpr{
									pos: position{line: 302, col: 20, offset: 9187},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 302, col: 20, offset: 9187},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 302, col: 30, offset: 9197},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 302, col: 35, offset: 9202},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 302, col: 42, offset: 9209},
								expr: &seqExpr{
									pos: position{line: 302, col: 43, offset: 9210},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 302, col: 43, offset: 9210},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 302, col: 52, offset: 9219},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 302, col: 57, offset: 9224},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 302, col: 61, offset: 9228},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 74, offset: 9241},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 302, col: 77, offset: 9244},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 302, col: 82, offset: 9249},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 93, offset: 9260},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 302, col: 95, offset: 9262},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 99, offset: 9266},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 302, col: 102, offset: 9269},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 302, col: 112, offset: 9279},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 302, col: 122, offset: 9289},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 302, col: 126, offset: 9293},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 302, col: 129, offset: 9296},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 302, col: 140, offset: 9307},
								expr: &ruleRefExpr{
									pos:  position{line: 302, col: 140, offset: 9307},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 302, col: 148, offset: 9315},
							expr: &ruleRefExpr{
								pos:  position{line: 302, col: 148, offset: 9315},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 329, col: 1, offset: 9910},
			expr: &actionExpr{
				pos: position{line: 329, col: 16, offset: 9927},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 329, col: 16, offset: 9927},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 329, col: 21, offset: 9932},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 329, col: 21, offset: 9932},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 329, col: 30, offset: 9941},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 336, col: 1, offset: 10063},
			expr: &actionExpr{
				pos: position{line: 336, col: 10, offset: 10074},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 336, col: 10, offset: 10074},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 336, col: 10, offset: 10074},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 336, col: 19, offset: 10083},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 336, col: 22, offset: 10086},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 336, col: 26, offset: 10090},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 336, col: 29, offset: 10093},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 336, col: 40, offset: 10104},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 336, col: 50, offset: 10114},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 340, col: 1, offset: 10150},
			expr: &actionExpr{
				pos: position{line: 340, col: 13, offset: 10164},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 340, col: 13, offset: 10164},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 340, col: 18, offset: 10169},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 340, col: 18, offset: 10169},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 340, col: 29, offset: 10180},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 340, col: 45, offset: 10196},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 347, col: 1, offset: 10321},
			expr: &actionExpr{
				pos: position{line: 347, col: 12, offset: 10334},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 347, col: 13, offset: 10335},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 347, col: 13, offset: 10335},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 22, offset: 10344},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 31, offset: 10353},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 39, offset: 10361},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 47, offset: 10369},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 55, offset: 10377},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 66, offset: 10388},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 77, offset: 10399},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 351, col: 1, offset: 10459},
			expr: &actionExpr{
				pos: position{line: 351, col: 17, offset: 10477},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 351, col: 17, offset: 10477},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 351, col: 22, offset: 10482},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 351, col: 22, offset: 10482},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 351, col: 32, offset: 10492},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 351, col: 42, offset: 10502},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 355, col: 1, offset: 10537},
			expr: &actionExpr{
				pos: position{line: 355, col: 11, offset: 10549},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 355, col: 11, offset: 10549},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 355, col: 11, offset: 10549},
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 11, offset: 10549},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 355, col: 20, offset: 10558},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 27, offset: 10565},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 355, col: 30, offset: 10568},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 34, offset: 10572},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 44, offset: 10582},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 355, col: 47, offset: 10585},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 51, offset: 10589},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 355, col: 54, offset: 10592},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 60, offset: 10598},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 70, offset: 10608},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 355, col: 73, offset: 10611},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 363, col: 1, offset: 10734},
			expr: &actionExpr{
				pos: position{line: 363, col: 11, offset: 10746},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 363, col: 11, offset: 10746},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 363, col: 11, offset: 10746},
							expr: &ruleRefExpr{
								pos:  position{line: 363, col: 11, offset: 10746},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 363, col: 20, offset: 10755},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 363, col: 27, offset: 10762},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 363, col: 30, offset: 10765},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 363, col: 34, offset: 10769},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 363, col: 44, offset: 10779},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 363, col: 47, offset: 10782},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 370, col: 1, offset: 10873},
			expr: &actionExpr{
				pos: position{line: 370, col: 12, offset: 10886},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 370, col: 12, offset: 10886},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 370, col: 12, offset: 10886},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 20, offset: 10894},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 370, col: 23, offset: 10897},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 370, col: 27, offset: 10901},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 37, offset: 10911},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 370, col: 40, offset: 10914},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 377, col: 1, offset: 11006},
			expr: &actionExpr{
				pos: position{line: 377, col: 11, offset: 11018},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 377, col: 11, offset: 11018},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 377, col: 11, offset: 11018},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 377, col: 22, offset: 11029},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 377, col: 30, offset: 11037},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 381, col: 1, offset: 11074},
			expr: &choiceExpr{
				pos: position{line: 381, col: 14, offset: 11089},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 381, col: 14, offset: 11089},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 24, offset: 11099},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 41, offset: 11116},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 55, offset: 11130},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 66, offset: 11141},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 78, offset: 11153},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 383, col: 1, offset: 11165},
			expr: &actionExpr{
				pos: position{line: 383, col: 15, offset: 11181},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 383, col: 15, offset: 11181},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 383, col: 15, offset: 11181},
							expr: &charClassMatcher{
								pos:        position{line: 383, col: 15, offset: 11181},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 383, col: 21, offset: 11187},
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 21, offset: 11187},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 387, col: 1, offset: 11251},
			expr: &actionExpr{
				pos: position{line: 387, col: 18, offset: 11270},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 387, col: 18, offset: 11270},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 387, col: 18, offset: 11270},
							expr: &charClassMatcher{
								pos:        position{line: 387, col: 18, offset: 11270},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 387, col: 24, offset: 11276},
							expr: &ruleRefExpr{
								pos:  position{line: 387, col: 24, offset: 11276},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 387, col: 31, offset: 11283},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 387, col: 35, offset: 11287},
							expr: &ruleRefExpr{
								pos:  position{line: 387, col: 35, offset: 11287},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 387, col: 42, offset: 11294},
							expr: &seqExpr{
								pos: position{line: 387, col: 44, offset: 11296},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 387, col: 44, offset: 11296},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 387, col: 51, offset: 11303},
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
			pos:  position{line: 391, col: 1, offset: 11373},
			expr: &actionExpr{
				pos: position{line: 391, col: 13, offset: 11387},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 391, col: 13, offset: 11387},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 391, col: 13, offset: 11387},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 17, offset: 11391},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 391, col: 20, offset: 11394},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 391, col: 27, offset: 11401},
								expr: &seqExpr{
									pos: position{line: 391, col: 28, offset: 11402},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 391, col: 28, offset: 11402},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 391, col: 39, offset: 11413},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 391, col: 42, offset: 11416},
											expr: &ruleRefExpr{
												pos:  position{line: 391, col: 42, offset: 11416},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 391, col: 57, offset: 11431},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 62, offset: 11436},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 391, col: 65, offset: 11439},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 400, col: 1, offset: 11633},
			expr: &actionExpr{
				pos: position{line: 400, col: 12, offset: 11646},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 400, col: 12, offset: 11646},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 400, col: 12, offset: 11646},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 16, offset: 11650},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 400, col: 19, offset: 11653},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 400, col: 26, offset: 11660},
								expr: &seqExpr{
									pos: position{line: 400, col: 27, offset: 11661},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 400, col: 27, offset: 11661},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 38, offset: 11672},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 400, col: 41, offset: 11675},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 45, offset: 11679},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 48, offset: 11682},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 59, offset: 11693},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 400, col: 63, offset: 11697},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 400, col: 63, offset: 11697},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 400, col: 69, offset: 11703},
													expr: &litMatcher{
														pos:        position{line: 400, col: 70, offset: 11704},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 75, offset: 11709},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 400, col: 80, offset: 11714},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 420, col: 1, offset: 12264},
			expr: &ruleRefExpr{
				pos:  position{line: 420, col: 19, offset: 12284},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 422, col: 1, offset: 12291},
			expr: &actionExpr{
				pos: position{line: 422, col: 9, offset: 12301},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 422, col: 9, offset: 12301},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 422, col: 9, offset: 12301},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 422, col: 16, offset: 12308},
								expr: &seqExpr{
									pos: position{line: 422, col: 17, offset: 12309},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 422, col: 17, offset: 12309},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 422, col: 27, offset: 12319},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 422, col: 32, offset: 12324},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 40, offset: 12332},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 422, col: 43, offset: 12335},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 422, col: 48, offset: 12340},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 59, offset: 12351},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 422, col: 62, offset: 12354},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 422, col: 69, offset: 12361},
								expr: &ruleRefExpr{
									pos:  position{line: 422, col: 69, offset: 12361},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 77, offset: 12369},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 422, col: 80, offset: 12372},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 84, offset: 12376},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 422, col: 87, offset: 12379},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 422, col: 98, offset: 12390},
								expr: &seqExpr{
									pos: position{line: 422, col: 99, offset: 12391},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 422, col: 99, offset: 12391},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 422, col: 109, offset: 12401},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 422, col: 115, offset: 12407},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 422, col: 115, offset: 12407},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 422, col: 121, offset: 12413},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 138, offset: 12430},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 443, col: 1, offset: 12975},
			expr: &actionExpr{
				pos: position{line: 443, col: 19, offset: 12995},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 443, col: 19, offset: 12995,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 447, col: 1, offset: 13062},
			expr: &actionExpr{
				pos: position{line: 447, col: 10, offset: 13073},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 447, col: 10, offset: 13073},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 447, col: 10, offset: 13073},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 19, offset: 13082},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 22, offset: 13085},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 447, col: 34, offset: 13097},
							expr: &seqExpr{
								pos: position{line: 447, col: 35, offset: 13098},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 447, col: 35, offset: 13098},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 447, col: 39, offset: 13102},
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
			pos:  position{line: 452, col: 1, offset: 13233},
			expr: &choiceExpr{
				pos: position{line: 452, col: 15, offset: 13249},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 452, col: 16, offset: 13250},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 452, col: 16, offset: 13250},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 452, col: 20, offset: 13254},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 452, col: 31, offset: 13265},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 452, col: 38, offset: 13272},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 454, col: 1, offset: 13284},
			expr: &oneOrMoreExpr{
				pos: position{line: 454, col: 14, offset: 13299},
				expr: &charClassMatcher{
					pos:        position{line: 454, col: 14, offset: 13299},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 456, col: 1, offset: 13317},
			expr: &actionExpr{
				pos: position{line: 456, col: 13, offset: 13331},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 456, col: 13, offset: 13331},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 456, col: 13, offset: 13331},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 456, col: 20, offset: 13338},
								expr: &seqExpr{
									pos: position{line: 456, col: 21, offset: 13339},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 456, col: 21, offset: 13339},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 456, col: 31, offset: 13349},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 456, col: 36, offset: 13354},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 456, col: 41, offset: 13359},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 52, offset: 13370},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 456, col: 54, offset: 13372},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 58, offset: 13376},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 456, col: 61, offset: 13379},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 456, col: 65, offset: 13383},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 76, offset: 13394},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 472, col: 1, offset: 13905},
			expr: &actionExpr{
				pos: position{line: 472, col: 11, offset: 13917},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 472, col: 12, offset: 13918},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 472, col: 13, offset: 13919},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 472, col: 13, offset: 13919},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 472, col: 17, offset: 13923},
									expr: &choiceExpr{
										pos: position{line: 472, col: 18, offset: 13924},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 472, col: 18, offset: 13924},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 472, col: 25, offset: 13931},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 472, col: 32, offset: 13938},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 472, col: 40, offset: 13946},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 472, col: 40, offset: 13946},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 472, col: 45, offset: 13951},
									expr: &choiceExpr{
										pos: position{line: 472, col: 46, offset: 13952},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 472, col: 46, offset: 13952},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 472, col: 53, offset: 13959},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 472, col: 60, offset: 13966},
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
			pos:  position{line: 479, col: 1, offset: 14182},
			expr: &actionExpr{
				pos: position{line: 479, col: 14, offset: 14197},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 479, col: 14, offset: 14197},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 479, col: 14, offset: 14197},
							expr: &choiceExpr{
								pos: position{line: 479, col: 15, offset: 14198},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 479, col: 15, offset: 14198},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 479, col: 24, offset: 14207},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 479, col: 30, offset: 14213},
							expr: &choiceExpr{
								pos: position{line: 479, col: 31, offset: 14214},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 479, col: 31, offset: 14214},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 479, col: 40, offset: 14223},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 479, col: 48, offset: 14231},
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
			pos:  position{line: 483, col: 1, offset: 14286},
			expr: &charClassMatcher{
				pos:        position{line: 483, col: 17, offset: 14304},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 484, col: 1, offset: 14309},
			expr: &charClassMatcher{
				pos:        position{line: 484, col: 10, offset: 14320},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 485, col: 1, offset: 14329},
			expr: &charClassMatcher{
				pos:        position{line: 485, col: 9, offset: 14339},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 487, col: 1, offset: 14346},
			expr: &anyMatcher{
				line: 487, col: 14, offset: 14361,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 488, col: 1, offset: 14363},
			expr: &actionExpr{
				pos: position{line: 488, col: 13, offset: 14377},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 488, col: 13, offset: 14377},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 488, col: 13, offset: 14377},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 488, col: 20, offset: 14384},
							expr: &seqExpr{
								pos: position{line: 488, col: 22, offset: 14386},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 488, col: 22, offset: 14386},
										expr: &litMatcher{
											pos:        position{line: 488, col: 23, offset: 14387},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 488, col: 28, offset: 14392},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 488, col: 42, offset: 14406},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 494, col: 1, offset: 14586},
			expr: &choiceExpr{
				pos: position{line: 494, col: 11, offset: 14598},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 494, col: 11, offset: 14598},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 494, col: 30, offset: 14617},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 495, col: 1, offset: 14635},
			expr: &seqExpr{
				pos: position{line: 495, col: 20, offset: 14656},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 495, col: 20, offset: 14656},
						expr: &ruleRefExpr{
							pos:  position{line: 495, col: 21, offset: 14657},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 495, col: 31, offset: 14667},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 495, col: 36, offset: 14672},
						expr: &seqExpr{
							pos: position{line: 495, col: 38, offset: 14674},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 495, col: 38, offset: 14674},
									expr: &litMatcher{
										pos:        position{line: 495, col: 39, offset: 14675},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 495, col: 44, offset: 14680},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 495, col: 58, offset: 14694},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 496, col: 1, offset: 14699},
			expr: &seqExpr{
				pos: position{line: 496, col: 36, offset: 14736},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 496, col: 36, offset: 14736},
						expr: &ruleRefExpr{
							pos:  position{line: 496, col: 37, offset: 14737},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 496, col: 47, offset: 14747},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 496, col: 52, offset: 14752},
						expr: &seqExpr{
							pos: position{line: 496, col: 54, offset: 14754},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 496, col: 54, offset: 14754},
									expr: &choiceExpr{
										pos: position{line: 496, col: 57, offset: 14757},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 496, col: 57, offset: 14757},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 496, col: 64, offset: 14764},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 496, col: 70, offset: 14770},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 496, col: 84, offset: 14784},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 497, col: 1, offset: 14789},
			expr: &choiceExpr{
				pos: position{line: 497, col: 21, offset: 14811},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 497, col: 22, offset: 14812},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 497, col: 22, offset: 14812},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 497, col: 27, offset: 14817},
								expr: &seqExpr{
									pos: position{line: 497, col: 29, offset: 14819},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 497, col: 29, offset: 14819},
											expr: &ruleRefExpr{
												pos:  position{line: 497, col: 30, offset: 14820},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 497, col: 34, offset: 14824},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 497, col: 52, offset: 14842},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 497, col: 52, offset: 14842},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 497, col: 56, offset: 14846},
								expr: &seqExpr{
									pos: position{line: 497, col: 58, offset: 14848},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 497, col: 58, offset: 14848},
											expr: &ruleRefExpr{
												pos:  position{line: 497, col: 59, offset: 14849},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 497, col: 63, offset: 14853},
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
			pos:  position{line: 499, col: 1, offset: 14869},
			expr: &zeroOrMoreExpr{
				pos: position{line: 499, col: 6, offset: 14876},
				expr: &choiceExpr{
					pos: position{line: 499, col: 8, offset: 14878},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 499, col: 8, offset: 14878},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 499, col: 21, offset: 14891},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 499, col: 27, offset: 14897},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 500, col: 1, offset: 14908},
			expr: &zeroOrMoreExpr{
				pos: position{line: 500, col: 5, offset: 14914},
				expr: &choiceExpr{
					pos: position{line: 500, col: 7, offset: 14916},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 500, col: 7, offset: 14916},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 500, col: 20, offset: 14929},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 501, col: 1, offset: 14965},
			expr: &zeroOrMoreExpr{
				pos: position{line: 501, col: 6, offset: 14972},
				expr: &ruleRefExpr{
					pos:  position{line: 501, col: 6, offset: 14972},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 503, col: 1, offset: 14985},
			expr: &charClassMatcher{
				pos:        position{line: 503, col: 14, offset: 15000},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 504, col: 1, offset: 15008},
			expr: &litMatcher{
				pos:        position{line: 504, col: 7, offset: 15016},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 505, col: 1, offset: 15021},
			expr: &choiceExpr{
				pos: position{line: 505, col: 7, offset: 15029},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 505, col: 7, offset: 15029},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 505, col: 7, offset: 15029},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 505, col: 10, offset: 15032},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 505, col: 16, offset: 15038},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 505, col: 16, offset: 15038},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 505, col: 18, offset: 15040},
								expr: &ruleRefExpr{
									pos:  position{line: 505, col: 18, offset: 15040},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 505, col: 37, offset: 15059},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 505, col: 43, offset: 15065},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 505, col: 43, offset: 15065},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 505, col: 46, offset: 15068},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 507, col: 1, offset: 15073},
			expr: &notExpr{
				pos: position{line: 507, col: 7, offset: 15081},
				expr: &anyMatcher{
					line: 507, col: 8, offset: 15082,
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
