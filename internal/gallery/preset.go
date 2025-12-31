package gallery

// Preset represents a community configuration preset
type Preset struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Category    string `json:"category"` // e.g., "Cyberpunk", "Minimal Dark", "Retro Green"
	Downloads   int    `json:"downloads"`
	ConfigURL   string `json:"config_url"`
	ImageURL    string `json:"image_url,omitempty"` // Optional: ASCII art source image
	PreviewURL  string `json:"preview_url,omitempty"`
}

// PresetIndex represents the gallery index file
type PresetIndex struct {
	Version    string   `json:"version"`
	LastUpdate string   `json:"last_update"`
	Presets    []Preset `json:"presets"`
}

// Category represents a preset category for filtering
type Category struct {
	Name  string
	Count int
}

// SortOption represents how to sort presets
type SortOption int

const (
	SortByPopular SortOption = iota
	SortByNewest
	SortByName
)

func (s SortOption) String() string {
	switch s {
	case SortByPopular:
		return "popular"
	case SortByNewest:
		return "newest"
	case SortByName:
		return "name"
	default:
		return "popular"
	}
}
