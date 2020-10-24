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
	os.Setenv("FILE_URLS", fileURL)
	os.Setenv("LOAD_THROUGHPUT", "500")
	os.Setenv("NORMAL_THROUGHPUT", "5")
	os.Unsetenv("FILE_PATHS")
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

	// Both FILE_PATHS and FILE_URLS set
	os.Setenv("FILE_URLS", fileURL)
	os.Setenv("FILE_PATHS", "/tmp/file.tsv")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)

	// Only FILE_PATHS set
	os.Setenv("FILE_PATHS", "/tmp/file.tsv")
	os.Unsetenv("FILE_URLS")
	env, err = ParseEnv()
	assert.NoError(t, err)
	assert.NotNil(t, env)

	// Neither FILE_URLS or FILE_PATHS set
	os.Unsetenv("FILE_PATHS")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("FILE_URLS", fileURL)

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

	// Bad NORMAL_THROUGHPUT
	os.Setenv("NORMAL_THROUGHPUT", "bad")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)

	// Bad NORMAL_THROUGHPUT
	os.Setenv("NORMAL_THROUGHPUT", "-1")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)

	// Unset NORMAL_THROUGHPUT
	os.Unsetenv("NORMAL_THROUGHPUT")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("NORMAL_THROUGHPUT", "5")

	// Bad LOAD_THROUGHPUT
	os.Setenv("LOAD_THROUGHPUT", "bad")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)

	// Bad LOAD_THROUGHPUT
	os.Setenv("LOAD_THROUGHPUT", "0")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)

	// Unset LOAD_THROUGHPUT
	os.Unsetenv("LOAD_THROUGHPUT")
	env, err = ParseEnv()
	assert.Error(t, err)
	log.Printf("Error: %v", err)
	assert.Nil(t, env)
	os.Setenv("LOAD_THROUGHPUT", "500")

	// All bad
	os.Setenv("MIN_LOG_LEVEL", "WARN")
	os.Setenv("FILE_URLS", "bad")
	os.Setenv("FILE_PATHS", "bad")
	os.Unsetenv("DYNAMODB_TABLE_NAME")
	os.Unsetenv("AWS_REGION")
	os.Setenv("LOCAL_TEST", "bad")
	os.Setenv("NORMAL_THROUGHPUT", "0")
	os.Setenv("LOAD_THROUGHPUT", "0")
	env, err = ParseEnv()
	assert.Error(t, err)
	assert.Nil(t, env)
}
