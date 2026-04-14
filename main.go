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
	task      string
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
			sections = append(sections, statusStyle.Render("task: "+m.task))
		}
		sections = append(sections, "")
	case m.pomodoro:
		sections = append(sections, m.phaseStyle().Render(m.phaseLabel()))
		if m.task != "" && m.phase == phaseWork {
			sections = append(sections, statusStyle.Render("task: "+m.task))
		}
		sections = append(sections, "")
	default:
		if m.task != "" {
			sections = append(sections, statusStyle.Render("task: "+m.task), "")
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
			tasks[e.Task] += e.DurationSec
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

	fmt.Printf("%s %s\n\n",
		dim.Render("all-time "+noun+":"),
		accent.Render(fmtDur(allTimeTotal)))
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

	buildFrames()

	work := time.Duration(*workMin) * time.Minute
	m := model{
		durations: [3]time.Duration{
			work,
			time.Duration(*shortMin) * time.Minute,
			time.Duration(*longMin) * time.Minute,
		},
		rounds:    *rounds,
		task:      *task,
		sound:     *sound,
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
