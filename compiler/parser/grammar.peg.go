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
func toAnnotations(v interface{}) Annotations {
	if v == nil {
		return nil
	}
	return Annotations(v.([]*Annotation))
}

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 85, col: 1, offset: 2393},
			expr: &actionExpr{
				pos: position{line: 85, col: 12, offset: 2404},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 85, col: 12, offset: 2404},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 85, col: 12, offset: 2404},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 15, offset: 2407},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 85, col: 26, offset: 2418},
								expr: &seqExpr{
									pos: position{line: 85, col: 28, offset: 2420},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 85, col: 28, offset: 2420},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 85, col: 38, offset: 2430},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 85, col: 45, offset: 2437},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 85, col: 45, offset: 2437},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 85, col: 51, offset: 2443},
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
			pos:  position{line: 156, col: 1, offset: 5051},
			expr: &actionExpr{
				pos: position{line: 156, col: 16, offset: 5066},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 156, col: 16, offset: 5066,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 160, col: 1, offset: 5124},
			expr: &actionExpr{
				pos: position{line: 160, col: 14, offset: 5137},
				run: (*parser).callonStatement1,
				expr: &seqExpr{
					pos: position{line: 160, col: 14, offset: 5137},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 160, col: 14, offset: 5137},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 160, col: 21, offset: 5144},
								expr: &seqExpr{
									pos: position{line: 160, col: 22, offset: 5145},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 160, col: 22, offset: 5145},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 160, col: 32, offset: 5155},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 160, col: 37, offset: 5160},
							label: "statement",
							expr: &choiceExpr{
								pos: position{line: 160, col: 48, offset: 5171},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 160, col: 48, offset: 5171},
										name: "ThriftStatement",
									},
									&ruleRefExpr{
										pos:  position{line: 160, col: 66, offset: 5189},
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
			pos:  position{line: 173, col: 1, offset: 5660},
			expr: &choiceExpr{
				pos: position{line: 173, col: 20, offset: 5679},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 173, col: 20, offset: 5679},
						name: "Include",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 30, offset: 5689},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 42, offset: 5701},
						name: "Const",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 50, offset: 5709},
						name: "Enum",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 57, offset: 5716},
						name: "TypeDef",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 67, offset: 5726},
						name: "Struct",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 76, offset: 5735},
						name: "Exception",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 88, offset: 5747},
						name: "Union",
					},
					&ruleRefExpr{
						pos:  position{line: 173, col: 96, offset: 5755},
						name: "Service",
					},
				},
			},
		},
		{
			name: "Include",
			pos:  position{line: 175, col: 1, offset: 5764},
			expr: &actionExpr{
				pos: position{line: 175, col: 12, offset: 5775},
				run: (*parser).callonInclude1,
				expr: &seqExpr{
					pos: position{line: 175, col: 12, offset: 5775},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 175, col: 12, offset: 5775},
							val:        "include",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 22, offset: 5785},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 175, col: 24, offset: 5787},
							label: "file",
							expr: &ruleRefExpr{
								pos:  position{line: 175, col: 29, offset: 5792},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 175, col: 37, offset: 5800},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 183, col: 1, offset: 5977},
			expr: &actionExpr{
				pos: position{line: 183, col: 14, offset: 5990},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 183, col: 14, offset: 5990},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 183, col: 14, offset: 5990},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 26, offset: 6002},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 183, col: 28, offset: 6004},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 183, col: 34, offset: 6010},
								expr: &charClassMatcher{
									pos:        position{line: 183, col: 34, offset: 6010},
									val:        "[*a-z.-]",
									chars:      []rune{'*', '.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 44, offset: 6020},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 183, col: 46, offset: 6022},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 183, col: 49, offset: 6025},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 183, col: 60, offset: 6036},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Const",
			pos:  position{line: 190, col: 1, offset: 6161},
			expr: &actionExpr{
				pos: position{line: 190, col: 10, offset: 6170},
				run: (*parser).callonConst1,
				expr: &seqExpr{
					pos: position{line: 190, col: 10, offset: 6170},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 190, col: 10, offset: 6170},
							val:        "const",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 18, offset: 6178},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 20, offset: 6180},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 24, offset: 6184},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 34, offset: 6194},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 36, offset: 6196},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 41, offset: 6201},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 52, offset: 6212},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 190, col: 54, offset: 6214},
							val:        "=",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 58, offset: 6218},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 60, offset: 6220},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 190, col: 66, offset: 6226},
								name: "ConstValue",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 77, offset: 6237},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 190, col: 79, offset: 6239},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 190, col: 91, offset: 6251},
								expr: &ruleRefExpr{
									pos:  position{line: 190, col: 91, offset: 6251},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 190, col: 108, offset: 6268},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Enum",
			pos:  position{line: 199, col: 1, offset: 6462},
			expr: &actionExpr{
				pos: position{line: 199, col: 9, offset: 6470},
				run: (*parser).callonEnum1,
				expr: &seqExpr{
					pos: position{line: 199, col: 9, offset: 6470},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 199, col: 9, offset: 6470},
							val:        "enum",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 16, offset: 6477},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 18, offset: 6479},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 199, col: 23, offset: 6484},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 34, offset: 6495},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 199, col: 37, offset: 6498},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 41, offset: 6502},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 44, offset: 6505},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 199, col: 51, offset: 6512},
								expr: &seqExpr{
									pos: position{line: 199, col: 52, offset: 6513},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 199, col: 52, offset: 6513},
											name: "EnumValue",
										},
										&ruleRefExpr{
											pos:  position{line: 199, col: 62, offset: 6523},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 199, col: 67, offset: 6528},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 71, offset: 6532},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 199, col: 73, offset: 6534},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 199, col: 85, offset: 6546},
								expr: &ruleRefExpr{
									pos:  position{line: 199, col: 85, offset: 6546},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 199, col: 102, offset: 6563},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EnumValue",
			pos:  position{line: 223, col: 1, offset: 7225},
			expr: &actionExpr{
				pos: position{line: 223, col: 14, offset: 7238},
				run: (*parser).callonEnumValue1,
				expr: &seqExpr{
					pos: position{line: 223, col: 14, offset: 7238},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 223, col: 14, offset: 7238},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 223, col: 21, offset: 7245},
								expr: &seqExpr{
									pos: position{line: 223, col: 22, offset: 7246},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 223, col: 22, offset: 7246},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 223, col: 32, offset: 7256},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 223, col: 37, offset: 7261},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 223, col: 42, offset: 7266},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 223, col: 53, offset: 7277},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 223, col: 55, offset: 7279},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 223, col: 61, offset: 7285},
								expr: &seqExpr{
									pos: position{line: 223, col: 62, offset: 7286},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 223, col: 62, offset: 7286},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 223, col: 66, offset: 7290},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 223, col: 68, offset: 7292},
											name: "IntConstant",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 223, col: 82, offset: 7306},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 223, col: 84, offset: 7308},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 223, col: 96, offset: 7320},
								expr: &ruleRefExpr{
									pos:  position{line: 223, col: 96, offset: 7320},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 223, col: 113, offset: 7337},
							expr: &ruleRefExpr{
								pos:  position{line: 223, col: 113, offset: 7337},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "TypeDef",
			pos:  position{line: 239, col: 1, offset: 7735},
			expr: &actionExpr{
				pos: position{line: 239, col: 12, offset: 7746},
				run: (*parser).callonTypeDef1,
				expr: &seqExpr{
					pos: position{line: 239, col: 12, offset: 7746},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 239, col: 12, offset: 7746},
							val:        "typedef",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 22, offset: 7756},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 239, col: 24, offset: 7758},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 239, col: 28, offset: 7762},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 38, offset: 7772},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 239, col: 40, offset: 7774},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 239, col: 45, offset: 7779},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 56, offset: 7790},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 239, col: 58, offset: 7792},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 239, col: 70, offset: 7804},
								expr: &ruleRefExpr{
									pos:  position{line: 239, col: 70, offset: 7804},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 239, col: 87, offset: 7821},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Struct",
			pos:  position{line: 247, col: 1, offset: 7993},
			expr: &actionExpr{
				pos: position{line: 247, col: 11, offset: 8003},
				run: (*parser).callonStruct1,
				expr: &seqExpr{
					pos: position{line: 247, col: 11, offset: 8003},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 247, col: 11, offset: 8003},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 247, col: 20, offset: 8012},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 247, col: 22, offset: 8014},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 247, col: 25, offset: 8017},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Exception",
			pos:  position{line: 248, col: 1, offset: 8057},
			expr: &actionExpr{
				pos: position{line: 248, col: 14, offset: 8070},
				run: (*parser).callonException1,
				expr: &seqExpr{
					pos: position{line: 248, col: 14, offset: 8070},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 248, col: 14, offset: 8070},
							val:        "exception",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 248, col: 26, offset: 8082},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 248, col: 28, offset: 8084},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 248, col: 31, offset: 8087},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "Union",
			pos:  position{line: 249, col: 1, offset: 8138},
			expr: &actionExpr{
				pos: position{line: 249, col: 10, offset: 8147},
				run: (*parser).callonUnion1,
				expr: &seqExpr{
					pos: position{line: 249, col: 10, offset: 8147},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 249, col: 10, offset: 8147},
							val:        "union",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 249, col: 18, offset: 8155},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 249, col: 20, offset: 8157},
							label: "st",
							expr: &ruleRefExpr{
								pos:  position{line: 249, col: 23, offset: 8160},
								name: "StructLike",
							},
						},
					},
				},
			},
		},
		{
			name: "StructLike",
			pos:  position{line: 250, col: 1, offset: 8207},
			expr: &actionExpr{
				pos: position{line: 250, col: 15, offset: 8221},
				run: (*parser).callonStructLike1,
				expr: &seqExpr{
					pos: position{line: 250, col: 15, offset: 8221},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 250, col: 15, offset: 8221},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 20, offset: 8226},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 31, offset: 8237},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 250, col: 34, offset: 8240},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 38, offset: 8244},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 41, offset: 8247},
							label: "fields",
							expr: &ruleRefExpr{
								pos:  position{line: 250, col: 48, offset: 8254},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 250, col: 58, offset: 8264},
							val:        "}",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 62, offset: 8268},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 250, col: 64, offset: 8270},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 250, col: 76, offset: 8282},
								expr: &ruleRefExpr{
									pos:  position{line: 250, col: 76, offset: 8282},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 250, col: 93, offset: 8299},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "FieldList",
			pos:  position{line: 261, col: 1, offset: 8516},
			expr: &actionExpr{
				pos: position{line: 261, col: 14, offset: 8529},
				run: (*parser).callonFieldList1,
				expr: &labeledExpr{
					pos:   position{line: 261, col: 14, offset: 8529},
					label: "fields",
					expr: &zeroOrMoreExpr{
						pos: position{line: 261, col: 21, offset: 8536},
						expr: &seqExpr{
							pos: position{line: 261, col: 22, offset: 8537},
							exprs: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 261, col: 22, offset: 8537},
									name: "Field",
								},
								&ruleRefExpr{
									pos:  position{line: 261, col: 28, offset: 8543},
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
			pos:  position{line: 270, col: 1, offset: 8724},
			expr: &actionExpr{
				pos: position{line: 270, col: 10, offset: 8733},
				run: (*parser).callonField1,
				expr: &seqExpr{
					pos: position{line: 270, col: 10, offset: 8733},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 270, col: 10, offset: 8733},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 270, col: 17, offset: 8740},
								expr: &seqExpr{
									pos: position{line: 270, col: 18, offset: 8741},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 270, col: 18, offset: 8741},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 270, col: 28, offset: 8751},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 270, col: 33, offset: 8756},
							label: "id",
							expr: &ruleRefExpr{
								pos:  position{line: 270, col: 36, offset: 8759},
								name: "IntConstant",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 270, col: 48, offset: 8771},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 270, col: 50, offset: 8773},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 270, col: 54, offset: 8777},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 270, col: 56, offset: 8779},
							label: "mod",
							expr: &zeroOrOneExpr{
								pos: position{line: 270, col: 60, offset: 8783},
								expr: &ruleRefExpr{
									pos:  position{line: 270, col: 60, offset: 8783},
									name: "FieldModifier",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 270, col: 75, offset: 8798},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 270, col: 77, offset: 8800},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 270, col: 81, offset: 8804},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 270, col: 91, offset: 8814},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 270, col: 93, offset: 8816},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 270, col: 98, offset: 8821},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 270, col: 109, offset: 8832},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 270, col: 112, offset: 8835},
							label: "def",
							expr: &zeroOrOneExpr{
								pos: position{line: 270, col: 116, offset: 8839},
								expr: &seqExpr{
									pos: position{line: 270, col: 117, offset: 8840},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 270, col: 117, offset: 8840},
											val:        "=",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 270, col: 121, offset: 8844},
											name: "_",
										},
										&ruleRefExpr{
											pos:  position{line: 270, col: 123, offset: 8846},
											name: "ConstValue",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 270, col: 136, offset: 8859},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 270, col: 138, offset: 8861},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 270, col: 150, offset: 8873},
								expr: &ruleRefExpr{
									pos:  position{line: 270, col: 150, offset: 8873},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 270, col: 167, offset: 8890},
							expr: &ruleRefExpr{
								pos:  position{line: 270, col: 167, offset: 8890},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FieldModifier",
			pos:  position{line: 293, col: 1, offset: 9422},
			expr: &actionExpr{
				pos: position{line: 293, col: 18, offset: 9439},
				run: (*parser).callonFieldModifier1,
				expr: &choiceExpr{
					pos: position{line: 293, col: 19, offset: 9440},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 293, col: 19, offset: 9440},
							val:        "required",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 293, col: 32, offset: 9453},
							val:        "optional",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Service",
			pos:  position{line: 301, col: 1, offset: 9596},
			expr: &actionExpr{
				pos: position{line: 301, col: 12, offset: 9607},
				run: (*parser).callonService1,
				expr: &seqExpr{
					pos: position{line: 301, col: 12, offset: 9607},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 301, col: 12, offset: 9607},
							val:        "service",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 301, col: 22, offset: 9617},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 301, col: 24, offset: 9619},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 301, col: 29, offset: 9624},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 301, col: 40, offset: 9635},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 301, col: 42, offset: 9637},
							label: "extends",
							expr: &zeroOrOneExpr{
								pos: position{line: 301, col: 50, offset: 9645},
								expr: &seqExpr{
									pos: position{line: 301, col: 51, offset: 9646},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 301, col: 51, offset: 9646},
											val:        "extends",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 301, col: 61, offset: 9656},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 301, col: 64, offset: 9659},
											name: "Identifier",
										},
										&ruleRefExpr{
											pos:  position{line: 301, col: 75, offset: 9670},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 301, col: 80, offset: 9675},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 301, col: 83, offset: 9678},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 301, col: 87, offset: 9682},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 301, col: 90, offset: 9685},
							label: "methods",
							expr: &zeroOrMoreExpr{
								pos: position{line: 301, col: 98, offset: 9693},
								expr: &seqExpr{
									pos: position{line: 301, col: 99, offset: 9694},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 301, col: 99, offset: 9694},
											name: "Function",
										},
										&ruleRefExpr{
											pos:  position{line: 301, col: 108, offset: 9703},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 301, col: 114, offset: 9709},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 301, col: 114, offset: 9709},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 301, col: 120, offset: 9715},
									name: "EndOfServiceError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 301, col: 139, offset: 9734},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 301, col: 141, offset: 9736},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 301, col: 153, offset: 9748},
								expr: &ruleRefExpr{
									pos:  position{line: 301, col: 153, offset: 9748},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 301, col: 170, offset: 9765},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfServiceError",
			pos:  position{line: 318, col: 1, offset: 10206},
			expr: &actionExpr{
				pos: position{line: 318, col: 22, offset: 10227},
				run: (*parser).callonEndOfServiceError1,
				expr: &anyMatcher{
					line: 318, col: 22, offset: 10227,
				},
			},
		},
		{
			name: "Function",
			pos:  position{line: 322, col: 1, offset: 10296},
			expr: &actionExpr{
				pos: position{line: 322, col: 13, offset: 10308},
				run: (*parser).callonFunction1,
				expr: &seqExpr{
					pos: position{line: 322, col: 13, offset: 10308},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 322, col: 13, offset: 10308},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 322, col: 20, offset: 10315},
								expr: &seqExpr{
									pos: position{line: 322, col: 21, offset: 10316},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 322, col: 21, offset: 10316},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 322, col: 31, offset: 10326},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 322, col: 36, offset: 10331},
							label: "oneway",
							expr: &zeroOrOneExpr{
								pos: position{line: 322, col: 43, offset: 10338},
								expr: &seqExpr{
									pos: position{line: 322, col: 44, offset: 10339},
									exprs: []interface{}{
										&litMatcher{
											pos:        position{line: 322, col: 44, offset: 10339},
											val:        "oneway",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 322, col: 53, offset: 10348},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 322, col: 58, offset: 10353},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 322, col: 62, offset: 10357},
								name: "FunctionType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 322, col: 75, offset: 10370},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 322, col: 78, offset: 10373},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 322, col: 83, offset: 10378},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 322, col: 94, offset: 10389},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 322, col: 96, offset: 10391},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 322, col: 100, offset: 10395},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 322, col: 103, offset: 10398},
							label: "arguments",
							expr: &ruleRefExpr{
								pos:  position{line: 322, col: 113, offset: 10408},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 322, col: 123, offset: 10418},
							val:        ")",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 322, col: 127, offset: 10422},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 322, col: 130, offset: 10425},
							label: "exceptions",
							expr: &zeroOrOneExpr{
								pos: position{line: 322, col: 141, offset: 10436},
								expr: &ruleRefExpr{
									pos:  position{line: 322, col: 141, offset: 10436},
									name: "Throws",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 322, col: 149, offset: 10444},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 322, col: 151, offset: 10446},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 322, col: 163, offset: 10458},
								expr: &ruleRefExpr{
									pos:  position{line: 322, col: 163, offset: 10458},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 322, col: 180, offset: 10475},
							expr: &ruleRefExpr{
								pos:  position{line: 322, col: 180, offset: 10475},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "FunctionType",
			pos:  position{line: 350, col: 1, offset: 11126},
			expr: &actionExpr{
				pos: position{line: 350, col: 17, offset: 11142},
				run: (*parser).callonFunctionType1,
				expr: &labeledExpr{
					pos:   position{line: 350, col: 17, offset: 11142},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 350, col: 22, offset: 11147},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 350, col: 22, offset: 11147},
								val:        "void",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 350, col: 31, offset: 11156},
								name: "FieldType",
							},
						},
					},
				},
			},
		},
		{
			name: "Throws",
			pos:  position{line: 357, col: 1, offset: 11278},
			expr: &actionExpr{
				pos: position{line: 357, col: 11, offset: 11288},
				run: (*parser).callonThrows1,
				expr: &seqExpr{
					pos: position{line: 357, col: 11, offset: 11288},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 357, col: 11, offset: 11288},
							val:        "throws",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 357, col: 20, offset: 11297},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 357, col: 23, offset: 11300},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 357, col: 27, offset: 11304},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 357, col: 30, offset: 11307},
							label: "exceptions",
							expr: &ruleRefExpr{
								pos:  position{line: 357, col: 41, offset: 11318},
								name: "FieldList",
							},
						},
						&litMatcher{
							pos:        position{line: 357, col: 51, offset: 11328},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FieldType",
			pos:  position{line: 361, col: 1, offset: 11364},
			expr: &actionExpr{
				pos: position{line: 361, col: 14, offset: 11377},
				run: (*parser).callonFieldType1,
				expr: &labeledExpr{
					pos:   position{line: 361, col: 14, offset: 11377},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 361, col: 19, offset: 11382},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 361, col: 19, offset: 11382},
								name: "BaseType",
							},
							&ruleRefExpr{
								pos:  position{line: 361, col: 30, offset: 11393},
								name: "ContainerType",
							},
							&ruleRefExpr{
								pos:  position{line: 361, col: 46, offset: 11409},
								name: "Identifier",
							},
						},
					},
				},
			},
		},
		{
			name: "BaseType",
			pos:  position{line: 368, col: 1, offset: 11534},
			expr: &actionExpr{
				pos: position{line: 368, col: 13, offset: 11546},
				run: (*parser).callonBaseType1,
				expr: &seqExpr{
					pos: position{line: 368, col: 13, offset: 11546},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 368, col: 13, offset: 11546},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 368, col: 18, offset: 11551},
								name: "BaseTypeName",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 368, col: 31, offset: 11564},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 368, col: 33, offset: 11566},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 368, col: 45, offset: 11578},
								expr: &ruleRefExpr{
									pos:  position{line: 368, col: 45, offset: 11578},
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
			pos:  position{line: 375, col: 1, offset: 11714},
			expr: &actionExpr{
				pos: position{line: 375, col: 17, offset: 11730},
				run: (*parser).callonBaseTypeName1,
				expr: &choiceExpr{
					pos: position{line: 375, col: 18, offset: 11731},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 375, col: 18, offset: 11731},
							val:        "bool",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 375, col: 27, offset: 11740},
							val:        "byte",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 375, col: 36, offset: 11749},
							val:        "i16",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 375, col: 44, offset: 11757},
							val:        "i32",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 375, col: 52, offset: 11765},
							val:        "i64",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 375, col: 60, offset: 11773},
							val:        "double",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 375, col: 71, offset: 11784},
							val:        "string",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 375, col: 82, offset: 11795},
							val:        "binary",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ContainerType",
			pos:  position{line: 379, col: 1, offset: 11842},
			expr: &actionExpr{
				pos: position{line: 379, col: 18, offset: 11859},
				run: (*parser).callonContainerType1,
				expr: &labeledExpr{
					pos:   position{line: 379, col: 18, offset: 11859},
					label: "typ",
					expr: &choiceExpr{
						pos: position{line: 379, col: 23, offset: 11864},
						alternatives: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 379, col: 23, offset: 11864},
								name: "MapType",
							},
							&ruleRefExpr{
								pos:  position{line: 379, col: 33, offset: 11874},
								name: "SetType",
							},
							&ruleRefExpr{
								pos:  position{line: 379, col: 43, offset: 11884},
								name: "ListType",
							},
						},
					},
				},
			},
		},
		{
			name: "MapType",
			pos:  position{line: 383, col: 1, offset: 11919},
			expr: &actionExpr{
				pos: position{line: 383, col: 12, offset: 11930},
				run: (*parser).callonMapType1,
				expr: &seqExpr{
					pos: position{line: 383, col: 12, offset: 11930},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 383, col: 12, offset: 11930},
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 12, offset: 11930},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 383, col: 21, offset: 11939},
							val:        "map<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 28, offset: 11946},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 383, col: 31, offset: 11949},
							label: "key",
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 35, offset: 11953},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 45, offset: 11963},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 383, col: 48, offset: 11966},
							val:        ",",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 52, offset: 11970},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 383, col: 55, offset: 11973},
							label: "value",
							expr: &ruleRefExpr{
								pos:  position{line: 383, col: 61, offset: 11979},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 71, offset: 11989},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 383, col: 74, offset: 11992},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 383, col: 78, offset: 11996},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 383, col: 80, offset: 11998},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 383, col: 92, offset: 12010},
								expr: &ruleRefExpr{
									pos:  position{line: 383, col: 92, offset: 12010},
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
			pos:  position{line: 392, col: 1, offset: 12208},
			expr: &actionExpr{
				pos: position{line: 392, col: 12, offset: 12219},
				run: (*parser).callonSetType1,
				expr: &seqExpr{
					pos: position{line: 392, col: 12, offset: 12219},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 392, col: 12, offset: 12219},
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 12, offset: 12219},
								name: "CppType",
							},
						},
						&litMatcher{
							pos:        position{line: 392, col: 21, offset: 12228},
							val:        "set<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 392, col: 28, offset: 12235},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 392, col: 31, offset: 12238},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 392, col: 35, offset: 12242},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 392, col: 45, offset: 12252},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 392, col: 48, offset: 12255},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 392, col: 52, offset: 12259},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 392, col: 54, offset: 12261},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 392, col: 66, offset: 12273},
								expr: &ruleRefExpr{
									pos:  position{line: 392, col: 66, offset: 12273},
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
			pos:  position{line: 400, col: 1, offset: 12435},
			expr: &actionExpr{
				pos: position{line: 400, col: 13, offset: 12447},
				run: (*parser).callonListType1,
				expr: &seqExpr{
					pos: position{line: 400, col: 13, offset: 12447},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 400, col: 13, offset: 12447},
							val:        "list<",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 21, offset: 12455},
							name: "WS",
						},
						&labeledExpr{
							pos:   position{line: 400, col: 24, offset: 12458},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 400, col: 28, offset: 12462},
								name: "FieldType",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 38, offset: 12472},
							name: "WS",
						},
						&litMatcher{
							pos:        position{line: 400, col: 41, offset: 12475},
							val:        ">",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 400, col: 45, offset: 12479},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 400, col: 47, offset: 12481},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 400, col: 59, offset: 12493},
								expr: &ruleRefExpr{
									pos:  position{line: 400, col: 59, offset: 12493},
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
			pos:  position{line: 408, col: 1, offset: 12656},
			expr: &actionExpr{
				pos: position{line: 408, col: 12, offset: 12667},
				run: (*parser).callonCppType1,
				expr: &seqExpr{
					pos: position{line: 408, col: 12, offset: 12667},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 408, col: 12, offset: 12667},
							val:        "cpp_type",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 408, col: 23, offset: 12678},
							label: "cppType",
							expr: &ruleRefExpr{
								pos:  position{line: 408, col: 31, offset: 12686},
								name: "Literal",
							},
						},
					},
				},
			},
		},
		{
			name: "ConstValue",
			pos:  position{line: 412, col: 1, offset: 12723},
			expr: &choiceExpr{
				pos: position{line: 412, col: 15, offset: 12737},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 412, col: 15, offset: 12737},
						name: "Literal",
					},
					&ruleRefExpr{
						pos:  position{line: 412, col: 25, offset: 12747},
						name: "BoolConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 412, col: 40, offset: 12762},
						name: "DoubleConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 412, col: 57, offset: 12779},
						name: "IntConstant",
					},
					&ruleRefExpr{
						pos:  position{line: 412, col: 71, offset: 12793},
						name: "ConstMap",
					},
					&ruleRefExpr{
						pos:  position{line: 412, col: 82, offset: 12804},
						name: "ConstList",
					},
					&ruleRefExpr{
						pos:  position{line: 412, col: 94, offset: 12816},
						name: "Identifier",
					},
				},
			},
		},
		{
			name: "TypeAnnotations",
			pos:  position{line: 414, col: 1, offset: 12828},
			expr: &actionExpr{
				pos: position{line: 414, col: 20, offset: 12847},
				run: (*parser).callonTypeAnnotations1,
				expr: &seqExpr{
					pos: position{line: 414, col: 20, offset: 12847},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 414, col: 20, offset: 12847},
							val:        "(",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 414, col: 24, offset: 12851},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 414, col: 27, offset: 12854},
							label: "annotations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 414, col: 39, offset: 12866},
								expr: &ruleRefExpr{
									pos:  position{line: 414, col: 39, offset: 12866},
									name: "TypeAnnotation",
								},
							},
						},
						&litMatcher{
							pos:        position{line: 414, col: 55, offset: 12882},
							val:        ")",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "TypeAnnotation",
			pos:  position{line: 422, col: 1, offset: 13046},
			expr: &actionExpr{
				pos: position{line: 422, col: 19, offset: 13064},
				run: (*parser).callonTypeAnnotation1,
				expr: &seqExpr{
					pos: position{line: 422, col: 19, offset: 13064},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 422, col: 19, offset: 13064},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 422, col: 24, offset: 13069},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 35, offset: 13080},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 422, col: 37, offset: 13082},
							label: "value",
							expr: &zeroOrOneExpr{
								pos: position{line: 422, col: 43, offset: 13088},
								expr: &actionExpr{
									pos: position{line: 422, col: 44, offset: 13089},
									run: (*parser).callonTypeAnnotation8,
									expr: &seqExpr{
										pos: position{line: 422, col: 44, offset: 13089},
										exprs: []interface{}{
											&litMatcher{
												pos:        position{line: 422, col: 44, offset: 13089},
												val:        "=",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 422, col: 48, offset: 13093},
												name: "__",
											},
											&labeledExpr{
												pos:   position{line: 422, col: 51, offset: 13096},
												label: "value",
												expr: &ruleRefExpr{
													pos:  position{line: 422, col: 57, offset: 13102},
													name: "Literal",
												},
											},
										},
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 422, col: 89, offset: 13134},
							expr: &ruleRefExpr{
								pos:  position{line: 422, col: 89, offset: 13134},
								name: "ListSeparator",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 422, col: 104, offset: 13149},
							name: "__",
						},
					},
				},
			},
		},
		{
			name: "BoolConstant",
			pos:  position{line: 433, col: 1, offset: 13345},
			expr: &actionExpr{
				pos: position{line: 433, col: 17, offset: 13361},
				run: (*parser).callonBoolConstant1,
				expr: &choiceExpr{
					pos: position{line: 433, col: 18, offset: 13362},
					alternatives: []interface{}{
						&litMatcher{
							pos:        position{line: 433, col: 18, offset: 13362},
							val:        "true",
							ignoreCase: false,
						},
						&litMatcher{
							pos:        position{line: 433, col: 27, offset: 13371},
							val:        "false",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "IntConstant",
			pos:  position{line: 437, col: 1, offset: 13426},
			expr: &actionExpr{
				pos: position{line: 437, col: 16, offset: 13441},
				run: (*parser).callonIntConstant1,
				expr: &seqExpr{
					pos: position{line: 437, col: 16, offset: 13441},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 437, col: 16, offset: 13441},
							expr: &charClassMatcher{
								pos:        position{line: 437, col: 16, offset: 13441},
								val:        "[-+]",
								chars:      []rune{'-', '+'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 437, col: 22, offset: 13447},
							expr: &ruleRefExpr{
								pos:  position{line: 437, col: 22, offset: 13447},
								name: "Digit",
							},
						},
					},
				},
			},
		},
		{
			name: "DoubleConstant",
			pos:  position{line: 441, col: 1, offset: 13511},
			expr: &actionExpr{
				pos: position{line: 441, col: 19, offset: 13529},
				run: (*parser).callonDoubleConstant1,
				expr: &seqExpr{
					pos: position{line: 441, col: 19, offset: 13529},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 441, col: 19, offset: 13529},
							expr: &charClassMatcher{
								pos:        position{line: 441, col: 19, offset: 13529},
								val:        "[+-]",
								chars:      []rune{'+', '-'},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 441, col: 25, offset: 13535},
							expr: &ruleRefExpr{
								pos:  position{line: 441, col: 25, offset: 13535},
								name: "Digit",
							},
						},
						&litMatcher{
							pos:        position{line: 441, col: 32, offset: 13542},
							val:        ".",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 441, col: 36, offset: 13546},
							expr: &ruleRefExpr{
								pos:  position{line: 441, col: 36, offset: 13546},
								name: "Digit",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 441, col: 43, offset: 13553},
							expr: &seqExpr{
								pos: position{line: 441, col: 45, offset: 13555},
								exprs: []interface{}{
									&charClassMatcher{
										pos:        position{line: 441, col: 45, offset: 13555},
										val:        "['Ee']",
										chars:      []rune{'\'', 'E', 'e', '\''},
										ignoreCase: false,
										inverted:   false,
									},
									&ruleRefExpr{
										pos:  position{line: 441, col: 52, offset: 13562},
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
			pos:  position{line: 445, col: 1, offset: 13632},
			expr: &actionExpr{
				pos: position{line: 445, col: 14, offset: 13645},
				run: (*parser).callonConstList1,
				expr: &seqExpr{
					pos: position{line: 445, col: 14, offset: 13645},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 445, col: 14, offset: 13645},
							val:        "[",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 18, offset: 13649},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 445, col: 21, offset: 13652},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 445, col: 28, offset: 13659},
								expr: &seqExpr{
									pos: position{line: 445, col: 29, offset: 13660},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 445, col: 29, offset: 13660},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 445, col: 40, offset: 13671},
											name: "__",
										},
										&zeroOrOneExpr{
											pos: position{line: 445, col: 43, offset: 13674},
											expr: &ruleRefExpr{
												pos:  position{line: 445, col: 43, offset: 13674},
												name: "ListSeparator",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 445, col: 58, offset: 13689},
											name: "__",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 445, col: 63, offset: 13694},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 445, col: 66, offset: 13697},
							val:        "]",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "ConstMap",
			pos:  position{line: 454, col: 1, offset: 13891},
			expr: &actionExpr{
				pos: position{line: 454, col: 13, offset: 13903},
				run: (*parser).callonConstMap1,
				expr: &seqExpr{
					pos: position{line: 454, col: 13, offset: 13903},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 454, col: 13, offset: 13903},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 454, col: 17, offset: 13907},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 454, col: 20, offset: 13910},
							label: "values",
							expr: &zeroOrMoreExpr{
								pos: position{line: 454, col: 27, offset: 13917},
								expr: &seqExpr{
									pos: position{line: 454, col: 28, offset: 13918},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 454, col: 28, offset: 13918},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 454, col: 39, offset: 13929},
											name: "__",
										},
										&litMatcher{
											pos:        position{line: 454, col: 42, offset: 13932},
											val:        ":",
											ignoreCase: false,
										},
										&ruleRefExpr{
											pos:  position{line: 454, col: 46, offset: 13936},
											name: "__",
										},
										&ruleRefExpr{
											pos:  position{line: 454, col: 49, offset: 13939},
											name: "ConstValue",
										},
										&ruleRefExpr{
											pos:  position{line: 454, col: 60, offset: 13950},
											name: "__",
										},
										&choiceExpr{
											pos: position{line: 454, col: 64, offset: 13954},
											alternatives: []interface{}{
												&litMatcher{
													pos:        position{line: 454, col: 64, offset: 13954},
													val:        ",",
													ignoreCase: false,
												},
												&andExpr{
													pos: position{line: 454, col: 70, offset: 13960},
													expr: &litMatcher{
														pos:        position{line: 454, col: 71, offset: 13961},
														val:        "}",
														ignoreCase: false,
													},
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 454, col: 76, offset: 13966},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 454, col: 81, offset: 13971},
							val:        "}",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "FrugalStatement",
			pos:  position{line: 474, col: 1, offset: 14521},
			expr: &ruleRefExpr{
				pos:  position{line: 474, col: 20, offset: 14540},
				name: "Scope",
			},
		},
		{
			name: "Scope",
			pos:  position{line: 476, col: 1, offset: 14547},
			expr: &actionExpr{
				pos: position{line: 476, col: 10, offset: 14556},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 476, col: 10, offset: 14556},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 476, col: 10, offset: 14556},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 476, col: 17, offset: 14563},
								expr: &seqExpr{
									pos: position{line: 476, col: 18, offset: 14564},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 476, col: 18, offset: 14564},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 476, col: 28, offset: 14574},
											name: "__",
										},
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 476, col: 33, offset: 14579},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 476, col: 41, offset: 14587},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 476, col: 44, offset: 14590},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 476, col: 49, offset: 14595},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 476, col: 60, offset: 14606},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 476, col: 63, offset: 14609},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 476, col: 70, offset: 14616},
								expr: &ruleRefExpr{
									pos:  position{line: 476, col: 70, offset: 14616},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 476, col: 78, offset: 14624},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 476, col: 81, offset: 14627},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 476, col: 85, offset: 14631},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 476, col: 88, offset: 14634},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 476, col: 99, offset: 14645},
								expr: &seqExpr{
									pos: position{line: 476, col: 100, offset: 14646},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 476, col: 100, offset: 14646},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 476, col: 110, offset: 14656},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 476, col: 116, offset: 14662},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 476, col: 116, offset: 14662},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 476, col: 122, offset: 14668},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 476, col: 139, offset: 14685},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 476, col: 141, offset: 14687},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 476, col: 153, offset: 14699},
								expr: &ruleRefExpr{
									pos:  position{line: 476, col: 153, offset: 14699},
									name: "TypeAnnotations",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 476, col: 170, offset: 14716},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 498, col: 1, offset: 15313},
			expr: &actionExpr{
				pos: position{line: 498, col: 20, offset: 15332},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 498, col: 20, offset: 15332,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 502, col: 1, offset: 15399},
			expr: &actionExpr{
				pos: position{line: 502, col: 11, offset: 15409},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 502, col: 11, offset: 15409},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 502, col: 11, offset: 15409},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 502, col: 20, offset: 15418},
							name: "__",
						},
						&ruleRefExpr{
							pos:  position{line: 502, col: 23, offset: 15421},
							name: "PrefixToken",
						},
						&zeroOrMoreExpr{
							pos: position{line: 502, col: 35, offset: 15433},
							expr: &seqExpr{
								pos: position{line: 502, col: 36, offset: 15434},
								exprs: []interface{}{
									&litMatcher{
										pos:        position{line: 502, col: 36, offset: 15434},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 502, col: 40, offset: 15438},
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
			pos:  position{line: 507, col: 1, offset: 15569},
			expr: &choiceExpr{
				pos: position{line: 507, col: 16, offset: 15584},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 507, col: 17, offset: 15585},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 507, col: 17, offset: 15585},
								val:        "{",
								ignoreCase: false,
							},
							&ruleRefExpr{
								pos:  position{line: 507, col: 21, offset: 15589},
								name: "PrefixWord",
							},
							&litMatcher{
								pos:        position{line: 507, col: 32, offset: 15600},
								val:        "}",
								ignoreCase: false,
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 507, col: 39, offset: 15607},
						name: "PrefixWord",
					},
				},
			},
		},
		{
			name: "PrefixWord",
			pos:  position{line: 509, col: 1, offset: 15619},
			expr: &oneOrMoreExpr{
				pos: position{line: 509, col: 15, offset: 15633},
				expr: &charClassMatcher{
					pos:        position{line: 509, col: 15, offset: 15633},
					val:        "[^\\r\\n\\t\\f .{}]",
					chars:      []rune{'\r', '\n', '\t', '\f', ' ', '.', '{', '}'},
					ignoreCase: false,
					inverted:   true,
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 511, col: 1, offset: 15651},
			expr: &actionExpr{
				pos: position{line: 511, col: 14, offset: 15664},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 511, col: 14, offset: 15664},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 511, col: 14, offset: 15664},
							label: "docstr",
							expr: &zeroOrOneExpr{
								pos: position{line: 511, col: 21, offset: 15671},
								expr: &seqExpr{
									pos: position{line: 511, col: 22, offset: 15672},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 511, col: 22, offset: 15672},
											name: "DocString",
										},
										&ruleRefExpr{
											pos:  position{line: 511, col: 32, offset: 15682},
											name: "__",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 511, col: 37, offset: 15687},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 511, col: 42, offset: 15692},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 511, col: 53, offset: 15703},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 511, col: 55, offset: 15705},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 511, col: 59, offset: 15709},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 511, col: 62, offset: 15712},
							label: "typ",
							expr: &ruleRefExpr{
								pos:  position{line: 511, col: 66, offset: 15716},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 511, col: 77, offset: 15727},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 511, col: 79, offset: 15729},
							label: "annotations",
							expr: &zeroOrOneExpr{
								pos: position{line: 511, col: 91, offset: 15741},
								expr: &ruleRefExpr{
									pos:  position{line: 511, col: 91, offset: 15741},
									name: "TypeAnnotations",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 511, col: 108, offset: 15758},
							expr: &ruleRefExpr{
								pos:  position{line: 511, col: 108, offset: 15758},
								name: "ListSeparator",
							},
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 528, col: 1, offset: 16344},
			expr: &actionExpr{
				pos: position{line: 528, col: 12, offset: 16355},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 528, col: 13, offset: 16356},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 528, col: 14, offset: 16357},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 528, col: 14, offset: 16357},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 528, col: 18, offset: 16361},
									expr: &choiceExpr{
										pos: position{line: 528, col: 19, offset: 16362},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 528, col: 19, offset: 16362},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 528, col: 26, offset: 16369},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 528, col: 33, offset: 16376},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 528, col: 41, offset: 16384},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 528, col: 41, offset: 16384},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 528, col: 46, offset: 16389},
									expr: &choiceExpr{
										pos: position{line: 528, col: 47, offset: 16390},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 528, col: 47, offset: 16390},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 528, col: 54, offset: 16397},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 528, col: 61, offset: 16404},
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
			pos:  position{line: 537, col: 1, offset: 16690},
			expr: &actionExpr{
				pos: position{line: 537, col: 15, offset: 16704},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 537, col: 15, offset: 16704},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 537, col: 15, offset: 16704},
							expr: &choiceExpr{
								pos: position{line: 537, col: 16, offset: 16705},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 537, col: 16, offset: 16705},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 537, col: 25, offset: 16714},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 537, col: 31, offset: 16720},
							expr: &choiceExpr{
								pos: position{line: 537, col: 32, offset: 16721},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 537, col: 32, offset: 16721},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 537, col: 41, offset: 16730},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 537, col: 49, offset: 16738},
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
			pos:  position{line: 541, col: 1, offset: 16793},
			expr: &charClassMatcher{
				pos:        position{line: 541, col: 18, offset: 16810},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 542, col: 1, offset: 16815},
			expr: &charClassMatcher{
				pos:        position{line: 542, col: 11, offset: 16825},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 543, col: 1, offset: 16834},
			expr: &charClassMatcher{
				pos:        position{line: 543, col: 10, offset: 16843},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 545, col: 1, offset: 16850},
			expr: &anyMatcher{
				line: 545, col: 15, offset: 16864,
			},
		},
		{
			name: "DocString",
			pos:  position{line: 546, col: 1, offset: 16866},
			expr: &actionExpr{
				pos: position{line: 546, col: 14, offset: 16879},
				run: (*parser).callonDocString1,
				expr: &seqExpr{
					pos: position{line: 546, col: 14, offset: 16879},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 546, col: 14, offset: 16879},
							val:        "/**@",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 546, col: 21, offset: 16886},
							expr: &seqExpr{
								pos: position{line: 546, col: 23, offset: 16888},
								exprs: []interface{}{
									&notExpr{
										pos: position{line: 546, col: 23, offset: 16888},
										expr: &litMatcher{
											pos:        position{line: 546, col: 24, offset: 16889},
											val:        "*/",
											ignoreCase: false,
										},
									},
									&ruleRefExpr{
										pos:  position{line: 546, col: 29, offset: 16894},
										name: "SourceChar",
									},
								},
							},
						},
						&litMatcher{
							pos:        position{line: 546, col: 43, offset: 16908},
							val:        "*/",
							ignoreCase: false,
						},
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 552, col: 1, offset: 17088},
			expr: &choiceExpr{
				pos: position{line: 552, col: 12, offset: 17099},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 552, col: 12, offset: 17099},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 552, col: 31, offset: 17118},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 553, col: 1, offset: 17136},
			expr: &seqExpr{
				pos: position{line: 553, col: 21, offset: 17156},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 553, col: 21, offset: 17156},
						expr: &ruleRefExpr{
							pos:  position{line: 553, col: 22, offset: 17157},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 553, col: 32, offset: 17167},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 553, col: 37, offset: 17172},
						expr: &seqExpr{
							pos: position{line: 553, col: 39, offset: 17174},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 553, col: 39, offset: 17174},
									expr: &litMatcher{
										pos:        position{line: 553, col: 40, offset: 17175},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 553, col: 45, offset: 17180},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 553, col: 59, offset: 17194},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 554, col: 1, offset: 17199},
			expr: &seqExpr{
				pos: position{line: 554, col: 37, offset: 17235},
				exprs: []interface{}{
					&notExpr{
						pos: position{line: 554, col: 37, offset: 17235},
						expr: &ruleRefExpr{
							pos:  position{line: 554, col: 38, offset: 17236},
							name: "DocString",
						},
					},
					&litMatcher{
						pos:        position{line: 554, col: 48, offset: 17246},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 554, col: 53, offset: 17251},
						expr: &seqExpr{
							pos: position{line: 554, col: 55, offset: 17253},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 554, col: 55, offset: 17253},
									expr: &choiceExpr{
										pos: position{line: 554, col: 58, offset: 17256},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 554, col: 58, offset: 17256},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 554, col: 65, offset: 17263},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 554, col: 71, offset: 17269},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 554, col: 85, offset: 17283},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 555, col: 1, offset: 17288},
			expr: &choiceExpr{
				pos: position{line: 555, col: 22, offset: 17309},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 555, col: 23, offset: 17310},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 555, col: 23, offset: 17310},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 555, col: 28, offset: 17315},
								expr: &seqExpr{
									pos: position{line: 555, col: 30, offset: 17317},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 555, col: 30, offset: 17317},
											expr: &ruleRefExpr{
												pos:  position{line: 555, col: 31, offset: 17318},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 555, col: 35, offset: 17322},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 555, col: 53, offset: 17340},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 555, col: 53, offset: 17340},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 555, col: 57, offset: 17344},
								expr: &seqExpr{
									pos: position{line: 555, col: 59, offset: 17346},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 555, col: 59, offset: 17346},
											expr: &ruleRefExpr{
												pos:  position{line: 555, col: 60, offset: 17347},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 555, col: 64, offset: 17351},
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
			pos:  position{line: 557, col: 1, offset: 17367},
			expr: &zeroOrMoreExpr{
				pos: position{line: 557, col: 7, offset: 17373},
				expr: &choiceExpr{
					pos: position{line: 557, col: 9, offset: 17375},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 557, col: 9, offset: 17375},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 557, col: 22, offset: 17388},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 557, col: 28, offset: 17394},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 558, col: 1, offset: 17405},
			expr: &zeroOrMoreExpr{
				pos: position{line: 558, col: 6, offset: 17410},
				expr: &choiceExpr{
					pos: position{line: 558, col: 8, offset: 17412},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 558, col: 8, offset: 17412},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 558, col: 21, offset: 17425},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 559, col: 1, offset: 17461},
			expr: &zeroOrMoreExpr{
				pos: position{line: 559, col: 7, offset: 17467},
				expr: &ruleRefExpr{
					pos:  position{line: 559, col: 7, offset: 17467},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 561, col: 1, offset: 17480},
			expr: &charClassMatcher{
				pos:        position{line: 561, col: 15, offset: 17494},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 562, col: 1, offset: 17502},
			expr: &litMatcher{
				pos:        position{line: 562, col: 8, offset: 17509},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 563, col: 1, offset: 17514},
			expr: &choiceExpr{
				pos: position{line: 563, col: 8, offset: 17521},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 563, col: 8, offset: 17521},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 563, col: 8, offset: 17521},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 563, col: 11, offset: 17524},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 563, col: 17, offset: 17530},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 563, col: 17, offset: 17530},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 563, col: 19, offset: 17532},
								expr: &ruleRefExpr{
									pos:  position{line: 563, col: 19, offset: 17532},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 563, col: 38, offset: 17551},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 563, col: 44, offset: 17557},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 563, col: 44, offset: 17557},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 563, col: 47, offset: 17560},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 565, col: 1, offset: 17565},
			expr: &notExpr{
				pos: position{line: 565, col: 8, offset: 17572},
				expr: &anyMatcher{
					line: 565, col: 9, offset: 17573,
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

func (c *current) onConst1(typ, name, value, annotations interface{}) (interface{}, error) {
	return &Constant{
		Name:        string(name.(Identifier)),
		Type:        typ.(*Type),
		Value:       value,
		Annotations: toAnnotations(annotations),
	}, nil
}

func (p *parser) callonConst1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onConst1(stack["typ"], stack["name"], stack["value"], stack["annotations"])
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
