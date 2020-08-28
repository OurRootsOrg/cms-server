package stddate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ourrootsorg/cms-server/stdtext"
)

const StdSuffix = "_std"

type CompoundDate struct {
	First  Date
	Second Date
	Type   CompoundType
}

func (cd CompoundDate) String() string {
	if cd.Type == CompoundNone {
		return fmt.Sprintf("%s", cd.First)
	} else {
		return fmt.Sprintf("%s: %s - %s", cd.Type, cd.First, cd.Second)
	}
}

func (cd CompoundDate) Encode() string {
	switch {
	case cd.Type == CompoundTwo:
		return cd.First.YearMmDd() + "," + cd.Second.YearMmDd()
	case cd.Type == CompoundRange:
		return cd.First.YearMmDd() + "," + cd.First.StartYearMmDd() + "-" + cd.Second.EndYearMmDd()
	case cd.First.Modifier != ModifierNone || cd.First.Quality == QualityEstimated:
		return cd.First.YearMmDd() + "," + cd.First.StartYearMmDd() + "-" + cd.First.EndYearMmDd()
	case cd.First.Quality == QualityAmbiguous:
		altDate := cd.First
		altDate.Month = cd.First.Day
		altDate.Day = cd.First.Month
		return cd.First.YearMmDd() + "," + altDate.YearMmDd()
	case cd.First.Double == DoubleDate:
		altDate := cd.First
		altDate.Year = cd.First.Year + 1
		return cd.First.YearMmDd() + "," + altDate.YearMmDd()
	default:
		return cd.First.YearMmDd()
	}
}

type Date struct {
	Day      int
	Month    int
	Year     int
	Double   DoubleType
	Modifier ModifierType
	Quality  QualityType
}

func (d Date) String() string {
	return fmt.Sprintf("%04d%02d%02d %s %s %s", d.Year, d.Month, d.Day, d.Modifier, d.Double, d.Quality)
}

func (d Date) YearMmDd() string {
	return fmt.Sprintf("%04d%02d%02d", d.Year, d.Month, d.Day)
}

func (d Date) StartYearMmDd() string {
	result := d
	switch {
	case d.Modifier == ModifierBefore || d.Quality == QualityEstimated:
		result.Year -= 10
		result.Month = 01
		result.Day = 01
	case d.Modifier == ModifierAbout:
		result.Year -= 1
		result.Month = 01
		result.Day = 01
	case d.Month == 0:
		result.Month = 01
		result.Day = 01
	case d.Day == 0:
		result.Day = 01
	}
	return result.YearMmDd()
}

func (d Date) StartYear() int {
	switch {
	case d.Modifier == ModifierBefore || d.Quality == QualityEstimated:
		return d.Year - 10
	case d.Modifier == ModifierAbout:
		return d.Year - 1
	}
	return d.Year
}

func (d Date) EndYearMmDd() string {
	result := d
	switch {
	case d.Modifier == ModifierAfter || d.Quality == QualityEstimated:
		result.Year += 10
		result.Month = 12
		result.Day = 31
	case d.Modifier == ModifierAbout:
		result.Year += 1
		result.Month = 12
		result.Day = 31
	case d.Month == 0:
		result.Month = 12
		result.Day = 31
	case d.Day == 0:
		result.Day = 31 // not necessary to get the exact last day of the month
	}
	return result.YearMmDd()
}

func (d Date) EndYear() int {
	switch {
	case d.Modifier == ModifierAfter || d.Quality == QualityEstimated:
		return d.Year + 10
	case d.Modifier == ModifierAbout:
		return d.Year + 1
	}
	return d.Year
}

type CompoundType int8

const (
	CompoundNone  CompoundType = iota
	CompoundRange              // first and second define a date range
	CompoundTwo                // first and second are two separate dates
)

func (ct CompoundType) String() string {
	switch ct {
	case CompoundNone:
		return ""
	case CompoundRange:
		return "Range"
	case CompoundTwo:
		return "Two"
	default:
		return "ERROR"
	}
}

type DoubleType int8

const (
	DoubleNone DoubleType = iota
	DoubleDate
)

func (dt DoubleType) String() string {
	switch dt {
	case DoubleNone:
		return ""
	case DoubleDate:
		return "DoubleDate"
	default:
		return "ERROR"
	}
}

type ModifierType int8

const (
	ModifierNone ModifierType = iota
	ModifierAbout
	ModifierBefore
	ModifierAfter
)

func (mt ModifierType) String() string {
	switch mt {
	case ModifierNone:
		return ""
	case ModifierAbout:
		return "About"
	case ModifierBefore:
		return "Before"
	case ModifierAfter:
		return "After"
	default:
		return "ERROR"
	}
}

