package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"

	"resourcegroupstaggingapi_discovery_proof_of_concept/pkg/discovery"
)

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,

		// Config: aws.Config{
		// 	Endpoint:                      aws.String("http://127.0.0.1:8000"),
		// 	CredentialsChainVerboseErrors: aws.Bool(true),
		// 	LogLevel:                      aws.LogLevel(aws.LogDebug),
		// 	DisableSSL: aws.Bool(true),
		// },
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	d := discovery.New(sess)
	err := d.Do(ctx, []string{"rds:db", "rds:cluster"}, map[string][]string{
		"shard": {}}, func() error {
		fmt.Printf("in callback fn\n")
		return nil
	})
	if err != nil {
		panic(err)
	}

}
