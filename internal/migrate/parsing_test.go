package migrate

import (
	"testing"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "JSON in markdown code block",
			content: `Here is the result:
` + "```json" + `
{
  "name": "User",
  "code": "public class User {}"
}
` + "```" + `
That's the converted entity.`,
			want: `{
  "name": "User",
  "code": "public class User {}"
}`,
		},
		{
			name: "Raw JSON without markdown",
			content: `{
  "name": "User",
  "code": "public class User {}"
}`,
			want: `{
  "name": "User",
  "code": "public class User {}"
}`,
		},
		{
			name:    "JSON with text before",
			content: `The converted code is: {"name": "User", "code": "class User {}"}`,
			want:    `{"name": "User", "code": "class User {}"}`,
		},
		{
			name: "Nested JSON objects",
			content: `{"entity": {"name": "User"}, "dto": {"name": "UserDTO"}}`,
			want:    `{"entity": {"name": "User"}, "dto": {"name": "UserDTO"}}`,
		},
		{
			name:    "Empty content",
			content: "",
			want:    "",
		},
		{
			name:    "No JSON in content",
			content: "This is just plain text without any JSON.",
			want:    "This is just plain text without any JSON.",
		},
		{
			name: "JSON in plain markdown code block (no language)",
			content: "Here is the result:\n```\n" + `{
  "name": "User",
  "code": "public class User {}"
}` + "\n```\nDone.",
			want: `{
  "name": "User",
  "code": "public class User {}"
}`,
		},
		{
			name: "JSON with curly braces in strings",
			content: `{"name": "Test", "code": "String s = \"hello {world}\";"}`,
			want:    `{"name": "Test", "code": "String s = \"hello {world}\";"}`,
		},
		{
			name: "JSON with newline after json marker",
			content: "```json\n{\n  \"skip\": false,\n  \"entity\": {\"name\": \"Test\"}\n}\n```",
			want:    "{\n  \"skip\": false,\n  \"entity\": {\"name\": \"Test\"}\n}",
		},
		{
			name: "JSON without closing backticks (truncated response)",
			content: "```json\n{\n  \"skip\": false,\n  \"entity\": {\"name\": \"Test\"}\n}",
			want:    "{\n  \"skip\": false,\n  \"entity\": {\"name\": \"Test\"}\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.content)
			if got != tt.want {
				t.Errorf("extractJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseEntityMigrationResult(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*EntityMigrationResult) bool
	}{
		{
			name: "Valid entity result",
			content: `{
				"entity": {"name": "User", "code": "public class User {}"},
				"dto": {"name": "UserRequest", "code": "public class UserRequest {}"},
				"response": {"name": "UserResponse", "code": "public class UserResponse {}"},
				"flyway_migration": "CREATE TABLE users (...)",
				"notes": ["Converted from JPA"],
				"requires_review": false,
				"review_reason": ""
			}`,
			wantErr: false,
			check: func(r *EntityMigrationResult) bool {
				return r.Entity.Name == "User" &&
					r.DTO.Name == "UserRequest" &&
					r.Response.Name == "UserResponse" &&
					len(r.Notes) == 1
			},
		},
		{
			name: "With markdown wrapper",
			content: "```json\n" + `{
				"entity": {"name": "Order", "code": "public class Order {}"},
				"dto": {"name": "OrderRequest", "code": ""},
				"response": {"name": "OrderResponse", "code": ""},
				"flyway_migration": "",
				"notes": [],
				"requires_review": true,
				"review_reason": "Complex relationships"
			}` + "\n```",
			wantErr: false,
			check: func(r *EntityMigrationResult) bool {
				return r.Entity.Name == "Order" && r.RequiresReview
			},
		},
		{
			name:    "Invalid JSON",
			content: `{"entity": {"name": "User"`,
			wantErr: true,
		},
		{
			name:    "Empty content",
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEntityMigrationResult(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseEntityMigrationResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("parseEntityMigrationResult() result validation failed")
			}
		})
	}
}

