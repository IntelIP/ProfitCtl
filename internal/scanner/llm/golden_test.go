package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoldenFilePromptBuilding(t *testing.T) {
	// Load the golden fixture
	fixturePath := "../../../test/fixtures/samples/go-webapp"
	goldenPath := "../../../test/fixtures/golden/go-webapp-expected.yml"

	// Verify fixture files exist
	require.FileExists(t, filepath.Join(fixturePath, "go.mod"))
	require.FileExists(t, goldenPath)

	// Read the go.mod file
	goModContent, err := os.ReadFile(filepath.Join(fixturePath, "go.mod"))
	require.NoError(t, err)

	// Create CodeContext with the fixture
	ctx := CodeContext{
		Files: map[string]string{
			"go.mod": string(goModContent),
		},
	}

	// Generate the prompt
	prompt := BuildDetectionPrompt(ctx)

	// Validate prompt structure
	assert.Contains(t, prompt, "# Codebase Analysis")
	assert.Contains(t, prompt, "## go.mod")
	assert.Contains(t, prompt, "```")
	assert.Contains(t, prompt, "module github.com/example/go-webapp")
	assert.Contains(t, prompt, "## Analysis Instructions")
	assert.Contains(t, prompt, `"services": [`)
	assert.Contains(t, prompt, `"dependencies": [`)
	assert.Contains(t, prompt, `"patterns": [`)

	// Validate truncation is applied (go.mod content should be present)
	assert.Contains(t, prompt, "github.com/gin-gonic/gin v1.9.1")
	assert.Contains(t, prompt, "gorm.io/gorm v1.25.5")
	assert.Contains(t, prompt, "github.com/redis/go-redis/v9 v9.3.0")

	// Ensure prompt is not excessively long
	assert.Less(t, len(prompt), 10000, "Prompt should be reasonable length")
}

func BenchmarkBuildDetectionPrompt(b *testing.B) {
	// Create a realistic CodeContext
	ctx := CodeContext{
		Files: map[string]string{
			"go.mod": `module github.com/example/app

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
	github.com/redis/go-redis/v9 v9.3.0
	github.com/aws/aws-sdk-go v1.48.0
	github.com/stripe/stripe-go/v76 v76.15.0
)`,
			"package.json": `{
  "name": "example-app",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.18.2",
    "redis": "^4.6.8",
    "stripe": "^12.0.0"
  }
}`,
			"Dockerfile": `FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]`,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildDetectionPrompt(ctx)
	}
}

func BenchmarkBuildDetectionPromptLarge(b *testing.B) {
	// Create a large CodeContext with many files
	ctx := CodeContext{
		Files: make(map[string]string),
	}

	// Add multiple large files
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("file%d.go", i)
		content := fmt.Sprintf(`package main

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/redis/go-redis/v9"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stripe/stripe-go/v76"
)

// Large file %d with lots of content
func main() {
	r := gin.Default()
	rdb := redis.NewClient(&redis.Options{})
	db, _ := gorm.Open(postgres.Open(""))
	awsSvc := aws.NewService()
	stripe.Key = "sk_test_..."

	r.GET("/api/v1/users", func(c *gin.Context) {
		// Lots of code here
		c.JSON(200, gin.H{
			"message": "Hello World",
			"data": []string{"item1", "item2", "item3"},
		})
	})

	r.Run(":8080")
}`, i)
		ctx.Files[filename] = content
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildDetectionPrompt(ctx)
	}
}
