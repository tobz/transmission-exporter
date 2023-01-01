package main

import (
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	transmission "github.com/tobz/transmission-exporter"
	"go.uber.org/zap"
)

const (
	namespace string = "transmission_"
)

// TorrentCollector has a transmission.Client to create torrent metrics
type TorrentCollector struct {
	logger *zap.Logger
	client *transmission.Client

	Status             *prometheus.Desc
	Added              *prometheus.Desc
	Finished           *prometheus.Desc
	Done               *prometheus.Desc
	Ratio              *prometheus.Desc
	Download           *prometheus.Desc
	Upload             *prometheus.Desc
	UploadedEver       *prometheus.Desc
	DownloadedEver     *prometheus.Desc
	PeersConnected     *prometheus.Desc
	PeersGettingFromUs *prometheus.Desc
	PeersSendingToUs   *prometheus.Desc

	recentlyActiveOnly bool

	torrentMap     map[int]transmission.Torrent
	torrentMapLock sync.Mutex
}

// NewTorrentCollector creates a new torrent collector with the transmission.Client
func NewTorrentCollector(logger *zap.Logger, client *transmission.Client) *TorrentCollector {
	const collectorNamespace = "torrent_"

	return &TorrentCollector{
		torrentMap: make(map[int]transmission.Torrent),
		logger:     logger,
		client:     client,

		Status: prometheus.NewDesc(
			namespace+collectorNamespace+"status",
			"Status of a torrent",
			[]string{"id", "name"},
			nil,
		),
		Added: prometheus.NewDesc(
			namespace+collectorNamespace+"added",
			"The unixtime time a torrent was added",
			[]string{"id", "name"},
			nil,
		),
		Finished: prometheus.NewDesc(
			namespace+collectorNamespace+"finished",
			"Indicates if a torrent is finished (1) or not (0)",
			[]string{"id", "name"},
			nil,
		),
		Done: prometheus.NewDesc(
			namespace+collectorNamespace+"done",
			"The percent of a torrent being done",
			[]string{"id", "name"},
			nil,
		),
		Ratio: prometheus.NewDesc(
			namespace+collectorNamespace+"ratio",
			"The upload ratio of a torrent",
			[]string{"id", "name"},
			nil,
		),
		Download: prometheus.NewDesc(
			namespace+collectorNamespace+"download_bytes",
			"The current download rate of a torrent in bytes",
			[]string{"id", "name"},
			nil,
		),
		Upload: prometheus.NewDesc(
			namespace+collectorNamespace+"upload_bytes",
			"The current upload rate of a torrent in bytes",
			[]string{"id", "name"},
			nil,
		),
		UploadedEver: prometheus.NewDesc(
			namespace+collectorNamespace+"uploaded_ever_bytes",
			"The amount of bytes that have been uploaded from a torrent ever",
			[]string{"id", "name"},
			nil,
		),
		DownloadedEver: prometheus.NewDesc(
			namespace+collectorNamespace+"downloaded_ever_bytes",
			"The amount of bytes that have been downloaded from a torrent ever",
			[]string{"id", "name"},
			nil,
		),
		PeersConnected: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_connected",
			"The quantity of peers connected on a torrent",
			[]string{"id", "name"},
			nil,
		),
		PeersGettingFromUs: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_getting_from_us",
			"The quantity of peers getting pieces of a torrent from us",
			[]string{"id", "name"},
			nil,
		),
		PeersSendingToUs: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_sending_to_us",
			"The quantity of peers sending pieces of a torrent to us",
			[]string{"id", "name"},
			nil,
		),
	}
}

// Describe implements the prometheus.Collector interface
func (tc *TorrentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- tc.Status
	ch <- tc.Added
	ch <- tc.Finished
	ch <- tc.Done
	ch <- tc.Ratio
	ch <- tc.Download
	ch <- tc.Upload
	ch <- tc.UploadedEver
	ch <- tc.DownloadedEver
	ch <- tc.PeersConnected
	ch <- tc.PeersGettingFromUs
	ch <- tc.PeersSendingToUs
}

// Collect implements the prometheus.Collector interface
func (tc *TorrentCollector) Collect(ch chan<- prometheus.Metric) {
	response, err := tc.client.GetTorrents(tc.recentlyActiveOnly)
	if err != nil {
		tc.logger.Error("Failed to get torrents from Transmission.", zap.Error(err))
		return
	}

	var activeTorrents []transmission.Torrent

	// Update our map of cached torrents, both adding any new torrents as well as deleting any
	// removed torrents. We'll create a new list after doing that to iterate over for metrics.
	tc.torrentMapLock.Lock()
	for _, t := range response.Torrents {
		tc.torrentMap[t.ID] = t
	}
	for _, id := range response.RemovedTorrents {
		delete(tc.torrentMap, id)
	}
	for _, t := range tc.torrentMap {
		activeTorrents = append(activeTorrents, t)
	}
	tc.torrentMapLock.Unlock()

	if len(activeTorrents) > 0 {
		tc.recentlyActiveOnly = true // only do this if successful
	}

	for _, t := range activeTorrents {
		var finished float64

		id := strconv.Itoa(t.ID)

		if t.IsFinished {
			finished = 1
		}

		ch <- prometheus.MustNewConstMetric(
			tc.Status,
			prometheus.GaugeValue,
			float64(t.Status),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Added,
			prometheus.GaugeValue,
			float64(t.Added),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Finished,
			prometheus.GaugeValue,
			finished,
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Done,
			prometheus.GaugeValue,
			t.PercentDone,
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Ratio,
			prometheus.GaugeValue,
			t.UploadRatio,
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Download,
			prometheus.GaugeValue,
			float64(t.RateDownload),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Upload,
			prometheus.GaugeValue,
			float64(t.RateUpload),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.UploadedEver,
			prometheus.GaugeValue,
			float64(t.UploadedEver),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.DownloadedEver,
			prometheus.GaugeValue,
			float64(t.DownloadedEver),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersConnected,
			prometheus.GaugeValue,
			float64(t.PeersConnected),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersGettingFromUs,
			prometheus.GaugeValue,
			float64(t.PeersGettingFromUs),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersSendingToUs,
			prometheus.GaugeValue,
			float64(t.PeersSendingToUs),
			id, t.Name,
		)
	}
}
