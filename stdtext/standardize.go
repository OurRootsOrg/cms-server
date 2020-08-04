package stdtext

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// try to match lucene's asciifolding
var transliterations = map[rune]string{
	'Æ': "AE",
	'Ð': "D",
	'Ł': "L",
	'Ø': "O",
	'Þ': "Th",
	'ß': "ss",
	'ẞ': "SS",
	'æ': "ae",
	'ð': "d",
	'ł': "l",
	'ø': "o",
	'þ': "th",
	'Œ': "OE",
	'œ': "oe",
}

type transliterator struct {
}

func (t transliterator) Transform(dst, src []byte, atEOF bool) (int, int, error) {
	var err error
	total := 0
	for i, w := 0, 0; i < len(src) && err == nil; i += w {
		var n int
		r, width := utf8.DecodeRune(src[i:])
		if d, ok := transliterations[r]; ok {
			n = copy(dst[total:], d)
			if n < len(d) {
				err = transform.ErrShortDst
			}
		} else {
			n = copy(dst[total:], src[i:i+width])
			if n < width {
				err = transform.ErrShortDst
			}
		}
		total += n
		w = width
	}

	return total, len(src), err
}

func (t transliterator) Reset() {
}

func AsciiFold(s string) string {
	var tl transliterator
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), tl, norm.NFC)
	result, _, err := transform.String(t, s)
	if err != nil {
		return s // return as-is
	}
	return result
}
