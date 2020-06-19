package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/rdsdataservice"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sts"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: dbconfig <environment-name>")
	}
	envName := os.Args[1]
	sess := session.Must(session.NewSession())
	// Get the secret ARNs from CF
	cfSvc := cloudformation.New(sess)
	leo, err := cfSvc.ListExports(&cloudformation.ListExportsInput{})
	if err != nil {
		log.Fatalf("Error listing exports: %v", err)
	}
	var masterSecretARN, appSecretARN, auroraDBClusterID *string
	for _, e := range leo.Exports {
		if *e.Name == envName+"-AuroraMasterSecretARN" {
			masterSecretARN = e.Value
		} else if *e.Name == envName+"-AuroraAppSecretARN" {
			appSecretARN = e.Value
		} else if *e.Name == envName+"-AuroraDBClusterID" {
			auroraDBClusterID = e.Value
		}
	}
	// Get the app username and password from Secrets Manager
	secretSvc := secretsmanager.New(session.New())
	gsvi := &secretsmanager.GetSecretValueInput{
		SecretId: appSecretARN,
	}
	gsvo, err := secretSvc.GetSecretValue(gsvi)
	if err != nil {
		log.Fatalf("Error getting app secret: %v", err)
	}
	var appSecretMap map[string]interface{}
	err = json.Unmarshal([]byte(*gsvo.SecretString), &appSecretMap)
	if err != nil {
		log.Fatalf("Error decoding master secret JSON: %v", err)
	}
	// Create the app user
	appUsername, ok := appSecretMap["username"]
	if !ok {
		log.Fatalf("App username was not found")
	}
	appPassword, ok := appSecretMap["password"]
	if !ok {
		log.Fatalf("App password was not found")
	}
	stsSvc := sts.New(sess)
	gcio, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatalf("Error calling GetCallerIdentity: %v", err)
	}

	createUserSQL := fmt.Sprintf("create user %s with encrypted password '%s'", appUsername, appPassword)
	rdsSvc := rdsdataservice.New(sess)
	auroraDBClusterARN := fmt.Sprintf("arn:aws:rds:%s:%s:cluster:%s", *rdsSvc.Config.Region, *gcio.Account, *auroraDBClusterID)
	log.Printf("Creating application user '%s'", appUsername)
	eso, err := rdsSvc.ExecuteStatement(&rdsdataservice.ExecuteStatementInput{
		ResourceArn: &auroraDBClusterARN,
		Database:    aws.String("cms"),
		SecretArn:   masterSecretARN,
		Sql:         &createUserSQL,
	})
	if err != nil {
		log.Fatalf("Error creating app user: %v", err)
	}
	log.Printf("%d records updated", *eso.NumberOfRecordsUpdated)
}
