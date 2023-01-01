package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tobz/transmission-exporter"
	"go.uber.org/zap"
)

// SessionCollector exposes session metrics
type SessionCollector struct {
	logger *zap.Logger
	client *transmission.Client

	AltSpeedDown     *prometheus.Desc
	AltSpeedUp       *prometheus.Desc
	CacheSize        *prometheus.Desc
	FreeSpace        *prometheus.Desc
	QueueDown        *prometheus.Desc
	QueueUp          *prometheus.Desc
	PeerLimitGlobal  *prometheus.Desc
	PeerLimitTorrent *prometheus.Desc
	SeedRatioLimit   *prometheus.Desc
	SpeedLimitDown   *prometheus.Desc
	SpeedLimitUp     *prometheus.Desc
	Version          *prometheus.Desc
}

// NewSessionCollector takes a transmission.Client and returns a SessionCollector
func NewSessionCollector(logger *zap.Logger, client *transmission.Client) *SessionCollector {
	return &SessionCollector{
		logger: logger,
		client: client,

		AltSpeedDown: prometheus.NewDesc(
			namespace+"alt_speed_down",
			"Alternative max global download speed",
			[]string{"enabled"},
			nil,
		),
		AltSpeedUp: prometheus.NewDesc(
			namespace+"alt_speed_up",
			"Alternative max global upload speed",
			[]string{"enabled"},
			nil,
		),
		CacheSize: prometheus.NewDesc(
			namespace+"cache_size_bytes",
			"Maximum size of the disk cache",
			nil,
			nil,
		),
		FreeSpace: prometheus.NewDesc(
			namespace+"free_space",
			"Free space left on device to download to",
			[]string{"download_dir", "incomplete_dir"},
			nil,
		),
		QueueDown: prometheus.NewDesc(
			namespace+"queue_down",
			"Max number of torrents to download at once",
			[]string{"enabled"},
			nil,
		),
		QueueUp: prometheus.NewDesc(
			namespace+"queue_up",
			"Max number of torrents to upload at once",
			[]string{"enabled"},
			nil,
		),
		PeerLimitGlobal: prometheus.NewDesc(
			namespace+"global_peer_limit",
			"Maximum global number of peers",
			nil,
			nil,
		),
		PeerLimitTorrent: prometheus.NewDesc(
			namespace+"torrent_peer_limit",
			"Maximum number of peers for a single torrent",
			nil,
			nil,
		),
		SeedRatioLimit: prometheus.NewDesc(
			namespace+"seed_ratio_limit",
			"The default seed ratio for torrents to use",
			[]string{"enabled"},
			nil,
		),
		SpeedLimitDown: prometheus.NewDesc(
			namespace+"speed_limit_down_bytes",
			"Max global download speed",
			[]string{"enabled"},
			nil,
		),
		SpeedLimitUp: prometheus.NewDesc(
			namespace+"speed_limit_up_bytes",
			"Max global upload speed",
			[]string{"enabled"},
			nil,
		),
		Version: prometheus.NewDesc(
			namespace+"version",
			"Transmission version as label",
			[]string{"version"},
			nil,
		),
	}
}

// Describe implements the prometheus.Collector interface
func (sc *SessionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.AltSpeedDown
	ch <- sc.AltSpeedUp
	ch <- sc.CacheSize
	ch <- sc.FreeSpace
	ch <- sc.QueueDown
	ch <- sc.QueueUp
	ch <- sc.PeerLimitGlobal
	ch <- sc.PeerLimitTorrent
	ch <- sc.SeedRatioLimit
	ch <- sc.SpeedLimitDown
	ch <- sc.SpeedLimitUp
	ch <- sc.Version
}

// Collect implements the prometheus.Collector interface
func (sc *SessionCollector) Collect(ch chan<- prometheus.Metric) {
	session, err := sc.client.GetSession()
	if err != nil {
		sc.logger.Error("Failed to get session from Transmission.", zap.Error(err))
		return
	}

	ch <- prometheus.MustNewConstMetric(
		sc.AltSpeedDown,
		prometheus.GaugeValue,
		float64(session.AltSpeedDown),
		NumericBool(session.AltSpeedEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.AltSpeedUp,
		prometheus.GaugeValue,
		float64(session.AltSpeedUp),
		NumericBool(session.AltSpeedEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.CacheSize,
		prometheus.GaugeValue,
		float64(session.CacheSizeMB*1024*1024),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.FreeSpace,
		prometheus.GaugeValue,
		float64(session.DownloadDirFreeSpace),
		session.DownloadDir, session.IncompleteDir,
	)
	ch <- prometheus.MustNewConstMetric(
		sc.QueueDown,
		prometheus.GaugeValue,
		float64(session.DownloadQueueSize),
		NumericBool(session.DownloadQueueEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.QueueUp,
		prometheus.GaugeValue,
		float64(session.SeedQueueSize),
		NumericBool(session.SeedQueueEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.PeerLimitGlobal,
		prometheus.GaugeValue,
		float64(session.PeerLimitGlobal),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.PeerLimitTorrent,
		prometheus.GaugeValue,
		float64(session.PeerLimitPerTorrent),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.SeedRatioLimit,
		prometheus.GaugeValue,
		float64(session.SeedRatioLimit),
		NumericBool(session.SeedRatioLimited),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.SpeedLimitDown,
		prometheus.GaugeValue,
		float64(session.SpeedLimitDown),
		NumericBool(session.SpeedLimitDownEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.SpeedLimitUp,
		prometheus.GaugeValue,
		float64(session.SpeedLimitUp),
		NumericBool(session.SpeedLimitUpEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.Version,
		prometheus.GaugeValue,
		float64(1),
		session.Version,
	)
}
