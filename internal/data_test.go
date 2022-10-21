package internal

import (
	"testing"
	"time"
)

func TestNewData(t *testing.T) {
	data := NewData()
	list := data.GetBarData("", 1, time.Hour, time.Date(2022, 1, 1, 1, 1, 1, 1, time.Local), time.Now())
	println(list.Len())
}
