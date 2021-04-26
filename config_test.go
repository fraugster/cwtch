package main

import (
	"testing"

	tcell "github.com/gdamore/tcell/v2"
)

func TestCombineAttributes(t *testing.T) {
	testData := []struct {
		FgAttrStr     string
		BgAttrStr     string
		ExpectedStyle tcell.Style
		ExpectErr     bool
	}{
		{"black", "", tcell.StyleDefault.Foreground(tcell.ColorBlack), false},
		{"", "", tcell.StyleDefault, false},
		{"foo", "", tcell.StyleDefault, true},
		{"green, bold", "", tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true), false},
		{"blue,reverse", "", tcell.StyleDefault.Foreground(tcell.ColorBlue).Reverse(true), false},
		{"yellow,", "", tcell.StyleDefault, true},
		{"red, foo", "", tcell.StyleDefault, true},
	}

	for idx, tt := range testData {
		style, err := combineAttributes(tt.FgAttrStr, tt.BgAttrStr)
		if (err != nil) != tt.ExpectErr {
			t.Fatalf("%d. expected error = %t, err = %v", idx, tt.ExpectErr, err)
		}

		if style != tt.ExpectedStyle {
			t.Fatalf("%d. expected attribute (%v) != returned attribute (%v)", idx, tt.ExpectedStyle, style)
		}

		t.Logf("%d. input = %q %q output = %d, %v", idx, tt.FgAttrStr, tt.BgAttrStr, style, err)
	}
}

func TestLoadConfigFile(t *testing.T) {
	testData := []struct {
		File      string
		ExpectErr bool
	}{
		{"examples/kubectl.yml", false},
		{"testdata/does-not-exist.yml", false}, // non-existent files are simply ignored
		{"testdata/invalid.yml", true},
		{"testdata/invalid_regex.yml", true},
		{"testdata/invalid_regex2.yml", true},
		{"testdata/invalid_color1.yml", true},
		{"testdata/invalid_color2.yml", true},
	}

	for idx, tt := range testData {
		cfg, err := loadConfigFile(tt.File)
		if (err != nil) != tt.ExpectErr {
			t.Errorf("%d. cfg = %v expected error = %t, err = %v", idx, cfg, tt.ExpectErr, err)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	_, err := loadConfig("examples/kubectl.yml", "testdata")
	if err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}
}
