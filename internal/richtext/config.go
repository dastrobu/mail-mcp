package richtext

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed config/default_styles.yaml
var defaultStylesYAML []byte

// RenderingConfig holds the raw YAML configuration (for parsing only)
type RenderingConfig struct {
	Defaults StyleConfig            `yaml:"defaults"`
	Styles   map[string]StyleConfig `yaml:"styles"`
}

// PrefixConfig defines styling for a prefix (from YAML)
type PrefixConfig struct {
	Content *string `yaml:"content,omitempty"` // Text to prepend
	Font    *string `yaml:"font,omitempty"`
	Size    *int    `yaml:"size,omitempty"`
	Color   *string `yaml:"color,omitempty"` // Web color format (#RRGGBB)
}

// StyleConfig defines raw styling properties from YAML (for parsing only)
type StyleConfig struct {
	Font         *string       `yaml:"font,omitempty"`
	Size         *int          `yaml:"size,omitempty"`
	Color        *string       `yaml:"color,omitempty"` // Web color format (#RRGGBB)
	MarginTop    *int          `yaml:"margin_top,omitempty"`
	MarginBottom *int          `yaml:"margin_bottom,omitempty"`
	Prefix       *PrefixConfig `yaml:"prefix,omitempty"` // Prefix styling
}

// PreparedConfig holds pre-computed styles ready for rendering.
// All styles have defaults merged and colors converted to AppleRGB.
type PreparedConfig struct {
	styles map[string]PreparedStyle
}

// PreparedPrefix defines a fully resolved prefix style
type PreparedPrefix struct {
	Content *string
	Font    *string
	Size    *int
	Color   *AppleRGB // Already converted, nil if not specified
}

// PreparedStyle defines a fully resolved style with colors already converted
type PreparedStyle struct {
	Font         *string
	Size         *int
	Color        *AppleRGB // Already converted, nil if not specified
	MarginTop    *int
	MarginBottom *int
	Prefix       *PreparedPrefix // Prefix styling
}

// AppleRGB represents Apple Mail's 16-bit RGB color format (0-65535)
type AppleRGB [3]int

// LoadConfig loads rendering configuration from a file path.
// If path is empty, uses the embedded default configuration.
// Returns PreparedConfig with all styles pre-merged and colors pre-converted.
func LoadConfig(path string) (*PreparedConfig, error) {
	var data []byte
	var err error

	if path == "" {
		data = defaultStylesYAML
	} else {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}
	}

	var config RenderingConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate the raw configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Prepare the config by merging defaults and converting colors
	prepared, err := prepareConfig(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare configuration: %w", err)
	}

	return prepared, nil
}

// validateConfig ensures the raw configuration has required fields and valid values
func validateConfig(config *RenderingConfig) error {
	// Validate defaults
	if config.Defaults.Font == nil || *config.Defaults.Font == "" {
		return fmt.Errorf("defaults.font is required")
	}
	if config.Defaults.Size == nil || *config.Defaults.Size <= 0 {
		return fmt.Errorf("defaults.size must be positive")
	}
	// Color is optional in defaults, but validate if present
	if config.Defaults.Color != nil && *config.Defaults.Color != "" {
		if _, err := webColorToAppleRGB(*config.Defaults.Color); err != nil {
			return fmt.Errorf("invalid defaults.color: %w", err)
		}
	}

	// Validate each style's colors (only if present)
	for name, style := range config.Styles {
		if style.Color != nil && *style.Color != "" {
			if _, err := webColorToAppleRGB(*style.Color); err != nil {
				return fmt.Errorf("invalid color in style %s: %w", name, err)
			}
		}
		if style.Size != nil && *style.Size < 0 {
			return fmt.Errorf("size in style %s cannot be negative", name)
		}
	}

	return nil
}

