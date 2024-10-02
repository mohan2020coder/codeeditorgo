package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CodeRequest struct {
	Code string `json:"code"`
}

type ExecutionResult struct {
	Code   string `json:"code"`
	Output string `json:"output"`
	Time   string `json:"time"`
}

var collection *mongo.Collection

func main() {
	// Update to use mongo.Connect instead of mongo.NewClient
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println("MongoDB client error:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check the connection
	if err = client.Ping(ctx, nil); err != nil {
		fmt.Println("MongoDB connection error:", err)
		return
	}

	collection = client.Database("code_editor").Collection("executions")

	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		// or specify specific origins with:
		// AllowOrigins: []string{"http://localhost:3000"}, // Replace with your frontend origin
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	// Route for executing Go code
	r.POST("/execute", func(c *gin.Context) {
		var codeReq CodeRequest
		if err := c.BindJSON(&codeReq); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		output, err := runCodeInDocker(codeReq.Code)
		if err != nil {
			c.JSON(500, gin.H{"output": err.Error()})
			return
		}

		execution := ExecutionResult{
			Code:   codeReq.Code,
			Output: output,
			Time:   time.Now().Format(time.RFC3339),
		}

		_, err = collection.InsertOne(context.Background(), execution)
		if err != nil {
			fmt.Println("MongoDB insertion error:", err)
		}

		c.JSON(200, gin.H{"output": output})
	})

	r.Run(":8080")
}

// runCodeInDocker runs the provided code inside a Docker container
// runCodeInDocker runs the provided code inside a Docker container
func runCodeInDocker(code string) (string, error) {
	// Save the code to a temporary file in the user's home directory
	tmpFile, err := os.CreateTemp(os.Getenv("HOME"), "code-*.go") // Change the path here if needed
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(code)
	if err != nil {
		return "", err
	}

	// Command to run the Docker container with the Go code
	cmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"-v", tmpFile.Name()+":/app/code.go",
		"--memory", "256m", // Increase memory limit
		"--cpus", "1", // You can also increase CPU limit if needed
		"go-sandbox",
		"go", "run", "/app/code.go",
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("execution failed: %s", out.String())
	}

	return out.String(), nil
}
