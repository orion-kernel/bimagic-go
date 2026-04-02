package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const VERSION = "v1.6.2"

var theme = map[string]string{
	"BIMAGIC_PRIMARY":   "212",
	"BIMAGIC_SECONDARY": "51",
	"BIMAGIC_SUCCESS":   "46",
	"BIMAGIC_ERROR":     "196",
	"BIMAGIC_WARNING":   "214",
	"BIMAGIC_INFO":      "39",
	"BIMAGIC_MUTED":     "240",
	"BANNER_COLOR_1":    "51",
	"BANNER_COLOR_2":    "45",
	"BANNER_COLOR_3":    "39",
	"BANNER_COLOR_4":    "99",
	"BANNER_COLOR_5":    "135",
}

func main() {
	// 1. Handle Version Flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Bimagic Git Wizard %s\n", VERSION)
		os.Exit(0)
	}

	// 2. Load Config & Theme
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "bimagic")
	themeFile := filepath.Join(configDir, "theme.wz")
	loadTheme(themeFile)

	// 3. Ensure gum is installed
	if !hasCmd("gum") {
		fmt.Println("Error: gum is not installed.")
		fmt.Println("Please install it: https://github.com/charmbracelet/gum")
		os.Exit(1)
	}

	// 4. Parse CLI arguments
	var cliMode, cliURL, cliMsg, cliDepth string
	cliInteractive := false

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-d":
			cliMode = "clone"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				cliURL = args[i+1]
				i++
			}
		case "--depth":
			if i+1 < len(args) {
				cliDepth = args[i+1]
				i++
			}
		case "-i":
			cliInteractive = true
		case "-z":
			cliMode = "lazy"
			if i+1 < len(args) {
				cliMsg = args[i+1]
				i++
			}
		case "-s":
			cliMode = "status"
		case "-u":
			cliMode = "undo"
		case "-g":
			cliMode = "graph"
		case "-p":
			cliMode = "pull"
		case "-a", "--architect":
			cliMode = "architect"
		default:
			if cliMode == "clone" && cliURL == "" {
				cliURL = args[i]
			} else if cliMode == "lazy" && cliMsg == "" {
				cliMsg = args[i]
			}
		}
	}

	// 5. Handle direct CLI Modes
	switch cliMode {
	case "clone":
		if cliURL == "" {
			printError("Error: Repository URL required with -d")
			os.Exit(1)
		}
		cloneRepo(cliURL, cliInteractive, cliDepth)
		os.Exit(0)
	case "status":
		showRepoStatus()
		os.Exit(0)
	case "pull":
		pullLatestChanges()
		os.Exit(0)
	case "graph":
		if !isGitRepo() {
			printError("Not a git repository!")
			os.Exit(1)
		}
		prettyGitLog()
		os.Exit(0)
	case "architect":
		summonGitignore()
		os.Exit(0)
	case "undo":
		timeTurner()
		os.Exit(0)
	case "lazy":
		lazyWizard(cliMsg)
		os.Exit(0)
	}

	// Warn if credentials are not set
	if os.Getenv("GITHUB_USER") == "" || os.Getenv("GITHUB_TOKEN") == "" {
		printWarning("GITHUB_USER or GITHUB_TOKEN not set. Defaulting to SSH/Local mode.")
	}

	// 6. Welcome Banner Logic
	versionFile := filepath.Join(configDir, "version")
	storedVersion := ""
	if b, err := os.ReadFile(versionFile); err == nil {
		storedVersion = strings.TrimSpace(string(b))
	}

	if storedVersion != VERSION {
		showWelcomeBanner(versionFile, configDir, storedVersion == "")
	} else {
		fmt.Println("Welcome to the Git Wizard! Let's work some magic...\n")
	}

	// 7. Interactive Main Loop
	for {
		clearScreen()
		showRepoStatus()

		options := []string{
			" Clone repository",
			" Init new repo",
			" Add files",
			" Commit changes",
			" Push to remote",
			" Pull latest changes",
			" Create/switch branch",
			" Set remote",
			"󱖫 Show status",
			" Contributor Statistics",
			" Git graph",
			"󰓗 Summon the Architect (.gitignore)",
			"󰮘 Remove files/folders (rm)",
			" Merge branches",
			" Uninitialize repo",
			"󰔪 Summon the Resurrection Stone (Recover lost code)",
			"󰁯 Revert commit(s)",
			"󰓗 Stash operations",
			"󰈈 The Scrying Glass (Quick View)",
			"󰿅 Exit",
		}

		choice := gumChoose(
			" Choose your spell: (j/k to navigate)",
			" ",
			theme["BIMAGIC_PRIMARY"],
			options...,
		)
		fmt.Println()

		switch choice {
		case " Clone repository":
			repoURL := gumInput("Enter repository URL", "")
			if repoURL == "" {
				continue
			}
			repoDepth := gumInput("Enter depth (empty for full clone)", "")
			cloneMode := gumChoose("", "", "", "Standard Clone", "Interactive (Select files)")
			cloneRepo(repoURL, cloneMode == "Interactive (Select files)", repoDepth)

		case "󰓗 Stash operations":
			stashOperations()

		case " Init new repo":
			dirname := gumInput("Enter repo directory name (or '.' for current dir)", "")
			if dirname == "" {
				printWarning("Operation cancelled.")
				continue
			}
			if dirname == "." {
				printCommand("git init")
				runGitCmd("init")
				cwd, _ := os.Getwd()
				printStatus("Repo initialized in current directory: " + cwd)
			} else {
				os.MkdirAll(dirname, 0o755)
				originalDir, _ := os.Getwd()
				os.Chdir(dirname)
				printCommand("git init")
				runGitCmd("init")
				branch := getGitOutput("symbolic-ref", "--short", "HEAD")
				if branch == "master" {
					printCommand("git branch -M main")
					runGitCmd("branch", "-M", "main")
					fmt.Println("Default branch renamed from 'master' to 'main' in " + dirname)
				}
				os.Chdir(originalDir)
				printStatus("Repo initialized in new directory: " + dirname)
			}

		case " Add files":
			addFilesLogic()

		case " Commit changes":
			commitMode := gumChoose("", "", "", "󰦥 Magic Commit (Builder)", "󱐋 Quick Commit (One-line)")
			if commitMode == "󰦥 Magic Commit (Builder)" {
				commitWizard()
			} else {
				msg := gumInput("Enter commit message", "")
				if msg == "" {
					printWarning("No commit message provided. Cancelled.")
					continue
				}
				if gumConfirm("Commit changes?") {
					printCommand(`git commit -m "` + msg + `"`)
					err := runGitCmd("commit", "-m", msg)
					if err == nil {
						printStatus("Commit done!")
					} else {
						printStatus("Commit cancelled.") // Note: bash logic says commit done if it succeeds, but prints cancelled if not. Actually, bash says if git commit -m; then print_status else print_status "Commit cancelled."
					}
				} else {
					printStatus("Commit cancelled.")
				}
			}

		case " Push to remote":
			pushToRemote()

		case " Pull latest changes":
			pullChangesInteractive()

		case " Create/switch branch":
			createSwitchBranch()

		case " Set remote":
			setupRemote("origin")

		case "󱖫 Show status":
			runCmdOutToScreen("git", "status")

		case "󰮘 Remove files/folders (rm)":
			removeFilesLogic()

		case " Uninitialize repo":
			uninitializeRepo()

		case "󰈈 The Scrying Glass (Quick View)":
			scryingGlass()

		case "󰿅 Exit":
			if gumConfirm("Are you sure you want to exit?") {
				fmt.Println("Git Wizard vanishes in a puff of smoke...")
				os.Exit(0)
			} else {
				continue
			}

		case " Merge branches":
			mergeBranches()

		case " Contributor Statistics":
			showContributorStats()

		case "󰓗 Summon the Architect (.gitignore)":
			summonGitignore()

		case "󰔪 Summon the Resurrection Stone (Recover lost code)":
			resurrectCommit()

		case "󰁯 Revert commit(s)":
			revertCommits()

		case " Git graph":
			drawGitGraphBox()
			gumSpin("Drawing git graph...", "sleep", "2")
			prettyGitLog()

		default:
			printError("Invalid choice! Try again.")
			fmt.Println("Git Wizard vanishes in a puff of smoke...")
			break
		}

		fmt.Println()
		gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_MUTED"]), "Press Enter to continue...")
		waitForEnter()
	}
}

