package main

import (
	"log"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/prometheus/procfs"
	blockdevice "github.com/prometheus/procfs/blockdevice"
)

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

	memGauge.Percent = min(100, int(*meminfo.MemAvailable*100 / *meminfo.MemTotal))
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

	cpuInfo := getStat(fSystem)

	cpuGauge.Percent = min(100, int(cpuInfo.CPUTotal.User+cpuInfo.CPUTotal.System))
	cpuGauge.SetRect(40, 4, 30, 3)

	if cpuGauge.Percent < 50 {
		cpuGauge.BarColor = ui.ColorGreen
	} else if cpuGauge.Percent < 75 {
		cpuGauge.BarColor = ui.ColorYellow
	} else {
		cpuGauge.BarColor = ui.ColorRed
	}

	diskGauge := widgets.NewGauge()

	diskGauge.Title = "Disk Usage"

	diskStats := getDiskStats()

	diskGauge.Percent = min(100, int(diskStats[0].IOsInProgress))
	diskGauge.SetRect(80, 4, 30, 3)

	if diskGauge.Percent < 50 {
		diskGauge.BarColor = ui.ColorGreen
	} else if diskGauge.Percent < 75 {
		diskGauge.BarColor = ui.ColorYellow
	} else {
		diskGauge.BarColor = ui.ColorRed
	}

	grid.Set(ui.NewRow(
		1.0,
		ui.NewCol(0.2, memGauge),
		ui.NewCol(0.2, cpuGauge),
		ui.NewCol(0.2, diskGauge),
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
			switch event.ID { // event string/identifier
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
