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

type namespace struct {
	scope     string
	namespace string
}

type typeDef struct {
	name string
	typ  *Type
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
			pos:  position{line: 75, col: 1, offset: 2362},
			expr: &actionExpr{
				pos: position{line: 75, col: 11, offset: 2374},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 75, col: 11, offset: 2374},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 75, col: 11, offset: 2374},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 75, col: 14, offset: 2377},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 75, col: 25, offset: 2388},
								expr: &seqExpr{
									pos: position{line: 75, col: 27, offset: 2390},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 75, col: 27, offset: 2390},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 75, col: 37, offset: 2400},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 75, col: 44, offset: 2407},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 75, col: 44, offset: 2407},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 75, col: 50, offset: 2413},
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
			pos:  position{line: 126, col: 1, offset: 4420},
			expr: &actionExpr{
				pos: position{line: 126, col: 15, offset: 4436},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 126, col: 15, offset: 4436,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 134, col: 1, offset: 4739},
			expr: &choiceExpr{
				pos: position{line: 134, col: 13, offset: 4753},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 134, col: 13, offset: 4753},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 134, col: 25, offset: 4765},
						name: "Scope",
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 136, col: 1, offset: 4772},
			expr: &actionExpr{
				pos: position{line: 136, col: 13, offset: 4786},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 136, col: 13, offset: 4786},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 136, col: 13, offset: 4786},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 136, col: 25, offset: 4798},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 136, col: 27, offset: 4800},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 136, col: 33, offset: 4806},
								expr: &charClassMatcher{
									pos:        position{line: 136, col: 33, offset: 4806},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 136, col: 42, offset: 4815},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 136, col: 44, offset: 4817},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 136, col: 47, offset: 4820},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 136, col: 58, offset: 4831},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 147, col: 1, offset: 5229},
			expr: &actionExpr{
				pos: position{line: 147, col: 9, offset: 5239},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 147, col: 9, offset: 5239},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 147, col: 9, offset: 5239},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 147, col: 17, offset: 5247},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 147, col: 19, offset: 5249},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 147, col: 23, offset: 5253},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 147, col: 33, offset: 5263},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 147, col: 35, offset: 5265},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 147, col: 40, offset: 5270},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 147, col: 51, offset: 5281},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 147, col: 53, offset: 5283},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 147, col: 57, offset: 5287},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 147, col: 59, offset: 5289},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 147, col: 65, offset: 5295},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 147, col: 76, offset: 5306},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 155, col: 1, offset: 5438},
			expr: &actionExpr{
				pos: position{line: 155, col: 8, offset: 5447},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 155, col: 8, offset: 5447},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 155, col: 8, offset: 5447},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 155, col: 15, offset: 5454},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 155, col: 17, offset: 5456},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 155, col: 22, offset: 5461},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 155, col: 33, offset: 5472},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 155, col: 36, offset: 5475},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 155, col: 40, offset: 5479},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 155, col: 43, offset: 5482},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 155, col: 50, offset: 5489},
								expr: &seqExpr{
									pos: position{line: 155, col: 51, offset: 5490},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 155, col: 51, offset: 5490},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 155, col: 61, offset: 5500},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 155, col: 66, offset: 5505},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 155, col: 70, offset: 5509},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 178, col: 1, offset: 6121},
			expr: &actionExpr{
				pos: position{line: 178, col: 13, offset: 6135},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 178, col: 13, offset: 6135},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 178, col: 13, offset: 6135},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 178, col: 18, offset: 6140},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 178, col: 29, offset: 6151},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 178, col: 31, offset: 6153},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 178, col: 37, offset: 6159},
								expr: &seqExpr{
									pos: position{line: 178, col: 38, offset: 6160},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 178, col: 38, offset: 6160},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 178, col: 42, offset: 6164},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 178, col: 44, offset: 6166},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 178, col: 58, offset: 6180},
							expr: &ruleRefExpr{
								pos:  position{line: 178, col: 58, offset: 6180},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 189, col: 1, offset: 6392},
			expr: &actionExpr{
				pos: position{line: 189, col: 11, offset: 6404},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 189, col: 11, offset: 6404},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 189, col: 11, offset: 6404},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 21, offset: 6414},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 23, offset: 6416},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 27, offset: 6420},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 37, offset: 6430},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 39, offset: 6432},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 44, offset: 6437},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 55, offset: 6448},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 196, col: 1, offset: 6556},
			expr: &actionExpr{
				pos: position{line: 196, col: 10, offset: 6567},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 196, col: 10, offset: 6567},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 196, col: 10, offset: 6567},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 196, col: 19, offset: 6576},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 196, col: 21, offset: 6578},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 24, offset: 6581},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 197, col: 1, offset: 6621},
			expr: &actionExpr{
				pos: position{line: 197, col: 13, offset: 6635},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 197, col: 13, offset: 6635},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 197, col: 13, offset: 6635},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 197, col: 25, offset: 6647},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 197, col: 27, offset: 6649},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 197, col: 30, offset: 6652},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 198, col: 1, offset: 6703},
			expr: &actionExpr{
				pos: position{line: 198, col: 9, offset: 6713},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 198, col: 9, offset: 6713},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 198, col: 9, offset: 6713},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 17, offset: 6721},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 19, offset: 6723},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 198, col: 22, offset: 6726},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 199, col: 1, offset: 6773},
			expr: &actionExpr{
				pos: position{line: 199, col: 14, offset: 6788},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 199, col: 14, offset: 6788},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 199, col: 14, offset: 6788},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 199, col: 19, offset: 6793},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 30, offset: 6804},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 199, col: 33, offset: 6807},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 37, offset: 6811},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 40, offset: 6814},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 199, col: 47, offset: 6821},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 199, col: 57, offset: 6831},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 61, offset: 6835},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 209, col: 1, offset: 6996},
			expr: &actionExpr{
				pos: position{line: 209, col: 13, offset: 7010},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 209, col: 13, offset: 7010},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 209, col: 20, offset: 7017},
						expr: &seqExpr{
							pos: position{line: 209, col: 21, offset: 7018},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 209, col: 21, offset: 7018},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 209, col: 27, offset: 7024},
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
			pos:  position{line: 218, col: 1, offset: 7205},
			expr: &actionExpr{
				pos: position{line: 218, col: 9, offset: 7215},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 218, col: 9, offset: 7215},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 218, col: 9, offset: 7215},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 218, col: 12, offset: 7218},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 218, col: 24, offset: 7230},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 218, col: 26, offset: 7232},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 218, col: 30, offset: 7236},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 218, col: 32, offset: 7238},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 218, col: 36, offset: 7242},
								expr: &ruleRefExpr{
									pos:  position{line: 218, col: 36, offset: 7242},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 218, col: 46, offset: 7252},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 218, col: 48, offset: 7254},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 218, col: 52, offset: 7258},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 218, col: 62, offset: 7268},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 218, col: 64, offset: 7270},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 218, col: 69, offset: 7275},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 218, col: 80, offset: 7286},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 218, col: 83, offset: 7289},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 218, col: 87, offset: 7293},
								expr: &seqExpr{
									pos: position{line: 218, col: 88, offset: 7294},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 218, col: 88, offset: 7294},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 218, col: 92, offset: 7298},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 218, col: 94, offset: 7300},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 218, col: 107, offset: 7313},
							expr: &ruleRefExpr{
								pos:  position{line: 218, col: 107, offset: 7313},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 233, col: 1, offset: 7624},
			expr: &actionExpr{
				pos: position{line: 233, col: 12, offset: 7637},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 233, col: 13, offset: 7638},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 233, col: 13, offset: 7638},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 233, col: 26, offset: 7651},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 237, col: 1, offset: 7725},
			expr: &actionExpr{
				pos: position{line: 237, col: 11, offset: 7737},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 237, col: 11, offset: 7737},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 237, col: 11, offset: 7737},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 21, offset: 7747},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 23, offset: 7749},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 28, offset: 7754},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 39, offset: 7765},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 41, offset: 7767},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 237, col: 49, offset: 7775},
								expr: &seqExpr{
									pos: position{line: 237, col: 50, offset: 7776},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 237, col: 50, offset: 7776},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 237, col: 60, offset: 7786},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 237, col: 63, offset: 7789},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 237, col: 74, offset: 7800},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 79, offset: 7805},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 237, col: 82, offset: 7808},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 86, offset: 7812},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 89, offset: 7815},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 237, col: 97, offset: 7823},
								expr: &seqExpr{
									pos: position{line: 237, col: 98, offset: 7824},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 237, col: 98, offset: 7824},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 237, col: 107, offset: 7833},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 237, col: 113, offset: 7839},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 237, col: 113, offset: 7839},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 237, col: 119, offset: 7845},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 138, offset: 7864},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 252, col: 1, offset: 8259},
			expr: &actionExpr{
				pos: position{line: 252, col: 21, offset: 8281},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 252, col: 21, offset: 8281,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 256, col: 1, offset: 8350},
			expr: &actionExpr{
				pos: position{line: 256, col: 12, offset: 8363},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 256, col: 12, offset: 8363},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 256, col: 12, offset: 8363},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 256, col: 19, offset: 8370},
								expr: &seqExpr{
									pos: position{line: 256, col: 20, offset: 8371},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 256, col: 20, offset: 8371},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 256, col: 29, offset: 8380},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 256, col: 34, offset: 8385},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 38, offset: 8389},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 51, offset: 8402},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 54, offset: 8405},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 59, offset: 8410},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 70, offset: 8421},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 256, col: 72, offset: 8423},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 76, offset: 8427},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 79, offset: 8430},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 89, offset: 8440},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 256, col: 99, offset: 8450},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 103, offset: 8454},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 106, offset: 8457},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 256, col: 117, offset: 8468},
								expr: &ruleRefExpr{
									pos:  position{line: 256, col: 117, offset: 8468},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 256, col: 125, offset: 8476},
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 125, offset: 8476},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 279, col: 1, offset: 8944},
			expr: &actionExpr{
				pos: position{line: 279, col: 16, offset: 8961},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 279, col: 16, offset: 8961},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 279, col: 21, offset: 8966},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 279, col: 21, offset: 8966},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 279, col: 30, offset: 8975},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 286, col: 1, offset: 9097},
			expr: &actionExpr{
				pos: position{line: 286, col: 10, offset: 9108},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 286, col: 10, offset: 9108},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 286, col: 10, offset: 9108},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 286, col: 19, offset: 9117},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 286, col: 22, offset: 9120},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 286, col: 26, offset: 9124},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 286, col: 29, offset: 9127},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 286, col: 40, offset: 9138},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 286, col: 50, offset: 9148},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 290, col: 1, offset: 9184},
			expr: &actionExpr{
				pos: position{line: 290, col: 13, offset: 9198},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 290, col: 13, offset: 9198},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 290, col: 18, offset: 9203},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 290, col: 18, offset: 9203},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 290, col: 29, offset: 9214},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 290, col: 45, offset: 9230},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "DefinitionType",
			pos:  position{line: 297, col: 1, offset: 9355},
			expr: &actionExpr{
				pos: position{line: 297, col: 18, offset: 9374},
				run: (*parser).callonDefinitionType1,
				expr: &labeledExpr{
					pos:   position{line: 297, col: 18, offset: 9374},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 297, col: 23, offset: 9379},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 297, col: 23, offset: 9379},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 297, col: 34, offset: 9390},
								name: "ContainerType",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 301, col: 1, offset: 9430},
			expr: &actionExpr{
				pos: position{line: 301, col: 12, offset: 9443},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 301, col: 13, offset: 9444},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 301, col: 13, offset: 9444},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 301, col: 22, offset: 9453},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 301, col: 31, offset: 9462},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 301, col: 39, offset: 9470},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 301, col: 47, offset: 9478},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 301, col: 55, offset: 9486},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 301, col: 66, offset: 9497},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 301, col: 77, offset: 9508},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 305, col: 1, offset: 9568},
			expr: &actionExpr{
				pos: position{line: 305, col: 17, offset: 9586},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 305, col: 17, offset: 9586},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 305, col: 22, offset: 9591},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 305, col: 22, offset: 9591},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 305, col: 32, offset: 9601},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 305, col: 42, offset: 9611},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 309, col: 1, offset: 9646},
			expr: &actionExpr{
				pos: position{line: 309, col: 11, offset: 9658},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 309, col: 11, offset: 9658},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 309, col: 11, offset: 9658},
							expr: &ruleRefExpr{
								pos:  position{line: 309, col: 11, offset: 9658},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 309, col: 20, offset: 9667},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 309, col: 27, offset: 9674},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 309, col: 30, offset: 9677},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 309, col: 34, offset: 9681},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 309, col: 44, offset: 9691},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 309, col: 47, offset: 9694},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 309, col: 51, offset: 9698},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 309, col: 54, offset: 9701},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 309, col: 60, offset: 9707},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 309, col: 70, offset: 9717},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 309, col: 73, offset: 9720},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 317, col: 1, offset: 9843},
			expr: &actionExpr{
				pos: position{line: 317, col: 11, offset: 9855},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 317, col: 11, offset: 9855},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 317, col: 11, offset: 9855},
							expr: &ruleRefExpr{
								pos:  position{line: 317, col: 11, offset: 9855},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 317, col: 20, offset: 9864},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 317, col: 27, offset: 9871},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 317, col: 30, offset: 9874},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 317, col: 34, offset: 9878},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 317, col: 44, offset: 9888},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 317, col: 47, offset: 9891},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 324, col: 1, offset: 9982},
			expr: &actionExpr{
				pos: position{line: 324, col: 12, offset: 9995},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 324, col: 12, offset: 9995},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 324, col: 12, offset: 9995},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 20, offset: 10003},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 324, col: 23, offset: 10006},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 324, col: 27, offset: 10010},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 324, col: 37, offset: 10020},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 324, col: 40, offset: 10023},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 331, col: 1, offset: 10115},
			expr: &actionExpr{
				pos: position{line: 331, col: 11, offset: 10127},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 331, col: 11, offset: 10127},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 331, col: 11, offset: 10127},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 331, col: 22, offset: 10138},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 331, col: 30, offset: 10146},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 335, col: 1, offset: 10183},
			expr: &choiceExpr{
				pos: position{line: 335, col: 14, offset: 10198},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 335, col: 14, offset: 10198},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 335, col: 24, offset: 10208},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 335, col: 41, offset: 10225},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 335, col: 55, offset: 10239},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 335, col: 66, offset: 10250},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 335, col: 78, offset: 10262},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 337, col: 1, offset: 10274},
			expr: &actionExpr{
				pos: position{line: 337, col: 15, offset: 10290},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 337, col: 15, offset: 10290},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 337, col: 15, offset: 10290},
							expr: &charClassMatcher{
								pos:        position{line: 337, col: 15, offset: 10290},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 337, col: 21, offset: 10296},
							expr: &ruleRefExpr{
								pos:  position{line: 337, col: 21, offset: 10296},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 341, col: 1, offset: 10360},
			expr: &actionExpr{
				pos: position{line: 341, col: 18, offset: 10379},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 341, col: 18, offset: 10379},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 341, col: 18, offset: 10379},
							expr: &charClassMatcher{
								pos:        position{line: 341, col: 18, offset: 10379},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 341, col: 24, offset: 10385},
							expr: &ruleRefExpr{
								pos:  position{line: 341, col: 24, offset: 10385},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 341, col: 31, offset: 10392},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 341, col: 35, offset: 10396},
							expr: &ruleRefExpr{
								pos:  position{line: 341, col: 35, offset: 10396},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 341, col: 42, offset: 10403},
							expr: &seqExpr{
								pos: position{line: 341, col: 44, offset: 10405},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 341, col: 44, offset: 10405},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 341, col: 51, offset: 10412},
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
			pos:  position{line: 345, col: 1, offset: 10482},
			expr: &actionExpr{
				pos: position{line: 345, col: 13, offset: 10496},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 345, col: 13, offset: 10496},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 345, col: 13, offset: 10496},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 345, col: 17, offset: 10500},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 345, col: 20, offset: 10503},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 345, col: 27, offset: 10510},
								expr: &seqExpr{
									pos: position{line: 345, col: 28, offset: 10511},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 345, col: 28, offset: 10511},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 345, col: 39, offset: 10522},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 345, col: 42, offset: 10525},
											expr: &ruleRefExpr{
												pos:  position{line: 345, col: 42, offset: 10525},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 345, col: 57, offset: 10540},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 345, col: 62, offset: 10545},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 345, col: 65, offset: 10548},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 354, col: 1, offset: 10742},
			expr: &actionExpr{
				pos: position{line: 354, col: 12, offset: 10755},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 354, col: 12, offset: 10755},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 354, col: 12, offset: 10755},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 354, col: 16, offset: 10759},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 354, col: 19, offset: 10762},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 354, col: 26, offset: 10769},
								expr: &seqExpr{
									pos: position{line: 354, col: 27, offset: 10770},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 354, col: 27, offset: 10770},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 354, col: 38, offset: 10781},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 354, col: 41, offset: 10784},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 354, col: 45, offset: 10788},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 354, col: 48, offset: 10791},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 354, col: 59, offset: 10802},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 354, col: 63, offset: 10806},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 354, col: 63, offset: 10806},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 354, col: 69, offset: 10812},
													expr: &litMatcher{
														pos:        position{line: 354, col: 70, offset: 10813},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 354, col: 75, offset: 10818},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 354, col: 80, offset: 10823},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Scope",
			pos:  position{line: 374, col: 1, offset: 11373},
			expr: &actionExpr{
				pos: position{line: 374, col: 9, offset: 11383},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 374, col: 9, offset: 11383},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 374, col: 9, offset: 11383},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 374, col: 17, offset: 11391},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 374, col: 20, offset: 11394},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 374, col: 25, offset: 11399},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 374, col: 36, offset: 11410},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 374, col: 39, offset: 11413},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 374, col: 43, offset: 11417},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 374, col: 46, offset: 11420},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 374, col: 53, offset: 11427},
								expr: &ruleRefExpr{
									pos:  position{line: 374, col: 53, offset: 11427},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 374, col: 61, offset: 11435},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 374, col: 64, offset: 11438},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 374, col: 75, offset: 11449},
								expr: &seqExpr{
									pos: position{line: 374, col: 76, offset: 11450},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 374, col: 76, offset: 11450},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 374, col: 86, offset: 11460},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 374, col: 92, offset: 11466},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 374, col: 92, offset: 11466},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 374, col: 98, offset: 11472},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 374, col: 115, offset: 11489},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 391, col: 1, offset: 11987},
			expr: &actionExpr{
				pos: position{line: 391, col: 19, offset: 12007},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 391, col: 19, offset: 12007,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 395, col: 1, offset: 12078},
			expr: &actionExpr{
				pos: position{line: 395, col: 10, offset: 12089},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 395, col: 10, offset: 12089},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 395, col: 10, offset: 12089},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 395, col: 19, offset: 12098},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 395, col: 21, offset: 12100},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 395, col: 26, offset: 12105},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 395, col: 34, offset: 12113},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 399, col: 1, offset: 12167},
			expr: &actionExpr{
				pos: position{line: 399, col: 13, offset: 12181},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 399, col: 13, offset: 12181},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 399, col: 13, offset: 12181},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 399, col: 18, offset: 12186},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 29, offset: 12197},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 399, col: 31, offset: 12199},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 35, offset: 12203},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 38, offset: 12206},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 399, col: 44, offset: 12212},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 55, offset: 12223},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 411, col: 1, offset: 12630},
			expr: &actionExpr{
				pos: position{line: 411, col: 11, offset: 12642},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 411, col: 12, offset: 12643},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 411, col: 13, offset: 12644},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 411, col: 13, offset: 12644},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 411, col: 17, offset: 12648},
									expr: &choiceExpr{
										pos: position{line: 411, col: 18, offset: 12649},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 411, col: 18, offset: 12649},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 411, col: 25, offset: 12656},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 411, col: 32, offset: 12663},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 411, col: 40, offset: 12671},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 411, col: 40, offset: 12671},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 411, col: 45, offset: 12676},
									expr: &choiceExpr{
										pos: position{line: 411, col: 46, offset: 12677},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 411, col: 46, offset: 12677},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 411, col: 53, offset: 12684},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 411, col: 60, offset: 12691},
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
			pos:  position{line: 418, col: 1, offset: 12927},
			expr: &actionExpr{
				pos: position{line: 418, col: 14, offset: 12942},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 418, col: 14, offset: 12942},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 418, col: 14, offset: 12942},
							expr: &choiceExpr{
								pos: position{line: 418, col: 15, offset: 12943},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 418, col: 15, offset: 12943},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 418, col: 24, offset: 12952},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 418, col: 30, offset: 12958},
							expr: &choiceExpr{
								pos: position{line: 418, col: 31, offset: 12959},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 418, col: 31, offset: 12959},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 418, col: 40, offset: 12968},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 418, col: 48, offset: 12976},
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
			pos:  position{line: 422, col: 1, offset: 13035},
			expr: &charClassMatcher{
				pos:        position{line: 422, col: 17, offset: 13053},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 423, col: 1, offset: 13058},
			expr: &charClassMatcher{
				pos:        position{line: 423, col: 10, offset: 13069},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 424, col: 1, offset: 13078},
			expr: &charClassMatcher{
				pos:        position{line: 424, col: 9, offset: 13088},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 426, col: 1, offset: 13095},
			expr: &anyMatcher{
				line: 426, col: 14, offset: 13110,
			},
		},
		{
			name: "Comment",
			pos:  position{line: 427, col: 1, offset: 13112},
			expr: &choiceExpr{
				pos: position{line: 427, col: 11, offset: 13124},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 427, col: 11, offset: 13124},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 427, col: 30, offset: 13143},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 428, col: 1, offset: 13161},
			expr: &seqExpr{
				pos: position{line: 428, col: 20, offset: 13182},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 428, col: 20, offset: 13182},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 428, col: 25, offset: 13187},
						expr: &seqExpr{
							pos: position{line: 428, col: 27, offset: 13189},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 428, col: 27, offset: 13189},
									expr: &litMatcher{
										pos:        position{line: 428, col: 28, offset: 13190},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 428, col: 33, offset: 13195},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 428, col: 47, offset: 13209},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 429, col: 1, offset: 13214},
			expr: &seqExpr{
				pos: position{line: 429, col: 36, offset: 13251},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 429, col: 36, offset: 13251},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 429, col: 41, offset: 13256},
						expr: &seqExpr{
							pos: position{line: 429, col: 43, offset: 13258},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 429, col: 43, offset: 13258},
									expr: &choiceExpr{
										pos: position{line: 429, col: 46, offset: 13261},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 429, col: 46, offset: 13261},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 429, col: 53, offset: 13268},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 429, col: 59, offset: 13274},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 429, col: 73, offset: 13288},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 430, col: 1, offset: 13293},
			expr: &choiceExpr{
				pos: position{line: 430, col: 21, offset: 13315},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 430, col: 22, offset: 13316},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 430, col: 22, offset: 13316},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 430, col: 27, offset: 13321},
								expr: &seqExpr{
									pos: position{line: 430, col: 29, offset: 13323},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 430, col: 29, offset: 13323},
											expr: &ruleRefExpr{
												pos:  position{line: 430, col: 30, offset: 13324},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 430, col: 34, offset: 13328},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 430, col: 52, offset: 13346},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 430, col: 52, offset: 13346},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 430, col: 56, offset: 13350},
								expr: &seqExpr{
									pos: position{line: 430, col: 58, offset: 13352},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 430, col: 58, offset: 13352},
											expr: &ruleRefExpr{
												pos:  position{line: 430, col: 59, offset: 13353},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 430, col: 63, offset: 13357},
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
			pos:  position{line: 432, col: 1, offset: 13373},
			expr: &zeroOrMoreExpr{
				pos: position{line: 432, col: 6, offset: 13380},
				expr: &choiceExpr{
					pos: position{line: 432, col: 8, offset: 13382},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 432, col: 8, offset: 13382},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 432, col: 21, offset: 13395},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 432, col: 27, offset: 13401},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 433, col: 1, offset: 13412},
			expr: &zeroOrMoreExpr{
				pos: position{line: 433, col: 5, offset: 13418},
				expr: &choiceExpr{
					pos: position{line: 433, col: 7, offset: 13420},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 433, col: 7, offset: 13420},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 433, col: 20, offset: 13433},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 434, col: 1, offset: 13469},
			expr: &zeroOrMoreExpr{
				pos: position{line: 434, col: 6, offset: 13476},
				expr: &ruleRefExpr{
					pos:  position{line: 434, col: 6, offset: 13476},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 436, col: 1, offset: 13489},
			expr: &charClassMatcher{
				pos:        position{line: 436, col: 14, offset: 13504},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 437, col: 1, offset: 13512},
			expr: &litMatcher{
				pos:        position{line: 437, col: 7, offset: 13520},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 438, col: 1, offset: 13525},
			expr: &choiceExpr{
				pos: position{line: 438, col: 7, offset: 13533},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 438, col: 7, offset: 13533},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 438, col: 7, offset: 13533},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 438, col: 10, offset: 13536},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 438, col: 16, offset: 13542},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 438, col: 16, offset: 13542},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 438, col: 18, offset: 13544},
								expr: &ruleRefExpr{
									pos:  position{line: 438, col: 18, offset: 13544},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 438, col: 37, offset: 13563},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 438, col: 43, offset: 13569},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 438, col: 43, offset: 13569},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 438, col: 46, offset: 13572},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 440, col: 1, offset: 13577},
			expr: &notExpr{
				pos: position{line: 440, col: 7, offset: 13585},
				expr: &anyMatcher{
					line: 440, col: 8, offset: 13586,
				},
			},
		},
	},
}

