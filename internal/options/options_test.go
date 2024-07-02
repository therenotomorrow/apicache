package options

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kxnes/go-interviews/apicache/internal/apicache"
	"github.com/kxnes/go-interviews/apicache/internal/fs"
)

const testdata = "../../test/testdata/"

func TestLoad(t *testing.T) {
	cases := []struct {
		name string
		want *Options
		err  string
	}{
		{
			name: "valid.json",
			want: &Options{
				APICache: &apicache.Options{Addr: "127.0.0.1:8080"},
				FileSystem: &fs.Options{
					MaxConn: 10,
					Timeout: 10,
				},
				Driver: &Optional{
					Name: "redis",
					Addr: "127.0.0.1:6379",
				},
			},
		},
		{
			name: "invalid.json",
			want: nil,
			err:  "decode config error",
		},
		{
			name: "not-exist.json",
			want: nil,
			err:  "open config error",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				err := recover()

				if c.err == "" {
					if err != nil {
						t.Errorf("panic = %v not expected", err.(string))
					}
					return
				}

				if !strings.HasPrefix(err.(string), c.err) {
					t.Errorf("panic got = %v, want = %v", err, c.err)
				}
			}()

			got := Load(testdata + c.name)

			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Load() = %v, want = %v", got, c.want)
			}
		})
	}
}
