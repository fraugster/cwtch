package main

import (
	"context"
	"reflect"
	"regexp"
	"testing"
	"time"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
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

func TestInputLoop(t *testing.T) {
	scr := tcell.NewSimulationScreen("utf-8")
	if err := scr.Init(); err != nil {
		t.Fatalf("scr.Init failed: %v", err)
	}
	defer scr.Fini()

	cfg := &config{
		cmd:    `echo "hello world"`,
		wait:   100 * time.Millisecond,
		groups: []*configGroup{{}},
	}

	doneC := make(chan struct{})

	go func() {
		defer func() {
			doneC <- struct{}{}
		}()
		inputLoop(scr, cfg)
	}()

	time.Sleep(50 * time.Millisecond)

	expectText(t, scr, 0, 2, "hello world ")

	time.Sleep(100 * time.Millisecond)

	scr.InjectKey(tcell.KeyCtrlC, 0, 0)

	select {
	case <-doneC:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("after sending Ctrl-C, input loop didn't return in a timely manner")
	}
}

func TestRunCommand(t *testing.T) {
	ctx := context.Background()

	scr := tcell.NewSimulationScreen("utf-8")
	if err := scr.Init(); err != nil {
		t.Fatalf("scr.Init failed: %v", err)
	}
	defer scr.Fini()

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

	cfg.cmd = `echo "hello world"`
	runCommand(ctx, scr, cfg)

	expectText(t, scr, 0, 0, "Every 2s:")
	expectText(t, scr, 0, 2, "hello world")

	expectStyle(t, scr, 0, 2, tcell.StyleDefault, 2)
	expectStyle(t, scr, 2, 2, tcell.StyleDefault.Foreground(tcell.ColorGreen), 2)
	expectStyle(t, scr, 4, 2, tcell.StyleDefault, 7)

	cfg.cmd = `echo "bye"`
	runCommand(ctx, scr, cfg)

	expectText(t, scr, 0, 2, "bye        ")
	expectStyle(t, scr, 0, 2, tcell.StyleDefault, 11)
}

func expectText(t *testing.T, scr tcell.SimulationScreen, x, y int, expectedText string) {
	contents, width, _ := scr.GetContents()

	pos := y*width + x
	endPos := pos + runewidth.StringWidth(expectedText)
	if endPos > len(contents) {
		t.Errorf("expected end position of text %q is past end of screen contents", expectedText)
		return
	}

	contents = contents[pos:endPos]

	idx := 0
	for _, r := range expectedText {
		if !reflect.DeepEqual(contents[idx].Runes, []rune{r}) {
			t.Errorf("rune at row %d column %d differs: expected %c, got %s", y, x, r, string(contents[idx].Runes))
		}
		idx += runewidth.RuneWidth(r)
	}
}

func expectStyle(t *testing.T, scr tcell.SimulationScreen, x, y int, expectedStyle tcell.Style, numCols int) {
	contents, width, _ := scr.GetContents()

	pos := y*width + x
	endPos := pos + numCols
	if endPos > len(contents) {
		t.Errorf("expected end position is past end of screen contents")
		return
	}

	contents = contents[pos:endPos]
	for idx, c := range contents {
		if c.Style != expectedStyle {
			t.Fatalf("contents at row %d column %d differs: expected style %v, got %v", y, x+idx, expectedStyle, c.Style)
		}
	}
}
