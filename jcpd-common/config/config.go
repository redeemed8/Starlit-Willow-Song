package config

type MapConfig struct {
	Map_ map[string]string
}

func (m *MapConfig) Get(key string) string {
	return m.Map_[key]
}

var M = loadConfigMap()

const User = "jcpduser"

func loadConfigMap() MapConfig {
	return MapConfig{
		map[string]string{
			User: "127.0.0.1:7071",
		},
	}
}
