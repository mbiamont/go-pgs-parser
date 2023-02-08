package main

import (
	"fmt"
	"github.com/mbiamont/go-pgs-parser/pgs"
	"os"
	"time"
)

func main() {
	parser := pgs.NewPgsParser()

	err := parser.ConvertToPngImages("./example/input.sup", func(index int, startTime time.Duration) (*os.File, error) {
		return os.Create(fmt.Sprintf("./example/subs/input.%d.%s.png", index, formatTimeCode(startTime)))
	})
	if err != nil {
		panic(err)
	}
}

func formatTimeCode(timeCode time.Duration) string {
	hours := int64(timeCode.Hours())
	timeCode -= time.Duration(hours * int64(time.Hour))

	minutes := int64(timeCode.Minutes())
	timeCode -= time.Duration(minutes * int64(time.Minute))

	seconds := int64(timeCode.Seconds())
	timeCode -= time.Duration(seconds * int64(time.Second))

	millis := timeCode.Milliseconds()

	return fmt.Sprintf("%02d-%02d-%02d-%03d", hours, minutes, seconds, millis)
}
