package apimtracer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

type ApimTracePayload struct {
	ApiName          string  `json:"api_name"`
	SubscriptionName string  `json:"subscription_name"`
	ProductName      string  `json:"product_name"`
	TraceParent      *string `json:"traceparent,omitempty"`
	ParentSpanId     *string `json:"parent_span_id,omitempty"`
	HttpTarget       string  `json:"http_target"`
	HttpMethod       string  `json:"http_method"`
	HttpHost         string  `json:"http_host"`
	HttpStatusCode   int     `json:"http_status_code"`
	HttpClientIp     string  `json:"http_client_ip"`
	RequestStartTime int64   `json:"request_start_time"`
	RequestEndTime   int64   `json:"request_end_time"`
}

func mapToTrace(eventData string) (*ptrace.Traces, error) {
	traces := ptrace.NewTraces()

	// convert eventData json string to a ApimTracePayload struct
	apimTracePayload := ApimTracePayload{}
	err := json.Unmarshal([]byte(eventData), &apimTracePayload)
	if err != nil {
		return nil, err
	}

	resourceSpan := traces.ResourceSpans().AppendEmpty()
	resource := resourceSpan.Resource()
	resourceAttrs := resource.Attributes()
	resourceAttrs.PutStr("apim.product.name", apimTracePayload.ProductName)
	resourceAttrs.PutStr(conventions.AttributeServiceName, "apim")
	resourceAttrs.PutStr("apim.api.name", apimTracePayload.ApiName)
	resourceAttrs.PutStr("apim.subscription.name", apimTracePayload.SubscriptionName)

	scopeSpans := resourceSpan.ScopeSpans().AppendEmpty()

	apiSpan := scopeSpans.Spans().AppendEmpty()
	apiSpan.SetName(fmt.Sprintf("apim: %s %s %s", apimTracePayload.HttpMethod, apimTracePayload.ApiName, apimTracePayload.HttpTarget))
	// example traceparent: 00-a9b7d8bbcecf8c7d9d24c69cc86f694d-2defb8587ed6dfef-01
	// format: version-trace_id-parent_id-trace_flags
	fmt.Printf("Extracting span and trace id from traceparent %s", *apimTracePayload.TraceParent)
	traceString := (*apimTracePayload.TraceParent)[3:35]
	traceId, err := uuid.Parse(traceString)
	if err != nil {
		return nil, fmt.Errorf("error converting traceId to uuid: traceString: %s err: %v", traceString, err)
	} else {
		apiSpan.SetTraceID(pcommon.TraceID(traceId))
	}

	decodedSpanId, err := hex.DecodeString((*apimTracePayload.TraceParent)[36:52])
	apiSpan.SetSpanID(pcommon.SpanID(decodedSpanId))
	fmt.Printf("Extracted TraceId=%s and SpanId=%s, apiSpanId=%s", (*apimTracePayload.TraceParent)[3:35], (*apimTracePayload.TraceParent)[36:52], apiSpan.SpanID().String())
	if apimTracePayload.ParentSpanId != nil {
		apiSpan.SetParentSpanID(pcommon.SpanID([]byte(*apimTracePayload.ParentSpanId)))
	}

	apiSpan.SetKind(ptrace.SpanKindServer)
	apiSpan.Status().SetCode(ptrace.StatusCodeUnset)
	// request start and end time are provided in unix milliseconds
	apiSpan.SetStartTimestamp(pcommon.NewTimestampFromTime(time.UnixMilli(apimTracePayload.RequestStartTime))) // example: OTel Start time     : 2024-12-10 14:26:32.435836303 +0000 UTC
	apiSpan.SetEndTimestamp(pcommon.NewTimestampFromTime(time.UnixMilli(apimTracePayload.RequestEndTime)))

	attributes := apiSpan.Attributes()
	attributes.PutStr(conventions.AttributeHTTPTarget, apimTracePayload.HttpTarget)
	attributes.PutStr(conventions.AttributeHTTPMethod, apimTracePayload.HttpMethod)
	attributes.PutStr(conventions.AttributeHTTPHost, apimTracePayload.HttpHost)
	attributes.PutStr(conventions.AttributeHTTPClientIP, apimTracePayload.HttpClientIp)
	attributes.PutInt(conventions.AttributeHTTPStatusCode, int64(apimTracePayload.HttpStatusCode))
	return &traces, nil
}
