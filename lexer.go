package gpp

import (
    "bufio"
    "io"
    "strings"
    "fmt"
    "regexp"
)

type Token struct {
    Type string
    Value string
}

type TokenMatcher struct {
    Regexp *regexp.Regexp
    TokenType string
}

type ParseError struct {
    Source string
    Line int
    Message string
}

// "(.*)"\pZ+(.*)$
var matchLangLine *regexp.Regexp = regexp.MustCompile("\"(.*)\"\\pZ+(.*)$")

// ^[^"\pZ]+$
var matchTokenType *regexp.Regexp = regexp.MustCompile("^[^\"\\pZ]+$")

var specialTypes = []string{"@skip", "@fail", "@conc"}

func (e ParseError) Error() string {
    if e.Line < 1 {
        return e.Message
    } else {
        return fmt.Sprintf("%v: Line %v: %v", e.Source, e.Line, e.Message)
    }
}

func ParseLang(langReader io.Reader) ([]TokenMatcher, error) {
    langScanner := bufio.NewScanner(langReader)
    var lang []TokenMatcher
    var err error
    lineNum := 0

    for langScanner.Scan() && err == nil {
        lineNum++
        line := langScanner.Text()
        line = strings.TrimSpace(line)

        if len(line) > 0 && line[0] != '#' {
            submatchIdxs := matchLangLine.FindStringSubmatchIndex(line)
            var reStr, tokenType string
            var re *regexp.Regexp
            var reErr error

            if len(submatchIdxs) != 6 || submatchIdxs[0] != 0 ||
                    submatchIdxs[1] != len(line) {
                err = ParseError{
                    "Lang def",
                    lineNum,
                    "Misformatted line",
                }
            }

            if err == nil {
                reStr = line[submatchIdxs[2]:submatchIdxs[3]]
                tokenType = line[submatchIdxs[4]:submatchIdxs[5]]
                re, reErr = regexp.Compile("^" + reStr)
            }

            if err != nil {
                // do nothing
            } else if reErr != nil {
                err = ParseError{
                    "Lang def",
                    lineNum,
                    fmt.Sprintf("Regexp:\n\t%v", reErr),
                }
            } else if !matchTokenType.MatchString(tokenType) {
                err = ParseError{
                    "Lang def",
                    lineNum,
                    "Misformatted token type",
                }
            } else if re.NumSubexp() > 1 {
                err = ParseError{
                    "Lang def",
                    lineNum,
                    "Multiple capturing groups exist in regexp",
                }
            } else {
                lang = append(lang, TokenMatcher{re, tokenType})
            }

            if err == nil && tokenType[0] == '@' {
                foundType := false

                for _, specialType := range specialTypes {
                    if specialType == tokenType {
                        foundType = true
                    }
                }

                if !foundType {
                    err = ParseError{
                        "Lang def",
                        lineNum,
                        "Token type " + tokenType + " is reserved",
                    }
                }
            }
        }
    }

    if err == nil {
        err = langScanner.Err()
    }

    return lang, err
}

func lexLine(lang []TokenMatcher, tokStream chan Token, lineSlice string,
lineNum int) error {

    var err error

    for len(lineSlice) > 0 {
        for i, matcher := range lang {
            submatches := matcher.Regexp.FindStringSubmatch(lineSlice)

            if len(submatches) > 0 {
                var tokenValue string

                if len(submatches) > 1 {
                    tokenValue = submatches[1]
                } else {
                    tokenValue = submatches[0]
                }

                if matcher.TokenType == "@skip" {
                    // do nothing
                } else if matcher.TokenType == "@fail" {
                    err = ParseError{
                        "Input",
                        lineNum,
                        "Found illegal token: " + tokenValue,
                    }
                } else {
                    tokStream <- Token{ matcher.TokenType, tokenValue }
                }

                if len(submatches[0]) > 0 {
                    lineSlice = lineSlice[len(submatches[0]):]
                } else {
                    lineSlice = lineSlice[1:]
                }
                break
            } else if i == len(lang) - 1 {
                err = ParseError{
                    "Input",
                    lineNum,
                    "Could not extract token from " + lineSlice,
                }
            }
        }
    }

    return err
}

func Lex(lang []TokenMatcher, inputReader io.Reader,
        tokStream chan Token)  {
    input := bufio.NewScanner(inputReader)
    lineNum := 0
    var err error

    for err == nil && input.Scan() {
        lineNum++
        line := input.Text()
        err = lexLine(lang, tokStream, line, lineNum)
    }

    close(tokStream)
}


