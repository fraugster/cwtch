package main

import (
	"testing"

	termbox "github.com/nsf/termbox-go"
)

func TestCombineAttributes(t *testing.T) {
	testData := []struct {
		AttrStr      string
		ExpectedAttr termbox.Attribute
		ExpectErr    bool
	}{
		{"black", termbox.ColorBlack, false},
		{"", termbox.ColorDefault, false},
		{"foo", 0, true},
		{"green, bold", termbox.ColorGreen | termbox.AttrBold, false},
		{"blue,reverse", termbox.ColorBlue | termbox.AttrReverse, false},
		{"yellow,", 0, true},
		{"red, foo", 0, true},
	}

	for idx, tt := range testData {
		attr, err := combineAttributes(tt.AttrStr)
		if (err != nil) != tt.ExpectErr {
			t.Fatalf("%d. expected error = %t, err = %v", idx, tt.ExpectErr, err)
		}

		if attr != tt.ExpectedAttr {
			t.Fatalf("%d. expected attribute (%d) != returned attribute (%d)", idx, tt.ExpectedAttr, attr)
		}

		t.Logf("%d. input = %q output = %d, %v", idx, tt.AttrStr, attr, err)
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
			t.Errorf("%d. expected error = %t, err = %v", idx, tt.ExpectErr, err)
		}

		t.Logf("%d. input = %q output = %#v, %v", idx, tt.File, cfg, err)
	}
}

func TestLoadConfig(t *testing.T) {
	_, err := loadConfig("examples/kubectl.yml", "testdata")
	if err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}
}
