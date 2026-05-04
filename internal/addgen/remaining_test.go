package addgen

import (
	"strings"
	"testing"
)

// --- Event ---

func TestGenerateEvent(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	result, err := GenerateEvent(ctx, EventOpts{
		Name:   "OrderShipped",
		Fields: "orderId:string,shippedAt:instant,carrierRef:string?",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Created) != 1 {
		t.Fatalf("expected 1 file, got %v", result.Created)
	}
	body := readPath(t, project, result.Created[0])
	wants := []string{
		"package com.example.demo.model.events;",
		"public record OrderShipped(",
		"@NotNull String orderId",
		"@NotNull Instant shippedAt",
		"@Nullable String carrierRef",
		"import java.time.Instant;",
	}
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Errorf("event missing %q\ngot:\n%s", w, body)
		}
	}
}

func TestGenerateEvent_Errors(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	cases := []struct {
		name    string
		opts    EventOpts
		wantErr string
	}{
		{"empty name", EventOpts{Fields: "x:string"}, "name is required"},
		{"lowercase name", EventOpts{Name: "myEvent", Fields: "x:string"}, "PascalCase"},
		{"empty fields", EventOpts{Name: "Foo"}, "must not be empty"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GenerateEvent(ctx, tc.opts)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

// --- Service ---

func TestGenerateService_NoEntity(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	result, err := GenerateService(ctx, ServiceOpts{Name: "NotificationService"})
	if err != nil {
		t.Fatal(err)
	}
	body := readPath(t, project, result.Created[0])
	wants := []string{
		"package com.example.demo.shared.service;",
		"@Service",
		"public class NotificationService {",
		"public NotificationService()",
		"throw new UnsupportedOperationException",
	}
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Errorf("service missing %q\ngot:\n%s", w, body)
		}
	}
	// No repo import when --entity is unset
	if strings.Contains(body, "Repository") {
		t.Errorf("expected no Repository reference without --entity, got:\n%s", body)
	}
}

func TestGenerateService_WithEntity_SQL(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	result, err := GenerateService(ctx, ServiceOpts{Name: "OrderService", Entity: "Order"})
	if err != nil {
		t.Fatal(err)
	}
	body := readPath(t, project, result.Created[0])
	wants := []string{
		"import com.example.demo.sqldatastore.repository.OrderRepository;",
		"private final OrderRepository orderRepository;",
		"public OrderService(OrderRepository orderRepository) {",
		"this.orderRepository = orderRepository;",
	}
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Errorf("service missing %q\ngot:\n%s", w, body)
		}
	}
}

func TestGenerateService_WithEntity_Mongo(t *testing.T) {
	project := setupProject(t, map[string]string{
		".trabuco.json": `{
  "version": "1.13.2", "projectName": "demo", "groupId": "com.example.demo",
  "artifactId": "demo", "javaVersion": "21",
  "modules": ["Model", "NoSQLDatastore", "Shared", "API"], "noSqlDatabase": "mongodb"
}`,
	})
	ctx := mustCtx(t, project)
	result, err := GenerateService(ctx, ServiceOpts{Name: "OrderService", Entity: "Order"})
	if err != nil {
		t.Fatal(err)
	}
	body := readPath(t, project, result.Created[0])
	if !strings.Contains(body, "import com.example.demo.nosqldatastore.repository.OrderDocumentRepository;") {
		t.Errorf("expected Mongo document repository import:\n%s", body)
	}
}

func TestGenerateService_Errors(t *testing.T) {
	project := setupProject(t, map[string]string{
		".trabuco.json": `{"version":"1.13.2","projectName":"demo","groupId":"com.example.demo","artifactId":"demo","javaVersion":"21","modules":["Model","API"]}`,
	})
	ctx := mustCtx(t, project)
	_, err := GenerateService(ctx, ServiceOpts{Name: "Foo"})
	if err == nil || !strings.Contains(err.Error(), "Shared module") {
		t.Fatalf("expected Shared module error, got %v", err)
	}
}

// --- Job ---

