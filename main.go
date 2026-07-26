package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateSetupBin state = iota
	stateSetupDir
	stateSetupPort
	stateSelectModel
	stateSelectCtx
	stateRunning
	stateSelectDrive
	stateSettingsMenu // New state for settings configuration management
)

var (
	contexts        = []string{"8192", "16384", "32768", "65536"}
	settingsOptions = []string{"🔧 Change llamafile Binary Path", "📁 Change Models Directory", "🔌 Change Server Port Allocation", "❌ Clear Config & Restart Wizard", "↩ Return to Model Selection"}

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFCC")).MarginBottom(1)
	statusStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#005FDF")).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	urlStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	loadingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900")).Italic(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#00FFFF")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	inputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#333333")).Padding(0, 1)
	folderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1C40F")).Bold(true)
	driveStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")).Bold(true)
)

type Config struct {
	LlamafilePath string `json:"llamafile_path"`
	ModelsDir     string `json:"models_dir"`
	Port          string `json:"port"`
}

type statusMsg string

type model struct {
	state            state
	savedState       state
	config           Config
	inputValue       string
	models           []string
	cursor           int
	selectedModel    string
	selectedCtx      string
	status           string
	isReady          bool
	currentBrowseDir string
	browseItems      []os.DirEntry
	availableDrives  []string
}

func initialModel() model {
	m := model{status: "Initializing..."}
	file, err := os.ReadFile("config.json")
	if err == nil {
		var cfg Config
		if json.Unmarshal(file, &cfg) == nil && cfg.Port != "" {
			m.config = cfg
			if m.loadModels() {
				m.state = stateSelectModel
				return m
			}
		}
	}

	dir, _ := os.Getwd()
	m.currentBrowseDir = dir
	m.state = stateSetupBin
	m.updateBrowseList()
	m.status = "Setup Mode - Select llamafile executable"
	return m
}

func (m *model) updateBrowseList() {
	entries, err := os.ReadDir(m.currentBrowseDir)
	if err != nil {
		return
	}

	m.browseItems = []os.DirEntry{}
	for _, entry := range entries {
		if m.state == stateSetupBin {
			if entry.IsDir() || strings.HasSuffix(strings.ToLower(entry.Name()), ".exe") || filepath.Separator == '/' {
				m.browseItems = append(m.browseItems, entry)
			}
		} else if m.state == stateSetupDir {
			if entry.IsDir() {
				m.browseItems = append(m.browseItems, entry)
			}
		}
	}
	m.cursor = 0
}

func (m *model) loadModels() bool {
	files, err := filepath.Glob(filepath.Join(m.config.ModelsDir, "*.gguf"))
	if err != nil || len(files) == 0 {
		m.models = []string{"No .gguf files found in configured directory!"}
		return false
	}

	m.models = []string{}
	for _, f := range files {
		m.models = append(m.models, filepath.Base(f))
	}
	m.status = "Idle - Ready to launch"
	return true
}

