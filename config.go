package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	tcell "github.com/gdamore/tcell/v2"
)

type config struct {
	wait    time.Duration
	noTitle bool
	groups  []*configGroup
}

type highlight struct {
	Regex      string `yaml:"regex"`
	Foreground string `yaml:"fg"`
	Background string `yaml:"bg"`

	rx    *regexp.Regexp `yaml:"-"`
	style tcell.Style    `yaml:"-"`
}

type configGroup struct {
	file       string         `yaml:"-"`
	CmdRegex   string         `yaml:"cmd_regex"`
	cmdrx      *regexp.Regexp `yaml:"-"`
	Highlights []highlight    `yaml:"highlights"`
}

func loadConfig(configFile, configDir string) (*config, error) {
	cfg := &config{}

	d, err := os.Open(configDir)
	if err == nil {
		files, err := d.Readdirnames(-1)
		if err != nil {
			return nil, fmt.Errorf("couldn't read configuration directory %q: %v", configDir, err)
		}

		for _, file := range files {
			filename := filepath.Join(configDir, file)
			cfgGroup, err := loadConfigFile(filename)
			if err != nil {
				fmt.Printf("WARN: couldn't load configuration file %q: %v\n", filename, err)
				continue
			}
			cfg.groups = append(cfg.groups, cfgGroup)
		}

		sort.Slice(cfg.groups, func(i, j int) bool {
			return cfg.groups[i].file < cfg.groups[j].file
		})
	}

	cfgGroup, err := loadConfigFile(configFile)
	if err != nil {
		fmt.Printf("WARN: couldn't load configuration file %q: %v\n", configFile, err)
		return cfg, nil
	}

	cfg.groups = append(cfg.groups, cfgGroup)

	return cfg, nil
}

func loadConfigFile(file string) (*configGroup, error) {
	cfg := &configGroup{}

	f, err := os.Open(file)
	if err != nil {
		return cfg, nil
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("WARN: couldn't close file %s: %v\n", file, err)
		}
	}()

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	cfg.file = file
	cfg.cmdrx, err = regexp.Compile(cfg.CmdRegex)
	if err != nil {
		return nil, fmt.Errorf("couldn't compile regular expression %q: %v", cfg.CmdRegex, err)
	}

	for idx, hl := range cfg.Highlights {

		cfg.Highlights[idx].style, err = combineAttributes(hl.Foreground, hl.Background)
		if err != nil {
			return nil, err
		}

		cfg.Highlights[idx].rx, err = regexp.Compile(hl.Regex)
		if err != nil {
			return nil, fmt.Errorf("couldn't compile regular expression %q: %v", hl.Regex, err)
		}
	}

	return cfg, nil
}

func combineAttributes(fgAttr, bgAttr string) (style tcell.Style, err error) {
	style = tcell.StyleDefault

	colorMap := map[string]tcell.Color{
		"black":   tcell.ColorBlack,
		"red":     tcell.ColorRed,
		"green":   tcell.ColorGreen,
		"yellow":  tcell.ColorYellow,
		"blue":    tcell.ColorBlue,
		"magenta": tcell.ColorFuchsia,
		"cyan":    tcell.ColorAqua,
		"white":   tcell.ColorWhite,
	}

	fgAttr = strings.TrimSpace(fgAttr)
	if fgAttr != "" {
		fgAttrs := strings.Split(fgAttr, ",")
		for _, an := range fgAttrs {
			an = strings.TrimSpace(an)
			switch a := strings.ToLower(an); a {
			case "bold":
				style = style.Bold(true)
			case "underline":
				style = style.Underline(true)
			case "reverse":
				style = style.Reverse(true)
			case "italic":
				style = style.Italic(true)
			case "blink":
				style = style.Blink(true)
			case "dim":
				style = style.Dim(true)
			case "strikethrough":
				style = style.StrikeThrough(true)
			default:
				color, ok := colorMap[a]
				if !ok {
					return tcell.StyleDefault, fmt.Errorf("unknown foreground colour or attribute %q", an)

				}
				style = style.Foreground(color)
			}
		}
	}

	bgAttr = strings.TrimSpace(bgAttr)
	if bgAttr != "" {
		bgAttrs := strings.Split(bgAttr, ",")
		for _, an := range bgAttrs {
			color, ok := colorMap[strings.ToLower(an)]
			if !ok {
				return tcell.StyleDefault, fmt.Errorf("unknown background colour %q", an)

			}
			style = style.Background(color)
		}
	}

	return style, nil
}
