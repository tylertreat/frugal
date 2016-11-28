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
				pos: position{line: 77, col: 12, offset: 2175},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 77, col: 12, offset: 2175},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 77, col: 12, offset: 2175},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 77, col: 15, offset: 2178},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 77, col: 26, offset: 2189},
								expr: &seqExpr{
									pos: position{line: 77, col: 28, offset: 2191},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 77, col: 28, offset: 2191},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 77, col: 38, offset: 2201},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 77, col: 45, offset: 2208},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 77, col: 45, offset: 2208},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 77, col: 51, offset: 2214},
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
			pos:  position{line: 148, col: 1, offset: 4822},
			expr: &actionExpr{
				pos: position{line: 148, col: 16, offset: 4837},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 148, col: 16, offset: 4837,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 152, col: 1, offset: 4895},
			expr: &actionExpr{
				pos: position{line: 152, col: 14, offset: 4908},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 152, col: 14, offset: 4908},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 152, col: 14, offset: 4908},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 152, col: 21, offset: 4915},
								expr: &seqExpr{
									pos: position{line: 152, col: 22, offset: 4916},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 152, col: 22, offset: 4916},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 152, col: 32, offset: 4926},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 152, col: 37, offset: 4931},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 152, col: 48, offset: 4942},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 152, col: 48, offset: 4942},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 152, col: 66, offset: 4960},
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
			pos:  position{line: 165, col: 1, offset: 5431},
			expr: &choiceExpr{
				pos: position{line: 165, col: 20, offset: 5450},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 165, col: 20, offset: 5450},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 30, offset: 5460},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 42, offset: 5472},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 50, offset: 5480},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 57, offset: 5487},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 67, offset: 5497},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 76, offset: 5506},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 88, offset: 5518},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 165, col: 96, offset: 5526},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 167, col: 1, offset: 5535},
			expr: &actionExpr{
				pos: position{line: 167, col: 12, offset: 5546},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 167, col: 12, offset: 5546},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 167, col: 12, offset: 5546},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 167, col: 22, offset: 5556},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 167, col: 24, offset: 5558},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 167, col: 29, offset: 5563},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 167, col: 37, offset: 5571},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 175, col: 1, offset: 5748},
			expr: &actionExpr{
				pos: position{line: 175, col: 14, offset: 5761},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 175, col: 14, offset: 5761},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 14, offset: 5761},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 26, offset: 5773},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 28, offset: 5775},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 175, col: 34, offset: 5781},
								expr: &charClassMatcher{
									pos:        position{line: 175, col: 34, offset: 5781},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 44, offset: 5791},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 46, offset: 5793},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 49, offset: 5796},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 60, offset: 5807},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 182, col: 1, offset: 5932},
			expr: &actionExpr{
				pos: position{line: 182, col: 10, offset: 5941},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 182, col: 10, offset: 5941},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 182, col: 10, offset: 5941},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 18, offset: 5949},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 20, offset: 5951},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 24, offset: 5955},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 34, offset: 5965},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 36, offset: 5967},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 41, offset: 5972},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 52, offset: 5983},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 182, col: 54, offset: 5985},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 58, offset: 5989},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 60, offset: 5991},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 66, offset: 5997},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 77, offset: 6008},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 190, col: 1, offset: 6140},
			expr: &actionExpr{
				pos: position{line: 190, col: 9, offset: 6148},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 190, col: 9, offset: 6148},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 190, col: 9, offset: 6148},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 16, offset: 6155},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 18, offset: 6157},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 23, offset: 6162},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 34, offset: 6173},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 190, col: 37, offset: 6176},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 41, offset: 6180},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 44, offset: 6183},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 190, col: 51, offset: 6190},
								expr: &seqExpr{
									pos: position{line: 190, col: 52, offset: 6191},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 190, col: 52, offset: 6191},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 190, col: 62, offset: 6201},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 190, col: 67, offset: 6206},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 71, offset: 6210},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 213, col: 1, offset: 6811},
			expr: &actionExpr{
				pos: position{line: 213, col: 14, offset: 6824},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 213, col: 14, offset: 6824},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 213, col: 14, offset: 6824},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 21, offset: 6831},
								expr: &seqExpr{
									pos: position{line: 213, col: 22, offset: 6832},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 213, col: 22, offset: 6832},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 32, offset: 6842},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 213, col: 37, offset: 6847},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 42, offset: 6852},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 213, col: 53, offset: 6863},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 213, col: 55, offset: 6865},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 61, offset: 6871},
								expr: &seqExpr{
									pos: position{line: 213, col: 62, offset: 6872},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 213, col: 62, offset: 6872},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 66, offset: 6876},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 68, offset: 6878},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 213, col: 82, offset: 6892},
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 82, offset: 6892},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 228, col: 1, offset: 7228},
			expr: &actionExpr{
				pos: position{line: 228, col: 12, offset: 7239},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 228, col: 12, offset: 7239},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 12, offset: 7239},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 22, offset: 7249},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 24, offset: 7251},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 28, offset: 7255},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 38, offset: 7265},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 40, offset: 7267},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 45, offset: 7272},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 56, offset: 7283},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 235, col: 1, offset: 7392},
			expr: &actionExpr{
				pos: position{line: 235, col: 11, offset: 7402},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 235, col: 11, offset: 7402},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 11, offset: 7402},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 20, offset: 7411},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 22, offset: 7413},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 25, offset: 7416},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 236, col: 1, offset: 7456},
			expr: &actionExpr{
				pos: position{line: 236, col: 14, offset: 7469},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 236, col: 14, offset: 7469},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 236, col: 14, offset: 7469},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 236, col: 26, offset: 7481},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 236, col: 28, offset: 7483},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 236, col: 31, offset: 7486},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 237, col: 1, offset: 7537},
			expr: &actionExpr{
				pos: position{line: 237, col: 10, offset: 7546},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 237, col: 10, offset: 7546},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 237, col: 10, offset: 7546},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 18, offset: 7554},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 20, offset: 7556},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 23, offset: 7559},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 238, col: 1, offset: 7606},
			expr: &actionExpr{
				pos: position{line: 238, col: 15, offset: 7620},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 238, col: 15, offset: 7620},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 238, col: 15, offset: 7620},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 20, offset: 7625},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 31, offset: 7636},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 238, col: 34, offset: 7639},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 38, offset: 7643},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 41, offset: 7646},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 48, offset: 7653},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 238, col: 58, offset: 7663},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 62, offset: 7667},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 248, col: 1, offset: 7828},
			expr: &actionExpr{
				pos: position{line: 248, col: 14, offset: 7841},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 248, col: 14, offset: 7841},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 248, col: 21, offset: 7848},
						expr: &seqExpr{
							pos: position{line: 248, col: 22, offset: 7849},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 248, col: 22, offset: 7849},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 248, col: 28, offset: 7855},
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
			pos:  position{line: 257, col: 1, offset: 8036},
			expr: &actionExpr{
				pos: position{line: 257, col: 10, offset: 8045},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 257, col: 10, offset: 8045},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 257, col: 10, offset: 8045},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 17, offset: 8052},
								expr: &seqExpr{
									pos: position{line: 257, col: 18, offset: 8053},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 257, col: 18, offset: 8053},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 28, offset: 8063},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 257, col: 33, offset: 8068},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 36, offset: 8071},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 48, offset: 8083},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 257, col: 50, offset: 8085},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 54, offset: 8089},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 56, offset: 8091},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 60, offset: 8095},
								expr: &ruleRefExpr{
									pos:  position{line: 257, col: 60, offset: 8095},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 75, offset: 8110},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 77, offset: 8112},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 81, offset: 8116},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 91, offset: 8126},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 93, offset: 8128},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 98, offset: 8133},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 109, offset: 8144},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 112, offset: 8147},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 116, offset: 8151},
								expr: &seqExpr{
									pos: position{line: 257, col: 117, offset: 8152},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 257, col: 117, offset: 8152},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 121, offset: 8156},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 123, offset: 8158},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 257, col: 136, offset: 8171},
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 136, offset: 8171},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 279, col: 1, offset: 8633},
			expr: &actionExpr{
				pos: position{line: 279, col: 18, offset: 8650},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 279, col: 19, offset: 8651},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 279, col: 19, offset: 8651},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 279, col: 32, offset: 8664},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 287, col: 1, offset: 8807},
			expr: &actionExpr{
				pos: position{line: 287, col: 12, offset: 8818},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 287, col: 12, offset: 8818},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 287, col: 12, offset: 8818},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 22, offset: 8828},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 24, offset: 8830},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 287, col: 29, offset: 8835},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 40, offset: 8846},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 42, offset: 8848},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 287, col: 50, offset: 8856},
								expr: &seqExpr{
									pos: position{line: 287, col: 51, offset: 8857},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 287, col: 51, offset: 8857},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 61, offset: 8867},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 64, offset: 8870},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 75, offset: 8881},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 80, offset: 8886},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 287, col: 83, offset: 8889},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 87, offset: 8893},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 287, col: 90, offset: 8896},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 287, col: 98, offset: 8904},
								expr: &seqExpr{
									pos: position{line: 287, col: 99, offset: 8905},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 287, col: 99, offset: 8905},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 287, col: 108, offset: 8914},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 287, col: 114, offset: 8920},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 287, col: 114, offset: 8920},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 287, col: 120, offset: 8926},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 287, col: 139, offset: 8945},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 303, col: 1, offset: 9329},
			expr: &actionExpr{
				pos: position{line: 303, col: 22, offset: 9350},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 303, col: 22, offset: 9350,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 307, col: 1, offset: 9419},
			expr: &actionExpr{
				pos: position{line: 307, col: 13, offset: 9431},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 307, col: 13, offset: 9431},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 307, col: 13, offset: 9431},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 20, offset: 9438},
								expr: &seqExpr{
									pos: position{line: 307, col: 21, offset: 9439},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 307, col: 21, offset: 9439},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 31, offset: 9449},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 307, col: 36, offset: 9454},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 43, offset: 9461},
								expr: &seqExpr{
									pos: position{line: 307, col: 44, offset: 9462},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 307, col: 44, offset: 9462},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 307, col: 53, offset: 9471},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 307, col: 58, offset: 9476},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 62, offset: 9480},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 75, offset: 9493},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 78, offset: 9496},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 83, offset: 9501},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 94, offset: 9512},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 307, col: 96, offset: 9514},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 100, offset: 9518},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 103, offset: 9521},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 113, offset: 9531},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 307, col: 123, offset: 9541},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 307, col: 127, offset: 9545},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 307, col: 130, offset: 9548},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 307, col: 141, offset: 9559},
								expr: &ruleRefExpr{
									pos:  position{line: 307, col: 141, offset: 9559},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 307, col: 149, offset: 9567},
							expr: &ruleRefExpr{
								pos:  position{line: 307, col: 149, offset: 9567},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 334, col: 1, offset: 10162},
			expr: &actionExpr{
				pos: position{line: 334, col: 17, offset: 10178},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 334, col: 17, offset: 10178},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 334, col: 22, offset: 10183},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 334, col: 22, offset: 10183},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 334, col: 31, offset: 10192},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 341, col: 1, offset: 10314},
			expr: &actionExpr{
				pos: position{line: 341, col: 11, offset: 10324},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 341, col: 11, offset: 10324},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 341, col: 11, offset: 10324},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 341, col: 20, offset: 10333},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 341, col: 23, offset: 10336},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 341, col: 27, offset: 10340},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 341, col: 30, offset: 10343},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 341, col: 41, offset: 10354},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 341, col: 51, offset: 10364},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 345, col: 1, offset: 10400},
			expr: &actionExpr{
				pos: position{line: 345, col: 14, offset: 10413},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 345, col: 14, offset: 10413},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 345, col: 19, offset: 10418},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 345, col: 19, offset: 10418},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 30, offset: 10429},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 46, offset: 10445},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 352, col: 1, offset: 10570},
			expr: &actionExpr{
				pos: position{line: 352, col: 13, offset: 10582},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 352, col: 14, offset: 10583},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 352, col: 14, offset: 10583},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 23, offset: 10592},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 32, offset: 10601},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 40, offset: 10609},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 48, offset: 10617},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 56, offset: 10625},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 67, offset: 10636},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 352, col: 78, offset: 10647},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 356, col: 1, offset: 10707},
			expr: &actionExpr{
				pos: position{line: 356, col: 18, offset: 10724},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 356, col: 18, offset: 10724},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 356, col: 23, offset: 10729},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 356, col: 23, offset: 10729},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 33, offset: 10739},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 356, col: 43, offset: 10749},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 360, col: 1, offset: 10784},
			expr: &actionExpr{
				pos: position{line: 360, col: 12, offset: 10795},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 360, col: 12, offset: 10795},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 360, col: 12, offset: 10795},
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 12, offset: 10795},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 360, col: 21, offset: 10804},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 28, offset: 10811},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 360, col: 31, offset: 10814},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 35, offset: 10818},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 45, offset: 10828},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 360, col: 48, offset: 10831},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 52, offset: 10835},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 360, col: 55, offset: 10838},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 360, col: 61, offset: 10844},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 360, col: 71, offset: 10854},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 360, col: 74, offset: 10857},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 368, col: 1, offset: 10980},
			expr: &actionExpr{
				pos: position{line: 368, col: 12, offset: 10991},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 368, col: 12, offset: 10991},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 368, col: 12, offset: 10991},
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 12, offset: 10991},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 368, col: 21, offset: 11000},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 28, offset: 11007},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 368, col: 31, offset: 11010},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 35, offset: 11014},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 45, offset: 11024},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 368, col: 48, offset: 11027},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 375, col: 1, offset: 11118},
			expr: &actionExpr{
				pos: position{line: 375, col: 13, offset: 11130},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 375, col: 13, offset: 11130},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 375, col: 13, offset: 11130},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 375, col: 21, offset: 11138},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 375, col: 24, offset: 11141},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 28, offset: 11145},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 375, col: 38, offset: 11155},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 375, col: 41, offset: 11158},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 382, col: 1, offset: 11250},
			expr: &actionExpr{
				pos: position{line: 382, col: 12, offset: 11261},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 382, col: 12, offset: 11261},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 382, col: 12, offset: 11261},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 382, col: 23, offset: 11272},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 31, offset: 11280},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 386, col: 1, offset: 11317},
			expr: &choiceExpr{
				pos: position{line: 386, col: 15, offset: 11331},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 386, col: 15, offset: 11331},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 25, offset: 11341},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 40, offset: 11356},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 57, offset: 11373},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 71, offset: 11387},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 82, offset: 11398},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 386, col: 94, offset: 11410},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 388, col: 1, offset: 11422},
			expr: &actionExpr{
				pos: position{line: 388, col: 17, offset: 11438},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 388, col: 18, offset: 11439},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 388, col: 18, offset: 11439},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 388, col: 27, offset: 11448},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 392, col: 1, offset: 11503},
			expr: &actionExpr{
				pos: position{line: 392, col: 16, offset: 11518},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 392, col: 16, offset: 11518},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 392, col: 16, offset: 11518},
							expr: &charClassMatcher{
								pos:        position{line: 392, col: 16, offset: 11518},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 392, col: 22, offset: 11524},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 22, offset: 11524},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 396, col: 1, offset: 11588},
			expr: &actionExpr{
				pos: position{line: 396, col: 19, offset: 11606},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 396, col: 19, offset: 11606},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 396, col: 19, offset: 11606},
							expr: &charClassMatcher{
								pos:        position{line: 396, col: 19, offset: 11606},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 396, col: 25, offset: 11612},
							expr: &ruleRefExpr{
								pos:  position{line: 396, col: 25, offset: 11612},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 396, col: 32, offset: 11619},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 396, col: 36, offset: 11623},
							expr: &ruleRefExpr{
								pos:  position{line: 396, col: 36, offset: 11623},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 396, col: 43, offset: 11630},
							expr: &seqExpr{
								pos: position{line: 396, col: 45, offset: 11632},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 396, col: 45, offset: 11632},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 396, col: 52, offset: 11639},
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
			pos:  position{line: 400, col: 1, offset: 11709},
			expr: &actionExpr{
				pos: position{line: 400, col: 14, offset: 11722},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 400, col: 14, offset: 11722},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 400, col: 14, offset: 11722},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 18, offset: 11726},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 400, col: 21, offset: 11729},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 400, col: 28, offset: 11736},
								expr: &seqExpr{
									pos: position{line: 400, col: 29, offset: 11737},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 400, col: 29, offset: 11737},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 40, offset: 11748},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 400, col: 43, offset: 11751},
											expr: &ruleRefExpr{
												pos:  position{line: 400, col: 43, offset: 11751},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 400, col: 58, offset: 11766},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 63, offset: 11771},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 400, col: 66, offset: 11774},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 409, col: 1, offset: 11968},
			expr: &actionExpr{
				pos: position{line: 409, col: 13, offset: 11980},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 409, col: 13, offset: 11980},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 409, col: 13, offset: 11980},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 409, col: 17, offset: 11984},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 409, col: 20, offset: 11987},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 409, col: 27, offset: 11994},
								expr: &seqExpr{
									pos: position{line: 409, col: 28, offset: 11995},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 409, col: 28, offset: 11995},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 39, offset: 12006},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 409, col: 42, offset: 12009},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 46, offset: 12013},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 49, offset: 12016},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 60, offset: 12027},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 409, col: 64, offset: 12031},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 409, col: 64, offset: 12031},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 409, col: 70, offset: 12037},
													expr: &litMatcher{
														pos:        position{line: 409, col: 71, offset: 12038},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 409, col: 76, offset: 12043},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 409, col: 81, offset: 12048},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 429, col: 1, offset: 12598},
			expr: &ruleRefExpr{
				pos:  position{line: 429, col: 20, offset: 12617},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 431, col: 1, offset: 12624},
			expr: &actionExpr{
				pos: position{line: 431, col: 10, offset: 12633},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 431, col: 10, offset: 12633},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 431, col: 10, offset: 12633},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 431, col: 17, offset: 12640},
								expr: &seqExpr{
									pos: position{line: 431, col: 18, offset: 12641},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 431, col: 18, offset: 12641},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 431, col: 28, offset: 12651},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 431, col: 33, offset: 12656},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 41, offset: 12664},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 431, col: 44, offset: 12667},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 431, col: 49, offset: 12672},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 60, offset: 12683},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 431, col: 63, offset: 12686},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 431, col: 70, offset: 12693},
								expr: &ruleRefExpr{
									pos:  position{line: 431, col: 70, offset: 12693},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 78, offset: 12701},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 431, col: 81, offset: 12704},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 85, offset: 12708},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 431, col: 88, offset: 12711},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 431, col: 99, offset: 12722},
								expr: &seqExpr{
									pos: position{line: 431, col: 100, offset: 12723},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 431, col: 100, offset: 12723},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 431, col: 110, offset: 12733},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 431, col: 116, offset: 12739},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 431, col: 116, offset: 12739},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 431, col: 122, offset: 12745},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 431, col: 139, offset: 12762},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 452, col: 1, offset: 13307},
			expr: &actionExpr{
				pos: position{line: 452, col: 20, offset: 13326},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 452, col: 20, offset: 13326,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 456, col: 1, offset: 13393},
			expr: &actionExpr{
				pos: position{line: 456, col: 11, offset: 13403},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 456, col: 11, offset: 13403},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 456, col: 11, offset: 13403},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 20, offset: 13412},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 456, col: 23, offset: 13415},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 456, col: 35, offset: 13427},
							expr: &seqExpr{
								pos: position{line: 456, col: 36, offset: 13428},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 456, col: 36, offset: 13428},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 456, col: 40, offset: 13432},
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
			pos:  position{line: 461, col: 1, offset: 13563},
			expr: &choiceExpr{
				pos: position{line: 461, col: 16, offset: 13578},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 461, col: 17, offset: 13579},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 461, col: 17, offset: 13579},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 461, col: 21, offset: 13583},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 461, col: 32, offset: 13594},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 461, col: 39, offset: 13601},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 463, col: 1, offset: 13613},
			expr: &oneOrMoreExpr{
				pos: position{line: 463, col: 15, offset: 13627},
				expr: &charClassMatcher{
					pos:        position{line: 463, col: 15, offset: 13627},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 465, col: 1, offset: 13645},
			expr: &actionExpr{
				pos: position{line: 465, col: 14, offset: 13658},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 465, col: 14, offset: 13658},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 465, col: 14, offset: 13658},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 465, col: 21, offset: 13665},
								expr: &seqExpr{
									pos: position{line: 465, col: 22, offset: 13666},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 465, col: 22, offset: 13666},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 465, col: 32, offset: 13676},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 465, col: 37, offset: 13681},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 465, col: 42, offset: 13686},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 465, col: 53, offset: 13697},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 465, col: 55, offset: 13699},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 465, col: 59, offset: 13703},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 465, col: 62, offset: 13706},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 465, col: 66, offset: 13710},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 465, col: 77, offset: 13721},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 481, col: 1, offset: 14232},
			expr: &actionExpr{
				pos: position{line: 481, col: 12, offset: 14243},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 481, col: 13, offset: 14244},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 481, col: 14, offset: 14245},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 481, col: 14, offset: 14245},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 481, col: 18, offset: 14249},
									expr: &choiceExpr{
										pos: position{line: 481, col: 19, offset: 14250},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 481, col: 19, offset: 14250},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 481, col: 26, offset: 14257},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 481, col: 33, offset: 14264},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 481, col: 41, offset: 14272},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 481, col: 41, offset: 14272},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 481, col: 46, offset: 14277},
									expr: &choiceExpr{
										pos: position{line: 481, col: 47, offset: 14278},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 481, col: 47, offset: 14278},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 481, col: 54, offset: 14285},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 481, col: 61, offset: 14292},
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
			pos:  position{line: 490, col: 1, offset: 14578},
			expr: &actionExpr{
				pos: position{line: 490, col: 15, offset: 14592},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 490, col: 15, offset: 14592},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 490, col: 15, offset: 14592},
							expr: &choiceExpr{
								pos: position{line: 490, col: 16, offset: 14593},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 490, col: 16, offset: 14593},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 490, col: 25, offset: 14602},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 490, col: 31, offset: 14608},
							expr: &choiceExpr{
								pos: position{line: 490, col: 32, offset: 14609},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 490, col: 32, offset: 14609},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 490, col: 41, offset: 14618},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 490, col: 49, offset: 14626},
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
			pos:  position{line: 494, col: 1, offset: 14681},
			expr: &charClassMatcher{
				pos:        position{line: 494, col: 18, offset: 14698},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 495, col: 1, offset: 14703},
			expr: &charClassMatcher{
				pos:        position{line: 495, col: 11, offset: 14713},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 496, col: 1, offset: 14722},
			expr: &charClassMatcher{
				pos:        position{line: 496, col: 10, offset: 14731},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 498, col: 1, offset: 14738},
			expr: &anyMatcher{
				line: 498, col: 15, offset: 14752,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 499, col: 1, offset: 14754},
			expr: &actionExpr{
				pos: position{line: 499, col: 14, offset: 14767},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 499, col: 14, offset: 14767},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 499, col: 14, offset: 14767},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 499, col: 21, offset: 14774},
							expr: &seqExpr{
								pos: position{line: 499, col: 23, offset: 14776},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 499, col: 23, offset: 14776},
										expr: &litMatcher{
											pos:        position{line: 499, col: 24, offset: 14777},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 499, col: 29, offset: 14782},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 499, col: 43, offset: 14796},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 505, col: 1, offset: 14976},
			expr: &choiceExpr{
				pos: position{line: 505, col: 12, offset: 14987},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 505, col: 12, offset: 14987},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 505, col: 31, offset: 15006},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 506, col: 1, offset: 15024},
			expr: &seqExpr{
				pos: position{line: 506, col: 21, offset: 15044},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 506, col: 21, offset: 15044},
						expr: &ruleRefExpr{
							pos:  position{line: 506, col: 22, offset: 15045},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 506, col: 32, offset: 15055},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 506, col: 37, offset: 15060},
						expr: &seqExpr{
							pos: position{line: 506, col: 39, offset: 15062},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 506, col: 39, offset: 15062},
									expr: &litMatcher{
										pos:        position{line: 506, col: 40, offset: 15063},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 506, col: 45, offset: 15068},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 506, col: 59, offset: 15082},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 507, col: 1, offset: 15087},
			expr: &seqExpr{
				pos: position{line: 507, col: 37, offset: 15123},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 507, col: 37, offset: 15123},
						expr: &ruleRefExpr{
							pos:  position{line: 507, col: 38, offset: 15124},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 507, col: 48, offset: 15134},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 507, col: 53, offset: 15139},
						expr: &seqExpr{
							pos: position{line: 507, col: 55, offset: 15141},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 507, col: 55, offset: 15141},
									expr: &choiceExpr{
										pos: position{line: 507, col: 58, offset: 15144},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 507, col: 58, offset: 15144},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 507, col: 65, offset: 15151},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 507, col: 71, offset: 15157},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 507, col: 85, offset: 15171},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 508, col: 1, offset: 15176},
			expr: &choiceExpr{
				pos: position{line: 508, col: 22, offset: 15197},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 508, col: 23, offset: 15198},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 508, col: 23, offset: 15198},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 508, col: 28, offset: 15203},
								expr: &seqExpr{
									pos: position{line: 508, col: 30, offset: 15205},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 508, col: 30, offset: 15205},
											expr: &ruleRefExpr{
												pos:  position{line: 508, col: 31, offset: 15206},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 508, col: 35, offset: 15210},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 508, col: 53, offset: 15228},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 508, col: 53, offset: 15228},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 508, col: 57, offset: 15232},
								expr: &seqExpr{
									pos: position{line: 508, col: 59, offset: 15234},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 508, col: 59, offset: 15234},
											expr: &ruleRefExpr{
												pos:  position{line: 508, col: 60, offset: 15235},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 508, col: 64, offset: 15239},
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
			pos:  position{line: 510, col: 1, offset: 15255},
			expr: &zeroOrMoreExpr{
				pos: position{line: 510, col: 7, offset: 15261},
				expr: &choiceExpr{
					pos: position{line: 510, col: 9, offset: 15263},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 510, col: 9, offset: 15263},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 510, col: 22, offset: 15276},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 510, col: 28, offset: 15282},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 511, col: 1, offset: 15293},
			expr: &zeroOrMoreExpr{
				pos: position{line: 511, col: 6, offset: 15298},
				expr: &choiceExpr{
					pos: position{line: 511, col: 8, offset: 15300},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 511, col: 8, offset: 15300},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 511, col: 21, offset: 15313},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 512, col: 1, offset: 15349},
			expr: &zeroOrMoreExpr{
				pos: position{line: 512, col: 7, offset: 15355},
				expr: &ruleRefExpr{
					pos:  position{line: 512, col: 7, offset: 15355},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 514, col: 1, offset: 15368},
			expr: &charClassMatcher{
				pos:        position{line: 514, col: 15, offset: 15382},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 515, col: 1, offset: 15390},
			expr: &litMatcher{
				pos:        position{line: 515, col: 8, offset: 15397},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 516, col: 1, offset: 15402},
			expr: &choiceExpr{
				pos: position{line: 516, col: 8, offset: 15409},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 516, col: 8, offset: 15409},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 516, col: 8, offset: 15409},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 516, col: 11, offset: 15412},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 516, col: 17, offset: 15418},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 516, col: 17, offset: 15418},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 516, col: 19, offset: 15420},
								expr: &ruleRefExpr{
									pos:  position{line: 516, col: 19, offset: 15420},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 516, col: 38, offset: 15439},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 516, col: 44, offset: 15445},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 516, col: 44, offset: 15445},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 516, col: 47, offset: 15448},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 518, col: 1, offset: 15453},
			expr: &notExpr{
				pos: position{line: 518, col: 8, offset: 15460},
				expr: &anyMatcher{
					line: 518, col: 9, offset: 15461,
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
