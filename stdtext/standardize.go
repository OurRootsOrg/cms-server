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
	'Ǽ': "Ae",
	'Ǝ': "Ae",
	'Ə': "Ae",
	'ǝ': "ae",
	'ǽ': "ae",
	'ǣ': "ae",
	'æ': "ae",
	'Ð': "D",
	'ð': "d",
	'Đ': "Dj",
	'đ': "dj",
	'Ł': "L",
	'ł': "l",
	'Ø': "O",
	'ø': "o",
	'Œ': "OE",
	'Ǿ': "Oe",
	'œ': "oe",
	'Þ': "Th",
	'þ': "th",
	'ẞ': "SS",
	'ß': "ss",
	'Ĳ': "Y",
	'ĳ': "y",
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

func HasByteOrderMark(s string) bool {
	return len(s) > 3 && s[0] == 0xEF && s[1] == 0xBB && s[2] == 0xBF
}

func RemoveByteOrderMark(s string) string {
	return s[3:]
}
