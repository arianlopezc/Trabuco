package mcp

import (
	"strings"
	"testing"
)

// =============================================================================
// buildSystemDesign: multi-service decomposition
// =============================================================================

func TestSystemDesign_UserAndNotificationServices(t *testing.T) {
	design := buildSystemDesign("I need a user service and a notification service")
	if len(design.Services) != 2 {
		t.Fatalf("Expected 2 services, got %d: %v", len(design.Services), serviceNames(design.Services))
	}
	assertServiceExists(t, design, "user-service", "rest-api")
	assertServiceExists(t, design, "notification-service", "background-processing")
}

func TestSystemDesign_APIGatewayAndBackendServices(t *testing.T) {
	design := buildSystemDesign("System with an API gateway, user service, and order service")
	assertServiceExists(t, design, "api-gateway", "microservice-light")
	assertServiceExists(t, design, "user-service", "rest-api")
	assertServiceExists(t, design, "order-service", "rest-api")
}

func TestSystemDesign_NoDuplicateFromOverlappingKeywords(t *testing.T) {
	// "api gateway" should NOT also create a separate "gateway" service
	design := buildSystemDesign("Build an API gateway for the platform")
	gatewayCount := 0
	for _, svc := range design.Services {
		if svc.Name == "gateway" || svc.Name == "api-gateway" {
			gatewayCount++
		}
	}
	if gatewayCount != 1 {
		t.Errorf("Expected exactly 1 gateway service, got %d: %v", gatewayCount, serviceNames(design.Services))
	}
}

func TestSystemDesign_NotificationNoDuplicate(t *testing.T) {
	// "notification service" should NOT also create a standalone "notification" service
	design := buildSystemDesign("We need a notification service for emails")
	notifCount := 0
	for _, svc := range design.Services {
		if svc.Name == "notification" || svc.Name == "notification-service" {
			notifCount++
		}
	}
	if notifCount != 1 {
		t.Errorf("Expected exactly 1 notification service, got %d: %v", notifCount, serviceNames(design.Services))
	}
}

func TestSystemDesign_AnalyticsNoDuplicate(t *testing.T) {
	// "analytics service" should NOT also create a standalone "analytics" service
	design := buildSystemDesign("We need an analytics service for tracking")
	analyticsCount := 0
	for _, svc := range design.Services {
		if svc.Name == "analytics" || svc.Name == "analytics-service" {
			analyticsCount++
		}
	}
	if analyticsCount != 1 {
		t.Errorf("Expected exactly 1 analytics service, got %d: %v", analyticsCount, serviceNames(design.Services))
	}
}

func TestSystemDesign_SearchServiceGetsNoSQL(t *testing.T) {
	design := buildSystemDesign("search service for product catalog")
	svc := findService(design, "search-service")
	if svc == nil {
		t.Fatal("Expected search-service")
	}
	if svc.Pattern != "rest-api-nosql" {
		t.Errorf("Expected rest-api-nosql pattern for search, got '%s'", svc.Pattern)
	}
	if !strings.Contains(svc.Modules, "NoSQLDatastore") {
		t.Errorf("Expected NoSQLDatastore module for search, got '%s'", svc.Modules)
	}
}

func TestSystemDesign_AnalyticsGetsEventDriven(t *testing.T) {
	design := buildSystemDesign("analytics service for tracking user behavior")
	svc := findService(design, "analytics-service")
	if svc == nil {
		t.Fatal("Expected analytics-service")
	}
	if svc.Pattern != "event-driven" {
		t.Errorf("Expected event-driven pattern for analytics, got '%s'", svc.Pattern)
	}
	if !strings.Contains(svc.Modules, "EventConsumer") {
		t.Errorf("Expected EventConsumer module for analytics, got '%s'", svc.Modules)
	}
}

