package statsd

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

// Client type defines the relevant properties of a StatsD connection.
type Client struct {
	Host string
	Port int
	conn net.Conn
}

// New is a factory method to initialize udp connection
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
func New(host string, port int) *Client {
	client := Client{Host: host, Port: port}
	client.Open()
	return &client
}

// Open opens udp connection, called by default client factory
func (client *Client) Open() {
	connectionString := fmt.Sprintf("%s:%d", client.Host, client.Port)
	conn, err := net.Dial("udp", connectionString)
	if err != nil {
		logger.Error.Println(err)
	}
	client.conn = conn
}

// Close closes udp connection
func (client *Client) Close() {
	client.conn.Close()
}

// Timing logs timing information (in milliseconds) without sampling
//
// Usage:
//
//     import (
//         "statsd"
//         "time"
//     )
//
//     client := statsd.New('localhost', 8125)
//     t1 := time.Now()
//     expensiveCall()
//     t2 := time.Now()
//     duration := int64(t2.Sub(t1)/time.Millisecond)
//     client.Timing("foo.time", duration)
func (client *Client) Timing(stat string, time int64) {
	updateString := fmt.Sprintf("%d|ms", time)
	stats := map[string]string{stat: updateString}
	client.Send(stats, 1)
}

// TimingWithSampleRate logs timing information (in milliseconds) with sampling
//
// Usage:
//
//     import (
//         "statsd"
//         "time"
//     )
//
//     client := statsd.New('localhost', 8125)
//     t1 := time.Now()
//     expensiveCall()
//     t2 := time.Now()
//     duration := int64(t2.Sub(t1)/time.Millisecond)
//     client.TimingWithSampleRate("foo.time", duration, 0.2)
func (client *Client) TimingWithSampleRate(stat string, time int64, sampleRate float32) {
	updateString := fmt.Sprintf("%d|ms", time)
	stats := map[string]string{stat: updateString}
	client.Send(stats, sampleRate)
}

// IncrementCounter increments one stat counter without sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.IncrementCounter('foo.bar')
func (client *Client) IncrementCounter(stat string) {
	stats := []string{stat}
	client.UpdateCounters(stats, 1, 1)
}

// IncrementCounterByValue increments one stat counter by value provided without sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.IncrementCounterByValue('foo.bar', 5)
func (client *Client) IncrementCounterByValue(stat string, val int) {
	stats := []string{stat}
	client.UpdateCounters(stats, val, 1)
}

// IncrementCounterWithSampling increments one stat counter with sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.IncrementCounterWithSampling('foo.bar', 0.2)
func (client *Client) IncrementCounterWithSampling(stat string, sampleRate float32) {
	stats := []string{stat}
	client.UpdateCounters(stats[:], 1, sampleRate)
}

// DecrementCounter decrements one stat counter without sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.DecrementCounter('foo.bar')
func (client *Client) DecrementCounter(stat string) {
	stats := []string{stat}
	client.UpdateCounters(stats[:], -1, 1)
}

// DecrementCounterWithSampling decrements one stat counter with sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.DecrementCounterWithSampling('foo.bar', 0.2)
func (client *Client) DecrementCounterWithSampling(stat string, sampleRate float32) {
	stats := []string{stat}
	client.UpdateCounters(stats[:], -1, sampleRate)
}

// UpdateCounters arbitrarily updates a list of stat counters by a delta
func (client *Client) UpdateCounters(stats []string, delta int, sampleRate float32) {
	statsToSend := make(map[string]string)
	for _, stat := range stats {
		updateString := fmt.Sprintf("%d|c", delta)
		statsToSend[stat] = updateString
	}
	client.Send(statsToSend, sampleRate)
}

// IncrementGauge increments one stat gauge
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.IncrementGauge('foo.bar')
func (client *Client) IncrementGauge(stat string) {
	stats := []string{stat}
	client.UpdateGauges(stats, 1, 1)
}

// IncrementGaugeByValue increments one stat gauge by value provided
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.IncrementGaugeByValue('foo.bar', 5)
func (client *Client) IncrementGaugeByValue(stat string, val int) {
	stats := []string{stat}
	client.UpdateGauges(stats, val, 1)
}

// DecrementGauge decrements one stat gauge without sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.DecrementGauge('foo.bar')
func (client *Client) DecrementGauge(stat string) {
	stats := []string{stat}
	client.UpdateGauges(stats[:], -1, 1)
}

// DecrementGaugeByValue fecrements one stat gauge by value provided
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.DecrementGaugeByValue('foo.bar')
func (client *Client) DecrementGaugeByValue(stat string, val int) {
	stats := []string{stat}
	client.UpdateGauges(stats[:], -val, 1)
}

// UpdateGauges arbitrarily updates a list of stat gagues by a delta
func (client *Client) UpdateGauges(stats []string, delta int, sampleRate float32) {
	statsToSend := make(map[string]string)
	for _, stat := range stats {
		updateString := fmt.Sprintf("%d|g", delta)
		statsToSend[stat] = updateString
	}
	client.Send(statsToSend, sampleRate)
}

// Send will send data to statsd daemon by UDP
func (client *Client) Send(data map[string]string, sampleRate float32) {
	sampledData := make(map[string]string)
	if sampleRate < 1 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		rNum := r.Float32()
		if rNum <= sampleRate {
			for stat, value := range data {
				sampledUpdateString := fmt.Sprintf("%s|@%f", value, sampleRate)
				sampledData[stat] = sampledUpdateString
			}
		}
	} else {
		sampledData = data
	}

	for k, v := range sampledData {
		update_string := fmt.Sprintf("%s:%s", k, v)
		_, err := fmt.Fprintf(client.conn, update_string)
		if err != nil {
			logger.Error.Println(err)
		}
	}
}
