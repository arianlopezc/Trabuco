# Trabuco CDK Extension -- Complete Design Document

## Table of Contents

1. [Overview](#1-overview)
2. [New ProjectConfig Fields](#2-new-projectconfig-fields)
3. [New CDK Module Types](#3-new-cdk-module-types)
4. [Function Module Redesign](#4-function-module-redesign)
5. [Build System](#5-build-system)
6. [Template Inventory](#6-template-inventory)
7. [AI Agent Integration for CDK](#7-ai-agent-integration-for-cdk)
8. [Cost Optimization Constructs](#8-cost-optimization-constructs)
9. [CLI UX](#9-cli-ux)
10. [Go Implementation Details](#10-go-implementation-details)
11. [Testing Strategy](#11-testing-strategy)

---

## 1. Overview

### Design Philosophy

CDK projects are a distinct **deployment target**, not a bolt-on to the existing Spring Boot generation path. When a user selects `--target=cdk`, the entire generation pipeline shifts: the build system changes from Maven to Gradle, the runtime modules produce Lambda handlers instead of Spring Boot applications, infrastructure is defined in Java CDK stacks, and the CI pipeline deploys via CDK Pipelines instead of Docker images.

The existing `Model` and `Shared` modules are reused as-is (they contain domain objects and business logic with no runtime coupling). The runtime modules (`API`, `Worker`, `EventConsumer`) are replaced by Lambda-native equivalents. New CDK infrastructure modules are added.

### Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Build system | Gradle (Kotlin DSL) | CDK Java projects use Gradle by convention; better Lambda packaging with Shadow plugin; faster builds than Maven for CDK synth/deploy cycles |
| Lambda runtime | Plain AWS SDK handlers (no framework) | Fastest cold start; smallest JAR; SnapStart compatible. Quarkus/Micronaut offered as opt-in via `--lambda-framework` flag |
| CDK version | CDK v2 (aws-cdk-lib) | v1 is deprecated; v2 is the only supported version |
| Stack organization | One stack per concern (Stateful, Api, Async, Monitoring, Pipeline) | Matches AWS Well-Architected; enables independent deployments |
| DynamoDB over RDS | DynamoDB is the default datastore | Truly serverless; no VPC required; pay-per-request; RDS Proxy offered as opt-in |

---

## 2. New ProjectConfig Fields

### Additions to `config/project.go`

```go
type ProjectConfig struct {
    // --- Existing fields (unchanged) ---
    ProjectName string
    GroupID     string
    ArtifactID  string
    JavaVersion         string
    JavaVersionDetected bool
    Modules     []string
    Database    string
    NoSQLDatabase string
    MessageBroker string
    AIAgents    []string
    CIProvider  string
    IncludeCLAUDEMD bool // deprecated

    // --- New CDK fields ---

    // DeploymentTarget selects the generation pipeline.
    // "springboot" (default, current behavior) or "cdk" (new serverless pipeline).
    DeploymentTarget string

    // AWS-specific configuration (only when DeploymentTarget == "cdk")
    AWSRegion       string   // e.g., "us-east-1" (default)
    AWSAccount      string   // optional, for Pipeline stack; e.g., "123456789012"
    CDKModules      []string // e.g., ["CDK-Infra", "CDK-StatefulStack", "CDK-ApiStack", ...]

    // Lambda configuration
    LambdaFramework string // "none" (plain AWS SDK, default), "quarkus", "micronaut"
    LambdaMemoryMB  int    // default 512
    EnableSnapStart bool   // default true for Java 21

    // CDK-specific datastore (replaces Database/NoSQLDatabase for CDK)
    // DynamoDB is always available in CDK mode. RDS is opt-in.
    EnableDynamoDB  bool   // default true
    EnableRDS       bool   // default false (requires VPC)
    EnableS3        bool   // default false
    EnableCognito   bool   // default false

    // Async infrastructure
    EnableSQS          bool // default false
    EnableEventBridge  bool // default false
    EnableStepFunctions bool // default false
}
```

### New Constants in `config/modules.go`

```go
// Deployment target constants
const (
    TargetSpringBoot = "springboot"
    TargetCDK        = "cdk"
)

// CDK module name constants
const (
    ModuleCDKInfra      = "CDK-Infra"
    ModuleCDKStateful   = "CDK-StatefulStack"
    ModuleCDKApi        = "CDK-ApiStack"
    ModuleCDKAsync      = "CDK-AsyncStack"
    ModuleCDKMonitoring = "CDK-MonitoringStack"
    ModuleCDKPipeline   = "CDK-Pipeline"
)

// Lambda function module constants (CDK runtime modules)
const (
    ModuleLambdaApi     = "LambdaApi"
    ModuleLambdaWorker  = "LambdaWorker"
    ModuleLambdaEvents  = "LambdaEvents"
)

// Lambda framework constants
const (
    LambdaFrameworkNone      = "none"
    LambdaFrameworkQuarkus   = "quarkus"
    LambdaFrameworkMicronaut = "micronaut"
)
```

### New Helper Methods on ProjectConfig

```go
func (c *ProjectConfig) IsCDK() bool {
    return c.DeploymentTarget == TargetCDK
}

func (c *ProjectConfig) IsSpringBoot() bool {
    return c.DeploymentTarget == "" || c.DeploymentTarget == TargetSpringBoot
}

func (c *ProjectConfig) HasCDKModule(name string) bool {
    for _, m := range c.CDKModules {
        if m == name { return true }
    }
    return false
}

func (c *ProjectConfig) UsesPlainLambda() bool {
    return c.LambdaFramework == "" || c.LambdaFramework == LambdaFrameworkNone
}

func (c *ProjectConfig) UsesQuarkus() bool {
    return c.LambdaFramework == LambdaFrameworkQuarkus
}

func (c *ProjectConfig) UsesMicronaut() bool {
    return c.LambdaFramework == LambdaFrameworkMicronaut
}

func (c *ProjectConfig) NeedsVPC() bool {
    return c.EnableRDS
}

func (c *ProjectConfig) HasAnyAsyncService() bool {
    return c.EnableSQS || c.EnableEventBridge || c.EnableStepFunctions
}

func (c *ProjectConfig) CDKAppClassName() string {
    return c.ProjectNamePascal() + "App"
}

func (c *ProjectConfig) StackPrefix() string {
    return c.ProjectNamePascal()
}
```

### Updated ProjectMetadata

```go
type ProjectMetadata struct {
    // ... existing fields ...
    DeploymentTarget string   `json:"deploymentTarget,omitempty"`
    AWSRegion        string   `json:"awsRegion,omitempty"`
    CDKModules       []string `json:"cdkModules,omitempty"`
    LambdaFramework  string   `json:"lambdaFramework,omitempty"`
}
```

---

## 3. New CDK Module Types

### 3.1 CDK-Infra (Required, Auto-included)

**Description:** Core CDK application scaffolding. Defines the CDK App entry point, stage abstraction for multi-environment deployments, and shared configuration (environment names, account/region lookups).

**Dependencies:** None (this is the CDK root).

**CDK Constructs:** `App`, `Stage`, `Environment`.

**Files Generated:**

| Template Path | Output Path | Description |
|--------------|-------------|-------------|
| `cdk/infra/CdkApp.java.tmpl` | `infra/src/main/java/{pkg}/infra/{Name}App.java` | CDK App entry point |
| `cdk/infra/AppStage.java.tmpl` | `infra/src/main/java/{pkg}/infra/{Name}Stage.java` | Stage aggregating all stacks |
| `cdk/infra/EnvironmentConfig.java.tmpl` | `infra/src/main/java/{pkg}/infra/EnvironmentConfig.java` | Environment enum (dev/staging/prod) with account/region |
| `cdk/infra/Tags.java.tmpl` | `infra/src/main/java/{pkg}/infra/Tags.java` | Standard tagging utility |
| `gradle/infra.build.gradle.kts.tmpl` | `infra/build.gradle.kts` | Gradle build for CDK infra module |
| `cdk/cdk.json.tmpl` | `cdk.json` | CDK toolkit configuration |
| `cdk/infra/test/AppStageTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/{Name}StageTest.java` | Snapshot test for synthesized stacks |

### 3.2 CDK-StatefulStack

**Description:** Stateful AWS resources that persist data and survive redeployments. DynamoDB tables, S3 buckets, Cognito user pools. These are placed in a separate stack so they can be deployed independently and have `RemovalPolicy.RETAIN`.

**Dependencies:** CDK-Infra.

**CDK Constructs:** `Table` (DynamoDB), `Bucket` (S3), `UserPool` (Cognito), `StringParameter` (SSM for cross-stack references).

**Files Generated:**

| Template Path | Output Path | Description |
|--------------|-------------|-------------|
| `cdk/stateful/StatefulStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/StatefulStack.java` | Main stateful stack |
| `cdk/stateful/DynamoDbConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/DynamoDbTables.java` | DynamoDB table definitions (conditional on `EnableDynamoDB`) |
| `cdk/stateful/S3Construct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/S3Buckets.java` | S3 bucket definitions (conditional on `EnableS3`) |
| `cdk/stateful/CognitoConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/CognitoAuth.java` | Cognito user pool (conditional on `EnableCognito`) |
| `cdk/stateful/RdsConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/RdsDatabase.java` | RDS + RDS Proxy (conditional on `EnableRDS`) |
| `cdk/stateful/test/StatefulStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/StatefulStackTest.java` | Assertion tests for stateful resources |

### 3.3 CDK-ApiStack

**Description:** API Gateway REST API backed by Lambda functions. Creates the API Gateway, Lambda integrations, custom domain (optional), WAF association, and CloudWatch logging.

**Dependencies:** CDK-Infra, CDK-StatefulStack (for table grants), LambdaApi module.

**CDK Constructs:** `RestApi`, `LambdaIntegration`, `Function` (via L3 `OptimizedJavaFunction`), `CfnStage`, `LogGroup`.

**Files Generated:**

| Template Path | Output Path | Description |
|--------------|-------------|-------------|
| `cdk/api/ApiStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/ApiStack.java` | API Gateway + Lambda stack |
| `cdk/api/ApiLambdaConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/ApiLambdas.java` | Lambda function definitions for API |
| `cdk/api/test/ApiStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/ApiStackTest.java` | Assertion tests |

### 3.4 CDK-AsyncStack

**Description:** Asynchronous processing infrastructure. SQS queues (with DLQ), EventBridge rules, Step Functions state machines, and the Lambda functions that process them.

**Dependencies:** CDK-Infra, CDK-StatefulStack, LambdaWorker or LambdaEvents module.

**CDK Constructs:** `Queue` (SQS), `Rule` (EventBridge), `StateMachine` (Step Functions), `SqsEventSource`, `Function`.

**Files Generated:**

| Template Path | Output Path | Description |
|--------------|-------------|-------------|
| `cdk/async/AsyncStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/AsyncStack.java` | Async processing stack |
| `cdk/async/SqsConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/SqsQueues.java` | SQS queues with DLQ (conditional on `EnableSQS`) |
| `cdk/async/EventBridgeConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/EventBridgeRules.java` | EventBridge bus and rules (conditional on `EnableEventBridge`) |
| `cdk/async/StepFunctionsConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/StepFunctionsWorkflow.java` | Step Functions state machine (conditional on `EnableStepFunctions`) |
| `cdk/async/test/AsyncStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/AsyncStackTest.java` | Assertion tests |

### 3.5 CDK-MonitoringStack

**Description:** CloudWatch dashboards, alarms, and SNS topics for operational visibility. Generates per-function dashboards, composite alarms, and anomaly detection.

**Dependencies:** CDK-Infra, CDK-ApiStack or CDK-AsyncStack (for metric sources).

**CDK Constructs:** `Dashboard`, `GraphWidget`, `Alarm`, `CompositeAlarm`, `Topic` (SNS), `SnsAction`.

**Files Generated:**

| Template Path | Output Path | Description |
|--------------|-------------|-------------|
| `cdk/monitoring/MonitoringStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/MonitoringStack.java` | Monitoring stack |
| `cdk/monitoring/LambdaDashboard.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/LambdaDashboard.java` | Per-function CloudWatch dashboard |
| `cdk/monitoring/ApiAlarms.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/ApiAlarms.java` | API Gateway 4xx/5xx alarms |
| `cdk/monitoring/test/MonitoringStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/MonitoringStackTest.java` | Assertion tests |

### 3.6 CDK-Pipeline

**Description:** CDK Pipelines CI/CD that self-mutates. Connects to GitHub, runs synth, deploys through dev -> staging -> prod with manual approval gates.

**Dependencies:** CDK-Infra (uses AppStage).

**CDK Constructs:** `CodePipeline`, `ShellStep`, `ManualApprovalStep`, `CodePipelineSource.gitHub()`.

**Files Generated:**

| Template Path | Output Path | Description |
|--------------|-------------|-------------|
| `cdk/pipeline/PipelineStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/PipelineStack.java` | CDK Pipelines stack |
| `cdk/pipeline/test/PipelineStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/PipelineStackTest.java` | Pipeline assertion tests |

### CDK Module Registry

```go
var CDKModuleRegistry = []Module{
    {
        Name:        ModuleCDKInfra,
        Description: "CDK app, stages, environment config",
        Required:    true,
        Internal:    false,
        Dependencies: []string{},
    },
    {
        Name:        ModuleCDKStateful,
        Description: "DynamoDB, S3, Cognito, RDS (stateful resources)",
        Required:    false,
        Internal:    false,
        Dependencies: []string{ModuleCDKInfra},
    },
    {
        Name:        ModuleCDKApi,
        Description: "API Gateway + Lambda REST API",
        Required:    false,
        Internal:    false,
        Dependencies: []string{ModuleCDKInfra, ModuleCDKStateful},
    },
    {
        Name:        ModuleCDKAsync,
        Description: "SQS, EventBridge, Step Functions",
        Required:    false,
        Internal:    false,
        Dependencies: []string{ModuleCDKInfra, ModuleCDKStateful},
    },
    {
        Name:        ModuleCDKMonitoring,
        Description: "CloudWatch dashboards and alarms",
        Required:    false,
        Internal:    false,
        Dependencies: []string{ModuleCDKInfra},
    },
    {
        Name:        ModuleCDKPipeline,
        Description: "CDK Pipelines CI/CD (self-mutating)",
        Required:    false,
        Internal:    false,
        Dependencies: []string{ModuleCDKInfra},
    },
}
```

---

## 4. Function Module Redesign

### 4.1 Design Principle

Lambda handlers must be as lightweight as possible. Spring Boot is **never** used for Lambda in CDK mode. The default is plain AWS SDK `RequestHandler` implementations with no dependency injection framework. Quarkus and Micronaut are opt-in via `--lambda-framework`.

Each function module produces a **separate shadow JAR** via Gradle's Shadow plugin. This keeps deployment artifacts small (typically 5-15 MB for plain SDK, 20-40 MB for Quarkus native).

### 4.2 LambdaApi Module

Replaces the Spring Boot `API` module. Contains Lambda handlers that back API Gateway routes.

**Handler Pattern (Plain SDK):**

```java
package {groupId}.lambdaapi.handler;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;

public class GetPlaceholderHandler
    implements RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {

    private final PlaceholderService service;

    public GetPlaceholderHandler() {
        // Manual wiring -- no DI framework
        this.service = ServiceFactory.createPlaceholderService();
    }

    @Override
    public APIGatewayProxyResponseEvent handleRequest(
            APIGatewayProxyRequestEvent input, Context context) {
        // ...
    }
}
```

**Handler Pattern (Quarkus):**

```java
package {groupId}.lambdaapi.handler;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;
import jakarta.inject.Inject;
import jakarta.inject.Named;

@Named("getPlaceholder")
public class GetPlaceholderHandler
    implements RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {

    @Inject
    PlaceholderService service;

    @Override
    public APIGatewayProxyResponseEvent handleRequest(
            APIGatewayProxyRequestEvent input, Context context) {
        // ...
    }
}
```

**Files Generated:**

| Template Path | Output Path |
|--------------|-------------|
| `java/lambdaapi/handler/GetPlaceholderHandler.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/handler/GetPlaceholderHandler.java` |
| `java/lambdaapi/handler/CreatePlaceholderHandler.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/handler/CreatePlaceholderHandler.java` |
| `java/lambdaapi/handler/ListPlaceholdersHandler.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/handler/ListPlaceholdersHandler.java` |
| `java/lambdaapi/ServiceFactory.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/ServiceFactory.java` |
| `java/lambdaapi/util/ApiGatewayResponse.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/util/ApiGatewayResponse.java` |
| `java/lambdaapi/util/JsonUtil.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/util/JsonUtil.java` |
| `java/lambdaapi/test/GetPlaceholderHandlerTest.java.tmpl` | `lambda-api/src/test/java/{pkg}/lambdaapi/handler/GetPlaceholderHandlerTest.java` |
| `gradle/lambda-api.build.gradle.kts.tmpl` | `lambda-api/build.gradle.kts` |

### 4.3 LambdaWorker Module

Replaces the Spring Boot `Worker` module. Contains Lambda handlers triggered by SQS queues, scheduled EventBridge rules, or Step Functions tasks.

**Handler Pattern:**

```java
package {groupId}.lambdaworker.handler;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.SQSEvent;

public class ProcessPlaceholderHandler implements RequestHandler<SQSEvent, Void> {

    @Override
    public Void handleRequest(SQSEvent event, Context context) {
        for (SQSEvent.SQSMessage message : event.getRecords()) {
            // process message
        }
        return null;
    }
}
```

**Files Generated:**

| Template Path | Output Path |
|--------------|-------------|
| `java/lambdaworker/handler/ProcessPlaceholderHandler.java.tmpl` | `lambda-worker/src/main/java/{pkg}/lambdaworker/handler/ProcessPlaceholderHandler.java` |
| `java/lambdaworker/handler/ScheduledTaskHandler.java.tmpl` | `lambda-worker/src/main/java/{pkg}/lambdaworker/handler/ScheduledTaskHandler.java` |
| `java/lambdaworker/ServiceFactory.java.tmpl` | `lambda-worker/src/main/java/{pkg}/lambdaworker/ServiceFactory.java` |
| `java/lambdaworker/test/ProcessPlaceholderHandlerTest.java.tmpl` | `lambda-worker/src/test/java/{pkg}/lambdaworker/handler/ProcessPlaceholderHandlerTest.java` |
| `gradle/lambda-worker.build.gradle.kts.tmpl` | `lambda-worker/build.gradle.kts` |

### 4.4 LambdaEvents Module

Replaces the Spring Boot `EventConsumer` module. Contains Lambda handlers triggered by EventBridge events.

**Files Generated:**

| Template Path | Output Path |
|--------------|-------------|
| `java/lambdaevents/handler/PlaceholderEventHandler.java.tmpl` | `lambda-events/src/main/java/{pkg}/lambdaevents/handler/PlaceholderEventHandler.java` |
| `java/lambdaevents/ServiceFactory.java.tmpl` | `lambda-events/src/main/java/{pkg}/lambdaevents/ServiceFactory.java` |
| `java/lambdaevents/test/PlaceholderEventHandlerTest.java.tmpl` | `lambda-events/src/test/java/{pkg}/lambdaevents/handler/PlaceholderEventHandlerTest.java` |
| `gradle/lambda-events.build.gradle.kts.tmpl` | `lambda-events/build.gradle.kts` |

### 4.5 SnapStart Compatibility

All Lambda handlers are SnapStart-compatible by default when Java 21 is selected:

- No lazy initialization patterns that break SnapStart
- `CRaC` hooks generated in `ServiceFactory` for connection priming
- CDK stacks set `snapStart.applyOn = "PublishedVersions"` automatically
- Generated comment blocks explain SnapStart constraints

```java
// In ServiceFactory.java (plain SDK mode):
import org.crac.Core;
import org.crac.Resource;

public class ServiceFactory implements Resource {
    private static DynamoDbClient dynamoClient;

    static {
        Core.getGlobalContext().register(new ServiceFactory());
        dynamoClient = DynamoDbClient.create();
    }

    @Override
    public void beforeCheckpoint(org.crac.Context<? extends Resource> context) {
        // Connection priming for SnapStart
    }

    @Override
    public void afterRestore(org.crac.Context<? extends Resource> context) {
        // Re-establish connections after restore
    }
}
```

### 4.6 Module Reuse

The `Model` and `Shared` modules are **reused unchanged** in CDK mode. They contain only domain objects (Immutables) and business logic (services) with no Spring Boot dependencies. The Gradle build files for these modules strip out the Spring Boot parent and use plain Java library configuration:

```kotlin
// Model/build.gradle.kts
plugins {
    `java-library`
}

dependencies {
    api("org.immutables:value:2.10.1")
    annotationProcessor("org.immutables:value:2.10.1")
}
```

---

## 5. Build System

### 5.1 Gradle Over Maven for CDK

CDK projects use Gradle with Kotlin DSL. Rationale:

1. **CDK convention** -- `cdk init` generates Gradle projects; all AWS CDK Java examples use Gradle
2. **Shadow JAR** -- The Shadow plugin produces single-JAR Lambda deployment artifacts; Maven Shade is more verbose
3. **Build speed** -- Gradle's incremental compilation and configuration caching make `cdk synth` faster
4. **Multi-project** -- Gradle's composite builds and dependency substitution handle the CDK + Lambda multi-module layout more naturally

### 5.2 Root Build Configuration

**File:** `build.gradle.kts` (project root)

```kotlin
plugins {
    java
    id("com.github.johnrengelman.shadow") version "8.1.1" apply false
    id("com.diffplug.spotless") version "7.0.2"
}

allprojects {
    group = "{{.GroupID}}"
    version = "1.0-SNAPSHOT"

    repositories {
        mavenCentral()
    }
}

subprojects {
    apply(plugin = "java")
    apply(plugin = "com.diffplug.spotless")

    java {
        sourceCompatibility = JavaVersion.VERSION_{{.JavaVersion}}
        targetCompatibility = JavaVersion.VERSION_{{.JavaVersion}}
    }

    spotless {
        java {
            palantirJavaFormat("2.50.0")
            removeUnusedImports()
            trimTrailingWhitespace()
        }
    }

    tasks.withType<Test> {
        useJUnitPlatform()
    }
}
```

**File:** `settings.gradle.kts`

```kotlin
rootProject.name = "{{.ProjectName}}"

include("model")
include("shared")
{{- if .HasCDKModule "CDK-ApiStack"}}
include("lambda-api")
{{- end}}
{{- if .HasCDKModule "CDK-AsyncStack"}}
include("lambda-worker")
include("lambda-events")
{{- end}}
include("infra")
```

### 5.3 Lambda Module Build Configuration

Each Lambda module uses the Shadow plugin to produce a fat JAR:

```kotlin
// lambda-api/build.gradle.kts
plugins {
    java
    id("com.github.johnrengelman.shadow")
}

dependencies {
    implementation(project(":model"))
    implementation(project(":shared"))
    implementation("com.amazonaws:aws-lambda-java-core:1.2.3")
    implementation("com.amazonaws:aws-lambda-java-events:3.14.0")
    implementation("software.amazon.awssdk:dynamodb:2.29.51")
    implementation("com.fasterxml.jackson.core:jackson-databind:2.18.2")
    {{- if .EnableSnapStart}}
    implementation("io.github.crac:org-crac:0.1.3")
    {{- end}}

    testImplementation("org.junit.jupiter:junit-jupiter:5.11.4")
    testImplementation("org.mockito:mockito-core:5.14.2")
}

tasks.shadowJar {
    archiveClassifier.set("")
    mergeServiceFiles()
}
```

### 5.4 CDK Infra Build Configuration

```kotlin
// infra/build.gradle.kts
plugins {
    java
    application
}

application {
    mainClass.set("{{.GroupID}}.infra.{{.CDKAppClassName}}")
}

dependencies {
    implementation("software.amazon.awscdk:aws-cdk-lib:2.178.2")
    implementation("software.constructs:constructs:[10.0.0,11.0.0)")

    // Reference Lambda JARs for CDK Code.fromAsset()
    {{- if .HasCDKModule "CDK-ApiStack"}}
    implementation(project(":lambda-api"))
    {{- end}}
    {{- if .HasCDKModule "CDK-AsyncStack"}}
    implementation(project(":lambda-worker"))
    implementation(project(":lambda-events"))
    {{- end}}

    testImplementation("org.junit.jupiter:junit-jupiter:5.11.4")
    testImplementation("org.assertj:assertj-core:3.27.3")
}
```

---

## 6. Template Inventory

### 6.1 Gradle Build Templates (11 files)

| Template Path | Output Path |
|--------------|-------------|
| `gradle/root.build.gradle.kts.tmpl` | `build.gradle.kts` |
| `gradle/settings.gradle.kts.tmpl` | `settings.gradle.kts` |
| `gradle/gradle.properties.tmpl` | `gradle.properties` |
| `gradle/model.build.gradle.kts.tmpl` | `model/build.gradle.kts` |
| `gradle/shared.build.gradle.kts.tmpl` | `shared/build.gradle.kts` |
| `gradle/lambda-api.build.gradle.kts.tmpl` | `lambda-api/build.gradle.kts` |
| `gradle/lambda-worker.build.gradle.kts.tmpl` | `lambda-worker/build.gradle.kts` |
| `gradle/lambda-events.build.gradle.kts.tmpl` | `lambda-events/build.gradle.kts` |
| `gradle/infra.build.gradle.kts.tmpl` | `infra/build.gradle.kts` |
| `gradle/wrapper/gradle-wrapper.properties.tmpl` | `gradle/wrapper/gradle-wrapper.properties` |
| `gradle/libs.versions.toml.tmpl` | `gradle/libs.versions.toml` |

### 6.2 CDK Infrastructure Templates (20 files)

| Template Path | Output Path |
|--------------|-------------|
| `cdk/cdk.json.tmpl` | `cdk.json` |
| `cdk/infra/CdkApp.java.tmpl` | `infra/src/main/java/{pkg}/infra/{Name}App.java` |
| `cdk/infra/AppStage.java.tmpl` | `infra/src/main/java/{pkg}/infra/{Name}Stage.java` |
| `cdk/infra/EnvironmentConfig.java.tmpl` | `infra/src/main/java/{pkg}/infra/EnvironmentConfig.java` |
| `cdk/infra/Tags.java.tmpl` | `infra/src/main/java/{pkg}/infra/Tags.java` |
| `cdk/stateful/StatefulStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/StatefulStack.java` |
| `cdk/stateful/DynamoDbConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/DynamoDbTables.java` |
| `cdk/stateful/S3Construct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/S3Buckets.java` |
| `cdk/stateful/CognitoConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/CognitoAuth.java` |
| `cdk/stateful/RdsConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/RdsDatabase.java` |
| `cdk/api/ApiStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/ApiStack.java` |
| `cdk/api/ApiLambdaConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/ApiLambdas.java` |
| `cdk/async/AsyncStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/AsyncStack.java` |
| `cdk/async/SqsConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/SqsQueues.java` |
| `cdk/async/EventBridgeConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/EventBridgeRules.java` |
| `cdk/async/StepFunctionsConstruct.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/StepFunctionsWorkflow.java` |
| `cdk/monitoring/MonitoringStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/MonitoringStack.java` |
| `cdk/monitoring/LambdaDashboard.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/LambdaDashboard.java` |
| `cdk/monitoring/ApiAlarms.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/ApiAlarms.java` |
| `cdk/pipeline/PipelineStack.java.tmpl` | `infra/src/main/java/{pkg}/infra/stacks/PipelineStack.java` |

### 6.3 L3 Construct Templates (4 files)

| Template Path | Output Path |
|--------------|-------------|
| `cdk/constructs/OptimizedJavaFunction.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/OptimizedJavaFunction.java` |
| `cdk/constructs/EventDrivenWorker.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/EventDrivenWorker.java` |
| `cdk/constructs/SecuredApi.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/SecuredApi.java` |
| `cdk/constructs/MonitoredFunction.java.tmpl` | `infra/src/main/java/{pkg}/infra/constructs/MonitoredFunction.java` |

### 6.4 Lambda Java Source Templates (16 files)

| Template Path | Output Path |
|--------------|-------------|
| `java/lambdaapi/handler/GetPlaceholderHandler.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/handler/GetPlaceholderHandler.java` |
| `java/lambdaapi/handler/CreatePlaceholderHandler.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/handler/CreatePlaceholderHandler.java` |
| `java/lambdaapi/handler/ListPlaceholdersHandler.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/handler/ListPlaceholdersHandler.java` |
| `java/lambdaapi/ServiceFactory.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/ServiceFactory.java` |
| `java/lambdaapi/util/ApiGatewayResponse.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/util/ApiGatewayResponse.java` |
| `java/lambdaapi/util/JsonUtil.java.tmpl` | `lambda-api/src/main/java/{pkg}/lambdaapi/util/JsonUtil.java` |
| `java/lambdaworker/handler/ProcessPlaceholderHandler.java.tmpl` | `lambda-worker/src/main/java/{pkg}/lambdaworker/handler/ProcessPlaceholderHandler.java` |
| `java/lambdaworker/handler/ScheduledTaskHandler.java.tmpl` | `lambda-worker/src/main/java/{pkg}/lambdaworker/handler/ScheduledTaskHandler.java` |
| `java/lambdaworker/ServiceFactory.java.tmpl` | `lambda-worker/src/main/java/{pkg}/lambdaworker/ServiceFactory.java` |
| `java/lambdaevents/handler/PlaceholderEventHandler.java.tmpl` | `lambda-events/src/main/java/{pkg}/lambdaevents/handler/PlaceholderEventHandler.java` |
| `java/lambdaevents/ServiceFactory.java.tmpl` | `lambda-events/src/main/java/{pkg}/lambdaevents/ServiceFactory.java` |
| `java/lambdaapi/test/GetPlaceholderHandlerTest.java.tmpl` | `lambda-api/src/test/java/{pkg}/lambdaapi/handler/GetPlaceholderHandlerTest.java` |
| `java/lambdaapi/test/CreatePlaceholderHandlerTest.java.tmpl` | `lambda-api/src/test/java/{pkg}/lambdaapi/handler/CreatePlaceholderHandlerTest.java` |
| `java/lambdaworker/test/ProcessPlaceholderHandlerTest.java.tmpl` | `lambda-worker/src/test/java/{pkg}/lambdaworker/handler/ProcessPlaceholderHandlerTest.java` |
| `java/lambdaevents/test/PlaceholderEventHandlerTest.java.tmpl` | `lambda-events/src/test/java/{pkg}/lambdaevents/handler/PlaceholderEventHandlerTest.java` |
| `java/lambdacommon/LambdaLogger.java.tmpl` | `lambda-common/src/main/java/{pkg}/lambdacommon/LambdaLogger.java` |

### 6.5 CDK Test Templates (6 files)

| Template Path | Output Path |
|--------------|-------------|
| `cdk/infra/test/AppStageTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/{Name}StageTest.java` |
| `cdk/stateful/test/StatefulStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/StatefulStackTest.java` |
| `cdk/api/test/ApiStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/ApiStackTest.java` |
| `cdk/async/test/AsyncStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/AsyncStackTest.java` |
| `cdk/monitoring/test/MonitoringStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/MonitoringStackTest.java` |
| `cdk/pipeline/test/PipelineStackTest.java.tmpl` | `infra/src/test/java/{pkg}/infra/stacks/PipelineStackTest.java` |

### 6.6 CI/CD Templates (2 files)

| Template Path | Output Path |
|--------------|-------------|
| `github/workflows/cdk-ci.yml.tmpl` | `.github/workflows/ci.yml` |
| `github/workflows/cdk-deploy.yml.tmpl` | `.github/workflows/deploy.yml` |

### 6.7 Documentation Templates (4 files)

| Template Path | Output Path |
|--------------|-------------|
| `docs/cdk/README.md.tmpl` | `README.md` |
| `docs/cdk/CLAUDE.md.tmpl` | `CLAUDE.md` (and other agent files) |
| `docs/cdk/gitignore.tmpl` | `.gitignore` |
| `docs/cdk/ARCHITECTURE.md.tmpl` | `ARCHITECTURE.md` |

### 6.8 AI Prompt Templates for CDK (7 files)

| Template Path | Output Path |
|--------------|-------------|
| `ai/prompts/cdk/add-lambda.md.tmpl` | `.ai/prompts/add-lambda.md` |
| `ai/prompts/cdk/add-stack.md.tmpl` | `.ai/prompts/add-stack.md` |
| `ai/prompts/cdk/add-construct.md.tmpl` | `.ai/prompts/add-construct.md` |
| `ai/prompts/cdk/add-dynamodb-table.md.tmpl` | `.ai/prompts/add-dynamodb-table.md` |
| `ai/prompts/cdk/cdk-patterns.md.tmpl` | `.ai/prompts/cdk-patterns.md` |
| `ai/prompts/cdk/lambda-optimization.md.tmpl` | `.ai/prompts/lambda-optimization.md` |
| `ai/prompts/cdk/testing-guide.md.tmpl` | `.ai/prompts/testing-guide.md` |

### 6.9 SAM Local Development Templates (2 files)

| Template Path | Output Path |
|--------------|-------------|
| `sam/template.yml.tmpl` | `template.yml` |
| `sam/samconfig.toml.tmpl` | `samconfig.toml` |

### 6.10 Existing Templates Reused for CDK

The following existing templates are reused without modification:

- `java/model/ImmutableStyle.java.tmpl`
- `java/model/entities/Placeholder.java.tmpl`
- `java/model/dto/PlaceholderRequest.java.tmpl`
- `java/model/dto/PlaceholderResponse.java.tmpl`
- `java/shared/service/PlaceholderService.java.tmpl` (with template conditionals for CDK)
- `ai/checkpoint.json.tmpl`
- `ai/review-log.jsonl.tmpl`
- `ai/prompts/JAVA_CODE_QUALITY.md.tmpl`
- `ai/prompts/code-review.md.tmpl`
- `claude/settings.json.tmpl` (with CDK-aware commands)
- `claude/skills/commit.md.tmpl`
- `claude/skills/pr.md.tmpl`

### Total New Template Count

| Category | Count |
|----------|-------|
| Gradle build | 11 |
| CDK infrastructure | 20 |
| L3 constructs | 4 |
| Lambda Java source | 16 |
| CDK tests | 6 |
| CI/CD | 2 |
| Documentation | 4 |
| AI prompts | 7 |
| SAM local dev | 2 |
| **Total new** | **72** |
| Reused from existing | 11 |
| **Grand total for CDK project** | **83** |

---

## 7. AI Agent Integration for CDK

### 7.1 CLAUDE.md Template for CDK Projects

The CDK CLAUDE.md template (`templates/docs/cdk/CLAUDE.md.tmpl`) follows the same structure as the existing Spring Boot template but with CDK-specific content:

```
{{- if .Frontmatter -}}
---
{{.Frontmatter}}---
{{end -}}
# {{.ProjectName}}

AWS CDK serverless Java project{{if .EnableDynamoDB}} with DynamoDB{{end}}{{if .EnableRDS}} and RDS{{end}}{{if .HasCDKModule "CDK-ApiStack"}} — API Gateway + Lambda REST API{{end}}{{if .HasCDKModule "CDK-AsyncStack"}} — SQS/EventBridge async processing{{end}}.

## Code Quality (IMPORTANT)

**Code quality specification:** `{{.PromptsDir}}/JAVA_CODE_QUALITY.md`

(same quality table as existing CLAUDE.md)

## Build & Deploy Commands

| Command | Description |
|---------|-------------|
| `./gradlew build` | Build all modules |
| `./gradlew test` | Run all tests |
| `./gradlew :lambda-api:shadowJar` | Build API Lambda deployment JAR |
| `./gradlew :lambda-worker:shadowJar` | Build Worker Lambda deployment JAR |
| `cdk synth` | Synthesize CloudFormation templates |
| `cdk diff` | Show pending infrastructure changes |
| `cdk deploy --all` | Deploy all stacks to AWS |
| `cdk deploy {{.StackPrefix}}StatefulStack` | Deploy only stateful resources |
| `cdk destroy --all` | Tear down all stacks |
| `sam local start-api` | Run API locally with SAM |
| `./gradlew spotlessApply` | Auto-format all Java files |
| `./gradlew spotlessCheck` | Check formatting (CI) |

## Project Structure

```
{{.ProjectName}}/
├── model/                    # Domain objects (Immutables)
├── shared/                   # Business logic (services)
{{- if .HasCDKModule "CDK-ApiStack"}}
├── lambda-api/               # API Gateway Lambda handlers
│   └── src/main/java/{pkg}/lambdaapi/handler/
{{- end}}
{{- if .HasCDKModule "CDK-AsyncStack"}}
├── lambda-worker/            # SQS/scheduled Lambda handlers
├── lambda-events/            # EventBridge Lambda handlers
{{- end}}
├── infra/                    # CDK infrastructure code
│   └── src/main/java/{pkg}/infra/
│       ├── {Name}App.java    # CDK entry point
│       ├── {Name}Stage.java  # Multi-env stage
│       ├── stacks/           # Stack definitions
│       └── constructs/       # Reusable L3 constructs
├── build.gradle.kts          # Root build
├── cdk.json                  # CDK config
└── template.yml              # SAM local dev
```

## Module Dependencies

```
model              → (none)
shared             → model
lambda-api         → model, shared
lambda-worker      → model, shared
lambda-events      → model, shared
infra              → lambda-api, lambda-worker, lambda-events (for Code.fromAsset)
```

Never import from infra in Lambda modules. Never import between Lambda modules.

## CDK Patterns

| Pattern | Description |
|---------|-------------|
| L3 Constructs | Use `OptimizedJavaFunction` for all Lambda functions -- handles SnapStart, memory, timeout, logging |
| Stack References | Pass resources between stacks via constructor props, not SSM lookups |
| Removal Policy | `RETAIN` for DynamoDB/S3/RDS in production, `DESTROY` in dev |
| Tagging | Use `Tags.of(this).add()` via the Tags utility class |
| Environment Config | Use `EnvironmentConfig` enum, not hardcoded account/region strings |
{{- if .EnableSnapStart}}
| SnapStart | All Lambda functions use SnapStart; implement CRaC hooks in ServiceFactory |
{{- end}}

## Lambda Handler Rules

- One handler class per Lambda function (single-responsibility)
- Use `ServiceFactory` for dependency wiring, not constructor parameters
- Return structured `APIGatewayProxyResponseEvent` objects, not raw strings
- Always set CORS headers via the `ApiGatewayResponse` utility
- Log with Lambda Context logger, not SLF4J
- Keep handlers under 30 lines; extract business logic to Shared services

## CDK Stack Rules

- Each stack accepts its dependencies as typed props (e.g., `ApiStackProps`)
- Never use `Fn.importValue` -- pass constructs between stacks via Stage
- Use L3 constructs from `constructs/` package, not raw L1/L2 unless necessary
- Test stacks with assertion-based tests (Template.fromStack), not snapshot tests
```

### 7.2 New MCP Tools for CDK

The existing MCP server (`internal/mcp/tools.go`) gains these new tools when `DeploymentTarget == "cdk"`:

```go
func registerCDKTools(s *server.MCPServer, version string) {
    registerCDKSynth(s)       // Run `cdk synth` and return CloudFormation output
    registerCDKDiff(s)        // Run `cdk diff` and return change summary
    registerCDKDeploy(s)      // Run `cdk deploy` for a specific stack
    registerSAMLocalInvoke(s) // Invoke a Lambda locally with SAM
    registerCDKListStacks(s)  // List all stacks and their status
    registerLambdaLogs(s)     // Fetch recent CloudWatch logs for a Lambda
}
```

Tool descriptions for MCP:

| Tool Name | Description |
|-----------|-------------|
| `cdk_synth` | Synthesize CloudFormation templates. Returns the generated template for a given stack name. |
| `cdk_diff` | Compare deployed stack with local code. Returns a human-readable diff of changes. |
| `cdk_deploy` | Deploy a specific stack. Accepts stack name and optional `--require-approval never` flag. |
| `sam_local_invoke` | Invoke a Lambda function locally using SAM. Accepts function name and event JSON. |
| `cdk_list_stacks` | List all CDK stacks defined in the app with deployment status. |
| `lambda_logs` | Fetch the last N log events from CloudWatch for a specific Lambda function. |

### 7.3 New AI Prompt Templates

**`add-lambda.md`** -- Step-by-step playbook for adding a new Lambda function:
1. Create handler class in the appropriate lambda module
2. Add route/trigger in the corresponding CDK stack
3. Wire dependencies via ServiceFactory
4. Add unit tests
5. Run `cdk synth` to validate
6. Run `cdk diff` to verify changes

**`add-stack.md`** -- Playbook for adding a new CDK stack:
1. Create stack class in `infra/stacks/`
2. Define typed props interface
3. Add to the Stage class
4. Add assertion tests
5. Run `cdk synth`

**`add-construct.md`** -- Playbook for extracting reusable L3 constructs from stacks.

**`add-dynamodb-table.md`** -- Playbook for adding a new DynamoDB table with GSI/LSI, including the CDK definition and the Java DAO class.

**`cdk-patterns.md`** -- Reference for CDK patterns: cross-stack references, environment-specific config, conditional resources, custom resources.

**`lambda-optimization.md`** -- Guide for optimizing Lambda cold starts: SnapStart hooks, SDK client reuse, lazy initialization pitfalls, memory tuning, Powertools for AWS Lambda.

**`testing-guide.md`** -- CDK-specific testing guide covering CDK assertion tests (`Template.fromStack`), Lambda handler unit tests, SAM local integration tests, and end-to-end tests with deployed stacks.

---

## 8. Cost Optimization Constructs

### 8.1 OptimizedJavaFunction

The core L3 construct wrapping `Function` with Java-specific best practices baked in.

```java
package {groupId}.infra.constructs;

import software.amazon.awscdk.Duration;
import software.amazon.awscdk.services.lambda.*;
import software.amazon.awscdk.services.lambda.Runtime;
import software.amazon.awscdk.services.logs.*;
import software.constructs.Construct;

/**
 * L3 construct for Java Lambda functions with production defaults:
 * - SnapStart enabled (Java 21)
 * - ARM64 architecture (Graviton2 -- 20% cheaper)
 * - Structured logging via CloudWatch Logs
 * - X-Ray tracing enabled
 * - Reasonable memory/timeout defaults
 * - Log retention policy
 */
public class OptimizedJavaFunction extends Construct {

    private final Function function;
    private final Alias liveAlias;

    public OptimizedJavaFunction(Construct scope, String id, OptimizedJavaFunctionProps props) {
        super(scope, id);

        this.function = Function.Builder.create(this, "Function")
            .functionName(props.getFunctionName())
            .runtime(Runtime.JAVA_21)
            .architecture(Architecture.ARM_64)
            .handler(props.getHandler())
            .code(Code.fromAsset(props.getCodePath()))
            .memorySize(props.getMemorySize() != null ? props.getMemorySize() : 512)
            .timeout(props.getTimeout() != null ? props.getTimeout() : Duration.seconds(30))
            .tracing(Tracing.ACTIVE)
            .environment(props.getEnvironment())
            .snapStart(SnapStartConf.ON_PUBLISHED_VERSIONS)
            .logRetention(RetentionDays.TWO_WEEKS)
            .build();

        // SnapStart requires a version + alias
        var version = this.function.getCurrentVersion();
        this.liveAlias = Alias.Builder.create(this, "LiveAlias")
            .aliasName("live")
            .version(version)
            .build();
    }

    /** Returns the alias to use for integrations (API Gateway, event sources). */
    public IFunction getAliasOrFunction() {
        return this.liveAlias;
    }

    public Function getFunction() {
        return this.function;
    }
}
```

### 8.2 EventDrivenWorker

L3 construct that bundles an SQS queue with DLQ, a Lambda consumer, and CloudWatch alarms.

```java
/**
 * L3 construct for event-driven Lambda workers with production defaults:
 * - SQS queue with DLQ (maxReceiveCount = 3)
 * - Lambda triggered by SQS with batch size and concurrency controls
 * - CloudWatch alarm on DLQ message count > 0
 * - Visibility timeout auto-calculated from Lambda timeout
 */
public class EventDrivenWorker extends Construct {

    private final Queue queue;
    private final Queue deadLetterQueue;
    private final OptimizedJavaFunction handler;

    public EventDrivenWorker(Construct scope, String id, EventDrivenWorkerProps props) {
        super(scope, id);

        this.deadLetterQueue = Queue.Builder.create(this, "DLQ")
            .queueName(props.getQueueName() + "-dlq")
            .retentionPeriod(Duration.days(14))
            .build();

        this.queue = Queue.Builder.create(this, "Queue")
            .queueName(props.getQueueName())
            .visibilityTimeout(Duration.seconds(
                props.getLambdaTimeout().toSeconds() * 6))
            .deadLetterQueue(DeadLetterQueue.builder()
                .queue(this.deadLetterQueue)
                .maxReceiveCount(3)
                .build())
            .build();

        this.handler = new OptimizedJavaFunction(this, "Handler",
            props.toOptimizedFunctionProps());

        this.handler.getAliasOrFunction().addEventSource(
            SqsEventSource.Builder.create(this.queue)
                .batchSize(props.getBatchSize() != null ? props.getBatchSize() : 10)
                .maxConcurrency(props.getMaxConcurrency() != null ? props.getMaxConcurrency() : 5)
                .build());

        // DLQ alarm
        Alarm.Builder.create(this, "DLQAlarm")
            .metric(this.deadLetterQueue.metricApproximateNumberOfMessagesVisible())
            .threshold(1)
            .evaluationPeriods(1)
            .treatMissingData(TreatMissingData.NOT_BREACHING)
            .build();
    }

    public Queue getQueue() { return this.queue; }
}
```

### 8.3 SecuredApi

L3 construct that creates a REST API with Cognito authorization, request validation, WAF, and throttling.

```java
/**
 * L3 construct for API Gateway with production security:
 * - Cognito authorizer (if Cognito enabled)
 * - Request validators
 * - WAF WebACL association
 * - Throttling (10K req/sec burst, 5K sustained)
 * - Access logging to CloudWatch
 * - CORS configuration
 */
public class SecuredApi extends Construct {

    private final RestApi api;
    private final IAuthorizer authorizer;

    public SecuredApi(Construct scope, String id, SecuredApiProps props) {
        super(scope, id);

        this.api = RestApi.Builder.create(this, "Api")
            .restApiName(props.getApiName())
            .deployOptions(StageOptions.builder()
                .stageName(props.getStageName())
                .throttlingRateLimit(5000)
                .throttlingBurstLimit(10000)
                .accessLogDestination(new LogGroupLogDestination(
                    LogGroup.Builder.create(this, "AccessLog")
                        .retention(RetentionDays.ONE_MONTH)
                        .build()))
                .accessLogFormat(AccessLogFormat.jsonWithStandardFields())
                .tracingEnabled(true)
                .build())
            .defaultCorsPreflightOptions(CorsOptions.builder()
                .allowOrigins(Cors.ALL_ORIGINS)
                .allowMethods(Cors.ALL_METHODS)
                .build())
            .build();

        if (props.getUserPool() != null) {
            this.authorizer = CognitoUserPoolsAuthorizer.Builder
                .create(this, "Authorizer")
                .cognitoUserPools(List.of(props.getUserPool()))
                .build();
        } else {
            this.authorizer = null;
        }
    }

    public RestApi getApi() { return this.api; }
    public IAuthorizer getAuthorizer() { return this.authorizer; }
}
```

### 8.4 MonitoredFunction

L3 construct that wraps `OptimizedJavaFunction` and adds per-function CloudWatch dashboard widgets and alarms.

```java
/**
 * L3 construct that adds monitoring to OptimizedJavaFunction:
 * - Invocation count metric
 * - Error rate alarm (> 1% over 5 minutes)
 * - Duration P99 alarm (> 80% of timeout)
 * - Throttle alarm (any throttle events)
 * - ConcurrentExecutions metric
 * - Dashboard widgets (returned for aggregation)
 */
public class MonitoredFunction extends Construct {

    private final OptimizedJavaFunction function;
    private final List<IWidget> dashboardWidgets;

    public MonitoredFunction(Construct scope, String id, MonitoredFunctionProps props) {
        super(scope, id);

        this.function = new OptimizedJavaFunction(this, "Function", props.getFunctionProps());

        // Error rate alarm
        Alarm.Builder.create(this, "ErrorAlarm")
            .metric(this.function.getFunction().metricErrors()
                .with(MetricOptions.builder().period(Duration.minutes(5)).build()))
            .threshold(props.getErrorRateThreshold() != null ? props.getErrorRateThreshold() : 1)
            .evaluationPeriods(1)
            .build();

        // Duration P99 alarm (80% of timeout to catch slow functions before they time out)
        var timeoutSeconds = props.getFunctionProps().getTimeout() != null
            ? props.getFunctionProps().getTimeout().toSeconds() : 30;
        Alarm.Builder.create(this, "DurationAlarm")
            .metric(this.function.getFunction().metricDuration()
                .with(MetricOptions.builder()
                    .statistic("p99")
                    .period(Duration.minutes(5))
                    .build()))
            .threshold(timeoutSeconds * 0.8 * 1000) // milliseconds
            .evaluationPeriods(2)
            .build();

        // Build dashboard widgets
        this.dashboardWidgets = List.of(
            GraphWidget.Builder.create()
                .title(props.getFunctionProps().getFunctionName() + " Invocations")
                .left(List.of(this.function.getFunction().metricInvocations()))
                .right(List.of(this.function.getFunction().metricErrors()))
                .build(),
            GraphWidget.Builder.create()
                .title(props.getFunctionProps().getFunctionName() + " Duration")
                .left(List.of(this.function.getFunction().metricDuration()))
                .build()
        );
    }

    public OptimizedJavaFunction getFunction() { return this.function; }
    public List<IWidget> getDashboardWidgets() { return this.dashboardWidgets; }
}
```

---

## 9. CLI UX

### 9.1 New `--target` Flag

The primary change is a new `--target` flag on `trabuco init`:

```
trabuco init --target=cdk
```

When `--target=cdk`, the prompt flow changes entirely. When `--target=springboot` (or omitted), the existing flow is unchanged.

### 9.2 Updated Flag Set for `trabuco init`

```go
// Existing flags (unchanged)
initCmd.Flags().StringVar(&flagProjectName, "name", "", "Project name")
initCmd.Flags().StringVar(&flagGroupID, "group-id", "", "Group ID")
initCmd.Flags().StringVar(&flagJavaVersion, "java-version", "21", "Java version: 17, 21, or 25")
initCmd.Flags().StringVar(&flagAIAgents, "ai-agents", "", "AI agents: claude,cursor,copilot,codex")
initCmd.Flags().StringVar(&flagCI, "ci", "", "CI provider: github")
initCmd.Flags().BoolVar(&flagStrict, "strict", false, "Fail if Java not detected")
initCmd.Flags().BoolVar(&flagSkipBuild, "skip-build", false, "Skip build after generation")

// New flags for CDK
initCmd.Flags().StringVar(&flagTarget, "target", "springboot", "Deployment target: springboot, cdk")
initCmd.Flags().StringVar(&flagAWSRegion, "aws-region", "us-east-1", "AWS region (CDK only)")
initCmd.Flags().StringVar(&flagAWSAccount, "aws-account", "", "AWS account ID (CDK Pipeline only)")
initCmd.Flags().StringVar(&flagCDKModules, "cdk-modules", "", "CDK modules (comma-separated)")
initCmd.Flags().StringVar(&flagLambdaFramework, "lambda-framework", "none", "Lambda framework: none, quarkus, micronaut")
initCmd.Flags().IntVar(&flagLambdaMemory, "lambda-memory", 512, "Lambda memory in MB")
initCmd.Flags().BoolVar(&flagEnableSnapStart, "snapstart", true, "Enable SnapStart (Java 21 only)")
initCmd.Flags().BoolVar(&flagEnableDynamoDB, "dynamodb", true, "Enable DynamoDB (CDK only)")
initCmd.Flags().BoolVar(&flagEnableRDS, "rds", false, "Enable RDS with Proxy (CDK only, requires VPC)")
initCmd.Flags().BoolVar(&flagEnableS3, "s3", false, "Enable S3 bucket (CDK only)")
initCmd.Flags().BoolVar(&flagEnableCognito, "cognito", false, "Enable Cognito auth (CDK only)")
initCmd.Flags().BoolVar(&flagEnableSQS, "sqs", false, "Enable SQS queues (CDK only)")
initCmd.Flags().BoolVar(&flagEnableEventBridge, "eventbridge", false, "Enable EventBridge (CDK only)")
initCmd.Flags().BoolVar(&flagEnableStepFunctions, "stepfunctions", false, "Enable Step Functions (CDK only)")

// Existing Spring Boot flags remain but are ignored when --target=cdk
// (flagModules, flagDatabase, flagNoSQLDatabase, flagMessageBroker)
```

### 9.3 Interactive Prompt Flow for CDK

When `--target=cdk` is specified without all required flags:

```
╔════════════════════════════════════════════╗
║   Trabuco - AWS CDK Serverless Generator   ║
╚════════════════════════════════════════════╝

? Project name: my-platform
? Group ID: [com.company.myplatform]
? Java version: 21 (LTS) [detected]

? AWS Region: [us-east-1]

? Select CDK stacks to include:
  [x] CDK-Infra - CDK app, stages, environment config (required)
  [x] CDK-StatefulStack - DynamoDB, S3, Cognito, RDS
  [x] CDK-ApiStack - API Gateway + Lambda REST API
  [ ] CDK-AsyncStack - SQS, EventBridge, Step Functions
  [x] CDK-MonitoringStack - CloudWatch dashboards and alarms
  [ ] CDK-Pipeline - CDK Pipelines CI/CD

? Stateful resources to provision:
  [x] DynamoDB (Recommended - serverless NoSQL)
  [ ] S3 (Object storage)
  [ ] Cognito (User authentication)
  [ ] RDS + RDS Proxy (Relational database - requires VPC)

? (Only if CDK-AsyncStack selected) Async services:
  [x] SQS (Message queues with DLQ)
  [ ] EventBridge (Event routing)
  [ ] Step Functions (Workflow orchestration)

? Lambda framework:
  > None (Plain AWS SDK - fastest cold start)
    Quarkus (CDI + REST, GraalVM-ready)
    Micronaut (Compile-time DI, GraalVM-ready)

? Lambda memory (MB): [512]

? (Only if Java 21) Enable SnapStart? [Y/n]

? Generate AI agent context files:
  [x] Claude Code
  [ ] Cursor
  [ ] Copilot

? CI/CD:
  > None
    GitHub Actions (build + cdk synth)

─────────────────────────────────────────
  Project Summary
─────────────────────────────────────────
  Target:       AWS CDK (Serverless)
  Project:      my-platform
  Group ID:     com.company.myplatform
  Java:         21 (SnapStart enabled)
  Region:       us-east-1
  CDK Stacks:   CDK-Infra, CDK-StatefulStack, CDK-ApiStack, CDK-MonitoringStack
  Stateful:     DynamoDB
  Lambda:       Plain AWS SDK (512 MB)
  AI Agents:    Claude Code
─────────────────────────────────────────
```

### 9.4 Non-Interactive CDK Example

```bash
trabuco init \
  --target=cdk \
  --name=my-platform \
  --group-id=com.company.myplatform \
  --java-version=21 \
  --aws-region=us-east-1 \
  --cdk-modules=CDK-Infra,CDK-StatefulStack,CDK-ApiStack,CDK-MonitoringStack \
  --dynamodb \
  --lambda-framework=none \
  --lambda-memory=512 \
  --snapstart \
  --ai-agents=claude \
  --ci=github
```

### 9.5 Post-Generation Build

Instead of `mvn clean install`, CDK projects run:

```go
func runGradleBuild(projectDir string) error {
    cmd := exec.Command("./gradlew", "build", "-x", "test", "--console=plain")
    cmd.Dir = projectDir
    output, err := cmd.CombinedOutput()
    // ...
}
```

And optionally:

```go
func runCDKSynth(projectDir string) error {
    cmd := exec.Command("cdk", "synth", "--quiet")
    cmd.Dir = projectDir
    output, err := cmd.CombinedOutput()
    // ...
}
```

### 9.6 Docker Requirement Relaxation

CDK projects do **not** require Docker to be running at generation time (unlike Spring Boot projects which need Docker for Testcontainers). The Docker check in `runInit` is conditioned on `target == springboot`.

### 9.7 Prerequisite Checks for CDK

New checks added to `trabuco doctor` and the init flow:

```go
// Check Node.js (required for CDK CLI)
func checkNodeJS() DoctorCheck { ... }

// Check CDK CLI
func checkCDKCLI() DoctorCheck { ... }

// Check AWS CLI and credentials
func checkAWSCLI() DoctorCheck { ... }
func checkAWSCredentials() DoctorCheck { ... }

// Check SAM CLI (optional, for local dev)
func checkSAMCLI() DoctorCheck { ... }
```

---

## 10. Go Implementation Details

### 10.1 Generator Branching

The main `Generate()` method in `generator.go` branches based on deployment target:

```go
func (g *Generator) Generate() error {
    if g.config.IsCDK() {
        return g.generateCDK()
    }
    return g.generateSpringBoot() // existing logic, renamed
}
```

### 10.2 New Generator File: `generator/cdk.go`

```go
package generator

func (g *Generator) generateCDK() error {
    // 1. Create directory structure
    if err := g.createCDKDirectories(); err != nil { ... }

    // 2. Generate Gradle build files
    if err := g.generateGradleBuild(); err != nil { ... }

    // 3. Generate Model module (reuses existing templates)
    if err := g.generateModelModule(); err != nil { ... }

    // 4. Generate Shared module (reuses existing templates)
    if g.config.HasModule(config.ModuleShared) {
        if err := g.generateSharedModule(); err != nil { ... }
    }

    // 5. Generate Lambda function modules
    if g.config.HasCDKModule(config.ModuleCDKApi) {
        if err := g.generateLambdaApiModule(); err != nil { ... }
    }
    if g.config.HasCDKModule(config.ModuleCDKAsync) {
        if err := g.generateLambdaWorkerModule(); err != nil { ... }
        if err := g.generateLambdaEventsModule(); err != nil { ... }
    }

    // 6. Generate CDK infrastructure stacks
    if err := g.generateCDKInfra(); err != nil { ... }
    for _, cdkModule := range g.config.CDKModules {
        if err := g.generateCDKModule(cdkModule); err != nil { ... }
    }

    // 7. Generate L3 constructs
    if err := g.generateL3Constructs(); err != nil { ... }

    // 8. Generate documentation
    if err := g.generateCDKDocs(); err != nil { ... }

    // 9. Generate SAM template for local dev
    if err := g.generateSAMTemplate(); err != nil { ... }

    // 10. Initialize git
    if err := g.initGit(); err != nil { ... }

    // 11. Generate metadata
    if err := g.generateMetadata(g.version); err != nil { ... }

    return nil
}
```

### 10.3 Template Engine Changes

No changes to `templates/templates.go`. The existing `Engine` handles CDK templates identically -- they are just `.tmpl` files in new subdirectories under `templates/`.

### 10.4 Embed Changes

The `templates/embed.go` file already embeds all files recursively:

```go
//go:embed all:*
var FS embed.FS
```

New template directories (`cdk/`, `gradle/`, `sam/`, `java/lambdaapi/`, etc.) are picked up automatically.

### 10.5 Module Resolution for CDK

The existing `ResolveDependencies` function is extended:

```go
func ResolveCDKDependencies(selectedCDKModules []string, config *ProjectConfig) ([]string, []string) {
    // Returns (resolvedCDKModules, requiredLambdaModules)
    cdkSet := make(map[string]bool)
    for _, m := range selectedCDKModules { cdkSet[m] = true }

    // Always include CDK-Infra
    cdkSet[ModuleCDKInfra] = true

    // CDK-ApiStack requires LambdaApi
    // CDK-AsyncStack requires LambdaWorker and LambdaEvents
    // CDK-StatefulStack is implied by CDK-ApiStack or CDK-AsyncStack

    lambdaModules := []string{}
    if cdkSet[ModuleCDKApi] {
        cdkSet[ModuleCDKStateful] = true
        lambdaModules = append(lambdaModules, ModuleLambdaApi)
    }
    if cdkSet[ModuleCDKAsync] {
        cdkSet[ModuleCDKStateful] = true
        lambdaModules = append(lambdaModules, ModuleLambdaWorker, ModuleLambdaEvents)
    }

    // Return in registry order
    // ...
}
```

---

## 11. Testing Strategy

### 11.1 Template Compilation Tests

Extend the existing `gen_test.go` / `compilation_test.go` pattern:

```go
func TestCDKGeneration_ApiStack(t *testing.T) {
    cfg := &config.ProjectConfig{
        ProjectName:      "test-cdk",
        GroupID:           "com.test.cdk",
        ArtifactID:       "test-cdk",
        JavaVersion:      "21",
        DeploymentTarget: config.TargetCDK,
        AWSRegion:        "us-east-1",
        CDKModules:       []string{"CDK-Infra", "CDK-StatefulStack", "CDK-ApiStack"},
        Modules:          []string{"Model", "Shared"},
        EnableDynamoDB:   true,
        EnableSnapStart:  true,
        LambdaFramework:  "none",
        LambdaMemoryMB:   512,
    }

    gen, err := generator.NewWithVersionAt(cfg, "test", t.TempDir())
    require.NoError(t, err)
    require.NoError(t, gen.Generate())

    // Verify Gradle build compiles
    cmd := exec.Command("./gradlew", "build", "-x", "test")
    cmd.Dir = filepath.Join(t.TempDir())
    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "Gradle build failed: %s", string(output))

    // Verify CDK synth produces valid CloudFormation
    cmd = exec.Command("cdk", "synth", "--quiet")
    cmd.Dir = filepath.Join(t.TempDir())
    output, err = cmd.CombinedOutput()
    require.NoError(t, err, "CDK synth failed: %s", string(output))
}
```

### 11.2 Test Matrix

Generate and compile-test every meaningful CDK module combination:

| Test Case | CDK Modules | Lambda Framework | Datastores |
|-----------|-------------|-----------------|------------|
| Minimal API | Infra, Stateful, Api | none | DynamoDB |
| Full serverless | Infra, Stateful, Api, Async, Monitoring | none | DynamoDB, S3 |
| With Pipeline | Infra, Stateful, Api, Pipeline | none | DynamoDB |
| Quarkus API | Infra, Stateful, Api | quarkus | DynamoDB |
| Micronaut API | Infra, Stateful, Api | micronaut | DynamoDB |
| RDS + VPC | Infra, Stateful, Api | none | DynamoDB, RDS |
| Cognito auth | Infra, Stateful, Api | none | DynamoDB, Cognito |
| Async only | Infra, Stateful, Async | none | DynamoDB, SQS |
| EventBridge | Infra, Stateful, Async | none | DynamoDB, EventBridge |
| Step Functions | Infra, Stateful, Async | none | DynamoDB, StepFunctions |
| Kitchen sink | All stacks | none | All datastores |

---

## Appendix A: Generated Project Structure (Full Example)

For `trabuco init --target=cdk --name=my-platform --cdk-modules=CDK-Infra,CDK-StatefulStack,CDK-ApiStack,CDK-AsyncStack,CDK-MonitoringStack`:

```
my-platform/
├── .ai/
│   ├── checkpoint.json
│   ├── review-log.jsonl
│   ├── README.md
│   └── prompts/
│       ├── JAVA_CODE_QUALITY.md
│       ├── code-review.md
│       ├── testing-guide.md
│       ├── add-lambda.md
│       ├── add-stack.md
│       ├── add-construct.md
│       ├── add-dynamodb-table.md
│       ├── cdk-patterns.md
│       └── lambda-optimization.md
├── .claude/
│   ├── rules/                  (mirrors .ai/prompts/)
│   ├── settings.json
│   └── skills/
│       ├── commit/SKILL.md
│       ├── pr/SKILL.md
│       └── review/SKILL.md
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── deploy.yml
├── model/
│   ├── build.gradle.kts
│   └── src/main/java/com/company/myplatform/model/
│       ├── ImmutableStyle.java
│       ├── entities/
│       │   └── Placeholder.java
│       └── dto/
│           ├── PlaceholderRequest.java
│           └── PlaceholderResponse.java
├── shared/
│   ├── build.gradle.kts
│   └── src/
│       ├── main/java/com/company/myplatform/shared/
│       │   ├── config/
│       │   │   └── SharedConfig.java
│       │   └── service/
│       │       └── PlaceholderService.java
│       └── test/java/com/company/myplatform/shared/
│           └── service/
│               └── PlaceholderServiceTest.java
├── lambda-api/
│   ├── build.gradle.kts
│   └── src/
│       ├── main/java/com/company/myplatform/lambdaapi/
│       │   ├── ServiceFactory.java
│       │   ├── handler/
│       │   │   ├── GetPlaceholderHandler.java
│       │   │   ├── CreatePlaceholderHandler.java
│       │   │   └── ListPlaceholdersHandler.java
│       │   └── util/
│       │       ├── ApiGatewayResponse.java
│       │       └── JsonUtil.java
│       └── test/java/com/company/myplatform/lambdaapi/
│           └── handler/
│               ├── GetPlaceholderHandlerTest.java
│               └── CreatePlaceholderHandlerTest.java
├── lambda-worker/
│   ├── build.gradle.kts
│   └── src/
│       ├── main/java/com/company/myplatform/lambdaworker/
│       │   ├── ServiceFactory.java
│       │   └── handler/
│       │       ├── ProcessPlaceholderHandler.java
│       │       └── ScheduledTaskHandler.java
│       └── test/java/com/company/myplatform/lambdaworker/
│           └── handler/
│               └── ProcessPlaceholderHandlerTest.java
├── lambda-events/
│   ├── build.gradle.kts
│   └── src/
│       ├── main/java/com/company/myplatform/lambdaevents/
│       │   ├── ServiceFactory.java
│       │   └── handler/
│       │       └── PlaceholderEventHandler.java
│       └── test/java/com/company/myplatform/lambdaevents/
│           └── handler/
│               └── PlaceholderEventHandlerTest.java
├── infra/
│   ├── build.gradle.kts
│   └── src/
│       ├── main/java/com/company/myplatform/infra/
│       │   ├── MyPlatformApp.java
│       │   ├── MyPlatformStage.java
│       │   ├── EnvironmentConfig.java
│       │   ├── Tags.java
│       │   ├── stacks/
│       │   │   ├── StatefulStack.java
│       │   │   ├── ApiStack.java
│       │   │   ├── AsyncStack.java
│       │   │   └── MonitoringStack.java
│       │   └── constructs/
│       │       ├── OptimizedJavaFunction.java
│       │       ├── EventDrivenWorker.java
│       │       ├── SecuredApi.java
│       │       ├── MonitoredFunction.java
│       │       ├── DynamoDbTables.java
│       │       ├── ApiLambdas.java
│       │       ├── SqsQueues.java
│       │       ├── EventBridgeRules.java
│       │       ├── LambdaDashboard.java
│       │       └── ApiAlarms.java
│       └── test/java/com/company/myplatform/infra/
│           ├── MyPlatformStageTest.java
│           └── stacks/
│               ├── StatefulStackTest.java
│               ├── ApiStackTest.java
│               ├── AsyncStackTest.java
│               └── MonitoringStackTest.java
├── gradle/
│   ├── wrapper/
│   │   └── gradle-wrapper.properties
│   └── libs.versions.toml
├── build.gradle.kts
├── settings.gradle.kts
├── gradle.properties
├── cdk.json
├── template.yml                 (SAM local dev)
├── samconfig.toml
├── CLAUDE.md
├── AGENTS.md
├── ARCHITECTURE.md
├── README.md
├── .gitignore
├── .trabuco.json
└── .env.example
```

## Appendix B: Migration Path

Existing Spring Boot projects generated by Trabuco cannot be automatically converted to CDK projects. The `trabuco migrate` command is not extended for CDK. Instead, users start fresh with `trabuco init --target=cdk` and manually move their domain objects (Model) and business logic (Shared) into the new project. The Model and Shared module code is structurally identical between Spring Boot and CDK projects by design.
