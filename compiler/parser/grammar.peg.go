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

var g = &grammar{
	rules: []*rule{
		{
			name: "Grammar",
			pos:  position{line: 53, col: 1, offset: 1308},
			expr: &actionExpr{
				pos: position{line: 53, col: 11, offset: 1320},
				run: (*parser).callonGrammar1,
				expr: &seqExpr{
					pos: position{line: 53, col: 11, offset: 1320},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 53, col: 11, offset: 1320},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 53, col: 14, offset: 1323},
							label: "statements",
							expr: &zeroOrMoreExpr{
								pos: position{line: 53, col: 25, offset: 1334},
								expr: &seqExpr{
									pos: position{line: 53, col: 27, offset: 1336},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 53, col: 27, offset: 1336},
											name: "Statement",
										},
										&ruleRefExpr{
											pos:  position{line: 53, col: 37, offset: 1346},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 53, col: 44, offset: 1353},
							alternatives: []interface{}{
								&ruleRefExpr{
									pos:  position{line: 53, col: 44, offset: 1353},
									name: "EOF",
								},
								&ruleRefExpr{
									pos:  position{line: 53, col: 50, offset: 1359},
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
			pos:  position{line: 72, col: 1, offset: 1992},
			expr: &actionExpr{
				pos: position{line: 72, col: 15, offset: 2008},
				run: (*parser).callonSyntaxError1,
				expr: &anyMatcher{
					line: 72, col: 15, offset: 2008,
				},
			},
		},
		{
			name: "Statement",
			pos:  position{line: 76, col: 1, offset: 2070},
			expr: &choiceExpr{
				pos: position{line: 76, col: 13, offset: 2084},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 76, col: 13, offset: 2084},
						name: "Namespace",
					},
					&ruleRefExpr{
						pos:  position{line: 76, col: 25, offset: 2096},
						name: "Scope",
					},
				},
			},
		},
		{
			name: "Namespace",
			pos:  position{line: 78, col: 1, offset: 2103},
			expr: &actionExpr{
				pos: position{line: 78, col: 13, offset: 2117},
				run: (*parser).callonNamespace1,
				expr: &seqExpr{
					pos: position{line: 78, col: 13, offset: 2117},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 78, col: 13, offset: 2117},
							val:        "namespace",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 78, col: 25, offset: 2129},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 78, col: 27, offset: 2131},
							label: "scope",
							expr: &oneOrMoreExpr{
								pos: position{line: 78, col: 33, offset: 2137},
								expr: &charClassMatcher{
									pos:        position{line: 78, col: 33, offset: 2137},
									val:        "[a-z.-]",
									chars:      []rune{'.', '-'},
									ranges:     []rune{'a', 'z'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 78, col: 42, offset: 2146},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 78, col: 44, offset: 2148},
							label: "ns",
							expr: &ruleRefExpr{
								pos:  position{line: 78, col: 47, offset: 2151},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 78, col: 58, offset: 2162},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "Scope",
			pos:  position{line: 85, col: 1, offset: 2319},
			expr: &actionExpr{
				pos: position{line: 85, col: 9, offset: 2329},
				run: (*parser).callonScope1,
				expr: &seqExpr{
					pos: position{line: 85, col: 9, offset: 2329},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 85, col: 9, offset: 2329},
							val:        "scope",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 17, offset: 2337},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 19, offset: 2339},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 85, col: 24, offset: 2344},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 35, offset: 2355},
							name: "__",
						},
						&litMatcher{
							pos:        position{line: 85, col: 38, offset: 2358},
							val:        "{",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 42, offset: 2362},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 45, offset: 2365},
							label: "prefix",
							expr: &zeroOrOneExpr{
								pos: position{line: 85, col: 52, offset: 2372},
								expr: &ruleRefExpr{
									pos:  position{line: 85, col: 52, offset: 2372},
									name: "Prefix",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 60, offset: 2380},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 85, col: 63, offset: 2383},
							label: "operations",
							expr: &zeroOrMoreExpr{
								pos: position{line: 85, col: 74, offset: 2394},
								expr: &seqExpr{
									pos: position{line: 85, col: 75, offset: 2395},
									exprs: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 85, col: 75, offset: 2395},
											name: "Operation",
										},
										&ruleRefExpr{
											pos:  position{line: 85, col: 85, offset: 2405},
											name: "__",
										},
									},
								},
							},
						},
						&choiceExpr{
							pos: position{line: 85, col: 91, offset: 2411},
							alternatives: []interface{}{
								&litMatcher{
									pos:        position{line: 85, col: 91, offset: 2411},
									val:        "}",
									ignoreCase: false,
								},
								&ruleRefExpr{
									pos:  position{line: 85, col: 97, offset: 2417},
									name: "EndOfScopeError",
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 85, col: 114, offset: 2434},
							name: "EOS",
						},
					},
				},
			},
		},
		{
			name: "EndOfScopeError",
			pos:  position{line: 102, col: 1, offset: 2947},
			expr: &actionExpr{
				pos: position{line: 102, col: 19, offset: 2967},
				run: (*parser).callonEndOfScopeError1,
				expr: &anyMatcher{
					line: 102, col: 19, offset: 2967,
				},
			},
		},
		{
			name: "Prefix",
			pos:  position{line: 106, col: 1, offset: 3038},
			expr: &actionExpr{
				pos: position{line: 106, col: 10, offset: 3049},
				run: (*parser).callonPrefix1,
				expr: &seqExpr{
					pos: position{line: 106, col: 10, offset: 3049},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 106, col: 10, offset: 3049},
							val:        "prefix",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 106, col: 19, offset: 3058},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 106, col: 21, offset: 3060},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 106, col: 26, offset: 3065},
								name: "Literal",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 106, col: 34, offset: 3073},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "Operation",
			pos:  position{line: 110, col: 1, offset: 3128},
			expr: &actionExpr{
				pos: position{line: 110, col: 13, offset: 3142},
				run: (*parser).callonOperation1,
				expr: &seqExpr{
					pos: position{line: 110, col: 13, offset: 3142},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 110, col: 13, offset: 3142},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 110, col: 18, offset: 3147},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 110, col: 29, offset: 3158},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 110, col: 31, offset: 3160},
							val:        ":",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 110, col: 35, offset: 3164},
							name: "__",
						},
						&labeledExpr{
							pos:   position{line: 110, col: 38, offset: 3167},
							label: "param",
							expr: &ruleRefExpr{
								pos:  position{line: 110, col: 44, offset: 3173},
								name: "Identifier",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 110, col: 55, offset: 3184},
							name: "_",
						},
						&ruleRefExpr{
							pos:  position{line: 110, col: 57, offset: 3186},
							name: "EOL",
						},
					},
				},
			},
		},
		{
			name: "Literal",
			pos:  position{line: 118, col: 1, offset: 3353},
			expr: &actionExpr{
				pos: position{line: 118, col: 11, offset: 3365},
				run: (*parser).callonLiteral1,
				expr: &choiceExpr{
					pos: position{line: 118, col: 12, offset: 3366},
					alternatives: []interface{}{
						&seqExpr{
							pos: position{line: 118, col: 13, offset: 3367},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 118, col: 13, offset: 3367},
									val:        "\"",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 118, col: 17, offset: 3371},
									expr: &choiceExpr{
										pos: position{line: 118, col: 18, offset: 3372},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 118, col: 18, offset: 3372},
												val:        "\\\"",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 118, col: 25, offset: 3379},
												val:        "[^\"]",
												chars:      []rune{'"'},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 118, col: 32, offset: 3386},
									val:        "\"",
									ignoreCase: false,
								},
							},
						},
						&seqExpr{
							pos: position{line: 118, col: 40, offset: 3394},
							exprs: []interface{}{
								&litMatcher{
									pos:        position{line: 118, col: 40, offset: 3394},
									val:        "'",
									ignoreCase: false,
								},
								&zeroOrMoreExpr{
									pos: position{line: 118, col: 45, offset: 3399},
									expr: &choiceExpr{
										pos: position{line: 118, col: 46, offset: 3400},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 118, col: 46, offset: 3400},
												val:        "\\'",
												ignoreCase: false,
											},
											&charClassMatcher{
												pos:        position{line: 118, col: 53, offset: 3407},
												val:        "[^']",
												chars:      []rune{'\''},
												ignoreCase: false,
												inverted:   true,
											},
										},
									},
								},
								&litMatcher{
									pos:        position{line: 118, col: 60, offset: 3414},
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
			pos:  position{line: 125, col: 1, offset: 3650},
			expr: &actionExpr{
				pos: position{line: 125, col: 14, offset: 3665},
				run: (*parser).callonIdentifier1,
				expr: &seqExpr{
					pos: position{line: 125, col: 14, offset: 3665},
					exprs: []interface{}{
						&oneOrMoreExpr{
							pos: position{line: 125, col: 14, offset: 3665},
							expr: &choiceExpr{
								pos: position{line: 125, col: 15, offset: 3666},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 125, col: 15, offset: 3666},
										name: "Letter",
									},
									&litMatcher{
										pos:        position{line: 125, col: 24, offset: 3675},
										val:        "_",
										ignoreCase: false,
									},
								},
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 125, col: 30, offset: 3681},
							expr: &choiceExpr{
								pos: position{line: 125, col: 31, offset: 3682},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 125, col: 31, offset: 3682},
										name: "Letter",
									},
									&ruleRefExpr{
										pos:  position{line: 125, col: 40, offset: 3691},
										name: "Digit",
									},
									&charClassMatcher{
										pos:        position{line: 125, col: 48, offset: 3699},
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
			pos:  position{line: 129, col: 1, offset: 3758},
			expr: &charClassMatcher{
				pos:        position{line: 129, col: 17, offset: 3776},
				val:        "[,;]",
				chars:      []rune{',', ';'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Letter",
			pos:  position{line: 130, col: 1, offset: 3781},
			expr: &charClassMatcher{
				pos:        position{line: 130, col: 10, offset: 3792},
				val:        "[A-Za-z]",
				ranges:     []rune{'A', 'Z', 'a', 'z'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "Digit",
			pos:  position{line: 131, col: 1, offset: 3801},
			expr: &charClassMatcher{
				pos:        position{line: 131, col: 9, offset: 3811},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "SourceChar",
			pos:  position{line: 133, col: 1, offset: 3818},
			expr: &anyMatcher{
				line: 133, col: 14, offset: 3833,
			},
		},
		{
			name: "Comment",
			pos:  position{line: 134, col: 1, offset: 3835},
			expr: &choiceExpr{
				pos: position{line: 134, col: 11, offset: 3847},
				alternatives: []interface{}{
					&ruleRefExpr{
						pos:  position{line: 134, col: 11, offset: 3847},
						name: "MultiLineComment",
					},
					&ruleRefExpr{
						pos:  position{line: 134, col: 30, offset: 3866},
						name: "SingleLineComment",
					},
				},
			},
		},
		{
			name: "MultiLineComment",
			pos:  position{line: 135, col: 1, offset: 3884},
			expr: &seqExpr{
				pos: position{line: 135, col: 20, offset: 3905},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 135, col: 20, offset: 3905},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 135, col: 25, offset: 3910},
						expr: &seqExpr{
							pos: position{line: 135, col: 27, offset: 3912},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 135, col: 27, offset: 3912},
									expr: &litMatcher{
										pos:        position{line: 135, col: 28, offset: 3913},
										val:        "*/",
										ignoreCase: false,
									},
								},
								&ruleRefExpr{
									pos:  position{line: 135, col: 33, offset: 3918},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 135, col: 47, offset: 3932},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "MultiLineCommentNoLineTerminator",
			pos:  position{line: 136, col: 1, offset: 3937},
			expr: &seqExpr{
				pos: position{line: 136, col: 36, offset: 3974},
				exprs: []interface{}{
					&litMatcher{
						pos:        position{line: 136, col: 36, offset: 3974},
						val:        "/*",
						ignoreCase: false,
					},
					&zeroOrMoreExpr{
						pos: position{line: 136, col: 41, offset: 3979},
						expr: &seqExpr{
							pos: position{line: 136, col: 43, offset: 3981},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 136, col: 43, offset: 3981},
									expr: &choiceExpr{
										pos: position{line: 136, col: 46, offset: 3984},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 136, col: 46, offset: 3984},
												val:        "*/",
												ignoreCase: false,
											},
											&ruleRefExpr{
												pos:  position{line: 136, col: 53, offset: 3991},
												name: "EOL",
											},
										},
									},
								},
								&ruleRefExpr{
									pos:  position{line: 136, col: 59, offset: 3997},
									name: "SourceChar",
								},
							},
						},
					},
					&litMatcher{
						pos:        position{line: 136, col: 73, offset: 4011},
						val:        "*/",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "SingleLineComment",
			pos:  position{line: 137, col: 1, offset: 4016},
			expr: &choiceExpr{
				pos: position{line: 137, col: 21, offset: 4038},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 137, col: 22, offset: 4039},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 137, col: 22, offset: 4039},
								val:        "//",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 137, col: 27, offset: 4044},
								expr: &seqExpr{
									pos: position{line: 137, col: 29, offset: 4046},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 137, col: 29, offset: 4046},
											expr: &ruleRefExpr{
												pos:  position{line: 137, col: 30, offset: 4047},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 137, col: 34, offset: 4051},
											name: "SourceChar",
										},
									},
								},
							},
						},
					},
					&seqExpr{
						pos: position{line: 137, col: 52, offset: 4069},
						exprs: []interface{}{
							&litMatcher{
								pos:        position{line: 137, col: 52, offset: 4069},
								val:        "#",
								ignoreCase: false,
							},
							&zeroOrMoreExpr{
								pos: position{line: 137, col: 56, offset: 4073},
								expr: &seqExpr{
									pos: position{line: 137, col: 58, offset: 4075},
									exprs: []interface{}{
										&notExpr{
											pos: position{line: 137, col: 58, offset: 4075},
											expr: &ruleRefExpr{
												pos:  position{line: 137, col: 59, offset: 4076},
												name: "EOL",
											},
										},
										&ruleRefExpr{
											pos:  position{line: 137, col: 63, offset: 4080},
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
			pos:  position{line: 139, col: 1, offset: 4096},
			expr: &zeroOrMoreExpr{
				pos: position{line: 139, col: 6, offset: 4103},
				expr: &choiceExpr{
					pos: position{line: 139, col: 8, offset: 4105},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 139, col: 8, offset: 4105},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 139, col: 21, offset: 4118},
							name: "EOL",
						},
						&ruleRefExpr{
							pos:  position{line: 139, col: 27, offset: 4124},
							name: "Comment",
						},
					},
				},
			},
		},
		{
			name: "_",
			pos:  position{line: 140, col: 1, offset: 4135},
			expr: &zeroOrMoreExpr{
				pos: position{line: 140, col: 5, offset: 4141},
				expr: &choiceExpr{
					pos: position{line: 140, col: 7, offset: 4143},
					alternatives: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 140, col: 7, offset: 4143},
							name: "Whitespace",
						},
						&ruleRefExpr{
							pos:  position{line: 140, col: 20, offset: 4156},
							name: "MultiLineCommentNoLineTerminator",
						},
					},
				},
			},
		},
		{
			name: "WS",
			pos:  position{line: 141, col: 1, offset: 4192},
			expr: &zeroOrMoreExpr{
				pos: position{line: 141, col: 6, offset: 4199},
				expr: &ruleRefExpr{
					pos:  position{line: 141, col: 6, offset: 4199},
					name: "Whitespace",
				},
			},
		},
		{
			name: "Whitespace",
			pos:  position{line: 143, col: 1, offset: 4212},
			expr: &charClassMatcher{
				pos:        position{line: 143, col: 14, offset: 4227},
				val:        "[ \\t\\r]",
				chars:      []rune{' ', '\t', '\r'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 144, col: 1, offset: 4235},
			expr: &litMatcher{
				pos:        position{line: 144, col: 7, offset: 4243},
				val:        "\n",
				ignoreCase: false,
			},
		},
		{
			name: "EOS",
			pos:  position{line: 145, col: 1, offset: 4248},
			expr: &choiceExpr{
				pos: position{line: 145, col: 7, offset: 4256},
				alternatives: []interface{}{
					&seqExpr{
						pos: position{line: 145, col: 7, offset: 4256},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 145, col: 7, offset: 4256},
								name: "__",
							},
							&litMatcher{
								pos:        position{line: 145, col: 10, offset: 4259},
								val:        ";",
								ignoreCase: false,
							},
						},
					},
					&seqExpr{
						pos: position{line: 145, col: 16, offset: 4265},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 145, col: 16, offset: 4265},
								name: "_",
							},
							&zeroOrOneExpr{
								pos: position{line: 145, col: 18, offset: 4267},
								expr: &ruleRefExpr{
									pos:  position{line: 145, col: 18, offset: 4267},
									name: "SingleLineComment",
								},
							},
							&ruleRefExpr{
								pos:  position{line: 145, col: 37, offset: 4286},
								name: "EOL",
							},
						},
					},
					&seqExpr{
						pos: position{line: 145, col: 43, offset: 4292},
						exprs: []interface{}{
							&ruleRefExpr{
								pos:  position{line: 145, col: 43, offset: 4292},
								name: "__",
							},
							&ruleRefExpr{
								pos:  position{line: 145, col: 46, offset: 4295},
								name: "EOF",
							},
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 147, col: 1, offset: 4300},
			expr: &notExpr{
				pos: position{line: 147, col: 7, offset: 4308},
				expr: &anyMatcher{
					line: 147, col: 8, offset: 4309,
				},
			},
		},
	},
}

func (c *current) onGrammar1(statements interface{}) (interface{}, error) {
	frugal := &Frugal{
		Namespaces: make(map[string]string),
		Scopes:     make(map[string]*Scope),
	}
	stmts := toIfaceSlice(statements)
	for _, st := range stmts {
		switch v := st.([]interface{})[0].(type) {
		case *namespace:
			frugal.Namespaces[v.scope] = v.namespace
		case *Scope:
			frugal.Scopes[v.Name] = v
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

func (c *current) onScope1(name, prefix, operations interface{}) (interface{}, error) {
	ops := operations.([]interface{})
	scope := &Scope{
		Name:       string(name.(Identifier)),
		Operations: make(map[string]*Operation, len(ops)),
		Prefix:     defaultPrefix,
	}
	if prefix != nil {
		scope.Prefix = prefix.(*ScopePrefix)
	}
	for _, o := range ops {
		op := o.([]interface{})[0].(*Operation)
		scope.Operations[op.Name] = op
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
