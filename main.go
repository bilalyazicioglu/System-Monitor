package main

import (
	"fmt"
	"os"
	"os/user"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// periyodij update
type tickMsg time.Time

// TUI durumu
type model struct {
	// Kullanıcı bilgileri
	username string
	hostname string
	os       string
	platform string
	uptime   uint64
	// CPU
	cpuUsage float64
	// RAM
	ramTotal     uint64
	ramUsed      uint64
	ramAvailable uint64
	ramPercent   float64
}

func initialModel() model {
	m := model{}
	// user info al (constant)
	if u, err := user.Current(); err == nil {
		m.username = u.Username
	}
	// host info al (constant)
	if info, err := host.Info(); err == nil {
		m.hostname = info.Hostname
		m.os = friendlyOSName(info.OS)
		m.platform = info.Platform + " " + info.PlatformVersion
	}
	return m
}

func friendlyOSName(os string) string {
	switch os {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return os
	}
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

// uptime al
func getUptime() uint64 {
	uptime, _ := host.Uptime()
	return uptime
}

// uptime'ı okunabilir formata çevir
func formatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
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
		m.uptime = getUptime()
		return m, tickCmd()
	}
	return m, nil
}

func (m model) View() string {
	s := "╔════════════════════════════════════════╗\n"
	s += "║           SYSTEM MONITOR               ║\n"
	s += "╠════════════════════════════════════════╣\n"
	s += fmt.Sprintf("║  Kullanıcı:  %-25s ║\n", m.username)
	s += fmt.Sprintf("║  Hostname:   %-25s ║\n", m.hostname)
	s += fmt.Sprintf("║  OS:         %-25s ║\n", m.os+" "+m.platform)
	s += fmt.Sprintf("║  Uptime:     %-25s ║\n", formatUptime(m.uptime))
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
