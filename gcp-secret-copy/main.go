package main

import (
	"flag"
	"fmt"
	"os"
	"encoding/json"
	"context"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type Configuration struct {
	SourceProjects map[string]map[string]string `json:"source_projects"`
	Secrets        map[string]map[string][]string `json:"secrets"`
	TargetProjects map[string]map[string][]string `json:"target_projects"`
}

func loadConfig(configPath string) (*Configuration, error) {
	// Load the configuration from the specified path
	fmt.Println("Loading config from:", configPath)

	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config: %w", err)
	}
	defer f.Close()

	var config Configuration
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	// Check if the configuration contains valid source and target projects
	for customer, environment := range config.Secrets {
		for env, _ := range environment {
			if config.SourceProjects[customer][env] == "" {
				return nil, fmt.Errorf("source project for customer %s environment %s is empty", customer, env)
			}
			if len(config.TargetProjects[customer][env]) == 0 {
				return nil, fmt.Errorf("target project for customer %s environment %s is empty", customer, env)
			}
		}
	}

	return &config, nil
}

func getCustomers(config *Configuration, filter string) []string {
	customers := make([]string, 0, len(config.Secrets))
	for key := range config.Secrets {
		if (filter == "" || filter == key) {
			customers = append(customers, key)
		}
	}
	return customers
}

func readLatestVersion(ctx context.Context, client *secretmanager.Client, secret string) ([]byte, error){
	fmt.Println("Reading latest version of secret:", secret)

	resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secret + "/versions/latest",
	})
	if err != nil {
		fmt.Println("Error accessing secret version:", err)
		return nil, err
	}
	return resp.Payload.Data, nil
}

func addVersion(ctx context.Context, client *secretmanager.Client, secret string, data []byte) (string, error) {
	fmt.Println("Adding new version to secret:", secret)
	
	resp, err := client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret,
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	})
	if err != nil {
		fmt.Println("Error adding secret version:", err)
		return "", err
	}
	fmt.Println("Added secret version:", resp.Name)
	return resp.Name, nil
}

func copySecrets(config *Configuration, customer string, environment string, apply bool) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		fmt.Println("Error creating secret manager client:", err)
		return
	}
	defer client.Close()

	for env, secrets := range config.Secrets[customer] {
		if environment != "" && environment != env {
			continue
		}
		for _, secret := range secrets {
			sourceSecret := fmt.Sprintf("projects/%s/secrets/%s", config.SourceProjects[customer][env], secret)
			targetSecret := fmt.Sprintf("projects/%s/secrets/%s", config.TargetProjects[customer][env][0], secret)

			if apply {
				fmt.Println("Applying copy for secret:", secret)
				data, err := readLatestVersion(ctx, client, sourceSecret)
				if err != nil {
					fmt.Println("Error reading latest version of secret:", err)
					continue
				}
				_, err = addVersion(ctx, client, targetSecret, data)
				if err != nil {
					fmt.Println("Error adding secret version:", err)
					continue
				}
			} else {
				fmt.Println("Dry run: not applying copy for secret:", secret)
			}
		}
	}
}

func main() {
	apply := flag.Bool("apply", false, "Apply the changes")
	configPath := flag.String("config", "settings.json", "Path to the config file")
	customer := flag.String("customer", "", "Customer name. Limit the execution to this customer")
	environment := flag.String("env", "", "Environment name. Limit the execution to this environment")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	customers := getCustomers(config, *customer)
	fmt.Println("Customers to process:", customers)

	for _, customer := range customers {
		copySecrets(config, customer, *environment, *apply)
	}
}