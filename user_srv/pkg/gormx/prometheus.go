package gormx

import (
	"time"

	prome "github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

// 监控mysql的执行时间
type Callbacks struct {
	Vector *prome.SummaryVec
}

func (c Callbacks) Name() string {
	return "call_back_prome"
}

func (c Callbacks) Initialize(d *gorm.DB) error {
	return c.registerAll(d)
}

func NewCallbacks() *Callbacks {
	label := []string{"type", "table"}
	vec := prome.NewSummaryVec(prome.SummaryOpts{
		Namespace: "github.com/Numsina/tkshop",
		Subsystem: "products",
		Name:      "sql_exec_time",
		Help:      "统计sql语句的执行时间",
		Objectives: map[float64]float64{
			0.5:   0.1,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, label)

	prome.Register(vec)
	return &Callbacks{
		Vector: vec,
	}
}

func (c *Callbacks) registerAll(db *gorm.DB) error {
	err := db.Callback().Create().Before("*").Register("create_before", c.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Create().After("*").Register("create_After", c.After("create"))
	if err != nil {
		return err
	}
	err = db.Callback().Update().Before("*").Register("update_before", c.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Update().After("*").Register("update_After", c.After("update"))
	if err != nil {
		return err
	}
	err = db.Callback().Delete().Before("*").Register("delete_before", c.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Delete().After("*").Register("delete_After", c.After("delete"))
	if err != nil {
		return err
	}
	err = db.Callback().Query().Before("*").Register("query_before", c.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Query().After("*").Register("query_After", c.After("query"))
	if err != nil {
		return err
	}
	err = db.Callback().Raw().Before("*").Register("Raw_before", c.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Raw().After("*").Register("Raw_After", c.After("raw"))
	if err != nil {
		return err
	}
	err = db.Callback().Row().Before("*").Register("Row_before", c.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Row().After("*").Register("Row_After", c.After("Row"))
	if err != nil {
		return err
	}
	return nil
}

func (c Callbacks) Before() func(*gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c Callbacks) After(s string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		duration := time.Since(startTime)
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.Vector.WithLabelValues(s, table).Observe(float64(duration.Milliseconds()))
	}
}