func TestSystemDesign_EmailServiceGetsWorker(t *testing.T) {
	design := buildSystemDesign("email service for sending transactional emails")
	svc := findService(design, "email-service")
	if svc == nil {
		t.Fatal("Expected email-service")
	}
	if svc.Pattern != "background-processing" {
		t.Errorf("Expected background-processing pattern for email, got '%s'", svc.Pattern)
	}
	if !strings.Contains(svc.Modules, "Worker") {
		t.Errorf("Expected Worker module for email, got '%s'", svc.Modules)
	}
}

func TestSystemDesign_FallbackToSingleService(t *testing.T) {
	// When no specific service keywords match, should fall back to pattern scoring
	design := buildSystemDesign("REST API backend with PostgreSQL database")
	if len(design.Services) != 1 {
		t.Fatalf("Expected 1 fallback service, got %d", len(design.Services))
	}
	if design.Services[0].Name != "main-service" {
		t.Errorf("Expected fallback service named 'main-service', got '%s'", design.Services[0].Name)
	}
	// Should have a warning about no distinct boundaries
	foundWarning := false
	for _, w := range design.Warnings {
		if strings.Contains(w, "Could not identify distinct service boundaries") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("Expected warning about no distinct service boundaries")
	}
}

func TestSystemDesign_GibberishFallsBack(t *testing.T) {
	design := buildSystemDesign("xyzzy plugh nothing")
	// No service keywords match and no pattern keywords either
	foundWarning := false
	for _, w := range design.Warnings {
		if strings.Contains(w, "Could not identify") {
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Error("Expected warning for gibberish input")
	}
}

// =============================================================================
// System design: shared infrastructure detection
// =============================================================================

func TestSystemDesign_SharedInfraIncludesDB(t *testing.T) {
	design := buildSystemDesign("user service and order service")
	// Both use rest-api pattern which has postgresql
	foundDB := false
	for _, infra := range design.SharedInfra {
		if strings.Contains(infra, "PostgreSQL") {
			foundDB = true
		}
	}
	if !foundDB {
		t.Errorf("Expected PostgreSQL in shared infra, got %v", design.SharedInfra)
	}
}

func TestSystemDesign_SharedInfraIncludesBroker(t *testing.T) {
	design := buildSystemDesign("analytics service and logging service")
	// Both use event-driven pattern which has kafka
	foundBroker := false
	for _, infra := range design.SharedInfra {
		if strings.Contains(infra, "broker") || strings.Contains(infra, "Broker") {
			foundBroker = true
		}
	}
	if !foundBroker {
		t.Errorf("Expected message broker in shared infra, got %v", design.SharedInfra)
	}
}

func TestSystemDesign_CommunicationNotesForMultiService(t *testing.T) {
	design := buildSystemDesign("user service and notification service")
	if len(design.CommunicationNotes) == 0 {
		t.Error("Expected communication notes for multi-service design")
	}
}

func TestSystemDesign_NoCommunicationNotesForSingleService(t *testing.T) {
	// Single service fallback — no inter-service communication needed
	design := buildSystemDesign("REST API with database")
	if len(design.CommunicationNotes) != 0 {
		t.Error("Expected no communication notes for single service")
	}
}

// =============================================================================
// System design: unsupported requirement warnings
// =============================================================================

func TestSystemDesign_WarnsAboutAuth(t *testing.T) {
	design := buildSystemDesign("auth service with JWT login and user service")
	foundAuth := false
	for _, w := range design.Warnings {
		if strings.Contains(w, "Authentication") {
			foundAuth = true
		}
	}
	if !foundAuth {
		t.Error("Expected auth unsupported warning")
	}
}

func TestSystemDesign_WarnsAboutKubernetes(t *testing.T) {
	design := buildSystemDesign("deploy user service to kubernetes with helm charts")
	foundK8s := false
	for _, w := range design.Warnings {
		if strings.Contains(w, "Kubernetes") {
			foundK8s = true
		}
	}
	if !foundK8s {
		t.Error("Expected Kubernetes unsupported warning")
	}
}

// =============================================================================
// System design: service group IDs
// =============================================================================

func TestSystemDesign_ServiceGroupIDs(t *testing.T) {
	design := buildSystemDesign("user service and order service")
	for _, svc := range design.Services {
		if svc.GroupID == "" {
			t.Errorf("Service '%s' has empty group_id", svc.Name)
		}
		if !strings.HasPrefix(svc.GroupID, "com.company.") {
			t.Errorf("Service '%s' group_id should start with 'com.company.', got '%s'", svc.Name, svc.GroupID)
		}
	}
}

func TestSystemDesign_ServiceBoundaries(t *testing.T) {
	design := buildSystemDesign("user service and order service")
	for _, svc := range design.Services {
		if len(svc.Boundaries) == 0 {
			t.Errorf("Service '%s' should have boundaries", svc.Name)
		}
	}
}

// =============================================================================
// System design: nil-safe slices (JSON serialization)
// =============================================================================

func TestSystemDesign_NilSafeSlices(t *testing.T) {
	design := buildSystemDesign("xyzzy plugh nothing")
	if design.Warnings == nil {
		t.Error("Warnings should be empty slice, not nil")
	}
	if design.SharedInfra == nil {
		t.Error("SharedInfra should be empty slice, not nil")
	}
	if design.CommunicationNotes == nil {
		t.Error("CommunicationNotes should be empty slice, not nil")
	}
}

// =============================================================================
// buildSharedDockerCompose tests
// =============================================================================

func TestDockerCompose_PostgreSQL(t *testing.T) {
	services := []serviceConfig{
		{Name: "svc-a", Database: "postgresql"},
	}
	compose := buildSharedDockerCompose(services)
	if !strings.Contains(compose, "postgres:") {
		t.Error("Expected postgres service in compose")
	}
	if !strings.Contains(compose, "postgres_data:") {
		t.Error("Expected postgres_data volume")
	}
}

func TestDockerCompose_MySQL(t *testing.T) {
	services := []serviceConfig{
		{Name: "svc-a", Database: "mysql"},
	}
	compose := buildSharedDockerCompose(services)
	if !strings.Contains(compose, "mysql:") {
		t.Error("Expected mysql service in compose")
	}
	if !strings.Contains(compose, "mysql_data:") {
		t.Error("Expected mysql_data volume")
	}
}

func TestDockerCompose_MongoDB(t *testing.T) {
	services := []serviceConfig{
		{Name: "svc-a", NoSQLDatabase: "mongodb"},
	}
	compose := buildSharedDockerCompose(services)
	if !strings.Contains(compose, "mongodb:") {
		t.Error("Expected mongodb service in compose")
	}
	if !strings.Contains(compose, "mongo_data:") {
		t.Error("Expected mongo_data volume")
	}
}

func TestDockerCompose_Redis(t *testing.T) {
	services := []serviceConfig{
		{Name: "svc-a", NoSQLDatabase: "redis"},
	}
	compose := buildSharedDockerCompose(services)
	if !strings.Contains(compose, "redis:") {
		t.Error("Expected redis service in compose")
	}
	if !strings.Contains(compose, "redis_data:") {
		t.Error("Expected redis_data volume")
	}
}

func TestDockerCompose_Kafka(t *testing.T) {
	services := []serviceConfig{
		{Name: "svc-a", MessageBroker: "kafka"},
	}
	compose := buildSharedDockerCompose(services)
	if !strings.Contains(compose, "kafka:") {
		t.Error("Expected kafka service in compose")
	}
	if !strings.Contains(compose, "KAFKA_") {
		t.Error("Expected Kafka environment variables")
	}
}

func TestDockerCompose_RabbitMQ(t *testing.T) {
	services := []serviceConfig{
		{Name: "svc-a", MessageBroker: "rabbitmq"},
	}
	compose := buildSharedDockerCompose(services)
	if !strings.Contains(compose, "rabbitmq:") {
		t.Error("Expected rabbitmq service in compose")
	}
	if !strings.Contains(compose, "15672") {
		t.Error("Expected RabbitMQ management port")
	}
}

func TestDockerCompose_MixedInfrastructure(t *testing.T) {
	// Multiple services with different infrastructure needs
	services := []serviceConfig{
		{Name: "api-svc", Database: "postgresql"},
		{Name: "search-svc", NoSQLDatabase: "mongodb"},
		{Name: "events-svc", Database: "postgresql", MessageBroker: "kafka"},
	}
	compose := buildSharedDockerCompose(services)
	if !strings.Contains(compose, "postgres:") {
		t.Error("Expected postgres")
	}
	if !strings.Contains(compose, "mongodb:") {
		t.Error("Expected mongodb")
	}
	if !strings.Contains(compose, "kafka:") {
		t.Error("Expected kafka")
	}
	// Should NOT have MySQL, Redis, or RabbitMQ
	if strings.Contains(compose, "mysql:") {
		t.Error("Should NOT have mysql")
	}
	if strings.Contains(compose, "redis:") {
		t.Error("Should NOT have redis")
	}
	if strings.Contains(compose, "rabbitmq:") {
		t.Error("Should NOT have rabbitmq")
	}
}

func TestDockerCompose_DeduplicatesInfra(t *testing.T) {
	// Two services both using PostgreSQL should produce ONE postgres entry
	services := []serviceConfig{
		{Name: "svc-a", Database: "postgresql"},
		{Name: "svc-b", Database: "postgresql"},
	}
	compose := buildSharedDockerCompose(services)
	count := strings.Count(compose, "postgres:")
	// "postgres:" appears once as service name and once in volume reference;
	// should not have two postgres service definitions
	if count > 2 {
		t.Errorf("Expected at most 2 occurrences of 'postgres:' (service + volume), got %d", count)
	}
}

func TestDockerCompose_NoServicesNoVolumes(t *testing.T) {
	// Services with no infrastructure
	services := []serviceConfig{
		{Name: "stateless-svc"},
	}
	compose := buildSharedDockerCompose(services)
	if strings.Contains(compose, "volumes:") {
		t.Error("Expected no volumes section for stateless services")
	}
}

// =============================================================================
// Complex multi-service scenarios
// =============================================================================

func TestSystemDesign_EcommercePlatform(t *testing.T) {
	design := buildSystemDesign("E-commerce platform with catalog service, order service, payment service, and notification service")
	if len(design.Services) < 4 {
		t.Errorf("Expected at least 4 services, got %d: %v", len(design.Services), serviceNames(design.Services))
	}
	assertServiceExists(t, design, "catalog-service", "rest-api")
	assertServiceExists(t, design, "order-service", "rest-api")
	assertServiceExists(t, design, "payment-service", "rest-api")
	assertServiceExists(t, design, "notification-service", "background-processing")
}

func TestSystemDesign_MicroservicesWithGateway(t *testing.T) {
	design := buildSystemDesign("API gateway with user service, inventory service, and analytics service")
	assertServiceExists(t, design, "api-gateway", "microservice-light")
	assertServiceExists(t, design, "user-service", "rest-api")
	assertServiceExists(t, design, "inventory-service", "rest-api")
	assertServiceExists(t, design, "analytics-service", "event-driven")

	// Gateway should be lightweight (Model, Shared, API only)
	gw := findService(design, "api-gateway")
	if gw != nil && strings.Contains(gw.Modules, "SQLDatastore") {
		t.Error("Gateway should not have SQLDatastore")
	}
}

func TestSystemDesign_ModuleCoverageAllPatterns(t *testing.T) {
	// Design a system that exercises all distinct module combinations
	design := buildSystemDesign(
		"System with: API gateway for routing, " +
			"user service for accounts, " +
			"search service for products, " +
			"notification service for emails, " +
			"analytics service for tracking, " +
			"email service for transactional messages")

	// Collect all unique modules across all services
	allModules := make(map[string]bool)
	for _, svc := range design.Services {
		for _, mod := range strings.Split(svc.Modules, ",") {
			allModules[strings.TrimSpace(mod)] = true
		}
	}

	// Should cover: Model, SQLDatastore, NoSQLDatastore, Shared, API, Worker, EventConsumer
	expected := []string{"Model", "Shared", "API", "SQLDatastore", "NoSQLDatastore", "Worker", "EventConsumer"}
	for _, mod := range expected {
		if !allModules[mod] {
			t.Errorf("Expected module '%s' to appear across services. Got modules: %v", mod, allModules)
		}
	}
}

// =============================================================================
// New service types added in polish pass
// =============================================================================

func TestSystemDesign_BillingService(t *testing.T) {
	design := buildSystemDesign("billing service for invoices and subscriptions")
	assertServiceExists(t, design, "billing-service", "rest-api")
}

func TestSystemDesign_ReportingService(t *testing.T) {
	design := buildSystemDesign("reporting service for generating daily reports")
	assertServiceExists(t, design, "reporting-service", "background-processing")
}

func TestSystemDesign_SchedulingService(t *testing.T) {
	design := buildSystemDesign("scheduling service for cron jobs")
	assertServiceExists(t, design, "scheduling-service", "background-processing")
}

func TestSystemDesign_WorkflowService(t *testing.T) {
	design := buildSystemDesign("workflow service for multi-step processes")
	assertServiceExists(t, design, "workflow-service", "background-processing")
}

func TestSystemDesign_AdminService(t *testing.T) {
	design := buildSystemDesign("admin service for internal management")
	assertServiceExists(t, design, "admin-service", "rest-api")
}

func TestSystemDesign_WebhookService(t *testing.T) {
	design := buildSystemDesign("webhook service to receive Stripe callbacks")
	assertServiceExists(t, design, "webhook-service", "rest-api")
}

func TestSystemDesign_IntegrationService(t *testing.T) {
	design := buildSystemDesign("integration service for third-party APIs")
	assertServiceExists(t, design, "integration-service", "rest-api")
}

func TestSystemDesign_AuditService(t *testing.T) {
	design := buildSystemDesign("audit service for tracking changes")
	assertServiceExists(t, design, "audit-service", "event-driven")
}

func TestSystemDesign_CacheService(t *testing.T) {
	design := buildSystemDesign("cache service for session storage")
	assertServiceExists(t, design, "cache-service", "rest-api-nosql")
}

func TestSystemDesign_ImportService(t *testing.T) {
	design := buildSystemDesign("import service for batch data ingestion")
	assertServiceExists(t, design, "import-service", "worker-only")
}

func TestSystemDesign_DataService(t *testing.T) {
	design := buildSystemDesign("data service for persistence layer")
	assertServiceExists(t, design, "data-service", "rest-api")
}

// =============================================================================
// Complex scenario with new service types
// =============================================================================

func TestSystemDesign_SaaSPlatform(t *testing.T) {
	design := buildSystemDesign(
		"SaaS platform with: user service, billing service, " +
			"notification service, admin service, and audit service")
	if len(design.Services) < 5 {
		t.Errorf("Expected at least 5 services, got %d: %v", len(design.Services), serviceNames(design.Services))
	}
	assertServiceExists(t, design, "user-service", "rest-api")
	assertServiceExists(t, design, "billing-service", "rest-api")
	assertServiceExists(t, design, "notification-service", "background-processing")
	assertServiceExists(t, design, "admin-service", "rest-api")
	assertServiceExists(t, design, "audit-service", "event-driven")
}

// =============================================================================
// Test helpers
// =============================================================================

func serviceNames(services []serviceDesign) []string {
	names := make([]string, len(services))
	for i, s := range services {
		names[i] = s.Name
	}
	return names
}

func findService(design *systemDesign, name string) *serviceDesign {
	for i := range design.Services {
		if design.Services[i].Name == name {
			return &design.Services[i]
		}
	}
	return nil
}

func assertServiceExists(t *testing.T, design *systemDesign, name, expectedPattern string) {
	t.Helper()
	svc := findService(design, name)
	if svc == nil {
		t.Errorf("Expected service '%s' not found. Services: %v", name, serviceNames(design.Services))
		return
	}
	if svc.Pattern != expectedPattern {
		t.Errorf("Service '%s': expected pattern '%s', got '%s'", name, expectedPattern, svc.Pattern)
	}
}
