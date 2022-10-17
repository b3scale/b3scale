package metrics

import (
	"context"
	"net/url"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/store"
)

// Metric descriptors used by the metrics collector
var (
	meetingAttendeesDesc = prometheus.NewDesc(
		"b3scale_meeting_attendees",
		"Number of attendees in the cluster",
		[]string{
			// Frontend Key
			"frontend",
			// Backend host
			"backend",
			// Type is either "audio" or "video"
			"type",
			// MeetingID
			"meeting",
		}, nil)

	meetingDurationsDesc = prometheus.NewDesc(
		"b3scale_meeting_durations",
		"Duration of meetings in the cluster",
		[]string{
			// Frontend Key
			"frontend",
			// Backend Host
			"backend",
			// MeetingID
			"meeting",
		}, nil)

	frontendMeetingsDesc = prometheus.NewDesc(
		"b3scale_frontend_meetings",
		"Number of meetings per frontend",
		[]string{
			// Frontend Key
			"frontend",
		}, nil)

	backendMeetingsDesc = prometheus.NewDesc(
		"b3scale_backend_meetings",
		"Number of meetings per backend",
		[]string{
			// Backend Host
			"backend",
		}, nil)
)

// The Collector will gather metrics from the b3scale
// cluster about current meetings
//   - attendees [ frontend, backend, type, meeting ]
//   - durations [ meeting ]
//
type Collector struct{}

// Describe the collector
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- meetingAttendeesDesc
	ch <- meetingDurationsDesc
	ch <- frontendMeetingsDesc
}

// Collect metrics from store
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	// Create context, acquire database connection and start tx
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := store.Acquire(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not collect metrics")
		return
	}
	defer conn.Release()
	ctx = store.ContextWithConnection(ctx, conn)
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not collect metrics")
		return
	}
	defer tx.Rollback(ctx)

	// Collect attendees and meetings
	if err := c.collectMeetingMetrics(ctx, tx, ch); err != nil {
		log.Error().Err(err).Msg("could not collect metrics for meetings")
	}

}

// Collect attendee metrics
func (c Collector) collectMeetingMetrics(
	ctx context.Context,
	tx pgx.Tx,
	ch chan<- prometheus.Metric,
) error {
	meetings, err := store.GetMeetingStates(ctx, tx, store.Q())
	if err != nil {
		return err
	}

	// Get front and backends, so we can lookup the keys and hosts
	feKeys, err := getFrontendKeys(ctx, tx)
	if err != nil {
		return err
	}
	beHosts, err := getBackendHosts(ctx, tx)
	if err != nil {
		return err
	}

	// Collect meeting count per frontend and backend
	feMeetings := map[string]float64{}
	beMeetings := map[string]float64{}

	// For each meeting count audio and video attendees
	for _, m := range meetings {
		var ac, vc float64
		for _, a := range m.Meeting.Attendees {
			if a.HasVideo {
				vc++
			} else {
				ac++
			}
		}

		fkey := "undefined"
		if m.FrontendID != nil {
			fkey = feKeys[*m.FrontendID]
		}
		host := "undefined"
		if m.BackendID != nil {
			host = beHosts[*m.BackendID]
		}

		// Count meetings per frontend and backend
		if _, ok := feMeetings[fkey]; !ok {
			feMeetings[fkey] = 0
		}
		feMeetings[fkey]++
		if _, ok := beMeetings[host]; !ok {
			beMeetings[host] = 0
		}
		beMeetings[host]++

		duration := time.Now().UTC().Sub(m.CreatedAt)

		// Audio attendees count
		ch <- prometheus.MustNewConstMetric(
			meetingAttendeesDesc, prometheus.GaugeValue,
			ac, fkey, host, "audio", m.ID,
		)
		// Video attendees count
		ch <- prometheus.MustNewConstMetric(
			meetingAttendeesDesc, prometheus.GaugeValue,
			vc, fkey, host, "video", m.ID,
		)

		// If meeting is running, we log the duration
		if m.Meeting.Running {
			ch <- prometheus.MustNewConstMetric(
				meetingDurationsDesc, prometheus.CounterValue,
				duration.Seconds(), fkey, host, m.ID,
			)
		}
	}

	for fkey, count := range feMeetings {
		ch <- prometheus.MustNewConstMetric(
			frontendMeetingsDesc, prometheus.GaugeValue,
			count, fkey,
		)
	}

	for host, count := range beMeetings {
		ch <- prometheus.MustNewConstMetric(
			backendMeetingsDesc, prometheus.GaugeValue,
			count, host,
		)
	}

	return nil
}

// Get all frontend keys and map to IDs
func getFrontendKeys(ctx context.Context, tx pgx.Tx) (map[string]string, error) {
	states, err := store.GetFrontendStates(ctx, tx, store.Q())
	if err != nil {
		return nil, err
	}
	keys := make(map[string]string)
	for _, s := range states {
		keys[s.ID] = s.Frontend.Key
	}
	return keys, nil
}

// Get all backend hosts (and strip down the host to it's domain)
func getBackendHosts(ctx context.Context, tx pgx.Tx) (map[string]string, error) {
	states, err := store.GetBackendStates(ctx, tx, store.Q())
	if err != nil {
		return nil, err
	}
	hosts := make(map[string]string)
	for _, s := range states {
		hosts[s.ID] = hostname(s.Backend.Host)
	}
	return hosts, nil
}

// Get the hostname only from the parsed url
func hostname(h string) string {
	u, err := url.Parse(h)
	if err != nil {
		return "undefined"
	}
	return u.Host
}
