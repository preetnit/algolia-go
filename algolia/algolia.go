package algolia

import (
	"fmt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/preetnit/algolia-go/config"
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
