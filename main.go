package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

var logo = strings.TrimLeft(`
████████╗██╗  ██╗ █████╗ ██╗  ██╗██╗  ██╗ █████╗ ██╗     ██╗
╚══██╔══╝██║  ██║██╔══██╗██║ ██╔╝██║ ██╔╝██╔══██╗██║     ██║
   ██║   ███████║███████║█████╔╝ █████╔╝ ███████║██║     ██║
   ██║   ██╔══██║██╔══██║██╔═██╗ ██╔═██╗ ██╔══██║██║     ██║
   ██║   ██║  ██║██║  ██║██║  ██╗██║  ██╗██║  ██║███████╗██║
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝
`, "\n")

const (
	tomatoRX   = 13.0
	tomatoRY   = 6.5
	stemHeight = 2

	numIdleVariants = 8  // outer ring shimmer cycles
	numSpinFrames   = 24 // full 360° spin

	animInterval  = 70 * time.Millisecond
	spinCooldown  = 120 // ~8.4s between tricks
	spinSpeed     = 2   // anim ticks per spin frame
	blinkPeriod   = 20  // ~1.4s between blinks
	blinkHold     = 3   // blink duration
	maxJumpAmp    = 4   // max rows of jump during spin
	idleBounceAmp = 1   // continuous bounce amplitude
)

var (
	spriteW = int(math.Round(tomatoRX*2)) + 3
	spriteH = int(math.Round(tomatoRY*2)) + 1 + stemHeight
	stripH  = spriteH + maxJumpAmp
)

// Ghostty-style layered palette, in tomato tones.
var (
	colBrightRed = lipgloss.Color("#FF5252")
	colDarkRed   = lipgloss.Color("#7F1212")
	colEye       = lipgloss.Color("#1A0000")
	colStem      = lipgloss.Color("#43A047")
	colDim       = lipgloss.Color("240")
	colRed       = lipgloss.Color("#E53935")
	colGreen     = lipgloss.Color("#43A047")

	logoStyle = lipgloss.NewStyle().Foreground(colRed).Bold(true)
	timerBox  = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colRed).
			Padding(1, 3).
			Align(lipgloss.Center)
	timeStyle   = lipgloss.NewStyle().Foreground(colRed).Bold(true)
	statusStyle = lipgloss.NewStyle().Foreground(colDim).Italic(true)
	doneStyle   = lipgloss.NewStyle().Foreground(colGreen).Bold(true)
	helpStyle   = lipgloss.NewStyle().Foreground(colDim)

	stemStyle  = lipgloss.NewStyle().Foreground(colStem).Bold(true)
	outerStyle = lipgloss.NewStyle().Foreground(colDarkRed).Bold(true)
	innerStyle = lipgloss.NewStyle().Foreground(colBrightRed).Bold(true)
	eyeStyle   = lipgloss.NewStyle().Foreground(colEye).Bold(true)

	workLabelStyle  = lipgloss.NewStyle().Foreground(colBrightRed).Bold(true)
	breakLabelStyle = lipgloss.NewStyle().Foreground(colGreen).Bold(true)
)

type cellClass uint8

const (
	clsEmpty cellClass = iota
	clsStem
	clsOuter
	clsInner
	clsEye
)

type cell struct {
	ch  rune
	cls cellClass
}

// Pre-baked frames.
var idleFrames [][][]cell // shimmer-only variants, static body + upright stem
var spinFrames [][][]cell // full 360° rotation (body shading + stem tilt)

func buildFrames() {
	idleFrames = make([][][]cell, numIdleVariants)
	for i := 0; i < numIdleVariants; i++ {
		idleFrames[i] = buildTomato(0, i, false)
	}
	spinFrames = make([][][]cell, numSpinFrames)
	for f := 0; f < numSpinFrames; f++ {
		angle := 2 * math.Pi * float64(f) / float64(numSpinFrames)
		// During spin, shimmer also ticks through its variants.
		spinFrames[f] = buildTomato(angle, f, true)
	}
}

// buildTomato builds one frame.
//   - angle: body rotation (affects inner highlight + stem tilt) — only used
//     when spinning is true; otherwise treat as 0.
//   - shimmerOffset: rotates outer-ring character pattern.
//   - spinning: if true, inner body shows rotating highlight; if false, body
//     has a fixed highlight (top-left) for depth only.
func buildTomato(angle float64, shimmerOffset int, spinning bool) [][]cell {
	grid := make([][]cell, spriteH)
	for i := range grid {
		grid[i] = make([]cell, spriteW)
	}

	cx := float64(spriteW) / 2.0
	cy := float64(stemHeight) + tomatoRY + 0.5

	// Highlight position — rotating during spin, fixed top-left in idle.
	var hx, hy float64
	if spinning {
		hx = tomatoRX * 0.4 * math.Cos(angle)
		hy = tomatoRY * 0.4 * math.Sin(angle)
	} else {
		hx = -tomatoRX * 0.25
		hy = -tomatoRY * 0.35
	}

	// Ghostty-style density palette. Inner body uses the densest chars
	// (@$%), outer halo uses medium-density (x=+*%$@) like Ghostty's b-class.
	innerRamp := []rune("@@@$%")
	ringChars := []rune("x=+*%$@=")

	for y := stemHeight; y < spriteH; y++ {
		for x := 0; x < spriteW; x++ {
			px := float64(x) - cx + 0.5
			py := float64(y) - cy + 0.5

			nx := px / tomatoRX
			ny := py / tomatoRY
			d := math.Sqrt(nx*nx + ny*ny)

			switch {
			case d > 0.99:
				// outside body
			case d > 0.82:
				// Outer ring with rotating shimmer pattern.
				angAround := math.Atan2(py, px)
				idx := int((angAround+math.Pi)/(2*math.Pi)*float64(len(ringChars))) + shimmerOffset
				idx = ((idx % len(ringChars)) + len(ringChars)) % len(ringChars)
				grid[y][x] = cell{ringChars[idx], clsOuter}
			case d > 0.74:
				// gap
			default:
				// Eyes stay upright — fixed in body frame. Empty (space)
				// char gives a crisp hole like Ghostty's eyes.
				if inEye(px, py, false) {
					grid[y][x] = cell{' ', clsEye}
					continue
				}
				hd := math.Sqrt(math.Pow(px-hx, 2) + math.Pow((py-hy)*2, 2))
				maxD := 2 * tomatoRY
				t := 1.0 - hd/maxD
				if t < 0 {
					t = 0
				}
				if t > 1 {
					t = 1
				}
				idx := int(t * float64(len(innerRamp)-1))
				if idx < 0 {
					idx = 0
				}
				if idx >= len(innerRamp) {
					idx = len(innerRamp) - 1
				}
				grid[y][x] = cell{innerRamp[idx], clsInner}
			}
		}
	}

	drawStem(grid, angle, int(math.Round(cx)), spinning)
	return grid
}

func inEye(px, py float64, blink bool) bool {
	const eyeY = -1.5
	const eyeDX = 4.5
	check := func(ecx float64) bool {
		dx := px - ecx
		dy := py - eyeY
		if blink {
			return math.Abs(dy) < 0.55 && math.Abs(dx) < 1.8
		}
		return (dx*dx)/3.0+(dy*dy)*0.6 < 1.0
	}
	return check(-eyeDX) || check(eyeDX)
}

func drawStem(grid [][]cell, angle float64, cx int, spinning bool) {
	set := func(y, x int, ch rune) {
		if y >= 0 && y < spriteH && x >= 0 && x < spriteW {
			grid[y][x] = cell{ch, clsStem}
		}
	}
	set(1, cx-1, '\\')
	set(1, cx, '|')
	set(1, cx+1, '/')

	lean := 0.0
	if spinning {
		lean = math.Cos(angle)
	}
	var pos int
	var leaf rune
	switch {
	case lean > 0.7:
		pos, leaf = cx+2, '/'
	case lean > 0.3:
		pos, leaf = cx+1, '/'
	case lean > -0.3:
		pos, leaf = cx, '|'
	case lean > -0.7:
		pos, leaf = cx-1, '\\'
	default:
		pos, leaf = cx-2, '\\'
	}
	set(0, pos, leaf)
}

type tickMsg time.Time
type animMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func animTick() tea.Cmd {
	return tea.Tick(animInterval, func(t time.Time) tea.Msg { return animMsg(t) })
}

type animState uint8

const (
	stateIdle animState = iota
	stateSpin
)

type phase uint8

const (
	phaseWork phase = iota
	phaseShort
	phaseLong
	phaseTimer
)

func (p phase) String() string {
	switch p {
	case phaseWork:
		return "work"
	case phaseShort:
		return "short_break"
	case phaseLong:
		return "long_break"
	case phaseTimer:
		return "timer"
	}
	return "unknown"
}

type config struct {
	Work   int    `json:"work"`
	Short  int    `json:"short"`
	Long   int    `json:"long"`
	Rounds int    `json:"rounds"`
	Sound  string `json:"sound"`
}

var defaultConfig = config{Work: 25, Short: 5, Long: 15, Rounds: 4, Sound: ""}

// playSound plays the given notification sound.
//   - "" or "default" / "beep" → cross-platform beeper
//   - macOS: resolved as /System/Library/Sounds/<name>.aiff (or absolute path)
//   - other OSes: always falls back to beep
func playSound(name string) {
	if name == "" || name == "default" || name == "beep" {
		_ = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		return
	}
	if runtime.GOOS == "darwin" {
		path := name
		if !strings.HasPrefix(name, "/") {
			path = "/System/Library/Sounds/" + name + ".aiff"
		}
		cmd := exec.Command("afplay", path)
		if err := cmd.Start(); err != nil {
			_ = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
			return
		}
		go cmd.Wait()
		return
	}
	_ = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
}

func thakkaliDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "thakkali"), nil
}

func loadConfig() config {
	cfg := defaultConfig
	dir, err := thakkaliDir()
	if err != nil {
		return cfg
	}
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		// Create default config on first run.
		_ = os.MkdirAll(dir, 0o755)
		if b, e := json.MarshalIndent(cfg, "", "  "); e == nil {
			_ = os.WriteFile(path, b, 0o644)
		}
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

type logEntry struct {
	Timestamp   string `json:"timestamp"`
	Phase       string `json:"phase"`
	DurationSec int    `json:"duration_sec"`
	Task        string `json:"task,omitempty"`
	TaskID      *int   `json:"task_id,omitempty"`
	TaskPrefix  string `json:"task_prefix,omitempty"`
	Project     string `json:"project,omitempty"`
}

func appendLog(e logEntry) {
	dir, err := thakkaliDir()
	if err != nil {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, "log.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	_, _ = f.Write(append(b, '\n'))
}

type model struct {
	durations [3]time.Duration
	rounds    int
	task       string // display title (resolved from TSK-N or free text)
	taskID     *int   // non-nil when -task TSK-N resolved to a tracked task
	taskPrefix string // ID prefix for the tracked task ("TSK", "THA", ...)
	project    string // project tag carried from the tracked task
	sound     string
	pomodoro  bool // full cycle: work + short/long breaks across rounds
	minimal   bool // hide logo + tomato animation
	countUp   bool // stopwatch mode — count up from 0 rather than down
	hideHelp  bool // hide footer hint lines

	remaining time.Duration
	running   bool
	done      bool // only used in simple-timer (non-pomodoro) mode
	phase     phase
	round     int // current work round, 1..rounds

	elapsed   time.Duration // count-up elapsed time (timer mode)
	target    time.Duration // optional soft goal in timer mode; 0 = none
	targetHit bool          // latched once elapsed crosses target

	width   int
	spriteX int

	animN     int
	state     animState
	stateN    int
	spinFrame int
}

// taskDisplay formats the session task label. When tied to a tracked task,
// prefix with the PREFIX-N id and append the project badge.
func (m *model) taskDisplay() string {
	if m.taskID == nil {
		return m.task
	}
	p := m.taskPrefix
	if p == "" {
		p = "TSK"
	}
	s := fmt.Sprintf("%s-%d %s", p, *m.taskID, m.task)
	if m.project != "" {
		s += " #" + m.project
	}
	return s
}

func (m *model) advance() {
	from := m.phase
	fromDur := m.durations[from]
	fromTask := m.task

	entry := logEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Phase:       from.String(),
		DurationSec: int(fromDur.Seconds()),
	}
	if from == phaseWork {
		entry.Task = fromTask
		entry.TaskID = m.taskID
		entry.TaskPrefix = m.taskPrefix
		entry.Project = m.project
	}
	go appendLog(entry)

	if !m.pomodoro {
		m.remaining = 0
		m.running = false
		m.done = true
		sound := m.sound
		go func() {
			_ = beeep.Notify("Thakkali", "Timer complete!", "")
			playSound(sound)
		}()
		return
	}


	switch m.phase {
	case phaseWork:
		if m.round >= m.rounds {
			m.phase = phaseLong
		} else {
			m.phase = phaseShort
		}
	case phaseShort:
		m.phase = phaseWork
		m.round++
	case phaseLong:
		m.phase = phaseWork
		m.round = 1
	}
	m.remaining = m.durations[m.phase]
	m.running = true
	go notifyTransition(from, m.phase, m.sound)
}

// logTimerSession writes a stopwatch session to log.jsonl. Called on quit and
// on reset so in-progress tracking isn't silently discarded.
func (m *model) logTimerSession() {
	if !m.countUp || m.elapsed < time.Second {
		return
	}
	appendLog(logEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Phase:       phaseTimer.String(),
		DurationSec: int(m.elapsed.Seconds()),
		Task:        m.task,
		TaskID:      m.taskID,
		TaskPrefix:  m.taskPrefix,
		Project:     m.project,
	})
}

