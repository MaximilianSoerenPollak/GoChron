package chron

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/shopspring/decimal"
)

type dataError struct {
	msg string
}

func (e *dataError) Error() string {
	return e.msg
}

var fractional bool

func fmtDuration(dur time.Duration) string {
	return fmtHours(decimal.NewFromFloat(dur.Hours()))
}

func fmtHours(hours decimal.Decimal) string {
	if fractional {
		return hours.StringFixed(2)
	} else {
		return fmt.Sprintf(
			"%s,%02s",
			hours.Floor(), // hours
			hours.Sub(hours.Floor()).
				Mul(decimal.NewFromFloat(.6)).
				Mul(decimal.NewFromInt(100)).
				Floor())
	}
}

var ansiStyleRegexp = regexp.MustCompile(`\x1b[[\d;]*m`)

func placeOverlay(bg, fg string, width, row, col int) string {
	wrappedBG := ansi.Hardwrap(bg, width, true)

	overlay := overlayStyle.Render(fg)

	bgLines := strings.Split(wrappedBG, "\n")
	overlayLines := strings.Split(overlay, "\n")

	for i, overlayLine := range overlayLines {
		bgLine := bgLines[i+row] // TODO: index handling
		if len(bgLine) < col {
			bgLine += strings.Repeat(" ", col-len(bgLine)) // add padding
		}

		bgLeft := ansi.Truncate(bgLine, col, "")
		bgRight := truncateLeft(bgLine, col+ansi.StringWidth(overlayLine))

		bgLines[i+row] = bgLeft + overlayLine + bgRight
	}

	result := strings.Join(bgLines, "\n")
	return result
}

func truncateLeft(line string, padding int) string {
	if strings.Contains(line, "\n") {
		panic("line must not contain newline")
	}

	// NOTE: line has no newline, so [strings.Join] after [strings.Split] is safe.
	wrapped := strings.Split(ansi.Hardwrap(line, padding, true), "\n")
	if len(wrapped) == 1 {
		return ""
	}

	var ansiStyle string
	ansiStyles := ansiStyleRegexp.FindAllString(wrapped[0], -1)
	if l := len(ansiStyles); l > 0 {
		ansiStyle = ansiStyles[l-1]
	}

	return ansiStyle + strings.Join(wrapped[1:], "")
}