// --- UI & Helper Functions ---

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func waitForEnter() {
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func getAnsiEsc(color string) string {
	color = strings.TrimSpace(color)
	if strings.HasPrefix(color, "#") && len(color) == 7 {
		r, _ := strconv.ParseInt(color[1:3], 16, 64)
		g, _ := strconv.ParseInt(color[3:5], 16, 64)
		b, _ := strconv.ParseInt(color[5:7], 16, 64)
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
	}
	return fmt.Sprintf("\033[38;5;%sm", color)
}

func playSound(soundType string) {
	switch soundType {
	case "success":
		fmt.Print("\a")
		for i := 0; i < 2; i++ {
			fmt.Print("\a")
			time.Sleep(100 * time.Millisecond)
		}
	case "error":
		for i := 0; i < 3; i++ {
			fmt.Print("\a")
			time.Sleep(50 * time.Millisecond)
		}
	case "warning":
		fmt.Print("\a")
	case "magic":
		for i := 0; i < 3; i++ {
			fmt.Print("\a")
			time.Sleep(200 * time.Millisecond)
		}
	case "progress":
		fmt.Print("\a")
	}
}

func printCommand(cmd string) {
	gray := getAnsiEsc(theme["BIMAGIC_MUTED"])
	purple := getAnsiEsc(theme["BIMAGIC_PRIMARY"])
	nc := "\033[0m"
	// Note: Bimagic uses WHITE for command output but didn't define it in the bash script.
	// Falling back to terminal default by just not adding a color or using a standard sequence.
	fmt.Printf("%s %sCommand:%s %s%s\n", gray, purple, nc, cmd, nc)
}

func printStatus(msg string) {
	gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_PRIMARY"]), msg)
	playSound("success")
}

func printError(msg string) {
	gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_ERROR"]), msg)
	playSound("error")
}

func printWarning(msg string) {
	gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_WARNING"]), msg)
	playSound("warning")
}

func drawProgressBar(label string, percent int) {
	width := 30
	filled := (percent * width) / 100
	empty := width - filled

	bar := strings.Repeat("█", filled)
	bg := strings.Repeat("░", empty)

	c := getAnsiEsc(theme["BIMAGIC_INFO"])
	p := getAnsiEsc(theme["BIMAGIC_PRIMARY"])
	g := getAnsiEsc(theme["BIMAGIC_MUTED"])
	y := getAnsiEsc(theme["BIMAGIC_WARNING"])
	nc := "\033[0m"

	fmt.Printf("\r\033[K%s%-20s%s [%s%s%s%s%s%s] %s%3d%%%s", c, label, nc, p, bar, nc, g, bg, nc, y, percent, nc)
}

func generateBar(percentage float64) string {
	width := 20
	intPercentage := int(percentage)
	if intPercentage > 100 {
		intPercentage = 100
	}
	filled := (intPercentage * width) / 100
	empty := width - filled

	colors := []string{
		getAnsiEsc(theme["BANNER_COLOR_1"]),
		getAnsiEsc(theme["BANNER_COLOR_2"]),
		getAnsiEsc(theme["BANNER_COLOR_3"]),
		getAnsiEsc(theme["BANNER_COLOR_4"]),
		getAnsiEsc(theme["BANNER_COLOR_5"]),
	}
	gray := getAnsiEsc(theme["BIMAGIC_MUTED"])
	nc := "\033[0m"

	bar := gray + "[" + nc
	for i := 0; i < filled; i++ {
		colorIdx := (i * 5) / width
		if colorIdx > 4 {
			colorIdx = 4
		}
		bar += colors[colorIdx] + "█"
	}
	bar += gray
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	bar += "]" + nc
	return bar
}

func loadTheme(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := strings.Trim(parts[1], `"' `)
			if _, exists := theme[key]; exists {
				theme[key] = val
			}
		}
	}
}

// --- External Tools (Gum, Git) ---

