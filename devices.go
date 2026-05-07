package collector

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wguner/tailscale-exporter/tailscale"
)

// DeviceCollector implements prometheus.Collector for Tailscale device metrics.
type DeviceCollector struct {
	client *tailscale.Client

	deviceOnline      *prometheus.Desc
	deviceLastSeen    *prometheus.Desc
	deviceKeyExpiry   *prometheus.Desc
	devicesTotal      *prometheus.Desc
	deviceUpdateAvail *prometheus.Desc
}

// NewDeviceCollector initialises all metric descriptors.
func NewDeviceCollector(client *tailscale.Client) *DeviceCollector {
	labels := []string{"id", "hostname", "os"}

	return &DeviceCollector{
		client: client,

		deviceOnline: prometheus.NewDesc(
			"tailscale_device_online",
			"1 if the device was seen within the last 5 minutes, 0 otherwise.",
			labels, nil,
		),
		deviceLastSeen: prometheus.NewDesc(
			"tailscale_device_last_seen_timestamp_seconds",
			"Unix timestamp of the most recent time the device was seen.",
			labels, nil,
		),
		deviceKeyExpiry: prometheus.NewDesc(
			"tailscale_device_key_expiry_seconds",
			"Seconds until the device auth key expires. Negative if already expired.",
			labels, nil,
		),
		devicesTotal: prometheus.NewDesc(
			"tailscale_devices_total",
			"Total number of devices registered in the tailnet.",
			nil, nil,
		),
		deviceUpdateAvail: prometheus.NewDesc(
			"tailscale_device_update_available",
			"1 if a Tailscale client update is available on the device.",
			labels, nil,
		),
	}
}

// Describe sends all metric descriptors to the channel — required by prometheus.Collector.
func (c *DeviceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.deviceOnline
	ch <- c.deviceLastSeen
	ch <- c.deviceKeyExpiry
	ch <- c.devicesTotal
	ch <- c.deviceUpdateAvail
}

// Collect fetches current device data and emits metrics — called on every scrape.
func (c *DeviceCollector) Collect(ch chan<- prometheus.Metric) {
	devices, err := c.client.GetDevices()
	if err != nil {
		log.Printf("[error] fetching devices from Tailscale API: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.devicesTotal,
		prometheus.GaugeValue,
		float64(len(devices)),
	)

	now := time.Now()

	for _, d := range devices {
		labels := []string{d.ID, d.Hostname, d.OS}

		// Online: seen within the last 5 minutes
		online := 0.0
		if now.Sub(d.LastSeen) < 5*time.Minute {
			online = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.deviceOnline, prometheus.GaugeValue, online, labels...)

		// Last seen as a Unix timestamp
		ch <- prometheus.MustNewConstMetric(
			c.deviceLastSeen,
			prometheus.GaugeValue,
			float64(d.LastSeen.Unix()),
			labels...,
		)

		// Key expiry — skip devices with expiry disabled (e.g. tagged devices)
		if !d.KeyExpiryDisabled {
			expirySeconds := d.Expires.Sub(now).Seconds()
			ch <- prometheus.MustNewConstMetric(
				c.deviceKeyExpiry,
				prometheus.GaugeValue,
				expirySeconds,
				labels...,
			)
		}

		// Whether a client update is available
		updateAvail := 0.0
		if d.UpdateAvailable {
			updateAvail = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.deviceUpdateAvail,
			prometheus.GaugeValue,
			updateAvail,
			labels...,
		)
	}
}
