package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/prometheus/procfs"
	blockdevice "github.com/prometheus/procfs/blockdevice"
)

var prevStat procfs.Stat

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

	// uptimeIdle, err := strconv.ParseFloat(uptimeSplit[1], 64)
	// if err != nil {
	// 	log.Println(err)

	// 	return 0, 0, 0
	// }

	days, hours, minutes := convertSeconds(int(uptimeTotal))

	return days, hours, minutes
}

func drawFunction() {
	fSystem, err := procfs.NewFS("/proc")
	if err != nil {
		log.Println(err)

		return
	}

	grid := ui.NewGrid()
	grid.SetRect(0, 0, 200, 3)

	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		log.Println(err)

		grid.Title = ""
	} else {
		grid.Title = string(data)
	}

	memGauge := widgets.NewGauge()

	memGauge.Title = "Memory Usage"

	meminfo := getmeminfo(fSystem)

	memGauge.Percent = min(100, 100-int(*meminfo.MemAvailable*100 / *meminfo.MemTotal))
	memGauge.SetRect(0, 0, 30, 3)

	if memGauge.Percent < 50 {
		memGauge.BarColor = ui.ColorGreen
	} else if memGauge.Percent < 75 {
		memGauge.BarColor = ui.ColorYellow
	} else {
		memGauge.BarColor = ui.ColorRed
	}

	cpuGauge := widgets.NewGauge()

	cpuGauge.Title = "CPU Usage"

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

	cpuGauge.SetRect(40, 4, 30, 3)

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

	ioWait.Percent = min(100, int((stat.CPUTotal.Iowait-prevStat.CPUTotal.Iowait)/totald))

	if ioWait.Percent < 50 {
		ioWait.BarColor = ui.ColorGreen
	} else if ioWait.Percent < 75 {
		ioWait.BarColor = ui.ColorYellow
	} else {
		ioWait.BarColor = ui.ColorRed
	}

	diskParagraph := widgets.NewParagraph()

	diskParagraph.Title = "Disk Usage"

	diskStats := getDiskStats()
	diskParagraph.SetRect(80, 4, 30, 3)

	diskParagraph.Text = strconv.FormatUint(diskStats[2].IOStats.ReadSectors*512, 10)

	uptimeParagraph := widgets.NewParagraph()

	uptimeParagraph.Title = "Uptime"

	days, hours, minutes := getUptime()

	uptimeParagraph.Text = strconv.Itoa(days) + "d " + strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m"

	netWid := widgets.NewParagraph()

	netWid.Title = "Network"

	grid.Set(ui.NewRow(
		1.0,
		ui.NewCol(0.16, memGauge),
		ui.NewCol(0.16, cpuGauge),
		ui.NewCol(0.08, ioWait),
		ui.NewCol(0.16, diskParagraph),
		ui.NewCol(0.1, uptimeParagraph),
		ui.NewCol(0.1, netWid),
	))

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