type QualityType int

const (
	QualityNone QualityType = iota
	QualityEstimated
	QualityAmbiguous
)

func (qt QualityType) String() string {
	switch qt {
	case QualityNone:
		return ""
	case QualityEstimated:
		return "Estimated"
	case QualityAmbiguous:
		return "Ambiguous"
	default:
		return "ERROR"
	}
}

func Standardize(s string) *CompoundDate {
	// get tokens
	scanner := NewScanner(strings.NewReader(stdtext.AsciiFold(s)))
	tokens := []token{}
	for t := scanner.Scan(); t.tokType != typeEOF; t = scanner.Scan() {
		tokens = append(tokens, t)
	}

	// parse tokens
	pos := 0

	// Is this an early year (before minYear)?
	if d, _ := parseEarlyYear(tokens, pos); d != nil {
		return &CompoundDate{First: *d}
	}

	// ignore noise at the beginning
	for pos < len(tokens) && tokens[pos].tokType == typeNoise {
		pos++
	}

	// is this a compound date?
	if cd, _ := parseCompound(tokens, pos); cd != nil {
		return cd
	}

	// is this a quarter?
	if d, _ := parseQuarterYear(tokens, pos); d != nil {
		if d.Month > 10 {
			d.Month = 10
		}
		return &CompoundDate{First: Date{Month: d.Month, Year: d.Year}, Second: Date{Month: d.Month + 2, Year: d.Year}, Type: CompoundRange}
	}

	// skip past compound tokens
	for pos < len(tokens) &&
		(tokens[pos].tokType == typeBetween ||
			tokens[pos].tokType == typeFrom ||
			tokens[pos].tokType == typeTo ||
			tokens[pos].tokType == typeAnd ||
			tokens[pos].tokType == typeOr ||
			tokens[pos].tokType == typeNoise ||
			tokens[pos].tokType == typeSeparator) {
		pos++
	}

	// is this a single date?
	if d, _ := parseDate(tokens, pos); d != nil {
		return &CompoundDate{First: *d}
	}

	// is this a month year?
	for pos < len(tokens) && tokens[pos].tokType != typeMonthAlpha && tokens[pos].tokType != typeYear {
		pos++
	}
	if d, _ := parseDate(tokens, pos); d != nil {
		return &CompoundDate{First: *d}
	}

	// is this just a year?
	for pos < len(tokens) && tokens[pos].tokType != typeYear {
		pos++
	}
	if d, _ := parseDate(tokens, pos); d != nil {
		return &CompoundDate{First: *d}
	}

	return nil
}

func parseCompound(tokens []token, start int) (*CompoundDate, int) {
	pos := start
	var typ tokenType
	var isEstimated bool
	cd := &CompoundDate{}

	// is this estimated?
	if pos < len(tokens) && tokens[pos].tokType == typeEstimated {
		isEstimated = true
		pos++
	}

	// is this a compound date?
	if pos < len(tokens) && tokens[pos].tokType == typeBetween {
		typ = typeBetween
		pos++
	} else if pos < len(tokens) && tokens[pos].tokType == typeFrom {
		typ = typeFrom
		pos++
	}
	d, pos := parseDate(tokens, pos)
	switch {
	case d != nil:
		cd.First = *d
	case pos < len(tokens) && tokens[pos].tokType == typeMonthAlpha:
		cd.First = Date{Month: GetMonthNum(tokens[pos].value)}
		pos++
	case pos < len(tokens) && (tokens[pos].tokType == typeMonthDay || tokens[pos].tokType == typeDay):
		day, _ := strconv.Atoi(tokens[pos].value)
		cd.First = Date{Day: day}
		pos++
		if pos < len(tokens) && tokens[pos].tokType == typeMonthAlpha {
			cd.First.Month = GetMonthNum(tokens[pos].value)
			pos++
		}
	default:
		// couldn't find a compound date indicator
		return nil, start
	}

	// is this a range or two dates?
	switch {
	case pos < len(tokens) &&
		((typ == typeBetween && (tokens[pos].tokType == typeAnd || (tokens[pos].tokType == typeSeparator && tokens[pos].value == "-"))) ||
			(typ != typeBetween && (tokens[pos].tokType == typeTo || (tokens[pos].tokType == typeSeparator && tokens[pos].value == "-")))):
		cd.Type = CompoundRange
		pos++
	case pos < len(tokens) && typ != typeBetween && typ != typeFrom:
		cd.Type = CompoundTwo
		for pos < len(tokens) && (tokens[pos].tokType == typeNoise || tokens[pos].tokType == typeOr || tokens[pos].tokType == typeSeparator) {
			pos++
		}
	}

	// parse the second date
	d, pos = parseDate(tokens, pos)
	if d == nil {
		return nil, start
	}
	cd.Second = *d

	// invalid date
	if cd.First.Year == 0 && cd.Second.Year == 0 {
		return nil, start
	}

	if isEstimated {
		cd.First.Quality = QualityEstimated
		cd.Second.Quality = QualityEstimated
	}

	// BETWEEN MONTH_ALPHA|DAY AND date
	if cd.First.Month == 0 {
		cd.First.Month = cd.Second.Month
	}
	if cd.First.Year == 0 {
		cd.First.Year = cd.Second.Year
	}

	// turn range into two dates if second date is before first
	if cd.Type == CompoundRange && cd.First.YearMmDd() > cd.Second.YearMmDd() {
		cd.Type = CompoundTwo
		cd.Second, cd.First = cd.First, cd.Second
	}

	return cd, pos
}