// switchMode reconfigures the live model into countdown / pomodoro / timer without
// quitting. If we're leaving an in-progress timer session, flush it to the log
// first so tracked time isn't silently dropped.
func (m *model) switchMode(pomodoro, countUp bool) {
	if m.countUp {
		m.logTimerSession()
	}
	m.pomodoro = pomodoro
	m.countUp = countUp
	m.elapsed = 0
	m.targetHit = false
	m.done = false
	m.round = 1
	if countUp {
		m.phase = phaseTimer
		m.remaining = 0
	} else {
		m.phase = phaseWork
		m.remaining = m.durations[phaseWork]
	}
	m.running = true
}

func notifyTransition(from, to phase, sound string) {
	var msg string
	switch to {
	case phaseWork:
		msg = "Break's over — back to work!"
	case phaseShort:
		msg = "Work complete — take a short break!"
	case phaseLong:
		msg = "All rounds done — enjoy a long break!"
	}
	_ = from
	_ = beeep.Notify("Thakkali", msg, "")
	playSound(sound)
}

func (m model) phaseLabel() string {
	switch m.phase {
	case phaseWork:
		return fmt.Sprintf("work  %d / %d", m.round, m.rounds)
	case phaseShort:
		return "short break"
	case phaseLong:
		return "long break"
	}
	return ""
}

func (m model) phaseStyle() lipgloss.Style {
	if m.phase == phaseWork {
		return workLabelStyle
	}
	return breakLabelStyle
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), animTick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.logTimerSession()
			return m, tea.Quit
		case " ":
			m.running = !m.running
		case "r":
			if m.countUp {
				// Preserve the in-progress session so an accidental reset
				// doesn't silently drop tracked time.
				m.logTimerSession()
				m.elapsed = 0
				m.targetHit = false
				m.running = true
			} else {
				m.remaining = m.durations[m.phase]
				m.running = true
				m.done = false
			}
		case "s":
			if m.pomodoro {
				m.advance()
			}
		case "m":
			m.minimal = !m.minimal
		case "h":
			m.hideHelp = !m.hideHelp
		case "1":
			m.switchMode(false, false)
		case "2":
			m.switchMode(true, false)
		case "3":
			m.switchMode(false, true)
		case "+", "=":
			if m.countUp {
				if m.target < 24*time.Hour {
					if m.target == 0 {
						m.target = time.Minute
					} else {
						m.target += time.Minute
					}
					if m.elapsed < m.target {
						m.targetHit = false
					}
				}
			} else if m.durations[m.phase] < 120*time.Minute {
				m.durations[m.phase] += time.Minute
				m.remaining = m.durations[m.phase]
			}
		case "-", "_":
			if m.countUp {
				if m.target > time.Minute {
					m.target -= time.Minute
				} else {
					m.target = 0
					m.targetHit = false
				}
			} else if m.durations[m.phase] > time.Minute {
				m.durations[m.phase] -= time.Minute
				m.remaining = m.durations[m.phase]
			}
		}
	case tickMsg:
		if m.running {
			if m.countUp {
				m.elapsed += time.Second
				if m.target > 0 && !m.targetHit && m.elapsed >= m.target {
					m.targetHit = true
					sound := m.sound
					go func() {
						_ = beeep.Notify("Thakkali", "Target reached — keep going or wrap up.", "")
						playSound(sound)
					}()
				}
			} else if !m.done {
				m.remaining -= time.Second
				if m.remaining <= 0 {
					m.advance()
				}
			}
		}
		return m, tick()
	case animMsg:
		m.animN++
		m.stateN++

		w := m.width
		if w <= 0 {
			w = 80
		}

		switch m.state {
		case stateIdle:
			// Rolling horizontally while idle.
			m.spriteX = (m.spriteX + 1) % (w + spriteW)
			if m.stateN >= spinCooldown {
				m.state = stateSpin
				m.stateN = 0
				m.spinFrame = 0
			}
		case stateSpin:
			// Pause horizontal motion — tomato spins in place.
			if m.stateN%spinSpeed == 0 {
				m.spinFrame++
				if m.spinFrame >= numSpinFrames {
					m.state = stateIdle
					m.stateN = 0
					m.spinFrame = 0
				}
			}
		}
		return m, animTick()
	}
	return m, nil
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	if total >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", total/3600, (total%3600)/60, total%60)
	}
	return fmt.Sprintf("%02d:%02d", total/60, total%60)
}

// ANSI Shadow digit glyphs — same font family as the THAKKALI logo.
var digitGlyphs = map[rune][]string{
	'0': {
		` ██████╗ `,
		`██╔═████╗`,
		`██║██╔██║`,
		`████╔╝██║`,
		`╚██████╔╝`,
		` ╚═════╝ `,
	},
	'1': {
		` ██╗`,
		`███║`,
		`╚██║`,
		` ██║`,
		` ██║`,
		` ╚═╝`,
	},
	'2': {
		`██████╗ `,
		`╚════██╗`,
		` █████╔╝`,
		`██╔═══╝ `,
		`███████╗`,
		`╚══════╝`,
	},
	'3': {
		`██████╗ `,
		`╚════██╗`,
		` █████╔╝`,
		` ╚═══██╗`,
		`██████╔╝`,
		`╚═════╝ `,
	},
	'4': {
		`██╗  ██╗`,
		`██║  ██║`,
		`███████║`,
		`╚════██║`,
		`     ██║`,
		`     ╚═╝`,
	},
	'5': {
		`███████╗`,
		`██╔════╝`,
		`███████╗`,
		`╚════██║`,
		`███████║`,
		`╚══════╝`,
	},
	'6': {
		` ██████╗ `,
		`██╔════╝ `,
		`███████╗ `,
		`██╔═══██╗`,
		`╚██████╔╝`,
		` ╚═════╝ `,
	},
	'7': {
		`███████╗`,
		`╚════██║`,
		`    ██╔╝`,
		`   ██╔╝ `,
		`   ██║  `,
		`   ╚═╝  `,
	},
	'8': {
		` █████╗ `,
		`██╔══██╗`,
		`╚█████╔╝`,
		`██╔══██╗`,
		`╚█████╔╝`,
		` ╚════╝ `,
	},
	'9': {
		` █████╗ `,
		`██╔══██╗`,
		`╚██████║`,
		` ╚═══██║`,
		` █████╔╝`,
		` ╚════╝ `,
	},
	':': {
		`   `,
		`██╗`,
		`╚═╝`,
		`██╗`,
		`╚═╝`,
		`   `,
	},
}

func renderBigTime(text string) string {
	const glyphH = 6
	rows := make([]string, glyphH)
	for _, ch := range text {
		g, ok := digitGlyphs[ch]
		if !ok {
			continue
		}
		for i := 0; i < glyphH; i++ {
			rows[i] += g[i] + " "
		}
	}
	return strings.Join(rows, "\n")
}

func styleFor(cls cellClass) lipgloss.Style {
	switch cls {
	case clsStem:
		return stemStyle
	case clsOuter:
		return outerStyle
	case clsInner:
		return innerStyle
	case clsEye:
		return eyeStyle
	}
	return lipgloss.NewStyle()
}

// jumpOffset returns vertical offset in rows (0..maxJumpAmp).
func (m model) jumpOffset() int {
	// Continuous gentle bounce.
	base := float64(idleBounceAmp) * math.Abs(math.Sin(float64(m.animN)*0.18))

	// Big parabolic jump during spin.
	if m.state == stateSpin {
		phase := float64(m.spinFrame) / float64(numSpinFrames)
		arc := 4 * phase * (1 - phase) // 0..1..0
		return int(math.Round(base + float64(maxJumpAmp)*arc))
	}
	return int(math.Round(base))
}

func (m model) currentFrame() [][]cell {
	if m.state == stateSpin {
		return spinFrames[m.spinFrame%numSpinFrames]
	}
	// Shimmer cycles slowly in idle.
	idx := (m.animN / 2) % numIdleVariants
	return idleFrames[idx]
}

// applyBlink overlays closed-eye cells onto the frame when blinking.
func applyBlink(frame [][]cell) [][]cell {
	// Shallow-copy rows we'll modify; this frame is used read-only otherwise.
	out := make([][]cell, len(frame))
	for i, row := range frame {
		out[i] = row
	}
	cx := float64(spriteW) / 2.0
	cy := float64(stemHeight) + tomatoRY + 0.5
	for y := stemHeight; y < spriteH; y++ {
		rowCopy := make([]cell, len(out[y]))
		copy(rowCopy, out[y])
		modified := false
		for x := 0; x < spriteW; x++ {
			if rowCopy[x].cls != clsEye && rowCopy[x].cls != clsInner {
				continue
			}
			px := float64(x) - cx + 0.5
			py := float64(y) - cy + 0.5
			if inEye(px, py, true) {
				rowCopy[x] = cell{'-', clsEye}
				modified = true
			} else if rowCopy[x].cls == clsEye {
				// Eye cell that's not in blink slit — revert to body.
				rowCopy[x] = cell{'@', clsInner}
				modified = true
			}
		}
		if modified {
			out[y] = rowCopy
		}
	}
	return out
}

func (m model) renderSprite() string {
	w := m.width
	if w <= 0 {
		w = 80
	}

	frame := m.currentFrame()
	if m.animN%blinkPeriod < blinkHold {
		frame = applyBlink(frame)
	}

	rows := make([][]cell, stripH)
	for i := range rows {
		rows[i] = make([]cell, w)
		for j := range rows[i] {
			rows[i][j] = cell{' ', clsEmpty}
		}
	}

	// yBase: row where sprite top lands. Higher jumpOffset = sprite higher
	// up = smaller yBase.
	yBase := maxJumpAmp - m.jumpOffset()

	stamp := func(startX int) {
		for r, line := range frame {
			destY := yBase + r
			if destY < 0 || destY >= stripH {
				continue
			}
			for i, c := range line {
				x := startX + i
				if x < 0 || x >= w {
					continue
				}
				if c.cls != clsEmpty {
					rows[destY][x] = c
				}
			}
		}
	}
	stamp(m.spriteX)
	stamp(m.spriteX - (w + spriteW))

	var out strings.Builder
	for r, row := range rows {
		i := 0
		for i < len(row) {
			j := i
			for j < len(row) && row[j].cls == row[i].cls {
				j++
			}
			var seg strings.Builder
			for k := i; k < j; k++ {
				seg.WriteRune(row[k].ch)
			}
			if row[i].cls == clsEmpty {
				out.WriteString(seg.String())
			} else {
				out.WriteString(styleFor(row[i].cls).Render(seg.String()))
			}
			i = j
		}
		if r < len(rows)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func (m model) View() string {
	var status string
	switch {
	case m.done:
		status = doneStyle.Render("✔ complete")
	case m.running:
		status = statusStyle.Render("● running")
	default:
		status = statusStyle.Render("‖ paused")
	}

	var display time.Duration
	if m.countUp {
		display = m.elapsed
	} else {
		display = m.remaining
	}
	timeRender := timeStyle
	if m.countUp && m.targetHit {
		timeRender = doneStyle
	}
	timer := timeRender.Render(renderBigTime(formatDuration(display)))

	var helpText string
	switch {
	case m.countUp:
		helpText = "space — pause/resume · r — reset · m — minimal · h — toggle help · +/− — target · q — save & quit"
	case m.pomodoro:
		helpText = "space — pause/resume · r — reset · s — skip · m — minimal · h — toggle help · +/− — adjust · q — quit"
	default:
		helpText = "space — pause/resume · r — reset · m — minimal · h — toggle help · +/− — adjust · q — quit"
	}
	modeHelp := "1 — countdown · 2 — pomodoro · 3 — timer"
	help := helpStyle.Copy().Align(lipgloss.Center).Render(helpText + "\n" + modeHelp)

	sections := []string{""}
	if !m.minimal {
		sections = append(sections,
			logoStyle.Render(logo),
			"",
			m.renderSprite(),
			"",
		)
	}
	switch {
	case m.countUp:
		label := "timer"
		if m.target > 0 {
			mark := "target"
			if m.targetHit {
				mark = "target ✔"
			}
			label = fmt.Sprintf("timer · %s %s", mark, formatDuration(m.target))
		}
		labelStyle := workLabelStyle
		if m.targetHit {
			labelStyle = breakLabelStyle
		}
		sections = append(sections, labelStyle.Render(label))
		if m.task != "" {
			sections = append(sections, statusStyle.Render("task: "+m.taskDisplay()))
		}
		sections = append(sections, "")
	case m.pomodoro:
		sections = append(sections, m.phaseStyle().Render(m.phaseLabel()))
		if m.task != "" && m.phase == phaseWork {
			sections = append(sections, statusStyle.Render("task: "+m.taskDisplay()))
		}
		sections = append(sections, "")
	default:
		if m.task != "" {
			sections = append(sections, statusStyle.Render("task: "+m.taskDisplay()), "")
		}
	}
	sections = append(sections,
		timer,
		"",
		status,
	)
	if !m.hideHelp {
		sections = append(sections, "", help)
	}

	return lipgloss.JoinVertical(lipgloss.Center, sections...)
}

func readLog() []logEntry {
	dir, err := thakkaliDir()
	if err != nil {
		return nil
	}
	f, err := os.Open(filepath.Join(dir, "log.jsonl"))
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []logEntry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		var e logEntry
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			out = append(out, e)
		}
	}
	return out
}

