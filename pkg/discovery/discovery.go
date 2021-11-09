package discovery

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
)

type DiscoverFn func() error
type HandlerFn func(ctx context.Context, arn []*string, fn DiscoverFn) error

var (
	resourceTypeRDSDatabaseInstanceRE = regexp.MustCompile("^arn:aws:rds:.*?:.*?:db:.*")
	resourceTypeRDSDatabaseClusterRE  = regexp.MustCompile("^arn:aws:rds:.*?:.*?:cluster:.*")
	resourceTypeElasticCacheRE        = regexp.MustCompile("jakfjdfkadjk")
)

type DiscoverAndDo struct {
	taggingSVC resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI
	rdsSVC     rdsiface.RDSAPI
}

func New(sess *session.Session) *DiscoverAndDo {
	return &DiscoverAndDo{
		taggingSVC: resourcegroupstaggingapi.New(sess),
		rdsSVC:     rds.New(sess),
	}
}

// Do
// the callback func 'onDiscover' will be invoked for each page of discovered objects
func (d *DiscoverAndDo) Do(ctx context.Context, resourceTypeFilters []string, tagFilters map[string][]string, onDiscover DiscoverFn) error {
	tf := make([]*resourcegroupstaggingapi.TagFilter, 0, 10)
	for k, v := range tagFilters {
		tf = append(tf, &resourcegroupstaggingapi.TagFilter{Key: aws.String(k), Values: aws.StringSlice(v)})
	}

	input := resourcegroupstaggingapi.GetResourcesInput{
		ResourceTypeFilters: aws.StringSlice(resourceTypeFilters),
		TagFilters:          tf,
	}

	pageNum := 0
	err := d.taggingSVC.GetResourcesPages(&input,
		func(page *resourcegroupstaggingapi.GetResourcesOutput, lastPage bool) bool {
			pageNum++
			fmt.Println(page)
			var handlerFn HandlerFn
			handlerFns := make(map[string]HandlerFn)
			arns := make([]*string, 0, 100)
			for i := range page.ResourceTagMappingList {
				arn := page.ResourceTagMappingList[i].ResourceARN
				if arn == nil {
					return false //fmt.Errorf("whoa, that's super bad nil ARN from aws")
				}

				arns = append(arns, arn)
				switch {
				case resourceTypeRDSDatabaseClusterRE.MatchString(*arn):
					fmt.Printf("found resource type matching %v, arn: %v\n", resourceTypeRDSDatabaseInstanceRE, *arn)
					handlerFns[*arn] = d.DiscoveryHandlerDatabaseCluster
				case resourceTypeRDSDatabaseInstanceRE.MatchString(*arn):
					fmt.Printf("found resource type matching %v, arn: %v\n", resourceTypeRDSDatabaseInstanceRE, *arn)
					handlerFns[*arn] = d.DiscoveryHandlerDatabaseInstance
				case resourceTypeElasticCacheRE.MatchString(*arn):
					fmt.Printf("found resource type matching %v, arn: %v\n", resourceTypeRDSDatabaseInstanceRE, *arn)
					handlerFns[*arn] = d.DiscoveryHandlerPrintf
				default:
					fmt.Printf("no handlers for resource type arn %v\n", *arn)
				}
			}

			// TODO: review for the size of the ARN list, what's reasonable? 10?, 100?
			err := handlerFn(ctx, arns, onDiscover)
			if err != nil {
				panic(err)
			}

			return !lastPage
		})
	if err != nil {
		panic(err)
	}
	return nil
}

func (d *DiscoverAndDo) DiscoveryHandlerDatabaseCluster(ctx context.Context, arns []*string, fn DiscoverFn) error {
	fmt.Printf("discovery handler databse cluster func\n")
	return nil
}

func (d *DiscoverAndDo) DiscoveryHandlerDatabaseInstance(ctx context.Context, arns []*string, fn DiscoverFn) error {
	fmt.Printf("discovery handler databse cluster func\n")
	input := rds.DescribeDBInstancesInput{
		Filters: []*rds.Filter{{
			Name:   aws.String("db-instance-id"),
			Values: arns,
		}},
	}

	err := d.rdsSVC.DescribeDBInstancesPagesWithContext(ctx, &input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		for j := range page.DBInstances {
			fmt.Printf("%+v", page.DBInstances[j])
		}
		return false
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *DiscoverAndDo) DiscoveryHandlerPrintf(ctx context.Context, arns []*string, fn DiscoverFn) error {
	fmt.Printf("discovery handler printf func\n")
	return nil
}
