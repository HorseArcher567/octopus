package otelsvc

import (
	"github.com/k8s-practice/octopus/pkg/log"
	"github.com/k8s-practice/octopus/pkg/service"
	"github.com/k8s-practice/octopus/pkg/util/structure"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.10.0"
	"sync"
)

type Config struct {
	OpenTelemetry struct {
		Enabled      bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Name         string `json:"name,omitempty" yaml:"name,omitempty"`
		Namespace    string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
		Service      string `json:"service,omitempty" yaml:"service,omitempty"`
		SpanExporter struct {
			Name              string `json:"name,omitempty" yaml:"name,omitempty"`
			CollectorEndpoint string `json:"collectorEndpoint,omitempty" yaml:"collectorEndpoint,omitempty"`
		} `json:"spanExporter,omitempty" yaml:"spanExporter,omitempty"`
	} `json:"openTelemetry,omitempty" yaml:"openTelemetry,omitempty"`
}

type builder struct {
	sync.Once
	service *Service
}

func (b *builder) Build(bootConfig map[interface{}]interface{}, tag string) service.Entry {
	b.Do(func() {
		var conf Config
		if err := structure.UnmarshalWithTag(bootConfig, &conf, tag); err != nil {
			log.Panicln(err)
			return
		}

		svc := &Service{
			enabled: conf.OpenTelemetry.Enabled,
			name:    conf.OpenTelemetry.Name,
		}
		if conf.OpenTelemetry.Enabled {
			exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
				jaeger.WithEndpoint(conf.OpenTelemetry.SpanExporter.CollectorEndpoint)))
			if err != nil {
				log.Panicln(err)
				return
			}

			tp := sdkTrace.NewTracerProvider(
				sdkTrace.WithResource(resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceNamespaceKey.String(conf.OpenTelemetry.Namespace),
					semconv.ServiceNameKey.String(conf.OpenTelemetry.Service),
				)),
				sdkTrace.WithSampler(sdkTrace.ParentBased(sdkTrace.AlwaysSample())),
				sdkTrace.WithSyncer(exporter),
			)
			otel.SetTracerProvider(tp)
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{}))
		}

		b.service = svc
	})

	return b.service
}
