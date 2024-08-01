package config_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/kxnes/go-interviews/apicache/internal/config"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"github.com/stretchr/testify/require"
)

var errParseEnv = errors.New("parse env error: " +
	"Debug(\"invalid\"): strconv.ParseBool: parsing \"invalid\": invalid syntax")

func TestUnitNew(t *testing.T) {
	t.Parallel()

	wantJSON := "{\"debug\":true,\"server\":{\"address\":\"0.0.0.0:8080\",\"shutdownTimeout\":1000000000}," +
		"\"driver\":{\"name\":\"machine\",\"address\":\"http://test.loc\",\"maxConn\":10,\"connTimeout\":1000000000}}"

	got, err := config.New(toolkit.EnvFile())

	gotJSON, errJ := json.Marshal(got)
	if errJ != nil {
		panic(errJ)
	}

	toolkit.Assert(t, toolkit.Got(err, string(gotJSON)), toolkit.Want(wantJSON, nil))
}

func TestUnitMustNew(t *testing.T) {
	type args struct {
		filename string
	}

	tests := []struct {
		name string
		args args
	}{
		{name: "success", args: args{filename: toolkit.EnvFile()}},
		{name: "failure", args: args{filename: ".env.invalid"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "failure" {
				t.Setenv("APICACHE_DEBUG", "invalid")

				require.Panics(t, func() {
					_ = config.MustNew(test.args.filename)
				})
			} else {
				require.NotPanics(t, func() {
					_ = config.MustNew(test.args.filename)
				})
			}
		})
	}
}

func TestUnitSettingsEnvDependent(t *testing.T) {
	t.Setenv("APICACHE_DEBUG", "invalid")

	obj, err := config.New(toolkit.EnvFile())

	toolkit.Assert(t, toolkit.Got(err, obj), toolkit.Want[*config.Settings](nil, errParseEnv))

	t.Setenv("APICACHE_DEBUG", "true")
	t.Setenv("APICACHE_DRIVER_NAME", "invalid")

	obj, err = config.New(toolkit.EnvFile())

	toolkit.Assert(t, toolkit.Got(err, obj), toolkit.Want[*config.Settings](nil, config.ErrInvalidDriver))

	for _, driverName := range []string{"machine", "memcached", "redis"} {
		t.Setenv("APICACHE_DRIVER_NAME", driverName)

		obj, _ = config.New(toolkit.EnvFile())

		toolkit.Assert(t, toolkit.Got(nil, obj.Driver.Name), toolkit.Want[config.Driver](config.Driver(driverName), nil))
	}
}
