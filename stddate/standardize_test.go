package stddate_test

import (
	"testing"

	"github.com/ourrootsorg/cms-server/stddate"
	"github.com/stretchr/testify/assert"
)

func TestStandardizeDate(t *testing.T) {
	tests := []struct {
		text    string
		date    *stddate.CompoundDate
		encoded string
	}{
		{
			text: "1 Jan 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      1,
					Month:    1,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000101",
		},
		{
			text: "Jan 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    1,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000100",
		},
		{
			text: "1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000000",
		},
		{
			text: "Feb 25, 1759/60",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      25,
					Month:    2,
					Year:     1759,
					Double:   stddate.DoubleDate,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "17590225|17600225",
		},
		{
			text: "ABT 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierAbout,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000000|18990101|19011231",
		},
		{
			text: "EST. 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityEstimated,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000000|18900101|19101231",
		},
		{
			text: "2/23/1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      23,
					Month:    2,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000223",
		},
		{
			text: "2/3/1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      3,
					Month:    2,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityAmbiguous,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000203|19000302",
		},
		{
			text: "before 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierBefore,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000000|18900101|19001231",
		},
		{
			text: "AFT 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierAfter,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{},
				Type:   stddate.CompoundNone,
			},
			encoded: "19000000|19000101|19101231",
		},
		{
			text: "between 1900 and 1910",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1910,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Type: stddate.CompoundRange,
			},
			encoded: "19000000|19000101|19101231",
		},
		{
			text: "1900 to 1910",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{
					Day:      0,
					Month:    0,
					Year:     1910,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Type: stddate.CompoundRange,
			},
			encoded: "19000000|19000101|19101231",
		},
		{
			text: "JAN QTR 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    1,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{
					Day:      0,
					Month:    3,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Type: stddate.CompoundRange,
			},
			encoded: "19000100|19000101|19000331",
		},
		{
			text: "JAN FEB MAR 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      0,
					Month:    1,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{
					Day:      0,
					Month:    3,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Type: stddate.CompoundRange,
			},
			encoded: "19000100|19000101|19000331",
		},
		{
			text: "5 or 15 jan 1900",
			date: &stddate.CompoundDate{
				First: stddate.Date{
					Day:      5,
					Month:    1,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Second: stddate.Date{
					Day:      15,
					Month:    1,
					Year:     1900,
					Double:   stddate.DoubleNone,
					Modifier: stddate.ModifierNone,
					Quality:  stddate.QualityNone,
				},
				Type: stddate.CompoundTwo,
			},
			encoded: "19000105|19000115",
		},
	}

	for _, test := range tests {
		assert.EqualValues(t, test.date, stddate.Standardize(test.text))
		assert.EqualValues(t, test.encoded, test.date.Encode())
	}
}
