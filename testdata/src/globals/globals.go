package globals

type Config struct{}

var def = &Config{} // want def:"nonnil"

var reassigned = &Config{}

var never *Config

var escaped = &Config{}

var initialized *Config // want initialized:"nonnil"

func init() {
	initialized = &Config{}
}

func Reset() {
	reassigned = nil
}

func Take() **Config { // want Take:"nonnil r0"
	return &escaped
}

func Default() *Config { // want Default:"nonnil r0"
	return def
}

func Reassigned() *Config {
	return reassigned // want `from package variable reassigned`
}

func Never() *Config {
	return never // want `from package variable never`
}

func Escaped() *Config {
	return escaped // want `from package variable escaped`
}

func Initialized() *Config { // want Initialized:"nonnil r0"
	return initialized
}
