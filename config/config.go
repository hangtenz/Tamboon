package config

type Env struct {
	OmisePublicKey string
	OmiseSecretKey string
	MaxGoRoutine   int
}

//Config variable
func LoadDevEnv() *Env {
	return &Env{
		OmisePublicKey: "pkey_test_5nctkg8j7wjdoi6laqa",
		OmiseSecretKey: "skey_test_5nctdm87qtmttlv8gi9",
		MaxGoRoutine:   2,
	}
}
