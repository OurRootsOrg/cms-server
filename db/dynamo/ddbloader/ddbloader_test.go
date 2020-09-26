package main

import (
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEnv(t *testing.T) {
	const fileURL = "file:///../testdata/place_settings.tsv"
	// All defaults
	os.Setenv("MIN_LOG_LEVEL", "")
	os.Setenv("DYNAMODB_TABLE_NAME", "test-table")
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("FILE_URL", fileURL)
	os.Unsetenv("FILE_PATH")
	env, err := ParseEnv()
	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, "DEBUG", env.MinLogLevel) // default
	assert.Equal(t, "test-table", env.DynamoDBTableName)
	assert.Equal(t, "us-east-1", env.Region)
	assert.Equal(t, false, env.LocalTest)
	_, err = url.Parse(fileURL)
	assert.NoError(t, err)

	// Bad MIN_LOG_LEVEL
	os.Setenv("MIN_LOG_LEVEL", "WARN")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("MIN_LOG_LEVEL", "")

	// Bad FILE_URL
	os.Setenv("FILE_URL", "bad")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("FILE_URL", fileURL)

	// Both FILE_PATH and FILE_URL set
	os.Setenv("FILE_PATH", "/tmp/file.tsv")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)

	// Only FILE_PATH set
	os.Setenv("FILE_PATH", "/tmp/file.tsv")
	os.Unsetenv("FILE_URL")
	env, err = ParseEnv()
	assert.NoError(t, err)
	assert.NotNil(t, env)

	// Neither FILE_URL or FILE_PATH set
	os.Unsetenv("FILE_PATH")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("FILE_URL", fileURL)

	// Missing DYNAMODB_TABLE_NAME
	os.Unsetenv("DYNAMODB_TABLE_NAME")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("DYNAMODB_TABLE_NAME", "test-table")

	// Missing AWS_REGION
	os.Unsetenv("AWS_REGION")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("AWS_REGION", "us-east-1")

	// Bad LOCAL_TEST
	os.Setenv("LOCAL_TEST", "bad")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)

	// True LOCAL_TEST
	os.Setenv("LOCAL_TEST", "true")
	env, err = ParseEnv()
	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, true, env.LocalTest)

	// All bad
	os.Setenv("MIN_LOG_LEVEL", "WARN")
	os.Setenv("FILE_URL", "bad")
	os.Setenv("FILE_PATH", "bad")
	os.Unsetenv("DYNAMODB_TABLE_NAME")
	os.Unsetenv("AWS_REGION")
	os.Setenv("LOCAL_TEST", "bad")
	env, err = ParseEnv()
	assert.Error(t, err)
	assert.Nil(t, env)
}
