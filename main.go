package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tcell "github.com/gdamore/tcell/v2"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use: "cwtch",
		Run: cwtchMain,
	}

	rootCmd.PersistentFlags().IntP("interval", "n", 2, "seconds to wait between updates")
	rootCmd.PersistentFlags().String("config", os.ExpandEnv("$HOME/.cwtch.yml"), "configuration file")
	rootCmd.PersistentFlags().String("config-dir", os.ExpandEnv("$HOME/.cwtch"), "configuration directory")
	rootCmd.PersistentFlags().BoolP("no-title", "t", false, "turn off header")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func cwtchMain(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		if err := cmd.Help(); err != nil {
			fmt.Printf("WARN: showing help failed: %v\n", err)
		}
		os.Exit(1)
	}

	command := strings.Join(args, " ")

	configFile, err := cmd.PersistentFlags().GetString("config")
	if err != nil {
		fmt.Printf("ERROR: failed to find configuration variable config: %v\n", err)
		os.Exit(1)
	}

	configDir, err := cmd.PersistentFlags().GetString("config-dir")
	if err != nil {
		fmt.Printf("ERROR: failed to find configuration variable config-dir: %v\n", err)
		os.Exit(1)
	}

	cfg, err := loadConfig(configFile, configDir)
	if err != nil {
		fmt.Printf("ERROR: failed load configuration file %s: %v\n", configFile, err)
		os.Exit(1)
	}

	waitSeconds, err := cmd.PersistentFlags().GetInt("interval")
	if err != nil {
		fmt.Printf("ERROR: failed to find configuration variable interval: %v\n", err)
		os.Exit(1)
	}

	cfg.wait = time.Duration(waitSeconds) * time.Second

	noTitle, err := cmd.PersistentFlags().GetBool("no-title")
	if err != nil {
		fmt.Printf("ERROR: failed to find configuration variable no-title: %v\n", err)
		os.Exit(1)
	}

	cfg.noTitle = noTitle

	cfg.cmd = command

	scr, err := tcell.NewScreen()
	if err != nil {
		fmt.Printf("ERROR: couldn't create new screen: %v\n", err)
		os.Exit(1)
	}

	if err := scr.Init(); err != nil {
		fmt.Printf("ERROR: couldn't initialize screen: %v\n", err)
		os.Exit(1)
	}
	defer scr.Fini()

	inputLoop(scr, cfg)
}

func inputLoop(scr tcell.Screen, cfg *config) {
	resizeC := make(chan struct{}, 1)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		defer cancelFunc()
		for {
			event := scr.PollEvent()
			switch ev := event.(type) {
			case *tcell.EventResize:
				resizeC <- struct{}{}
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlC {
					return
				}
			}
		}
	}()

	runCommand(ctx, scr, cfg)

	ticker := time.NewTicker(cfg.wait)
	for {
		select {
		case <-ctx.Done():
			return
		case <-resizeC:
			runCommand(ctx, scr, cfg)
		case <-ticker.C:
			runCommand(ctx, scr, cfg)
		}
	}
}

func runCommand(ctx context.Context, scr tcell.Screen, cfg *config) {
	defer func() {
		scr.Show()
	}()

	scr.Clear()

	var cfgGroup *configGroup

	for _, group := range cfg.groups {
		if group.CmdRegex == "" {
			cfgGroup = group
			break
		}

		if group.cmdrx.Match([]byte(cfg.cmd)) {
			cfgGroup = group
			break
		}
	}

	width, height := scr.Size()

	y := 0

	if !cfg.noTitle {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = os.Getenv("HOSTNAME")
		}

		dateStr := fmt.Sprintf("%s: %s", hostname, time.Now().Format(time.UnixDate))
		writexy(scr, width-runewidth.StringWidth(dateStr), 0, dateStr)

		everyPrefix := fmt.Sprintf("Every %s: ", cfg.wait)

		/* the -1 is for the space between cmdline and hostname, the -3 is for the ... */
		cmdlineMaxLen := width - len(dateStr) - len(everyPrefix) - 1 - 3

		shownCmdline := cfg.cmd
		if cmdlineMaxLen <= 0 {
			shownCmdline = ""
		} else if len(shownCmdline) > cmdlineMaxLen {
			shownCmdline = shownCmdline[:cmdlineMaxLen] + "..."
		}

		writexy(scr, 0, 0, everyPrefix+shownCmdline)

		scr.ShowCursor(width-1, height-1)

		y = 2 // if we show a title, we start at line 2.
	}

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", cfg.cmd)
	output, err := cmd.Output()
	if err != nil {
		writexy(scr, 0, 2, fmt.Sprintf("ERROR: failed to run %q: %v", cfg.cmd, err))
		return
	}

	outputStr := string(output)
	outputLines := strings.Split(outputStr, "\n")

	const tabWidth = 8

	for _, line := range outputLines {
		coloredLine := highlightLine(line, cfgGroup)

		x := 0
		for i, r := range coloredLine {
			scr.SetContent(x, y, r.r, nil, r.style)
			if r.r == '\t' {
				x += tabWidth - i%tabWidth
			} else {
				x += runewidth.RuneWidth(r.r)
			}
			if x >= width {
				x = 0
				y++
			}
		}

		y++
		if y >= height {
			break
		}
	}
}

func writexy(scr tcell.Screen, x, y int, str string) {
	for _, ch := range str {
		scr.SetContent(x, y, ch, nil, tcell.StyleDefault)
		x += runewidth.RuneWidth(ch)
	}
}

type pos struct {
	r     rune
	style tcell.Style
}

func highlightLine(line string, cfg *configGroup) []pos {
	runes := make([]pos, len(line))
	for i, r := range line {
		runes[i].r = r
	}

	if cfg != nil {
		for _, hl := range cfg.Highlights {
			indexes := hl.rx.FindAllStringIndex(line, -1)
			for _, indexPair := range indexes {
				for i := indexPair[0]; i < indexPair[1]; i++ {
					runes[i].style = hl.style
				}
			}
		}
	}

	return runes
}
