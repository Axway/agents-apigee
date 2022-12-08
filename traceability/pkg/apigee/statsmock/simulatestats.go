package statsmock

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/Axway/agents-apigee/client/pkg/apigee"
	"github.com/Axway/agents-apigee/client/pkg/apigee/models"
	"github.com/Axway/agents-apigee/traceability/pkg/apigee/definitions"
)

const maxNumAPIs = 5
const maxNumTxns = 5

type simulate struct {
	client        *apigee.ApigeeClient
	products      apigee.Products
	messageCounts map[string][]int
	mapMutex      sync.Mutex
	logger        log.FieldLogger
}

func NewStatsMock(client *apigee.ApigeeClient, products apigee.Products) definitions.StatsClient {
	return &simulate{
		client:        client,
		products:      products,
		logger:        log.NewFieldLogger().WithComponent("simulate").WithPackage("statsmock"),
		mapMutex:      sync.Mutex{},
		messageCounts: make(map[string][]int),
	}
}

func (s *simulate) GetEnvironments() []string {
	return s.client.GetEnvironments()
}

func (s *simulate) GetProduct(productName string) (*models.ApiProduct, error) {
	return s.client.GetProduct(productName)
}

func (s *simulate) GetStats(env, dimension, metricSelect string, start, end time.Time) (*models.Metrics, error) {
	numProducts := rand.Intn(maxNumAPIs)

	response := &models.Metrics{
		Environments: []models.MetricsEnvironments{
			{
				Name:       env,
				Dimensions: []models.MetricsDimensions{},
			},
		},
	}

	startProd := rand.Intn(len(s.products) - numProducts)
	for i := startProd; i < startProd+numProducts; i++ {
		p := s.products[i]
		response.Environments[0].Dimensions = append(response.Environments[0].Dimensions, s.genProductMetrics(p, start, end))
	}

	s.printStats()

	return response, nil
}

func (s *simulate) printStats() {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	for product, stats := range s.messageCounts {
		s.logger.
			WithField("product", product).
			WithField("total", stats[0]).
			WithField("success", stats[1]).
			WithField("policyErr", stats[2]).
			WithField("serverErr", stats[3]).
			Trace("*****stats*****")
	}
}

func (s *simulate) genProductMetrics(productName string, start, end time.Time) models.MetricsDimensions {
	s.mapMutex.Lock()
	if _, ok := s.messageCounts[productName]; !ok {
		s.messageCounts[productName] = []int{0, 0, 0, 0}
	}
	s.mapMutex.Unlock()
	metric := models.MetricsDimensions{
		Name: productName,
		Metrics: []models.MetricsMetrics{
			{
				Name:         "sum(message_count)",
				MetricValues: []models.MetricsValues{},
			},
			{
				Name:         "sum(policy_error)",
				MetricValues: []models.MetricsValues{},
			},
			{
				Name:         "sum(target_error)",
				MetricValues: []models.MetricsValues{},
			},
			{
				Name:         "avg(total_response_time)",
				MetricValues: []models.MetricsValues{},
			},
		},
	}

	numMinutes := int(math.RoundToEven(end.Sub(start).Minutes()))

	for i := 0; i < numMinutes; i++ {
		timestamp := start.Add(time.Minute * time.Duration(i)).UnixMilli()
		var pErr int
		var tErr int
		total := rand.Intn(maxNumTxns)
		if total > 0 {
			pErr = rand.Intn(total)
		}
		if total-pErr > 0 {
			tErr = rand.Intn(total - pErr)
		}
		resTime := rand.Intn(100)
		s.mapMutex.Lock()
		s.messageCounts[productName][0] += total
		s.messageCounts[productName][1] += (total - pErr - tErr)
		s.messageCounts[productName][2] += pErr
		s.messageCounts[productName][3] += tErr
		s.mapMutex.Unlock()

		metric.Metrics[0].MetricValues = append(metric.Metrics[0].MetricValues, models.MetricsValues{
			Timestamp: timestamp,
			Value:     fmt.Sprintf("%v.0", total),
		})
		metric.Metrics[1].MetricValues = append(metric.Metrics[1].MetricValues, models.MetricsValues{
			Timestamp: timestamp,
			Value:     fmt.Sprintf("%v.0", pErr),
		})
		metric.Metrics[2].MetricValues = append(metric.Metrics[2].MetricValues, models.MetricsValues{
			Timestamp: timestamp,
			Value:     fmt.Sprintf("%v.0", tErr),
		})
		metric.Metrics[3].MetricValues = append(metric.Metrics[3].MetricValues, models.MetricsValues{
			Timestamp: timestamp,
			Value:     fmt.Sprintf("%v.0", resTime),
		})
	}

	return metric
}
