package cfgimpl

import (
	"context"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

const (
	maxRowPerQuery = 100
)

type DatabaseDataPump struct {
	connString string

	p                *pgxpool.Pool
	eventPumpChannel chan configapi.Event

	startScan    chan struct{}
	startOnce    sync.Once
	updateScanId atomic.Int64
	closeCh      chan struct{}
}

func NewDatabaseDataPump(connString string) *DatabaseDataPump {
	return &DatabaseDataPump{
		connString:       connString,
		closeCh:          make(chan struct{}),
		startScan:        make(chan struct{}),
		eventPumpChannel: make(chan configapi.Event),
	}
}

func (d *DatabaseDataPump) logError(msg string, err error) {
	if err != nil {
		log.Println("[ERROR]", msg, err)
	} else {
		log.Println("[ERROR]", msg)
	}
}

func (d *DatabaseDataPump) Startup() error {
	c, err := pgxpool.ParseConfig(d.connString)
	if err != nil {
		return err
	}
	p, err := pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return err
	}
	d.p = p
	go d.scanLoop()
	return nil
}

func (d *DatabaseDataPump) Stop() error {
	d.p.Close()
	close(d.closeCh)
	close(d.eventPumpChannel)
	return nil
}

func (d *DatabaseDataPump) EventChannel() <-chan configapi.Event {
	return d.eventPumpChannel
}

func (d *DatabaseDataPump) TriggerDumpToChannel() <-chan configapi.Event {
	ch := make(chan configapi.Event, 1024)
	go func() {
		var lastUpdateScanId int64
		// dump db and
		dumpTask := func() {
			// acquire database connections
			var c *pgxpool.Conn
			for {
				var err error
				c, err = d.p.Acquire(context.Background())
				if err != nil {
					d.logError("acquire database connection error", err)
					time.Sleep(1 * time.Second) // retry
					continue
				}
				break
			}
			defer c.Release()
			// do data retrieving
			for {
				time.Sleep(1 * time.Second) // retry
				maxId, err := d.queryMaxSequence(c)
				if err != nil {
					d.logError("query max sequence error", err)
					continue
				}
				lastUpdateScanId = maxId
				var start int64 = 0
				for {
					list, err := d.queryConfigurations(start, lastUpdateScanId, c)
					if err != nil {
						d.logError("query configurations error", err)
						time.Sleep(1 * time.Second) // wait for next trial
						continue
					}
					if len(list) == 0 {
						break // finish all records
					}
					start = list[len(list)-1].Seq
					// send configurations
					for _, v := range list {
						ch <- struct {
							Configuration *configapi.Configuration
							Created       bool
							Modified      bool
							Deleted       bool
						}{Configuration: v.Configuration, Created: v.Status == 0, Modified: false, Deleted: v.Status == 1}
					}
				}
				break
			}
		}
		dumpTask()

		// finally allow event scan start
		close(ch)
		d.updateScanId.Store(lastUpdateScanId)
		d.startOnce.Do(func() {
			close(d.startScan)
		})
	}()
	return ch
}

func (d *DatabaseDataPump) scanLoop() {
	// wait for start
	<-d.startScan
	log.Println("full dump triggered, start scanLoop...")
	// loop scan
Overall:
	for {
		select {
		case _, ok := <-d.closeCh:
			if !ok {
				break Overall
			}
		default:
		}

		// query updates
		lastUpdateScanId := d.updateScanId.Load()
		retrieveFn := func() bool {
			c, err := d.p.Acquire(context.Background())
			if err != nil {
				d.logError("scanLoop acquire database connection error", err)
				time.Sleep(1 * time.Second)
				return true
			}
			defer c.Release()
			data, err := d.queryConfigurations(lastUpdateScanId, math.MaxInt64, c)
			if err != nil {
				d.logError("scanLoop query configurations error", err)
				time.Sleep(1 * time.Second)
				return true
			}
			if len(data) == 0 {
				// no more data, continue with short rest
				return false
			}
			for _, v := range data {
				d.eventPumpChannel <- configapi.Event{
					Configuration: v.Configuration,
					Created:       v.Status == 0,
					Modified:      false,
					Deleted:       v.Status == 1,
				}
			}
			// update seq number to mark as read
			d.updateScanId.Store(data[len(data)-1].Seq)
			return true
		}
		if retrieveFn() {
			// has more updates or error occurs, continue immediately
			//FIXME need to figure out error and wait some time in order to avoid use up all the resources
			continue
		} else {
			// wait short period for next check
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (d *DatabaseDataPump) queryMaxSequence(c *pgxpool.Conn) (int64, error) {
	rows, err := c.Query(context.Background(), "select max(sequence) from configuration")
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, nil
	}
	var maxId *int64
	if err := rows.Scan(&maxId); err != nil {
		return 0, err
	}
	// if no record, nil will be retrieved
	if maxId == nil {
		return 0, nil
	}
	return *maxId, nil
}

func (d *DatabaseDataPump) queryConfigurations(startExcluded int64, maxIdIncluded int64, c *pgxpool.Conn) (res []struct {
	Seq           int64
	Status        int
	Configuration *configapi.Configuration
}, err error) {
	rows, err := c.Query(context.Background(), "select raw_cfg_value, sequence, cfg_status from configuration where sequence > $1 and sequence <= $2 order by sequence asc limit $3",
		startExcluded, maxIdIncluded, maxRowPerQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var data []byte
		var seq int64
		var status int
		if err := rows.Scan(&data, &seq, &status); err != nil {
			return nil, err
		}
		var cfg configapi.Configuration
		if err := cbor.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		res = append(res, struct {
			Seq           int64
			Status        int
			Configuration *configapi.Configuration
		}{Seq: seq, Configuration: &cfg, Status: status})
	}
	return res, nil
}
