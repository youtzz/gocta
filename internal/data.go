package internal

import (
	"container/list"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	expb "newgitlab.com/xquant/exchange-protocols/protocols/src/go"
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

func (d *Data) GetBarData(symbol string, _ Exchange, interval time.Duration, start, end time.Time) *list.List {
	rows, err := d.db.Raw("select * from bars_btcusdt_1d where datetime between ? and ?;", start, end).Rows()
	if err != nil {
		panic(err)
		return nil
	}

	list := list.New()
	for rows.Next() {
		var bar expb.BarData
		if err := rows.Scan(&bar); err != nil {
			panic(err)
			return nil
		}

		list.PushBack(bar)
	}

	return list
}

func (d *Data) GetTickData(symbol string, exchange Exchange, interval time.Duration, start, end time.Time) *list.List {
	panic("implement me")
}