func fmtDur(sec int) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func bar(val, max, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := int(float64(val)/float64(max)*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// statsBanner returns a thick block-letter line spanning the logo width with
// " STATS " centered inside, so the section header reads as part of the logo
// rather than a tiny afterthought.
func statsBanner() string {
	return statsBannerWithLabel(" STATS ")
}

func printExamples() {
	header := lipgloss.NewStyle().Foreground(colRed).Bold(true)
	cmdStyle := lipgloss.NewStyle().Foreground(colBrightRed)
	dim := lipgloss.NewStyle().Foreground(colDim)

	section := func(title string, items [][2]string) {
		fmt.Println()
		fmt.Println(header.Render(title))
		fmt.Println()
		for _, it := range items {
			fmt.Printf("  %s\n", cmdStyle.Render(it[0]))
			fmt.Printf("    %s\n\n", dim.Render(it[1]))
		}
	}

	fmt.Println()
	fmt.Println(logoStyle.Render(logo))
	fmt.Println(logoStyle.Render(statsBannerWithLabel(" EXAMPLES ")))

	section("countdown mode (default)", [][2]string{
		{"thakkali", "25-minute countdown — the default, no flags needed"},
		{"thakkali -w 45", "45-minute countdown"},
		{"thakkali -w 50 -t \"deep work\"", "tag the session so it shows up under top tasks in stats"},
		{"thakkali -w 25 -t TSK-3", "tie the session to a tracked task — stats roll up per task and project"},
		{"thakkali -w 30 -m", "minimal mode — hide the logo and tomato animation"},
		{"thakkali -w 25 -S Glass", "use macOS Glass sound when the timer ends"},
	})

	section("pomodoro mode", [][2]string{
		{"thakkali -p", "classic Pomodoro — 25 / 5 / 15, four rounds"},
		{"thakkali -p -t \"ship phase 7\"", "Pomodoro with a task tag (work phases only)"},
		{"thakkali -p -w 50 -s 10 -l 20 -r 3", "longer work blocks, 10-min short break, 20-min long break, 3 rounds"},
		{"thakkali -p -w 1 -s 1 -l 1 -r 2", "smoke test — full cycle in ~5 minutes"},
		{"thakkali -p -m -S Hero", "minimal Pomodoro with Hero notification sound"},
	})

	section("timer / stopwatch mode", [][2]string{
		{"thakkali -T", "open-ended stopwatch — counts up until you quit"},
		{"thakkali -T -t \"code review\"", "stopwatch tagged with a task"},
		{"thakkali -T -target 45m -t \"debug prod\"", "soft 45-minute goal — beeps on reach, keeps running"},
		{"thakkali -T -target 1h30m -t \"design doc\"", "longer goals accept any time.ParseDuration string"},
		{"thakkali -T -m -t \"meeting\"", "minimal stopwatch for a quiet corner of your terminal"},
	})

	section("task management", [][2]string{
		{"thakkali task add \"Review auth PR\" -p auth -d 2026-05-01", "capture a new task with a project tag and planned due date"},
		{"thakkali task list", "show active tasks (doing + todo) from the task file in play"},
		{"thakkali task list -state all -p thakkali", "show every task scoped to the #thakkali project"},
		{"thakkali task move TSK-3 doing", "transition a task; @begin is auto-stamped on save"},
		{"thakkali task done TSK-3", "mark done; @done is auto-stamped"},
		{"thakkali task bulk", "open the task file in $EDITOR for free-form bulk capture"},
		{"thakkali task show TSK-3", "full details for one task"},
		{"thakkali todo", "interactive TUI: j/k move · space cycle · n new · e edit · d delete · / filter"},
		{"thakkali kanban", "three-column board: h/l switch column · </> move task · same n/e/d/space/?"},
		{"thakkali gantt -view week", "horizontal bars over a 14-day window for tasks with @start / @due"},
		{"thakkali gantt -view year", "yearly view — useful for big planned ranges"},
		{"thakkali activity", "GitHub-style 52-week heatmap of logged sessions from log.jsonl"},
	})

	section("stats", [][2]string{
		{"thakkali stats", "today + last 7 days, both Pomodoro and Timer sections"},
		{"thakkali stats -days 30", "wider window"},
		{"thakkali stats -mode pomodoro", "only Pomodoro / countdown sessions"},
		{"thakkali stats -m timer -days 14", "only stopwatch sessions, last 14 days (short flag)"},
	})

	fmt.Println(dim.Render("  in-app keys: space — pause/resume · r — reset · m — minimal · h — toggle help · +/− — adjust · 1/2/3 — mode · q — quit"))
	fmt.Println()
}

// statsBannerWithLabel is the generalized form of statsBanner — used by
// printExamples to produce a matching banner with a different label.
func statsBannerWithLabel(label string) string {
	width := lipgloss.Width(strings.Split(logo, "\n")[0])
	pad := width - lipgloss.Width(label)
	if pad < 2 {
		return label
	}
	left := pad / 2
	right := pad - left
	return strings.Repeat("█", left) + label + strings.Repeat("█", right)
}

func runStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	days := fs.Int("days", 7, "number of past days to include in the chart")
	mode := fs.String("mode", "all", "which sessions to show: all | pomodoro | timer")
	fs.StringVar(mode, "m", "all", "shorthand for -mode")
	_ = fs.Parse(args)

	switch *mode {
	case "all", "pomodoro", "timer":
	default:
		fmt.Fprintf(os.Stderr, "error: -mode must be all, pomodoro, or timer (got %q)\n", *mode)
		os.Exit(2)
	}

	entries := readLog()
	if len(entries) == 0 {
		fmt.Println("no sessions logged yet — run `thakkali` to get started.")
		return
	}

	fmt.Println()
	fmt.Println(logoStyle.Render(logo))
	fmt.Println(logoStyle.Render(statsBanner()))

	if *mode == "all" || *mode == "pomodoro" {
		renderStatsSection(entries, *days, "pomodoro", "work", map[string]bool{"work": true})
	}
	if *mode == "all" || *mode == "timer" {
		renderStatsSection(entries, *days, "timer", "timer", map[string]bool{"timer": true})
	}
}

// renderStatsSection prints one stats block (today / last-N-days / bar chart /
// top tasks / all-time) filtered to the given set of logEntry.Phase values.
// `title` is the section heading; `noun` is the word used in summary lines
// ("work" or "timer").
func renderStatsSection(entries []logEntry, days int, title, noun string, phases map[string]bool) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeStart := todayStart.AddDate(0, 0, -(days - 1))

	var (
		todayTotal    int
		todaySessions int
		rangeTotal    int
		allTimeTotal  int
	)
	perDay := make(map[string]int)
	tasks := make(map[string]int)
	projects := make(map[string]int)
	trackedSec := make(map[int]int)
	trackedSessions := make(map[int]int)
	trackedTitle := make(map[int]string)
	trackedPrefix := make(map[int]string)

	for _, e := range entries {
		if !phases[e.Phase] {
			continue
		}
		t, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			continue
		}
		tLocal := t.Local()
		allTimeTotal += e.DurationSec
		if tLocal.Before(rangeStart) {
			continue
		}
		dateKey := tLocal.Format("2006-01-02")
		perDay[dateKey] += e.DurationSec
		rangeTotal += e.DurationSec
		if !tLocal.Before(todayStart) {
			todayTotal += e.DurationSec
			todaySessions++
		}
		if e.Task != "" {
			label := e.Task
			if e.TaskID != nil {
				label = fmt.Sprintf("%s %s", logRef(e), e.Task)
			}
			tasks[label] += e.DurationSec
		}
		if e.Project != "" {
			projects["#"+e.Project] += e.DurationSec
		}
		if e.TaskID != nil {
			id := *e.TaskID
			trackedSec[id] += e.DurationSec
			trackedSessions[id]++
			if e.Task != "" {
				trackedTitle[id] = e.Task
			}
			if e.TaskPrefix != "" {
				trackedPrefix[id] = e.TaskPrefix
			}
		}
	}

	header := lipgloss.NewStyle().Foreground(colRed).Bold(true)
	subHeader := lipgloss.NewStyle().Foreground(colRed).Bold(true).Underline(true)
	dim := lipgloss.NewStyle().Foreground(colDim)
	accent := lipgloss.NewStyle().Foreground(colBrightRed)

	fmt.Println()
	fmt.Println(subHeader.Render(strings.ToUpper(title)))

	if allTimeTotal == 0 {
		fmt.Printf("  %s\n\n", dim.Render("no "+noun+" sessions logged yet."))
		return
	}

	fmt.Println()
	fmt.Println(header.Render("today"))
	fmt.Printf("  %-6s %s  %s\n",
		noun,
		accent.Render(fmtDur(todayTotal)),
		dim.Render(fmt.Sprintf("(%d sessions)", todaySessions)))
	fmt.Println()

	fmt.Println(header.Render(fmt.Sprintf("last %d days", days)))
	fmt.Printf("  total  %s\n", accent.Render(fmtDur(rangeTotal)))
	fmt.Println()

	maxSec := 0
	for _, v := range perDay {
		if v > maxSec {
			maxSec = v
		}
	}
	const barWidth = 24
	for i := 0; i < days; i++ {
		d := rangeStart.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		sec := perDay[key]
		label := d.Format("Mon Jan 02")
		marker := ""
		if key == todayStart.Format("2006-01-02") {
			marker = dim.Render(" ← today")
		}
		fmt.Printf("  %s  %s  %s%s\n",
			label,
			accent.Render(bar(sec, maxSec, barWidth)),
			fmtDur(sec),
			marker)
	}
	fmt.Println()

	if len(tasks) > 0 {
		fmt.Println(header.Render("top tasks"))
		type tk struct {
			name string
			sec  int
		}
		list := make([]tk, 0, len(tasks))
		for k, v := range tasks {
			list = append(list, tk{k, v})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].sec > list[j].sec })
		limit := 5
		if len(list) < limit {
			limit = len(list)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("  %-30s  %s\n", list[i].name, accent.Render(fmtDur(list[i].sec)))
		}
		fmt.Println()
	}

	if len(projects) > 0 {
		fmt.Println(header.Render("top projects"))
		type pj struct {
			name string
			sec  int
		}
		list := make([]pj, 0, len(projects))
		for k, v := range projects {
			list = append(list, pj{k, v})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].sec > list[j].sec })
		limit := 5
		if len(list) < limit {
			limit = len(list)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("  %-30s  %s\n", list[i].name, accent.Render(fmtDur(list[i].sec)))
		}
		fmt.Println()
	}

	if len(trackedSec) > 0 {
		renderTrackedTasksBlock(trackedSec, trackedSessions, trackedTitle, trackedPrefix, header, accent, dim)
	}

	fmt.Printf("%s %s\n\n",
		dim.Render("all-time "+noun+":"),
		accent.Render(fmtDur(allTimeTotal)))
}

