package metrics

import (
	"context"
	"github.com/k8s-practice/octopus/pkg/util/ignore"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// GrpcServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a gRPC server.
type GrpcServerMetrics struct {
	startedRequest *prometheus.CounterVec
	handledRequest *prometheus.CounterVec

	receivedStreamMsg *prometheus.CounterVec
	sentStreamMsg     *prometheus.CounterVec

	countsHandlingTimeEnabled bool
	countsHandlingTimeOpts    prometheus.HistogramOpts
	countsHandlingTime        *prometheus.HistogramVec
}

// NewGrpcServerMetrics returns a ServerMetrics object. Use a new instance of
// ServerMetrics when not using the default Prometheus metrics registry, for
// example when wanting to control which metrics are added to a registry as
// opposed to automatically adding metrics via init functions.
func NewGrpcServerMetrics(namespace string, subsystem string) *GrpcServerMetrics {
	return &GrpcServerMetrics{
		startedRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_server_started_total",
				Help:      "Total number of RPCs started on the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		handledRequest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_server_handled_total",
				Help:      "Total number of RPCs completed on the server, regardless of success or failure.",
			}, []string{"grpc_type", "grpc_service", "grpc_method", "grpc_code"}),
		receivedStreamMsg: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_server_msg_received_total",
				Help:      "Total number of RPC stream messages received on the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		sentStreamMsg: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "grpc_server_msg_sent_total",
				Help:      "Total number of gRPC stream messages sent by the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		countsHandlingTimeEnabled: false,
		countsHandlingTimeOpts: prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "grpc_server_handling_seconds",
			Help:      "Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.",
			Buckets:   prometheus.DefBuckets,
		},
		countsHandlingTime: nil,
	}
}

// EnableCountsHandlingTime enables histograms being registered when
// registering the ServerMetrics on a Prometheus registry. Histograms can be
// expensive on Prometheus servers. It takes options to configure histogram
// options such as the defined buckets.
func (m *GrpcServerMetrics) EnableCountsHandlingTime() {
	if !m.countsHandlingTimeEnabled {
		m.countsHandlingTime = prometheus.NewHistogramVec(
			m.countsHandlingTimeOpts,
			[]string{"grpc_type", "grpc_service", "grpc_method"},
		)
	}
	m.countsHandlingTimeEnabled = true
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *GrpcServerMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.startedRequest.Describe(ch)
	m.handledRequest.Describe(ch)
	m.receivedStreamMsg.Describe(ch)
	m.sentStreamMsg.Describe(ch)
	if m.countsHandlingTimeEnabled {
		m.countsHandlingTime.Describe(ch)
	}
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *GrpcServerMetrics) Collect(ch chan<- prometheus.Metric) {
	m.startedRequest.Collect(ch)
	m.handledRequest.Collect(ch)
	m.receivedStreamMsg.Collect(ch)
	m.sentStreamMsg.Collect(ch)
	if m.countsHandlingTimeEnabled {
		m.countsHandlingTime.Collect(ch)
	}
}

// InitializeMetrics initializes all metrics, with their appropriate null
// value, for all gRPC methods registered on a gRPC server. This is useful, to
// ensure that all metrics exist when collecting and querying.
func (m *GrpcServerMetrics) InitializeMetrics(server *grpc.Server) {
	serviceInfo := server.GetServiceInfo()
	for serviceName, info := range serviceInfo {
		for _, mInfo := range info.Methods {
			m.preRegisterMethod(serviceName, &mInfo)
		}
	}
}

// preRegisterMethod is invoked on Register of a Server, allowing all gRPC services labels to be pre-populated.
func (m *GrpcServerMetrics) preRegisterMethod(serviceName string, mInfo *grpc.MethodInfo) {
	methodName := mInfo.Name
	methodType := grpcTypeFromMethodInfo(mInfo)
	// These are just references (no increments), as just referencing will create the labels but not set values.
	ignore.Results(m.startedRequest.GetMetricWithLabelValues(methodType, serviceName, methodName))
	if methodType != Unary {
		ignore.Results(m.receivedStreamMsg.GetMetricWithLabelValues(methodType, serviceName, methodName))
	}
	ignore.Results(m.sentStreamMsg.GetMetricWithLabelValues(methodType, serviceName, methodName))
	if m.countsHandlingTimeEnabled {
		ignore.Results(m.countsHandlingTime.GetMetricWithLabelValues(methodType, serviceName, methodName))
	}
	for _, code := range grpcCodes {
		ignore.Results(m.handledRequest.GetMetricWithLabelValues(methodType, serviceName, methodName, code.String()))
	}
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *GrpcServerMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		monitor := newGrpcServerMonitor(m, Unary, info.FullMethod)
		resp, err := handler(ctx, req)
		s, _ := status.FromError(err)
		monitor.Handled(s.Code())
		return resp, err
	}
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (m *GrpcServerMetrics) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		monitor := newGrpcServerMonitor(m, grpcStreamType(info), info.FullMethod)
		err := handler(srv, &monitoredServerStream{ss, monitor})
		s, _ := status.FromError(err)
		monitor.Handled(s.Code())
		return err
	}
}

type grpcServerMonitor struct {
	metrics     *GrpcServerMetrics
	grpcType    string
	serviceName string
	methodName  string
	startTime   time.Time
}

func newGrpcServerMonitor(metrics *GrpcServerMetrics, grpcType string, fullMethod string) *grpcServerMonitor {
	monitor := &grpcServerMonitor{
		metrics:  metrics,
		grpcType: grpcType,
	}
	if monitor.metrics.countsHandlingTimeEnabled {
		monitor.startTime = time.Now()
	}
	monitor.serviceName, monitor.methodName = splitGrpcMethodName(fullMethod)
	monitor.metrics.startedRequest.WithLabelValues(monitor.grpcType, monitor.serviceName, monitor.methodName).Inc()
	return monitor
}

func (m *grpcServerMonitor) ReceivedMessage() {
	m.metrics.receivedStreamMsg.WithLabelValues(m.grpcType, m.serviceName, m.methodName).Inc()
}

func (m *grpcServerMonitor) SentMessage() {
	m.metrics.sentStreamMsg.WithLabelValues(m.grpcType, m.serviceName, m.methodName).Inc()
}

func (m *grpcServerMonitor) Handled(code codes.Code) {
	m.metrics.handledRequest.WithLabelValues(m.grpcType, m.serviceName, m.methodName, code.String()).Inc()
	if m.metrics.countsHandlingTimeEnabled {
		m.metrics.countsHandlingTime.WithLabelValues(m.grpcType, m.serviceName, m.methodName).Observe(time.Since(m.startTime).Seconds())
	}
}

// monitoredServerStream wraps grpc.ServerStream allowing each Sent/Recv of message to increment counters.
type monitoredServerStream struct {
	grpc.ServerStream
	monitor *grpcServerMonitor
}

func (ss *monitoredServerStream) SendMsg(m interface{}) error {
	err := ss.ServerStream.SendMsg(m)
	if err == nil {
		ss.monitor.SentMessage()
	}
	return err
}

func (ss *monitoredServerStream) RecvMsg(m interface{}) error {
	err := ss.ServerStream.RecvMsg(m)
	if err == nil {
		ss.monitor.ReceivedMessage()
	}
	return err
}
