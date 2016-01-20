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
			pos:  position{line: 152, col: 1, offset: 4845},
			expr: &actionExpr{
				pos: position{line: 152, col: 15, offset: 4861},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 152, col: 15, offset: 4861,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 156, col: 1, offset: 4919},
			expr: &actionExpr{
				pos: position{line: 156, col: 13, offset: 4933},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 156, col: 13, offset: 4933},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 156, col: 13, offset: 4933},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 156, col: 20, offset: 4940},
								expr: &seqExpr{
									pos: position{line: 156, col: 21, offset: 4941},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 156, col: 21, offset: 4941},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 156, col: 31, offset: 4951},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 156, col: 36, offset: 4956},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 156, col: 47, offset: 4967},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 156, col: 47, offset: 4967},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 156, col: 65, offset: 4985},
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
			pos:  position{line: 169, col: 1, offset: 5456},
			expr: &choiceExpr{
				pos: position{line: 169, col: 19, offset: 5476},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 169, col: 19, offset: 5476},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 29, offset: 5486},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 41, offset: 5498},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 49, offset: 5506},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 56, offset: 5513},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 66, offset: 5523},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 75, offset: 5532},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 87, offset: 5544},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 169, col: 95, offset: 5552},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 171, col: 1, offset: 5561},
			expr: &actionExpr{
				pos: position{line: 171, col: 11, offset: 5573},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 171, col: 11, offset: 5573},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 171, col: 11, offset: 5573},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 21, offset: 5583},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 171, col: 23, offset: 5585},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 171, col: 28, offset: 5590},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 171, col: 36, offset: 5598},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 175, col: 1, offset: 5646},
			expr: &actionExpr{
				pos: position{line: 175, col: 13, offset: 5660},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 175, col: 13, offset: 5660},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 13, offset: 5660},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 25, offset: 5672},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 27, offset: 5674},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 175, col: 33, offset: 5680},
								expr: &charClassMatcher{
									pos:        position{line: 175, col: 33, offset: 5680},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 42, offset: 5689},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 44, offset: 5691},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 47, offset: 5694},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 58, offset: 5705},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 182, col: 1, offset: 5830},
			expr: &actionExpr{
				pos: position{line: 182, col: 9, offset: 5840},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 182, col: 9, offset: 5840},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 182, col: 9, offset: 5840},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 17, offset: 5848},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 19, offset: 5850},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 23, offset: 5854},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 33, offset: 5864},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 35, offset: 5866},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 40, offset: 5871},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 51, offset: 5882},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 182, col: 53, offset: 5884},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 57, offset: 5888},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 182, col: 59, offset: 5890},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 182, col: 65, offset: 5896},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 182, col: 76, offset: 5907},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 190, col: 1, offset: 6039},
			expr: &actionExpr{
				pos: position{line: 190, col: 8, offset: 6048},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 190, col: 8, offset: 6048},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 190, col: 8, offset: 6048},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 15, offset: 6055},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 17, offset: 6057},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 22, offset: 6062},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 33, offset: 6073},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 190, col: 36, offset: 6076},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 40, offset: 6080},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 43, offset: 6083},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 190, col: 50, offset: 6090},
								expr: &seqExpr{
									pos: position{line: 190, col: 51, offset: 6091},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 190, col: 51, offset: 6091},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 190, col: 61, offset: 6101},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 190, col: 66, offset: 6106},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 70, offset: 6110},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 213, col: 1, offset: 6722},
			expr: &actionExpr{
				pos: position{line: 213, col: 13, offset: 6736},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 213, col: 13, offset: 6736},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 213, col: 13, offset: 6736},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 20, offset: 6743},
								expr: &seqExpr{
									pos: position{line: 213, col: 21, offset: 6744},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 213, col: 21, offset: 6744},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 31, offset: 6754},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 213, col: 36, offset: 6759},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 41, offset: 6764},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 213, col: 52, offset: 6775},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 213, col: 54, offset: 6777},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 213, col: 60, offset: 6783},
								expr: &seqExpr{
									pos: position{line: 213, col: 61, offset: 6784},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 213, col: 61, offset: 6784},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 65, offset: 6788},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 213, col: 67, offset: 6790},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 213, col: 81, offset: 6804},
							expr: &ruleRefExpr{
								pos:  position{line: 213, col: 81, offset: 6804},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 228, col: 1, offset: 7140},
			expr: &actionExpr{
				pos: position{line: 228, col: 11, offset: 7152},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 228, col: 11, offset: 7152},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 11, offset: 7152},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 21, offset: 7162},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 23, offset: 7164},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 27, offset: 7168},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 37, offset: 7178},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 39, offset: 7180},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 44, offset: 7185},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 55, offset: 7196},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 235, col: 1, offset: 7305},
			expr: &actionExpr{
				pos: position{line: 235, col: 10, offset: 7316},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 235, col: 10, offset: 7316},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 235, col: 10, offset: 7316},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 235, col: 19, offset: 7325},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 235, col: 21, offset: 7327},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 235, col: 24, offset: 7330},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 236, col: 1, offset: 7370},
			expr: &actionExpr{
				pos: position{line: 236, col: 13, offset: 7384},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 236, col: 13, offset: 7384},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 236, col: 13, offset: 7384},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 236, col: 25, offset: 7396},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 236, col: 27, offset: 7398},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 236, col: 30, offset: 7401},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 237, col: 1, offset: 7452},
			expr: &actionExpr{
				pos: position{line: 237, col: 9, offset: 7462},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 237, col: 9, offset: 7462},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 237, col: 9, offset: 7462},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 237, col: 17, offset: 7470},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 237, col: 19, offset: 7472},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 237, col: 22, offset: 7475},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 238, col: 1, offset: 7522},
			expr: &actionExpr{
				pos: position{line: 238, col: 14, offset: 7537},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 238, col: 14, offset: 7537},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 238, col: 14, offset: 7537},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 19, offset: 7542},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 30, offset: 7553},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 238, col: 33, offset: 7556},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 37, offset: 7560},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7563},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 47, offset: 7570},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 238, col: 57, offset: 7580},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 61, offset: 7584},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 248, col: 1, offset: 7745},
			expr: &actionExpr{
				pos: position{line: 248, col: 13, offset: 7759},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 248, col: 13, offset: 7759},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 248, col: 20, offset: 7766},
						expr: &seqExpr{
							pos: position{line: 248, col: 21, offset: 7767},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 248, col: 21, offset: 7767},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 248, col: 27, offset: 7773},
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
			pos:  position{line: 257, col: 1, offset: 7954},
			expr: &actionExpr{
				pos: position{line: 257, col: 9, offset: 7964},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 257, col: 9, offset: 7964},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 257, col: 9, offset: 7964},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 16, offset: 7971},
								expr: &seqExpr{
									pos: position{line: 257, col: 17, offset: 7972},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 257, col: 17, offset: 7972},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 27, offset: 7982},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 257, col: 32, offset: 7987},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 35, offset: 7990},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 47, offset: 8002},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 257, col: 49, offset: 8004},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 53, offset: 8008},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 55, offset: 8010},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 59, offset: 8014},
								expr: &ruleRefExpr{
									pos:  position{line: 257, col: 59, offset: 8014},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 69, offset: 8024},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 71, offset: 8026},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 75, offset: 8030},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 85, offset: 8040},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 87, offset: 8042},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 92, offset: 8047},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 257, col: 103, offset: 8058},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 257, col: 106, offset: 8061},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 257, col: 110, offset: 8065},
								expr: &seqExpr{
									pos: position{line: 257, col: 111, offset: 8066},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 257, col: 111, offset: 8066},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 115, offset: 8070},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 257, col: 117, offset: 8072},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 257, col: 130, offset: 8085},
							expr: &ruleRefExpr{
								pos:  position{line: 257, col: 130, offset: 8085},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 276, col: 1, offset: 8504},
			expr: &actionExpr{
				pos: position{line: 276, col: 12, offset: 8517},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 276, col: 13, offset: 8518},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 276, col: 13, offset: 8518},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 276, col: 26, offset: 8531},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 280, col: 1, offset: 8605},
			expr: &actionExpr{
				pos: position{line: 280, col: 11, offset: 8617},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 280, col: 11, offset: 8617},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 280, col: 11, offset: 8617},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 21, offset: 8627},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 280, col: 23, offset: 8629},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 280, col: 28, offset: 8634},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 39, offset: 8645},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 280, col: 41, offset: 8647},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 280, col: 49, offset: 8655},
								expr: &seqExpr{
									pos: position{line: 280, col: 50, offset: 8656},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 280, col: 50, offset: 8656},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 60, offset: 8666},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 63, offset: 8669},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 74, offset: 8680},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 79, offset: 8685},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 280, col: 82, offset: 8688},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 86, offset: 8692},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 280, col: 89, offset: 8695},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 280, col: 97, offset: 8703},
								expr: &seqExpr{
									pos: position{line: 280, col: 98, offset: 8704},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 280, col: 98, offset: 8704},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 280, col: 107, offset: 8713},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 280, col: 113, offset: 8719},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 280, col: 113, offset: 8719},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 280, col: 119, offset: 8725},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 280, col: 138, offset: 8744},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 296, col: 1, offset: 9128},
			expr: &actionExpr{
				pos: position{line: 296, col: 21, offset: 9150},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 296, col: 21, offset: 9150,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 300, col: 1, offset: 9219},
			expr: &actionExpr{
				pos: position{line: 300, col: 12, offset: 9232},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 300, col: 12, offset: 9232},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 300, col: 12, offset: 9232},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 19, offset: 9239},
								expr: &seqExpr{
									pos: position{line: 300, col: 20, offset: 9240},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 300, col: 20, offset: 9240},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 30, offset: 9250},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 300, col: 35, offset: 9255},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 42, offset: 9262},
								expr: &seqExpr{
									pos: position{line: 300, col: 43, offset: 9263},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 300, col: 43, offset: 9263},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 52, offset: 9272},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 300, col: 57, offset: 9277},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 61, offset: 9281},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 74, offset: 9294},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 77, offset: 9297},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 82, offset: 9302},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 93, offset: 9313},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 300, col: 95, offset: 9315},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 99, offset: 9319},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 102, offset: 9322},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 112, offset: 9332},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 300, col: 122, offset: 9342},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 126, offset: 9346},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 129, offset: 9349},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 140, offset: 9360},
								expr: &ruleRefExpr{
									pos:  position{line: 300, col: 140, offset: 9360},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 300, col: 148, offset: 9368},
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 148, offset: 9368},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 327, col: 1, offset: 9959},
			expr: &actionExpr{
				pos: position{line: 327, col: 16, offset: 9976},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 327, col: 16, offset: 9976},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 327, col: 21, offset: 9981},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 327, col: 21, offset: 9981},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 327, col: 30, offset: 9990},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 334, col: 1, offset: 10112},
			expr: &actionExpr{
				pos: position{line: 334, col: 10, offset: 10123},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 334, col: 10, offset: 10123},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 334, col: 10, offset: 10123},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 334, col: 19, offset: 10132},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 334, col: 22, offset: 10135},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 334, col: 26, offset: 10139},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 334, col: 29, offset: 10142},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 334, col: 40, offset: 10153},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 334, col: 50, offset: 10163},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 338, col: 1, offset: 10199},
			expr: &actionExpr{
				pos: position{line: 338, col: 13, offset: 10213},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 338, col: 13, offset: 10213},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 338, col: 18, offset: 10218},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 338, col: 18, offset: 10218},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 338, col: 29, offset: 10229},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 338, col: 45, offset: 10245},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 345, col: 1, offset: 10370},
			expr: &actionExpr{
				pos: position{line: 345, col: 12, offset: 10383},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 345, col: 13, offset: 10384},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 345, col: 13, offset: 10384},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 22, offset: 10393},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 31, offset: 10402},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 39, offset: 10410},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 47, offset: 10418},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 55, offset: 10426},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 66, offset: 10437},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 345, col: 77, offset: 10448},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 349, col: 1, offset: 10508},
			expr: &actionExpr{
				pos: position{line: 349, col: 17, offset: 10526},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 349, col: 17, offset: 10526},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 349, col: 22, offset: 10531},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 349, col: 22, offset: 10531},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 32, offset: 10541},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 42, offset: 10551},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 353, col: 1, offset: 10586},
			expr: &actionExpr{
				pos: position{line: 353, col: 11, offset: 10598},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 353, col: 11, offset: 10598},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 353, col: 11, offset: 10598},
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 11, offset: 10598},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 353, col: 20, offset: 10607},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 27, offset: 10614},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 353, col: 30, offset: 10617},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 34, offset: 10621},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 44, offset: 10631},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 353, col: 47, offset: 10634},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 51, offset: 10638},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 353, col: 54, offset: 10641},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 353, col: 60, offset: 10647},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 353, col: 70, offset: 10657},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 353, col: 73, offset: 10660},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 361, col: 1, offset: 10783},
			expr: &actionExpr{
				pos: position{line: 361, col: 11, offset: 10795},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 361, col: 11, offset: 10795},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 361, col: 11, offset: 10795},
							expr: &ruleRefExpr{
								pos:  position{line: 361, col: 11, offset: 10795},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 361, col: 20, offset: 10804},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 361, col: 27, offset: 10811},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 361, col: 30, offset: 10814},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 361, col: 34, offset: 10818},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 361, col: 44, offset: 10828},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 361, col: 47, offset: 10831},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 368, col: 1, offset: 10922},
			expr: &actionExpr{
				pos: position{line: 368, col: 12, offset: 10935},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 368, col: 12, offset: 10935},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 368, col: 12, offset: 10935},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 20, offset: 10943},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 368, col: 23, offset: 10946},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 27, offset: 10950},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 37, offset: 10960},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 368, col: 40, offset: 10963},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 375, col: 1, offset: 11055},
			expr: &actionExpr{
				pos: position{line: 375, col: 11, offset: 11067},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 375, col: 11, offset: 11067},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 375, col: 11, offset: 11067},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 375, col: 22, offset: 11078},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 375, col: 30, offset: 11086},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 379, col: 1, offset: 11123},
			expr: &choiceExpr{
				pos: position{line: 379, col: 14, offset: 11138},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 379, col: 14, offset: 11138},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 24, offset: 11148},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 41, offset: 11165},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 55, offset: 11179},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 66, offset: 11190},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 379, col: 78, offset: 11202},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 381, col: 1, offset: 11214},
			expr: &actionExpr{
				pos: position{line: 381, col: 15, offset: 11230},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 381, col: 15, offset: 11230},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 381, col: 15, offset: 11230},
							expr: &charClassMatcher{
								pos:        position{line: 381, col: 15, offset: 11230},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 381, col: 21, offset: 11236},
							expr: &ruleRefExpr{
								pos:  position{line: 381, col: 21, offset: 11236},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 385, col: 1, offset: 11300},
			expr: &actionExpr{
				pos: position{line: 385, col: 18, offset: 11319},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 385, col: 18, offset: 11319},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 385, col: 18, offset: 11319},
							expr: &charClassMatcher{
								pos:        position{line: 385, col: 18, offset: 11319},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 385, col: 24, offset: 11325},
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 24, offset: 11325},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 385, col: 31, offset: 11332},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 385, col: 35, offset: 11336},
							expr: &ruleRefExpr{
								pos:  position{line: 385, col: 35, offset: 11336},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 385, col: 42, offset: 11343},
							expr: &seqExpr{
								pos: position{line: 385, col: 44, offset: 11345},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 385, col: 44, offset: 11345},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 385, col: 51, offset: 11352},
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
			pos:  position{line: 389, col: 1, offset: 11422},
			expr: &actionExpr{
				pos: position{line: 389, col: 13, offset: 11436},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 389, col: 13, offset: 11436},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 389, col: 13, offset: 11436},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 17, offset: 11440},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 389, col: 20, offset: 11443},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 389, col: 27, offset: 11450},
								expr: &seqExpr{
									pos: position{line: 389, col: 28, offset: 11451},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 389, col: 28, offset: 11451},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 389, col: 39, offset: 11462},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 389, col: 42, offset: 11465},
											expr: &ruleRefExpr{
												pos:  position{line: 389, col: 42, offset: 11465},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 389, col: 57, offset: 11480},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 389, col: 62, offset: 11485},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 389, col: 65, offset: 11488},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 398, col: 1, offset: 11682},
			expr: &actionExpr{
				pos: position{line: 398, col: 12, offset: 11695},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 398, col: 12, offset: 11695},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 398, col: 12, offset: 11695},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 398, col: 16, offset: 11699},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 398, col: 19, offset: 11702},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 398, col: 26, offset: 11709},
								expr: &seqExpr{
									pos: position{line: 398, col: 27, offset: 11710},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 398, col: 27, offset: 11710},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 38, offset: 11721},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 398, col: 41, offset: 11724},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 45, offset: 11728},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 48, offset: 11731},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 59, offset: 11742},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 398, col: 63, offset: 11746},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 398, col: 63, offset: 11746},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 398, col: 69, offset: 11752},
													expr: &litMatcher{
														pos:        position{line: 398, col: 70, offset: 11753},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 398, col: 75, offset: 11758},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 398, col: 80, offset: 11763},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 418, col: 1, offset: 12313},
			expr: &choiceExpr{
				pos: position{line: 418, col: 19, offset: 12333},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 418, col: 19, offset: 12333},
						name: "Scope",
					},
					&ruleRefExpr{
						pos:  position{line: 418, col: 27, offset: 12341},
						name: "Async",
					},
				},
			},
		},
		{
			name: "Scope",
			pos:  position{line: 420, col: 1, offset: 12348},
			expr: &actionExpr{
				pos: position{line: 420, col: 9, offset: 12358},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 420, col: 9, offset: 12358},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 420, col: 9, offset: 12358},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 420, col: 16, offset: 12365},
								expr: &seqExpr{
									pos: position{line: 420, col: 17, offset: 12366},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 420, col: 17, offset: 12366},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 420, col: 27, offset: 12376},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 420, col: 32, offset: 12381},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 40, offset: 12389},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 43, offset: 12392},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 420, col: 48, offset: 12397},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 59, offset: 12408},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 420, col: 62, offset: 12411},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 66, offset: 12415},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 69, offset: 12418},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 420, col: 76, offset: 12425},
								expr: &ruleRefExpr{
									pos:  position{line: 420, col: 76, offset: 12425},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 84, offset: 12433},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 420, col: 87, offset: 12436},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 420, col: 98, offset: 12447},
								expr: &seqExpr{
									pos: position{line: 420, col: 99, offset: 12448},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 420, col: 99, offset: 12448},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 420, col: 109, offset: 12458},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 420, col: 115, offset: 12464},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 420, col: 115, offset: 12464},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 420, col: 121, offset: 12470},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 420, col: 138, offset: 12487},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 441, col: 1, offset: 13032},
			expr: &actionExpr{
				pos: position{line: 441, col: 19, offset: 13052},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 441, col: 19, offset: 13052,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 445, col: 1, offset: 13119},
			expr: &actionExpr{
				pos: position{line: 445, col: 10, offset: 13130},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 445, col: 10, offset: 13130},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 445, col: 10, offset: 13130},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 19, offset: 13139},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 445, col: 21, offset: 13141},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 445, col: 26, offset: 13146},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 34, offset: 13154},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 449, col: 1, offset: 13204},
			expr: &actionExpr{
				pos: position{line: 449, col: 13, offset: 13218},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 449, col: 13, offset: 13218},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 449, col: 13, offset: 13218},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 449, col: 20, offset: 13225},
								expr: &seqExpr{
									pos: position{line: 449, col: 21, offset: 13226},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 449, col: 21, offset: 13226},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 449, col: 31, offset: 13236},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 449, col: 36, offset: 13241},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 449, col: 41, offset: 13246},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 52, offset: 13257},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 449, col: 54, offset: 13259},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 58, offset: 13263},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 449, col: 61, offset: 13266},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 449, col: 67, offset: 13272},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 449, col: 78, offset: 13283},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Async",
			pos:  position{line: 461, col: 1, offset: 13544},
			expr: &actionExpr{
				pos: position{line: 461, col: 9, offset: 13554},
				run: (*parser).callonAsync1,
				expr: &seqExpr{
					pos: position{line: 461, col: 9, offset: 13554},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 461, col: 9, offset: 13554},
							val:        "async",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 17, offset: 13562},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 19, offset: 13564},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 461, col: 24, offset: 13569},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 35, offset: 13580},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 37, offset: 13582},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 461, col: 45, offset: 13590},
								expr: &seqExpr{
									pos: position{line: 461, col: 46, offset: 13591},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 461, col: 46, offset: 13591},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 56, offset: 13601},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 59, offset: 13604},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 70, offset: 13615},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 75, offset: 13620},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 461, col: 78, offset: 13623},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 82, offset: 13627},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 461, col: 85, offset: 13630},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 461, col: 93, offset: 13638},
								expr: &seqExpr{
									pos: position{line: 461, col: 94, offset: 13639},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 461, col: 94, offset: 13639},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 461, col: 103, offset: 13648},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 461, col: 109, offset: 13654},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 461, col: 109, offset: 13654},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 461, col: 115, offset: 13660},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 461, col: 134, offset: 13679},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 481, col: 1, offset: 14310},
			expr: &actionExpr{
				pos: position{line: 481, col: 11, offset: 14322},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 481, col: 12, offset: 14323},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 481, col: 13, offset: 14324},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 481, col: 13, offset: 14324},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 481, col: 17, offset: 14328},
									expr: &choiceExpr{
										pos: position{line: 481, col: 18, offset: 14329},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 481, col: 18, offset: 14329},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 481, col: 25, offset: 14336},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 481, col: 32, offset: 14343},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 481, col: 40, offset: 14351},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 481, col: 40, offset: 14351},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 481, col: 45, offset: 14356},
									expr: &choiceExpr{
										pos: position{line: 481, col: 46, offset: 14357},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 481, col: 46, offset: 14357},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 481, col: 53, offset: 14364},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 481, col: 60, offset: 14371},
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
			pos:  position{line: 488, col: 1, offset: 14587},
			expr: &actionExpr{
				pos: position{line: 488, col: 14, offset: 14602},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 488, col: 14, offset: 14602},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 488, col: 14, offset: 14602},
							expr: &choiceExpr{
								pos: position{line: 488, col: 15, offset: 14603},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 488, col: 15, offset: 14603},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 488, col: 24, offset: 14612},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 488, col: 30, offset: 14618},
							expr: &choiceExpr{
								pos: position{line: 488, col: 31, offset: 14619},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 488, col: 31, offset: 14619},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 488, col: 40, offset: 14628},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 488, col: 48, offset: 14636},
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
			pos:  position{line: 492, col: 1, offset: 14691},
			expr: &charClassMatcher{
				pos:        position{line: 492, col: 17, offset: 14709},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 493, col: 1, offset: 14714},
			expr: &charClassMatcher{
				pos:        position{line: 493, col: 10, offset: 14725},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 494, col: 1, offset: 14734},
			expr: &charClassMatcher{
				pos:        position{line: 494, col: 9, offset: 14744},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 496, col: 1, offset: 14751},
			expr: &anyMatcher{
				line: 496, col: 14, offset: 14766,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 497, col: 1, offset: 14768},
			expr: &actionExpr{
				pos: position{line: 497, col: 13, offset: 14782},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 497, col: 13, offset: 14782},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 497, col: 13, offset: 14782},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 497, col: 20, offset: 14789},
							expr: &seqExpr{
								pos: position{line: 497, col: 22, offset: 14791},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 497, col: 22, offset: 14791},
										expr: &litMatcher{
											pos:        position{line: 497, col: 23, offset: 14792},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 497, col: 28, offset: 14797},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 497, col: 42, offset: 14811},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 503, col: 1, offset: 14991},
			expr: &choiceExpr{
				pos: position{line: 503, col: 11, offset: 15003},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 503, col: 11, offset: 15003},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 503, col: 30, offset: 15022},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 504, col: 1, offset: 15040},
			expr: &seqExpr{
				pos: position{line: 504, col: 20, offset: 15061},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 504, col: 20, offset: 15061},
						expr: &ruleRefExpr{
							pos:  position{line: 504, col: 21, offset: 15062},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 504, col: 31, offset: 15072},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 504, col: 36, offset: 15077},
						expr: &seqExpr{
							pos: position{line: 504, col: 38, offset: 15079},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 504, col: 38, offset: 15079},
									expr: &litMatcher{
										pos:        position{line: 504, col: 39, offset: 15080},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 504, col: 44, offset: 15085},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 504, col: 58, offset: 15099},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 505, col: 1, offset: 15104},
			expr: &seqExpr{
				pos: position{line: 505, col: 36, offset: 15141},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 505, col: 36, offset: 15141},
						expr: &ruleRefExpr{
							pos:  position{line: 505, col: 37, offset: 15142},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 505, col: 47, offset: 15152},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 505, col: 52, offset: 15157},
						expr: &seqExpr{
							pos: position{line: 505, col: 54, offset: 15159},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 505, col: 54, offset: 15159},
									expr: &choiceExpr{
										pos: position{line: 505, col: 57, offset: 15162},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 505, col: 57, offset: 15162},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 505, col: 64, offset: 15169},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 505, col: 70, offset: 15175},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 505, col: 84, offset: 15189},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 506, col: 1, offset: 15194},
			expr: &choiceExpr{
				pos: position{line: 506, col: 21, offset: 15216},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 506, col: 22, offset: 15217},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 506, col: 22, offset: 15217},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 506, col: 27, offset: 15222},
								expr: &seqExpr{
									pos: position{line: 506, col: 29, offset: 15224},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 506, col: 29, offset: 15224},
											expr: &ruleRefExpr{
												pos:  position{line: 506, col: 30, offset: 15225},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 506, col: 34, offset: 15229},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 506, col: 52, offset: 15247},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 506, col: 52, offset: 15247},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 506, col: 56, offset: 15251},
								expr: &seqExpr{
									pos: position{line: 506, col: 58, offset: 15253},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 506, col: 58, offset: 15253},
											expr: &ruleRefExpr{
												pos:  position{line: 506, col: 59, offset: 15254},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 506, col: 63, offset: 15258},
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
			pos:  position{line: 508, col: 1, offset: 15274},
			expr: &zeroOrMoreExpr{
				pos: position{line: 508, col: 6, offset: 15281},
				expr: &choiceExpr{
					pos: position{line: 508, col: 8, offset: 15283},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 508, col: 8, offset: 15283},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 21, offset: 15296},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 508, col: 27, offset: 15302},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 509, col: 1, offset: 15313},
			expr: &zeroOrMoreExpr{
				pos: position{line: 509, col: 5, offset: 15319},
				expr: &choiceExpr{
					pos: position{line: 509, col: 7, offset: 15321},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 509, col: 7, offset: 15321},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 509, col: 20, offset: 15334},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 510, col: 1, offset: 15370},
			expr: &zeroOrMoreExpr{
				pos: position{line: 510, col: 6, offset: 15377},
				expr: &ruleRefExpr{
					pos:  position{line: 510, col: 6, offset: 15377},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 512, col: 1, offset: 15390},
			expr: &charClassMatcher{
				pos:        position{line: 512, col: 14, offset: 15405},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 513, col: 1, offset: 15413},
			expr: &litMatcher{
				pos:        position{line: 513, col: 7, offset: 15421},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 514, col: 1, offset: 15426},
			expr: &choiceExpr{
				pos: position{line: 514, col: 7, offset: 15434},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 514, col: 7, offset: 15434},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 7, offset: 15434},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 514, col: 10, offset: 15437},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 514, col: 16, offset: 15443},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 16, offset: 15443},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 514, col: 18, offset: 15445},
								expr: &ruleRefExpr{
									pos:  position{line: 514, col: 18, offset: 15445},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 514, col: 37, offset: 15464},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 514, col: 43, offset: 15470},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 514, col: 43, offset: 15470},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 514, col: 46, offset: 15473},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 516, col: 1, offset: 15478},
			expr: &notExpr{
				pos: position{line: 516, col: 7, offset: 15486},
				expr: &anyMatcher{
					line: 516, col: 8, offset: 15487,
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
