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
			pos:  position{line: 149, col: 1, offset: 4463},
			expr: &actionExpr{
				pos: position{line: 149, col: 15, offset: 4479},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 149, col: 15, offset: 4479,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 153, col: 1, offset: 4537},
			expr: &actionExpr{
				pos: position{line: 153, col: 13, offset: 4551},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 153, col: 13, offset: 4551},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 153, col: 13, offset: 4551},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 153, col: 20, offset: 4558},
								expr: &seqExpr{
									pos: position{line: 153, col: 21, offset: 4559},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 153, col: 21, offset: 4559},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 153, col: 31, offset: 4569},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 153, col: 36, offset: 4574},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 153, col: 47, offset: 4585},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 153, col: 47, offset: 4585},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 153, col: 65, offset: 4603},
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
			pos:  position{line: 166, col: 1, offset: 5074},
			expr: &choiceExpr{
				pos: position{line: 166, col: 19, offset: 5094},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 166, col: 19, offset: 5094},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 29, offset: 5104},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 41, offset: 5116},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 49, offset: 5124},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 56, offset: 5131},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 66, offset: 5141},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 75, offset: 5150},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 87, offset: 5162},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 166, col: 95, offset: 5170},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 168, col: 1, offset: 5179},
			expr: &actionExpr{
				pos: position{line: 168, col: 11, offset: 5191},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 168, col: 11, offset: 5191},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 168, col: 11, offset: 5191},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 168, col: 21, offset: 5201},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 168, col: 23, offset: 5203},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 168, col: 28, offset: 5208},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 168, col: 36, offset: 5216},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 172, col: 1, offset: 5264},
			expr: &actionExpr{
				pos: position{line: 172, col: 13, offset: 5278},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 172, col: 13, offset: 5278},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 172, col: 13, offset: 5278},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 25, offset: 5290},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 172, col: 27, offset: 5292},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 172, col: 33, offset: 5298},
								expr: &charClassMatcher{
									pos:        position{line: 172, col: 33, offset: 5298},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 42, offset: 5307},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 172, col: 44, offset: 5309},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 172, col: 47, offset: 5312},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 172, col: 58, offset: 5323},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 179, col: 1, offset: 5456},
			expr: &actionExpr{
				pos: position{line: 179, col: 9, offset: 5466},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 179, col: 9, offset: 5466},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 179, col: 9, offset: 5466},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 17, offset: 5474},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 179, col: 19, offset: 5476},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 179, col: 23, offset: 5480},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 33, offset: 5490},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 179, col: 35, offset: 5492},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 179, col: 40, offset: 5497},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 51, offset: 5508},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 179, col: 53, offset: 5510},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 57, offset: 5514},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 179, col: 59, offset: 5516},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 179, col: 65, offset: 5522},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 76, offset: 5533},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 187, col: 1, offset: 5665},
			expr: &actionExpr{
				pos: position{line: 187, col: 8, offset: 5674},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 187, col: 8, offset: 5674},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 187, col: 8, offset: 5674},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 187, col: 15, offset: 5681},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 187, col: 17, offset: 5683},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 187, col: 22, offset: 5688},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 187, col: 33, offset: 5699},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 187, col: 36, offset: 5702},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 187, col: 40, offset: 5706},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 187, col: 43, offset: 5709},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 187, col: 50, offset: 5716},
								expr: &seqExpr{
									pos: position{line: 187, col: 51, offset: 5717},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 187, col: 51, offset: 5717},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 187, col: 61, offset: 5727},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 187, col: 66, offset: 5732},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 187, col: 70, offset: 5736},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 210, col: 1, offset: 6348},
			expr: &actionExpr{
				pos: position{line: 210, col: 13, offset: 6362},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 210, col: 13, offset: 6362},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 210, col: 13, offset: 6362},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 210, col: 18, offset: 6367},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 210, col: 29, offset: 6378},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 210, col: 31, offset: 6380},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 210, col: 37, offset: 6386},
								expr: &seqExpr{
									pos: position{line: 210, col: 38, offset: 6387},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 210, col: 38, offset: 6387},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 210, col: 42, offset: 6391},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 210, col: 44, offset: 6393},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 210, col: 58, offset: 6407},
							expr: &ruleRefExpr{
								pos:  position{line: 210, col: 58, offset: 6407},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 221, col: 1, offset: 6619},
			expr: &actionExpr{
				pos: position{line: 221, col: 11, offset: 6631},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 221, col: 11, offset: 6631},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 221, col: 11, offset: 6631},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 221, col: 21, offset: 6641},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 221, col: 23, offset: 6643},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 221, col: 27, offset: 6647},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 221, col: 37, offset: 6657},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 221, col: 39, offset: 6659},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 221, col: 44, offset: 6664},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 221, col: 55, offset: 6675},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 228, col: 1, offset: 6784},
			expr: &actionExpr{
				pos: position{line: 228, col: 10, offset: 6795},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 228, col: 10, offset: 6795},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 228, col: 10, offset: 6795},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 228, col: 19, offset: 6804},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 228, col: 21, offset: 6806},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 24, offset: 6809},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 229, col: 1, offset: 6849},
			expr: &actionExpr{
				pos: position{line: 229, col: 13, offset: 6863},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 229, col: 13, offset: 6863},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 229, col: 13, offset: 6863},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 25, offset: 6875},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 229, col: 27, offset: 6877},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 229, col: 30, offset: 6880},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 230, col: 1, offset: 6931},
			expr: &actionExpr{
				pos: position{line: 230, col: 9, offset: 6941},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 230, col: 9, offset: 6941},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 230, col: 9, offset: 6941},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 230, col: 17, offset: 6949},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 230, col: 19, offset: 6951},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 230, col: 22, offset: 6954},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 231, col: 1, offset: 7001},
			expr: &actionExpr{
				pos: position{line: 231, col: 14, offset: 7016},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 231, col: 14, offset: 7016},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 231, col: 14, offset: 7016},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 231, col: 19, offset: 7021},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 231, col: 30, offset: 7032},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 231, col: 33, offset: 7035},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 231, col: 37, offset: 7039},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 231, col: 40, offset: 7042},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 231, col: 47, offset: 7049},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 231, col: 57, offset: 7059},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 231, col: 61, offset: 7063},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 241, col: 1, offset: 7224},
			expr: &actionExpr{
				pos: position{line: 241, col: 13, offset: 7238},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 241, col: 13, offset: 7238},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 241, col: 20, offset: 7245},
						expr: &seqExpr{
							pos: position{line: 241, col: 21, offset: 7246},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 241, col: 21, offset: 7246},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 241, col: 27, offset: 7252},
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
			pos:  position{line: 250, col: 1, offset: 7433},
			expr: &actionExpr{
				pos: position{line: 250, col: 9, offset: 7443},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 250, col: 9, offset: 7443},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 250, col: 9, offset: 7443},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 250, col: 16, offset: 7450},
								expr: &seqExpr{
									pos: position{line: 250, col: 17, offset: 7451},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 250, col: 17, offset: 7451},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 250, col: 27, offset: 7461},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 250, col: 32, offset: 7466},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 35, offset: 7469},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 47, offset: 7481},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 250, col: 49, offset: 7483},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 53, offset: 7487},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 55, offset: 7489},
							label: "req",
							expr: &zeroOrOneExpr{
								pos: position{line: 250, col: 59, offset: 7493},
								expr: &ruleRefExpr{
									pos:  position{line: 250, col: 59, offset: 7493},
									name: "FieldReq",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 69, offset: 7503},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 71, offset: 7505},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 75, offset: 7509},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 85, offset: 7519},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 87, offset: 7521},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 92, offset: 7526},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 103, offset: 7537},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 106, offset: 7540},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 250, col: 110, offset: 7544},
								expr: &seqExpr{
									pos: position{line: 250, col: 111, offset: 7545},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 250, col: 111, offset: 7545},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 250, col: 115, offset: 7549},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 250, col: 117, offset: 7551},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 250, col: 130, offset: 7564},
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 130, offset: 7564},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldReq",
			pos:  position{line: 269, col: 1, offset: 7983},
			expr: &actionExpr{
				pos: position{line: 269, col: 12, offset: 7996},
				run: (*parser).callonFieldReq1,
				expr: &choiceExpr{
					pos: position{line: 269, col: 13, offset: 7997},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 269, col: 13, offset: 7997},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 269, col: 26, offset: 8010},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 273, col: 1, offset: 8084},
			expr: &actionExpr{
				pos: position{line: 273, col: 11, offset: 8096},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 273, col: 11, offset: 8096},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 273, col: 11, offset: 8096},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 273, col: 21, offset: 8106},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 273, col: 23, offset: 8108},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 273, col: 28, offset: 8113},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 273, col: 39, offset: 8124},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 273, col: 41, offset: 8126},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 273, col: 49, offset: 8134},
								expr: &seqExpr{
									pos: position{line: 273, col: 50, offset: 8135},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 273, col: 50, offset: 8135},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 273, col: 60, offset: 8145},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 273, col: 63, offset: 8148},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 273, col: 74, offset: 8159},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 273, col: 79, offset: 8164},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 273, col: 82, offset: 8167},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 273, col: 86, offset: 8171},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 273, col: 89, offset: 8174},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 273, col: 97, offset: 8182},
								expr: &seqExpr{
									pos: position{line: 273, col: 98, offset: 8183},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 273, col: 98, offset: 8183},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 273, col: 107, offset: 8192},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 273, col: 113, offset: 8198},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 273, col: 113, offset: 8198},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 273, col: 119, offset: 8204},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 273, col: 138, offset: 8223},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 288, col: 1, offset: 8618},
			expr: &actionExpr{
				pos: position{line: 288, col: 21, offset: 8640},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 288, col: 21, offset: 8640,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 292, col: 1, offset: 8709},
			expr: &actionExpr{
				pos: position{line: 292, col: 12, offset: 8722},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 292, col: 12, offset: 8722},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 292, col: 12, offset: 8722},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 292, col: 19, offset: 8729},
								expr: &seqExpr{
									pos: position{line: 292, col: 20, offset: 8730},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 292, col: 20, offset: 8730},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 292, col: 30, offset: 8740},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 292, col: 35, offset: 8745},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 292, col: 42, offset: 8752},
								expr: &seqExpr{
									pos: position{line: 292, col: 43, offset: 8753},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 292, col: 43, offset: 8753},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 292, col: 52, offset: 8762},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 292, col: 57, offset: 8767},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 292, col: 61, offset: 8771},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 292, col: 74, offset: 8784},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 292, col: 77, offset: 8787},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 292, col: 82, offset: 8792},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 292, col: 93, offset: 8803},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 292, col: 95, offset: 8805},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 292, col: 99, offset: 8809},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 292, col: 102, offset: 8812},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 292, col: 112, offset: 8822},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 292, col: 122, offset: 8832},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 292, col: 126, offset: 8836},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 292, col: 129, offset: 8839},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 292, col: 140, offset: 8850},
								expr: &ruleRefExpr{
									pos:  position{line: 292, col: 140, offset: 8850},
									name: "Throws",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 292, col: 148, offset: 8858},
							expr: &ruleRefExpr{
								pos:  position{line: 292, col: 148, offset: 8858},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 319, col: 1, offset: 9449},
			expr: &actionExpr{
				pos: position{line: 319, col: 16, offset: 9466},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 319, col: 16, offset: 9466},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 319, col: 21, offset: 9471},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 319, col: 21, offset: 9471},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 319, col: 30, offset: 9480},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 326, col: 1, offset: 9602},
			expr: &actionExpr{
				pos: position{line: 326, col: 10, offset: 9613},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 326, col: 10, offset: 9613},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 326, col: 10, offset: 9613},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 326, col: 19, offset: 9622},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 326, col: 22, offset: 9625},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 326, col: 26, offset: 9629},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 326, col: 29, offset: 9632},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 326, col: 40, offset: 9643},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 326, col: 50, offset: 9653},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 330, col: 1, offset: 9689},
			expr: &actionExpr{
				pos: position{line: 330, col: 13, offset: 9703},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 330, col: 13, offset: 9703},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 330, col: 18, offset: 9708},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 330, col: 18, offset: 9708},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 330, col: 29, offset: 9719},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 330, col: 45, offset: 9735},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "DefinitionType",
			pos:  position{line: 337, col: 1, offset: 9860},
			expr: &actionExpr{
				pos: position{line: 337, col: 18, offset: 9879},
				run: (*parser).callonDefinitionType1,
				expr: &labeledExpr{
					pos:   position{line: 337, col: 18, offset: 9879},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 337, col: 23, offset: 9884},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 337, col: 23, offset: 9884},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 337, col: 34, offset: 9895},
								name: "ContainerType",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 341, col: 1, offset: 9935},
			expr: &actionExpr{
				pos: position{line: 341, col: 12, offset: 9948},
				run: (*parser).callonBaseType1,
				expr: &choiceExpr{
					pos: position{line: 341, col: 13, offset: 9949},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 341, col: 13, offset: 9949},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 22, offset: 9958},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 31, offset: 9967},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 39, offset: 9975},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 47, offset: 9983},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 55, offset: 9991},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 66, offset: 10002},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 341, col: 77, offset: 10013},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 345, col: 1, offset: 10073},
			expr: &actionExpr{
				pos: position{line: 345, col: 17, offset: 10091},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 345, col: 17, offset: 10091},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 345, col: 22, offset: 10096},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 345, col: 22, offset: 10096},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 32, offset: 10106},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 345, col: 42, offset: 10116},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 349, col: 1, offset: 10151},
			expr: &actionExpr{
				pos: position{line: 349, col: 11, offset: 10163},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 349, col: 11, offset: 10163},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 349, col: 11, offset: 10163},
							expr: &ruleRefExpr{
								pos:  position{line: 349, col: 11, offset: 10163},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 349, col: 20, offset: 10172},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 349, col: 27, offset: 10179},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 349, col: 30, offset: 10182},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 349, col: 34, offset: 10186},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 349, col: 44, offset: 10196},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 349, col: 47, offset: 10199},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 349, col: 51, offset: 10203},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 349, col: 54, offset: 10206},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 349, col: 60, offset: 10212},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 349, col: 70, offset: 10222},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 349, col: 73, offset: 10225},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 357, col: 1, offset: 10348},
			expr: &actionExpr{
				pos: position{line: 357, col: 11, offset: 10360},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 357, col: 11, offset: 10360},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 357, col: 11, offset: 10360},
							expr: &ruleRefExpr{
								pos:  position{line: 357, col: 11, offset: 10360},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 357, col: 20, offset: 10369},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 357, col: 27, offset: 10376},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 357, col: 30, offset: 10379},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 357, col: 34, offset: 10383},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 357, col: 44, offset: 10393},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 357, col: 47, offset: 10396},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 364, col: 1, offset: 10487},
			expr: &actionExpr{
				pos: position{line: 364, col: 12, offset: 10500},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 364, col: 12, offset: 10500},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 364, col: 12, offset: 10500},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 364, col: 20, offset: 10508},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 364, col: 23, offset: 10511},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 364, col: 27, offset: 10515},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 364, col: 37, offset: 10525},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 364, col: 40, offset: 10528},
							val:        ">",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 371, col: 1, offset: 10620},
			expr: &actionExpr{
				pos: position{line: 371, col: 11, offset: 10632},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 371, col: 11, offset: 10632},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 371, col: 11, offset: 10632},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 371, col: 22, offset: 10643},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 371, col: 30, offset: 10651},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 375, col: 1, offset: 10688},
			expr: &choiceExpr{
				pos: position{line: 375, col: 14, offset: 10703},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 375, col: 14, offset: 10703},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 375, col: 24, offset: 10713},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 375, col: 41, offset: 10730},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 375, col: 55, offset: 10744},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 375, col: 66, offset: 10755},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 375, col: 78, offset: 10767},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 377, col: 1, offset: 10779},
			expr: &actionExpr{
				pos: position{line: 377, col: 15, offset: 10795},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 377, col: 15, offset: 10795},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 377, col: 15, offset: 10795},
							expr: &charClassMatcher{
								pos:        position{line: 377, col: 15, offset: 10795},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 377, col: 21, offset: 10801},
							expr: &ruleRefExpr{
								pos:  position{line: 377, col: 21, offset: 10801},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 381, col: 1, offset: 10865},
			expr: &actionExpr{
				pos: position{line: 381, col: 18, offset: 10884},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 381, col: 18, offset: 10884},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 381, col: 18, offset: 10884},
							expr: &charClassMatcher{
								pos:        position{line: 381, col: 18, offset: 10884},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 381, col: 24, offset: 10890},
							expr: &ruleRefExpr{
								pos:  position{line: 381, col: 24, offset: 10890},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 381, col: 31, offset: 10897},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 381, col: 35, offset: 10901},
							expr: &ruleRefExpr{
								pos:  position{line: 381, col: 35, offset: 10901},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 381, col: 42, offset: 10908},
							expr: &seqExpr{
								pos: position{line: 381, col: 44, offset: 10910},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 381, col: 44, offset: 10910},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 381, col: 51, offset: 10917},
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
			pos:  position{line: 385, col: 1, offset: 10987},
			expr: &actionExpr{
				pos: position{line: 385, col: 13, offset: 11001},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 385, col: 13, offset: 11001},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 385, col: 13, offset: 11001},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 17, offset: 11005},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 385, col: 20, offset: 11008},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 385, col: 27, offset: 11015},
								expr: &seqExpr{
									pos: position{line: 385, col: 28, offset: 11016},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 385, col: 28, offset: 11016},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 385, col: 39, offset: 11027},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 385, col: 42, offset: 11030},
											expr: &ruleRefExpr{
												pos:  position{line: 385, col: 42, offset: 11030},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 385, col: 57, offset: 11045},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 385, col: 62, offset: 11050},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 385, col: 65, offset: 11053},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 394, col: 1, offset: 11247},
			expr: &actionExpr{
				pos: position{line: 394, col: 12, offset: 11260},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 394, col: 12, offset: 11260},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 394, col: 12, offset: 11260},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 394, col: 16, offset: 11264},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 394, col: 19, offset: 11267},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 394, col: 26, offset: 11274},
								expr: &seqExpr{
									pos: position{line: 394, col: 27, offset: 11275},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 394, col: 27, offset: 11275},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 394, col: 38, offset: 11286},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 394, col: 41, offset: 11289},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 394, col: 45, offset: 11293},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 394, col: 48, offset: 11296},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 394, col: 59, offset: 11307},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 394, col: 63, offset: 11311},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 394, col: 63, offset: 11311},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 394, col: 69, offset: 11317},
													expr: &litMatcher{
														pos:        position{line: 394, col: 70, offset: 11318},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 394, col: 75, offset: 11323},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 394, col: 80, offset: 11328},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 414, col: 1, offset: 11878},
			expr: &ruleRefExpr{
				pos:  position{line: 414, col: 19, offset: 11898},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 416, col: 1, offset: 11905},
			expr: &actionExpr{
				pos: position{line: 416, col: 9, offset: 11915},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 416, col: 9, offset: 11915},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 416, col: 9, offset: 11915},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 416, col: 16, offset: 11922},
								expr: &seqExpr{
									pos: position{line: 416, col: 17, offset: 11923},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 416, col: 17, offset: 11923},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 416, col: 27, offset: 11933},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 416, col: 32, offset: 11938},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 416, col: 40, offset: 11946},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 416, col: 43, offset: 11949},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 416, col: 48, offset: 11954},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 416, col: 59, offset: 11965},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 416, col: 62, offset: 11968},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 416, col: 66, offset: 11972},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 416, col: 69, offset: 11975},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 416, col: 76, offset: 11982},
								expr: &ruleRefExpr{
									pos:  position{line: 416, col: 76, offset: 11982},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 416, col: 84, offset: 11990},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 416, col: 87, offset: 11993},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 416, col: 98, offset: 12004},
								expr: &seqExpr{
									pos: position{line: 416, col: 99, offset: 12005},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 416, col: 99, offset: 12005},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 416, col: 109, offset: 12015},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 416, col: 115, offset: 12021},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 416, col: 115, offset: 12021},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 416, col: 121, offset: 12027},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 416, col: 138, offset: 12044},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 437, col: 1, offset: 12589},
			expr: &actionExpr{
				pos: position{line: 437, col: 19, offset: 12609},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 437, col: 19, offset: 12609,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 441, col: 1, offset: 12676},
			expr: &actionExpr{
				pos: position{line: 441, col: 10, offset: 12687},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 441, col: 10, offset: 12687},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 441, col: 10, offset: 12687},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 441, col: 19, offset: 12696},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 441, col: 21, offset: 12698},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 441, col: 26, offset: 12703},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 441, col: 34, offset: 12711},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 445, col: 1, offset: 12761},
			expr: &actionExpr{
				pos: position{line: 445, col: 13, offset: 12775},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 445, col: 13, offset: 12775},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 445, col: 13, offset: 12775},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 445, col: 20, offset: 12782},
								expr: &seqExpr{
									pos: position{line: 445, col: 21, offset: 12783},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 445, col: 21, offset: 12783},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 445, col: 31, offset: 12793},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 445, col: 36, offset: 12798},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 445, col: 41, offset: 12803},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 52, offset: 12814},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 445, col: 54, offset: 12816},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 58, offset: 12820},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 445, col: 61, offset: 12823},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 445, col: 67, offset: 12829},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 78, offset: 12840},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 461, col: 1, offset: 13342},
			expr: &actionExpr{
				pos: position{line: 461, col: 11, offset: 13354},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 461, col: 12, offset: 13355},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 461, col: 13, offset: 13356},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 461, col: 13, offset: 13356},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 461, col: 17, offset: 13360},
									expr: &choiceExpr{
										pos: position{line: 461, col: 18, offset: 13361},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 461, col: 18, offset: 13361},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 461, col: 25, offset: 13368},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 461, col: 32, offset: 13375},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 461, col: 40, offset: 13383},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 461, col: 40, offset: 13383},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 461, col: 45, offset: 13388},
									expr: &choiceExpr{
										pos: position{line: 461, col: 46, offset: 13389},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 461, col: 46, offset: 13389},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 461, col: 53, offset: 13396},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 461, col: 60, offset: 13403},
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
			pos:  position{line: 468, col: 1, offset: 13619},
			expr: &actionExpr{
				pos: position{line: 468, col: 14, offset: 13634},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 468, col: 14, offset: 13634},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 468, col: 14, offset: 13634},
							expr: &choiceExpr{
								pos: position{line: 468, col: 15, offset: 13635},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 468, col: 15, offset: 13635},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 468, col: 24, offset: 13644},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 468, col: 30, offset: 13650},
							expr: &choiceExpr{
								pos: position{line: 468, col: 31, offset: 13651},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 468, col: 31, offset: 13651},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 468, col: 40, offset: 13660},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 468, col: 48, offset: 13668},
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
			pos:  position{line: 472, col: 1, offset: 13723},
			expr: &charClassMatcher{
				pos:        position{line: 472, col: 17, offset: 13741},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 473, col: 1, offset: 13746},
			expr: &charClassMatcher{
				pos:        position{line: 473, col: 10, offset: 13757},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 474, col: 1, offset: 13766},
			expr: &charClassMatcher{
				pos:        position{line: 474, col: 9, offset: 13776},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 476, col: 1, offset: 13783},
			expr: &anyMatcher{
				line: 476, col: 14, offset: 13798,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 477, col: 1, offset: 13800},
			expr: &actionExpr{
				pos: position{line: 477, col: 13, offset: 13814},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 477, col: 13, offset: 13814},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 477, col: 13, offset: 13814},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 477, col: 20, offset: 13821},
							expr: &seqExpr{
								pos: position{line: 477, col: 22, offset: 13823},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 477, col: 22, offset: 13823},
										expr: &litMatcher{
											pos:        position{line: 477, col: 23, offset: 13824},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 477, col: 28, offset: 13829},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 477, col: 42, offset: 13843},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 483, col: 1, offset: 14023},
			expr: &choiceExpr{
				pos: position{line: 483, col: 11, offset: 14035},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 483, col: 11, offset: 14035},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 483, col: 30, offset: 14054},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 484, col: 1, offset: 14072},
			expr: &seqExpr{
				pos: position{line: 484, col: 20, offset: 14093},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 484, col: 20, offset: 14093},
						expr: &ruleRefExpr{
							pos:  position{line: 484, col: 21, offset: 14094},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 484, col: 31, offset: 14104},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 484, col: 36, offset: 14109},
						expr: &seqExpr{
							pos: position{line: 484, col: 38, offset: 14111},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 484, col: 38, offset: 14111},
									expr: &litMatcher{
										pos:        position{line: 484, col: 39, offset: 14112},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 484, col: 44, offset: 14117},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 484, col: 58, offset: 14131},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 485, col: 1, offset: 14136},
			expr: &seqExpr{
				pos: position{line: 485, col: 36, offset: 14173},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 485, col: 36, offset: 14173},
						expr: &ruleRefExpr{
							pos:  position{line: 485, col: 37, offset: 14174},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 485, col: 47, offset: 14184},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 485, col: 52, offset: 14189},
						expr: &seqExpr{
							pos: position{line: 485, col: 54, offset: 14191},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 485, col: 54, offset: 14191},
									expr: &choiceExpr{
										pos: position{line: 485, col: 57, offset: 14194},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 485, col: 57, offset: 14194},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 485, col: 64, offset: 14201},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 485, col: 70, offset: 14207},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 485, col: 84, offset: 14221},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 486, col: 1, offset: 14226},
			expr: &choiceExpr{
				pos: position{line: 486, col: 21, offset: 14248},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 486, col: 22, offset: 14249},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 486, col: 22, offset: 14249},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 486, col: 27, offset: 14254},
								expr: &seqExpr{
									pos: position{line: 486, col: 29, offset: 14256},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 486, col: 29, offset: 14256},
											expr: &ruleRefExpr{
												pos:  position{line: 486, col: 30, offset: 14257},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 486, col: 34, offset: 14261},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 486, col: 52, offset: 14279},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 486, col: 52, offset: 14279},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 486, col: 56, offset: 14283},
								expr: &seqExpr{
									pos: position{line: 486, col: 58, offset: 14285},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 486, col: 58, offset: 14285},
											expr: &ruleRefExpr{
												pos:  position{line: 486, col: 59, offset: 14286},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 486, col: 63, offset: 14290},
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
			pos:  position{line: 488, col: 1, offset: 14306},
			expr: &zeroOrMoreExpr{
				pos: position{line: 488, col: 6, offset: 14313},
				expr: &choiceExpr{
					pos: position{line: 488, col: 8, offset: 14315},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 488, col: 8, offset: 14315},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 488, col: 21, offset: 14328},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 488, col: 27, offset: 14334},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 489, col: 1, offset: 14345},
			expr: &zeroOrMoreExpr{
				pos: position{line: 489, col: 5, offset: 14351},
				expr: &choiceExpr{
					pos: position{line: 489, col: 7, offset: 14353},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 489, col: 7, offset: 14353},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 489, col: 20, offset: 14366},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 490, col: 1, offset: 14402},
			expr: &zeroOrMoreExpr{
				pos: position{line: 490, col: 6, offset: 14409},
				expr: &ruleRefExpr{
					pos:  position{line: 490, col: 6, offset: 14409},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 492, col: 1, offset: 14422},
			expr: &charClassMatcher{
				pos:        position{line: 492, col: 14, offset: 14437},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 493, col: 1, offset: 14445},
			expr: &litMatcher{
				pos:        position{line: 493, col: 7, offset: 14453},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 494, col: 1, offset: 14458},
			expr: &choiceExpr{
				pos: position{line: 494, col: 7, offset: 14466},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 494, col: 7, offset: 14466},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 494, col: 7, offset: 14466},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 494, col: 10, offset: 14469},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 494, col: 16, offset: 14475},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 494, col: 16, offset: 14475},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 494, col: 18, offset: 14477},
								expr: &ruleRefExpr{
									pos:  position{line: 494, col: 18, offset: 14477},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 494, col: 37, offset: 14496},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 494, col: 43, offset: 14502},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 494, col: 43, offset: 14502},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 494, col: 46, offset: 14505},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 496, col: 1, offset: 14510},
			expr: &notExpr{
				pos: position{line: 496, col: 7, offset: 14518},
				expr: &anyMatcher{
					line: 496, col: 8, offset: 14519,
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
