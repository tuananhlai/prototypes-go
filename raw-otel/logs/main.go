package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	collectorlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	collectorURL := "http://localhost:14318/v1/logs"

	req := buildExportLogsRequest("raw-otel-logs", "hello world!")
	res, err := exportLogs(collectorURL, req)
	if err != nil {
		log.Fatalf("error exporting logs: %v", err)
	}

	log.Printf("Request succeeded: %+v\n", res)
}

func buildExportLogsRequest(serviceName, body string) *collectorlogsv1.ExportLogsServiceRequest {
	resource := resourcev1.Resource{
		Attributes: []*commonv1.KeyValue{
			{
				Key: "service.name",
				Value: &commonv1.AnyValue{
					Value: &commonv1.AnyValue_StringValue{StringValue: serviceName},
				},
			},
		},
	}

	// LogRecord stores the actual log line.
	record := &logsv1.LogRecord{
		SeverityNumber: logsv1.SeverityNumber_SEVERITY_NUMBER_INFO,
		Body: &commonv1.AnyValue{
			Value: &commonv1.AnyValue_StringValue{StringValue: body},
		},
	}

	// ScopeLogs applies a scope for multiple log records.
	scopeLogs := &logsv1.ScopeLogs{
		LogRecords: []*logsv1.LogRecord{record},
		Scope: &commonv1.InstrumentationScope{
			Name: "github.com/tuananhlai/prototypes/raw-otel",
		},
	}

	// ResourceLogs includes metadata about the log producer.
	resourceLogs := &logsv1.ResourceLogs{
		Resource:  &resource,
		ScopeLogs: []*logsv1.ScopeLogs{scopeLogs},
	}

	retval := collectorlogsv1.ExportLogsServiceRequest{
		ResourceLogs: []*logsv1.ResourceLogs{resourceLogs},
	}

	return &retval
}

// exportLogs sends a batch of log records to otel collector.
func exportLogs(url string, req *collectorlogsv1.ExportLogsServiceRequest) (*collectorlogsv1.ExportLogsServiceResponse, error) {
	body, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("content-type", "application/x-protobuf")

	httpClient := http.Client{}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error invalid response status code: %v", resp.StatusCode)
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	exportLogsResponse := &collectorlogsv1.ExportLogsServiceResponse{}
	err = proto.Unmarshal(rawBody, exportLogsResponse)
	if err != nil {
		return nil, err
	}

	return exportLogsResponse, nil
}
