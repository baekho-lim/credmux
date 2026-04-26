package keychain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Profile struct {
	Name     string            `json:"name"`
	Services map[string]string `json:"services"` // platform → env var name
}

const profileFile = "profiles.json"

func profilePath() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".credmux", profileFile), nil
}

// Load returns the profile for `name`. Missing file or missing entry yields a
// sensible default mapping so demos work without prior `profile create`.
func Load(name string) (*Profile, error) {
	path, err := profilePath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultProfile(name), nil
		}
		return nil, fmt.Errorf("open profiles: %w", err)
	}
	defer f.Close()

	var all []Profile
	if err := json.NewDecoder(f).Decode(&all); err != nil {
		return nil, fmt.Errorf("decode profiles: %w", err)
	}
	for i := range all {
		if all[i].Name == name {
			return &all[i], nil
		}
	}
	return defaultProfile(name), nil
}

func defaultProfile(name string) *Profile {
	return &Profile{
		Name: name,
		Services: map[string]string{
			"vercel":    "VERCEL_TOKEN",
			"github":    "GITHUB_TOKEN",
			"supabase":  "SUPABASE_KEY",
			"stripe":    "STRIPE_SECRET",
			"anthropic": "ANTHROPIC_API_KEY",
		},
	}
}

// Inject reads each service's token from Keychain and returns an env-var map
// suitable for merging into a child process environment. Missing entries are
// skipped silently — a partially-seeded profile still opens.
func Inject(p *Profile) map[string]string {
	out := map[string]string{}
	for platform, envVar := range p.Services {
		v, err := Find(p.Name, platform)
		if err != nil {
			continue
		}
		out[envVar] = v
	}
	return out
}
