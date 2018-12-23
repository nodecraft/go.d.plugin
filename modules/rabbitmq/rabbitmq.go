package rabbitmq

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/netdata/go.d.plugin/modules"
	"github.com/netdata/go.d.plugin/pkg/stm"
	"github.com/netdata/go.d.plugin/pkg/web"
)

func init() {
	creator := modules.Creator{
		Create: func() modules.Module { return New() },
	}

	modules.Register("rabbitmq", creator)
}

// New creates Rabbitmq with default values
func New() *Rabbitmq {
	var (
		defURL         = "http://localhost:15672"
		defUsername    = "guest"
		defPassword    = "guest"
		defHTTPTimeout = time.Second
	)

	return &Rabbitmq{
		HTTP: web.HTTP{
			Request: web.Request{
				URL:      defURL,
				Username: defUsername,
				Password: defPassword,
			},
			Client: web.Client{Timeout: web.Duration{Duration: defHTTPTimeout}},
		},
		metrics: make(map[string]int64),
	}
}

// https://www.rabbitmq.com/monitoring.html
type (
	apiOverview struct {
		objectTotals `json:"object_totals" stm:"object_totals"`
		queueTotals  `json:"queue_totals" stm:"queue_totals"`
		messageStats `json:"message_stats" stm:"message_stats"`
	}
	apiNodes []node
)

type objectTotals struct {
	Consumers   int `stm:"consumers"`
	Queues      int `stm:"queues"`
	Exchanges   int `stm:"exchanges"`
	Connections int `stm:"connections"`
	Channels    int `stm:"channels"`
}

type queueTotals struct {
	MessagesReady          int `json:"messages_ready" stm:"messages_ready"`
	MessagesUnacknowledged int `json:"messages_unacknowledged" stm:"messages_unacknowledged"`
}

// https://rawcdn.githack.com/rabbitmq/rabbitmq-management/master/priv/www/doc/stats.html
type messageStats struct {
	Ack              int `stm:"ack"`
	Publish          int `stm:"publish"`
	PublishIn        int `json:"publish_in" stm:"publish_in"`
	PublishOut       int `json:"publish_out" stm:"publish_out"`
	Confirm          int `stm:"confirm"`
	Deliver          int `stm:"deliver"`
	DeliverNoAck     int `json:"deliver_no_ack" stm:"deliver_no_ack"`
	Get              int `stm:"get"`
	GetNoAck         int `json:"get_no_ack" stm:"get_no_ack"`
	DeliverGet       int `json:"deliver_get" stm:"deliver_get"`
	Redeliver        int `stm:"redeliver"`
	ReturnUnroutable int `json:"return_unroutable" stm:"return_unroutable"`
}

type node struct {
	FDUsed      int `json:"fd_used" stm:"fd_used"`
	MemUsed     int `json:"mem_used" stm:"mem_used"`
	SocketsUsed int `json:"sockets_used" stm:"sockets_used"`
	ProcUsed    int `json:"proc_used" stm:"proc_used"`
	DiskFree    int `json:"disk_free" stm:"disk_free"`
	RunQueue    int `json:"run_queue" stm:"run_queue"`
}

// Rabbitmq rabbitmq module
type Rabbitmq struct {
	modules.Base // should be embedded by every module

	web.HTTP `yaml:",inline"`

	reqOverview *http.Request
	reqNodes    *http.Request
	client      *http.Client

	overview apiOverview
	nodes    apiNodes

	metrics map[string]int64
}

// Cleanup makes cleanup
func (Rabbitmq) Cleanup() {}

// Init makes initialization
func (r *Rabbitmq) Init() bool {
	// create HTTP requests
	if err := r.createRequests(); err != nil {
		r.Errorf("error on creating request to %s : %s", r.URL, err)
		return false
	}

	// create HTTP client
	r.client = web.NewHTTPClient(r.Client)

	// post Init debug info
	r.Debugf("using URL %s", r.URL)
	r.Debugf("using timeout: %s", r.Timeout.Duration)

	return true
}

// Check makes check
func (r *Rabbitmq) Check() bool {
	return len(r.Collect()) > 0
}

// Charts creates Charts
func (Rabbitmq) Charts() *Charts {
	return charts.Copy()
}

// Collect collects stats
func (r *Rabbitmq) Collect() map[string]int64 {
	if err := r.collect(r.reqOverview, &r.overview); err != nil {
		r.Error(err)
		return nil
	}

	if err := r.collect(r.reqNodes, &r.nodes); err != nil {
		r.Error(err)
		return nil
	}

	r.metrics = make(map[string]int64)

	for k, v := range stm.ToMap(r.overview) {
		r.metrics[k] = v
	}

	if len(r.nodes) > 0 {
		for k, v := range stm.ToMap(r.nodes[0]) {
			r.metrics[k] = v
		}
	}

	return r.metrics
}

func (r *Rabbitmq) doRequest(req *http.Request) (*http.Response, error) {
	return r.client.Do(req)
}

func (r *Rabbitmq) collect(req *http.Request, stats interface{}) error {
	resp, err := r.doRequest(req)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s", req.URL, err)
	}

	defer func() {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %d", req.URL, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(stats); err != nil {
		return fmt.Errorf("erorr on decode request to %s : %s", req.URL, err)
	}

	return nil
}

func (r *Rabbitmq) createRequests() (err error) {
	r.URI = "/api/overview"

	if r.reqOverview, err = web.NewHTTPRequest(r.Request); err != nil {
		return err
	}

	r.URI = "/api/nodes"

	if r.reqNodes, err = web.NewHTTPRequest(r.Request); err != nil {
		return err
	}

	return nil
}