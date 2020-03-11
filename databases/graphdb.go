package databases

import (
	"context"
	"encoding/json"
	"fmt"
	"pandor/logger"
	"pandor/models"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

// NewClient builds a new Dgraph Client
func NewClient() (*grpc.ClientConn, *dgo.Dgraph, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	return d, dgo.NewDgraphClient(
		api.NewDgraphClient(d)), err
}

// DropAll data including schema from the dgraph instance. This is useful
// for small examples such as this, since it puts dgraph into a clean
// state.
func DropAll(dg *dgo.Dgraph) error {
	err := dg.Alter(context.Background(),
		&api.Operation{DropOp: api.Operation_ALL})
	return err
}

// LoadSchema loads a new schema in Dgraph
func LoadSchema(schema string, dg *dgo.Dgraph) error {
	op := &api.Operation{}
	op.Schema = schema

	ctx := context.Background()
	err := dg.Alter(ctx, op)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}
	return err
}

// AddArticle adds an article to Dgraph
func AddArticle(article models.Article, dg *dgo.Dgraph) (*api.Response, error) {
	mu := &api.Mutation{
		CommitNow: true,
	}
	pb, err := json.Marshal(article)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	mu.SetJson = pb
	ctx := context.Background()
	response, err := dg.NewTxn().Mutate(ctx, mu)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}
	return response, err
}

// QueryWithVars allows to send a query to Dgraph with a var dict
func QueryWithVars(query string, variables map[string]string, dg *dgo.Dgraph) (api.Response, error) {
	ctx := context.Background()
	resp, err := dg.NewTxn().QueryWithVars(ctx, query, variables)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	return *resp, err
}

// Query allows to send a query to Dgraph
func Query(query string, dg *dgo.Dgraph) (api.Response, error) {
	ctx := context.Background()
	resp, err := dg.NewTxn().Query(ctx, query)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}
	fmt.Println(string(resp.Json))

	return *resp, err
}

// GetAuthorUID gives the UID of a given author
func GetAuthorUID(name string, dg *dgo.Dgraph) (string, error) {
	variables := map[string]string{"$name": name}
	query := `query GetUID($name: string){
							getuid(func: eq(name, $name)){
								uid
						  }
						}`
	resp, err := QueryWithVars(query, variables, dg)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	type Root struct {
		Authors []models.Author `json:"getuid"`
	}

	var r Root
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	if len(r.Authors) == 0 {
		return "", fmt.Errorf("No Author Found")
	}

	return r.Authors[0].UID, nil
}

// IsArticleInDB returns if an Article is already in Dgraph
func IsArticleInDB(title string, dg *dgo.Dgraph) (bool, error) {
	return true, nil
}
