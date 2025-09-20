package bes

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultChannelSize = 10000
	defaultBatchSize = 100
	defaultSubmitInterval = 5
)

type ErrorCallback func(err error)

type BesArgs struct {
	Client     *elastic.Client
	Retry        int
	ChannelSize  int
	BatchSize    int
	SubmitInterval int64
	Snowflake *Snowflake
	Callback ErrorCallback
}

type IndexData struct {
	Data  interface{}
	Retry int
}

// BatES es输出器信息
type BatES struct {
	Client       *elastic.Client
	Input chan EsData
	done chan bool
	batchSize    int
	SubmitInterval int64
	retry        int
	DataMap      map[string]IndexData
	Snowflake *Snowflake
	Callback ErrorCallback
}

type EsData struct {
	Id    string
	Index string
	Data  interface{}
}

type MyRetry struct {
	backoff elastic.Backoff
}

func NewRetry(backoff elastic.Backoff) *MyRetry {
	return &MyRetry{backoff: backoff}
}

func (r *MyRetry) Retry(ctx context.Context, retry int, req *http.Request, resp *http.Response, err error) (time.Duration, bool, error) {
	if resp != nil && resp.StatusCode == 429 {
		wait, stop := r.backoff.Next(retry)
		return wait, stop, nil
	}
	return 0, false, nil
}

// NewBatES 获取新es输出器
func NewBatES(p BesArgs) *BatES {
	if p.Client == nil {
		panic("client is nil")
	}
	if p.Snowflake == nil {
		panic("snowflake is nil")
	}
	if p.ChannelSize == 0 {
		p.ChannelSize = defaultChannelSize
	}
	if p.BatchSize == 0 {
		p.BatchSize = defaultBatchSize
	}
	if p.SubmitInterval == 0 {
		p.SubmitInterval = defaultSubmitInterval
	}

	es := &BatES{
		Client:       p.Client,
		Input: make(chan EsData, p.ChannelSize),
		done: make(chan bool),
		batchSize:    p.BatchSize,
		SubmitInterval: p.SubmitInterval,
		retry:        p.Retry,
		DataMap:      make(map[string]IndexData),
		Snowflake: p.Snowflake,
		Callback: p.Callback,
	}
	go es.Run(es.Callback)
	return es
}

// Stop 停止es输出器
func (s *BatES) Stop() {
	s.done <- true
}

// Run 启动es输出器
func (s *BatES) Run(callback ErrorCallback) {
	ticker := time.NewTicker(time.Second * time.Duration(s.SubmitInterval))
	defer ticker.Stop()
	i := 0
	numCached := 0
	bulk := s.Client.Bulk()
	for {
		select {
		case item, ok := <-s.Input:
			i += 1
			if !ok {
				continue
			}
			id, _ := s.Snowflake.NextID()
			if item.Id == "" {
				item.Id = strconv.FormatInt(id, 10)
			}
			s.DataMap[item.Id] = IndexData{
				Data:  item.Data,
				Retry: 0,
			}
			bulk = bulk.Add(elastic.NewBulkIndexRequest().Index(item.Index).Doc(item.Data).Id(item.Id))
			numCached++
			if numCached >= s.batchSize {
				numCached = 0
				resp, err := bulk.Do(context.Background())
				if err != nil {
					callback(fmt.Errorf("es bulk request failed, err:%s", err))
				}
				if resp != nil {
					if resp.Failed() != nil {
						for _, f := range resp.Failed() {
							d, ok := s.DataMap[f.Id]
							if !ok {
								callback(fmt.Errorf("data id:%s index failed with data missing", f.Id))
							} else {
								callback(fmt.Errorf("insert data %+v failed, reason:%+v, cause_by:%+v", d.Data, f.Error.Reason, f.Error.CausedBy))
								if d.Retry >= s.retry {
									callback(fmt.Errorf("data:%+v failed %d times, drop", d.Data, d.Retry))
									delete(s.DataMap, f.Id)
								} else {
									bulk = bulk.Add(elastic.NewBulkIndexRequest().Index(f.Index).Doc(d.Data).Id(f.Id))
									d.Retry += 1
									s.DataMap[f.Id] = d
									numCached += 1
								}
							}
						}
					}
					if resp.Succeeded() != nil {
						for _, r := range resp.Succeeded() {
							delete(s.DataMap, r.Id)
						}
					}
				}
			}

		case <-ticker.C:
			i = 0
			if numCached != 0 {
				numCached = 0
				resp, err := bulk.Do(context.Background())
				if err != nil {
					callback(fmt.Errorf("es bulk request failed, err:%s", err))
				}
				if resp != nil {
					if resp.Failed() != nil {
						for _, f := range resp.Failed() {
							d, ok := s.DataMap[f.Id]
							if !ok {
								callback(fmt.Errorf("data id:%s index failed with data missing", f.Id))
							} else {
								callback(fmt.Errorf("insert data %+v failed, reason:%+v, cause_by:%+v", d.Data, f.Error.Reason, f.Error.CausedBy))
								if d.Retry >= s.retry {
									callback(fmt.Errorf("data:%+v failed %d times, drop", d.Data, d.Retry))
									delete(s.DataMap, f.Id)
								} else {
									bulk = bulk.Add(elastic.NewBulkIndexRequest().Index(f.Index).Doc(d.Data).Id(f.Id))
									d.Retry += 1
									s.DataMap[f.Id] = d
									numCached += 1
								}
							}
						}
					}
					if resp.Succeeded() != nil {
						for _, r := range resp.Succeeded() {
							delete(s.DataMap, r.Id)
						}
					}
				}
			}
		case <-s.done:
			return
		}
	}
}



