package internal

import (
	"fmt"
	"testing"
	"time"
)

func TestNewData(t *testing.T) {
	data := NewData()
	list, err := data.GetBarData("", 1, time.Hour, time.Date(2022, 1, 1, 1, 1, 1, 1, time.Local), time.Now())
	if err != nil {
		panic(err)
	}

	for cur := list.Front(); cur.Next() != nil; cur = cur.Next() {
		bar := cur.Value.(BarData)
		fmt.Printf("%+v\n", bar.UpdatedAt.AsTime().String())
	}
}
