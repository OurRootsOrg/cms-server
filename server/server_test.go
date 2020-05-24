package main

import (
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEnv(t *testing.T) {
	// All defaults
	os.Unsetenv("LAMBDA_TASK_ROOT")
	os.Setenv("BASE_URL", "")
	os.Setenv("MIN_LOG_LEVEL", "")
	os.Setenv("DATABASE_URL", "postgres://ourroots:password@localhost:5432/ourroots?sslmode=disable")
	env, err := ParseEnv()
	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, "http://localhost:3000", env.BaseURLString)
	u, err := url.ParseRequestURI("http://localhost:3000")
	assert.NoError(t, err)
	assert.Equal(t, false, env.IsLambda)
	assert.Equal(t, u, env.BaseURL)
	assert.Equal(t, "DEBUG", env.MinLogLevel) // default
	assert.Equal(t, "postgres://ourroots:password@localhost:5432/ourroots?sslmode=disable", env.DatabaseURL)

	// Test Lambda
	os.Setenv("LAMBDA_TASK_ROOT", "/tmp")
	env, err = ParseEnv()
	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, true, env.IsLambda)
	os.Unsetenv("LAMBDA_TASK_ROOT")

	// Bad MIN_LOG_LEVEL
	os.Setenv("MIN_LOG_LEVEL", "WARN")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("MIN_LOG_LEVEL", "")

	// Bad BASE_URL
	os.Setenv("BASE_URL", "bad")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("BASE_URL", "")

	// Missing DATABASE_URL
	os.Unsetenv("DATABASE_URL")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)

	// Bad DATABASE_URL
	os.Setenv("DATABASE_URL", "baddb")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("DATABASE_URL", "")

	// All bad
	os.Setenv("BASE_URL", "bad")
	os.Setenv("MIN_LOG_LEVEL", "WARN")
	os.Setenv("DATABASE_URL", "baddb")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
}
