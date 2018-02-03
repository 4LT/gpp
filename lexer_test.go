package gpp

import (
    "testing"
    "bytes"
)

func TestParseLang(t *testing.T) {
    var buffer bytes.Buffer
    buffer.WriteString(
        "#comment\n" +
        "  \n" +
        "\"\"(.*)\"\" string\n" +
        "\"{\" @conc\n" +
        "\t#indented comment\n" +
        "  \"[0-9]*\" int  ")
    lang, err := ParseLang(&buffer)

    if err != nil {
        t.Errorf("Error returned by ParseLang: %v", err)
    } else if len(lang) != 3 {
        t.Errorf("Expected language length: %v, got: %v", 3, len(lang))
    } else {
        tokTypeErr := "Expected token type: %v, got: %v"
        regexpErr := "Expected regexp: %v, got: %v"
        tokenType := lang[0].TokenType
        regexp := lang[0].Regexp.String()

        if tokenType != "string" {
            t.Errorf(tokTypeErr, "string", tokenType)
        }

        if regexp != "^\"(.*)\"" {
            t.Errorf(regexpErr, "^\"(.*)\"", regexp)
        }

        tokenType = lang[1].TokenType
        regexp = lang[1].Regexp.String()

        if tokenType != "@conc" {
            t.Errorf(tokTypeErr, "@conc", tokenType)
        }

        if regexp != "^{" {
            t.Errorf(regexpErr, "^{", regexp)
        }

        tokenType = lang[2].TokenType
        regexp = lang[2].Regexp.String()

        if tokenType != "int" {
            t.Errorf(tokTypeErr, "int", tokenType)
        }

        if regexp != "^[0-9]*" {
            t.Errorf(regexpErr, "^[0-9]*", regexp)
        }
    }
}

func TestParseLangErrorNoInitQuote(t *testing.T) {
    var buffer bytes.Buffer
    buffer.WriteString("line \"must start with a\" quote")
    _, err := ParseLang(&buffer)

    if err == nil {
        t.Error("Line not starting with quote did not cause error")
    }
}

func TestParseLangErrorNoEndQuote(t *testing.T) {
    var buffer bytes.Buffer
    buffer.WriteString("\"regexp must have an ending quote")
    _, err := ParseLang(&buffer)

    if err == nil {
        t.Error("Regexp missing end quote did not cause error")
    }
}

func TestParseLangErrorMultipleTokensAfterRegexp(t *testing.T) {
    var buffer bytes.Buffer
    buffer.WriteString("\"foo\" bar baz")
    _, err := ParseLang(&buffer)

    if err == nil {
        t.Error("Multiple tokens after regexp did not cause error")
    }
}

func TestParseLangErrorBadRegexp(t *testing.T) {
    var buffer bytes.Buffer
    buffer.WriteString("\".*)\" badregexp")
    _, err := ParseLang(&buffer)

    if err == nil {
        t.Error("Badly formatted regexp did not cause error")
    }
}

func TestParseLangErrorReservedType(t *testing.T) {
    var buffer bytes.Buffer
    buffer.WriteString("\"foo\" @strange")
    _, err := ParseLang(&buffer)

    if err == nil {
        t.Error("Use of reserved token type did not cause error")
    }
}

func TestParseLangErrorMultipleCapGroups(t *testing.T) {
    var buffer bytes.Buffer
    buffer.WriteString("\"(foo)(bar)\" foobar")
    _, err := ParseLang(&buffer)

    if err == nil {
        t.Error("Multiple capturing groups did not cause error")
    }
}

func TestLex(t *testing.T) {
    var langBuf bytes.Buffer
    var inputBuf bytes.Buffer
    tokStream := make(chan Token, 20)
    expectedValues := []Token{
        {"string", "Hello"},
        {"int", "367"},
        {"int", "13"},
        {"string", "Good bye"},
    }

    langBuf.WriteString(
        "\"\"(.*)\"\"   string\n" +
        "\"[0-9]+\"     int\n" +
        "\"\\pZ*\"      @skip")
    lang, langErr := ParseLang(&langBuf)

    if (langErr != nil) {
        t.Error("Language def error: " + langErr.Error())
        return
    }

    inputBuf.WriteString("\"Hello\"   \t367\n13  \"Good bye\"   ")
    go Lex(lang, &inputBuf, tokStream)

    tokIdx := 0
    for tok := range tokStream {
        if tokIdx > 3 {
            t.Error("Received more than expected number of tokens")
        }

        expectedTok := expectedValues[tokIdx]

        if (tok.Type != expectedTok.Type) {
            t.Errorf("Expected token type: %v, got: %v", tok.Type,
                expectedTok.Type)
        }

        if (tok.Value != expectedTok.Value) {
            t.Errorf("Expected token value: %v, got: %v", tok.Value,
                expectedTok.Value)
        }

        tokIdx++
    }

    if tokIdx < 3 {
        t.Error("Received fewer than expected number of tokens")
    }
}