func (m *model) saveConfig() {
	data, _ := json.MarshalIndent(m.config, "", "  ")
	_ = os.WriteFile("config.json", data, 0644)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		m.status = string(msg)
		if strings.HasPrefix(string(msg), "Running") {
			m.isReady = true
		}
		return m, nil

	case tea.KeyMsg:
		if m.state != stateSetupPort {
			if msg.String() == "q" || msg.String() == "Q" || msg.String() == "ctrl+c" {
				shutdownBackgroundProcess()
				return m, tea.Quit
			}
		}

		// Trigger Settings view menu transition from home catalog index screen
		if m.state == stateSelectModel && (msg.String() == "s" || msg.String() == "S") {
			m.state = stateSettingsMenu
			m.cursor = 0
			m.status = "System Settings Layer Interface active"
			return m, nil
		}

		switch msg.String() {
		case "backspace":
			if m.state == stateSetupPort && len(m.inputValue) > 0 {
				m.inputValue = m.inputValue[:len(m.inputValue)-1]
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			max := 0
			parent := filepath.Dir(m.currentBrowseDir)
			hasParent := parent != m.currentBrowseDir

			switch m.state {
			case stateSettingsMenu:
				max = len(settingsOptions) - 1
			case stateSelectDrive:
				max = len(m.availableDrives) - 1
			case stateSelectModel:
				max = len(m.models) - 1
			case stateSelectCtx:
				max = len(contexts) - 1
			case stateSetupPort:
				max = 0
			case stateSetupBin, stateSetupDir:
				max = len(m.browseItems)
				if hasParent {
					max++
				}
			}
			if m.cursor < max {
				m.cursor++
			}

		case "enter":
			switch m.state {
			case stateSettingsMenu:
				switch m.cursor {
				case 0: // Change llamafile bin
					dir, _ := os.Getwd()
					m.currentBrowseDir = dir
					m.state = stateSetupBin
					m.updateBrowseList()
				case 1: // Change models path dir
					dir, _ := os.Getwd()
					m.currentBrowseDir = dir
					m.state = stateSetupDir
					m.updateBrowseList()
				case 2: // Change Port
					m.inputValue = m.config.Port
					m.state = stateSetupPort
				case 3: // Clear all configurations profiles profile variables
					_ = os.Remove("config.json")
					m.config = Config{}
					dir, _ := os.Getwd()
					m.currentBrowseDir = dir
					m.state = stateSetupBin
					m.updateBrowseList()
				case 4: // Return to listing menu
					m.loadModels()
					m.state = stateSelectModel
					m.cursor = 0
				}

			case stateSelectDrive:
				m.currentBrowseDir = m.availableDrives[m.cursor]
				m.state = m.savedState
				m.updateBrowseList()

			case stateSetupBin, stateSetupDir:
				parent := filepath.Dir(m.currentBrowseDir)
				hasParent := parent != m.currentBrowseDir

				if m.cursor == 0 {
					m.savedState = m.state
					m.availableDrives = getLogicalDrives()
					m.state = stateSelectDrive
					m.cursor = 0
					return m, nil
				}

				if hasParent && m.cursor == 1 {
					m.currentBrowseDir = parent
					m.updateBrowseList()
					return m, nil
				}

				virtualRows := 1
				if hasParent {
					virtualRows = 2
				}

				actualFileIdx := m.cursor - virtualRows
				if actualFileIdx < 0 || actualFileIdx >= len(m.browseItems) {
					return m, nil
				}

				item := m.browseItems[actualFileIdx]
				if item.IsDir() {
					m.currentBrowseDir = filepath.Join(m.currentBrowseDir, item.Name())
					m.updateBrowseList()
				} else if m.state == stateSetupBin {
					m.config.LlamafilePath = filepath.Join(m.currentBrowseDir, item.Name())
					// If adjusting settings step, save instantly and revert to parameters checklist context loop
					m.saveConfig()
					m.state = stateSettingsMenu
					m.cursor = 0
				}

			case stateSetupPort:
				portStr := strings.TrimSpace(m.inputValue)
				if portStr == "" {
					portStr = "8080"
				}
				m.config.Port = portStr
				m.inputValue = ""
				m.saveConfig()
				m.state = stateSettingsMenu
				m.cursor = 2

			case stateSelectModel:
				if len(m.models) == 1 && strings.Contains(m.models[0], "No .gguf") {
					return m, nil
				}
				m.selectedModel = m.models[m.cursor]
				m.state = stateSelectCtx
				m.cursor = 0

			case stateSelectCtx:
				m.selectedCtx = contexts[m.cursor]
				m.state = stateRunning
				m.isReady = false
				m.status = "Initializing backend process..."
				return m, m.startModelCmd()

			case stateRunning:
				m.shutdownModel()
				m.state = stateSelectModel
				m.cursor = 0
				m.status = "Idle - Model stopped"
			}

		case " ":
			if m.state == stateSetupDir {
				m.config.ModelsDir = m.currentBrowseDir
				m.saveConfig()
				m.state = stateSettingsMenu
				m.cursor = 1
			}

		default:
			if m.state == stateSetupPort {
				if len(msg.String()) == 1 {
					m.inputValue += msg.String()
				}
			}
		}
	}
	return m, nil
}