func TestParseRepositoryResult(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*RepositoryMigrationResult) bool
	}{
		{
			name: "Valid repository result",
			content: `{
				"name": "UserRepository",
				"code": "public interface UserRepository extends CrudRepository<User, Long> {}",
				"notes": ["Converted custom queries"],
				"requires_review": false,
				"review_reason": ""
			}`,
			wantErr: false,
			check: func(r *RepositoryMigrationResult) bool {
				return r.Name == "UserRepository" && len(r.Notes) == 1
			},
		},
		{
			name:    "Invalid JSON",
			content: `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRepositoryResult(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseRepositoryResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("parseRepositoryResult() result validation failed")
			}
		})
	}
}

func TestParseServiceResult(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*ServiceMigrationResult) bool
	}{
		{
			name: "Valid service result",
			content: `{
				"name": "UserService",
				"code": "public class UserService {}",
				"test_code": "public class UserServiceTest {}",
				"notes": [],
				"requires_review": false,
				"review_reason": ""
			}`,
			wantErr: false,
			check: func(r *ServiceMigrationResult) bool {
				return r.Name == "UserService" && r.TestCode != ""
			},
		},
		{
			name: "Service without test",
			content: `{
				"name": "SimpleService",
				"code": "public class SimpleService {}",
				"test_code": "",
				"notes": ["No test generated"],
				"requires_review": true,
				"review_reason": "Complex dependencies"
			}`,
			wantErr: false,
			check: func(r *ServiceMigrationResult) bool {
				return r.Name == "SimpleService" && r.TestCode == "" && r.RequiresReview
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseServiceResult(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseServiceResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("parseServiceResult() result validation failed")
			}
		})
	}
}

func TestParseControllerResult(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*ControllerMigrationResult) bool
	}{
		{
			name: "Valid controller result",
			content: `{
				"name": "UserController",
				"code": "@RestController public class UserController {}",
				"notes": ["Added validation"],
				"requires_review": false,
				"review_reason": ""
			}`,
			wantErr: false,
			check: func(r *ControllerMigrationResult) bool {
				return r.Name == "UserController"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseControllerResult(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseControllerResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("parseControllerResult() result validation failed")
			}
		})
	}
}

func TestParseJobResult(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*JobMigrationResult) bool
	}{
		{
			name: "Valid job result",
			content: `{
				"name": "DataCleanup",
				"job_request_code": "public class DataCleanupRequest implements JobRequest {}",
				"job_handler_code": "public class DataCleanupHandler implements JobRequestHandler<DataCleanupRequest> {}",
				"notes": ["Converted from @Scheduled"],
				"requires_review": false,
				"review_reason": ""
			}`,
			wantErr: false,
			check: func(r *JobMigrationResult) bool {
				return r.Name == "DataCleanup" && r.JobRequestCode != "" && r.JobHandlerCode != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJobResult(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseJobResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("parseJobResult() result validation failed")
			}
		})
	}
}

func TestParseEventResult(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*EventMigrationResult) bool
	}{
		{
			name: "Valid event result",
			content: `{
				"name": "UserCreated",
				"event_code": "public record UserCreatedEvent(String userId) {}",
				"listener_code": "@KafkaListener public class UserCreatedListener {}",
				"notes": [],
				"requires_review": false,
				"review_reason": ""
			}`,
			wantErr: false,
			check: func(r *EventMigrationResult) bool {
				return r.Name == "UserCreated" && r.EventCode != "" && r.ListenerCode != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEventResult(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseEventResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("parseEventResult() result validation failed")
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "Short string unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "Long string truncated",
			input:  "hello world this is a long string",
			maxLen: 10,
			want:   "hello worl...",
		},
		{
			name:   "Exact length unchanged",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "Empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate() = %q, want %q", got, tt.want)
			}
		})
	}
}
