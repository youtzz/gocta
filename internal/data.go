package internal

import (
	"container/list"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type Data struct {
	db *gorm.DB
}

func NewData() DataRepo {
	db, err := gorm.Open(mysql.Open("xg:xg1234@tcp(10.8.0.121:3306)/binance_data?charset=utf8mb4&parseTime=True&loc=Local"))
	if err != nil {
		panic(err)
	}

	return &Data{db}
}

type LoadFunc func(string, Exchange, time.Duration, time.Time, time.Time) (*list.List, error)

func (d *Data) GetBarData(symbol string, _ Exchange, interval time.Duration, start, end time.Time) (*list.List, error) {
	var raw []dbBarData
	err := d.db.Raw("select * from bars_btcusdt_1d where datetime between ? and ?;", start, end).Scan(&raw).Error
	if err != nil {
		return nil, err
	}

	list := list.New()
	for i := 0; i < len(raw); i++ {
		list.PushBack(raw[i].ToPb())
	}
	return list, nil
}

func (d *Data) GetTickData(symbol string, exchange Exchange, interval time.Duration, start, end time.Time) (*list.List, error) {
	panic("implement me")
}

type dbBarData struct {
	Symbol       string
	Exchange     string
	Datetime     time.Time
	Volume       float64
	Turnover     float64
	OpenInterest float64
	OpenPrice    float64
	HighPrice    float64
	LowPrice     float64
	ClosePrice   float64
	Interval     string
}

func (d dbBarData) ToPb() BarData {
	return BarData{
		Symbol: d.Symbol,
		//Exchange:     d.Exchange,
		UpdatedAt:    timestamppb.New(d.Datetime),
		Volume:       d.Volume,
		Turnover:     d.Turnover,
		OpenInterest: d.OpenInterest,
		OpenPrice:    d.OpenPrice,
		HighPrice:    d.HighPrice,
		LowPrice:     d.LowPrice,
		ClosePrice:   d.ClosePrice,
		//Interval:     d.Interval,
	}
}
