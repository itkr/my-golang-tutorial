package main

type Config struct {
	Email     string
	FilePath  string
	ProjectID string
}

func getConfig() map[string]Config {
	configures := make(map[string]Config, 1)
	configures["project-999"] = Config{
		Email:     "test@developer.gserviceaccount.com",
		FilePath:  "test.pem",
		ProjectID: "project-999",
	}
	return configures
}