// renderTrackedTasksBlock joins log entries tagged with a TaskID against the
// current task file, so the user sees per-task time alongside the task's
// current state and planned due date. Tasks deleted from the file still
// appear — the log is authoritative for historical time.
func renderTrackedTasksBlock(
	secs, sessions map[int]int, titles, prefixes map[int]string,
	header, accent, dim lipgloss.Style,
) {
	_, taskEntries, _ := readTaskFile()
	taskMap := make(map[int]*task)
	for _, fe := range taskEntries {
		if fe.task != nil && fe.task.ID > 0 {
			taskMap[fe.task.ID] = fe.task
		}
	}

	type row struct {
		id       int
		prefix   string
		title    string
		state    string
		due      string
		project  string
		sec      int
		sessions int
		overdue  bool
	}
	now := time.Now()
	var list []row
	for id, sec := range secs {
		r := row{id: id, sec: sec, sessions: sessions[id], title: titles[id], prefix: prefixes[id], state: "?"}
		if t, ok := taskMap[id]; ok {
			r.title = t.Title
			if t.Prefix != "" {
				r.prefix = t.Prefix
			}
			r.state = t.State.label()
			r.due = t.Due
			r.project = t.Project
			if d, ok := parseDateOnly(t.Due); ok && t.State != stateDone {
				if now.After(d.AddDate(0, 0, 1)) {
					r.overdue = true
				}
			}
		} else {
			r.state = "deleted"
		}
		if r.prefix == "" {
			r.prefix = "TSK"
		}
		list = append(list, r)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].sec > list[j].sec })
	limit := 10
	if len(list) < limit {
		limit = len(list)
	}

	fmt.Println(header.Render("tracked tasks"))
	for i := 0; i < limit; i++ {
		r := list[i]
		idLabel := accent.Render(fmt.Sprintf("%s-%d", r.prefix, r.id))
		title := r.title
		if len(title) > 24 {
			title = title[:23] + "…"
		}
		stateLabel := "[" + r.state + "]"
		if r.overdue {
			stateLabel = lipgloss.NewStyle().Foreground(colRed).Bold(true).Render("[overdue]")
		} else {
			stateLabel = dim.Render(stateLabel)
		}
		meta := ""
		if r.project != "" {
			meta += "#" + r.project
		}
		if r.due != "" {
			if meta != "" {
				meta += " · "
			}
			meta += "due " + r.due
		}
		sessionsLabel := fmt.Sprintf("%d sess", r.sessions)
		if r.sessions == 1 {
			sessionsLabel = "1 sess"
		}
		fmt.Printf("  %s  %-24s  %s  %s  %s  %s\n",
			idLabel,
			title,
			stateLabel,
			accent.Render(fmt.Sprintf("%8s", fmtDur(r.sec))),
			dim.Render(fmt.Sprintf("%8s", sessionsLabel)),
			dim.Render(meta))
	}
	fmt.Println()
}

