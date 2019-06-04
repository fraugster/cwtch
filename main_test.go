package main

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/nsf/termbox-go"
)

func TestHighlightLine(t *testing.T) {
	testData := []struct {
		Line           string
		Config         *configGroup
		ExpectedOutput []pos
	}{
		{"hello", nil, []pos{{r: 'h'}, {r: 'e'}, {r: 'l'}, {r: 'l'}, {r: 'o'}}},
		{
			"abc",
			&configGroup{
				Highlights: []highlight{
					highlight{
						rx: regexp.MustCompile("b"),
						fg: termbox.ColorGreen,
					},
				},
			},
			[]pos{{r: 'a'}, {r: 'b', fg: termbox.ColorGreen}, {r: 'c'}},
		},
		{
			"xyz",
			&configGroup{
				Highlights: []highlight{
					highlight{
						rx: regexp.MustCompile("^y"),
						fg: termbox.ColorGreen,
					},
				},
			},
			[]pos{{r: 'x'}, {r: 'y'}, {r: 'z'}},
		},
		{
			"我的狗",
			&configGroup{
				Highlights: []highlight{
					highlight{
						rx: regexp.MustCompile("的狗"),
						bg: termbox.AttrBold,
						fg: termbox.ColorYellow,
					},
				},
			},
			[]pos{
				{r: '我'}, {}, {},
				{r: '的', bg: termbox.AttrBold, fg: termbox.ColorYellow}, {bg: termbox.AttrBold, fg: termbox.ColorYellow}, {bg: termbox.AttrBold, fg: termbox.ColorYellow},
				{r: '狗', bg: termbox.AttrBold, fg: termbox.ColorYellow}, {bg: termbox.AttrBold, fg: termbox.ColorYellow}, {bg: termbox.AttrBold, fg: termbox.ColorYellow},
			},
		},
	}

	for idx, tt := range testData {
		output := highlightLine(tt.Line, tt.Config)
		if !reflect.DeepEqual(tt.ExpectedOutput, output) {
			t.Errorf("%d. expected output = %#v", idx, tt.ExpectedOutput)
			t.Errorf("%d.     real output = %#v", idx, output)
			continue
		}

		t.Logf("%d. input = %q, %#v output = %#v", idx, tt.Line, tt.Config, output)
	}
}