func parseDate(tokens []token, start int) (*Date, int) {
	pos := start
	date := &Date{}

	// ESTIMATED?
	if pos < len(tokens) && tokens[pos].tokType == typeEstimated {
		date.Quality = QualityEstimated
		pos++
	}

	// BEFORE|AFTER|ABOUT including "ABT AFT 1907"
loop:
	for pos < len(tokens) {
		switch tokens[pos].tokType {
		case typeBefore:
			date.Modifier = ModifierBefore
			pos++
		case typeAfter:
			date.Modifier = ModifierAfter
			pos++
		case typeAbout:
			date.Modifier = ModifierAbout
			pos++
		default:
			break loop
		}
	}

	// ESTIMATED?
	if pos < len(tokens) && tokens[pos].tokType == typeEstimated {
		date.Quality = QualityEstimated
		pos++
	}

	// date
	if d, p := parseMonthAlphaDayYear(tokens, pos); d != nil {
		date.Day = d.Day
		date.Month = d.Month
		date.Year = d.Year
		pos = p
	} else if d, p := parseDayMonthAlphaYear(tokens, pos); d != nil {
		date.Day = d.Day
		date.Month = d.Month
		date.Year = d.Year
		pos = p
	} else if d, p := parseMonthDayYear(tokens, pos); d != nil {
		date.Day = d.Day
		date.Month = d.Month
		date.Year = d.Year
		if d.Day >= 1 && d.Day <= 12 && d.Month > 0 {
			date.Quality = QualityAmbiguous
		}
		pos = p
	} else if d, p := parseDayMonthYear(tokens, pos); d != nil {
		date.Day = d.Day
		date.Month = d.Month
		date.Year = d.Year
		if d.Day >= 1 && d.Day <= 12 && d.Month > 0 {
			date.Quality = QualityAmbiguous
		}
		pos = p
	} else {
		return nil, start
	}

	// double date (e.g., 1756/7 or 1759/60)?
	if pos < len(tokens) && tokens[pos].tokType == typeSeparator && tokens[pos].value == "/" {
		if pos+1 < len(tokens) {
			if nextYear, err := strconv.Atoi(tokens[pos+1].value); err == nil {
				if (nextYear < 10 && (date.Year+1)%10 == nextYear) ||
					(nextYear < 100 && (date.Year+1)%100 == nextYear) ||
					(date.Year+1 == nextYear) {
					date.Double = DoubleDate
					pos += 2
				}
			}
		}
	}

	// ESTIMATED?
	if pos < len(tokens) && tokens[pos].tokType == typeEstimated {
		date.Quality = QualityEstimated
		pos++
	}

	return date, pos

}

