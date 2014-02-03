package aws

import (
	"github.com/robfig/revel"
	"github.com/ahmad972/goamz/aws"
	"github.com/ahmad972/goamz/ec2"
	"awsgraph/app/models"
	"time"
)

func ListInstances() models.Arbor {

	var arbor models.Arbor

	nodes := map[string]models.Node{}

	results := asyncApiCalls()
	//	   asyncApiCalls()
	for _, result := range results {

		for _,res := range result.Reservations {
			for _,instance := range res.Instances {

				var name, class string
				for _,tag := range instance.Tags {
					if tag.Key == "Name" {
						name = tag.Value
					}
					if tag.Key == "Class" {
						class = tag.Value
					}
				}

				revel.INFO.Printf("Instance: %s", instance.InstanceId)
				nodes[instance.InstanceId] = models.Node{
					Name:           name,
					Class:		class,
					Region:		instance.AvailZone,
				}
			}
		}
	}


	arbor.Nodes = nodes

	return arbor
}

func asyncApiCalls() []ec2.InstancesResp {
	ch := make(chan *ec2.InstancesResp)
	responses := []ec2.InstancesResp{}

	for name,region := range aws.Regions {
		go func(name string, region aws.Region) {
			revel.INFO.Printf("Fetching region: %s", name)
			auth, err := aws.GetAuth("", "", "", time.Time{})

			if err != nil {
				panic(err)
			}
			e := ec2.New(auth, region)
			resp, err := e.Instances(nil, nil)

			if err != nil {
				panic(err)
			}

			ch <- resp
		}(name, region)
	}

	for {
		select {
		case r := <-ch:
			revel.INFO.Printf("Region fetched with requestId: %s", r.RequestId)
			responses = append(responses, *r)
			if len(responses) == len(aws.Regions) {
				return responses
			}
		case <-time.After(15 * time.Second):
			revel.INFO.Print("waiting...")
		}
	}
	return responses
}
