package stddate

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
)

const minYear = 200
const maxYear = 2200
const eof = rune(0)

type token struct {
	tokType tokenType
	value   string
}

type tokenType int

const (
	typeEOF tokenType = iota
	typeNoise
	typeSt
	// separators
	typeSeparator
	// date parts
	typeDay
	typeMonthDay
	typeMonthAlpha
	typeYear
	// qualifiers
	typeAbout
	typeBefore
	typeAfter
	// ranges
	typeFrom
	typeTo
	typeBetween
	typeAnd
	typeOr
	typeQuarter
	// quality
	typeEstimated
)

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	r, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return r
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *Scanner) Scan() token {
	for {
		r := s.read()
		if isLetter(r) {
			s.unread()
			typ, value := s.scanWord()
			if typ != typeSt { // ignore st, nd, rd, th
				return token{tokType: typ, value: value}
			}
		} else if isNumber(r) {
			s.unread()
			typ, value := s.scanNumber()
			return token{tokType: typ, value: value}
		} else if isSeparator(r) {
			return token{tokType: typeSeparator, value: string(r)}
		} else if r == eof {
			return token{tokType: typeEOF, value: ""}
		}
	}
}

func (s *Scanner) scanWord() (tokenType, string) {
	var buf bytes.Buffer
	for {
		if r := s.read(); r == eof {
			break
		} else if !isLetter(r) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(r)
		}
	}
	value := buf.String()
	upper := strings.ToUpper(value)

	// keyword?
	switch upper {
	case "ABOUT":
		return typeAbout, value
	case "ABT":
		return typeAbout, value
	case "CIRCA":
		return typeAbout, value
	case "CA":
		return typeAbout, value
	case "C":
		return typeAbout, value
	case "BEFORE":
		return typeBefore, value
	case "BEF":
		return typeBefore, value
	case "AFTER":
		return typeAfter, value
	case "AFT":
		return typeAfter, value
	case "FROM":
		return typeFrom, value
	case "TO":
		return typeTo, value
	case "BETWEEN":
		return typeBetween, value
	case "BET":
		return typeBetween, value
	case "BETW":
		return typeBetween, value
	case "BTW":
		return typeBetween, value
	case "AND":
		return typeAnd, value
	case "OR":
		return typeOr, value
	case "ESTIMATED":
		return typeEstimated, value
	case "EST":
		return typeEstimated, value
	case "CALCULATED":
		return typeEstimated, value
	case "CALC":
		return typeEstimated, value
	case "CAL":
		return typeEstimated, value
	case "PROBABLY":
		return typeEstimated, value
	case "PROB":
		return typeEstimated, value
	case "QUARTER":
		return typeQuarter, value
	case "QTR":
		return typeQuarter, value
	case "Q":
		return typeQuarter, value
	case "ST":
		return typeSt, value
	case "ND":
		return typeSt, value
	case "RD":
		return typeSt, value
	case "TH":
		return typeSt, value
	}

	// month?
	if months[upper] > 0 {
		return typeMonthAlpha, value
	}

	return typeNoise, value
}

func (s *Scanner) scanNumber() (tokenType, string) {
	var buf bytes.Buffer
	for {
		if r := s.read(); r == eof {
			break
		} else if !isNumber(r) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(r)
		}
	}
	value := buf.String()
	number, _ := strconv.Atoi(value)

	if number >= 1 && number <= 12 {
		return typeMonthDay, value
	} else if number >= 1 && number <= 31 {
		return typeDay, value
	} else if number >= minYear && number <= maxYear {
		return typeYear, value
	}
	return typeNoise, value
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isNumber(r rune) bool {
	return r >= '0' && r <= '9'
}

func isSeparator(r rune) bool {
	return r == '-' || r == '/'
}