func TestGenerateJob(t *testing.T) {
	project := setupProject(t, map[string]string{
		".trabuco.json": `{
  "version": "1.13.2", "projectName": "demo", "groupId": "com.example.demo",
  "artifactId": "demo", "javaVersion": "21",
  "modules": ["Model", "SQLDatastore", "Shared", "API", "Worker"], "database": "postgresql"
}`,
	})
	ctx := mustCtx(t, project)
	result, err := GenerateJob(ctx, JobOpts{
		Name:    "ProcessShipment",
		Payload: "orderId:string,priority:integer",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"Model/src/main/java/com/example/demo/model/jobs/ProcessShipmentJobRequest.java",
		"Model/src/main/java/com/example/demo/model/jobs/ProcessShipmentJobRequestHandler.java",
		"Worker/src/main/java/com/example/demo/worker/handler/ProcessShipmentJobRequestHandler.java",
	}
	for _, p := range want {
		if !contains(result.Created, p) {
			t.Errorf("missing expected file %s in %v", p, result.Created)
		}
	}

	req := readPath(t, project, want[0])
	base := readPath(t, project, want[1])
	concrete := readPath(t, project, want[2])
	combined := req + base + concrete

	for _, w := range []string{
		"public record ProcessShipmentJobRequest(",
		"String orderId",
		"Integer priority",
		"implements JobRequest {",
		"return ProcessShipmentJobRequestHandler.class;",
		"public class ProcessShipmentJobRequestHandler implements JobRequestHandler<ProcessShipmentJobRequest>",
		"@Component\npublic class ProcessShipmentJobRequestHandler",
		"extends com.example.demo.model.jobs.ProcessShipmentJobRequestHandler",
		"@Job(name = \"ProcessShipment\")",
	} {
		if !strings.Contains(combined, w) {
			t.Errorf("job bundle missing %q", w)
		}
	}
}

func TestGenerateJob_RequiresWorker(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	_, err := GenerateJob(ctx, JobOpts{Name: "Foo", Payload: "x:string"})
	if err == nil || !strings.Contains(err.Error(), "Worker") {
		t.Fatalf("expected Worker module error, got %v", err)
	}
}

// --- Endpoint ---

func TestGenerateEndpoint_Plain(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	result, err := GenerateEndpoint(ctx, EndpointOpts{Name: "Health", Path: "/healthz"})
	if err != nil {
		t.Fatal(err)
	}
	body := readPath(t, project, result.Created[0])
	wants := []string{
		"@RestController",
		"@RequestMapping(\"/healthz\")",
		"public class HealthController {",
	}
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Errorf("plain endpoint missing %q\ngot:\n%s", w, body)
		}
	}
	if strings.Contains(body, "@PostMapping") {
		t.Errorf("plain endpoint should not have CRUD methods:\n%s", body)
	}
}

func TestGenerateEndpoint_CRUD(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	result, err := GenerateEndpoint(ctx, EndpointOpts{Name: "Order", Type: "crud"})
	if err != nil {
		t.Fatal(err)
	}
	body := readPath(t, project, result.Created[0])
	wants := []string{
		"@RequestMapping(\"/api/orders\")", // auto-derived
		"@PostMapping",
		"@GetMapping(\"/{id}\")",
		"@GetMapping",
		"@PutMapping(\"/{id}\")",
		"@DeleteMapping(\"/{id}\")",
		"public ResponseEntity<Object> create",
		"public ResponseEntity<Object> getById",
		"public ResponseEntity<Object> list",
		"public ResponseEntity<Object> update",
		"public ResponseEntity<Void> delete",
	}
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Errorf("CRUD endpoint missing %q\ngot:\n%s", w, body)
		}
	}
}

func TestGenerateEndpoint_Errors(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	cases := []struct {
		name string
		opts EndpointOpts
		want string
	}{
		{"empty name", EndpointOpts{}, "name is required"},
		{"lowercase name", EndpointOpts{Name: "order"}, "PascalCase"},
		{"unknown type", EndpointOpts{Name: "Order", Type: "rest"}, "must be plain or crud"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GenerateEndpoint(ctx, tc.opts)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
		})
	}
}

// --- StreamingEndpoint ---

func TestGenerateStreamingEndpoint(t *testing.T) {
	project := setupProject(t, map[string]string{
		".trabuco.json": `{
  "version": "1.13.2", "projectName": "demo", "groupId": "com.example.demo",
  "artifactId": "demo", "javaVersion": "21",
  "modules": ["Model", "Shared", "AIAgent", "API"]
}`,
	})
	ctx := mustCtx(t, project)
	result, err := GenerateStreamingEndpoint(ctx, StreamingEndpointOpts{Name: "Conversation"})
	if err != nil {
		t.Fatal(err)
	}
	body := readPath(t, project, result.Created[0])
	wants := []string{
		"package com.example.demo.aiagent.protocol;",
		"@RestController",
		"public class ConversationStreamController {",
		"@GetMapping(value = \"/api/agent/stream/conversation/{id}\", produces = MediaType.TEXT_EVENT_STREAM_VALUE)",
		"SseEmitter",
		"Thread.startVirtualThread",
	}
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Errorf("streaming endpoint missing %q\ngot:\n%s", w, body)
		}
	}
}

func TestGenerateStreamingEndpoint_RequiresAIAgent(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	_, err := GenerateStreamingEndpoint(ctx, StreamingEndpointOpts{Name: "Conversation"})
	if err == nil || !strings.Contains(err.Error(), "AIAgent") {
		t.Fatalf("expected AIAgent module error, got %v", err)
	}
}