func hasCmd(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runGitCmd(args ...string) error {
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func getGitOutput(args ...string) string {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func runCmdOutToScreen(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func isGitRepo() bool {
	return runGitCmd("rev-parse", "--git-dir") == nil
}

func getCurrentBranch() string {
	b := getGitOutput("branch", "--show-current")
	if b == "" {
		return "main"
	}
	return b
}

// --- Gum Wrappers ---

func gumChoose(header, cursor, cursorFg string, options ...string) string {
	args := []string{"choose"}
	if header != "" {
		args = append(args, "--header", header)
	}
	if cursor != "" {
		args = append(args, "--cursor", cursor)
	}
	if cursorFg != "" {
		args = append(args, "--cursor.foreground", cursorFg)
	}
	args = append(args, options...)

	cmd := exec.Command("gum", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func gumInput(placeholder, value string) string {
	args := []string{"input"}
	if placeholder != "" {
		args = append(args, "--placeholder", placeholder)
	}
	if value != "" {
		args = append(args, "--value", value)
	}
	cmd := exec.Command("gum", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func gumConfirm(prompt string) bool {
	cmd := exec.Command("gum", "confirm", prompt)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

func gumSpin(title string, cmdArgs ...string) bool {
	args := []string{"spin", "--title", title, "--"}
	args = append(args, cmdArgs...)
	cmd := exec.Command("gum", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

func gumStyleWithArgs(colorArg string, text string) {
	cmd := exec.Command("gum", "style", colorArg, text)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func gumFilterStdin(items, placeholder string, noLimit bool) string {
	args := []string{"filter"}
	if placeholder != "" {
		args = append(args, "--placeholder", placeholder)
	}
	if noLimit {
		args = append(args, "--no-limit")
	}
	cmd := exec.Command("gum", args...)
	cmd.Stdin = strings.NewReader(items)
	cmd.Stderr = os.Stderr
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

func gumWrite(placeholder string) string {
	args := []string{"write"}
	if placeholder != "" {
		args = append(args, "--placeholder", placeholder)
	}
	cmd := exec.Command("gum", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

// --- Specific Features ---

func showRepoStatus() {
	if !isGitRepo() {
		printWarning("Not inside a git repository!")
		return
	}

	branch := getCurrentBranch()
	upstream := getGitOutput("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")

	ahead, behind := "0", "0"
	if upstream != "" {
		if a := getGitOutput("rev-list", "--count", upstream+"..HEAD"); a != "" {
			ahead = a
		}
		if b := getGitOutput("rev-list", "--count", "HEAD.."+upstream); b != "" {
			behind = b
		}
	}

	status := "🟡 uncommitted"
	color := theme["BIMAGIC_WARNING"]

	// diff --quiet
	cleanIndex := runGitCmd("diff", "--quiet") == nil
	cleanCache := runGitCmd("diff", "--cached", "--quiet") == nil

	if cleanIndex && cleanCache {
		// check conflicts ls-files -u
		conflicts := getGitOutput("ls-files", "-u")
		if conflicts != "" {
			status = "🔴 conflicts"
			color = theme["BIMAGIC_ERROR"]
		} else {
			status = "🟢 clean"
			color = theme["BIMAGIC_SUCCESS"]
		}
	}

	displayUser := os.Getenv("GITHUB_USER")
	if displayUser == "" {
		displayUser = "SSH/Local"
	}

	content := fmt.Sprintf("GITHUB USER: %s\nBRANCH: %s\nAHEAD: %s | BEHIND: %s\nSTATUS: %s", displayUser, branch, ahead, behind, status)

	fmt.Println()
	cmd := exec.Command("gum", "style", "--border", "rounded", "--margin", "1 0", "--padding", "1 2", "--border-foreground", color, content)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func showWelcomeBanner(versionFile, configDir string, isFirstTime bool) {
	clearScreen()
	fmt.Printf("%s▗▖   ▄ ▄▄▄▄   ▗▄▖  ▗▄▄▖▄  ▗▄▄▖\033[0m\n", getAnsiEsc(theme["BANNER_COLOR_1"]))
	fmt.Printf("%s▐▌   ▄ █ █ █ ▐▌ ▐▌▐▌   ▄ ▐▌   \033[0m\n", getAnsiEsc(theme["BANNER_COLOR_2"]))
	fmt.Printf("%s▐▛▀▚▖█ █   █ ▐▛▀▜▌▐▌▝▜▌█ ▐▌   \033[0m\n", getAnsiEsc(theme["BANNER_COLOR_3"]))
	fmt.Printf("%s▐▙▄▞▘█       ▐▌ ▐▌▝▚▄▞▘█ ▝▚▄▄▖\033[0m\n", getAnsiEsc(theme["BANNER_COLOR_4"]))
	fmt.Printf("%s                              \033[0m\n", getAnsiEsc(theme["BANNER_COLOR_5"]))

	fmt.Println()
	gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_PRIMARY"]), "✨ Welcome to Bimagic Git Wizard "+VERSION+" ✨")
	fmt.Println()

	if isFirstTime {
		gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_SUCCESS"]), "It looks like this is your first time using Bimagic! Let's cast some spells.")
	} else {
		gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_SUCCESS"]), "Bimagic has been updated to "+VERSION+"! Enjoy the new magic.")
	}
	fmt.Println()

	os.MkdirAll(configDir, 0o755)
	os.WriteFile(versionFile, []byte(VERSION+"\n"), 0o644)

	gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_MUTED"]), "Press Enter to open the spellbook...")
	waitForEnter()
}

func cloneRepo(url string, interactive bool, depth string) {
	repoName := strings.TrimSuffix(filepath.Base(url), ".git")
	if _, err := os.Stat(repoName); !os.IsNotExist(err) {
		printError(fmt.Sprintf("Directory '%s' already exists.", repoName))
		return
	}

	depthArgs := []string{}
	if depth != "" {
		depthArgs = []string{"--depth", depth}
	}

	if interactive {
		printStatus(fmt.Sprintf("Initializing interactive clone for %s...", repoName))

		args := []string{"clone", "--filter=blob:none", "--no-checkout"}
		args = append(args, depthArgs...)
		args = append(args, url, repoName)

		cmdStr := fmt.Sprintf("git %s", strings.Join(args, " "))
		printCommand(cmdStr)
		printStatus("Cloning structure for " + repoName + "...")

		args = append([]string{"clone", "--progress", "--filter=blob:none", "--no-checkout"}, depthArgs...)
		args = append(args, url, repoName)
		runGitCloneWithProgress(args)

		if _, err := os.Stat(repoName); os.IsNotExist(err) {
			printError("Clone failed.")
			return
		}

		originalDir, _ := os.Getwd()
		os.Chdir(repoName)

		printStatus("Fetching file list...")
		allFiles := getGitOutput("ls-tree", "-r", "--name-only", "HEAD")
		selectedPaths := gumFilterStdin(allFiles, "Select files/folders to download (Space to select)", true)

		if selectedPaths == "" {
			printWarning("No files selected. Aborting checkout.")
			os.Chdir(originalDir)
			os.RemoveAll(repoName)
			return
		}

		gumSpin("Configuring sparse checkout...", "git", "sparse-checkout", "init", "--no-cone")

		setCmd := exec.Command("git", "sparse-checkout", "set", "--stdin")
		setCmd.Stdin = strings.NewReader(selectedPaths)
		setCmd.Run()

		printStatus("Downloading selected files...")
		runGitCloneWithProgress([]string{"checkout", "--progress", "HEAD"})

		os.Chdir(originalDir)
		printStatus("Successfully cloned selected files into '" + repoName + "'!")

	} else {
		args := []string{"clone"}
		args = append(args, depthArgs...)
		args = append(args, url, repoName)

		cmdStr := fmt.Sprintf("git %s", strings.Join(args, " "))
		printCommand(cmdStr)
		printStatus(fmt.Sprintf("Cloning %s into %s...", url, repoName))

		args = append([]string{"clone", "--progress"}, depthArgs...)
		args = append(args, url, repoName)
		success := runGitCloneWithProgress(args)

		if success {
			printStatus(fmt.Sprintf("Successfully cloned '%s' into '%s'!", url, repoName))
		} else {
			printError("Clone failed.")
		}
	}
}

// Special function to read stderr progress from git clone/checkout
func runGitCloneWithProgress(args []string) bool {
	cmd := exec.Command("git", args...)
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stderr)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	objRegex := regexp.MustCompile(`Receiving objects:\s+(\d+)%`)
	deltaRegex := regexp.MustCompile(`Resolving deltas:\s+(\d+)%`)
	updateRegex := regexp.MustCompile(`Updating files:\s+(\d+)%`)

	for scanner.Scan() {
		line := scanner.Text()
		if m := objRegex.FindStringSubmatch(line); len(m) > 1 {
			p, _ := strconv.Atoi(m[1])
			drawProgressBar("Receiving Objects", p)
		} else if m := deltaRegex.FindStringSubmatch(line); len(m) > 1 {
			p, _ := strconv.Atoi(m[1])
			drawProgressBar("Resolving Deltas", p)
		} else if m := updateRegex.FindStringSubmatch(line); len(m) > 1 {
			p, _ := strconv.Atoi(m[1])
			drawProgressBar("Updating Files", p)
		}
	}
	err := cmd.Wait()
	fmt.Println()
	return err == nil
}

func stashOperations() {
	for {
		stashChoice := gumChoose("Stash Operations", "", "",
			"󱉛 Push (Save) changes",
			"󱉙 Pop latest stash",
			" List stashes",
			" Apply specific stash",
			" Drop specific stash",
			"󰎟 Clear all stashes",
			"󰌍 Back",
		)

		switch stashChoice {
		case "󱉛 Push (Save) changes":
			msg := gumInput("Optional stash message", "")
			includeUntracked := ""
			if gumConfirm("Include untracked files?") {
				includeUntracked = "-u"
			}
			cmdStr := "git stash push"
			if includeUntracked != "" {
				cmdStr += " " + includeUntracked
			}
			if msg != "" {
				cmdStr += ` -m "` + msg + `"`
			}
			printCommand(cmdStr)
			args := []string{"stash", "push"}
			if includeUntracked != "" {
				args = append(args, includeUntracked)
			}
			if msg != "" {
				args = append(args, "-m", msg)
			}
			if runGitCmd(args...) == nil {
				printStatus("Changes stashed successfully!")
			} else {
				printError("Failed to stash changes.")
			}
		case "󱉙 Pop latest stash":
			printCommand("git stash pop")
			if runGitCmd("stash", "pop") == nil {
				printStatus("Stash popped successfully!")
			} else {
				printError("Failed to pop stash (possible conflicts).")
			}
		case " List stashes":
			stashes := getGitOutput("stash", "list")
			if stashes == "" {
				printWarning("No stashes found.")
			} else {
				cmd := exec.Command("gum", "style", "--border", "normal", "--padding", "0 1", stashes)
				cmd.Stdout = os.Stdout
				cmd.Run()
			}
		case " Apply specific stash":
			stashes := getGitOutput("stash", "list")
			if stashes == "" {
				printWarning("No stashes found.")
				continue
			}
			stashEntry := gumFilterStdin(stashes, "Select stash to apply", false)
			if stashEntry != "" {
				stashID := strings.SplitN(stashEntry, ":", 2)[0]
				printCommand("git stash apply " + stashID)
				if runGitCmd("stash", "apply", stashID) == nil {
					printStatus("Applied " + stashID)
				} else {
					printError("Failed to apply " + stashID)
				}
			}
		case " Drop specific stash":
			stashes := getGitOutput("stash", "list")
			if stashes == "" {
				printWarning("No stashes found.")
				continue
			}
			stashEntry := gumFilterStdin(stashes, "Select stash to drop", false)
			if stashEntry != "" {
				stashID := strings.SplitN(stashEntry, ":", 2)[0]
				if gumConfirm("Are you sure you want to drop " + stashID + "?") {
					printCommand("git stash drop " + stashID)
					if runGitCmd("stash", "drop", stashID) == nil {
						printStatus("Dropped " + stashID)
					} else {
						printError("Failed to drop " + stashID)
					}
				}
			}
		case "󰎟 Clear all stashes":
			if getGitOutput("stash", "list") == "" {
				printWarning("No stashes found.")
				continue
			}
			if gumConfirm("DANGER: This will delete ALL stashes. Continue?") {
				printCommand("git stash clear")
				if runGitCmd("stash", "clear") == nil {
					printStatus("All stashes cleared.")
				} else {
					printError("Failed to clear stashes.")
				}
			} else {
				printStatus("Operation cancelled.")
			}
		case "󰌍 Back":
			return
		}
		fmt.Println()
		gumStyleWithArgs(fmt.Sprintf("--foreground=%s", theme["BIMAGIC_MUTED"]), "Press Enter to continue...")
		waitForEnter()
	}
}

func addFilesLogic() {
	for {
		addChoice := gumChoose("", "", "", " Stage Files", "󰷊 Preview Files", "󰌍 Back")
		if addChoice == "󰌍 Back" {
			break
		}
		if addChoice == "󰷊 Preview Files" {
			scryingGlass()
			continue
		}

		out := getGitOutput("ls-files", "--others", "--modified", "--exclude-standard")
		filesList := "[ALL]\n" + out
		files := gumFilterStdin(filesList, "Select files to add", true)

		if files == "" {
			printWarning("No files selected.")
		} else {
			if strings.Contains(files, "[ALL]") {
				printCommand("git add .")
				runGitCmd("add", ".")
				printStatus("All files staged.")
			} else {
				for _, f := range strings.Split(files, "\n") {
					if f != "" {
						printCommand(`git add "` + f + `"`)
						runGitCmd("add", f)
					}
				}
				printStatus("Selected files staged.")
				fmt.Println(files)
			}
		}
		break
	}
}

func commitWizard() {
	nc := "\033[0m"
	fmt.Printf("%s=== The Alchemist's Commit ===%s\n", getAnsiEsc(theme["BIMAGIC_PRIMARY"]), nc)

	typeStr := gumChoose("Select change type:", "", "",
		"feat: A new feature",
		"fix: A bug fix",
		"docs: Documentation only changes",
		"style: Changes that do not affect the meaning of the code",
		"refactor: A code change that neither fixes a bug nor adds a feature",
		"perf: A code change that improves performance",
		"test: Adding missing tests or correcting existing tests",
		"chore: Changes to the build process or auxiliary tools",
	)
	if typeStr == "" {
		return
	}
	typeStr = strings.SplitN(typeStr, ":", 2)[0]

	scope := gumInput("Scope (optional, e.g., 'login', 'ui'). Press Enter to skip.", "")
	summary := gumInput("Short description (imperative mood, e.g., 'add generic login')", "")
	if summary == "" {
		printWarning("Summary is required!")
		return
	}

	body := ""
	if gumConfirm("Add a longer description (body)?") {
		body = gumWrite("Enter detailed description...")
	}

	breaking := ""
	if gumConfirm("Is this a BREAKING CHANGE?") {
		breaking = "!"
	}

	commitMsg := typeStr
	if scope != "" {
		commitMsg += "(" + scope + ")"
	}
	commitMsg += breaking + ": " + summary

	if body != "" {
		commitMsg += "\n\n" + body
	}

	fmt.Println()
	cmd := exec.Command("gum", "style", "--border", "rounded", "--border-foreground", theme["BIMAGIC_PRIMARY"], "--padding", "1 2", "PREVIEW:", commitMsg)
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println()

	if gumConfirm("Commit with this message?") {
		printCommand(`git commit -m "` + commitMsg + `"`)
		runGitCmd("commit", "-m", commitMsg)
		printStatus("󱝁 Mischief managed! (Commit successful)")
	} else {
		printWarning("Commit cancelled.")
	}
}

func pushToRemote() {
	branch := getCurrentBranch()
	remotesStr := getGitOutput("remote")

	var remote string
	if remotesStr == "" {
		printError("No remote set!")
		if setupRemote("origin") {
			remote = "origin"
		} else {
			return
		}
	} else {
		remotes := strings.Split(remotesStr, "\n")
		if len(remotes) == 1 {
			remote = remotes[0]
		} else {
			remote = gumFilterStdin(remotesStr, "Select remote to push to", false)
		}
	}

	if remote == "" {
		return
	}

	if gumConfirm(fmt.Sprintf("Push branch '%s' to '%s'?", branch, remote)) {
		fmt.Printf("Pushing branch '%s' to '%s'...\n", branch, remote)
		printCommand(fmt.Sprintf(`git push -u "%s" "%s"`, remote, branch))
		gumSpin("Pushing...", "git", "push", "-u", remote, branch)
	} else {
		printStatus("Push cancelled.")
	}
}

func pullLatestChanges() {
	printCommand("git fetch --all")
	if gumSpin("Fetching updates...", "git", "fetch", "--all") {
		printStatus("Fetch complete.")
	} else {
		printWarning("Fetch encountered issues during fetch.")
	}

	printCommand("git pull --all")
	if gumSpin("Pulling all...", "git", "pull", "--all") {
		printStatus("Pull all complete.")
	} else {
		printError("Pull failed. There might be conflicts or no upstream set.")
	}
}

func pullChangesInteractive() {
	if gumSpin("Fetching updates...", "git", "fetch", "--all") {
		printStatus("Fetch complete.")
	} else {
		printWarning("Fetch encountered issues.")
	}

	pullChoice := gumChoose("Select pull mode", "", "", "Pull specific branch", "Pull all")

	if pullChoice == "Pull all" {
		if gumConfirm("Run 'git pull --all'?") {
			printCommand("git pull --all")
			gumSpin("Pulling all...", "git", "pull", "--all")
			printStatus("Pull all complete.")
		} else {
			printStatus("Pull cancelled.")
		}
	} else if pullChoice == "Pull specific branch" {
		branch := gumInput("Enter branch to pull", "main")
		if branch == "" {
			branch = "main"
		}

		remotesStr := getGitOutput("remote")
		if remotesStr == "" {
			printError("No remote set! Cannot pull.")
			return
		}
		var remote string
		remotes := strings.Split(remotesStr, "\n")
		if len(remotes) == 1 {
			remote = remotes[0]
		} else {
			remote = gumFilterStdin(remotesStr, "Select remote to pull from", false)
		}

		if remote != "" {
			if gumConfirm(fmt.Sprintf("Pull branch '%s' from '%s'?", branch, remote)) {
				printCommand(fmt.Sprintf(`git pull "%s" "%s"`, remote, branch))
				gumSpin("Pulling...", "git", "pull", remote, branch)
			} else {
				printStatus("Pull cancelled.")
			}
		} else {
			printWarning("No remote selected.")
		}
	}
}

func createSwitchBranch() {
	currentBranch := getCurrentBranch()
	printStatus("Current branch: " + currentBranch)
	fmt.Println()
	printStatus("Available branches:")

	branchesOutput := getGitOutput("branch", "-a", "--format=%(refname:short)")
	branches := strings.Split(branchesOutput, "\n")
	uniqueBranches := make(map[string]bool)

	green := getAnsiEsc(theme["BIMAGIC_SUCCESS"])
	nc := "\033[0m"

	for _, b := range branches {
		b = strings.TrimSpace(b)
		if b == "" || uniqueBranches[b] {
			continue
		}
		uniqueBranches[b] = true
		if b == currentBranch {
			fmt.Printf("%s➤ %s%s (current)\n", green, b, nc)
		} else {
			fmt.Printf("  %s\n", b)
		}
	}
	fmt.Println()

	branchOpt := gumChoose("", "", "", "Switch to existing branch", "Create new branch")
	if branchOpt == "Switch to existing branch" {
		branchesOnly := getGitOutput("branch", "--format=%(refname:short)")
		existingBranch := gumFilterStdin(branchesOnly, "Select branch to switch to", false)
		if existingBranch != "" {
			printCommand(`git checkout "` + existingBranch + `"`)
			runGitCmd("checkout", existingBranch)
			printStatus("Switched to branch: " + existingBranch)
		} else {
			printWarning("No branch selected.")
		}
	} else if branchOpt == "Create new branch" {
		newBranch := gumInput("Enter new branch name", "")
		if newBranch != "" {
			printCommand(`git checkout -b "` + newBranch + `"`)
			runGitCmd("checkout", "-b", newBranch)
			printStatus("Created and switched to new branch: " + newBranch)
		} else {
			printError("No branch name provided.")
		}
	} else {
		printWarning("Operation cancelled.")
	}
}

func setupRemote(remoteName string) bool {
	if !isGitRepo() {
		printError("Not a git repository! Initialize it first.")
		return false
	}
	if remoteName == "" {
		remoteName = "origin"
	}

	protocol := gumChoose(fmt.Sprintf("Select protocol for '%s':", remoteName), "", "", "HTTPS (Token)", "SSH")
	var remoteURL string

	if protocol == "HTTPS (Token)" {
		ghUser := os.Getenv("GITHUB_USER")
		ghToken := os.Getenv("GITHUB_TOKEN")
		if ghUser == "" || ghToken == "" {
			printError("GITHUB_USER and GITHUB_TOKEN required for HTTPS!")
			return false
		}
		repoName := gumInput("Enter repo name (example: my-repo)", "")
		if repoName == "" {
			return false
		}
		repoName = strings.TrimSuffix(repoName, ".git") + ".git"
		remoteURL = fmt.Sprintf("https://%s@github.com/%s/%s", ghToken, ghUser, repoName)
	} else if protocol == "SSH" {
		remoteURL = gumInput("Enter SSH URL (e.g., git@github.com:user/repo.git)", "")
		if remoteURL == "" {
			return false
		}
	} else {
		printWarning("No protocol selected.")
		return false
	}

	if gumConfirm(fmt.Sprintf("Set remote '%s' to %s?", remoteName, remoteURL)) {
		printCommand(fmt.Sprintf(`git remote remove "%s"`, remoteName))
		runGitCmd("remote", "remove", remoteName)
		printCommand(fmt.Sprintf(`git remote add "%s" "%s"`, remoteName, remoteURL))
		runGitCmd("remote", "add", remoteName, remoteURL)
		printStatus(fmt.Sprintf(" Remote '%s' set to %s", remoteName, remoteURL))
		return true
	}
	printStatus("Operation cancelled.")
	return false
}

func removeFilesLogic() {
	for {
		removeChoice := gumChoose("", "", "", "Remove Files", "Preview Files", "Back")
		if removeChoice == "Back" {
			break
		}
		if removeChoice == "Preview Files" {
			scryingGlass()
			continue
		}

		out := getGitOutput("ls-files", "--cached", "--others", "--exclude-standard")
		files := gumFilterStdin(out, "Select files/folders to remove", true)

		if files == "" {
			printWarning("No files selected.")
			break
		}

		fmt.Println("Files selected for removal:")
		// equivalent to tput setaf 3
		fmt.Printf("\033[33m%s\033[0m\n\n", files)

		if gumConfirm("Confirm removal? This cannot be undone.") {
			for _, f := range strings.Split(files, "\n") {
				if f == "" {
					continue
				}
				if runGitCmd("ls-files", "--error-unmatch", f) == nil {
					printCommand(`git rm -rf "` + f + `"`)
					runGitCmd("rm", "-rf", f)
				} else {
					printCommand(`rm -rf "` + f + `"`)
					os.RemoveAll(f)
				}
			}
			printStatus("Selected files/folders have been removed.")
		} else {
			printStatus("Operation cancelled.")
		}
		break
	}
}

func uninitializeRepo() {
	printWarning("This will completely uninitialize the Git repository in this folder.")
	fmt.Println("This action will delete the .git directory and cannot be undone!\n")

	if gumConfirm("Are you sure you want to continue?") {
		if _, err := os.Stat(".git"); !os.IsNotExist(err) {
			printCommand("rm -rf .git")
			os.RemoveAll(".git")
			printStatus("Git repository has been uninitialized.")
		} else {
			printError("No .git directory found here. Nothing to do.")
		}
	} else {
		printStatus("Operation cancelled.")
	}
}

func mergeBranches() {
	currentBranch := getCurrentBranch()
	printStatus("You are on branch: " + currentBranch)
	fmt.Println()

	branchesOut := getGitOutput("branch", "--format=%(refname:short)")
	filtered := ""
	for _, b := range strings.Split(branchesOut, "\n") {
		if b != currentBranch {
			filtered += b + "\n"
		}
	}

	mergeBranch := gumFilterStdin(filtered, "Select branch to merge into "+currentBranch, false)

	if mergeBranch == "" {
		printWarning("No branch selected. Merge cancelled.")
	} else {
		if gumConfirm(fmt.Sprintf("Merge branch '%s' into '%s'?", mergeBranch, currentBranch)) {
			fmt.Printf("Merging branch '%s' into '%s'...\n", mergeBranch, currentBranch)
			printCommand(`git merge "` + mergeBranch + `"`)
			if gumSpin("Merging...", "git", "merge", mergeBranch) {
				printStatus("Merge successful!")
			} else {
				printError("Merge had conflicts! Resolve them manually.")
			}
		} else {
			printStatus("Merge cancelled.")
		}
	}
}

func drawGitGraphBox() {
	yellow := getAnsiEsc(theme["BIMAGIC_WARNING"])
	nc := "\033[0m"
	line1 := "Git graph"
	line2 := "[INFO] press 'q' to exit"

	fmt.Printf("%s╭%s╮%s\n", yellow, strings.Repeat("─", 30), nc)
	fmt.Printf("%s│%s %-28s %s│%s\n", yellow, nc, line1, yellow, nc)
	fmt.Printf("%s│%s %-28s %s│%s\n", yellow, nc, line2, yellow, nc)
	fmt.Printf("%s╰%s╯%s\n", yellow, strings.Repeat("─", 30), nc)
}

func prettyGitLog() {
	printCommand("git log --graph --oneline --decorate --all")
	runCmdOutToScreen("git", "log", "--graph", "--abbrev-commit", "--decorate", "--date=short", "--format=%C(auto)%h%Creset %C(blue)%ad%Creset %C(green)%an%Creset %C(yellow)%d%Creset %Creset%s", "--all")
}

func timeTurner() {
	printStatus("󱦟 Spinning the Time Turner...")

	if err := runGitCmd("rev-parse", "HEAD"); err != nil {
		printError("No commits to undo! This repo is empty.")
		os.Exit(1)
	}

	isInitialCommit := runGitCmd("rev-parse", "HEAD~1") != nil

	undoType := gumChoose("Select Undo Level:", "", "",
		"Soft (Undo commit, keep changes staged - Best for fixing typos)",
		"Mixed (Undo commit, keep changes unstaged - Best for splitting work)",
		"Hard (DESTROY changes - Revert to previous state)",
		"Cancel",
	)

	switch {
	case strings.HasPrefix(undoType, "Soft"):
		if isInitialCommit {
			printCommand("git update-ref -d HEAD")
			runGitCmd("update-ref", "-d", "HEAD")
		} else {
			printCommand("git reset --soft HEAD~1")
			runGitCmd("reset", "--soft", "HEAD~1")
		}
		printStatus("✨ Success! I undid the commit, but kept your files ready to commit again.")
	case strings.HasPrefix(undoType, "Mixed"):
		if isInitialCommit {
			printCommand("git update-ref -d HEAD")
			runGitCmd("update-ref", "-d", "HEAD")
			printCommand("git rm --cached -r -q .")
			runGitCmd("rm", "--cached", "-r", "-q", ".")
		} else {
			printCommand("git reset HEAD~1")
			runGitCmd("reset", "HEAD~1")
		}
		printStatus("󱞈 Success! I undid the commit and unstaged the files.")
	case strings.HasPrefix(undoType, "Hard"):
		if gumConfirm(" DANGER: This deletes your work forever. Are you sure?") {
			if isInitialCommit {
				printCommand("git update-ref -d HEAD")
				runGitCmd("update-ref", "-d", "HEAD")
				printCommand("git rm --cached -r -q .")
				runGitCmd("rm", "--cached", "-r", "-q", ".")
				printCommand("git clean -fd")
				runGitCmd("clean", "-fd")
			} else {
				printCommand("git reset --hard HEAD~1")
				runGitCmd("reset", "--hard", "HEAD~1")
			}
			printStatus("󱠇 Obliviate! The last commit and its changes are destroyed.")
		} else {
			printStatus("Operation cancelled.")
		}
	default:
		printStatus("Mischief managed (Cancelled).")
	}
}

func lazyWizard(cliMsg string) {
	if !isGitRepo() {
		printError("Not a git repository!")
		os.Exit(1)
	}
	if cliMsg == "" {
		printError("Error: Commit message required for Lazy Wizard (-z)")
		fmt.Println("Usage: bimagic -z \"commit message\"")
		os.Exit(1)
	}

	printStatus("  Lazy Wizard invoked!")

	printCommand("git add .")
	if gumSpin("Adding files...", "git", "add", ".") {
		printStatus("Files added.")
	} else {
		printError("Failed to add files.")
		os.Exit(1)
	}

	printCommand(`git commit -m "` + cliMsg + `"`)
	if runGitCmd("commit", "-m", cliMsg) == nil {
		printStatus("Committed: " + cliMsg)
	} else {
		printError("Commit failed (nothing to commit?)")
		os.Exit(1)
	}

	branch := getCurrentBranch()
	printStatus("Pushing to " + branch + "...")

	printCommand("git push")
	if gumSpin("Pushing...", "git", "push") {
		printStatus("󱝂 Magic complete!")
	} else {
		printWarning("Standard push failed. Trying to set upstream...")
		printCommand(`git push -u origin "` + branch + `"`)
		if gumSpin("Pushing (upstream)...", "git", "push", "-u", "origin", branch) {
			printStatus("󱝂 Magic complete (upstream set)!")
		} else {
			printError("Push failed.")
			os.Exit(1)
		}
	}
}

func summonGitignore() {
	if !hasCmd("curl") {
		printError("Error: curl is not installed. Required to summon the Architect.")
		return
	}

	printStatus("📜 Summoning the Architect...")

	if _, err := os.Stat(".gitignore"); !os.IsNotExist(err) {
		printWarning("A .gitignore file already exists in this directory.")
		if !gumConfirm("Do you want to overwrite it?") {
			printStatus("Operation cancelled.")
			return
		}
	}

	templates := []string{
		"Actionscript", "Ada", "Android", "Angular", "AppEngine", "ArchLinuxPackages", "Autotools",
		"C++", "C", "CMake", "CUDA", "CakePHP", "ChefCookbook", "Clojure", "CodeIgniter", "Composer",
		"Dart", "Delphi", "Dotnet", "Drupal", "Elixir", "Elm", "Erlang", "Flutter", "Fortran",
		"Go", "Godot", "Gradle", "Grails", "Haskell", "Haxe", "Java", "Jekyll", "Joomla", "Julia",
		"Kotlin", "Laravel", "Lua", "Magento", "Maven", "Nextjs", "Nim", "Nix", "Node", "Objective-C",
		"Opa", "Perl", "Phalcon", "PlayFramework", "Prestashop", "Processing", "Python", "Qt",
		"R", "ROS", "Rails", "Ruby", "Rust", "Scala", "Scheme", "Smalltalk", "Swift", "Symfony",
		"Terraform", "TeX", "Unity", "UnrealEngine", "VisualStudio", "WordPress", "Zig",
	}

	template := gumFilterStdin(strings.Join(templates, "\n"), "Search for a blueprint (e.g., Python, Node, Rust)...", false)

	if template == "" {
		printStatus("Cancelled.")
		return
	}

	printStatus("Drawing the magic circle for " + template + "...")
	url := "https://raw.githubusercontent.com/github/gitignore/main/" + template + ".gitignore"

	printCommand(`curl -sL "` + url + `" -o .gitignore`)
	if gumSpin("Fetching template...", "curl", "-sL", url, "-o", ".gitignore") {
		b, _ := os.ReadFile(".gitignore")
		if strings.Contains(string(b), "404: Not Found") {
			printError("Failed to summon template: 404 Not Found at " + url)
			os.Remove(".gitignore")
			return
		}
		printStatus("✨ .gitignore for " + template + " created successfully!")
	} else {
		printError("Failed to summon template. Check your internet connection.")
	}
}

func resurrectCommit() {
	printStatus("󰔪  Summoning the Resurrection Stone...")

	reflog := getGitOutput("reflog")
	if reflog == "" {
		printError("The ancient logs are empty. Nothing to resurrect.")
		return
	}

	fmt.Println("Search the ancient timelines for your lost code:")
	formattedReflog := getGitOutput("reflog", "--date=relative", "--format=%h %gd %C(blue)%ad%Creset %s")
	selectedLog := gumFilterStdin(formattedReflog, "Search for a lost commit message or hash...", false)

	if selectedLog == "" {
		printStatus("Resurrection cancelled.")
		return
	}

	targetHash := strings.Fields(selectedLog)[0]

	fmt.Println()
	cmd := exec.Command("gum", "style", "--border", "rounded", "--border-foreground", theme["BIMAGIC_PRIMARY"], "--padding", "1 2", "TARGET TIMELINE:", selectedLog)
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println()

	action := gumChoose("", "", "", "󰔱 Create a new branch here (Safest)", "  Hard Reset current branch to here (Dangerous)", "Cancel")

	switch action {
	case "󰔱 Create a new branch here (Safest)":
		newBranch := gumInput("Enter new branch name (e.g., recovered-code)", "")
		if newBranch != "" {
			printCommand(`git checkout -b "` + newBranch + `" "` + targetHash + `"`)
			runGitCmd("checkout", "-b", newBranch, targetHash)
			printStatus("󱝁 Timeline restored! You are now on branch: " + newBranch)
		}
	case "  Hard Reset current branch to here (Dangerous)":
		if gumConfirm("This will overwrite your CURRENT work. Are you absolutely sure?") {
			printCommand("git reset --hard " + targetHash)
			runGitCmd("reset", "--hard", targetHash)
			printStatus("󱝁 Timeline restored via hard reset!")
		}
	default:
		printStatus("The stone goes dormant.")
	}
}

func revertCommits() {
	printStatus("Fetching commit history...")
	fmt.Println()

	history := getGitOutput("log", "--oneline", "--decorate")
	commitsSelection := gumFilterStdin(history, "Select commit(s) to revert", true)

	if commitsSelection == "" {
		printWarning("No commit selected. Revert cancelled.")
		return
	}

	var hashes []string
	for _, line := range strings.Split(commitsSelection, "\n") {
		if line != "" {
			hashes = append(hashes, strings.Fields(line)[0])
		}
	}

	fmt.Println("You selected:")
	fmt.Println(strings.Join(hashes, "\n"))
	fmt.Println()

	if gumConfirm("Confirm revert?") {
		for _, c := range hashes {
			fmt.Printf("Reverting commit %s...\n", c)
			printCommand(fmt.Sprintf("git revert --no-edit %s", c))
			if runGitCmd("revert", "--no-edit", c) == nil {
				printStatus(fmt.Sprintf("Commit %s reverted.", c))
			} else {
				printError(fmt.Sprintf("Conflict occurred while reverting %s!", c))
				fmt.Println("Please resolve conflicts, then run:")
				fmt.Println("  git revert --continue")
				break
			}
		}
	} else {
		printStatus("Revert cancelled.")
	}
}

func scryingGlass() {
	printStatus("󰈈 Summoning the Scrying Glass...")

	if hasCmd("fzf") {
		previewCmd := "cat {}"
		if hasCmd("bat") {
			previewCmd = "bat --color=always --style=numbers {}"
		}

		gitOut := getGitOutput("ls-files", "--cached", "--others", "--exclude-standard")

		fzfCmd := exec.Command("fzf",
			"--preview", previewCmd,
			"--preview-window=right:60%",
			"--height=80%",
			"--layout=reverse",
			"--border",
			"--cycle",
			"--prompt=󰈈 Peer into: ",
			fmt.Sprintf("--color=bg+:-1,fg+:%s,hl:%s,hl+:%s,prompt:%s,pointer:%s,marker:%s,header:%s,spinner:%s,info:%s",
				theme["BIMAGIC_PRIMARY"], theme["BIMAGIC_SECONDARY"], theme["BIMAGIC_SECONDARY"],
				theme["BIMAGIC_INFO"], theme["BIMAGIC_PRIMARY"], theme["BIMAGIC_SUCCESS"],
				theme["BIMAGIC_PRIMARY"], theme["BIMAGIC_PRIMARY"], theme["BIMAGIC_MUTED"]),
		)
		fzfCmd.Stdin = strings.NewReader(gitOut)
		fzfCmd.Stderr = os.Stderr
		out, _ := fzfCmd.Output()
		file := strings.TrimSpace(string(out))

		if file == "" {
			printStatus("The glass goes dark (Cancelled).")
			return
		}

		if gumConfirm("Open in full pager?") {
			if hasCmd("bat") {
				batOut, _ := exec.Command("bat", "--color=always", file).Output()
				pager := exec.Command("gum", "pager")
				pager.Stdin = bytes.NewReader(batOut)
				pager.Stdout = os.Stdout
				pager.Run()
			} else {
				pager := exec.Command("gum", "pager")
				fileCont, _ := os.ReadFile(file)
				pager.Stdin = bytes.NewReader(fileCont)
				pager.Stdout = os.Stdout
				pager.Run()
			}
		}
	} else {
		gitOut := getGitOutput("ls-files", "--cached", "--others", "--exclude-standard")
		file := gumFilterStdin(gitOut, "Select a file to peer into...", false)

		if file == "" {
			printStatus("The glass goes dark (Cancelled).")
			return
		}

		printStatus("Peering into: " + file)
		if hasCmd("bat") {
			batOut, _ := exec.Command("bat", "--color=always", file).Output()
			pager := exec.Command("gum", "pager")
			pager.Stdin = bytes.NewReader(batOut)
			pager.Stdout = os.Stdout
			pager.Run()
		} else {
			pager := exec.Command("gum", "pager")
			fileCont, _ := os.ReadFile(file)
			pager.Stdin = bytes.NewReader(fileCont)
			pager.Stdout = os.Stdout
			pager.Run()
		}
	}
}

type authorStat struct {
	lines   int
	commits int
}

func showContributorStats() {
	if !isGitRepo() {
		printError("Not a git repository!")
		return
	}

	timeRange := gumChoose("Select time range", "", "", "Last 7 days", "Last 30 days", "Last 90 days", "Last year", "All time")
	since := ""
	switch timeRange {
	case "Last 7 days":
		since = "--since=7 days ago"
	case "Last 30 days":
		since = "--since=30 days ago"
	case "Last 90 days":
		since = "--since=3 months ago"
	case "Last year":
		since = "--since=1 year ago"
	case "All time":
		since = ""
	default:
		printWarning("No time range selected.")
		return
	}

	printStatus(fmt.Sprintf("Analyzing contributions (%s)...", timeRange))
	fmt.Println()

	args := []string{"log", "--pretty=format:COMMIT|%aN", "--numstat"}
	if since != "" {
		args = append(args, since)
	}

	cmd := exec.Command("git", args...)
	outBytes, err := cmd.Output()
	if err != nil || len(outBytes) == 0 {
		printError("No contribution data found for the selected period.")
		return
	}

	stats := make(map[string]*authorStat)
	totalLines := 0
	currentAuthor := ""

	scanner := bufio.NewScanner(bytes.NewReader(outBytes))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "COMMIT|") {
			parts := strings.SplitN(line, "|", 2)
			if len(parts) == 2 {
				currentAuthor = parts[1]
				if stats[currentAuthor] == nil {
					stats[currentAuthor] = &authorStat{}
				}
				stats[currentAuthor].commits++
			}
		} else if len(line) > 0 && (line[0] >= '0' && line[0] <= '9' || line[0] == '-') {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				added, deleted := 0, 0
				if parts[0] != "-" {
					added, _ = strconv.Atoi(parts[0])
				}
				if parts[1] != "-" {
					deleted, _ = strconv.Atoi(parts[1])
				}
				lines := added + deleted
				if currentAuthor != "" {
					stats[currentAuthor].lines += lines
					totalLines += lines
				}
			}
		}
	}

	if len(stats) == 0 {
		printError("No contribution data found for the selected period.")
		return
	}

	type authorInfo struct {
		name       string
		lines      int
		commits    int
		percentage float64
	}

	var authors []authorInfo
	for name, st := range stats {
		pct := float64(0)
		if totalLines > 0 {
			pct = (float64(st.lines) / float64(totalLines)) * 100
		}
		authors = append(authors, authorInfo{
			name:       strings.TrimSpace(name),
			lines:      st.lines,
			commits:    st.commits,
			percentage: pct,
		})
	}

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].lines > authors[j].lines
	})

	purple := getAnsiEsc(theme["BIMAGIC_PRIMARY"])
	blue := getAnsiEsc(theme["BIMAGIC_SECONDARY"])
	yellow := getAnsiEsc(theme["BIMAGIC_WARNING"])
	cyan := getAnsiEsc(theme["BIMAGIC_INFO"])
	nc := "\033[0m"

	fmt.Printf("%sContribution Report (%s)%s\n", purple, timeRange, nc)
	fmt.Println(strings.Repeat("─", 45))

	mostActiveAuthor := ""
	mostCommits := 0
	mostProductiveAuthor := ""
	highestAvg := 0

	for _, a := range authors {
		bar := generateBar(a.percentage)
		fmt.Printf("%s%-15s%s %s %s%5.1f%%%s (%s%d lines%s)\n", blue, a.name, nc, bar, yellow, a.percentage, nc, cyan, a.lines, nc)

		if a.commits > mostCommits {
			mostCommits = a.commits
			mostActiveAuthor = a.name
		}

		if a.commits > 0 {
			avgLines := a.lines / a.commits
			if avgLines > highestAvg {
				highestAvg = avgLines
				mostProductiveAuthor = a.name
			}
		}
	}

	fmt.Println()
	fmt.Printf("%sHighlights:%s\n", cyan, nc)
	if mostActiveAuthor != "" {
		fmt.Printf("%sMost Active:%s %s%s%s (%s%d commits%s)\n", blue, nc, purple, mostActiveAuthor, nc, yellow, mostCommits, nc)
	}
	if mostProductiveAuthor != "" {
		fmt.Printf("%sMost Productive:%s %s%s%s (%s%d lines/commit%s)\n", blue, nc, purple, mostProductiveAuthor, nc, yellow, highestAvg, nc)
	}
	fmt.Printf("%sTotal Contributors:%s %s%d%s\n", blue, nc, yellow, len(authors), nc)
}
