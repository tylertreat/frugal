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

// toAnnotations converts an interface{} to an Annotation slice.
func toAnnotations(v interface{}) []*Annotation {
	if v == nil {
		return nil
	}
	return v.([]*Annotation)
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 85, col: 1, offset: 2382},
			expr: &actionExpr{
				pos: position{line: 85, col: 12, offset: 2393},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 85, col: 12, offset: 2393},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 85, col: 12, offset: 2393},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 15, offset: 2396},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 85, col: 26, offset: 2407},
								expr: &seqExpr{
									pos: position{line: 85, col: 28, offset: 2409},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 85, col: 28, offset: 2409},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 85, col: 38, offset: 2419},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 85, col: 45, offset: 2426},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 85, col: 45, offset: 2426},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 85, col: 51, offset: 2432},
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
			pos:  position{line: 156, col: 1, offset: 5040},
			expr: &actionExpr{
				pos: position{line: 156, col: 16, offset: 5055},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 156, col: 16, offset: 5055,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 160, col: 1, offset: 5113},
			expr: &actionExpr{
				pos: position{line: 160, col: 14, offset: 5126},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 160, col: 14, offset: 5126},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 160, col: 14, offset: 5126},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 160, col: 21, offset: 5133},
								expr: &seqExpr{
									pos: position{line: 160, col: 22, offset: 5134},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 160, col: 22, offset: 5134},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 160, col: 32, offset: 5144},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 160, col: 37, offset: 5149},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 160, col: 48, offset: 5160},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 160, col: 48, offset: 5160},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 160, col: 66, offset: 5178},
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
			pos:  position{line: 173, col: 1, offset: 5649},
			expr: &choiceExpr{
				pos: position{line: 173, col: 20, offset: 5668},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 173, col: 20, offset: 5668},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 30, offset: 5678},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 42, offset: 5690},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 50, offset: 5698},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 57, offset: 5705},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 67, offset: 5715},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 76, offset: 5724},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 88, offset: 5736},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 96, offset: 5744},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 175, col: 1, offset: 5753},
			expr: &actionExpr{
				pos: position{line: 175, col: 12, offset: 5764},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 175, col: 12, offset: 5764},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 12, offset: 5764},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 22, offset: 5774},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 24, offset: 5776},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 29, offset: 5781},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 37, offset: 5789},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 183, col: 1, offset: 5966},
			expr: &actionExpr{
				pos: position{line: 183, col: 14, offset: 5979},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 183, col: 14, offset: 5979},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 183, col: 14, offset: 5979},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 26, offset: 5991},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 183, col: 28, offset: 5993},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 183, col: 34, offset: 5999},
								expr: &charClassMatcher{
									pos:        position{line: 183, col: 34, offset: 5999},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 44, offset: 6009},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 183, col: 46, offset: 6011},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 183, col: 49, offset: 6014},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 60, offset: 6025},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 190, col: 1, offset: 6150},
			expr: &actionExpr{
				pos: position{line: 190, col: 10, offset: 6159},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 190, col: 10, offset: 6159},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 190, col: 10, offset: 6159},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 18, offset: 6167},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 20, offset: 6169},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 24, offset: 6173},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 34, offset: 6183},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 36, offset: 6185},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 41, offset: 6190},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 52, offset: 6201},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 190, col: 54, offset: 6203},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 58, offset: 6207},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 60, offset: 6209},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 66, offset: 6215},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 77, offset: 6226},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 198, col: 1, offset: 6358},
			expr: &actionExpr{
				pos: position{line: 198, col: 9, offset: 6366},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 198, col: 9, offset: 6366},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 198, col: 9, offset: 6366},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 16, offset: 6373},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 18, offset: 6375},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 198, col: 23, offset: 6380},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 34, offset: 6391},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 198, col: 37, offset: 6394},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 41, offset: 6398},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 44, offset: 6401},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 198, col: 51, offset: 6408},
								expr: &seqExpr{
									pos: position{line: 198, col: 52, offset: 6409},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 198, col: 52, offset: 6409},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 198, col: 62, offset: 6419},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 198, col: 67, offset: 6424},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 71, offset: 6428},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 198, col: 73, offset: 6430},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 198, col: 85, offset: 6442},
								expr: &ruleRefExpr{
									pos:  position{line: 198, col: 85, offset: 6442},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 198, col: 102, offset: 6459},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 222, col: 1, offset: 7121},
			expr: &actionExpr{
				pos: position{line: 222, col: 14, offset: 7134},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 222, col: 14, offset: 7134},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 222, col: 14, offset: 7134},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 21, offset: 7141},
								expr: &seqExpr{
									pos: position{line: 222, col: 22, offset: 7142},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 222, col: 22, offset: 7142},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 32, offset: 7152},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 222, col: 37, offset: 7157},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 222, col: 42, offset: 7162},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 222, col: 53, offset: 7173},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 222, col: 55, offset: 7175},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 61, offset: 7181},
								expr: &seqExpr{
									pos: position{line: 222, col: 62, offset: 7182},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 222, col: 62, offset: 7182},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 66, offset: 7186},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 222, col: 68, offset: 7188},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 222, col: 82, offset: 7202},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 222, col: 84, offset: 7204},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 222, col: 96, offset: 7216},
								expr: &ruleRefExpr{
									pos:  position{line: 222, col: 96, offset: 7216},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 222, col: 113, offset: 7233},
							expr: &ruleRefExpr{
								pos:  position{line: 222, col: 113, offset: 7233},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 238, col: 1, offset: 7631},
			expr: &actionExpr{
				pos: position{line: 238, col: 12, offset: 7642},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 238, col: 12, offset: 7642},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 238, col: 12, offset: 7642},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 22, offset: 7652},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 24, offset: 7654},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 28, offset: 7658},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 38, offset: 7668},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 40, offset: 7670},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 45, offset: 7675},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 56, offset: 7686},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 58, offset: 7688},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 238, col: 70, offset: 7700},
								expr: &ruleRefExpr{
									pos:  position{line: 238, col: 70, offset: 7700},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 87, offset: 7717},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 246, col: 1, offset: 7889},
			expr: &actionExpr{
				pos: position{line: 246, col: 11, offset: 7899},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 246, col: 11, offset: 7899},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 246, col: 11, offset: 7899},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 246, col: 20, offset: 7908},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 246, col: 22, offset: 7910},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 246, col: 25, offset: 7913},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 247, col: 1, offset: 7953},
			expr: &actionExpr{
				pos: position{line: 247, col: 14, offset: 7966},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 247, col: 14, offset: 7966},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 247, col: 14, offset: 7966},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 247, col: 26, offset: 7978},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 247, col: 28, offset: 7980},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 247, col: 31, offset: 7983},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 248, col: 1, offset: 8034},
			expr: &actionExpr{
				pos: position{line: 248, col: 10, offset: 8043},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 248, col: 10, offset: 8043},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 248, col: 10, offset: 8043},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 18, offset: 8051},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 20, offset: 8053},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 23, offset: 8056},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 249, col: 1, offset: 8103},
			expr: &actionExpr{
				pos: position{line: 249, col: 15, offset: 8117},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 249, col: 15, offset: 8117},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 249, col: 15, offset: 8117},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 20, offset: 8122},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 31, offset: 8133},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 249, col: 34, offset: 8136},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 38, offset: 8140},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 41, offset: 8143},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 48, offset: 8150},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 249, col: 58, offset: 8160},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 62, offset: 8164},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 64, offset: 8166},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 249, col: 76, offset: 8178},
								expr: &ruleRefExpr{
									pos:  position{line: 249, col: 76, offset: 8178},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 93, offset: 8195},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 260, col: 1, offset: 8412},
			expr: &actionExpr{
				pos: position{line: 260, col: 14, offset: 8425},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 260, col: 14, offset: 8425},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 260, col: 21, offset: 8432},
						expr: &seqExpr{
							pos: position{line: 260, col: 22, offset: 8433},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 260, col: 22, offset: 8433},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 260, col: 28, offset: 8439},
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
			pos:  position{line: 269, col: 1, offset: 8620},
			expr: &actionExpr{
				pos: position{line: 269, col: 10, offset: 8629},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 269, col: 10, offset: 8629},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 269, col: 10, offset: 8629},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 17, offset: 8636},
								expr: &seqExpr{
									pos: position{line: 269, col: 18, offset: 8637},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 269, col: 18, offset: 8637},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 28, offset: 8647},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 269, col: 33, offset: 8652},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 36, offset: 8655},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 48, offset: 8667},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 269, col: 50, offset: 8669},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 54, offset: 8673},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 56, offset: 8675},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 60, offset: 8679},
								expr: &ruleRefExpr{
									pos:  position{line: 269, col: 60, offset: 8679},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 75, offset: 8694},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 77, offset: 8696},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 81, offset: 8700},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 91, offset: 8710},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 93, offset: 8712},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 98, offset: 8717},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 109, offset: 8728},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 112, offset: 8731},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 116, offset: 8735},
								expr: &seqExpr{
									pos: position{line: 269, col: 117, offset: 8736},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 269, col: 117, offset: 8736},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 121, offset: 8740},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 269, col: 123, offset: 8742},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 269, col: 136, offset: 8755},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 269, col: 138, offset: 8757},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 269, col: 150, offset: 8769},
								expr: &ruleRefExpr{
									pos:  position{line: 269, col: 150, offset: 8769},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 269, col: 167, offset: 8786},
							expr: &ruleRefExpr{
								pos:  position{line: 269, col: 167, offset: 8786},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 292, col: 1, offset: 9318},
			expr: &actionExpr{
				pos: position{line: 292, col: 18, offset: 9335},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 292, col: 19, offset: 9336},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 292, col: 19, offset: 9336},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 292, col: 32, offset: 9349},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 300, col: 1, offset: 9492},
			expr: &actionExpr{
				pos: position{line: 300, col: 12, offset: 9503},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 300, col: 12, offset: 9503},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 300, col: 12, offset: 9503},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 22, offset: 9513},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 24, offset: 9515},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 300, col: 29, offset: 9520},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 40, offset: 9531},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 42, offset: 9533},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 50, offset: 9541},
								expr: &seqExpr{
									pos: position{line: 300, col: 51, offset: 9542},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 300, col: 51, offset: 9542},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 61, offset: 9552},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 64, offset: 9555},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 75, offset: 9566},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 80, offset: 9571},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 300, col: 83, offset: 9574},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 87, offset: 9578},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 90, offset: 9581},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 300, col: 98, offset: 9589},
								expr: &seqExpr{
									pos: position{line: 300, col: 99, offset: 9590},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 300, col: 99, offset: 9590},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 300, col: 108, offset: 9599},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 300, col: 114, offset: 9605},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 300, col: 114, offset: 9605},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 300, col: 120, offset: 9611},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 139, offset: 9630},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 300, col: 141, offset: 9632},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 300, col: 153, offset: 9644},
								expr: &ruleRefExpr{
									pos:  position{line: 300, col: 153, offset: 9644},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 300, col: 170, offset: 9661},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 317, col: 1, offset: 10102},
			expr: &actionExpr{
				pos: position{line: 317, col: 22, offset: 10123},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 317, col: 22, offset: 10123,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 321, col: 1, offset: 10192},
			expr: &actionExpr{
				pos: position{line: 321, col: 13, offset: 10204},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 321, col: 13, offset: 10204},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 321, col: 13, offset: 10204},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 20, offset: 10211},
								expr: &seqExpr{
									pos: position{line: 321, col: 21, offset: 10212},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 321, col: 21, offset: 10212},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 321, col: 31, offset: 10222},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 321, col: 36, offset: 10227},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 43, offset: 10234},
								expr: &seqExpr{
									pos: position{line: 321, col: 44, offset: 10235},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 321, col: 44, offset: 10235},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 321, col: 53, offset: 10244},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 321, col: 58, offset: 10249},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 62, offset: 10253},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 75, offset: 10266},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 78, offset: 10269},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 83, offset: 10274},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 94, offset: 10285},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 321, col: 96, offset: 10287},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 100, offset: 10291},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 103, offset: 10294},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 113, offset: 10304},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 321, col: 123, offset: 10314},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 127, offset: 10318},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 130, offset: 10321},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 141, offset: 10332},
								expr: &ruleRefExpr{
									pos:  position{line: 321, col: 141, offset: 10332},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 321, col: 149, offset: 10340},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 321, col: 151, offset: 10342},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 321, col: 163, offset: 10354},
								expr: &ruleRefExpr{
									pos:  position{line: 321, col: 163, offset: 10354},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 321, col: 180, offset: 10371},
							expr: &ruleRefExpr{
								pos:  position{line: 321, col: 180, offset: 10371},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 349, col: 1, offset: 11022},
			expr: &actionExpr{
				pos: position{line: 349, col: 17, offset: 11038},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 349, col: 17, offset: 11038},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 349, col: 22, offset: 11043},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 349, col: 22, offset: 11043},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 349, col: 31, offset: 11052},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 356, col: 1, offset: 11174},
			expr: &actionExpr{
				pos: position{line: 356, col: 11, offset: 11184},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 356, col: 11, offset: 11184},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 356, col: 11, offset: 11184},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 20, offset: 11193},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 356, col: 23, offset: 11196},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 356, col: 27, offset: 11200},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 356, col: 30, offset: 11203},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 356, col: 41, offset: 11214},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 356, col: 51, offset: 11224},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 360, col: 1, offset: 11260},
			expr: &actionExpr{
				pos: position{line: 360, col: 14, offset: 11273},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 360, col: 14, offset: 11273},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 360, col: 19, offset: 11278},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 360, col: 19, offset: 11278},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 360, col: 30, offset: 11289},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 360, col: 46, offset: 11305},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 367, col: 1, offset: 11430},
			expr: &actionExpr{
				pos: position{line: 367, col: 13, offset: 11442},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 367, col: 13, offset: 11442},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 367, col: 13, offset: 11442},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 367, col: 18, offset: 11447},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 367, col: 31, offset: 11460},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 367, col: 33, offset: 11462},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 367, col: 45, offset: 11474},
								expr: &ruleRefExpr{
									pos:  position{line: 367, col: 45, offset: 11474},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "BaseTypeName",
			pos:  position{line: 374, col: 1, offset: 11610},
			expr: &actionExpr{
				pos: position{line: 374, col: 17, offset: 11626},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 374, col: 18, offset: 11627},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 374, col: 18, offset: 11627},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 27, offset: 11636},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 36, offset: 11645},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 44, offset: 11653},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 52, offset: 11661},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 60, offset: 11669},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 71, offset: 11680},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 374, col: 82, offset: 11691},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 378, col: 1, offset: 11738},
			expr: &actionExpr{
				pos: position{line: 378, col: 18, offset: 11755},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 378, col: 18, offset: 11755},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 378, col: 23, offset: 11760},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 378, col: 23, offset: 11760},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 378, col: 33, offset: 11770},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 378, col: 43, offset: 11780},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 382, col: 1, offset: 11815},
			expr: &actionExpr{
				pos: position{line: 382, col: 12, offset: 11826},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 382, col: 12, offset: 11826},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 382, col: 12, offset: 11826},
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 12, offset: 11826},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 382, col: 21, offset: 11835},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 28, offset: 11842},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 31, offset: 11845},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 35, offset: 11849},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 45, offset: 11859},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 382, col: 48, offset: 11862},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 52, offset: 11866},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 55, offset: 11869},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 382, col: 61, offset: 11875},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 71, offset: 11885},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 382, col: 74, offset: 11888},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 382, col: 78, offset: 11892},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 382, col: 80, offset: 11894},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 382, col: 92, offset: 11906},
								expr: &ruleRefExpr{
									pos:  position{line: 382, col: 92, offset: 11906},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SetType",
			pos:  position{line: 391, col: 1, offset: 12104},
			expr: &actionExpr{
				pos: position{line: 391, col: 12, offset: 12115},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 391, col: 12, offset: 12115},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 391, col: 12, offset: 12115},
							expr: &ruleRefExpr{
								pos:  position{line: 391, col: 12, offset: 12115},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 391, col: 21, offset: 12124},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 28, offset: 12131},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 391, col: 31, offset: 12134},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 391, col: 35, offset: 12138},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 45, offset: 12148},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 391, col: 48, offset: 12151},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 391, col: 52, offset: 12155},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 391, col: 54, offset: 12157},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 391, col: 66, offset: 12169},
								expr: &ruleRefExpr{
									pos:  position{line: 391, col: 66, offset: 12169},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ListType",
			pos:  position{line: 399, col: 1, offset: 12331},
			expr: &actionExpr{
				pos: position{line: 399, col: 13, offset: 12343},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 399, col: 13, offset: 12343},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 399, col: 13, offset: 12343},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 21, offset: 12351},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 24, offset: 12354},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 399, col: 28, offset: 12358},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 38, offset: 12368},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 399, col: 41, offset: 12371},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 399, col: 45, offset: 12375},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 399, col: 47, offset: 12377},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 399, col: 59, offset: 12389},
								expr: &ruleRefExpr{
									pos:  position{line: 399, col: 59, offset: 12389},
									name: "TypeAnnotations",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "CppType",
			pos:  position{line: 407, col: 1, offset: 12552},
			expr: &actionExpr{
				pos: position{line: 407, col: 12, offset: 12563},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 407, col: 12, offset: 12563},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 407, col: 12, offset: 12563},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 407, col: 23, offset: 12574},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 407, col: 31, offset: 12582},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 411, col: 1, offset: 12619},
			expr: &choiceExpr{
				pos: position{line: 411, col: 15, offset: 12633},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 411, col: 15, offset: 12633},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 25, offset: 12643},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 40, offset: 12658},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 57, offset: 12675},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 71, offset: 12689},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 82, offset: 12700},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 411, col: 94, offset: 12712},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 413, col: 1, offset: 12724},
			expr: &actionExpr{
				pos: position{line: 413, col: 20, offset: 12743},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 413, col: 20, offset: 12743},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 413, col: 20, offset: 12743},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 413, col: 24, offset: 12747},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 413, col: 27, offset: 12750},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 413, col: 39, offset: 12762},
								expr: &ruleRefExpr{
									pos:  position{line: 413, col: 39, offset: 12762},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 413, col: 55, offset: 12778},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 421, col: 1, offset: 12942},
			expr: &actionExpr{
				pos: position{line: 421, col: 19, offset: 12960},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 421, col: 19, offset: 12960},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 421, col: 19, offset: 12960},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 421, col: 24, offset: 12965},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 421, col: 35, offset: 12976},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 421, col: 37, offset: 12978},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 421, col: 43, offset: 12984},
								expr: &actionExpr{
									pos: position{line: 421, col: 44, offset: 12985},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 421, col: 44, offset: 12985},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 421, col: 44, offset: 12985},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 421, col: 48, offset: 12989},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 421, col: 51, offset: 12992},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 421, col: 57, offset: 12998},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 421, col: 89, offset: 13030},
							expr: &ruleRefExpr{
								pos:  position{line: 421, col: 89, offset: 13030},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 421, col: 104, offset: 13045},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 432, col: 1, offset: 13241},
			expr: &actionExpr{
				pos: position{line: 432, col: 17, offset: 13257},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 432, col: 18, offset: 13258},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 432, col: 18, offset: 13258},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 432, col: 27, offset: 13267},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 436, col: 1, offset: 13322},
			expr: &actionExpr{
				pos: position{line: 436, col: 16, offset: 13337},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 436, col: 16, offset: 13337},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 436, col: 16, offset: 13337},
							expr: &charClassMatcher{
								pos:        position{line: 436, col: 16, offset: 13337},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 436, col: 22, offset: 13343},
							expr: &ruleRefExpr{
								pos:  position{line: 436, col: 22, offset: 13343},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 440, col: 1, offset: 13407},
			expr: &actionExpr{
				pos: position{line: 440, col: 19, offset: 13425},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 440, col: 19, offset: 13425},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 440, col: 19, offset: 13425},
							expr: &charClassMatcher{
								pos:        position{line: 440, col: 19, offset: 13425},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 440, col: 25, offset: 13431},
							expr: &ruleRefExpr{
								pos:  position{line: 440, col: 25, offset: 13431},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 440, col: 32, offset: 13438},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 440, col: 36, offset: 13442},
							expr: &ruleRefExpr{
								pos:  position{line: 440, col: 36, offset: 13442},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 440, col: 43, offset: 13449},
							expr: &seqExpr{
								pos: position{line: 440, col: 45, offset: 13451},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 440, col: 45, offset: 13451},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 440, col: 52, offset: 13458},
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
			pos:  position{line: 444, col: 1, offset: 13528},
			expr: &actionExpr{
				pos: position{line: 444, col: 14, offset: 13541},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 444, col: 14, offset: 13541},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 444, col: 14, offset: 13541},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 444, col: 18, offset: 13545},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 444, col: 21, offset: 13548},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 444, col: 28, offset: 13555},
								expr: &seqExpr{
									pos: position{line: 444, col: 29, offset: 13556},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 444, col: 29, offset: 13556},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 444, col: 40, offset: 13567},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 444, col: 43, offset: 13570},
											expr: &ruleRefExpr{
												pos:  position{line: 444, col: 43, offset: 13570},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 444, col: 58, offset: 13585},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 444, col: 63, offset: 13590},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 444, col: 66, offset: 13593},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 453, col: 1, offset: 13787},
			expr: &actionExpr{
				pos: position{line: 453, col: 13, offset: 13799},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 453, col: 13, offset: 13799},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 453, col: 13, offset: 13799},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 453, col: 17, offset: 13803},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 453, col: 20, offset: 13806},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 453, col: 27, offset: 13813},
								expr: &seqExpr{
									pos: position{line: 453, col: 28, offset: 13814},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 453, col: 28, offset: 13814},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 39, offset: 13825},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 453, col: 42, offset: 13828},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 46, offset: 13832},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 49, offset: 13835},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 60, offset: 13846},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 453, col: 64, offset: 13850},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 453, col: 64, offset: 13850},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 453, col: 70, offset: 13856},
													expr: &litMatcher{
														pos:        position{line: 453, col: 71, offset: 13857},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 453, col: 76, offset: 13862},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 453, col: 81, offset: 13867},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 473, col: 1, offset: 14417},
			expr: &ruleRefExpr{
				pos:  position{line: 473, col: 20, offset: 14436},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 475, col: 1, offset: 14443},
			expr: &actionExpr{
				pos: position{line: 475, col: 10, offset: 14452},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 475, col: 10, offset: 14452},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 475, col: 10, offset: 14452},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 475, col: 17, offset: 14459},
								expr: &seqExpr{
									pos: position{line: 475, col: 18, offset: 14460},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 475, col: 18, offset: 14460},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 475, col: 28, offset: 14470},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 475, col: 33, offset: 14475},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 41, offset: 14483},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 475, col: 44, offset: 14486},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 475, col: 49, offset: 14491},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 60, offset: 14502},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 475, col: 63, offset: 14505},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 475, col: 70, offset: 14512},
								expr: &ruleRefExpr{
									pos:  position{line: 475, col: 70, offset: 14512},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 78, offset: 14520},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 475, col: 81, offset: 14523},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 85, offset: 14527},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 475, col: 88, offset: 14530},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 475, col: 99, offset: 14541},
								expr: &seqExpr{
									pos: position{line: 475, col: 100, offset: 14542},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 475, col: 100, offset: 14542},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 475, col: 110, offset: 14552},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 475, col: 116, offset: 14558},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 475, col: 116, offset: 14558},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 475, col: 122, offset: 14564},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 139, offset: 14581},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 475, col: 141, offset: 14583},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 475, col: 153, offset: 14595},
								expr: &ruleRefExpr{
									pos:  position{line: 475, col: 153, offset: 14595},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 475, col: 170, offset: 14612},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 497, col: 1, offset: 15209},
			expr: &actionExpr{
				pos: position{line: 497, col: 20, offset: 15228},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 497, col: 20, offset: 15228,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 501, col: 1, offset: 15295},
			expr: &actionExpr{
				pos: position{line: 501, col: 11, offset: 15305},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 501, col: 11, offset: 15305},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 501, col: 11, offset: 15305},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 501, col: 20, offset: 15314},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 501, col: 23, offset: 15317},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 501, col: 35, offset: 15329},
							expr: &seqExpr{
								pos: position{line: 501, col: 36, offset: 15330},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 501, col: 36, offset: 15330},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 501, col: 40, offset: 15334},
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
			pos:  position{line: 506, col: 1, offset: 15465},
			expr: &choiceExpr{
				pos: position{line: 506, col: 16, offset: 15480},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 506, col: 17, offset: 15481},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 506, col: 17, offset: 15481},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 506, col: 21, offset: 15485},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 506, col: 32, offset: 15496},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 506, col: 39, offset: 15503},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 508, col: 1, offset: 15515},
			expr: &oneOrMoreExpr{
				pos: position{line: 508, col: 15, offset: 15529},
				expr: &charClassMatcher{
					pos:        position{line: 508, col: 15, offset: 15529},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 510, col: 1, offset: 15547},
			expr: &actionExpr{
				pos: position{line: 510, col: 14, offset: 15560},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 510, col: 14, offset: 15560},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 510, col: 14, offset: 15560},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 510, col: 21, offset: 15567},
								expr: &seqExpr{
									pos: position{line: 510, col: 22, offset: 15568},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 510, col: 22, offset: 15568},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 510, col: 32, offset: 15578},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 510, col: 37, offset: 15583},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 510, col: 42, offset: 15588},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 510, col: 53, offset: 15599},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 510, col: 55, offset: 15601},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 510, col: 59, offset: 15605},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 510, col: 62, offset: 15608},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 510, col: 66, offset: 15612},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 510, col: 77, offset: 15623},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 510, col: 79, offset: 15625},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 510, col: 91, offset: 15637},
								expr: &ruleRefExpr{
									pos:  position{line: 510, col: 91, offset: 15637},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 510, col: 108, offset: 15654},
							expr: &ruleRefExpr{
								pos:  position{line: 510, col: 108, offset: 15654},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 527, col: 1, offset: 16240},
			expr: &actionExpr{
				pos: position{line: 527, col: 12, offset: 16251},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 527, col: 13, offset: 16252},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 527, col: 14, offset: 16253},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 527, col: 14, offset: 16253},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 527, col: 18, offset: 16257},
									expr: &choiceExpr{
										pos: position{line: 527, col: 19, offset: 16258},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 527, col: 19, offset: 16258},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 527, col: 26, offset: 16265},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 527, col: 33, offset: 16272},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 527, col: 41, offset: 16280},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 527, col: 41, offset: 16280},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 527, col: 46, offset: 16285},
									expr: &choiceExpr{
										pos: position{line: 527, col: 47, offset: 16286},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 527, col: 47, offset: 16286},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 527, col: 54, offset: 16293},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 527, col: 61, offset: 16300},
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
			pos:  position{line: 536, col: 1, offset: 16586},
			expr: &actionExpr{
				pos: position{line: 536, col: 15, offset: 16600},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 536, col: 15, offset: 16600},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 536, col: 15, offset: 16600},
							expr: &choiceExpr{
								pos: position{line: 536, col: 16, offset: 16601},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 536, col: 16, offset: 16601},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 536, col: 25, offset: 16610},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 536, col: 31, offset: 16616},
							expr: &choiceExpr{
								pos: position{line: 536, col: 32, offset: 16617},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 536, col: 32, offset: 16617},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 536, col: 41, offset: 16626},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 536, col: 49, offset: 16634},
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
			pos:  position{line: 540, col: 1, offset: 16689},
			expr: &charClassMatcher{
				pos:        position{line: 540, col: 18, offset: 16706},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 541, col: 1, offset: 16711},
			expr: &charClassMatcher{
				pos:        position{line: 541, col: 11, offset: 16721},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 542, col: 1, offset: 16730},
			expr: &charClassMatcher{
				pos:        position{line: 542, col: 10, offset: 16739},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 544, col: 1, offset: 16746},
			expr: &anyMatcher{
				line: 544, col: 15, offset: 16760,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 545, col: 1, offset: 16762},
			expr: &actionExpr{
				pos: position{line: 545, col: 14, offset: 16775},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 545, col: 14, offset: 16775},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 545, col: 14, offset: 16775},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 545, col: 21, offset: 16782},
							expr: &seqExpr{
								pos: position{line: 545, col: 23, offset: 16784},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 545, col: 23, offset: 16784},
										expr: &litMatcher{
											pos:        position{line: 545, col: 24, offset: 16785},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 545, col: 29, offset: 16790},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 545, col: 43, offset: 16804},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 551, col: 1, offset: 16984},
			expr: &choiceExpr{
				pos: position{line: 551, col: 12, offset: 16995},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 551, col: 12, offset: 16995},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 551, col: 31, offset: 17014},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 552, col: 1, offset: 17032},
			expr: &seqExpr{
				pos: position{line: 552, col: 21, offset: 17052},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 552, col: 21, offset: 17052},
						expr: &ruleRefExpr{
							pos:  position{line: 552, col: 22, offset: 17053},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 552, col: 32, offset: 17063},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 552, col: 37, offset: 17068},
						expr: &seqExpr{
							pos: position{line: 552, col: 39, offset: 17070},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 552, col: 39, offset: 17070},
									expr: &litMatcher{
										pos:        position{line: 552, col: 40, offset: 17071},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 552, col: 45, offset: 17076},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 552, col: 59, offset: 17090},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 553, col: 1, offset: 17095},
			expr: &seqExpr{
				pos: position{line: 553, col: 37, offset: 17131},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 553, col: 37, offset: 17131},
						expr: &ruleRefExpr{
							pos:  position{line: 553, col: 38, offset: 17132},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 553, col: 48, offset: 17142},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 553, col: 53, offset: 17147},
						expr: &seqExpr{
							pos: position{line: 553, col: 55, offset: 17149},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 553, col: 55, offset: 17149},
									expr: &choiceExpr{
										pos: position{line: 553, col: 58, offset: 17152},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 553, col: 58, offset: 17152},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 553, col: 65, offset: 17159},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 553, col: 71, offset: 17165},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 553, col: 85, offset: 17179},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 554, col: 1, offset: 17184},
			expr: &choiceExpr{
				pos: position{line: 554, col: 22, offset: 17205},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 554, col: 23, offset: 17206},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 554, col: 23, offset: 17206},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 554, col: 28, offset: 17211},
								expr: &seqExpr{
									pos: position{line: 554, col: 30, offset: 17213},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 554, col: 30, offset: 17213},
											expr: &ruleRefExpr{
												pos:  position{line: 554, col: 31, offset: 17214},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 554, col: 35, offset: 17218},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 554, col: 53, offset: 17236},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 554, col: 53, offset: 17236},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 554, col: 57, offset: 17240},
								expr: &seqExpr{
									pos: position{line: 554, col: 59, offset: 17242},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 554, col: 59, offset: 17242},
											expr: &ruleRefExpr{
												pos:  position{line: 554, col: 60, offset: 17243},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 554, col: 64, offset: 17247},
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
			pos:  position{line: 556, col: 1, offset: 17263},
			expr: &zeroOrMoreExpr{
				pos: position{line: 556, col: 7, offset: 17269},
				expr: &choiceExpr{
					pos: position{line: 556, col: 9, offset: 17271},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 556, col: 9, offset: 17271},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 556, col: 22, offset: 17284},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 556, col: 28, offset: 17290},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 557, col: 1, offset: 17301},
			expr: &zeroOrMoreExpr{
				pos: position{line: 557, col: 6, offset: 17306},
				expr: &choiceExpr{
					pos: position{line: 557, col: 8, offset: 17308},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 557, col: 8, offset: 17308},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 557, col: 21, offset: 17321},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 558, col: 1, offset: 17357},
			expr: &zeroOrMoreExpr{
				pos: position{line: 558, col: 7, offset: 17363},
				expr: &ruleRefExpr{
					pos:  position{line: 558, col: 7, offset: 17363},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 560, col: 1, offset: 17376},
			expr: &charClassMatcher{
				pos:        position{line: 560, col: 15, offset: 17390},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 561, col: 1, offset: 17398},
			expr: &litMatcher{
				pos:        position{line: 561, col: 8, offset: 17405},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 562, col: 1, offset: 17410},
			expr: &choiceExpr{
				pos: position{line: 562, col: 8, offset: 17417},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 562, col: 8, offset: 17417},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 562, col: 8, offset: 17417},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 562, col: 11, offset: 17420},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 562, col: 17, offset: 17426},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 562, col: 17, offset: 17426},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 562, col: 19, offset: 17428},
								expr: &ruleRefExpr{
									pos:  position{line: 562, col: 19, offset: 17428},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 562, col: 38, offset: 17447},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 562, col: 44, offset: 17453},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 562, col: 44, offset: 17453},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 562, col: 47, offset: 17456},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 564, col: 1, offset: 17461},
			expr: &notExpr{
				pos: position{line: 564, col: 8, offset: 17468},
				expr: &anyMatcher{
					line: 564, col: 9, offset: 17469,
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

func (c *current) onEnum1(name, values, annotations interface{}) (interface{}, error) {
	vs := toIfaceSlice(values)
	en := &Enum{
		Name:        string(name.(Identifier)),
		Values:      make([]*EnumValue, len(vs)),
		Annotations: toAnnotations(annotations),
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
	return p.cur.onEnum1(stack["name"], stack["values"], stack["annotations"])
}

func (c *current) onEnumValue1(docstr, name, value, annotations interface{}) (interface{}, error) {
	ev := &EnumValue{
		Name:        string(name.(Identifier)),
		Value:       -1,
		Annotations: toAnnotations(annotations),
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
	return p.cur.onEnumValue1(stack["docstr"], stack["name"], stack["value"], stack["annotations"])
}

func (c *current) onTypeDef1(typ, name, annotations interface{}) (interface{}, error) {
	return &TypeDef{
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonTypeDef1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeDef1(stack["typ"], stack["name"], stack["annotations"])
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

func (c *current) onStructLike1(name, fields, annotations interface{}) (interface{}, error) {
	st := &Struct{
		Name:        string(name.(Identifier)),
		Annotations: toAnnotations(annotations),
	}
	if fields != nil {
		st.Fields = fields.([]*Field)
	}
	return st, nil
}

func (p *parser) callonStructLike1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onStructLike1(stack["name"], stack["fields"], stack["annotations"])
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

func (c *current) onField1(docstr, id, mod, typ, name, def, annotations interface{}) (interface{}, error) {
	f := &Field{
		ID:          int(id.(int64)),
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Annotations: toAnnotations(annotations),
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
	return p.cur.onField1(stack["docstr"], stack["id"], stack["mod"], stack["typ"], stack["name"], stack["def"], stack["annotations"])
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

func (c *current) onService1(name, extends, methods, annotations interface{}) (interface{}, error) {
	ms := methods.([]interface{})
	svc := &Service{
		Name:        string(name.(Identifier)),
		Methods:     make([]*Method, len(ms)),
		Annotations: toAnnotations(annotations),
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
	return p.cur.onService1(stack["name"], stack["extends"], stack["methods"], stack["annotations"])
}

func (c *current) onEndOfServiceError1() (interface{}, error) {
	return nil, errors.New("parser: expected end of service")
}

func (p *parser) callonEndOfServiceError1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onEndOfServiceError1()
}

func (c *current) onFunction1(docstr, oneway, typ, name, arguments, exceptions, annotations interface{}) (interface{}, error) {
	m := &Method{
		Name:        string(name.(Identifier)),
		Annotations: toAnnotations(annotations),
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
	return p.cur.onFunction1(stack["docstr"], stack["oneway"], stack["typ"], stack["name"], stack["arguments"], stack["exceptions"], stack["annotations"])
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

func (c *current) onBaseType1(name, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        name.(string),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonBaseType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBaseType1(stack["name"], stack["annotations"])
}

func (c *current) onBaseTypeName1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonBaseTypeName1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onBaseTypeName1()
}

func (c *current) onContainerType1(typ interface{}) (interface{}, error) {
	return typ, nil
}

func (p *parser) callonContainerType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onContainerType1(stack["typ"])
}

func (c *current) onMapType1(key, value, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "map",
		KeyType:     key.(*Type),
		ValueType:   value.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonMapType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onMapType1(stack["key"], stack["value"], stack["annotations"])
}

func (c *current) onSetType1(typ, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "set",
		ValueType:   typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonSetType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onSetType1(stack["typ"], stack["annotations"])
}

func (c *current) onListType1(typ, annotations interface{}) (interface{}, error) {
	return &Type{
		Name:        "list",
		ValueType:   typ.(*Type),
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonListType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onListType1(stack["typ"], stack["annotations"])
}

func (c *current) onCppType1(cppType interface{}) (interface{}, error) {
	return cppType, nil
}

func (p *parser) callonCppType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onCppType1(stack["cppType"])
}

func (c *current) onTypeAnnotations1(annotations interface{}) (interface{}, error) {
	var anns []*Annotation
	for _, ann := range annotations.([]interface{}) {
		anns = append(anns, ann.(*Annotation))
	}
	return anns, nil
}

func (p *parser) callonTypeAnnotations1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotations1(stack["annotations"])
}

func (c *current) onTypeAnnotation8(value interface{}) (interface{}, error) {
	return value, nil
}

func (p *parser) callonTypeAnnotation8() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotation8(stack["value"])
}

func (c *current) onTypeAnnotation1(name, value interface{}) (interface{}, error) {
	var optValue string
	if value != nil {
		optValue = value.(string)
	}
	return &Annotation{
		Name:  string(name.(Identifier)),
		Value: optValue,
	}, nil
}

func (p *parser) callonTypeAnnotation1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onTypeAnnotation1(stack["name"], stack["value"])
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

func (c *current) onScope1(docstr, name, prefix, operations, annotations interface{}) (interface{}, error) {
	ops := operations.([]interface{})
	scope := &Scope{
		Name:        string(name.(Identifier)),
		Operations:  make([]*Operation, len(ops)),
		Prefix:      defaultPrefix,
		Annotations: toAnnotations(annotations),
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
	return p.cur.onScope1(stack["docstr"], stack["name"], stack["prefix"], stack["operations"], stack["annotations"])
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

func (c *current) onOperation1(docstr, name, typ, annotations interface{}) (interface{}, error) {
	o := &Operation{
		Name:        string(name.(Identifier)),
		Type:        &Type{Name: string(typ.(Identifier))},
		Annotations: toAnnotations(annotations),
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
	return p.cur.onOperation1(stack["docstr"], stack["name"], stack["typ"], stack["annotations"])
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
