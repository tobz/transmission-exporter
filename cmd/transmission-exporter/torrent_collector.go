package main

import (
	"log"
	"strconv"
	"sync"

	transmission "github.com/metalmatze/transmission-exporter"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace string = "transmission_"
)

// TorrentCollector has a transmission.Client to create torrent metrics
type TorrentCollector struct {
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

	cachedTorrents     map[string]transmission.Torrent
	cachedTorrentsLock sync.Mutex
}

// NewTorrentCollector creates a new torrent collector with the transmission.Client
func NewTorrentCollector(client *transmission.Client) *TorrentCollector {
	const collectorNamespace = "torrent_"

	return &TorrentCollector{
		cachedTorrents: make(map[string]transmission.Torrent),
		client:         client,

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
	torrents, err := tc.client.GetTorrents(tc.recentlyActiveOnly)
	if err != nil {
		log.Printf("failed to get torrents: %v", err)
		return
	}
	tc.cachedTorrentsLock.Lock()
	var realTorrentsList []transmission.Torrent
	for _, t := range torrents {
		tc.cachedTorrents[t.HashString] = t
	}
	for _, t := range tc.cachedTorrents {
		realTorrentsList = append(realTorrentsList, t)
	}
	tc.cachedTorrentsLock.Unlock()

	if len(realTorrentsList) > 0 {
		tc.recentlyActiveOnly = true // only do this if successful
	}

	for _, t := range realTorrentsList {
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
