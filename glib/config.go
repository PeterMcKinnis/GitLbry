package glib

import (
	"encoding/json"
	"os"
	"path"
)

type glConfig struct {
	// Default configuration
	Default glRepoConfig;

	// Configuration for idividual repositories
	ByUrl map[string]glRepoConfig;
}

type glRepoConfig struct {
	PushAs *glChannel
}

type glChannel struct {
	ClaimId string
	Name string 
}

func newConfig() *glConfig {
	return &glConfig{
		ByUrl: map[string]glRepoConfig{},
	}
}

func loadConfig() *glConfig {

	configPath, err := os.UserConfigDir();
	if err != nil {
		return newConfig();
	}
	path := path.Join(configPath, "gitlbry", "config.json");

	fid, err := os.Open(path);
	if err != nil {
		return newConfig();
	}
	defer fid.Close()

	var result *glConfig;
	err = json.NewDecoder(fid).Decode(&result);
	if err != nil {
		return newConfig();
	}
	return result;
}

func (c *glConfig) save() error {

	// Convert to json
	b, err := json.Marshal(c);

	// Create config directory if necessary
	configPath, err := os.UserConfigDir();
	if err != nil {
		return err;
	}
	dir := path.Join(configPath, "gitlbry");
	err = os.MkdirAll(dir, 0777);
	if err != nil {
		return err;
	}

	path := path.Join(configPath, "gitlbry", "config.json");
	return os.WriteFile(path, b, 0666);

}