// Set at build time via -ldflags "-X main.version=... -X main.commit=... -X main.date=..."
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "stats" {
		runStats(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "task" {
		runTask(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "todo" {
		runTodo(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "kanban" {
		runKanban(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "gantt" {
		runGantt(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "activity" {
		runActivity(os.Args[2:])
		return
	}

	cfg := loadConfig()

	workMin := new(int)
	flag.IntVar(workMin, "work", cfg.Work, "timer length in minutes")
	flag.IntVar(workMin, "w", cfg.Work, "shorthand for -work")

	shortMin := new(int)
	flag.IntVar(shortMin, "short", cfg.Short, "short break length in minutes (Pomodoro)")
	flag.IntVar(shortMin, "s", cfg.Short, "shorthand for -short")

	longMin := new(int)
	flag.IntVar(longMin, "long", cfg.Long, "long break length in minutes (Pomodoro)")
	flag.IntVar(longMin, "l", cfg.Long, "shorthand for -long")

	rounds := new(int)
	flag.IntVar(rounds, "rounds", cfg.Rounds, "work rounds before a long break (Pomodoro)")
	flag.IntVar(rounds, "r", cfg.Rounds, "shorthand for -rounds")

	task := new(string)
	flag.StringVar(task, "task", "", "task description to tag the session")
	flag.StringVar(task, "t", "", "shorthand for -task")

	pomodoro := new(bool)
	flag.BoolVar(pomodoro, "pomodoro", false, "full Pomodoro cycle — work + breaks + rounds")
	flag.BoolVar(pomodoro, "p", false, "shorthand for -pomodoro")

	timerMode := new(bool)
	flag.BoolVar(timerMode, "timer", false, "stopwatch mode — count up from 0 to track a task")
	flag.BoolVar(timerMode, "T", false, "shorthand for -timer")

	target := new(string)
	flag.StringVar(target, "target", "", "soft goal for -timer (e.g. 45m, 1h30m); notifies on reach, keeps running")

	minimal := new(bool)
	flag.BoolVar(minimal, "minimal", false, "hide logo and tomato animation")
	flag.BoolVar(minimal, "m", false, "shorthand for -minimal")

	sound := new(string)
	flag.StringVar(sound, "sound", cfg.Sound, `notification sound (macOS: "Glass", "Ping", "Hero", etc; "" or "default" for beep)`)
	flag.StringVar(sound, "S", cfg.Sound, "shorthand for -sound")

	showVersion := new(bool)
	flag.BoolVar(showVersion, "version", false, "print version and exit")
	flag.BoolVar(showVersion, "v", false, "shorthand for -version")

	showExamples := new(bool)
	flag.BoolVar(showExamples, "examples", false, "print usage examples for all modes and exit")
	flag.BoolVar(showExamples, "e", false, "shorthand for -examples")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "Thakkali — terminal Pomodoro timer with a rolling-tomato animation.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintln(w, "  thakkali [flags]")
		fmt.Fprintln(w, "  thakkali stats [-days N] [-mode all|pomodoro|timer]")
		fmt.Fprintln(w, "  thakkali task  add|list|move|done|rm|bulk|show ...")
		fmt.Fprintln(w, "  thakkali todo  — interactive TUI for the task file")
		fmt.Fprintln(w, "  thakkali kanban — three-column TODO/DOING/DONE board")
		fmt.Fprintln(w, "  thakkali gantt [-view week|month|year]")
		fmt.Fprintln(w, "  thakkali activity [-weeks N]")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Flags:")
		rows := [][2]string{
			{"-work, -w <int>", fmt.Sprintf("timer length in minutes (default %d)", cfg.Work)},
			{"-pomodoro, -p", "full Pomodoro cycle — work + breaks + rounds"},
			{"-timer, -T", "stopwatch mode — count up to track a task"},
			{"-target <dur>", "soft goal for -timer (e.g. 45m, 1h30m)"},
			{"-short, -s <int>", fmt.Sprintf("short break length in minutes, Pomodoro (default %d)", cfg.Short)},
			{"-long, -l <int>", fmt.Sprintf("long break length in minutes, Pomodoro (default %d)", cfg.Long)},
			{"-rounds, -r <int>", fmt.Sprintf("work rounds before a long break, Pomodoro (default %d)", cfg.Rounds)},
			{"-task, -t <string>", "task description to tag the session"},
			{"-minimal, -m", "hide logo and tomato animation"},
			{"-sound, -S <string>", `notification sound (macOS: "Glass", "Ping", "Hero", etc; "" for beep)`},
			{"-examples, -e", "print usage examples for all modes"},
			{"-version, -v", "print version and exit"},
		}
		maxW := 0
		for _, r := range rows {
			if len(r[0]) > maxW {
				maxW = len(r[0])
			}
		}
		for _, r := range rows {
			fmt.Fprintf(w, "  %-*s  %s\n", maxW, r[0], r[1])
		}
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("thakkali %s (%s) built %s\n", version, commit, date)
		return
	}

	if *showExamples {
		printExamples()
		return
	}

	if *timerMode && *pomodoro {
		fmt.Fprintln(os.Stderr, "error: -timer and -pomodoro are mutually exclusive")
		os.Exit(2)
	}

	var targetDur time.Duration
	if *target != "" {
		if !*timerMode {
			fmt.Fprintln(os.Stderr, "error: -target requires -timer")
			os.Exit(2)
		}
		d, err := time.ParseDuration(*target)
		if err != nil || d <= 0 {
			fmt.Fprintf(os.Stderr, "error: invalid -target %q (try 45m, 1h30m)\n", *target)
			os.Exit(2)
		}
		targetDur = d
	}

	// Resolve `-task TSK-N` against the current task file: replace the display
	// title with the tracked task's title, and carry its TaskID + Project
	// into the session log so stats can roll up per-task and per-project.
	var resolvedTaskID *int
	var resolvedPrefix, resolvedProject string
	if _, id, ok := parseTaskRef(*task); ok {
		path, entries, err := readTaskFile()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
			os.Exit(1)
		}
		tk := findTask(entries, *task)
		if tk == nil {
			fmt.Fprintf(os.Stderr, "error: no task with id %s\n", *task)
			os.Exit(2)
		}
		*task = tk.Title
		idCopy := id
		resolvedTaskID = &idCopy
		resolvedPrefix = tk.Prefix
		if resolvedPrefix == "" {
			resolvedPrefix = fileIDPrefix(path)
		}
		resolvedProject = tk.Project

		// Phase D: auto-promote todo → doing on first tagged session so the
		// @begin stamp marks when work actually started, not when the task was
		// captured. Backwards moves and explicit `task move` still win because
		// stampActuals reconciles state on every save.
		if tk.State == stateTodo {
			tk.State = stateDoing
			if werr := writeTaskFile(path, entries); werr != nil {
				fmt.Fprintf(os.Stderr, "warn: could not auto-stamp %s-%d: %v\n", resolvedPrefix, id, werr)
			}
		}
	}

	buildFrames()

	work := time.Duration(*workMin) * time.Minute
	m := model{
		durations: [3]time.Duration{
			work,
			time.Duration(*shortMin) * time.Minute,
			time.Duration(*longMin) * time.Minute,
		},
		rounds:     *rounds,
		task:       *task,
		taskID:     resolvedTaskID,
		taskPrefix: resolvedPrefix,
		project:    resolvedProject,
		sound:      *sound,
		pomodoro:  *pomodoro,
		minimal:   *minimal,
		countUp:   *timerMode,
		target:    targetDur,
		remaining: work,
		running:   true,
		phase:     phaseWork,
		round:     1,
	}
	if *timerMode {
		m.phase = phaseTimer
		m.remaining = 0
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("error:", err)
	}
}

// ---------------------------------------------------------------------------
// Task management (Phase 8 / A1)
// ---------------------------------------------------------------------------

type taskState uint8

const (
	stateTodo taskState = iota
	stateDoing
	stateDone
)

func (s taskState) marker() string {
	switch s {
	case stateDoing:
		return "*"
	case stateDone:
		return "x"
	default:
		return " "
	}
}

func (s taskState) label() string {
	switch s {
	case stateDoing:
		return "doing"
	case stateDone:
		return "done"
	default:
		return "todo"
	}
}

func parseStateLabel(s string) (taskState, bool) {
	switch strings.ToLower(s) {
	case "todo", "t":
		return stateTodo, true
	case "doing", "d", "in-progress", "wip":
		return stateDoing, true
	case "done", "x":
		return stateDone, true
	}
	return 0, false
}

// task is the in-memory representation of one checklist line.
type task struct {
	ID      int       // 0 = unassigned
	Prefix  string    // ID prefix, e.g. "TSK" or "THA"; empty = fall back to "TSK"
	State   taskState
	Title   string
	Project string
	Start   string    // YYYY-MM-DD, user-set
	Due     string    // YYYY-MM-DD, user-set
	Begin   time.Time // auto-stamped when moved to doing
	Done    time.Time // auto-stamped when moved to done
	Created time.Time
	Extras  []string  // unknown tokens preserved verbatim for round-trip
}

// ref returns the canonical "PREFIX-N" identifier for display. Falls back to
// "TSK" when no prefix is stored on the task.
func (t *task) ref() string {
	if t.ID == 0 {
		return ""
	}
	p := t.Prefix
	if p == "" {
		p = "TSK"
	}
	return fmt.Sprintf("%s-%d", p, t.ID)
}

// fileIDPrefix returns the ID prefix to use for tasks in this file:
//   - the global task file → "TSK"
//   - a project-local file → derived from the containing directory name,
//     e.g. /Users/.../Thakkali/thakkali.md → "THAK"
func fileIDPrefix(path string) string {
	if dir, err := thakkaliDir(); err == nil && dir != "" {
		if strings.HasPrefix(path, dir+string(filepath.Separator)) || path == filepath.Join(dir, "tasks.md") {
			return "TSK"
		}
	}
	parent := filepath.Dir(path)
	base := filepath.Base(parent)
	if base == ".thakkali" {
		base = filepath.Base(filepath.Dir(parent))
	}
	return derivePrefix(base)
}

func derivePrefix(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		}
	}
	s := strings.ToUpper(b.String())
	if len(s) > 4 {
		s = s[:4]
	}
	if len(s) < 2 {
		return "TSK"
	}
	return s
}

// logRef formats a log entry's tracked task id using its stored prefix,
// falling back to "TSK" for entries written before v2 storage.
func logRef(e logEntry) string {
	if e.TaskID == nil {
		return ""
	}
	p := e.TaskPrefix
	if p == "" {
		p = "TSK"
	}
	return fmt.Sprintf("%s-%d", p, *e.TaskID)
}

// fileEntry preserves file order across tasks and free-form lines.
type fileEntry struct {
	task *task  // non-nil if this entry is a task
	raw  string // otherwise the original line
}

const taskHeading = "# Thakkali tasks"

// taskFilePath returns the resolved task-file path and whether it already
// exists on disk.
//
// Discovery order:
//  1. ./thakkali.md in cwd
//  2. ./.thakkali/tasks.md in cwd
//  3. Fallback to global <thakkaliDir()>/tasks.md
//
// When `forWrite` is true and no file exists anywhere, the write target is
// ./thakkali.md in cwd so project-local is the default — colleagues sharing
// the same repo see the same task list.
func taskFilePath() (string, bool, error) {
	return resolveTaskPath(false)
}

func taskFilePathForWrite() (string, bool, error) {
	return resolveTaskPath(true)
}

func resolveTaskPath(forWrite bool) (string, bool, error) {
	candidates := []string{
		"thakkali.md",
		filepath.Join(".thakkali", "tasks.md"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs, true, nil
		}
	}
	dir, err := thakkaliDir()
	if err != nil {
		return "", false, err
	}
	global := filepath.Join(dir, "tasks.md")
	if st, err := os.Stat(global); err == nil && !st.IsDir() {
		return global, true, nil
	}
	if forWrite {
		abs, _ := filepath.Abs("thakkali.md")
		return abs, false, nil
	}
	return global, false, nil
}

func readTaskFile() (string, []fileEntry, error) {
	return readTaskFileMode(false)
}

func readTaskFileForWrite() (string, []fileEntry, error) {
	return readTaskFileMode(true)
}

func readTaskFileMode(forWrite bool) (string, []fileEntry, error) {
	path, exists, err := resolveTaskPath(forWrite)
	if err != nil {
		return "", nil, err
	}
	if !exists {
		return path, nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return path, nil, err
	}
	entries := parseTaskFile(string(data))
	return path, entries, nil
}

func parseTaskFile(src string) []fileEntry {
	var out []fileEntry
	for _, line := range strings.Split(strings.TrimRight(src, "\n"), "\n") {
		if t, ok := parseTaskLine(line); ok {
			tt := t
			out = append(out, fileEntry{task: &tt})
			continue
		}
		out = append(out, fileEntry{raw: line})
	}
	return out
}

// parseTaskLine recognizes `- [ ] ...`, `- [*] ...`, `- [x] ...` lines.
func parseTaskLine(line string) (task, bool) {
	s := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(s, "- [") || len(s) < 6 || s[4] != ']' || s[5] != ' ' {
		return task{}, false
	}
	var st taskState
	switch s[3] {
	case ' ':
		st = stateTodo
	case '*', 'o', '/':
		st = stateDoing
	case 'x', 'X':
		st = stateDone
	default:
		return task{}, false
	}
	rest := strings.TrimSpace(s[6:])

	t := task{State: st}
	tokens := strings.Fields(rest)
	var titleWords []string
	titleClosed := false
	for _, tok := range tokens {
		switch {
		case !titleClosed && t.ID == 0 && isTaskRefToken(tok):
			pfx, n, _ := splitTaskRef(tok)
			t.ID = n
			t.Prefix = pfx
			continue
		case strings.HasPrefix(tok, "#") && len(tok) > 1:
			titleClosed = true
			if t.Project == "" {
				t.Project = tok[1:]
			} else {
				t.Extras = append(t.Extras, tok)
			}
		case strings.HasPrefix(tok, "@") && strings.Contains(tok, ":"):
			titleClosed = true
			kv := tok[1:]
			idx := strings.Index(kv, ":")
			key, val := kv[:idx], kv[idx+1:]
			if !assignField(&t, key, val) {
				t.Extras = append(t.Extras, tok)
			}
		default:
			if titleClosed {
				t.Extras = append(t.Extras, tok)
			} else {
				titleWords = append(titleWords, tok)
			}
		}
	}
	t.Title = strings.Join(titleWords, " ")
	return t, true
}

func assignField(t *task, key, val string) bool {
	switch key {
	case "start":
		t.Start = val
	case "due":
		t.Due = val
	case "begin":
		if ts, err := time.Parse(time.RFC3339, val); err == nil {
			t.Begin = ts
		} else {
			return false
		}
	case "done":
		if ts, err := time.Parse(time.RFC3339, val); err == nil {
			t.Done = ts
		} else {
			return false
		}
	case "created":
		if ts, err := time.Parse(time.RFC3339, val); err == nil {
			t.Created = ts
		} else {
			return false
		}
	default:
		return false
	}
	return true
}

// renderTaskLine is the inverse of parseTaskLine.
func renderTaskLine(t task) string {
	var b strings.Builder
	fmt.Fprintf(&b, "- [%s] ", t.State.marker())
	if t.ID > 0 {
		p := t.Prefix
		if p == "" {
			p = "TSK"
		}
		fmt.Fprintf(&b, "%s-%d ", p, t.ID)
	}
	if t.Title != "" {
		b.WriteString(t.Title)
	}
	if t.Project != "" {
		b.WriteString(" #")
		b.WriteString(t.Project)
	}
	if t.Start != "" {
		b.WriteString(" @start:")
		b.WriteString(t.Start)
	}
	if t.Due != "" {
		b.WriteString(" @due:")
		b.WriteString(t.Due)
	}
	if !t.Begin.IsZero() {
		b.WriteString(" @begin:")
		b.WriteString(t.Begin.UTC().Format(time.RFC3339))
	}
	if !t.Done.IsZero() {
		b.WriteString(" @done:")
		b.WriteString(t.Done.UTC().Format(time.RFC3339))
	}
	if !t.Created.IsZero() {
		b.WriteString(" @created:")
		b.WriteString(t.Created.UTC().Format(time.RFC3339))
	}
	for _, extra := range t.Extras {
		b.WriteByte(' ')
		b.WriteString(extra)
	}
	return b.String()
}

// assignIDs assigns monotonic IDs to any unassigned tasks. Returns updated
// entries.
func assignIDs(entries []fileEntry) {
	max := 0
	for _, e := range entries {
		if e.task != nil && e.task.ID > max {
			max = e.task.ID
		}
	}
	for _, e := range entries {
		if e.task != nil && e.task.ID == 0 {
			max++
			e.task.ID = max
		}
	}
}

// stampActuals enforces stamp/state consistency on every save:
//   - todo:  clear @begin and @done (never started)
//   - doing: stamp @begin if missing; clear @done (no longer complete)
//   - done:  stamp both if missing
//
// Uniform behavior whether the state change came from CLI, TUI, kanban, or
// $EDITOR bulk edit.
func stampActuals(entries []fileEntry) {
	now := time.Now().UTC().Truncate(time.Second)
	for _, e := range entries {
		if e.task == nil {
			continue
		}
		t := e.task
		if t.Created.IsZero() {
			t.Created = now
		}
		switch t.State {
		case stateTodo:
			t.Begin = time.Time{}
			t.Done = time.Time{}
		case stateDoing:
			if t.Begin.IsZero() {
				t.Begin = now
			}
			t.Done = time.Time{}
		case stateDone:
			if t.Begin.IsZero() {
				t.Begin = now
			}
			if t.Done.IsZero() {
				t.Done = now
			}
		}
	}
}

// writeTaskFile rewrites the task file. Creates the parent directory and a
// heading on first write.
func writeTaskFile(path string, entries []fileEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	assignIDs(entries)
	stampActuals(entries)
	filePrefix := fileIDPrefix(path)
	for _, e := range entries {
		if e.task != nil && e.task.Prefix == "" {
			e.task.Prefix = filePrefix
		}
	}

	hasHeading := false
	for _, e := range entries {
		if e.task == nil && strings.TrimSpace(e.raw) == taskHeading {
			hasHeading = true
			break
		}
	}

	var b strings.Builder
	if !hasHeading {
		b.WriteString(taskHeading)
		b.WriteString("\n\n")
	}
	for _, e := range entries {
		if e.task != nil {
			b.WriteString(renderTaskLine(*e.task))
		} else {
			b.WriteString(e.raw)
		}
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// findTask resolves an id string (e.g. "TSK-3", "THA-3", or "3") to an
// entry pointer. The prefix is accepted for ergonomics but ignored for
// matching — IDs are unique within a single task file.
func findTask(entries []fileEntry, idArg string) *task {
	s := strings.TrimSpace(idArg)
	if _, n, ok := splitTaskRef(s); ok {
		for _, e := range entries {
			if e.task != nil && e.task.ID == n {
				return e.task
			}
		}
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return nil
	}
	for _, e := range entries {
		if e.task != nil && e.task.ID == n {
			return e.task
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Task CLI dispatcher
// ---------------------------------------------------------------------------

func runTask(args []string) {
	if len(args) == 0 {
		taskUsage(os.Stdout)
		return
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "add":
		taskAdd(rest)
	case "list", "ls":
		taskList(rest)
	case "move", "mv":
		taskMove(rest)
	case "done":
		taskDone(rest)
	case "rm", "del", "delete":
		taskRm(rest)
	case "bulk", "edit":
		taskBulk(rest)
	case "show":
		taskShow(rest)
	case "help", "-h", "--help":
		taskUsage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown task subcommand %q\n\n", sub)
		taskUsage(os.Stderr)
		os.Exit(2)
	}
}

func taskUsage(w *os.File) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  thakkali task add \"title\" [-p project] [-s YYYY-MM-DD] [-d YYYY-MM-DD]")
	fmt.Fprintln(w, "  thakkali task list [-state todo|doing|done|all] [-p project]")
	fmt.Fprintln(w, "  thakkali task move <id> <todo|doing|done>")
	fmt.Fprintln(w, "  thakkali task done <id>")
	fmt.Fprintln(w, "  thakkali task rm   <id>")
	fmt.Fprintln(w, "  thakkali task show <id>")
	fmt.Fprintln(w, "  thakkali task bulk")
}

func taskAdd(args []string) {
	fs := flag.NewFlagSet("task add", flag.ExitOnError)
	project := fs.String("p", "", "project tag (stored as #project)")
	start := fs.String("s", "", "planned start date YYYY-MM-DD")
	due := fs.String("d", "", "planned due date YYYY-MM-DD")
	_ = fs.Parse(reorderFlags(args))

	positional := fs.Args()
	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "error: task add requires a title")
		taskUsage(os.Stderr)
		os.Exit(2)
	}
	title := strings.TrimSpace(strings.Join(positional, " "))
	if title == "" {
		fmt.Fprintln(os.Stderr, "error: title cannot be empty")
		os.Exit(2)
	}
	if *start != "" && !isDate(*start) {
		fmt.Fprintf(os.Stderr, "error: -s must be YYYY-MM-DD (got %q)\n", *start)
		os.Exit(2)
	}
	if *due != "" && !isDate(*due) {
		fmt.Fprintf(os.Stderr, "error: -d must be YYYY-MM-DD (got %q)\n", *due)
		os.Exit(2)
	}

	path, entries, err := readTaskFileForWrite()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
		os.Exit(1)
	}
	t := task{
		State:   stateTodo,
		Title:   title,
		Project: strings.TrimPrefix(*project, "#"),
		Start:   *start,
		Due:     *due,
		Created: time.Now().UTC().Truncate(time.Second),
	}
	entries = append(entries, fileEntry{task: &t})
	if err := writeTaskFile(path, entries); err != nil {
		fmt.Fprintf(os.Stderr, "error writing task file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("added %s %s  %s\n", t.ref(), t.Title, lipgloss.NewStyle().Foreground(colDim).Render("→ "+path))
}

func taskList(args []string) {
	fs := flag.NewFlagSet("task list", flag.ExitOnError)
	state := fs.String("state", "active", "filter: todo | doing | done | active | all")
	project := fs.String("p", "", "filter by project tag")
	_ = fs.Parse(reorderFlags(args))

	path, entries, err := readTaskFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
		os.Exit(1)
	}

	dim := lipgloss.NewStyle().Foreground(colDim)
	header := lipgloss.NewStyle().Foreground(colRed).Bold(true)
	accent := lipgloss.NewStyle().Foreground(colBrightRed)

	fmt.Println(dim.Render("file: " + path))
	if len(entries) == 0 {
		fmt.Println(dim.Render("no tasks yet. run: thakkali task add \"...\""))
		return
	}

	want := func(s taskState) bool {
		switch *state {
		case "all":
			return true
		case "todo":
			return s == stateTodo
		case "doing":
			return s == stateDoing
		case "done":
			return s == stateDone
		default: // active
			return s != stateDone
		}
	}
	projectFilter := strings.TrimPrefix(*project, "#")

	// Group DOING → TODO → DONE.
	groups := [3][]task{}
	order := []taskState{stateDoing, stateTodo, stateDone}
	for _, e := range entries {
		if e.task == nil {
			continue
		}
		if !want(e.task.State) {
			continue
		}
		if projectFilter != "" && e.task.Project != projectFilter {
			continue
		}
		groups[e.task.State] = append(groups[e.task.State], *e.task)
	}

	total := 0
	for _, st := range order {
		total += len(groups[st])
	}
	if total == 0 {
		fmt.Println(dim.Render("(no matching tasks)"))
		return
	}

	for _, st := range order {
		rows := groups[st]
		if len(rows) == 0 {
			continue
		}
		fmt.Println()
		fmt.Println(header.Render(strings.ToUpper(st.label())))
		for _, t := range rows {
			id := "     "
			if ref := t.ref(); ref != "" {
				id = ref
			}
			meta := taskMetaSuffix(t)
			fmt.Printf("  %s  %s%s\n",
				accent.Render(id),
				t.Title,
				dim.Render(meta))
		}
	}
	fmt.Println()
}

func taskMetaSuffix(t task) string {
	parts := []string{}
	if t.Project != "" {
		parts = append(parts, "#"+t.Project)
	}
	if t.Start != "" {
		parts = append(parts, "start "+t.Start)
	}
	if t.Due != "" {
		parts = append(parts, "due "+t.Due)
	}
	if !t.Begin.IsZero() && t.State == stateDoing {
		parts = append(parts, "began "+humanAgo(t.Begin))
	}
	if !t.Done.IsZero() && t.State == stateDone {
		parts = append(parts, "done "+humanAgo(t.Done))
	}
	if len(parts) == 0 {
		return ""
	}
	return "  (" + strings.Join(parts, " · ") + ")"
}

func humanAgo(ts time.Time) string {
	d := time.Since(ts)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func taskMove(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "error: task move <id> <todo|doing|done>")
		os.Exit(2)
	}
	state, ok := parseStateLabel(args[1])
	if !ok {
		fmt.Fprintf(os.Stderr, "error: invalid state %q (want todo | doing | done)\n", args[1])
		os.Exit(2)
	}
	transitionTask(args[0], state)
}

func taskDone(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "error: task done <id>")
		os.Exit(2)
	}
	transitionTask(args[0], stateDone)
}

func transitionTask(idArg string, newState taskState) {
	path, entries, err := readTaskFileForWrite()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
		os.Exit(1)
	}
	t := findTask(entries, idArg)
	if t == nil {
		fmt.Fprintf(os.Stderr, "error: no task with id %q in %s\n", idArg, path)
		os.Exit(2)
	}
	prev := t.State
	t.State = newState
	if err := writeTaskFile(path, entries); err != nil {
		fmt.Fprintf(os.Stderr, "error writing task file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s %s: %s → %s\n", t.ref(), t.Title, prev.label(), newState.label())
}

func taskRm(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "error: task rm <id>")
		os.Exit(2)
	}
	path, entries, err := readTaskFileForWrite()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
		os.Exit(1)
	}
	idx := -1
	for i, e := range entries {
		if e.task != nil {
			if match := findTask([]fileEntry{e}, args[0]); match != nil {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		fmt.Fprintf(os.Stderr, "error: no task with id %q\n", args[0])
		os.Exit(2)
	}
	removed := entries[idx].task
	entries = append(entries[:idx], entries[idx+1:]...)
	if err := writeTaskFile(path, entries); err != nil {
		fmt.Fprintf(os.Stderr, "error writing task file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("removed %s %s\n", removed.ref(), removed.Title)
}

func taskShow(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "error: task show <id>")
		os.Exit(2)
	}
	_, entries, err := readTaskFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
		os.Exit(1)
	}
	t := findTask(entries, args[0])
	if t == nil {
		fmt.Fprintf(os.Stderr, "error: no task with id %q\n", args[0])
		os.Exit(2)
	}
	accent := lipgloss.NewStyle().Foreground(colBrightRed).Bold(true)
	dim := lipgloss.NewStyle().Foreground(colDim)
	fmt.Println()
	fmt.Printf("%s  %s\n", accent.Render(t.ref()), t.Title)
	fmt.Printf("  %s %s\n", dim.Render("state:  "), t.State.label())
	if t.Project != "" {
		fmt.Printf("  %s #%s\n", dim.Render("project:"), t.Project)
	}
	if t.Start != "" {
		fmt.Printf("  %s %s\n", dim.Render("start:  "), t.Start)
	}
	if t.Due != "" {
		fmt.Printf("  %s %s\n", dim.Render("due:    "), t.Due)
	}
	if !t.Begin.IsZero() {
		fmt.Printf("  %s %s (%s)\n", dim.Render("began:  "), t.Begin.Local().Format(time.RFC3339), humanAgo(t.Begin))
	}
	if !t.Done.IsZero() {
		fmt.Printf("  %s %s (%s)\n", dim.Render("done:   "), t.Done.Local().Format(time.RFC3339), humanAgo(t.Done))
	}
	if !t.Created.IsZero() {
		fmt.Printf("  %s %s\n", dim.Render("created:"), t.Created.Local().Format(time.RFC3339))
	}
	if len(t.Extras) > 0 {
		fmt.Printf("  %s %s\n", dim.Render("extras: "), strings.Join(t.Extras, " "))
	}
	fmt.Println()
}

func taskBulk(_ []string) {
	path, _, err := taskFilePathForWrite()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving task file: %v\n", err)
		os.Exit(1)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(path, []byte(taskHeading+"\n\n"), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
	if err := openEditor(path); err != nil {
		fmt.Fprintf(os.Stderr, "error launching editor: %v\n", err)
		os.Exit(1)
	}
	// Reparse + rewrite so IDs are assigned and actuals stamped.
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error re-reading task file: %v\n", err)
		os.Exit(1)
	}
	entries := parseTaskFile(string(data))
	if err := writeTaskFile(path, entries); err != nil {
		fmt.Fprintf(os.Stderr, "error rewriting task file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("saved %s\n", path)
}

func openEditor(path string) error {
	candidates := []string{os.Getenv("EDITOR"), "nvim", "vim"}
	var editor string
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if _, err := exec.LookPath(c); err == nil {
			editor = c
			break
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found — set $EDITOR or install nvim/vim")
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// reorderFlags shuffles a slice so flag tokens come first and positional
// args come last. Lets users write `task add "title" -p proj` — the usual
// Go `flag` package stops at the first positional otherwise.
// Known value-taking short flags for task subcommands:
var taskValueFlags = map[string]bool{
	"-p": true, "-s": true, "-d": true, "-state": true, "--state": true,
}

func reorderFlags(args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") && a != "-" && a != "--" {
			flags = append(flags, a)
			// If the flag takes a value and it's a separate token, grab it.
			if taskValueFlags[a] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
			continue
		}
		positional = append(positional, a)
	}
	return append(flags, positional...)
}

func isDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// parseTaskRef returns the prefix and numeric ID for any "<PREFIX>-N" string
// (case-insensitive on the prefix). The prefix is normalized to uppercase.
func parseTaskRef(s string) (string, int, bool) {
	return splitTaskRef(s)
}

// splitTaskRef accepts any "<ALPHA>-<DIGITS>" token and returns its upper-
// cased prefix and numeric id. Used by both the parser and the CLI.
func splitTaskRef(s string) (string, int, bool) {
	s = strings.TrimSpace(s)
	idx := strings.IndexByte(s, '-')
	if idx < 2 || idx >= len(s)-1 {
		return "", 0, false
	}
	pfx := s[:idx]
	for _, r := range pfx {
		if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') {
			return "", 0, false
		}
	}
	n, err := strconv.Atoi(s[idx+1:])
	if err != nil || n <= 0 {
		return "", 0, false
	}
	return strings.ToUpper(pfx), n, true
}

func isTaskRefToken(s string) bool {
	_, _, ok := splitTaskRef(s)
	return ok
}

// ---------------------------------------------------------------------------
// Task TUI (Phase 8 / A2) — `thakkali todo`
// ---------------------------------------------------------------------------

type todoMode int

const (
	todoModeNormal todoMode = iota
	todoModeFilter
	todoModeNew
	todoModeEdit
)

type todoModel struct {
	path      string
	entries   []fileEntry
	visible   []*task // flat DOING → TODO → DONE, filter-applied
	cursor    int
	mode      todoMode
	input     string
	filter    string
	editingID int
	errMsg    string
	showHelp  bool
	width     int
	height    int
}

func runTodo(_ []string) {
	var m todoModel
	if err := m.reload(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func (m *todoModel) reload() error {
	path, entries, err := readTaskFileForWrite()
	if err != nil {
		return err
	}
	m.path = path
	m.entries = entries
	m.rebuildVisible()
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	return nil
}

func (m *todoModel) rebuildVisible() {
	m.visible = m.visible[:0]
	f := strings.ToLower(m.filter)
	for _, st := range []taskState{stateDoing, stateTodo, stateDone} {
		for i := range m.entries {
			t := m.entries[i].task
			if t == nil || t.State != st {
				continue
			}
			if f != "" &&
				!strings.Contains(strings.ToLower(t.Title), f) &&
				!strings.Contains(strings.ToLower(t.Project), f) {
				continue
			}
			m.visible = append(m.visible, t)
		}
	}
}

func (m *todoModel) selected() *task {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil
	}
	return m.visible[m.cursor]
}

// save writes the file, reloads from disk, and restores the selection by ID
// so the cursor stays on the same task across rewrites.
func (m *todoModel) save() {
	sid := 0
	if t := m.selected(); t != nil {
		sid = t.ID
	}
	if err := writeTaskFile(m.path, m.entries); err != nil {
		m.errMsg = "save: " + err.Error()
		return
	}
	m.errMsg = ""
	_ = m.reload()
	if sid > 0 {
		for i, t := range m.visible {
			if t.ID == sid {
				m.cursor = i
				break
			}
		}
	}
}

func (m todoModel) Init() tea.Cmd { return nil }

func (m todoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.mode == todoModeNormal {
			return m.updateNormal(msg)
		}
		return m.updateInput(msg)
	}
	return m, nil
}

func (m todoModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.visible)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		if len(m.visible) > 0 {
			m.cursor = len(m.visible) - 1
		}
	case " ", "enter":
		m.cycle()
	case "n":
		m.mode = todoModeNew
		m.input = ""
	case "e":
		if t := m.selected(); t != nil {
			m.mode = todoModeEdit
			m.input = t.Title
			m.editingID = t.ID
		}
	case "d":
		m.deleteSelected()
	case "/":
		m.mode = todoModeFilter
		m.input = m.filter
	case "c":
		if m.filter != "" {
			m.filter = ""
			m.rebuildVisible()
		}
	case "r":
		_ = m.reload()
	case "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m todoModel) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Esc in filter mode keeps the current filter but exits typing.
		// In new/edit mode it discards the draft.
		m.mode = todoModeNormal
		m.input = ""
		return m, nil
	case tea.KeyEnter:
		m.commit()
		return m, nil
	case tea.KeyBackspace:
		if len(m.input) > 0 {
			r := []rune(m.input)
			m.input = string(r[:len(r)-1])
			if m.mode == todoModeFilter {
				m.filter = m.input
				m.rebuildVisible()
			}
		}
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.input += " "
		}
		if m.mode == todoModeFilter {
			m.filter = m.input
			m.rebuildVisible()
		}
		return m, nil
	}
	return m, nil
}

