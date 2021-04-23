package main

import (
	"context"
	"reflect"
	"regexp"
	"testing"
	"time"

	tcell "github.com/gdamore/tcell/v2"
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
					{
						rx:    regexp.MustCompile("b"),
						style: tcell.StyleDefault.Foreground(tcell.ColorGreen),
					},
				},
			},
			[]pos{{r: 'a'}, {r: 'b', style: tcell.StyleDefault.Foreground(tcell.ColorGreen)}, {r: 'c'}},
		},
		{
			"xyz",
			&configGroup{
				Highlights: []highlight{
					{
						rx:    regexp.MustCompile("^y"),
						style: tcell.StyleDefault.Foreground(tcell.ColorGreen),
					},
				},
			},
			[]pos{{r: 'x'}, {r: 'y'}, {r: 'z'}},
		},
		{
			"我的狗",
			&configGroup{
				Highlights: []highlight{
					{
						rx:    regexp.MustCompile("的狗"),
						style: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
					},
				},
			},
			[]pos{
				{r: '我'}, {}, {},
				{
					r:     '的',
					style: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
				},
				{
					style: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
				},
				{
					style: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
				},
				{
					r:     '狗',
					style: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
				},
				{
					style: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
				},
				{
					style: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
				},
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

func TestRunCommand(t *testing.T) {
	ctx := context.Background()

	scr := tcell.NewSimulationScreen("utf-8")
	if err := scr.Init(); err != nil {
		t.Fatalf("scr.Init failed: %v", err)
	}

	cfg := &config{
		wait: 2 * time.Second,
		groups: []*configGroup{
			{
				Highlights: []highlight{
					{
						rx:    regexp.MustCompile("ll"),
						style: tcell.StyleDefault.Foreground(tcell.ColorGreen),
					},
				},
			},
		},
	}

	runCommand(ctx, scr, `echo "hello world"`, cfg)

	contents, width, height := scr.GetContents()
	_ = width
	_ = height

	expectText(t, contents, "Every 2s:")

	expectText(t, contents[2*width:], "hello world")

	expectStyle(t, contents[2*width:2*width+2], tcell.StyleDefault)
	expectStyle(t, contents[2*width+2:2*width+4], tcell.StyleDefault.Foreground(tcell.ColorGreen))
	expectStyle(t, contents[2*width+4:2*width+11], tcell.StyleDefault)

}

func expectText(t *testing.T, contents []tcell.SimCell, expectedText string) {
	idx := 0
	for _, r := range expectedText {
		if !reflect.DeepEqual(contents[idx].Runes, []rune{r}) {
			t.Fatalf("contents at index %d differs: expected %c, got %s", idx, r, string(contents[idx].Runes))
		}
		idx++
	}
}

func expectStyle(t *testing.T, contents []tcell.SimCell, expectedStyle tcell.Style) {
	for idx, c := range contents {
		if c.Style != expectedStyle {
			t.Fatalf("contents at index %d differs: expected style %v, got %v", idx, expectedStyle, c.Style)
		}
	}
}
