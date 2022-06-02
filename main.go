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

type envFlags []string

func (env *envFlags) String() string {
	return strings.Join(*env, "; ")
}

func (env *envFlags) Set(value string) error {
	envKeys := strings.Split(value, ",")
	for _, envKey := range envKeys {
		*env = append(*env, envKey)
	}
	return nil
}

var env envFlags
var diff bool
var prefix string
var path string

func init() {
	flag.Var(&env, "e", "The environment namespace(s) to pull the variables from")
	flag.StringVar(&path, "t", "/env", "Path to pull variables from. Defaults to /env")
	flag.BoolVar(&diff, "d", false, "True to only pull environment variables not present")
	flag.StringVar(&prefix, "p", "", "Prefix each line of the output")
	flag.Parse()
	if len(env) == 0 {
		log.Fatal("You must supply at least one environment name")
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("configuration error: %s", err.Error())
	}
	client := ssm.NewFromConfig(cfg)
	for _, envKey := range env {
		if err := ProcessEnvKey(envKey, client); err != nil {
			log.Fatalf("error requesting parameters: %s", err.Error())
		}
	}
}

func ProcessEnvKey(envKey string, client ssm.GetParametersByPathAPIClient) error {
	parameters, err := GetParametersByPath(fmt.Sprintf("%s/%s", path, envKey), client)
	if err != nil {
		return err
	}
	if diff {
		for _, key := range GetEnvironmentKeys() {
			delete(parameters, key)
		}
	}
	for key, value := range parameters {
		fmt.Printf("%s%s='%s'\n", prefix, key, value)
	}
	return nil
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
