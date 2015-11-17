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
			pos:  position{line: 147, col: 1, offset: 4367},
			expr: &actionExpr{
				pos: position{line: 147, col: 15, offset: 4383},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 147, col: 15, offset: 4383,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 151, col: 1, offset: 4441},
			expr: &actionExpr{
				pos: position{line: 151, col: 13, offset: 4455},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 151, col: 13, offset: 4455},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 151, col: 13, offset: 4455},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 151, col: 20, offset: 4462},
								expr: &seqExpr{
									pos: position{line: 151, col: 21, offset: 4463},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 151, col: 21, offset: 4463},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 151, col: 31, offset: 4473},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 151, col: 36, offset: 4478},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 151, col: 47, offset: 4489},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 151, col: 47, offset: 4489},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 151, col: 65, offset: 4507},
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
			pos:  position{line: 164, col: 1, offset: 4978},
			expr: &choiceExpr{
				pos: position{line: 164, col: 19, offset: 4998},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 164, col: 19, offset: 4998},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 29, offset: 5008},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 41, offset: 5020},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 49, offset: 5028},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 56, offset: 5035},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 66, offset: 5045},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 75, offset: 5054},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 87, offset: 5066},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 164, col: 95, offset: 5074},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 166, col: 1, offset: 5083},
			expr: &actionExpr{
				pos: position{line: 166, col: 11, offset: 5095},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 166, col: 11, offset: 5095},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 166, col: 11, offset: 5095},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 166, col: 21, offset: 5105},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 166, col: 23, offset: 5107},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 166, col: 28, offset: 5112},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 166, col: 36, offset: 5120},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 170, col: 1, offset: 5168},
			expr: &actionExpr{
				pos: position{line: 170, col: 13, offset: 5182},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 170, col: 13, offset: 5182},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 170, col: 13, offset: 5182},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 25, offset: 5194},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 27, offset: 5196},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 170, col: 33, offset: 5202},
								expr: &charClassMatcher{
									pos:        position{line: 170, col: 33, offset: 5202},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 42, offset: 5211},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 44, offset: 5213},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 170, col: 47, offset: 5216},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 58, offset: 5227},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 177, col: 1, offset: 5360},
			expr: &actionExpr{
				pos: position{line: 177, col: 9, offset: 5370},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 177, col: 9, offset: 5370},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 177, col: 9, offset: 5370},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 17, offset: 5378},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 19, offset: 5380},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 23, offset: 5384},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 33, offset: 5394},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 35, offset: 5396},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 40, offset: 5401},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 51, offset: 5412},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 177, col: 53, offset: 5414},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 57, offset: 5418},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 177, col: 59, offset: 5420},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 177, col: 65, offset: 5426},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 177, col: 76, offset: 5437},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 185, col: 1, offset: 5569},
			expr: &actionExpr{
				pos: position{line: 185, col: 8, offset: 5578},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 185, col: 8, offset: 5578},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 185, col: 8, offset: 5578},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 15, offset: 5585},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 185, col: 17, offset: 5587},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 185, col: 22, offset: 5592},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 33, offset: 5603},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 185, col: 36, offset: 5606},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 40, offset: 5610},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 185, col: 43, offset: 5613},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 185, col: 50, offset: 5620},
								expr: &seqExpr{
									pos: position{line: 185, col: 51, offset: 5621},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 185, col: 51, offset: 5621},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 185, col: 61, offset: 5631},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 185, col: 66, offset: 5636},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 185, col: 70, offset: 5640},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 208, col: 1, offset: 6252},
			expr: &actionExpr{
				pos: position{line: 208, col: 13, offset: 6266},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 208, col: 13, offset: 6266},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 208, col: 13, offset: 6266},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 208, col: 18, offset: 6271},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 208, col: 29, offset: 6282},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 208, col: 31, offset: 6284},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 208, col: 37, offset: 6290},
								expr: &seqExpr{
									pos: position{line: 208, col: 38, offset: 6291},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 208, col: 38, offset: 6291},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 42, offset: 6295},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 208, col: 44, offset: 6297},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 208, col: 58, offset: 6311},
							expr: &ruleRefExpr{
								pos:  position{line: 208, col: 58, offset: 6311},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 219, col: 1, offset: 6523},
			expr: &actionExpr{
				pos: position{line: 219, col: 11, offset: 6535},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 219, col: 11, offset: 6535},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 219, col: 11, offset: 6535},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 219, col: 21, offset: 6545},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 219, col: 23, offset: 6547},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 219, col: 27, offset: 6551},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 219, col: 37, offset: 6561},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 219, col: 39, offset: 6563},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 219, col: 44, offset: 6568},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 219, col: 55, offset: 6579},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 226, col: 1, offset: 6688},
			expr: &actionExpr{
				pos: position{line: 226, col: 10, offset: 6699},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 226, col: 10, offset: 6699},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 226, col: 10, offset: 6699},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 226, col: 19, offset: 6708},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 226, col: 21, offset: 6710},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 226, col: 24, offset: 6713},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 227, col: 1, offset: 6753},
			expr: &actionExpr{
				pos: position{line: 227, col: 13, offset: 6767},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 227, col: 13, offset: 6767},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 227, col: 13, offset: 6767},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 227, col: 25, offset: 6779},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 227, col: 27, offset: 6781},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 227, col: 30, offset: 6784},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 228, col: 1, offset: 6835},
			expr: &actionExpr{
				pos: position{line: 228, col: 9, offset: 6845},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 228, col: 9, offset: 6845},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 9, offset: 6845},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 17, offset: 6853},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 19, offset: 6855},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 22, offset: 6858},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 229, col: 1, offset: 6905},
			expr: &actionExpr{
				pos: position{line: 229, col: 14, offset: 6920},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 229, col: 14, offset: 6920},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 229, col: 14, offset: 6920},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 19, offset: 6925},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 30, offset: 6936},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 229, col: 33, offset: 6939},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 37, offset: 6943},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 229, col: 40, offset: 6946},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 47, offset: 6953},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 229, col: 57, offset: 6963},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 61, offset: 6967},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 239, col: 1, offset: 7128},
			expr: &actionExpr{
				pos: position{line: 239, col: 13, offset: 7142},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 239, col: 13, offset: 7142},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 239, col: 20, offset: 7149},
						expr: &seqExpr{
							pos: position{line: 239, col: 21, offset: 7150},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 239, col: 21, offset: 7150},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 239, col: 27, offset: 7156},
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
			pos:  position{line: 248, col: 1, offset: 7337},
			expr: &actionExpr{
				pos: position{line: 248, col: 9, offset: 7347},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 248, col: 9, offset: 7347},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 248, col: 9, offset: 7347},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 248, col: 16, offset: 7354},
								expr: &seqExpr{
									pos: position{line: 248, col: 17, offset: 7355},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 248, col: 17, offset: 7355},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 248, col: 27, offset: 7365},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 248, col: 32, offset: 7370},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 35, offset: 7373},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 47, offset: 7385},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 248, col: 49, offset: 7387},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 53, offset: 7391},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 55, offset: 7393},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 248, col: 59, offset: 7397},
								expr: &ruleRefExpr{
									pos:  position{line: 248, col: 59, offset: 7397},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 69, offset: 7407},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 71, offset: 7409},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 75, offset: 7413},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 85, offset: 7423},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 87, offset: 7425},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 92, offset: 7430},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 103, offset: 7441},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 106, offset: 7444},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 248, col: 110, offset: 7448},
								expr: &seqExpr{
									pos: position{line: 248, col: 111, offset: 7449},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 248, col: 111, offset: 7449},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 248, col: 115, offset: 7453},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 248, col: 117, offset: 7455},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 248, col: 130, offset: 7468},
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 130, offset: 7468},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 267, col: 1, offset: 7887},
			expr: &actionExpr{
				pos: position{line: 267, col: 12, offset: 7900},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 267, col: 13, offset: 7901},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 267, col: 13, offset: 7901},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 267, col: 26, offset: 7914},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 271, col: 1, offset: 7988},
			expr: &actionExpr{
				pos: position{line: 271, col: 11, offset: 8000},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 271, col: 11, offset: 8000},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 271, col: 11, offset: 8000},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 21, offset: 8010},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 23, offset: 8012},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 271, col: 28, offset: 8017},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 39, offset: 8028},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 41, offset: 8030},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 271, col: 49, offset: 8038},
								expr: &seqExpr{
									pos: position{line: 271, col: 50, offset: 8039},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 271, col: 50, offset: 8039},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 60, offset: 8049},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 63, offset: 8052},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 74, offset: 8063},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 79, offset: 8068},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 271, col: 82, offset: 8071},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 86, offset: 8075},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 271, col: 89, offset: 8078},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 271, col: 97, offset: 8086},
								expr: &seqExpr{
									pos: position{line: 271, col: 98, offset: 8087},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 271, col: 98, offset: 8087},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 271, col: 107, offset: 8096},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 271, col: 113, offset: 8102},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 271, col: 113, offset: 8102},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 271, col: 119, offset: 8108},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 271, col: 138, offset: 8127},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 286, col: 1, offset: 8522},
			expr: &actionExpr{
				pos: position{line: 286, col: 21, offset: 8544},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 286, col: 21, offset: 8544,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 290, col: 1, offset: 8613},
			expr: &actionExpr{
				pos: position{line: 290, col: 12, offset: 8626},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 290, col: 12, offset: 8626},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 290, col: 12, offset: 8626},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 290, col: 19, offset: 8633},
								expr: &seqExpr{
									pos: position{line: 290, col: 20, offset: 8634},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 290, col: 20, offset: 8634},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 290, col: 30, offset: 8644},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 290, col: 35, offset: 8649},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 290, col: 42, offset: 8656},
								expr: &seqExpr{
									pos: position{line: 290, col: 43, offset: 8657},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 290, col: 43, offset: 8657},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 290, col: 52, offset: 8666},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 290, col: 57, offset: 8671},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 290, col: 61, offset: 8675},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 290, col: 74, offset: 8688},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 290, col: 77, offset: 8691},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 290, col: 82, offset: 8696},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 290, col: 93, offset: 8707},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 290, col: 95, offset: 8709},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 290, col: 99, offset: 8713},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 290, col: 102, offset: 8716},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 290, col: 112, offset: 8726},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 290, col: 122, offset: 8736},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 290, col: 126, offset: 8740},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 290, col: 129, offset: 8743},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 290, col: 140, offset: 8754},
								expr: &ruleRefExpr{
									pos:  position{line: 290, col: 140, offset: 8754},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 290, col: 148, offset: 8762},
							expr: &ruleRefExpr{
								pos:  position{line: 290, col: 148, offset: 8762},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 317, col: 1, offset: 9353},
			expr: &actionExpr{
				pos: position{line: 317, col: 16, offset: 9370},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 317, col: 16, offset: 9370},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 317, col: 21, offset: 9375},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 317, col: 21, offset: 9375},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 317, col: 30, offset: 9384},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 324, col: 1, offset: 9506},
			expr: &actionExpr{
				pos: position{line: 324, col: 10, offset: 9517},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 324, col: 10, offset: 9517},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 324, col: 10, offset: 9517},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 19, offset: 9526},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 324, col: 22, offset: 9529},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 26, offset: 9533},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 29, offset: 9536},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 40, offset: 9547},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 324, col: 50, offset: 9557},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 328, col: 1, offset: 9593},
			expr: &actionExpr{
				pos: position{line: 328, col: 13, offset: 9607},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 328, col: 13, offset: 9607},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 328, col: 18, offset: 9612},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 328, col: 18, offset: 9612},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 328, col: 29, offset: 9623},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 328, col: 45, offset: 9639},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "DefinitionType",
			pos:  position{line: 335, col: 1, offset: 9764},
			expr: &actionExpr{
				pos: position{line: 335, col: 18, offset: 9783},
				run: (*parser).callonDefinitionType1,
				expr: &labeledExpr{
					pos:   position{line: 335, col: 18, offset: 9783},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 335, col: 23, offset: 9788},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 335, col: 23, offset: 9788},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 335, col: 34, offset: 9799},
								name: "ContainerType",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 339, col: 1, offset: 9839},
			expr: &actionExpr{
				pos: position{line: 339, col: 12, offset: 9852},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 339, col: 13, offset: 9853},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 339, col: 13, offset: 9853},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 22, offset: 9862},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 31, offset: 9871},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 39, offset: 9879},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 47, offset: 9887},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 55, offset: 9895},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 66, offset: 9906},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 339, col: 77, offset: 9917},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 343, col: 1, offset: 9977},
			expr: &actionExpr{
				pos: position{line: 343, col: 17, offset: 9995},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 343, col: 17, offset: 9995},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 343, col: 22, offset: 10000},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 343, col: 22, offset: 10000},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 343, col: 32, offset: 10010},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 343, col: 42, offset: 10020},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 347, col: 1, offset: 10055},
			expr: &actionExpr{
				pos: position{line: 347, col: 11, offset: 10067},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 347, col: 11, offset: 10067},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 347, col: 11, offset: 10067},
							expr: &ruleRefExpr{
								pos:  position{line: 347, col: 11, offset: 10067},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 347, col: 20, offset: 10076},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 27, offset: 10083},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 347, col: 30, offset: 10086},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 347, col: 34, offset: 10090},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 44, offset: 10100},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 347, col: 47, offset: 10103},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 51, offset: 10107},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 347, col: 54, offset: 10110},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 347, col: 60, offset: 10116},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 347, col: 70, offset: 10126},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 347, col: 73, offset: 10129},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 355, col: 1, offset: 10252},
			expr: &actionExpr{
				pos: position{line: 355, col: 11, offset: 10264},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 355, col: 11, offset: 10264},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 355, col: 11, offset: 10264},
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 11, offset: 10264},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 355, col: 20, offset: 10273},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 27, offset: 10280},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 355, col: 30, offset: 10283},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 34, offset: 10287},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 44, offset: 10297},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 355, col: 47, offset: 10300},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 362, col: 1, offset: 10391},
			expr: &actionExpr{
				pos: position{line: 362, col: 12, offset: 10404},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 362, col: 12, offset: 10404},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 362, col: 12, offset: 10404},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 362, col: 20, offset: 10412},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 362, col: 23, offset: 10415},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 362, col: 27, offset: 10419},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 362, col: 37, offset: 10429},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 362, col: 40, offset: 10432},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 369, col: 1, offset: 10524},
			expr: &actionExpr{
				pos: position{line: 369, col: 11, offset: 10536},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 369, col: 11, offset: 10536},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 369, col: 11, offset: 10536},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 369, col: 22, offset: 10547},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 369, col: 30, offset: 10555},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 373, col: 1, offset: 10592},
			expr: &choiceExpr{
				pos: position{line: 373, col: 14, offset: 10607},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 373, col: 14, offset: 10607},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 24, offset: 10617},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 41, offset: 10634},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 55, offset: 10648},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 66, offset: 10659},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 373, col: 78, offset: 10671},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 375, col: 1, offset: 10683},
			expr: &actionExpr{
				pos: position{line: 375, col: 15, offset: 10699},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 375, col: 15, offset: 10699},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 375, col: 15, offset: 10699},
							expr: &charClassMatcher{
								pos:        position{line: 375, col: 15, offset: 10699},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 375, col: 21, offset: 10705},
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 21, offset: 10705},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 379, col: 1, offset: 10769},
			expr: &actionExpr{
				pos: position{line: 379, col: 18, offset: 10788},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 379, col: 18, offset: 10788},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 379, col: 18, offset: 10788},
							expr: &charClassMatcher{
								pos:        position{line: 379, col: 18, offset: 10788},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 379, col: 24, offset: 10794},
							expr: &ruleRefExpr{
								pos:  position{line: 379, col: 24, offset: 10794},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 379, col: 31, offset: 10801},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 379, col: 35, offset: 10805},
							expr: &ruleRefExpr{
								pos:  position{line: 379, col: 35, offset: 10805},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 379, col: 42, offset: 10812},
							expr: &seqExpr{
								pos: position{line: 379, col: 44, offset: 10814},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 379, col: 44, offset: 10814},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 379, col: 51, offset: 10821},
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
			pos:  position{line: 383, col: 1, offset: 10891},
			expr: &actionExpr{
				pos: position{line: 383, col: 13, offset: 10905},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 383, col: 13, offset: 10905},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 383, col: 13, offset: 10905},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 17, offset: 10909},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 383, col: 20, offset: 10912},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 383, col: 27, offset: 10919},
								expr: &seqExpr{
									pos: position{line: 383, col: 28, offset: 10920},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 383, col: 28, offset: 10920},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 383, col: 39, offset: 10931},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 383, col: 42, offset: 10934},
											expr: &ruleRefExpr{
												pos:  position{line: 383, col: 42, offset: 10934},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 383, col: 57, offset: 10949},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 62, offset: 10954},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 383, col: 65, offset: 10957},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 392, col: 1, offset: 11151},
			expr: &actionExpr{
				pos: position{line: 392, col: 12, offset: 11164},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 392, col: 12, offset: 11164},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 392, col: 12, offset: 11164},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 392, col: 16, offset: 11168},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 392, col: 19, offset: 11171},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 392, col: 26, offset: 11178},
								expr: &seqExpr{
									pos: position{line: 392, col: 27, offset: 11179},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 392, col: 27, offset: 11179},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 38, offset: 11190},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 392, col: 41, offset: 11193},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 45, offset: 11197},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 48, offset: 11200},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 59, offset: 11211},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 392, col: 63, offset: 11215},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 392, col: 63, offset: 11215},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 392, col: 69, offset: 11221},
													expr: &litMatcher{
														pos:        position{line: 392, col: 70, offset: 11222},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 392, col: 75, offset: 11227},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 392, col: 80, offset: 11232},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 412, col: 1, offset: 11782},
			expr: &ruleRefExpr{
				pos:  position{line: 412, col: 19, offset: 11802},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 414, col: 1, offset: 11809},
			expr: &actionExpr{
				pos: position{line: 414, col: 9, offset: 11819},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 414, col: 9, offset: 11819},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 414, col: 9, offset: 11819},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 414, col: 16, offset: 11826},
								expr: &seqExpr{
									pos: position{line: 414, col: 17, offset: 11827},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 414, col: 17, offset: 11827},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 414, col: 27, offset: 11837},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 414, col: 32, offset: 11842},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 40, offset: 11850},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 414, col: 43, offset: 11853},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 414, col: 48, offset: 11858},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 59, offset: 11869},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 414, col: 62, offset: 11872},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 66, offset: 11876},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 414, col: 69, offset: 11879},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 414, col: 76, offset: 11886},
								expr: &ruleRefExpr{
									pos:  position{line: 414, col: 76, offset: 11886},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 84, offset: 11894},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 414, col: 87, offset: 11897},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 414, col: 98, offset: 11908},
								expr: &seqExpr{
									pos: position{line: 414, col: 99, offset: 11909},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 414, col: 99, offset: 11909},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 414, col: 109, offset: 11919},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 414, col: 115, offset: 11925},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 414, col: 115, offset: 11925},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 414, col: 121, offset: 11931},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 138, offset: 11948},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 435, col: 1, offset: 12493},
			expr: &actionExpr{
				pos: position{line: 435, col: 19, offset: 12513},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 435, col: 19, offset: 12513,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 439, col: 1, offset: 12580},
			expr: &actionExpr{
				pos: position{line: 439, col: 10, offset: 12591},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 439, col: 10, offset: 12591},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 439, col: 10, offset: 12591},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 439, col: 19, offset: 12600},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 439, col: 21, offset: 12602},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 439, col: 26, offset: 12607},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 439, col: 34, offset: 12615},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 443, col: 1, offset: 12665},
			expr: &actionExpr{
				pos: position{line: 443, col: 13, offset: 12679},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 443, col: 13, offset: 12679},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 443, col: 13, offset: 12679},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 443, col: 20, offset: 12686},
								expr: &seqExpr{
									pos: position{line: 443, col: 21, offset: 12687},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 443, col: 21, offset: 12687},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 443, col: 31, offset: 12697},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 443, col: 36, offset: 12702},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 41, offset: 12707},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 443, col: 52, offset: 12718},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 443, col: 54, offset: 12720},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 443, col: 58, offset: 12724},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 443, col: 61, offset: 12727},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 443, col: 67, offset: 12733},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 443, col: 78, offset: 12744},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 459, col: 1, offset: 13246},
			expr: &actionExpr{
				pos: position{line: 459, col: 11, offset: 13258},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 459, col: 12, offset: 13259},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 459, col: 13, offset: 13260},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 459, col: 13, offset: 13260},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 459, col: 17, offset: 13264},
									expr: &choiceExpr{
										pos: position{line: 459, col: 18, offset: 13265},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 459, col: 18, offset: 13265},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 459, col: 25, offset: 13272},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 459, col: 32, offset: 13279},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 459, col: 40, offset: 13287},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 459, col: 40, offset: 13287},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 459, col: 45, offset: 13292},
									expr: &choiceExpr{
										pos: position{line: 459, col: 46, offset: 13293},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 459, col: 46, offset: 13293},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 459, col: 53, offset: 13300},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 459, col: 60, offset: 13307},
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
			pos:  position{line: 466, col: 1, offset: 13523},
			expr: &actionExpr{
				pos: position{line: 466, col: 14, offset: 13538},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 466, col: 14, offset: 13538},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 466, col: 14, offset: 13538},
							expr: &choiceExpr{
								pos: position{line: 466, col: 15, offset: 13539},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 466, col: 15, offset: 13539},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 466, col: 24, offset: 13548},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 466, col: 30, offset: 13554},
							expr: &choiceExpr{
								pos: position{line: 466, col: 31, offset: 13555},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 466, col: 31, offset: 13555},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 466, col: 40, offset: 13564},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 466, col: 48, offset: 13572},
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
			pos:  position{line: 470, col: 1, offset: 13627},
			expr: &charClassMatcher{
				pos:        position{line: 470, col: 17, offset: 13645},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 471, col: 1, offset: 13650},
			expr: &charClassMatcher{
				pos:        position{line: 471, col: 10, offset: 13661},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 472, col: 1, offset: 13670},
			expr: &charClassMatcher{
				pos:        position{line: 472, col: 9, offset: 13680},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 474, col: 1, offset: 13687},
			expr: &anyMatcher{
				line: 474, col: 14, offset: 13702,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 475, col: 1, offset: 13704},
			expr: &actionExpr{
				pos: position{line: 475, col: 13, offset: 13718},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 475, col: 13, offset: 13718},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 475, col: 13, offset: 13718},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 475, col: 20, offset: 13725},
							expr: &seqExpr{
								pos: position{line: 475, col: 22, offset: 13727},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 475, col: 22, offset: 13727},
										expr: &litMatcher{
											pos:        position{line: 475, col: 23, offset: 13728},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 475, col: 28, offset: 13733},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 475, col: 42, offset: 13747},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 481, col: 1, offset: 13927},
			expr: &choiceExpr{
				pos: position{line: 481, col: 11, offset: 13939},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 481, col: 11, offset: 13939},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 481, col: 30, offset: 13958},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 482, col: 1, offset: 13976},
			expr: &seqExpr{
				pos: position{line: 482, col: 20, offset: 13997},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 482, col: 20, offset: 13997},
						expr: &ruleRefExpr{
							pos:  position{line: 482, col: 21, offset: 13998},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 482, col: 31, offset: 14008},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 482, col: 36, offset: 14013},
						expr: &seqExpr{
							pos: position{line: 482, col: 38, offset: 14015},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 482, col: 38, offset: 14015},
									expr: &litMatcher{
										pos:        position{line: 482, col: 39, offset: 14016},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 482, col: 44, offset: 14021},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 482, col: 58, offset: 14035},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 483, col: 1, offset: 14040},
			expr: &seqExpr{
				pos: position{line: 483, col: 36, offset: 14077},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 483, col: 36, offset: 14077},
						expr: &ruleRefExpr{
							pos:  position{line: 483, col: 37, offset: 14078},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 483, col: 47, offset: 14088},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 483, col: 52, offset: 14093},
						expr: &seqExpr{
							pos: position{line: 483, col: 54, offset: 14095},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 483, col: 54, offset: 14095},
									expr: &choiceExpr{
										pos: position{line: 483, col: 57, offset: 14098},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 483, col: 57, offset: 14098},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 483, col: 64, offset: 14105},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 483, col: 70, offset: 14111},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 483, col: 84, offset: 14125},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 484, col: 1, offset: 14130},
			expr: &choiceExpr{
				pos: position{line: 484, col: 21, offset: 14152},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 484, col: 22, offset: 14153},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 484, col: 22, offset: 14153},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 484, col: 27, offset: 14158},
								expr: &seqExpr{
									pos: position{line: 484, col: 29, offset: 14160},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 484, col: 29, offset: 14160},
											expr: &ruleRefExpr{
												pos:  position{line: 484, col: 30, offset: 14161},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 484, col: 34, offset: 14165},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 484, col: 52, offset: 14183},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 484, col: 52, offset: 14183},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 484, col: 56, offset: 14187},
								expr: &seqExpr{
									pos: position{line: 484, col: 58, offset: 14189},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 484, col: 58, offset: 14189},
											expr: &ruleRefExpr{
												pos:  position{line: 484, col: 59, offset: 14190},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 484, col: 63, offset: 14194},
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
			pos:  position{line: 486, col: 1, offset: 14210},
			expr: &zeroOrMoreExpr{
				pos: position{line: 486, col: 6, offset: 14217},
				expr: &choiceExpr{
					pos: position{line: 486, col: 8, offset: 14219},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 486, col: 8, offset: 14219},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 486, col: 21, offset: 14232},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 486, col: 27, offset: 14238},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 487, col: 1, offset: 14249},
			expr: &zeroOrMoreExpr{
				pos: position{line: 487, col: 5, offset: 14255},
				expr: &choiceExpr{
					pos: position{line: 487, col: 7, offset: 14257},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 487, col: 7, offset: 14257},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 487, col: 20, offset: 14270},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 488, col: 1, offset: 14306},
			expr: &zeroOrMoreExpr{
				pos: position{line: 488, col: 6, offset: 14313},
				expr: &ruleRefExpr{
					pos:  position{line: 488, col: 6, offset: 14313},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 490, col: 1, offset: 14326},
			expr: &charClassMatcher{
				pos:        position{line: 490, col: 14, offset: 14341},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 491, col: 1, offset: 14349},
			expr: &litMatcher{
				pos:        position{line: 491, col: 7, offset: 14357},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 492, col: 1, offset: 14362},
			expr: &choiceExpr{
				pos: position{line: 492, col: 7, offset: 14370},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 492, col: 7, offset: 14370},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 492, col: 7, offset: 14370},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 492, col: 10, offset: 14373},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 492, col: 16, offset: 14379},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 492, col: 16, offset: 14379},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 492, col: 18, offset: 14381},
								expr: &ruleRefExpr{
									pos:  position{line: 492, col: 18, offset: 14381},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 492, col: 37, offset: 14400},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 492, col: 43, offset: 14406},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 492, col: 43, offset: 14406},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 492, col: 46, offset: 14409},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 494, col: 1, offset: 14414},
			expr: &notExpr{
				pos: position{line: 494, col: 7, offset: 14422},
				expr: &anyMatcher{
					line: 494, col: 8, offset: 14423,
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
