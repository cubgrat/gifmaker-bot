package domain

// Config represents application configuration
type Config struct {
	Bot struct {
		Token string `yaml:"token"`
	} `yaml:"bot"`
	GIF struct {
		Quality string `yaml:"quality"`
		FPS     int    `yaml:"fps"`
		Width   int    `yaml:"width"`
		Colors  int    `yaml:"colors"`
	} `yaml:"gif"`
	Processing struct {
		MaxConcurrent    int `yaml:"max_concurrent"`
		MaxVideoDuration int `yaml:"max_video_duration"`
	} `yaml:"processing"`
}