// MONTH_ALPHA DAY YEAR
func parseMonthAlphaDayYear(tokens []token, start int) (*Date, int) {
	pos := start
	d := &Date{}

	if pos < len(tokens) && tokens[pos].tokType == typeMonthAlpha {
		d.Month = GetMonthNum(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	if pos < len(tokens) && (tokens[pos].tokType == typeMonthDay || tokens[pos].tokType == typeDay) {
		d.Day, _ = strconv.Atoi(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	if pos < len(tokens) && tokens[pos].tokType == typeYear {
		d.Year, _ = strconv.Atoi(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	return d, pos
}

// DAY? separator? MONTH_ALPHA separator? YEAR
func parseDayMonthAlphaYear(tokens []token, start int) (*Date, int) {
	pos := start
	d := &Date{}

	if pos < len(tokens) && (tokens[pos].tokType == typeMonthDay || tokens[pos].tokType == typeDay) {
		d.Day, _ = strconv.Atoi(tokens[pos].value)
		pos++
	}

	if pos < len(tokens) && tokens[pos].tokType == typeSeparator {
		pos++
	}

	if pos < len(tokens) && tokens[pos].tokType == typeMonthAlpha {
		d.Month = GetMonthNum(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	if pos < len(tokens) && tokens[pos].tokType == typeSeparator {
		pos++
	}

	if pos < len(tokens) && tokens[pos].tokType == typeYear {
		d.Year, _ = strconv.Atoi(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	return d, pos
}

// MONTH_DAY separator? DAY separator? YEAR
func parseMonthDayYear(tokens []token, start int) (*Date, int) {
	pos := start
	d := &Date{}

	if pos < len(tokens) && tokens[pos].tokType == typeMonthDay {
		d.Month, _ = strconv.Atoi(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	if pos < len(tokens) && tokens[pos].tokType == typeSeparator {
		pos++
	}

	if pos < len(tokens) && (tokens[pos].tokType == typeMonthDay || tokens[pos].tokType == typeDay) {
		d.Day, _ = strconv.Atoi(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	if pos != len(tokens) && tokens[pos].tokType == typeSeparator {
		pos++
	}

	if pos < len(tokens) && tokens[pos].tokType == typeYear {
		d.Year, _ = strconv.Atoi(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	return d, pos
}

// DAY? separator? MONTH_DAY? separator? YEAR
func parseDayMonthYear(tokens []token, start int) (*Date, int) {
	pos := start
	d := &Date{}

	if pos < len(tokens) && (tokens[pos].tokType == typeMonthDay || tokens[pos].tokType == typeDay) {
		d.Day, _ = strconv.Atoi(tokens[pos].value)
		pos++
	}

	if pos < len(tokens) && tokens[pos].tokType == typeSeparator {
		pos++
	}

	if pos < len(tokens) && tokens[pos].tokType == typeMonthDay {
		d.Month, _ = strconv.Atoi(tokens[pos].value)
		pos++
	}

	if pos != len(tokens) && tokens[pos].tokType == typeSeparator {
		pos++
	}

	if pos < len(tokens) && tokens[pos].tokType == typeYear {
		d.Year, _ = strconv.Atoi(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	if d.Day > 0 && d.Day <= 12 && d.Month == 0 {
		d.Month = d.Day
		d.Day = 0
	}
	return d, pos
}

// MONTH_ALPHA quarter YEAR or
// MONTH_ALPHA (separator? MONTH_ALPHA+1) separator? MONTH_ALPHA+2 YEAR
func parseQuarterYear(tokens []token, start int) (*Date, int) {
	pos := start
	d := &Date{}

	if pos < len(tokens) && tokens[pos].tokType == typeMonthAlpha {
		d.Month = GetMonthNum(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	if pos < len(tokens) && tokens[pos].tokType == typeQuarter {
		pos++
	} else {
		if pos < len(tokens) && tokens[pos].tokType == typeSeparator {
			pos++
		}
		if pos < len(tokens) && tokens[pos].tokType == typeMonthAlpha && GetMonthNum(tokens[pos].value) == d.Month+2 {
			pos++
		} else {
			if pos < len(tokens) && tokens[pos].tokType == typeMonthAlpha && GetMonthNum(tokens[pos].value) == d.Month+1 {
				pos++
			} else {
				return nil, start
			}
			if pos < len(tokens) && tokens[pos].tokType == typeSeparator {
				pos++
			}
			if pos < len(tokens) && tokens[pos].tokType == typeMonthAlpha && GetMonthNum(tokens[pos].value) == d.Month+2 {
				pos++
			} else {
				return nil, start
			}
		}
	}

	if pos < len(tokens) && tokens[pos].tokType == typeYear {
		d.Year, _ = strconv.Atoi(tokens[pos].value)
		pos++
	} else {
		return nil, start
	}

	return d, pos
}

// YEAR AD|BC|CE and year < minYear
func parseEarlyYear(tokens []token, start int) (*Date, int) {
	pos := start
	d := &Date{}

	pos++ // skip past possible year
	for pos < len(tokens) {
		if tokens[pos].value == "BC" || tokens[pos].value == "AD" || tokens[pos].value == "CE" {
			if y, err := strconv.Atoi(tokens[pos-1].value); err == nil {
				if tokens[pos].value == "BC" {
					y = 1
				}
				if y < minYear {
					d.Year = y
					return d, pos + 1
				}
			}
		}
		pos++
	}

	// couldn't find an early year
	return nil, start
}