func (m *todoModel) commit() {
	switch m.mode {
	case todoModeFilter:
		m.filter = m.input
		m.rebuildVisible()
	case todoModeNew:
		title := strings.TrimSpace(m.input)
		if title != "" {
			t := task{
				State:   stateTodo,
				Title:   title,
				Created: time.Now().UTC().Truncate(time.Second),
			}
			m.entries = append(m.entries, fileEntry{task: &t})
			m.save()
		}
	case todoModeEdit:
		title := strings.TrimSpace(m.input)
		if title != "" {
			for i := range m.entries {
				if m.entries[i].task != nil && m.entries[i].task.ID == m.editingID {
					m.entries[i].task.Title = title
					break
				}
			}
			m.save()
		}
	}
	m.mode = todoModeNormal
	m.input = ""
}

func (m *todoModel) cycle() {
	t := m.selected()
	if t == nil {
		return
	}
	switch t.State {
	case stateTodo:
		t.State = stateDoing
	case stateDoing:
		t.State = stateDone
	case stateDone:
		t.State = stateTodo
	}
	m.save()
}

func (m *todoModel) deleteSelected() {
	t := m.selected()
	if t == nil {
		return
	}
	id := t.ID
	for i, e := range m.entries {
		if e.task != nil && e.task.ID == id {
			m.entries = append(m.entries[:i], m.entries[i+1:]...)
			break
		}
	}
	m.save()
}

