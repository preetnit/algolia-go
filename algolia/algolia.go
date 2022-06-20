package algolia

import (
	"encoding/json"
	"fmt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/preetnit/algolia-go/config"
	"io"
	"sync"
)

type Record struct {
	ObjectID               string `json:"objectID"`
	Name                   string `json:"hot_keywords"`
	ProductImportanceScore string `json:"product_importance_score"`
}

func UpdateIndex(cfg *config.Application, operations []search.BatchOperationIndexed, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Updating the index ", cfg.AlgoliaIndexName)
	client := search.NewClient(cfg.AlgoliaAppId, cfg.AlgoliaAPIKey)
	res, err := client.MultipleBatch(operations)
	fmt.Println("Complete update ", res, err)
}

func ProcessData(results *io.PipeReader, cfg *config.Application) {
	var wg sync.WaitGroup
	resReader := json.NewDecoder(results)
	var operations []search.BatchOperationIndexed
	var record Record
	for {
		err := resReader.Decode(&record)
		if err == io.EOF {
			fmt.Println("EOF")
			wg.Add(1)
			go UpdateIndex(cfg, operations, &wg)
			break
		}

		fmt.Printf("Record %v\n", record)
		operations = append(operations, search.BatchOperationIndexed{
			IndexName: cfg.AlgoliaIndexName,
			BatchOperation: search.BatchOperation{
				Action: search.PartialUpdateObjectNoCreate,
				Body:   record,
			},
		})

		if len(operations) == cfg.AlgoliaOpsBatchSize {
			wg.Add(1)
			go UpdateIndex(cfg, operations, &wg)
			operations = nil
		}
	}
	wg.Wait()
}
