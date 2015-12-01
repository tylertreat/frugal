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
			pos:  position{line: 151, col: 1, offset: 4615},
			expr: &actionExpr{
				pos: position{line: 151, col: 15, offset: 4631},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 151, col: 15, offset: 4631,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 155, col: 1, offset: 4689},
			expr: &actionExpr{
				pos: position{line: 155, col: 13, offset: 4703},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 155, col: 13, offset: 4703},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 155, col: 13, offset: 4703},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 155, col: 20, offset: 4710},
								expr: &seqExpr{
									pos: position{line: 155, col: 21, offset: 4711},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 155, col: 21, offset: 4711},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 155, col: 31, offset: 4721},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 155, col: 36, offset: 4726},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 155, col: 47, offset: 4737},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 155, col: 47, offset: 4737},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 155, col: 65, offset: 4755},
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
			pos:  position{line: 168, col: 1, offset: 5226},
			expr: &choiceExpr{
				pos: position{line: 168, col: 19, offset: 5246},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 168, col: 19, offset: 5246},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 29, offset: 5256},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 41, offset: 5268},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 49, offset: 5276},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 56, offset: 5283},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 66, offset: 5293},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 75, offset: 5302},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 87, offset: 5314},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 168, col: 95, offset: 5322},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 170, col: 1, offset: 5331},
			expr: &actionExpr{
				pos: position{line: 170, col: 11, offset: 5343},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 170, col: 11, offset: 5343},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 170, col: 11, offset: 5343},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 21, offset: 5353},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 170, col: 23, offset: 5355},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 170, col: 28, offset: 5360},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 170, col: 36, offset: 5368},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 174, col: 1, offset: 5416},
			expr: &actionExpr{
				pos: position{line: 174, col: 13, offset: 5430},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 174, col: 13, offset: 5430},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 174, col: 13, offset: 5430},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 174, col: 25, offset: 5442},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 174, col: 27, offset: 5444},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 174, col: 33, offset: 5450},
								expr: &charClassMatcher{
									pos:        position{line: 174, col: 33, offset: 5450},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 174, col: 42, offset: 5459},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 174, col: 44, offset: 5461},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 174, col: 47, offset: 5464},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 174, col: 58, offset: 5475},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 181, col: 1, offset: 5608},
			expr: &actionExpr{
				pos: position{line: 181, col: 9, offset: 5618},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 181, col: 9, offset: 5618},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 181, col: 9, offset: 5618},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 17, offset: 5626},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 19, offset: 5628},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 181, col: 23, offset: 5632},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 33, offset: 5642},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 35, offset: 5644},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 181, col: 40, offset: 5649},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 51, offset: 5660},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 181, col: 53, offset: 5662},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 57, offset: 5666},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 181, col: 59, offset: 5668},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 181, col: 65, offset: 5674},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 181, col: 76, offset: 5685},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 189, col: 1, offset: 5817},
			expr: &actionExpr{
				pos: position{line: 189, col: 8, offset: 5826},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 189, col: 8, offset: 5826},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 189, col: 8, offset: 5826},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 15, offset: 5833},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 17, offset: 5835},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 189, col: 22, offset: 5840},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 33, offset: 5851},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 189, col: 36, offset: 5854},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 40, offset: 5858},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 189, col: 43, offset: 5861},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 189, col: 50, offset: 5868},
								expr: &seqExpr{
									pos: position{line: 189, col: 51, offset: 5869},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 189, col: 51, offset: 5869},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 189, col: 61, offset: 5879},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 189, col: 66, offset: 5884},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 189, col: 70, offset: 5888},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 212, col: 1, offset: 6500},
			expr: &actionExpr{
				pos: position{line: 212, col: 13, offset: 6514},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 212, col: 13, offset: 6514},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 212, col: 13, offset: 6514},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 212, col: 20, offset: 6521},
								expr: &seqExpr{
									pos: position{line: 212, col: 21, offset: 6522},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 212, col: 21, offset: 6522},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 212, col: 31, offset: 6532},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 212, col: 36, offset: 6537},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 212, col: 41, offset: 6542},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 212, col: 52, offset: 6553},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 212, col: 54, offset: 6555},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 212, col: 60, offset: 6561},
								expr: &seqExpr{
									pos: position{line: 212, col: 61, offset: 6562},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 212, col: 61, offset: 6562},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 212, col: 65, offset: 6566},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 212, col: 67, offset: 6568},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 212, col: 81, offset: 6582},
							expr: &ruleRefExpr{
								pos:  position{line: 212, col: 81, offset: 6582},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 227, col: 1, offset: 6918},
			expr: &actionExpr{
				pos: position{line: 227, col: 11, offset: 6930},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 227, col: 11, offset: 6930},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 227, col: 11, offset: 6930},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 227, col: 21, offset: 6940},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 227, col: 23, offset: 6942},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 227, col: 27, offset: 6946},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 227, col: 37, offset: 6956},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 227, col: 39, offset: 6958},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 227, col: 44, offset: 6963},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 227, col: 55, offset: 6974},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 234, col: 1, offset: 7083},
			expr: &actionExpr{
				pos: position{line: 234, col: 10, offset: 7094},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 234, col: 10, offset: 7094},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 234, col: 10, offset: 7094},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 234, col: 19, offset: 7103},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 234, col: 21, offset: 7105},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 234, col: 24, offset: 7108},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 235, col: 1, offset: 7148},
			expr: &actionExpr{
				pos: position{line: 235, col: 13, offset: 7162},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 235, col: 13, offset: 7162},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 13, offset: 7162},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 25, offset: 7174},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 27, offset: 7176},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 30, offset: 7179},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 236, col: 1, offset: 7230},
			expr: &actionExpr{
				pos: position{line: 236, col: 9, offset: 7240},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 236, col: 9, offset: 7240},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 236, col: 9, offset: 7240},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 236, col: 17, offset: 7248},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 236, col: 19, offset: 7250},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 236, col: 22, offset: 7253},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 237, col: 1, offset: 7300},
			expr: &actionExpr{
				pos: position{line: 237, col: 14, offset: 7315},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 237, col: 14, offset: 7315},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 237, col: 14, offset: 7315},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 19, offset: 7320},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 30, offset: 7331},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 237, col: 33, offset: 7334},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 37, offset: 7338},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 40, offset: 7341},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 47, offset: 7348},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 237, col: 57, offset: 7358},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 61, offset: 7362},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 247, col: 1, offset: 7523},
			expr: &actionExpr{
				pos: position{line: 247, col: 13, offset: 7537},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 247, col: 13, offset: 7537},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 247, col: 20, offset: 7544},
						expr: &seqExpr{
							pos: position{line: 247, col: 21, offset: 7545},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 247, col: 21, offset: 7545},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 247, col: 27, offset: 7551},
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
			pos:  position{line: 256, col: 1, offset: 7732},
			expr: &actionExpr{
				pos: position{line: 256, col: 9, offset: 7742},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 256, col: 9, offset: 7742},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 256, col: 9, offset: 7742},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 256, col: 16, offset: 7749},
								expr: &seqExpr{
									pos: position{line: 256, col: 17, offset: 7750},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 256, col: 17, offset: 7750},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 256, col: 27, offset: 7760},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 256, col: 32, offset: 7765},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 35, offset: 7768},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 47, offset: 7780},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 256, col: 49, offset: 7782},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 53, offset: 7786},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 55, offset: 7788},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 256, col: 59, offset: 7792},
								expr: &ruleRefExpr{
									pos:  position{line: 256, col: 59, offset: 7792},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 69, offset: 7802},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 71, offset: 7804},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 75, offset: 7808},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 85, offset: 7818},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 87, offset: 7820},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 92, offset: 7825},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 256, col: 103, offset: 7836},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 256, col: 106, offset: 7839},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 256, col: 110, offset: 7843},
								expr: &seqExpr{
									pos: position{line: 256, col: 111, offset: 7844},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 256, col: 111, offset: 7844},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 256, col: 115, offset: 7848},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 256, col: 117, offset: 7850},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 256, col: 130, offset: 7863},
							expr: &ruleRefExpr{
								pos:  position{line: 256, col: 130, offset: 7863},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 275, col: 1, offset: 8282},
			expr: &actionExpr{
				pos: position{line: 275, col: 12, offset: 8295},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 275, col: 13, offset: 8296},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 275, col: 13, offset: 8296},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 275, col: 26, offset: 8309},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 279, col: 1, offset: 8383},
			expr: &actionExpr{
				pos: position{line: 279, col: 11, offset: 8395},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 279, col: 11, offset: 8395},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 279, col: 11, offset: 8395},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 279, col: 21, offset: 8405},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 279, col: 23, offset: 8407},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 279, col: 28, offset: 8412},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 279, col: 39, offset: 8423},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 279, col: 41, offset: 8425},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 279, col: 49, offset: 8433},
								expr: &seqExpr{
									pos: position{line: 279, col: 50, offset: 8434},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 279, col: 50, offset: 8434},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 279, col: 60, offset: 8444},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 279, col: 63, offset: 8447},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 279, col: 74, offset: 8458},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 279, col: 79, offset: 8463},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 279, col: 82, offset: 8466},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 279, col: 86, offset: 8470},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 279, col: 89, offset: 8473},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 279, col: 97, offset: 8481},
								expr: &seqExpr{
									pos: position{line: 279, col: 98, offset: 8482},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 279, col: 98, offset: 8482},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 279, col: 107, offset: 8491},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 279, col: 113, offset: 8497},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 279, col: 113, offset: 8497},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 279, col: 119, offset: 8503},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 279, col: 138, offset: 8522},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 294, col: 1, offset: 8917},
			expr: &actionExpr{
				pos: position{line: 294, col: 21, offset: 8939},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 294, col: 21, offset: 8939,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 298, col: 1, offset: 9008},
			expr: &actionExpr{
				pos: position{line: 298, col: 12, offset: 9021},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 298, col: 12, offset: 9021},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 298, col: 12, offset: 9021},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 298, col: 19, offset: 9028},
								expr: &seqExpr{
									pos: position{line: 298, col: 20, offset: 9029},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 298, col: 20, offset: 9029},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 298, col: 30, offset: 9039},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 298, col: 35, offset: 9044},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 298, col: 42, offset: 9051},
								expr: &seqExpr{
									pos: position{line: 298, col: 43, offset: 9052},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 298, col: 43, offset: 9052},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 298, col: 52, offset: 9061},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 298, col: 57, offset: 9066},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 298, col: 61, offset: 9070},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 298, col: 74, offset: 9083},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 298, col: 77, offset: 9086},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 298, col: 82, offset: 9091},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 298, col: 93, offset: 9102},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 298, col: 95, offset: 9104},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 298, col: 99, offset: 9108},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 298, col: 102, offset: 9111},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 298, col: 112, offset: 9121},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 298, col: 122, offset: 9131},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 298, col: 126, offset: 9135},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 298, col: 129, offset: 9138},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 298, col: 140, offset: 9149},
								expr: &ruleRefExpr{
									pos:  position{line: 298, col: 140, offset: 9149},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 298, col: 148, offset: 9157},
							expr: &ruleRefExpr{
								pos:  position{line: 298, col: 148, offset: 9157},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 325, col: 1, offset: 9748},
			expr: &actionExpr{
				pos: position{line: 325, col: 16, offset: 9765},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 325, col: 16, offset: 9765},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 325, col: 21, offset: 9770},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 325, col: 21, offset: 9770},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 325, col: 30, offset: 9779},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 332, col: 1, offset: 9901},
			expr: &actionExpr{
				pos: position{line: 332, col: 10, offset: 9912},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 332, col: 10, offset: 9912},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 332, col: 10, offset: 9912},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 332, col: 19, offset: 9921},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 332, col: 22, offset: 9924},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 332, col: 26, offset: 9928},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 332, col: 29, offset: 9931},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 332, col: 40, offset: 9942},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 332, col: 50, offset: 9952},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 336, col: 1, offset: 9988},
			expr: &actionExpr{
				pos: position{line: 336, col: 13, offset: 10002},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 336, col: 13, offset: 10002},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 336, col: 18, offset: 10007},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 336, col: 18, offset: 10007},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 336, col: 29, offset: 10018},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 336, col: 45, offset: 10034},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "DefinitionType",
			pos:  position{line: 343, col: 1, offset: 10159},
			expr: &actionExpr{
				pos: position{line: 343, col: 18, offset: 10178},
				run: (*parser).callonDefinitionType1,
				expr: &labeledExpr{
					pos:   position{line: 343, col: 18, offset: 10178},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 343, col: 23, offset: 10183},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 343, col: 23, offset: 10183},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 343, col: 34, offset: 10194},
								name: "ContainerType",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 347, col: 1, offset: 10234},
			expr: &actionExpr{
				pos: position{line: 347, col: 12, offset: 10247},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 347, col: 13, offset: 10248},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 347, col: 13, offset: 10248},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 22, offset: 10257},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 31, offset: 10266},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 39, offset: 10274},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 47, offset: 10282},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 55, offset: 10290},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 66, offset: 10301},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 347, col: 77, offset: 10312},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 351, col: 1, offset: 10372},
			expr: &actionExpr{
				pos: position{line: 351, col: 17, offset: 10390},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 351, col: 17, offset: 10390},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 351, col: 22, offset: 10395},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 351, col: 22, offset: 10395},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 351, col: 32, offset: 10405},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 351, col: 42, offset: 10415},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 355, col: 1, offset: 10450},
			expr: &actionExpr{
				pos: position{line: 355, col: 11, offset: 10462},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 355, col: 11, offset: 10462},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 355, col: 11, offset: 10462},
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 11, offset: 10462},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 355, col: 20, offset: 10471},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 27, offset: 10478},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 355, col: 30, offset: 10481},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 34, offset: 10485},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 44, offset: 10495},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 355, col: 47, offset: 10498},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 51, offset: 10502},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 355, col: 54, offset: 10505},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 355, col: 60, offset: 10511},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 355, col: 70, offset: 10521},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 355, col: 73, offset: 10524},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 363, col: 1, offset: 10647},
			expr: &actionExpr{
				pos: position{line: 363, col: 11, offset: 10659},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 363, col: 11, offset: 10659},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 363, col: 11, offset: 10659},
							expr: &ruleRefExpr{
								pos:  position{line: 363, col: 11, offset: 10659},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 363, col: 20, offset: 10668},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 363, col: 27, offset: 10675},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 363, col: 30, offset: 10678},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 363, col: 34, offset: 10682},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 363, col: 44, offset: 10692},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 363, col: 47, offset: 10695},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 370, col: 1, offset: 10786},
			expr: &actionExpr{
				pos: position{line: 370, col: 12, offset: 10799},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 370, col: 12, offset: 10799},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 370, col: 12, offset: 10799},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 20, offset: 10807},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 370, col: 23, offset: 10810},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 370, col: 27, offset: 10814},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 370, col: 37, offset: 10824},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 370, col: 40, offset: 10827},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 377, col: 1, offset: 10919},
			expr: &actionExpr{
				pos: position{line: 377, col: 11, offset: 10931},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 377, col: 11, offset: 10931},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 377, col: 11, offset: 10931},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 377, col: 22, offset: 10942},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 377, col: 30, offset: 10950},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 381, col: 1, offset: 10987},
			expr: &choiceExpr{
				pos: position{line: 381, col: 14, offset: 11002},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 381, col: 14, offset: 11002},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 24, offset: 11012},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 41, offset: 11029},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 55, offset: 11043},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 66, offset: 11054},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 381, col: 78, offset: 11066},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 383, col: 1, offset: 11078},
			expr: &actionExpr{
				pos: position{line: 383, col: 15, offset: 11094},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 383, col: 15, offset: 11094},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 383, col: 15, offset: 11094},
							expr: &charClassMatcher{
								pos:        position{line: 383, col: 15, offset: 11094},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 383, col: 21, offset: 11100},
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 21, offset: 11100},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 387, col: 1, offset: 11164},
			expr: &actionExpr{
				pos: position{line: 387, col: 18, offset: 11183},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 387, col: 18, offset: 11183},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 387, col: 18, offset: 11183},
							expr: &charClassMatcher{
								pos:        position{line: 387, col: 18, offset: 11183},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 387, col: 24, offset: 11189},
							expr: &ruleRefExpr{
								pos:  position{line: 387, col: 24, offset: 11189},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 387, col: 31, offset: 11196},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 387, col: 35, offset: 11200},
							expr: &ruleRefExpr{
								pos:  position{line: 387, col: 35, offset: 11200},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 387, col: 42, offset: 11207},
							expr: &seqExpr{
								pos: position{line: 387, col: 44, offset: 11209},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 387, col: 44, offset: 11209},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 387, col: 51, offset: 11216},
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
			pos:  position{line: 391, col: 1, offset: 11286},
			expr: &actionExpr{
				pos: position{line: 391, col: 13, offset: 11300},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 391, col: 13, offset: 11300},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 391, col: 13, offset: 11300},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 17, offset: 11304},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 391, col: 20, offset: 11307},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 391, col: 27, offset: 11314},
								expr: &seqExpr{
									pos: position{line: 391, col: 28, offset: 11315},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 391, col: 28, offset: 11315},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 391, col: 39, offset: 11326},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 391, col: 42, offset: 11329},
											expr: &ruleRefExpr{
												pos:  position{line: 391, col: 42, offset: 11329},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 391, col: 57, offset: 11344},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 62, offset: 11349},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 391, col: 65, offset: 11352},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 400, col: 1, offset: 11546},
			expr: &actionExpr{
				pos: position{line: 400, col: 12, offset: 11559},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 400, col: 12, offset: 11559},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 400, col: 12, offset: 11559},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 16, offset: 11563},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 400, col: 19, offset: 11566},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 400, col: 26, offset: 11573},
								expr: &seqExpr{
									pos: position{line: 400, col: 27, offset: 11574},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 400, col: 27, offset: 11574},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 38, offset: 11585},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 400, col: 41, offset: 11588},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 45, offset: 11592},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 48, offset: 11595},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 59, offset: 11606},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 400, col: 63, offset: 11610},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 400, col: 63, offset: 11610},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 400, col: 69, offset: 11616},
													expr: &litMatcher{
														pos:        position{line: 400, col: 70, offset: 11617},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 75, offset: 11622},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 400, col: 80, offset: 11627},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 420, col: 1, offset: 12177},
			expr: &ruleRefExpr{
				pos:  position{line: 420, col: 19, offset: 12197},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 422, col: 1, offset: 12204},
			expr: &actionExpr{
				pos: position{line: 422, col: 9, offset: 12214},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 422, col: 9, offset: 12214},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 422, col: 9, offset: 12214},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 422, col: 16, offset: 12221},
								expr: &seqExpr{
									pos: position{line: 422, col: 17, offset: 12222},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 422, col: 17, offset: 12222},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 422, col: 27, offset: 12232},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 422, col: 32, offset: 12237},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 40, offset: 12245},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 422, col: 43, offset: 12248},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 422, col: 48, offset: 12253},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 59, offset: 12264},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 422, col: 62, offset: 12267},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 66, offset: 12271},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 422, col: 69, offset: 12274},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 422, col: 76, offset: 12281},
								expr: &ruleRefExpr{
									pos:  position{line: 422, col: 76, offset: 12281},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 84, offset: 12289},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 422, col: 87, offset: 12292},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 422, col: 98, offset: 12303},
								expr: &seqExpr{
									pos: position{line: 422, col: 99, offset: 12304},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 422, col: 99, offset: 12304},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 422, col: 109, offset: 12314},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 422, col: 115, offset: 12320},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 422, col: 115, offset: 12320},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 422, col: 121, offset: 12326},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 138, offset: 12343},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 443, col: 1, offset: 12888},
			expr: &actionExpr{
				pos: position{line: 443, col: 19, offset: 12908},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 443, col: 19, offset: 12908,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 447, col: 1, offset: 12975},
			expr: &actionExpr{
				pos: position{line: 447, col: 10, offset: 12986},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 447, col: 10, offset: 12986},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 447, col: 10, offset: 12986},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 19, offset: 12995},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 447, col: 21, offset: 12997},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 447, col: 26, offset: 13002},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 447, col: 34, offset: 13010},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 451, col: 1, offset: 13060},
			expr: &actionExpr{
				pos: position{line: 451, col: 13, offset: 13074},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 451, col: 13, offset: 13074},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 451, col: 13, offset: 13074},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 451, col: 20, offset: 13081},
								expr: &seqExpr{
									pos: position{line: 451, col: 21, offset: 13082},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 451, col: 21, offset: 13082},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 451, col: 31, offset: 13092},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 451, col: 36, offset: 13097},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 451, col: 41, offset: 13102},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 451, col: 52, offset: 13113},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 451, col: 54, offset: 13115},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 451, col: 58, offset: 13119},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 451, col: 61, offset: 13122},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 451, col: 67, offset: 13128},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 451, col: 78, offset: 13139},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 467, col: 1, offset: 13641},
			expr: &actionExpr{
				pos: position{line: 467, col: 11, offset: 13653},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 467, col: 12, offset: 13654},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 467, col: 13, offset: 13655},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 467, col: 13, offset: 13655},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 467, col: 17, offset: 13659},
									expr: &choiceExpr{
										pos: position{line: 467, col: 18, offset: 13660},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 467, col: 18, offset: 13660},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 467, col: 25, offset: 13667},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 467, col: 32, offset: 13674},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 467, col: 40, offset: 13682},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 467, col: 40, offset: 13682},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 467, col: 45, offset: 13687},
									expr: &choiceExpr{
										pos: position{line: 467, col: 46, offset: 13688},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 467, col: 46, offset: 13688},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 467, col: 53, offset: 13695},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 467, col: 60, offset: 13702},
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
			pos:  position{line: 474, col: 1, offset: 13918},
			expr: &actionExpr{
				pos: position{line: 474, col: 14, offset: 13933},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 474, col: 14, offset: 13933},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 474, col: 14, offset: 13933},
							expr: &choiceExpr{
								pos: position{line: 474, col: 15, offset: 13934},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 474, col: 15, offset: 13934},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 474, col: 24, offset: 13943},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 474, col: 30, offset: 13949},
							expr: &choiceExpr{
								pos: position{line: 474, col: 31, offset: 13950},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 474, col: 31, offset: 13950},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 474, col: 40, offset: 13959},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 474, col: 48, offset: 13967},
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
			pos:  position{line: 478, col: 1, offset: 14022},
			expr: &charClassMatcher{
				pos:        position{line: 478, col: 17, offset: 14040},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 479, col: 1, offset: 14045},
			expr: &charClassMatcher{
				pos:        position{line: 479, col: 10, offset: 14056},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 480, col: 1, offset: 14065},
			expr: &charClassMatcher{
				pos:        position{line: 480, col: 9, offset: 14075},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 482, col: 1, offset: 14082},
			expr: &anyMatcher{
				line: 482, col: 14, offset: 14097,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 483, col: 1, offset: 14099},
			expr: &actionExpr{
				pos: position{line: 483, col: 13, offset: 14113},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 483, col: 13, offset: 14113},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 483, col: 13, offset: 14113},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 483, col: 20, offset: 14120},
							expr: &seqExpr{
								pos: position{line: 483, col: 22, offset: 14122},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 483, col: 22, offset: 14122},
										expr: &litMatcher{
											pos:        position{line: 483, col: 23, offset: 14123},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 483, col: 28, offset: 14128},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 483, col: 42, offset: 14142},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 489, col: 1, offset: 14322},
			expr: &choiceExpr{
				pos: position{line: 489, col: 11, offset: 14334},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 489, col: 11, offset: 14334},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 489, col: 30, offset: 14353},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 490, col: 1, offset: 14371},
			expr: &seqExpr{
				pos: position{line: 490, col: 20, offset: 14392},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 490, col: 20, offset: 14392},
						expr: &ruleRefExpr{
							pos:  position{line: 490, col: 21, offset: 14393},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 490, col: 31, offset: 14403},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 490, col: 36, offset: 14408},
						expr: &seqExpr{
							pos: position{line: 490, col: 38, offset: 14410},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 490, col: 38, offset: 14410},
									expr: &litMatcher{
										pos:        position{line: 490, col: 39, offset: 14411},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 490, col: 44, offset: 14416},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 490, col: 58, offset: 14430},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 491, col: 1, offset: 14435},
			expr: &seqExpr{
				pos: position{line: 491, col: 36, offset: 14472},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 491, col: 36, offset: 14472},
						expr: &ruleRefExpr{
							pos:  position{line: 491, col: 37, offset: 14473},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 491, col: 47, offset: 14483},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 491, col: 52, offset: 14488},
						expr: &seqExpr{
							pos: position{line: 491, col: 54, offset: 14490},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 491, col: 54, offset: 14490},
									expr: &choiceExpr{
										pos: position{line: 491, col: 57, offset: 14493},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 491, col: 57, offset: 14493},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 491, col: 64, offset: 14500},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 491, col: 70, offset: 14506},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 491, col: 84, offset: 14520},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 492, col: 1, offset: 14525},
			expr: &choiceExpr{
				pos: position{line: 492, col: 21, offset: 14547},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 492, col: 22, offset: 14548},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 492, col: 22, offset: 14548},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 492, col: 27, offset: 14553},
								expr: &seqExpr{
									pos: position{line: 492, col: 29, offset: 14555},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 492, col: 29, offset: 14555},
											expr: &ruleRefExpr{
												pos:  position{line: 492, col: 30, offset: 14556},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 492, col: 34, offset: 14560},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 492, col: 52, offset: 14578},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 492, col: 52, offset: 14578},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 492, col: 56, offset: 14582},
								expr: &seqExpr{
									pos: position{line: 492, col: 58, offset: 14584},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 492, col: 58, offset: 14584},
											expr: &ruleRefExpr{
												pos:  position{line: 492, col: 59, offset: 14585},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 492, col: 63, offset: 14589},
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
			pos:  position{line: 494, col: 1, offset: 14605},
			expr: &zeroOrMoreExpr{
				pos: position{line: 494, col: 6, offset: 14612},
				expr: &choiceExpr{
					pos: position{line: 494, col: 8, offset: 14614},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 494, col: 8, offset: 14614},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 494, col: 21, offset: 14627},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 494, col: 27, offset: 14633},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 495, col: 1, offset: 14644},
			expr: &zeroOrMoreExpr{
				pos: position{line: 495, col: 5, offset: 14650},
				expr: &choiceExpr{
					pos: position{line: 495, col: 7, offset: 14652},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 495, col: 7, offset: 14652},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 495, col: 20, offset: 14665},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 496, col: 1, offset: 14701},
			expr: &zeroOrMoreExpr{
				pos: position{line: 496, col: 6, offset: 14708},
				expr: &ruleRefExpr{
					pos:  position{line: 496, col: 6, offset: 14708},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 498, col: 1, offset: 14721},
			expr: &charClassMatcher{
				pos:        position{line: 498, col: 14, offset: 14736},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 499, col: 1, offset: 14744},
			expr: &litMatcher{
				pos:        position{line: 499, col: 7, offset: 14752},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 500, col: 1, offset: 14757},
			expr: &choiceExpr{
				pos: position{line: 500, col: 7, offset: 14765},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 500, col: 7, offset: 14765},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 500, col: 7, offset: 14765},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 500, col: 10, offset: 14768},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 500, col: 16, offset: 14774},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 500, col: 16, offset: 14774},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 500, col: 18, offset: 14776},
								expr: &ruleRefExpr{
									pos:  position{line: 500, col: 18, offset: 14776},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 500, col: 37, offset: 14795},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 500, col: 43, offset: 14801},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 500, col: 43, offset: 14801},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 500, col: 46, offset: 14804},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 502, col: 1, offset: 14809},
			expr: &notExpr{
				pos: position{line: 502, col: 7, offset: 14817},
				expr: &anyMatcher{
					line: 502, col: 8, offset: 14818,
				},
			},
		},
	},
}

func (c *current) onGrammar1(statements interface{}) (interface{}, error) {
	thrift := &Thrift{
		Includes:     []*Include{},
		Namespaces:   make(map[string]string),
		Typedefs:     []*TypeDef{},
		Constants:    make(map[string]*Constant),
		Enums:        make(map[string]*Enum),
		Structs:      make(map[string]*Struct),
		Exceptions:   make(map[string]*Struct),
		Unions:       make(map[string]*Struct),
		Services:     make(map[string]*Service),
		typedefIndex: make(map[string]*TypeDef),
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
			thrift.Typedefs = append(thrift.Typedefs, v)
			thrift.typedefIndex[v.Name] = v
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
