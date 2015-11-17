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
			pos:  position{line: 85, col: 1, offset: 2739},
			expr: &actionExpr{
				pos: position{line: 85, col: 11, offset: 2751},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 85, col: 11, offset: 2751},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 85, col: 11, offset: 2751},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 14, offset: 2754},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 85, col: 25, offset: 2765},
								expr: &seqExpr{
									pos: position{line: 85, col: 27, offset: 2767},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 85, col: 27, offset: 2767},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 85, col: 37, offset: 2777},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 85, col: 44, offset: 2784},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 85, col: 44, offset: 2784},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 85, col: 50, offset: 2790},
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
			pos:  position{line: 147, col: 1, offset: 5363},
			expr: &actionExpr{
				pos: position{line: 147, col: 15, offset: 5379},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 147, col: 15, offset: 5379,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 151, col: 1, offset: 5441},
			expr: &actionExpr{
				pos: position{line: 151, col: 13, offset: 5455},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 151, col: 13, offset: 5455},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 151, col: 13, offset: 5455},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 151, col: 20, offset: 5462},
								expr: &seqExpr{
									pos: position{line: 151, col: 21, offset: 5463},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 151, col: 21, offset: 5463},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 151, col: 31, offset: 5473},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 151, col: 36, offset: 5478},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 151, col: 47, offset: 5489},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 151, col: 47, offset: 5489},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 151, col: 65, offset: 5507},
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
			pos:  position{line: 164, col: 1, offset: 6010},
			expr: &choiceExpr{
				pos: position{line: 164, col: 19, offset: 6030},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 164, col: 19, offset: 6030},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 29, offset: 6040},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 41, offset: 6052},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 49, offset: 6060},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 56, offset: 6067},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 66, offset: 6077},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 75, offset: 6086},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 87, offset: 6098},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 95, offset: 6106},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 166, col: 1, offset: 6115},
			expr: &actionExpr{
				pos: position{line: 166, col: 11, offset: 6127},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 166, col: 11, offset: 6127},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 166, col: 11, offset: 6127},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 166, col: 21, offset: 6137},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 166, col: 23, offset: 6139},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 166, col: 28, offset: 6144},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 166, col: 36, offset: 6152},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 170, col: 1, offset: 6204},
			expr: &actionExpr{
				pos: position{line: 170, col: 13, offset: 6218},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 170, col: 13, offset: 6218},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 170, col: 13, offset: 6218},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 25, offset: 6230},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 27, offset: 6232},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 170, col: 33, offset: 6238},
								expr: &charClassMatcher{
									pos:        position{line: 170, col: 33, offset: 6238},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 42, offset: 6247},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 44, offset: 6249},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 170, col: 47, offset: 6252},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 58, offset: 6263},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 177, col: 1, offset: 6420},
			expr: &actionExpr{
				pos: position{line: 177, col: 9, offset: 6430},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 177, col: 9, offset: 6430},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 177, col: 9, offset: 6430},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 17, offset: 6438},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 19, offset: 6440},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 23, offset: 6444},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 33, offset: 6454},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 35, offset: 6456},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 40, offset: 6461},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 51, offset: 6472},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 177, col: 53, offset: 6474},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 57, offset: 6478},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 59, offset: 6480},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 65, offset: 6486},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 76, offset: 6497},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 185, col: 1, offset: 6629},
			expr: &actionExpr{
				pos: position{line: 185, col: 8, offset: 6638},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 185, col: 8, offset: 6638},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 185, col: 8, offset: 6638},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 15, offset: 6645},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 185, col: 17, offset: 6647},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 185, col: 22, offset: 6652},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 33, offset: 6663},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 185, col: 36, offset: 6666},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 40, offset: 6670},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 185, col: 43, offset: 6673},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 185, col: 50, offset: 6680},
								expr: &seqExpr{
									pos: position{line: 185, col: 51, offset: 6681},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 185, col: 51, offset: 6681},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 185, col: 61, offset: 6691},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 185, col: 66, offset: 6696},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 70, offset: 6700},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 208, col: 1, offset: 7312},
			expr: &actionExpr{
				pos: position{line: 208, col: 13, offset: 7326},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 208, col: 13, offset: 7326},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 208, col: 13, offset: 7326},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 208, col: 18, offset: 7331},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 208, col: 29, offset: 7342},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 208, col: 31, offset: 7344},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 208, col: 37, offset: 7350},
								expr: &seqExpr{
									pos: position{line: 208, col: 38, offset: 7351},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 208, col: 38, offset: 7351},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 42, offset: 7355},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 44, offset: 7357},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 208, col: 58, offset: 7371},
							expr: &ruleRefExpr{
								pos:  position{line: 208, col: 58, offset: 7371},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 219, col: 1, offset: 7583},
			expr: &actionExpr{
				pos: position{line: 219, col: 11, offset: 7595},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 219, col: 11, offset: 7595},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 219, col: 11, offset: 7595},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 219, col: 21, offset: 7605},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 219, col: 23, offset: 7607},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 219, col: 27, offset: 7611},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 219, col: 37, offset: 7621},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 219, col: 39, offset: 7623},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 219, col: 44, offset: 7628},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 219, col: 55, offset: 7639},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 226, col: 1, offset: 7748},
			expr: &actionExpr{
				pos: position{line: 226, col: 10, offset: 7759},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 226, col: 10, offset: 7759},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 226, col: 10, offset: 7759},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 226, col: 19, offset: 7768},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 226, col: 21, offset: 7770},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 226, col: 24, offset: 7773},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 227, col: 1, offset: 7813},
			expr: &actionExpr{
				pos: position{line: 227, col: 13, offset: 7827},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 227, col: 13, offset: 7827},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 227, col: 13, offset: 7827},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 227, col: 25, offset: 7839},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 227, col: 27, offset: 7841},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 227, col: 30, offset: 7844},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 228, col: 1, offset: 7895},
			expr: &actionExpr{
				pos: position{line: 228, col: 9, offset: 7905},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 228, col: 9, offset: 7905},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 9, offset: 7905},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 17, offset: 7913},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 19, offset: 7915},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 22, offset: 7918},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 229, col: 1, offset: 7965},
			expr: &actionExpr{
				pos: position{line: 229, col: 14, offset: 7980},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 229, col: 14, offset: 7980},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 229, col: 14, offset: 7980},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 19, offset: 7985},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 30, offset: 7996},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 229, col: 33, offset: 7999},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 37, offset: 8003},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 229, col: 40, offset: 8006},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 47, offset: 8013},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 229, col: 57, offset: 8023},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 61, offset: 8027},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 239, col: 1, offset: 8188},
			expr: &actionExpr{
				pos: position{line: 239, col: 13, offset: 8202},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 239, col: 13, offset: 8202},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 239, col: 20, offset: 8209},
						expr: &seqExpr{
							pos: position{line: 239, col: 21, offset: 8210},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 239, col: 21, offset: 8210},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 239, col: 27, offset: 8216},
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
			pos:  position{line: 248, col: 1, offset: 8397},
			expr: &actionExpr{
				pos: position{line: 248, col: 9, offset: 8407},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 248, col: 9, offset: 8407},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 248, col: 9, offset: 8407},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 248, col: 16, offset: 8414},
								expr: &seqExpr{
									pos: position{line: 248, col: 17, offset: 8415},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 248, col: 17, offset: 8415},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 248, col: 27, offset: 8425},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 248, col: 32, offset: 8430},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 35, offset: 8433},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 47, offset: 8445},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 248, col: 49, offset: 8447},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 53, offset: 8451},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 55, offset: 8453},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 248, col: 59, offset: 8457},
								expr: &ruleRefExpr{
									pos:  position{line: 248, col: 59, offset: 8457},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 69, offset: 8467},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 71, offset: 8469},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 75, offset: 8473},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 85, offset: 8483},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 87, offset: 8485},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 92, offset: 8490},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 103, offset: 8501},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 106, offset: 8504},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 248, col: 110, offset: 8508},
								expr: &seqExpr{
									pos: position{line: 248, col: 111, offset: 8509},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 248, col: 111, offset: 8509},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 248, col: 115, offset: 8513},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 248, col: 117, offset: 8515},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 248, col: 130, offset: 8528},
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 130, offset: 8528},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 267, col: 1, offset: 8962},
			expr: &actionExpr{
				pos: position{line: 267, col: 12, offset: 8975},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 267, col: 13, offset: 8976},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 267, col: 13, offset: 8976},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 267, col: 26, offset: 8989},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 271, col: 1, offset: 9063},
			expr: &actionExpr{
				pos: position{line: 271, col: 11, offset: 9075},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 271, col: 11, offset: 9075},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 271, col: 11, offset: 9075},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 21, offset: 9085},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 23, offset: 9087},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 271, col: 28, offset: 9092},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 39, offset: 9103},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 41, offset: 9105},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 271, col: 49, offset: 9113},
								expr: &seqExpr{
									pos: position{line: 271, col: 50, offset: 9114},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 271, col: 50, offset: 9114},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 60, offset: 9124},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 63, offset: 9127},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 74, offset: 9138},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 79, offset: 9143},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 271, col: 82, offset: 9146},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 86, offset: 9150},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 89, offset: 9153},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 271, col: 97, offset: 9161},
								expr: &seqExpr{
									pos: position{line: 271, col: 98, offset: 9162},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 271, col: 98, offset: 9162},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 107, offset: 9171},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 271, col: 113, offset: 9177},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 271, col: 113, offset: 9177},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 119, offset: 9183},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 138, offset: 9202},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 286, col: 1, offset: 9597},
			expr: &actionExpr{
				pos: position{line: 286, col: 21, offset: 9619},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 286, col: 21, offset: 9619,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 290, col: 1, offset: 9688},
			expr: &actionExpr{
				pos: position{line: 290, col: 12, offset: 9701},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 290, col: 12, offset: 9701},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 290, col: 12, offset: 9701},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 290, col: 19, offset: 9708},
								expr: &seqExpr{
									pos: position{line: 290, col: 20, offset: 9709},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 290, col: 20, offset: 9709},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 290, col: 30, offset: 9719},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 290, col: 35, offset: 9724},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 290, col: 42, offset: 9731},
								expr: &seqExpr{
									pos: position{line: 290, col: 43, offset: 9732},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 290, col: 43, offset: 9732},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 290, col: 52, offset: 9741},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 290, col: 57, offset: 9746},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 290, col: 61, offset: 9750},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 290, col: 74, offset: 9763},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 290, col: 77, offset: 9766},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 290, col: 82, offset: 9771},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 290, col: 93, offset: 9782},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 290, col: 95, offset: 9784},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 290, col: 99, offset: 9788},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 290, col: 102, offset: 9791},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 290, col: 112, offset: 9801},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 290, col: 122, offset: 9811},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 290, col: 126, offset: 9815},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 290, col: 129, offset: 9818},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 290, col: 140, offset: 9829},
								expr: &ruleRefExpr{
									pos:  position{line: 290, col: 140, offset: 9829},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 290, col: 148, offset: 9837},
							expr: &ruleRefExpr{
								pos:  position{line: 290, col: 148, offset: 9837},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 317, col: 1, offset: 10428},
			expr: &actionExpr{
				pos: position{line: 317, col: 16, offset: 10445},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 317, col: 16, offset: 10445},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 317, col: 21, offset: 10450},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 317, col: 21, offset: 10450},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 317, col: 30, offset: 10459},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 324, col: 1, offset: 10581},
			expr: &actionExpr{
				pos: position{line: 324, col: 10, offset: 10592},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 324, col: 10, offset: 10592},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 324, col: 10, offset: 10592},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 19, offset: 10601},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 324, col: 22, offset: 10604},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 26, offset: 10608},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 29, offset: 10611},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 40, offset: 10622},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 324, col: 50, offset: 10632},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 328, col: 1, offset: 10668},
			expr: &actionExpr{
				pos: position{line: 328, col: 13, offset: 10682},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 328, col: 13, offset: 10682},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 328, col: 18, offset: 10687},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 328, col: 18, offset: 10687},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 328, col: 29, offset: 10698},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 328, col: 45, offset: 10714},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "DefinitionType",
			pos:  position{line: 335, col: 1, offset: 10839},
			expr: &actionExpr{
				pos: position{line: 335, col: 18, offset: 10858},
				run: (*parser).callonDefinitionType1,
				expr: &labeledExpr{
					pos:   position{line: 335, col: 18, offset: 10858},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 335, col: 23, offset: 10863},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 335, col: 23, offset: 10863},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 335, col: 34, offset: 10874},
								name: "ContainerType",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 339, col: 1, offset: 10914},
			expr: &actionExpr{
				pos: position{line: 339, col: 12, offset: 10927},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 339, col: 13, offset: 10928},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 339, col: 13, offset: 10928},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 22, offset: 10937},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 31, offset: 10946},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 39, offset: 10954},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 47, offset: 10962},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 55, offset: 10970},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 66, offset: 10981},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 77, offset: 10992},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 343, col: 1, offset: 11052},
			expr: &actionExpr{
				pos: position{line: 343, col: 17, offset: 11070},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 343, col: 17, offset: 11070},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 343, col: 22, offset: 11075},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 343, col: 22, offset: 11075},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 343, col: 32, offset: 11085},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 343, col: 42, offset: 11095},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 347, col: 1, offset: 11130},
			expr: &actionExpr{
				pos: position{line: 347, col: 11, offset: 11142},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 347, col: 11, offset: 11142},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 347, col: 11, offset: 11142},
							expr: &ruleRefExpr{
								pos:  position{line: 347, col: 11, offset: 11142},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 347, col: 20, offset: 11151},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 27, offset: 11158},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 347, col: 30, offset: 11161},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 347, col: 34, offset: 11165},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 44, offset: 11175},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 347, col: 47, offset: 11178},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 51, offset: 11182},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 347, col: 54, offset: 11185},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 347, col: 60, offset: 11191},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 70, offset: 11201},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 347, col: 73, offset: 11204},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 355, col: 1, offset: 11327},
			expr: &actionExpr{
				pos: position{line: 355, col: 11, offset: 11339},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 355, col: 11, offset: 11339},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 355, col: 11, offset: 11339},
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 11, offset: 11339},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 355, col: 20, offset: 11348},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 27, offset: 11355},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 355, col: 30, offset: 11358},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 34, offset: 11362},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 44, offset: 11372},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 355, col: 47, offset: 11375},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 362, col: 1, offset: 11466},
			expr: &actionExpr{
				pos: position{line: 362, col: 12, offset: 11479},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 362, col: 12, offset: 11479},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 362, col: 12, offset: 11479},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 362, col: 20, offset: 11487},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 362, col: 23, offset: 11490},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 362, col: 27, offset: 11494},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 362, col: 37, offset: 11504},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 362, col: 40, offset: 11507},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 369, col: 1, offset: 11599},
			expr: &actionExpr{
				pos: position{line: 369, col: 11, offset: 11611},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 369, col: 11, offset: 11611},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 369, col: 11, offset: 11611},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 369, col: 22, offset: 11622},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 369, col: 30, offset: 11630},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 373, col: 1, offset: 11667},
			expr: &choiceExpr{
				pos: position{line: 373, col: 14, offset: 11682},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 373, col: 14, offset: 11682},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 24, offset: 11692},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 41, offset: 11709},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 55, offset: 11723},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 66, offset: 11734},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 78, offset: 11746},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 375, col: 1, offset: 11758},
			expr: &actionExpr{
				pos: position{line: 375, col: 15, offset: 11774},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 375, col: 15, offset: 11774},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 375, col: 15, offset: 11774},
							expr: &charClassMatcher{
								pos:        position{line: 375, col: 15, offset: 11774},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 375, col: 21, offset: 11780},
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 21, offset: 11780},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 379, col: 1, offset: 11844},
			expr: &actionExpr{
				pos: position{line: 379, col: 18, offset: 11863},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 379, col: 18, offset: 11863},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 379, col: 18, offset: 11863},
							expr: &charClassMatcher{
								pos:        position{line: 379, col: 18, offset: 11863},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 379, col: 24, offset: 11869},
							expr: &ruleRefExpr{
								pos:  position{line: 379, col: 24, offset: 11869},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 379, col: 31, offset: 11876},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 379, col: 35, offset: 11880},
							expr: &ruleRefExpr{
								pos:  position{line: 379, col: 35, offset: 11880},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 379, col: 42, offset: 11887},
							expr: &seqExpr{
								pos: position{line: 379, col: 44, offset: 11889},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 379, col: 44, offset: 11889},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 379, col: 51, offset: 11896},
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
			pos:  position{line: 383, col: 1, offset: 11966},
			expr: &actionExpr{
				pos: position{line: 383, col: 13, offset: 11980},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 383, col: 13, offset: 11980},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 383, col: 13, offset: 11980},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 17, offset: 11984},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 383, col: 20, offset: 11987},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 383, col: 27, offset: 11994},
								expr: &seqExpr{
									pos: position{line: 383, col: 28, offset: 11995},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 383, col: 28, offset: 11995},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 383, col: 39, offset: 12006},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 383, col: 42, offset: 12009},
											expr: &ruleRefExpr{
												pos:  position{line: 383, col: 42, offset: 12009},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 383, col: 57, offset: 12024},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 62, offset: 12029},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 383, col: 65, offset: 12032},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 392, col: 1, offset: 12226},
			expr: &actionExpr{
				pos: position{line: 392, col: 12, offset: 12239},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 392, col: 12, offset: 12239},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 392, col: 12, offset: 12239},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 392, col: 16, offset: 12243},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 392, col: 19, offset: 12246},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 392, col: 26, offset: 12253},
								expr: &seqExpr{
									pos: position{line: 392, col: 27, offset: 12254},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 392, col: 27, offset: 12254},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 38, offset: 12265},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 392, col: 41, offset: 12268},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 45, offset: 12272},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 48, offset: 12275},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 59, offset: 12286},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 392, col: 63, offset: 12290},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 392, col: 63, offset: 12290},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 392, col: 69, offset: 12296},
													expr: &litMatcher{
														pos:        position{line: 392, col: 70, offset: 12297},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 75, offset: 12302},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 392, col: 80, offset: 12307},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 412, col: 1, offset: 12857},
			expr: &ruleRefExpr{
				pos:  position{line: 412, col: 19, offset: 12877},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 414, col: 1, offset: 12884},
			expr: &actionExpr{
				pos: position{line: 414, col: 9, offset: 12894},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 414, col: 9, offset: 12894},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 414, col: 9, offset: 12894},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 414, col: 16, offset: 12901},
								expr: &seqExpr{
									pos: position{line: 414, col: 17, offset: 12902},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 414, col: 17, offset: 12902},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 414, col: 27, offset: 12912},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 414, col: 32, offset: 12917},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 40, offset: 12925},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 414, col: 43, offset: 12928},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 414, col: 48, offset: 12933},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 59, offset: 12944},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 414, col: 62, offset: 12947},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 66, offset: 12951},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 414, col: 69, offset: 12954},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 414, col: 76, offset: 12961},
								expr: &ruleRefExpr{
									pos:  position{line: 414, col: 76, offset: 12961},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 84, offset: 12969},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 414, col: 87, offset: 12972},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 414, col: 98, offset: 12983},
								expr: &seqExpr{
									pos: position{line: 414, col: 99, offset: 12984},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 414, col: 99, offset: 12984},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 414, col: 109, offset: 12994},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 414, col: 115, offset: 13000},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 414, col: 115, offset: 13000},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 414, col: 121, offset: 13006},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 138, offset: 13023},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 435, col: 1, offset: 13672},
			expr: &actionExpr{
				pos: position{line: 435, col: 19, offset: 13692},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 435, col: 19, offset: 13692,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 439, col: 1, offset: 13763},
			expr: &actionExpr{
				pos: position{line: 439, col: 10, offset: 13774},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 439, col: 10, offset: 13774},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 439, col: 10, offset: 13774},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 439, col: 19, offset: 13783},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 439, col: 21, offset: 13785},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 439, col: 26, offset: 13790},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 439, col: 34, offset: 13798},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 443, col: 1, offset: 13852},
			expr: &actionExpr{
				pos: position{line: 443, col: 13, offset: 13866},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 443, col: 13, offset: 13866},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 443, col: 13, offset: 13866},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 443, col: 20, offset: 13873},
								expr: &seqExpr{
									pos: position{line: 443, col: 21, offset: 13874},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 443, col: 21, offset: 13874},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 443, col: 31, offset: 13884},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 443, col: 36, offset: 13889},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 41, offset: 13894},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 443, col: 52, offset: 13905},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 443, col: 54, offset: 13907},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 443, col: 58, offset: 13911},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 443, col: 61, offset: 13914},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 67, offset: 13920},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 443, col: 78, offset: 13931},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 459, col: 1, offset: 14485},
			expr: &actionExpr{
				pos: position{line: 459, col: 11, offset: 14497},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 459, col: 12, offset: 14498},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 459, col: 13, offset: 14499},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 459, col: 13, offset: 14499},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 459, col: 17, offset: 14503},
									expr: &choiceExpr{
										pos: position{line: 459, col: 18, offset: 14504},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 459, col: 18, offset: 14504},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 459, col: 25, offset: 14511},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 459, col: 32, offset: 14518},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 459, col: 40, offset: 14526},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 459, col: 40, offset: 14526},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 459, col: 45, offset: 14531},
									expr: &choiceExpr{
										pos: position{line: 459, col: 46, offset: 14532},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 459, col: 46, offset: 14532},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 459, col: 53, offset: 14539},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 459, col: 60, offset: 14546},
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
			pos:  position{line: 466, col: 1, offset: 14782},
			expr: &actionExpr{
				pos: position{line: 466, col: 14, offset: 14797},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 466, col: 14, offset: 14797},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 466, col: 14, offset: 14797},
							expr: &choiceExpr{
								pos: position{line: 466, col: 15, offset: 14798},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 466, col: 15, offset: 14798},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 466, col: 24, offset: 14807},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 466, col: 30, offset: 14813},
							expr: &choiceExpr{
								pos: position{line: 466, col: 31, offset: 14814},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 466, col: 31, offset: 14814},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 466, col: 40, offset: 14823},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 466, col: 48, offset: 14831},
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
			pos:  position{line: 470, col: 1, offset: 14890},
			expr: &charClassMatcher{
				pos:        position{line: 470, col: 17, offset: 14908},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 471, col: 1, offset: 14913},
			expr: &charClassMatcher{
				pos:        position{line: 471, col: 10, offset: 14924},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 472, col: 1, offset: 14933},
			expr: &charClassMatcher{
				pos:        position{line: 472, col: 9, offset: 14943},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 474, col: 1, offset: 14950},
			expr: &anyMatcher{
				line: 474, col: 14, offset: 14965,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 475, col: 1, offset: 14967},
			expr: &actionExpr{
				pos: position{line: 475, col: 13, offset: 14981},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 475, col: 13, offset: 14981},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 475, col: 13, offset: 14981},
							val:        "/**",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 475, col: 19, offset: 14987},
							label: "comment",
							expr: &zeroOrMoreExpr{
								pos: position{line: 475, col: 27, offset: 14995},
								expr: &seqExpr{
									pos: position{line: 475, col: 29, offset: 14997},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 475, col: 29, offset: 14997},
											expr: &litMatcher{
												pos:        position{line: 475, col: 30, offset: 14998},
												val:        "*/",
												ignoreCase: false,
											},
										},
										&ruleRefExpr{
											pos:  position{line: 475, col: 35, offset: 15003},
											name: "SourceChar",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 475, col: 49, offset: 15017},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 482, col: 1, offset: 15265},
			expr: &choiceExpr{
				pos: position{line: 482, col: 11, offset: 15277},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 482, col: 11, offset: 15277},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 482, col: 30, offset: 15296},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 483, col: 1, offset: 15314},
			expr: &seqExpr{
				pos: position{line: 483, col: 20, offset: 15335},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 483, col: 20, offset: 15335},
						expr: &ruleRefExpr{
							pos:  position{line: 483, col: 21, offset: 15336},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 483, col: 31, offset: 15346},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 483, col: 36, offset: 15351},
						expr: &seqExpr{
							pos: position{line: 483, col: 38, offset: 15353},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 483, col: 38, offset: 15353},
									expr: &litMatcher{
										pos:        position{line: 483, col: 39, offset: 15354},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 483, col: 44, offset: 15359},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 483, col: 58, offset: 15373},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 484, col: 1, offset: 15378},
			expr: &seqExpr{
				pos: position{line: 484, col: 36, offset: 15415},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 484, col: 36, offset: 15415},
						expr: &ruleRefExpr{
							pos:  position{line: 484, col: 37, offset: 15416},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 484, col: 47, offset: 15426},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 484, col: 52, offset: 15431},
						expr: &seqExpr{
							pos: position{line: 484, col: 54, offset: 15433},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 484, col: 54, offset: 15433},
									expr: &choiceExpr{
										pos: position{line: 484, col: 57, offset: 15436},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 484, col: 57, offset: 15436},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 484, col: 64, offset: 15443},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 484, col: 70, offset: 15449},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 484, col: 84, offset: 15463},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 485, col: 1, offset: 15468},
			expr: &choiceExpr{
				pos: position{line: 485, col: 21, offset: 15490},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 485, col: 22, offset: 15491},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 485, col: 22, offset: 15491},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 485, col: 27, offset: 15496},
								expr: &seqExpr{
									pos: position{line: 485, col: 29, offset: 15498},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 485, col: 29, offset: 15498},
											expr: &ruleRefExpr{
												pos:  position{line: 485, col: 30, offset: 15499},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 485, col: 34, offset: 15503},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 485, col: 52, offset: 15521},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 485, col: 52, offset: 15521},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 485, col: 56, offset: 15525},
								expr: &seqExpr{
									pos: position{line: 485, col: 58, offset: 15527},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 485, col: 58, offset: 15527},
											expr: &ruleRefExpr{
												pos:  position{line: 485, col: 59, offset: 15528},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 485, col: 63, offset: 15532},
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
			pos:  position{line: 487, col: 1, offset: 15548},
			expr: &zeroOrMoreExpr{
				pos: position{line: 487, col: 6, offset: 15555},
				expr: &choiceExpr{
					pos: position{line: 487, col: 8, offset: 15557},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 487, col: 8, offset: 15557},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 487, col: 21, offset: 15570},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 487, col: 27, offset: 15576},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 488, col: 1, offset: 15587},
			expr: &zeroOrMoreExpr{
				pos: position{line: 488, col: 5, offset: 15593},
				expr: &choiceExpr{
					pos: position{line: 488, col: 7, offset: 15595},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 488, col: 7, offset: 15595},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 488, col: 20, offset: 15608},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 489, col: 1, offset: 15644},
			expr: &zeroOrMoreExpr{
				pos: position{line: 489, col: 6, offset: 15651},
				expr: &ruleRefExpr{
					pos:  position{line: 489, col: 6, offset: 15651},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 491, col: 1, offset: 15664},
			expr: &charClassMatcher{
				pos:        position{line: 491, col: 14, offset: 15679},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 492, col: 1, offset: 15687},
			expr: &litMatcher{
				pos:        position{line: 492, col: 7, offset: 15695},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 493, col: 1, offset: 15700},
			expr: &choiceExpr{
				pos: position{line: 493, col: 7, offset: 15708},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 493, col: 7, offset: 15708},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 493, col: 7, offset: 15708},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 493, col: 10, offset: 15711},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 493, col: 16, offset: 15717},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 493, col: 16, offset: 15717},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 493, col: 18, offset: 15719},
								expr: &ruleRefExpr{
									pos:  position{line: 493, col: 18, offset: 15719},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 493, col: 37, offset: 15738},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 493, col: 43, offset: 15744},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 493, col: 43, offset: 15744},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 493, col: 46, offset: 15747},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 495, col: 1, offset: 15752},
			expr: &notExpr{
				pos: position{line: 495, col: 7, offset: 15760},
				expr: &anyMatcher{
					line: 495, col: 8, offset: 15761,
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
		Thrift: thrift,
		Scopes: []*Scope{},
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

func (c *current) onDocString1(comment interface{}) (interface{}, error) {
	chars := make([]byte, len(comment.([]interface{})))
	for i, arr := range comment.([]interface{}) {
		chars[i] = arr.([]interface{})[1].([]byte)[0]
	}
	return strings.TrimSpace(string(chars)), nil
}

func (p *parser) callonDocString1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDocString1(stack["comment"])
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