func (m todoModel) View() string {
	header := lipgloss.NewStyle().Foreground(colRed).Bold(true)
	accent := lipgloss.NewStyle().Foreground(colBrightRed)
	dim := lipgloss.NewStyle().Foreground(colDim)
	doingStyle := lipgloss.NewStyle().Foreground(colBrightRed).Bold(true)
	doneStyle2 := lipgloss.NewStyle().Foreground(colDim).Strikethrough(true)
	cursorMark := lipgloss.NewStyle().Foreground(colBrightRed).Bold(true).Render("▸")
	errStyle := lipgloss.NewStyle().Foreground(colRed)

	var b strings.Builder
	b.WriteString(header.Render("THAKKALI · todo"))
	b.WriteString("\n")
	b.WriteString(dim.Render(m.path))
	b.WriteString("\n")
	if m.filter != "" {
		b.WriteString(dim.Render(fmt.Sprintf("filter: %q", m.filter)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if len(m.visible) == 0 {
		if m.filter != "" {
			b.WriteString(dim.Render("(no matches — press c to clear filter)"))
		} else {
			b.WriteString(dim.Render("(empty — press n to capture a task)"))
		}
		b.WriteString("\n")
	} else {
		var lastState taskState = 255
		for i, t := range m.visible {
			if t.State != lastState {
				if i > 0 {
					b.WriteString("\n")
				}
				b.WriteString(header.Render(strings.ToUpper(t.State.label())))
				b.WriteString("\n")
				lastState = t.State
			}
			marker := " "
			if i == m.cursor {
				marker = cursorMark
			}
			id := "     "
			if ref := t.ref(); ref != "" {
				id = ref
			}
			title := t.Title
			switch t.State {
			case stateDoing:
				title = doingStyle.Render(title)
			case stateDone:
				title = doneStyle2.Render(title)
			}
			meta := taskMetaSuffix(*t)
			fmt.Fprintf(&b, "%s %s  %s%s\n", marker, accent.Render(id), title, dim.Render(meta))
		}
	}

	b.WriteString("\n")
	switch m.mode {
	case todoModeNew:
		b.WriteString(accent.Render("new: ") + m.input + "▋")
	case todoModeEdit:
		b.WriteString(accent.Render("edit: ") + m.input + "▋")
	case todoModeFilter:
		b.WriteString(accent.Render("/") + m.input + "▋")
	default:
		if m.showHelp {
			b.WriteString(dim.Render("j/k move · g/G top/bot · space cycle · n new · e edit · d delete · / filter · c clear · r reload · ? help · q quit"))
		} else {
			b.WriteString(dim.Render("? help · q quit"))
		}
	}
	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.errMsg))
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Kanban TUI (Phase 8 / B) — `thakkali kanban`
// ---------------------------------------------------------------------------

type kanbanModel struct {
	path      string
	entries   []fileEntry
	cols      [3][]*task // indexed by taskState
	cursor    [3]int
	focus     taskState
	mode      todoMode
	input     string
	filter    string
	editingID int
	errMsg    string
	showHelp  bool
	width     int
	height    int
}

func runKanban(_ []string) {
	m := kanbanModel{focus: stateTodo, width: 120, height: 30}
	if err := m.reload(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func (m *kanbanModel) reload() error {
	path, entries, err := readTaskFileForWrite()
	if err != nil {
		return err
	}
	m.path = path
	m.entries = entries
	m.rebuildCols()
	for i := 0; i < 3; i++ {
		if m.cursor[i] >= len(m.cols[i]) {
			m.cursor[i] = len(m.cols[i]) - 1
		}
		if m.cursor[i] < 0 {
			m.cursor[i] = 0
		}
	}
	return nil
}

func (m *kanbanModel) rebuildCols() {
	for i := range m.cols {
		m.cols[i] = m.cols[i][:0]
	}
	f := strings.ToLower(m.filter)
	for i := range m.entries {
		t := m.entries[i].task
		if t == nil {
			continue
		}
		if f != "" &&
			!strings.Contains(strings.ToLower(t.Title), f) &&
			!strings.Contains(strings.ToLower(t.Project), f) {
			continue
		}
		m.cols[t.State] = append(m.cols[t.State], t)
	}
}

func (m *kanbanModel) selected() *task {
	col := m.cols[m.focus]
	c := m.cursor[m.focus]
	if c < 0 || c >= len(col) {
		return nil
	}
	return col[c]
}

func (m *kanbanModel) save() {
	sid := 0
	if t := m.selected(); t != nil {
		sid = t.ID
	}
	if err := writeTaskFile(m.path, m.entries); err != nil {
		m.errMsg = "save: " + err.Error()
		return
	}
	m.errMsg = ""
	_ = m.reload()
	if sid > 0 {
		for st := range m.cols {
			for i, t := range m.cols[st] {
				if t.ID == sid {
					m.focus = taskState(st)
					m.cursor[st] = i
					return
				}
			}
		}
	}
}

func (m kanbanModel) Init() tea.Cmd { return nil }

func (m kanbanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.mode == todoModeNormal {
			return m.updateNormal(msg)
		}
		return m.updateInput(msg)
	}
	return m, nil
}

func (m kanbanModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "h", "left":
		if m.focus > stateTodo {
			m.focus--
		}
	case "l", "right":
		if m.focus < stateDone {
			m.focus++
		}
	case "j", "down":
		c := m.cursor[m.focus]
		if c < len(m.cols[m.focus])-1 {
			m.cursor[m.focus] = c + 1
		}
	case "k", "up":
		if m.cursor[m.focus] > 0 {
			m.cursor[m.focus]--
		}
	case "g", "home":
		m.cursor[m.focus] = 0
	case "G", "end":
		if n := len(m.cols[m.focus]); n > 0 {
			m.cursor[m.focus] = n - 1
		}
	case " ", "enter":
		m.cycle()
	case ">", "shift+right", "L":
		m.shift(1)
	case "<", "shift+left", "H":
		m.shift(-1)
	case "n":
		m.mode = todoModeNew
		m.input = ""
	case "e":
		if t := m.selected(); t != nil {
			m.mode = todoModeEdit
			m.input = t.Title
			m.editingID = t.ID
		}
	case "d":
		m.deleteSelected()
	case "/":
		m.mode = todoModeFilter
		m.input = m.filter
	case "c":
		if m.filter != "" {
			m.filter = ""
			m.rebuildCols()
		}
	case "r":
		_ = m.reload()
	case "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m kanbanModel) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = todoModeNormal
		m.input = ""
		return m, nil
	case tea.KeyEnter:
		m.commit()
		return m, nil
	case tea.KeyBackspace:
		if len(m.input) > 0 {
			r := []rune(m.input)
			m.input = string(r[:len(r)-1])
			if m.mode == todoModeFilter {
				m.filter = m.input
				m.rebuildCols()
			}
		}
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.input += " "
		}
		if m.mode == todoModeFilter {
			m.filter = m.input
			m.rebuildCols()
		}
		return m, nil
	}
	return m, nil
}

func (m *kanbanModel) commit() {
	switch m.mode {
	case todoModeFilter:
		m.filter = m.input
		m.rebuildCols()
	case todoModeNew:
		title := strings.TrimSpace(m.input)
		if title != "" {
			t := task{
				State:   m.focus,
				Title:   title,
				Created: time.Now().UTC().Truncate(time.Second),
			}
			m.entries = append(m.entries, fileEntry{task: &t})
			m.save()
		}
	case todoModeEdit:
		title := strings.TrimSpace(m.input)
		if title != "" {
			for i := range m.entries {
				if m.entries[i].task != nil && m.entries[i].task.ID == m.editingID {
					m.entries[i].task.Title = title
					break
				}
			}
			m.save()
		}
	}
	m.mode = todoModeNormal
	m.input = ""
}

func (m *kanbanModel) cycle() {
	t := m.selected()
	if t == nil {
		return
	}
	switch t.State {
	case stateTodo:
		t.State = stateDoing
	case stateDoing:
		t.State = stateDone
	case stateDone:
		t.State = stateTodo
	}
	m.save()
}

// shift moves the selected task one column over (dir = -1 or +1) without
// wrapping. Auto-stamping fires inside writeTaskFile on save.
func (m *kanbanModel) shift(dir int) {
	t := m.selected()
	if t == nil {
		return
	}
	next := int(t.State) + dir
	if next < int(stateTodo) || next > int(stateDone) {
		return
	}
	t.State = taskState(next)
	m.save()
}

func (m *kanbanModel) deleteSelected() {
	t := m.selected()
	if t == nil {
		return
	}
	id := t.ID
	for i, e := range m.entries {
		if e.task != nil && e.task.ID == id {
			m.entries = append(m.entries[:i], m.entries[i+1:]...)
			break
		}
	}
	m.save()
}

func (m kanbanModel) View() string {
	header := lipgloss.NewStyle().Foreground(colRed).Bold(true)
	dim := lipgloss.NewStyle().Foreground(colDim)
	accent := lipgloss.NewStyle().Foreground(colBrightRed)
	errStyle := lipgloss.NewStyle().Foreground(colRed)

	colW := (m.width - 8) / 3
	if colW < 18 {
		colW = 18
	}

	var b strings.Builder
	b.WriteString(header.Render("THAKKALI · kanban"))
	b.WriteString("\n")
	b.WriteString(dim.Render(m.path))
	b.WriteString("\n")
	if m.filter != "" {
		b.WriteString(dim.Render(fmt.Sprintf("filter: %q", m.filter)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	cols := []string{
		m.renderColumn(stateTodo, colW),
		m.renderColumn(stateDoing, colW),
		m.renderColumn(stateDone, colW),
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	b.WriteString("\n")

	switch m.mode {
	case todoModeNew:
		b.WriteString(accent.Render(fmt.Sprintf("new in %s: ", strings.ToUpper(m.focus.label()))) + m.input + "▋")
	case todoModeEdit:
		b.WriteString(accent.Render("edit: ") + m.input + "▋")
	case todoModeFilter:
		b.WriteString(accent.Render("/") + m.input + "▋")
	default:
		if m.showHelp {
			b.WriteString(dim.Render("h/l switch col · j/k move · </> shift · space cycle · n new · e edit · d delete · / filter · c clear · r reload · ? help · q quit"))
		} else {
			b.WriteString(dim.Render("? help · q quit"))
		}
	}
	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.errMsg))
	}
	return b.String()
}

func (m kanbanModel) renderColumn(st taskState, width int) string {
	dim := lipgloss.NewStyle().Foreground(colDim)
	accent := lipgloss.NewStyle().Foreground(colBrightRed)
	doingTitle := lipgloss.NewStyle().Foreground(colBrightRed).Bold(true)
	doneTitle := lipgloss.NewStyle().Foreground(colDim).Strikethrough(true)
	cursorMark := lipgloss.NewStyle().Foreground(colBrightRed).Bold(true).Render("▸")

	titleStyle := lipgloss.NewStyle().Foreground(colDim).Bold(true)
	if st == m.focus {
		titleStyle = lipgloss.NewStyle().Foreground(colBrightRed).Bold(true)
	}

	var b strings.Builder
	heading := fmt.Sprintf("%s  (%d)", strings.ToUpper(st.label()), len(m.cols[st]))
	b.WriteString(titleStyle.Render(heading))
	b.WriteString("\n")
	b.WriteString(dim.Render(strings.Repeat("─", width-2)))
	b.WriteString("\n")

	if len(m.cols[st]) == 0 {
		b.WriteString(dim.Render("  (empty)"))
		b.WriteString("\n")
	}

	for i, t := range m.cols[st] {
		marker := "  "
		if st == m.focus && i == m.cursor[st] {
			marker = cursorMark + " "
		}
		id := ""
		if ref := t.ref(); ref != "" {
			id = ref + " "
		}
		title := t.Title
		switch t.State {
		case stateDoing:
			title = doingTitle.Render(title)
		case stateDone:
			title = doneTitle.Render(title)
		}
		line := marker + accent.Render(id) + title
		b.WriteString(truncateANSI(line, width-2))
		b.WriteString("\n")
		if t.Project != "" || t.Due != "" {
			meta := ""
			if t.Project != "" {
				meta += "#" + t.Project
			}
			if t.Due != "" {
				if meta != "" {
					meta += "  "
				}
				meta += "due " + t.Due
			}
			b.WriteString(dim.Render(truncateANSI("    "+meta, width-2)))
			b.WriteString("\n")
		}
	}

	border := lipgloss.RoundedBorder()
	borderColor := colDim
	if st == m.focus {
		borderColor = colBrightRed
	}
	return lipgloss.NewStyle().
		Width(width).
		Border(border).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(b.String())
}

// truncateANSI is a width-safe truncate using lipgloss.Width to handle ANSI
// escape sequences in the source string.
func truncateANSI(s string, max int) string {
	if lipgloss.Width(s) <= max {
		return s
	}
	return lipgloss.NewStyle().MaxWidth(max).Render(s)
}

// ---------------------------------------------------------------------------
// Gantt (Phase 8 / C) — `thakkali gantt -view week|month|year`
// ---------------------------------------------------------------------------

func runGantt(args []string) {
	fs := flag.NewFlagSet("gantt", flag.ExitOnError)
	view := fs.String("view", "month", "week | month | year")
	fs.StringVar(view, "v", "month", "shorthand for -view")
	_ = fs.Parse(args)

	path, entries, err := readTaskFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading task file: %v\n", err)
		os.Exit(1)
	}

	type ganttRow struct {
		t              *task
		start, end     time.Time // planned window (start..due)
		actStart, actE time.Time // actual window (begin..done|now)
		hasPlanned     bool
		hasActual      bool
	}
	var rows []ganttRow
	for i := range entries {
		t := entries[i].task
		if t == nil {
			continue
		}
		var r ganttRow
		r.t = t
		if s, ok := parseDateOnly(t.Start); ok {
			r.start = s
			r.hasPlanned = true
		}
		if d, ok := parseDateOnly(t.Due); ok {
			r.end = d
			r.hasPlanned = true
		}
		if r.hasPlanned {
			if r.start.IsZero() {
				r.start = r.end
			}
			if r.end.IsZero() {
				r.end = r.start
			}
		}
		if !t.Begin.IsZero() {
			r.actStart = t.Begin.Local()
			r.hasActual = true
		}
		if r.hasActual {
			if !t.Done.IsZero() {
				r.actE = t.Done.Local()
			} else {
				r.actE = time.Now()
			}
		}
		if r.hasPlanned || r.hasActual {
			rows = append(rows, r)
		}
	}

	if len(rows) == 0 {
		fmt.Println("no tasks with @start / @due / @begin / @done — nothing to plot.")
		fmt.Printf("file: %s\n", path)
		return
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var winStart, winEnd time.Time
	var width int
	switch *view {
	case "week", "w":
		winStart = today.AddDate(0, 0, -3)
		winEnd = today.AddDate(0, 0, 11) // 14-day window
		width = 56
	case "year", "y":
		winStart = today.AddDate(0, 0, -30)
		winEnd = today.AddDate(0, 11, 0) // ~12 months
		width = 96
	default: // month
		winStart = today.AddDate(0, 0, -7)
		winEnd = today.AddDate(0, 0, 38) // ~45 days
		width = 78
	}

	// Sort rows by planned start (or actual start if no plan).
	sort.SliceStable(rows, func(i, j int) bool {
		ai := rows[i].start
		if ai.IsZero() {
			ai = rows[i].actStart
		}
		aj := rows[j].start
		if aj.IsZero() {
			aj = rows[j].actStart
		}
		return ai.Before(aj)
	})

	header := lipgloss.NewStyle().Foreground(colRed).Bold(true)
	dim := lipgloss.NewStyle().Foreground(colDim)
	accent := lipgloss.NewStyle().Foreground(colBrightRed)
	plannedStyle := lipgloss.NewStyle().Foreground(colDarkRed)
	actualStyle := lipgloss.NewStyle().Foreground(colBrightRed).Bold(true)
	overlapStyle := lipgloss.NewStyle().Foreground(colBrightRed).Bold(true)
	todayStyle := lipgloss.NewStyle().Foreground(colGreen).Bold(true)

	const labelW = 32
	pad := strings.Repeat(" ", labelW)

	fmt.Println()
	fmt.Println(header.Render(fmt.Sprintf("THAKKALI · gantt (%s view)", strings.ToLower(*view))))
	fmt.Println(dim.Render("file: " + path))
	fmt.Println()

	// Header axis: print a date label every ~10 cols.
	axis := strings.Repeat(" ", width)
	axisRunes := []rune(axis)
	step := width / 6
	if step < 8 {
		step = 8
	}
	totalDays := winEnd.Sub(winStart).Hours() / 24
	for c := 0; c < width; c += step {
		d := winStart.AddDate(0, 0, int(float64(c)/float64(width)*totalDays))
		label := d.Format("Jan 02")
		for i, r := range label {
			if c+i < len(axisRunes) {
				axisRunes[c+i] = r
			}
		}
	}
	fmt.Println(pad + dim.Render(string(axisRunes)))

	// Tick row + today marker.
	tick := []rune(strings.Repeat("─", width))
	for c := 0; c < width; c += step {
		if c < len(tick) {
			tick[c] = '┬'
		}
	}
	todayCol := -1
	if !today.Before(winStart) && !today.After(winEnd) {
		todayCol = int(today.Sub(winStart).Hours() / 24 / totalDays * float64(width))
		if todayCol >= 0 && todayCol < len(tick) {
			tick[todayCol] = '│'
		}
	}
	fmt.Println(pad + dim.Render(string(tick[:todayCol+1])) + func() string {
		if todayCol < 0 {
			return ""
		}
		return todayStyle.Render(string(tick[todayCol])) + dim.Render(string(tick[todayCol+1:]))
	}())

	for _, r := range rows {
		label := truncateANSI(taskLabel(r.t), labelW-1)
		// pad label to labelW visually
		gap := labelW - lipgloss.Width(label)
		if gap < 1 {
			gap = 1
		}
		bar := makeGanttBar(width, winStart, winEnd, r.start, r.end, r.actStart, r.actE,
			plannedStyle, actualStyle, overlapStyle, todayCol, todayStyle)
		meta := ganttMeta(r.t, accent, dim)
		fmt.Println(label + strings.Repeat(" ", gap) + bar + "   " + meta)
	}
	fmt.Println()
	fmt.Println(dim.Render("legend: ") + plannedStyle.Render("█ planned") + dim.Render("  ") +
		actualStyle.Render("█ actual") + dim.Render("  ") + todayStyle.Render("│ today"))
	fmt.Println()
}

func taskLabel(t *task) string {
	id := ""
	if ref := t.ref(); ref != "" {
		id = ref + " "
	}
	return id + t.Title
}

func ganttMeta(t *task, accent, dim lipgloss.Style) string {
	parts := []string{}
	if t.Project != "" {
		parts = append(parts, accent.Render("#"+t.Project))
	}
	if t.Start != "" || t.Due != "" {
		s := t.Start
		if s == "" {
			s = "?"
		}
		d := t.Due
		if d == "" {
			d = "?"
		}
		parts = append(parts, dim.Render(s+" → "+d))
	}
	parts = append(parts, dim.Render("["+t.State.label()+"]"))
	return strings.Join(parts, " ")
}

func makeGanttBar(
	width int, winStart, winEnd, plStart, plEnd, acStart, acEnd time.Time,
	plannedStyle, actualStyle, overlapStyle lipgloss.Style,
	todayCol int, todayStyle lipgloss.Style,
) string {
	totalDays := winEnd.Sub(winStart).Hours() / 24
	if totalDays <= 0 {
		totalDays = 1
	}
	colOf := func(t time.Time) int {
		days := t.Sub(winStart).Hours() / 24
		col := int(days / totalDays * float64(width))
		if col < 0 {
			col = 0
		}
		if col >= width {
			col = width - 1
		}
		return col
	}

	// 0=empty, 1=planned, 2=actual, 3=overlap
	cells := make([]int, width)
	if !plStart.IsZero() && !plEnd.IsZero() {
		s, e := colOf(plStart), colOf(plEnd)
		for i := s; i <= e; i++ {
			cells[i] |= 1
		}
	}
	if !acStart.IsZero() && !acEnd.IsZero() {
		s, e := colOf(acStart), colOf(acEnd)
		for i := s; i <= e; i++ {
			cells[i] |= 2
		}
	}

	var b strings.Builder
	for i, v := range cells {
		ch := " "
		switch v {
		case 1:
			ch = plannedStyle.Render("█")
		case 2:
			ch = actualStyle.Render("█")
		case 3:
			ch = overlapStyle.Render("█")
		}
		if i == todayCol && v == 0 {
			ch = todayStyle.Render("│")
		}
		b.WriteString(ch)
	}
	return b.String()
}

func parseDateOnly(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// ---------------------------------------------------------------------------
// Activity heatmap (Phase 8 / C) — `thakkali activity`
// ---------------------------------------------------------------------------

func runActivity(args []string) {
	fs := flag.NewFlagSet("activity", flag.ExitOnError)
	weeks := fs.Int("weeks", 52, "weeks back from today")
	fs.IntVar(weeks, "w", 52, "shorthand for -weeks")
	_ = fs.Parse(args)
	if *weeks < 4 {
		*weeks = 4
	}
	if *weeks > 104 {
		*weeks = 104
	}

	entries := readLog()
	perDay := make(map[string]int)
	for _, e := range entries {
		t, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			continue
		}
		key := t.Local().Format("2006-01-02")
		perDay[key] += e.DurationSec
	}

	now := time.Now().Local()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// Anchor the rightmost column on the Saturday of this week.
	daysToSat := (6 - int(today.Weekday()) + 7) % 7
	endSat := today.AddDate(0, 0, daysToSat)
	startSun := endSat.AddDate(0, 0, -7*(*weeks)+1)

	// Find peak intensity for color scaling.
	maxSec := 0
	for d := startSun; !d.After(endSat); d = d.AddDate(0, 0, 1) {
		if v := perDay[d.Format("2006-01-02")]; v > maxSec {
			maxSec = v
		}
	}

	header := lipgloss.NewStyle().Foreground(colRed).Bold(true)
	dim := lipgloss.NewStyle().Foreground(colDim)
	accent := lipgloss.NewStyle().Foreground(colBrightRed)

	level := func(sec int) string {
		if sec == 0 {
			return dim.Render("·")
		}
		ratio := float64(sec) / float64(maxSec)
		switch {
		case ratio < 0.25:
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#5C0A0A")).Render("■")
		case ratio < 0.5:
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#9C2222")).Render("■")
		case ratio < 0.75:
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#D63333")).Render("■")
		default:
			return lipgloss.NewStyle().Foreground(colBrightRed).Bold(true).Render("■")
		}
	}

	fmt.Println()
	fmt.Println(header.Render(fmt.Sprintf("THAKKALI · activity (%d weeks)", *weeks)))
	fmt.Println(dim.Render(fmt.Sprintf("range: %s → %s", startSun.Format("2006-01-02"), endSat.Format("2006-01-02"))))
	fmt.Println()

	// Month label row above the grid. Each column = 1 week, 2 chars per cell
	// ("■ "). Place a month abbrev at the column for that week's Sunday if
	// it falls in the first 7 days of the month.
	monthRow := strings.Repeat(" ", *weeks*2)
	monthRunes := []rune(monthRow)
	for w := 0; w < *weeks; w++ {
		colSun := startSun.AddDate(0, 0, w*7)
		if colSun.Day() <= 7 {
			label := colSun.Format("Jan")
			for i, r := range label {
				pos := w*2 + i
				if pos < len(monthRunes) {
					monthRunes[pos] = r
				}
			}
		}
	}
	const dayLabelW = 5
	fmt.Print(strings.Repeat(" ", dayLabelW))
	fmt.Println(dim.Render(string(monthRunes)))

	// 7 rows: Sun..Sat
	dayLabels := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	for d := 0; d < 7; d++ {
		// Show every other label to reduce noise.
		lbl := "    "
		if d%2 == 1 {
			lbl = dim.Render(dayLabels[d]) + " "
		} else {
			lbl = "    "
		}
		fmt.Print(lbl)
		var b strings.Builder
		for w := 0; w < *weeks; w++ {
			cellDate := startSun.AddDate(0, 0, w*7+d)
			if cellDate.After(today) {
				b.WriteString("  ")
				continue
			}
			sec := perDay[cellDate.Format("2006-01-02")]
			b.WriteString(level(sec) + " ")
		}
		fmt.Println(b.String())
	}

	// Footer: legend + totals.
	fmt.Println()
	totalSec := 0
	activeDays := 0
	for d := startSun; !d.After(today); d = d.AddDate(0, 0, 1) {
		if v := perDay[d.Format("2006-01-02")]; v > 0 {
			totalSec += v
			activeDays++
		}
	}
	fmt.Printf("%s %s   %s %s\n",
		dim.Render("total:"), accent.Render(fmtDur(totalSec)),
		dim.Render("active days:"), accent.Render(fmt.Sprintf("%d", activeDays)))

	// Legend.
	fmt.Println()
	fmt.Print(dim.Render("less "))
	fmt.Print(level(0) + " ")
	fmt.Print(level(maxSec/8 + 1) + " ")
	fmt.Print(level(maxSec/3 + 1) + " ")
	fmt.Print(level(maxSec*2/3 + 1) + " ")
	fmt.Print(level(maxSec) + " ")
	fmt.Println(dim.Render("more"))
	fmt.Println()
}
