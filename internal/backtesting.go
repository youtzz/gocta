package internal

import (
	"container/list"
	"errors"
	"log"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	expb "newgitlab.com/xquant/exchange-protocols/protocols/src/go"

	"gocta/utils"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

func Evaluate(cfg EngineCfg) error {
	engine, err := newEngine(cfg)
	if err != nil {
		engine.logger.Println("创建引擎失败:", err.Error())
		return err
	}

	err = engine.loadData()
	if err != nil {
		engine.logger.Println("加载数据失败:", err.Error())
		return err
	}

	engine.runBackTesting()
	engine.calculateResult()
	engine.calculateStatistics(false)

	return nil
}

type BackTestingMod int

const (
	BAR = BackTestingMod(iota)
	TICK
)

type DataRepo interface {
	GetBarData(symbol string, exchange Exchange, interval time.Duration, start, end time.Time) (*list.List, error)
	GetTickData(symbol string, exchange Exchange, interval time.Duration, start, end time.Time) (*list.List, error)
}

type EngineCfg struct {
	Strategy
	DataRepo

	Symbol     string
	Start, End time.Time
	Interval   time.Duration
	Exchange   Exchange
	Capital    float64
	Rate       float64
	Size       float64
	Slippage   float64
	Inverse    bool
	AnnualDays int
	BackTestingMod
}

func (c EngineCfg) Check() error {
	if c.Symbol == "" {
		return errors.New("symbol不能为空")
	}
	if c.Strategy == nil {
		return errors.New("strategy不能为空")
	}
	if c.DataRepo == nil {
		return errors.New("dataRepo不能为空")
	}
	if c.Start.IsZero() {
		return errors.New("start不能为空")
	}

	return nil
}

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

type defaultLogger struct {
}

func (defaultLogger) Println(v ...any) {
	log.Println(v...)
}

func (defaultLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

type BackTestingEngine struct {
	EngineCfg

	logger Logger

	historyData *list.List

	bar      BarData
	tick     TickData
	datetime time.Time

	activeLimitOrders map[string]*OrderData
	trades            map[string]*TradeData
	tradeCount        int
	dailyDf           *dataframe.DataFrame
	dailyResults      map[string]*DailyResult
}

func newEngine(cfg EngineCfg) (*BackTestingEngine, error) {
	if err := cfg.Check(); err != nil {
		return nil, err
	}
	return &BackTestingEngine{
		EngineCfg:         cfg,
		logger:            defaultLogger{},
		historyData:       list.New(),
		activeLimitOrders: make(map[string]*OrderData),
		trades:            make(map[string]*TradeData),
		dailyResults:      make(map[string]*DailyResult),
	}, nil
}

func (b *BackTestingEngine) runBackTesting() {
	b.Strategy.OnInit()
	// todo 缺了让策略提前获得部分交易数据的功能
	b.logger.Println("策略初始化完成")

	b.Strategy.OnStart()
	b.logger.Println("开始回放历史数据")

	if b.historyData.Len() <= 1 {
		b.logger.Println("历史数据不足，回测终止")
		return
	}

	// todo 没有回放进度的功能
	for cur := b.historyData.Front(); cur.Next() != nil; cur = cur.Next() {
		if b.BackTestingMod == BAR {
			b.newBar(cur.Value.(BarData))
		} else {
			b.newTick(cur.Value.(TickData))
		}
	}

	b.Strategy.OnStop()
	b.logger.Println("历史数据回放结束")
}

func (b *BackTestingEngine) clear() {

}

func (b *BackTestingEngine) loadData() error {
	b.logger.Println("开始加载历史数据")

	if b.End.IsZero() {
		b.End = time.Now()
	}

	if !b.Start.Before(b.End) {
		return errors.New("起始日期必须小于结束日期")
	}

	b.historyData.Init()

	totalDays := int(b.End.Sub(b.Start).Hours() / 24)
	progressDays := Max(totalDays/10, 1)
	progressDelta := time.Hour * time.Duration(progressDays*24)
	intervalDelta := b.Interval

	start := b.Start
	end := b.Start.Add(progressDelta)
	progress := 0

	var howToLoad LoadFunc
	if b.BackTestingMod == BAR {
		howToLoad = b.DataRepo.GetBarData
	} else {
		howToLoad = b.DataRepo.GetTickData
	}

	for start.Before(b.End) {
		// 确保时间范围
		if end.After(b.End) {
			end = b.End
		}

		loaded, err := howToLoad(b.Symbol, b.Exchange, b.Interval, start, end)
		if err != nil {
			return err
		}
		b.historyData.PushBackList(loaded)

		progress += progressDays / totalDays
		progress = Min(progress, 1)

		start = end.Add(intervalDelta)
		end = end.Add(progressDelta)
	}

	b.logger.Println("历史数据加载完成，数据量:", b.historyData.Len())

	return nil
}

func (b *BackTestingEngine) newBar(bar BarData) {
	b.bar = bar
	b.datetime = bar.UpdatedAt.AsTime()

	b.crossLimitOrder()
	//b.crossStopOrder()
	b.Strategy.OnBar(bar)
}

func (b *BackTestingEngine) newTick(tick TickData) {

}

func (b *BackTestingEngine) crossLimitOrder() {
	var (
		longCrossPrice  float64
		shortCrossPrice float64
		longBestPrice   float64
		shortBestPrice  float64
	)
	if b.BackTestingMod == BAR {
		longCrossPrice = b.bar.LowPrice
		shortCrossPrice = b.bar.HighPrice
		longBestPrice = b.bar.OpenPrice
		shortBestPrice = b.bar.OpenPrice
	} else {
		longCrossPrice = b.tick.AskPrice_1
		shortCrossPrice = b.tick.BidPrice_1
		longBestPrice = longCrossPrice
		shortBestPrice = shortCrossPrice
	}

	for _, order := range b.activeLimitOrders {
		longCross := (order.Direction == expb.Direction_LONG) &&
			(order.Price >= longCrossPrice) &&
			(longCrossPrice > 0)

		shortCross := (order.Direction == expb.Direction_SHORT) &&
			(order.Price <= shortCrossPrice) &&
			(shortCrossPrice > 0)

		if !longCross || !shortCross {
			continue
		}

		order.Traded = order.Volume
		order.Status = expb.Status_ALL_TRADED

		b.Strategy.OnOrder(*order)
		if _, ok := b.activeLimitOrders[order.OrderNo]; ok {
			delete(b.activeLimitOrders, order.OrderNo)
		}

		b.tradeCount++

		var (
			tradePrice float64
			posChange  float64
		)

		if longCross {
			tradePrice = Min(order.Price, longBestPrice)
			posChange = order.Volume
		} else {
			tradePrice = Min(order.Price, shortBestPrice)
			posChange = -order.Volume
		}

		trade := TradeData{
			Symbol:    order.Symbol,
			Exchange:  order.Exchange,
			OrderNo:   order.OrderNo,
			TradeNo:   strconv.Itoa(b.tradeCount),
			Direction: order.Direction,
			Offset:    order.Offset,
			Price:     tradePrice,
			Volume:    order.Volume,
			UpdatedAt: timestamppb.New(b.datetime),
			Reference: "",
			//GatewayName: 0,
		}

		b.Strategy.OnPosition(posChange)
		b.Strategy.OnTrade(trade)

		b.trades[trade.TradeNo] = &trade
	}
}

func (b *BackTestingEngine) calculateResult(){
	res := dataframe.New()
	if len(b.trades) == 0 {
		b.logger.Println("成交记录为空，无法计算")
		return
	}
	for _, trade := range b.trades {
		date := trade.UpdatedAt.AsTime().Format("2006-01-02")
		dailyResult := b.dailyResults[date]
		dailyResult.AddTrade(trade)
	}

	var preClose, startPos float64

	for _, dailyResult := range b.dailyResults {
		dailyResult.CalculatePnl(preClose, startPos, b.Rate, b.Slippage, b.Size, b.Inverse)
		preClose = dailyResult.ClosePrice
		startPos = dailyResult.StartPos
	}
	results := make([]*DailyResult, 0)
	for k, v := range b.dailyResults {
		v.Date = k
		results = append(results, v)
	}

	res = dataframe.LoadStructs(results)
	b.dailyDf = &res
}

func (b *BackTestingEngine) calculateStatistics(output bool) map[string]any {
	var df *dataframe.DataFrame
	b.logger.Println("开始计算策略统计指标")
	startDate := ""
	endDate := ""
	totalDays := 0
	profitDays := 0
	lossDays := 0
	endBalance := 0.0
	maxDrawdown := 0.0
	maxDdpercent := 0.0
	maxDrawdownDuration := 0
	totalNetPnl := 0.0
	dailyNetPnl := 0.0
	totalCommission := 0.0
	dailyCommission := 0.0
	totalSlippage := 0.0
	dailySlippage := 0.0
	totalTurnover := 0.0
	dailyTurnover := 0.0
	totalTradeCount := 0.0
	dailyTradeCount := 0.0
	totalReturn := 0.0
	annualReturn := 0.0
	dailyReturn := 0.0
	returnStd := 0.0
	sharpeRatio := 0.0
	returnDrawdownRatio := 0.0

	if df == nil {
		df = b.dailyDf
	}
	netPnlSeries := df.Col("NetPnl")
	balanceSeries := utils.CumSumSeries(netPnlSeries, "Balance", b.Capital)
	preBalanceSeries := utils.ShiftSeries(balanceSeries, 1, "PreBalance")
	preBalanceSeries.Set(0, series.New(b.Capital, series.Float, "PreBalance"))

	// TODO: Highlevel
	balanceSeries.Rolling(1).Mean()

	drawdownSeries := utils.SubSeries(balanceSeries, df.Col("Highlevel"), "Drawdown")
	ddpercentSeries := utils.MutiSeries(utils.DivSeries(drawdownSeries, df.Col("Highlevel"), ""), 100, "ddpercent")
	startDate = b.Start.Format("2006-01-02")
	endDate = b.End.Format("2006-01-02")
	totalDays = balanceSeries.Len()
	profitDays = utils.GreaterSeries(netPnlSeries, 0)
	lossDays = utils.LowerSeries(netPnlSeries, 0)

	endBalance = balanceSeries.Float()[balanceSeries.Len()-1]
	maxDrawdown = drawdownSeries.Min()
	maxDdpercent = ddpercentSeries.Min()

	totalNetPnl = netPnlSeries.Sum()
	dailyNetPnl = totalNetPnl / float64(totalDays)
	
	totalCommission = df.Col("Commission").Sum()
	dailyCommission = totalCommission / float64(totalDays)

	totalSlippage = df.Col("Slippage").Sum()
	dailySlippage = totalSlippage / float64(totalDays)

	totalTurnover = df.Col("Turnover").Sum()
	dailyTurnover = totalTurnover / float64(totalDays)

	totalTradeCount = df.Col("TradeCount").Sum()
	dailyTradeCount = totalTradeCount / float64(totalDays)

	totalReturn = (endBalance / b.Capital - 1) * 100
	annualReturn = totalReturn / float64(totalDays) * float64(b.AnnualDays)
	
	// TODO: dailyReturn, returnStd, dailyRiskFree, sharpeRatio

	returnDrawdownRatio = -totalReturn / maxDdpercent

	if output {
		b.logger.Println("----------------------")
		b.logger.Printf("首个交易日:\t%v \n", startDate)
		b.logger.Printf("最后交易日:\t%v \n", endDate)

		b.logger.Printf("总交易日：\t%v \n", totalDays)
		b.logger.Printf("盈利交易日：\t%v \n", profitDays)
		b.logger.Printf("亏损交易日：\t%v \n", lossDays)

		b.logger.Printf("起始资金：\t%.2f \n", b.Capital)
		b.logger.Printf("结束资金：\t%.2f \n", endBalance)

		b.logger.Printf("总收益率：\t%.2f \n", totalReturn)
		b.logger.Printf("年化收益：\t%.2f \n", annualReturn)
		b.logger.Printf("最大回撤\t%.2f \n", maxDrawdown)
		b.logger.Printf("百分比最大回撤\t%.2f \n", maxDdpercent)
		// b.logger.Printf("最长回撤天数\t%.2f \n", maxDrawdownDuration)

		b.logger.Printf("总盈亏：\t%.2f \n", totalNetPnl)
		b.logger.Printf("总手续费：\t%.2f \n", totalCommission)
		b.logger.Printf("总滑点：\t%.2f \n", totalSlippage)
		b.logger.Printf("总成交金额：\t%.2f \n", totalTurnover)
		b.logger.Printf("总成交笔数：\t%v \n", totalTradeCount)

		b.logger.Printf("日均盈亏：\t%.2f \n", dailyNetPnl)
		b.logger.Printf("日均手续费：\t%.2f \n", dailyCommission)
		b.logger.Printf("日均滑点：\t%.2f \n", dailySlippage)
		b.logger.Printf("日均成交金额：\t%.2f \n", dailyTurnover)
		b.logger.Printf("日均成交笔数：\t%v \n", dailyTradeCount)

		// b.logger.Printf("日均收益率：\t%.2f \n", dailyReturn)
		// b.logger.Printf("收益标准差：\t%.2f \n", returnStd)
		// b.logger.Printf("Sharpe Ratio：\t%.2f \n", sharpeRatio)
		b.logger.Printf("收益回撤比：\t%.2f \n", returnDrawdownRatio)
	}

	statistics := map[string]any{
		"start_date": startDate,
		"end_date": endDate,
		"total_days": totalDays,
		"profit_days": profitDays,
		"loss_days": lossDays,
		"capital": b.Capital,
		"end_balance": endBalance,
		"max_drawdown": maxDrawdown,
		"max_ddpercent": maxDdpercent,
		"max_drawdown_duration": maxDrawdownDuration,
		"total_net_pnl": totalNetPnl,
		"daily_net_pnl": dailyNetPnl,
		"total_commission": totalCommission,
		"daily_commission": dailyCommission,
		"total_slippage": totalSlippage,
		"daily_slippage": dailySlippage,
		"total_turnover": totalTurnover,
		"daily_turnover": dailyTurnover,
		"total_trade_count": totalTradeCount,
		"daily_trade_count": dailyTradeCount,
		"total_return": totalReturn,
		"annual_return": annualReturn,
		"daily_return": dailyReturn,
		"return_std": returnStd,
		"sharpe_ratio": sharpeRatio,
		"return_drawdown_ratio": returnDrawdownRatio,
	}
	b.logger.Println("策略统计指标计算完成")
	return statistics
}