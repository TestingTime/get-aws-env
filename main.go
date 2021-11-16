package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

var env string
var diff bool
var prefix string
var path string

func init() {
	flag.StringVar(&env, "e", "", "The environment namespace to pull the variables from")
	flag.StringVar(&path, "t", "/env", "Path to pull variables from. Defaults to /env")
	flag.BoolVar(&diff, "d", false, "True to only pull environment variables not present")
	flag.StringVar(&prefix, "p", "", "Prefix each line of the output")
	flag.Parse()
	if env == "" {
		log.Fatal("You must supply the environment name")
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("configuration error: %s", err.Error())
	}
	client := ssm.NewFromConfig(cfg)
	parameters, err := GetParametersByPath(fmt.Sprintf("%s/%s", path, env), client)
	if err != nil {
		log.Fatalf("error requesting parameters: %s", err.Error())
	}
	if diff {
		for _, key := range GetEnvironmentKeys() {
			delete(parameters, key)
		}
	}
	for key, value := range parameters {
		fmt.Printf("%s%s='%s'\n", prefix, key, value)
	}
}

func GetParametersByPath(path string, client ssm.GetParametersByPathAPIClient) (map[string]string, error) {
	input := &ssm.GetParametersByPathInput{
		Path: &path,
	}
	parameters := make(map[string]string)
	var output *ssm.GetParametersByPathOutput
	var err error
	paginator := ssm.NewGetParametersByPathPaginator(client, input)
	for paginator.HasMorePages() {
		if output, err = paginator.NextPage(context.TODO()); err != nil {
			return nil, err
		}
		for _, param := range output.Parameters {
			parameters[(*param.Name)[len(path)+1:]] = *param.Value
		}
	}
	return parameters, nil
}

func GetEnvironmentKeys() []string {
	environ := os.Environ()
	for i := range environ {
		environ[i] = strings.Split(environ[i], "=")[0]
	}
	return environ
}
