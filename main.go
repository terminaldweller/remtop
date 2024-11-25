package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/prometheus/procfs"
	blockdevice "github.com/prometheus/procfs/blockdevice"
	"golang.org/x/sys/unix"
)

var prevStat procfs.Stat

func formatWithUnderscore(number int) string {
	numberStr := strconv.FormatInt(int64(number), 10)

	var formattedNumber string

	for i, digit := range numberStr {
		if i > 0 && (len(numberStr)-i)%3 == 0 {
			formattedNumber += "_"
		}

		formattedNumber += string(digit)
	}

	return formattedNumber
}

func BytesToString(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0

	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit

		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func convertSeconds(seconds int) (days, hours, minutes int) {
	days = seconds / (60 * 60 * 24)
	seconds = seconds % (60 * 60 * 24)
	hours = seconds / (60 * 60)
	seconds = seconds % (60 * 60)
	minutes = seconds / 60
	return days, hours, minutes
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func getDiskStats() []blockdevice.Diskstats {
	bdevice, err := blockdevice.NewFS("/proc", "/sys")
	if err != nil {
		log.Println(err)

		return nil
	}

	diskStats, err := bdevice.ProcDiskstats()
	if err != nil {
		log.Println(err)

		return nil
	}

	return diskStats
}

func getStat(fs procfs.FS) *procfs.Stat {
	statInfo, err := fs.Stat()
	if err != nil {
		log.Println(err)

		return nil
	}

	return &statInfo
}

func getmeminfo(fs procfs.FS) *procfs.Meminfo {
	meminfo, err := fs.Meminfo()
	if err != nil {
		log.Println(err)

		return nil
	}

	return &meminfo
}

func getUptime() (int, int, int) {
	uptime, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Println(err)

		return 0, 0, 0
	}

	uptimeStr := string(uptime)
	uptimeStr = strings.TrimSuffix(uptimeStr, "\n")
	uptimeSplit := strings.Split(uptimeStr, " ")

	uptimeTotal, err := strconv.ParseFloat(uptimeSplit[0], 64)
	if err != nil {
		log.Println(err)

		return 0, 0, 0
	}

	days, hours, minutes := convertSeconds(int(uptimeTotal))

	return days, hours, minutes
}

func drawFunction() {
	termWidth, termHeight := ui.TerminalDimensions()

	fSystem, err := procfs.NewFS("/proc")
	if err != nil {
		log.Println(err)

		return
	}

	grid := ui.NewGrid()
	grid.SetRect(1, 1, termWidth-3, termHeight-3)

	memGauge := widgets.NewGauge()
	memGauge.Border = true
	memGauge.TitleStyle = ui.NewStyle(ui.ColorGreen)
	memGauge.BorderStyle.Fg = ui.ColorCyan

	memGauge.Title = "Memory Usage"

	meminfo := getmeminfo(fSystem)

	memGauge.Percent = min(100, 100-int(*meminfo.MemAvailable*100 / *meminfo.MemTotal))

	if memGauge.Percent < 50 {
		memGauge.BarColor = ui.ColorGreen
	} else if memGauge.Percent < 75 {
		memGauge.BarColor = ui.ColorYellow
	} else {
		memGauge.BarColor = ui.ColorRed
	}

	cpuGauge := widgets.NewGauge()

	cpuGauge.Title = "CPU Usage"
	cpuGauge.BorderStyle = ui.NewStyle(ui.ColorCyan)
	cpuGauge.TitleStyle = ui.NewStyle(ui.ColorGreen)

	stat := getStat(fSystem)

	PrevIdle := prevStat.CPUTotal.Idle + prevStat.CPUTotal.Iowait
	Idle := stat.CPUTotal.Idle + stat.CPUTotal.Iowait

	prevNonIdle := prevStat.CPUTotal.User + prevStat.CPUTotal.Nice + prevStat.CPUTotal.System + prevStat.CPUTotal.IRQ + prevStat.CPUTotal.SoftIRQ + prevStat.CPUTotal.Steal
	nonIdle := stat.CPUTotal.User + stat.CPUTotal.Nice + stat.CPUTotal.System + stat.CPUTotal.IRQ + stat.CPUTotal.SoftIRQ + stat.CPUTotal.Steal

	prevTotal := PrevIdle + prevNonIdle
	total := Idle + nonIdle

	totald := total - prevTotal
	idled := Idle - PrevIdle

	cpuGauge.Percent = min(100, int(100*(totald-idled)/float64(totald)))

	cpuGauge.Border = true

	if cpuGauge.Percent < 50 {
		cpuGauge.BarColor = ui.ColorGreen
	} else if cpuGauge.Percent < 75 {
		cpuGauge.BarColor = ui.ColorYellow
	} else {
		cpuGauge.BarColor = ui.ColorRed
	}

	prevStat = *stat

	ioWait := widgets.NewGauge()

	ioWait.Title = "IO Wait"
	ioWait.BorderStyle = ui.NewStyle(ui.ColorCyan)
	ioWait.TitleStyle = ui.NewStyle(ui.ColorGreen)

	ioWait.Percent = min(100, int((stat.CPUTotal.Iowait-prevStat.CPUTotal.Iowait)/totald))

	if ioWait.Percent < 10 {
		ioWait.BarColor = ui.ColorGreen
	} else if ioWait.Percent < 50 {
		ioWait.BarColor = ui.ColorYellow
	} else {
		ioWait.BarColor = ui.ColorRed
	}

	diskParagraph := widgets.NewParagraph()

	diskParagraph.Title = "Disk Usage"
	diskParagraph.TextStyle = ui.NewStyle(ui.ColorBlue)
	diskParagraph.TitleStyle = ui.NewStyle(ui.ColorGreen)
	diskParagraph.BorderStyle = ui.NewStyle(ui.ColorCyan)

	diskStats := getDiskStats()
	diskParagraph.Border = true

	diskParagraph.Text = formatWithUnderscore(int(diskStats[2].WeightedIOTicks))

	uptimeParagraph := widgets.NewParagraph()

	uptimeParagraph.Title = "Uptime"

	uptimeParagraph.TitleStyle = ui.NewStyle(ui.ColorGreen)
	uptimeParagraph.BorderStyle = ui.NewStyle(ui.ColorCyan)

	days, hours, minutes := getUptime()

	uptimeParagraph.Text = strconv.Itoa(days) + "d " + strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m"

	if days >= 1 {
		uptimeParagraph.TextStyle = ui.NewStyle(ui.ColorGreen)
	} else {
		uptimeParagraph.TextStyle = ui.NewStyle(ui.ColorYellow)
	}

	netWid := widgets.NewParagraph()

	netWid.Title = "Network"

	netWid.TextStyle = ui.NewStyle(ui.ColorBlue)
	netWid.TitleStyle = ui.NewStyle(ui.ColorGreen)
	netWid.BorderStyle = ui.NewStyle(ui.ColorCyan)

	fNet, err := fSystem.NetDev()
	if err != nil {
		log.Println(err)
	}

	netWid.Text = BytesToString(int64(fNet.Total().RxBytes)) + "/" + BytesToString(int64(fNet.Total().TxBytes))

	loadAvg, err := fSystem.LoadAvg()
	if err != nil {
		log.Println(err)

		return
	}

	loadAvgParagraph := widgets.NewParagraph()

	loadAvgParagraph.Title = "Load Average"

	loadAvgParagraph.TextStyle = ui.NewStyle(ui.ColorBlue)
	loadAvgParagraph.TitleStyle = ui.NewStyle(ui.ColorGreen)
	loadAvgParagraph.BorderStyle = ui.NewStyle(ui.ColorCyan)

	loadAvgParagraph.Text = strconv.FormatFloat(loadAvg.Load1, 'f', 2, 64) + "/" + strconv.FormatFloat(loadAvg.Load5, 'f', 2, 64) + "/" + strconv.FormatFloat(loadAvg.Load15, 'f', 2, 64)

	entropyParagraph := widgets.NewParagraph()

	entropyParagraph.Title = "Entropy"

	entropyParagraph.TitleStyle = ui.NewStyle(ui.ColorGreen)
	entropyParagraph.BorderStyle = ui.NewStyle(ui.ColorCyan)

	rand, err := fSystem.KernelRandom()
	if err != nil {
		log.Println(err)
	}

	entropyParagraph.Text = strconv.FormatUint(*rand.EntropyAvaliable, 10)
	if *rand.EntropyAvaliable >= 256 {
		entropyParagraph.TextStyle = ui.NewStyle(ui.ColorGreen)
	} else {
		entropyParagraph.TextStyle = ui.NewStyle(ui.ColorYellow)
	}

	hostname := widgets.NewParagraph()

	hostname.Title = "hostname"

	hostname.Text, err = os.Hostname()
	if err != nil {
		log.Println(err)
	}

	hostname.TextStyle = ui.NewStyle(ui.ColorBlue)
	hostname.BorderStyle = ui.NewStyle(ui.ColorCyan)
	hostname.TitleStyle = ui.NewStyle(ui.ColorGreen)

	freeDiskSpace := widgets.NewParagraph()

	freeDiskSpace.Title = "Disk"
	freeDiskSpace.TextStyle = ui.NewStyle(ui.ColorBlue)
	freeDiskSpace.TitleStyle = ui.NewStyle(ui.ColorGreen)
	freeDiskSpace.BorderStyle = ui.NewStyle(ui.ColorCyan)

	var dStat unix.Statfs_t

	err = unix.Statfs("/", &dStat)
	if err != nil {
		log.Println(err)
	}

	freeDiskSpace.Text = BytesToString(int64(dStat.Bavail * uint64(dStat.Bsize)))

	grid.Set(
		ui.NewRow(
			.06,
			ui.NewCol(0.1, hostname),
			ui.NewCol(0.12, memGauge),
			ui.NewCol(0.12, cpuGauge),
			ui.NewCol(0.06, ioWait),
			ui.NewCol(0.06, freeDiskSpace),
			ui.NewCol(0.08, diskParagraph),
			ui.NewCol(0.08, uptimeParagraph),
			ui.NewCol(0.1, netWid),
			ui.NewCol(0.1, loadAvgParagraph),
			ui.NewCol(0.06, entropyParagraph),
		),
	)

	ui.Render(grid)
}

func main() {
	err := ui.Init()
	if err != nil {
		log.Println(err)

		return
	}
	defer ui.Close()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C

	for {
		select {
		case event := <-uiEvents:
			switch event.ID {
			case "q", "<C-c>":
				return
			case "<MouseLeft>":
			case "<Resize>":
			}
			switch event.Type {
			case ui.KeyboardEvent:
			}
		case <-ticker:
			drawFunction()
		}
	}
}
