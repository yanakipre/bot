package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/yanakipre/bot/internal/promtooling"
)

// Consumer
var (
	MBusMessageConsumptionDelayDuration = promtooling.NewHistogramVec(
		"cloud_mbus_ingress_message_delay",
		"Histogram of ingress messages production delay from now in seconds",
		[]float64{
			0.01, // 1ms
			0.05,
			0.1,
			0.2,
			0.4,
			0.6,
			0.8,
			1,
			3,
			6,
			10,
			20,
			40,
			60,
			90,
			120, // 2 min
			240,
			600,
			3600, // 1 hour
		},
		[]string{"bus", "topic", "cons_group", "cons_id"},
	)
	MBusMessagesConsumedTotal = promtooling.NewCounterVec(
		"cloud_mbus_messages_consumed_total",
		"Count messages consumed total",
		[]string{"bus", "topic", "cons_group", "cons_id"},
	)
	MBusMessagesConsumedProcessed = promtooling.NewCounterVec(
		"cloud_mbus_messages_consumed_processed",
		"Count messages consumed and successfully processed",
		[]string{"bus", "topic", "cons_group", "cons_id"},
	)
	MBusPendingAckMessages = promtooling.NewGaugeVec(
		"cloud_mbus_pending_ack_messages",
		"Count messages pending ack",
		[]string{"bus", "topic", "cons_group", "cons_id"},
	)
	MBusFetchedEntries = promtooling.NewGaugeVec(
		"cloud_mbus_fetched_entries",
		"Count entries fetched from stream",
		[]string{"bus", "topic", "cons_group", "cons_id"},
	)
	MBusUnprocessedAckErrors = promtooling.NewGaugeVec(
		"cloud_mbus_unprocessed_ack_errors",
		"Count unprocessed ack errors",
		[]string{"bus", "topic", "cons_group", "cons_id"},
	)
)

// Producer
var (
	MBusMessagePersistDuration = promtooling.NewHistogramVec(
		"cloud_mbus_egress_message_persist_duration",
		"Histogram of time to persist a messages to message bus in seconds",
		[]float64{
			0.01, // 1ms
			0.05,
			0.1,
			0.2,
			0.4,
			0.6,
			0.8,
			1,
			3,
			6,
			10,
			20,
		},
		[]string{"bus", "topic"},
	)
	MBusMessagesProducedTotal = promtooling.NewCounterVec(
		"cloud_mbus_messages_produced_total",
		"Count messages produced",
		[]string{"bus", "topic"},
	)
	MBusMessagesProducedSuccess = promtooling.NewCounterVec(
		"cloud_mbus_messages_produced_success",
		"Count messages produced and written successfully",
		[]string{"bus", "topic"},
	)
	MBusMessagesLastDelivered = promtooling.NewGaugeVec(
		"cloud_mbus_messages_last_delivered",
		"Timestamp part of ID of last delivered message",
		[]string{"bus", "topic", "cons_group"},
	)
	MBusMessagesFirstStored = promtooling.NewGaugeVec(
		"cloud_mbus_messages_first_stored",
		"Timestamp part of ID of first message stored for topic",
		[]string{"bus", "topic"},
	)
	MBusMessagesLastStored = promtooling.NewGaugeVec(
		"cloud_mbus_messages_last_stored",
		"Timestamp part of ID of last message stored for topic",
		[]string{"bus", "topic"},
	)
	MBusMessagesLen = promtooling.NewGaugeVec(
		"cloud_mbus_messages_len",
		"Current stream len",
		[]string{"bus", "topic"},
	)
	MBusMessagesLag = promtooling.NewGaugeVec(
		"cloud_mbus_messages_lag",
		"Reported lag",
		[]string{"bus", "topic", "cons_group"},
	)
	MBusMessagesPending = promtooling.NewGaugeVec(
		"cloud_mbus_messages_pending",
		"Number of messages pending ack",
		[]string{"bus", "topic", "cons_group"},
	)
)

func mbusMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		MBusMessageConsumptionDelayDuration,
		MBusMessagePersistDuration,
		MBusMessagesProducedTotal,
		MBusMessagesProducedSuccess,
		MBusMessagesConsumedProcessed,
		MBusMessagesConsumedTotal,
		MBusMessagesLastDelivered,
		MBusMessagesFirstStored,
		MBusMessagesLag,
		MBusMessagesPending,
		MBusMessagesLastStored,
		MBusMessagesLen,
		MBusPendingAckMessages,
		MBusFetchedEntries,
		MBusUnprocessedAckErrors,
	}
}
