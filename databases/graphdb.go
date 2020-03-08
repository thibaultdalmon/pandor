package databases

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pandor/models"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

// NewClient builds a new Dgraph Client
func NewClient() (*dgo.Dgraph, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d)), err
}

// Drop all data including schema from the dgraph instance. This is useful
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
		log.Fatal(err)
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
		log.Fatal(err)
	}

	mu.SetJson = pb
	ctx := context.Background()
	response, err := dg.NewTxn().Mutate(ctx, mu)
	if err != nil {
		log.Fatal(err)
	}
	return response, err
}

// Query allows to send a query to Dgraph
func Query(query string, variables map[string]string, dg *dgo.Dgraph) (*api.Response, error) {
	ctx := context.Background()
	resp, err := dg.NewTxn().QueryWithVars(ctx, query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return resp, err
}

// Format allows to extract the information from a Response
func Format(resp *api.Response) error {
	type Root struct {
		Me []models.Article `json:"me"`
	}

	var r Root
	err := json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(r, "", "\t")
	fmt.Printf("%s\n", out)

	return err
}
