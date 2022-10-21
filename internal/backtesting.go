package internal

import (
	"container/list"
	"log"
	"time"
)

type BackTestingMod int

const (
	BAR = BackTestingMod(iota)
	TICK
)

type DataRepo interface {
	GetBarData(symbol string, exchange Exchange, interval time.Duration, start, end time.Time) *list.List
	GetTickData(symbol string, exchange Exchange, interval time.Duration, start, end time.Time) *list.List
}

type engineCfg struct {
	symbol     string
	start, end time.Time
	interval   time.Duration
	exchange   Exchange

	BackTestingMod
}

type Logger interface {
	Println(v ...any)
}

type defaultLogger struct {
}

func (defaultLogger) Println(v ...any) {
	log.Println(v...)
}

type BackTestingEngine struct {
	engineCfg
	cta Strategy

	logger Logger
	repo   DataRepo

	historyData list.List
}

func NewBackTestingEngine(repo DataRepo) *BackTestingEngine {
	return nil
}

func (b *BackTestingEngine) SetStrategy(strategy Strategy) {
	b.cta = strategy
}

func (b *BackTestingEngine) Run() {
	b.loadData()
}

func (b *BackTestingEngine) loadData() {
	b.logger.Println("开始加载历史数据")

	if b.end.IsZero() {
		b.end = time.Now()
	}

	if !b.start.Before(b.end) {
		b.logger.Println("起始日期必须小于结束日期")
		return
	}

	b.historyData.Init() // 清除之前的数据

	totalDays := int(b.end.Sub(b.start).Hours() / 24)
	progressDays := Max(totalDays/10, 1)
	progressDelta := time.Duration(progressDays * 24)
	intervalDelta := b.interval

	start := b.start
	end := b.end.Add(progressDelta)
	progress := 0

	var howToLoad func(string, Exchange, time.Duration, time.Time, time.Time) *list.List
	if b.BackTestingMod == BAR {
		howToLoad = b.repo.GetBarData
	} else {
		howToLoad = b.repo.GetTickData
	}

	for start.Before(end) {
		loaded := howToLoad(b.symbol, b.exchange, b.interval, start, end)
		b.historyData.PushBackList(loaded)

		progress += progressDays / totalDays
		progress = Min(progress, 1)

		start = end.Add(intervalDelta)
		end = end.Add(progressDelta)
	}

	b.logger.Println("历史数据加载完成，数据量:", b.historyData.Len())
}