func (c *current) onGrammar1(statements interface{}) (interface{}, error) {
	thrift := &Thrift{
		Includes:   make(map[string]string),
		Namespaces: make(map[string]string),
		Typedefs:   make(map[string]*Type),
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
		switch v := st.([]interface{})[0].(type) {
		case *namespace:
			thrift.Namespaces[v.scope] = v.namespace
		case *Constant:
			thrift.Constants[v.Name] = v
		case *Enum:
			thrift.Enums[v.Name] = v
		case *typeDef:
			thrift.Typedefs[v.name] = v.typ
		case *Struct:
			thrift.Structs[v.Name] = v
		case exception:
			thrift.Exceptions[v.Name] = (*Struct)(v)
		case union:
			thrift.Unions[v.Name] = unionToStruct(v)
		case *Service:
			thrift.Services[v.Name] = v
		case include:
			name := string(v)
			if ix := strings.LastIndex(name, "."); ix > 0 {
				name = name[:ix]
			}
			thrift.Includes[name] = string(v)
		case *Scope:
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
	return &typeDef{
		name: string(name.(Identifier)),
		typ:  typ.(*Type),
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

func (c *current) onField1(id, req, typ, name, def interface{}) (interface{}, error) {
	f := &Field{
		ID:   int(id.(int64)),
		Name: string(name.(Identifier)),
		Type: typ.(*Type),
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
	return p.cur.onField1(stack["id"], stack["req"], stack["typ"], stack["name"], stack["def"])
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

func (c *current) onFunction1(oneway, typ, name, arguments, exceptions interface{}) (interface{}, error) {
	m := &Method{
		Name: string(name.(Identifier)),
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
	return p.cur.onFunction1(stack["oneway"], stack["typ"], stack["name"], stack["arguments"], stack["exceptions"])
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

func (c *current) onScope1(name, prefix, operations interface{}) (interface{}, error) {
	ops := operations.([]interface{})
	scope := &Scope{
		Name:       string(name.(Identifier)),
		Operations: make([]*Operation, len(ops)),
		Prefix:     defaultPrefix,
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
	return p.cur.onScope1(stack["name"], stack["prefix"], stack["operations"])
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

func (c *current) onOperation1(name, param interface{}) (interface{}, error) {
	o := &Operation{
		Name:  string(name.(Identifier)),
		Param: string(param.(Identifier)),
	}
	return o, nil
}

func (p *parser) callonOperation1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onOperation1(stack["name"], stack["param"])
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
