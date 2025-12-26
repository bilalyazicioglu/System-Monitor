package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// tickMsg periyodik güncelleme için kullanılır
type tickMsg time.Time

// model TUI durumunu tutar
type model struct {
	cpuUsage     float64 // CPU kullanım yüzdesi
	ramTotal     uint64  // Toplam RAM (bytes)
	ramUsed      uint64  // Kullanılan RAM (bytes)
	ramAvailable uint64  // Kullanılabilir RAM (bytes)
	ramPercent   float64 // RAM kullanım yüzdesi
}

func initialModel() model {
	return model{}
}

// tickCmd her saniye bir tick mesajı gönderir
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// cpu kullanımı al
func getCPUUsage() float64 {
	percentages, err := cpu.Percent(200*time.Millisecond, false)
	if err != nil || len(percentages) == 0 {
		return 0.0
	}
	return percentages[0]
}

// ram kullanımı al
func getRAMUsage() (total, used, available uint64, percent float64) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, 0, 0
	}
	// cache ve buffer'ları hariç tutar
	actualUsed := v.Total - v.Available
	actualPercent := (float64(actualUsed) / float64(v.Total)) * 100
	return v.Total, actualUsed, v.Available, actualPercent
}

// byte -> GB
func bytesToGB(bytes uint64) float64 {
	return float64(bytes) / (1024 * 1024 * 1024)
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tickMsg:
		// tick başı güncelle
		m.cpuUsage = getCPUUsage()
		m.ramTotal, m.ramUsed, m.ramAvailable, m.ramPercent = getRAMUsage()
		return m, tickCmd()
	}
	return m, nil
}

func (m model) View() string {
	s := "╔════════════════════════════════════════╗\n"
	s += "║           SYSTEM MONITOR               ║\n"
	s += "╠════════════════════════════════════════╣\n"
	s += fmt.Sprintf("║  CPU Kullanımı:  %6.2f%%               ║\n", m.cpuUsage)
	s += "╠════════════════════════════════════════╣\n"
	s += fmt.Sprintf("║  RAM Toplam:     %6.2f GB             ║\n", bytesToGB(m.ramTotal))
	s += fmt.Sprintf("║  RAM Kullanılan: %6.2f GB (%5.1f%%)    ║\n", bytesToGB(m.ramUsed), m.ramPercent)
	s += fmt.Sprintf("║  RAM Kullanılabilir: %6.2f GB         ║\n", bytesToGB(m.ramAvailable))
	s += "╚════════════════════════════════════════╝\n"
	s += "\n'q' tuşuna basarak çıkabilirsiniz.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