// prepareConfig merges defaults into each style and converts all colors
func prepareConfig(config *RenderingConfig) (*PreparedConfig, error) {
	// Convert default color once
	var defaultColor *AppleRGB
	if config.Defaults.Color != nil && *config.Defaults.Color != "" {
		rgb, err := webColorToAppleRGB(*config.Defaults.Color)
		if err != nil {
			return nil, fmt.Errorf("failed to convert defaults.color: %w", err)
		}
		defaultColor = &rgb
	}

	// Create prepared styles map
	styles := make(map[string]PreparedStyle)

	// Known style keys that should exist
	knownStyles := []string{
		"paragraph", "h1", "h2", "h3", "h4", "h5", "h6",
		"bold", "italic", "bold_italic", "strikethrough",
		"code", "code_block", "blockquote",
		"list", "list_item", "horizontal_rule", "link",
	}

	// Prepare each known style
	for _, styleName := range knownStyles {
		rawStyle, exists := config.Styles[styleName]
		if !exists {
			// Use defaults for missing styles
			styles[styleName] = PreparedStyle{
				Font:         config.Defaults.Font,
				Size:         config.Defaults.Size,
				Color:        defaultColor,
				MarginTop:    nil,
				MarginBottom: nil,
				Prefix:       nil,
			}
			continue
		}

		// Merge with defaults
		merged := PreparedStyle{
			Font:         config.Defaults.Font,
			Size:         config.Defaults.Size,
			Color:        defaultColor,
			MarginTop:    nil,
			MarginBottom: nil,
			Prefix:       nil,
		}

		if rawStyle.Font != nil {
			merged.Font = rawStyle.Font
		}
		if rawStyle.Size != nil {
			merged.Size = rawStyle.Size
		}
		if rawStyle.Color != nil && *rawStyle.Color != "" {
			rgb, err := webColorToAppleRGB(*rawStyle.Color)
			if err != nil {
				return nil, fmt.Errorf("failed to convert color in style %s: %w", styleName, err)
			}
			merged.Color = &rgb
		}
		if rawStyle.MarginTop != nil {
			merged.MarginTop = rawStyle.MarginTop
		}
		if rawStyle.MarginBottom != nil {
			merged.MarginBottom = rawStyle.MarginBottom
		}
		if rawStyle.Prefix != nil {
			preparedPrefix, err := preparePrefixConfig(rawStyle.Prefix, &merged)
			if err != nil {
				return nil, fmt.Errorf("failed to prepare prefix in style %s: %w", styleName, err)
			}
			merged.Prefix = preparedPrefix
		}

		styles[styleName] = merged
	}

	// Add any custom styles from config that aren't in knownStyles
	for styleName, rawStyle := range config.Styles {
		if _, exists := styles[styleName]; exists {
			continue // Already processed
		}

		// Merge custom style with defaults
		merged := PreparedStyle{
			Font:         config.Defaults.Font,
			Size:         config.Defaults.Size,
			Color:        defaultColor,
			MarginTop:    nil,
			MarginBottom: nil,
			Prefix:       nil,
		}

		if rawStyle.Font != nil {
			merged.Font = rawStyle.Font
		}
		if rawStyle.Size != nil {
			merged.Size = rawStyle.Size
		}
		if rawStyle.Color != nil && *rawStyle.Color != "" {
			rgb, err := webColorToAppleRGB(*rawStyle.Color)
			if err != nil {
				return nil, fmt.Errorf("failed to convert color in style %s: %w", styleName, err)
			}
			merged.Color = &rgb
		}
		if rawStyle.MarginTop != nil {
			merged.MarginTop = rawStyle.MarginTop
		}
		if rawStyle.MarginBottom != nil {
			merged.MarginBottom = rawStyle.MarginBottom
		}
		if rawStyle.Prefix != nil {
			preparedPrefix, err := preparePrefixConfig(rawStyle.Prefix, &merged)
			if err != nil {
				return nil, fmt.Errorf("failed to prepare prefix in style %s: %w", styleName, err)
			}
			merged.Prefix = preparedPrefix
		}

		styles[styleName] = merged
	}

	return &PreparedConfig{styles: styles}, nil
}

// preparePrefixConfig converts a PrefixConfig to PreparedPrefix
// Only sets fields that are explicitly provided in the prefix config.
// If not provided, the field is nil and will inherit from parent during rendering.
func preparePrefixConfig(prefix *PrefixConfig, parentStyle *PreparedStyle) (*PreparedPrefix, error) {
	prepared := &PreparedPrefix{
		Content: prefix.Content,
		Font:    nil,
		Size:    nil,
		Color:   nil,
	}

	// Only set fields that are explicitly provided in prefix config
	if prefix.Font != nil {
		prepared.Font = prefix.Font
	}
	if prefix.Size != nil {
		prepared.Size = prefix.Size
	}
	if prefix.Color != nil && *prefix.Color != "" {
		rgb, err := webColorToAppleRGB(*prefix.Color)
		if err != nil {
			return nil, fmt.Errorf("failed to convert prefix color: %w", err)
		}
		prepared.Color = &rgb
	}

	return prepared, nil
}

// webColorToAppleRGB converts web color format (#RRGGBB) to Apple's 16-bit RGB format
func webColorToAppleRGB(webColor string) (AppleRGB, error) {
	// Remove # prefix if present
	color := strings.TrimPrefix(webColor, "#")

	// Validate length
	if len(color) != 6 {
		return AppleRGB{}, fmt.Errorf("invalid color format %s: expected #RRGGBB", webColor)
	}

	// Parse hex values
	r, err := strconv.ParseInt(color[0:2], 16, 64)
	if err != nil {
		return AppleRGB{}, fmt.Errorf("invalid red component in %s: %w", webColor, err)
	}
	g, err := strconv.ParseInt(color[2:4], 16, 64)
	if err != nil {
		return AppleRGB{}, fmt.Errorf("invalid green component in %s: %w", webColor, err)
	}
	b, err := strconv.ParseInt(color[4:6], 16, 64)
	if err != nil {
		return AppleRGB{}, fmt.Errorf("invalid blue component in %s: %w", webColor, err)
	}

	// Convert 8-bit (0-255) to 16-bit (0-65535)
	// Formula: 16-bit = 8-bit × 257
	return AppleRGB{
		int(r * 257),
		int(g * 257),
		int(b * 257),
	}, nil
}

// GetStyle returns the prepared style for a given element type.
// Returns a zero-value style if the type doesn't exist (shouldn't happen with known styles).
func (c *PreparedConfig) GetStyle(elementType string) PreparedStyle {
	style, ok := c.styles[elementType]
	if !ok {
		log.Printf("WARNING: requested unknown style type %q, returning empty style\n", elementType)
		return PreparedStyle{}
	}
	return style
}
