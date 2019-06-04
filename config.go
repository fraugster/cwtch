package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	termbox "github.com/nsf/termbox-go"
)

type config struct {
	wait   time.Duration
	groups []*configGroup
}

type highlight struct {
	Regex      string `yaml:"regex"`
	Foreground string `yaml:"fg"`
	Background string `yaml:"bg"`

	rx *regexp.Regexp    `yaml:"-"`
	fg termbox.Attribute `yaml:"-"`
	bg termbox.Attribute `yaml:"-"`
}

type configGroup struct {
	file       string         `yaml:"-"`
	CmdRegex   string         `yaml:"cmd_regex"`
	cmdrx      *regexp.Regexp `yaml:"-"`
	Highlights []highlight    `yaml:"highlights"`
}

func loadConfig(configFile string, configDir string) (*config, error) {
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
				log.Printf("WARN: couldn't load configuration file %q: %v\n", filename, err)
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
		log.Printf("WARN: couldn't load configuration file %q: %v\n", configFile, err)
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
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	cfg.file = file
	cfg.cmdrx, err = regexp.Compile(cfg.CmdRegex)
	if err != nil {
		return nil, fmt.Errorf("couldn't compile regular expression %q: %v", cfg.CmdRegex, err)
	}

	for idx, hl := range cfg.Highlights {
		cfg.Highlights[idx].fg, err = combineAttributes(hl.Foreground)
		if err != nil {
			return nil, err
		}

		cfg.Highlights[idx].bg, err = combineAttributes(hl.Background)
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

var (
	attributeMap = map[string]termbox.Attribute{
		"black":     termbox.ColorBlack,
		"red":       termbox.ColorRed,
		"green":     termbox.ColorGreen,
		"yellow":    termbox.ColorYellow,
		"blue":      termbox.ColorBlue,
		"magenta":   termbox.ColorMagenta,
		"cyan":      termbox.ColorCyan,
		"white":     termbox.ColorWhite,
		"bold":      termbox.AttrBold,
		"underline": termbox.AttrUnderline,
		"reverse":   termbox.AttrReverse,
	}
)

func combineAttributes(attrStr string) (attr termbox.Attribute, err error) {
	attrStr = strings.TrimSpace(attrStr)
	if attrStr == "" {
		return attr, nil
	}

	attrs := strings.Split(attrStr, ",")
	for _, an := range attrs {
		an = strings.TrimSpace(an)
		attrValue, ok := attributeMap[strings.ToLower(an)]
		if !ok {
			return termbox.ColorDefault, fmt.Errorf("unknown colour or attribute %q", an)

		}
		attr |= attrValue
	}
	return attr, nil
}