func mRenderRows(m model, i int, label string, s *strings.Builder) {
	if m.cursor == i {
		s.WriteString(selectedStyle.Render(fmt.Sprintf(" > %s ", label)) + "\n")
	} else {
		s.WriteString(fmt.Sprintf("   %s\n", label))
	}
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(statusStyle.Render(fmt.Sprintf(" STATUS: %s ", m.status)) + "\n")
	s.WriteString(strings.Repeat("─", 75) + "\n\n")

	switch m.state {
	case stateSettingsMenu:
		s.WriteString(titleStyle.Render("Configuration Engine Control Settings Panel Matrix") + "\n")
		s.WriteString(dimStyle.Render(fmt.Sprintf("Current Llamafile: %s", m.config.LlamafilePath)) + "\n")
		s.WriteString(dimStyle.Render(fmt.Sprintf("Models Folder:     %s", m.config.ModelsDir)) + "\n")
		s.WriteString(dimStyle.Render(fmt.Sprintf("Port Mapping:      %s", m.config.Port)) + "\n\n")

		for i, opt := range settingsOptions {
			mRenderRows(m, i, opt, &s)
		}

	case stateSelectDrive:
		s.WriteString(titleStyle.Render("Select System Root Mount Target / Drive Allocation Logical Space:") + "\n\n")
		for i, drive := range m.availableDrives {
			mRenderRows(m, i, driveStyle.Render("💽 "+drive), &s)
		}

	case stateSetupBin, stateSetupDir:
		promptText := "Step 1: Select the llamafile executable binary:"
		if m.state == stateSetupDir {
			promptText = "Step 2: Navigate to your Models folder and press [SPACEBAR] to select it:"
		}
		s.WriteString(titleStyle.Render(promptText) + "\n")
		s.WriteString(dimStyle.Render(fmt.Sprintf("Current Directory Location Trace: %s", m.currentBrowseDir)) + "\n\n")

		idx := 0
		mRenderRows(m, idx, driveStyle.Render("💽 [Switch Drive / Root Location Space]"), &s)
		idx++

		parent := filepath.Dir(m.currentBrowseDir)
		if parent != m.currentBrowseDir {
			mRenderRows(m, idx, "↩ .. [Go Back]", &s)
			idx++
		}

		for _, item := range m.browseItems {
			label := ""
			if item.IsDir() {
				label = folderStyle.Render("📁 " + item.Name() + "/")
			} else {
				label = "📄 " + item.Name()
			}
			mRenderRows(m, idx, label, &s)
			idx++
		}

	case stateSetupPort:
		s.WriteString(titleStyle.Render("Step 3: Assign server listening port allocation:") + "\n")
		s.WriteString(inputStyle.Render(m.inputValue+"█") + "\n")

	case stateSelectModel:
		s.WriteString(titleStyle.Render(fmt.Sprintf("Select a GGUF Model from %s:", m.config.ModelsDir)) + "\n")
		for i, item := range m.models {
			mRenderRows(m, i, item, &s)
		}

	case stateSelectCtx:
		s.WriteString(titleStyle.Render(fmt.Sprintf("Select context window size for %s:", m.selectedModel)) + "\n")
		for i, ctx := range contexts {
			mRenderRows(m, i, fmt.Sprintf("%s tokens", ctx), &s)
		}

	case stateRunning:
		s.WriteString(titleStyle.Render("LLM Service Running Active Background Thread") + "\n")
		if m.isReady {
			s.WriteString(fmt.Sprintf("Local Endpoint API Base URL: %s\n\n", urlStyle.Render(fmt.Sprintf("http://127.0.0.1:%s", m.config.Port))))
		} else {
			s.WriteString(loadingStyle.Render("🔄 Loading weights into system memory buffers (parsing RAM)...") + "\n\n")
		}
		s.WriteString(selectedStyle.Render(" 🛑 Press ENTER to safely stop model ") + "\n")
	}

	footerInstructions := "[↑/↓] Navigate | [ENTER] Choose/Step-In | [Q] Exit App"
	if m.state == stateSelectModel {
		footerInstructions = "[↑/↓] Navigate | [ENTER] Select Model | [S] Open Settings Menu | [Q] Exit"
	} else if m.state == stateSetupDir {
		footerInstructions = "[↑/↓] Navigate | [ENTER] Step-In | [SPACEBAR] Confirm Selected Directory | [Q] Exit"
	}
	s.WriteString("\n" + dimStyle.Render(footerInstructions+"\n"))
	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel())

	// Log the start of execution
	fmt.Println("Starting application...")

	if _, err := p.Run(); err != nil {
		errMsg := "Execution Error: %v"
		log.Printf(errMsg, err)
		os.Exit(1)
	}

	// Log successful termination
	fmt.Println("Application terminated successfully.")
}
