package config

// *** THIS IS GENERATED CODE: DO NOT EDIT ***

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	falseVal = false
	trueVal  = true
)

func TestDuration_String(t *testing.T) {
	f := func(d time.Duration, exp string) {
		actual := Duration(d).String()
		if actual != exp {
			t.Errorf("String() for duration %d expected to return %s, but got %s", d, exp, actual)
		}
	}
	f(time.Second, "1s")
	f(time.Second*30, "30s")
	f(time.Minute, "1m0s")
	f(time.Second*90, "1m30s")
	f(0, "0s")
}

func TestDuration_JSON(t *testing.T) {
	f := func(d time.Duration, exp string) {
		v := Duration(d)
		bytes, err := json.Marshal(&v)
		if err != nil {
			t.Fatalf("Unable to json.Marshal our Duration of %+v: %v", v, err)
		}
		if string(bytes) != exp {
			t.Errorf("Marshaled duration expected to generate %v, but got %v", exp, string(bytes))
		}
		var decoded Duration
		if err := json.Unmarshal(bytes, &decoded); err != nil {
			t.Errorf("Got error trying to unmarshal %v to a Duration: %v", string(bytes), err)
		}
		if decoded != v {
			t.Errorf("Encoded/Decoded duration no longer equal!, original %v, round-tripped %v", v, decoded)
		}
	}
	f(0, `"0s"`)
	f(time.Second, `"1s"`)
	f(time.Minute*5, `"5m0s"`)
	f(time.Second*90, `"1m30s"`)
	f(time.Hour*2, `"2h0m0s"`)
	f(time.Millisecond*10, `"10ms"`)
}

func TestDuration_JSONDecode(t *testing.T) {
	f := func(j string, exp time.Duration) {
		var act Duration
		err := json.Unmarshal([]byte(j), &act)
		if err != nil {
			t.Fatalf("Unable to json.Unmarshal %s: %v", j, err)
		}
		if act.TimeDuration() != exp {
			t.Errorf("Expecting json of %s to production duration %s, but got %s", j, exp, act)
		}
	}
	f(`"5m"`, time.Minute*5)
	f(`120`, time.Second*120)
	f(`0`, 0)
	f(`"1m5s"`, time.Second*65)
}

func Test_overrideBool(t *testing.T) {
	d := &trueVal
	var zero *bool
	overrideBool(&d, &zero)
	require.NotEqual(t, d, zero, "overrideBool shouldn't have overriden the value as the override is the default/zero value. value now %v", d)
	o := &falseVal
	overrideBool(&d, &o)
	require.Equal(t, d, o, "overrideBool should of overriden the value but didn't. value %v, expecting %v", d, o)
}

func Test_overrideInt(t *testing.T) {
	d := -42
	var zero int
	overrideInt(&d, &zero)
	require.NotEqual(t, d, zero, "overrideInt shouldn't have overriden the value as the override is the default/zero value. value now %v", d)
	o := 42
	overrideInt(&d, &o)
	require.Equal(t, d, o, "overrideInt should of overriden the value but didn't. value %v, expecting %v", d, o)
}

func Test_overrideRepoLogLevelSlice(t *testing.T) {
	d := []RepoLogLevel{
		{
			Repo:    "one",
			Package: "one",
			Level:   "one"},
	}
	var zero []RepoLogLevel
	overrideRepoLogLevelSlice(&d, &zero)
	require.NotEqual(t, d, zero, "overrideRepoLogLevelSlice shouldn't have overriden the value as the override is the default/zero value. value now %v", d)
	o := []RepoLogLevel{
		{
			Repo:    "two",
			Package: "two",
			Level:   "two"},
	}
	overrideRepoLogLevelSlice(&d, &o)
	require.Equal(t, d, o, "overrideRepoLogLevelSlice should of overriden the value but didn't. value %v, expecting %v", d, o)
}

func Test_overrideString(t *testing.T) {
	d := "one"
	var zero string
	overrideString(&d, &zero)
	require.NotEqual(t, d, zero, "overrideString shouldn't have overriden the value as the override is the default/zero value. value now %v", d)
	o := "two"
	overrideString(&d, &o)
	require.Equal(t, d, o, "overrideString should of overriden the value but didn't. value %v, expecting %v", d, o)
}

func Test_overrideStrings(t *testing.T) {
	d := []string{"a"}
	var zero []string
	overrideStrings(&d, &zero)
	require.NotEqual(t, d, zero, "overrideStrings shouldn't have overriden the value as the override is the default/zero value. value now %v", d)
	o := []string{"b", "b"}
	overrideStrings(&d, &o)
	require.Equal(t, d, o, "overrideStrings should of overriden the value but didn't. value %v, expecting %v", d, o)
}

func TestConfiguration_overrideFrom(t *testing.T) {
	orig := Configuration{
		Region:      "one",
		Environment: "one",
		ServiceName: "one",
		ClusterName: "one",
		CryptoProv: CryptoProv{
			Default:   "one",
			Providers: []string{"a"}},
		Audit: Logger{
			Directory:  "one",
			MaxAgeDays: -42,
			MaxSizeMb:  -42},
		Logger: Logger{
			Directory:  "one",
			MaxAgeDays: -42,
			MaxSizeMb:  -42},
		LogLevels: []RepoLogLevel{
			{
				Repo:    "one",
				Package: "one",
				Level:   "one"},
		},
		TrustyClient: TrustyClient{
			Servers: []string{"a"},
			ClientTLS: TLSInfo{
				CertFile:       "one",
				KeyFile:        "one",
				TrustedCAFile:  "one",
				ClientCertAuth: &trueVal}}}
	dest := orig
	var zero Configuration
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "Configuration.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := Configuration{
		Region:      "two",
		Environment: "two",
		ServiceName: "two",
		ClusterName: "two",
		CryptoProv: CryptoProv{
			Default:   "two",
			Providers: []string{"b", "b"}},
		Audit: Logger{
			Directory:  "two",
			MaxAgeDays: 42,
			MaxSizeMb:  42},
		Logger: Logger{
			Directory:  "two",
			MaxAgeDays: 42,
			MaxSizeMb:  42},
		LogLevels: []RepoLogLevel{
			{
				Repo:    "two",
				Package: "two",
				Level:   "two"},
		},
		TrustyClient: TrustyClient{
			Servers: []string{"b", "b"},
			ClientTLS: TLSInfo{
				CertFile:       "two",
				KeyFile:        "two",
				TrustedCAFile:  "two",
				ClientCertAuth: &falseVal}}}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "Configuration.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := Configuration{
		Region: "one"}
	dest.overrideFrom(&o2)
	exp := o

	exp.Region = o2.Region
	require.Equal(t, dest, exp, "Configuration.overrideFrom should have overriden the field Region. value now %#v, expecting %#v", dest, exp)
}

func TestCryptoProv_overrideFrom(t *testing.T) {
	orig := CryptoProv{
		Default:   "one",
		Providers: []string{"a"}}
	dest := orig
	var zero CryptoProv
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "CryptoProv.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := CryptoProv{
		Default:   "two",
		Providers: []string{"b", "b"}}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "CryptoProv.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := CryptoProv{
		Default: "one"}
	dest.overrideFrom(&o2)
	exp := o

	exp.Default = o2.Default
	require.Equal(t, dest, exp, "CryptoProv.overrideFrom should have overriden the field Default. value now %#v, expecting %#v", dest, exp)
}

func TestLogger_overrideFrom(t *testing.T) {
	orig := Logger{
		Directory:  "one",
		MaxAgeDays: -42,
		MaxSizeMb:  -42}
	dest := orig
	var zero Logger
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "Logger.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := Logger{
		Directory:  "two",
		MaxAgeDays: 42,
		MaxSizeMb:  42}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "Logger.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := Logger{
		Directory: "one"}
	dest.overrideFrom(&o2)
	exp := o

	exp.Directory = o2.Directory
	require.Equal(t, dest, exp, "Logger.overrideFrom should have overriden the field Directory. value now %#v, expecting %#v", dest, exp)
}

func TestLogger_Getters(t *testing.T) {
	orig := Logger{
		Directory:  "one",
		MaxAgeDays: -42,
		MaxSizeMb:  -42}

	gv0 := orig.GetDirectory()
	require.Equal(t, orig.Directory, gv0, "Logger.GetDirectoryCfg() does not match")

	gv1 := orig.GetMaxAgeDays()
	require.Equal(t, orig.MaxAgeDays, gv1, "Logger.GetMaxAgeDaysCfg() does not match")

	gv2 := orig.GetMaxSizeMb()
	require.Equal(t, orig.MaxSizeMb, gv2, "Logger.GetMaxSizeMbCfg() does not match")

}

func TestMetrics_overrideFrom(t *testing.T) {
	orig := Metrics{
		Disabled: &trueVal,
		Provider: "one"}
	dest := orig
	var zero Metrics
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "Metrics.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := Metrics{
		Disabled: &falseVal,
		Provider: "two"}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "Metrics.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := Metrics{
		Disabled: &trueVal}
	dest.overrideFrom(&o2)
	exp := o

	exp.Disabled = o2.Disabled
	require.Equal(t, dest, exp, "Metrics.overrideFrom should have overriden the field Disabled. value now %#v, expecting %#v", dest, exp)
}

func TestMetrics_Getters(t *testing.T) {
	orig := Metrics{
		Disabled: &trueVal,
		Provider: "one"}

	gv0 := orig.GetDisabled()
	require.Equal(t, orig.Disabled, &gv0, "Metrics.GetDisabled() does not match")

	gv1 := orig.GetProvider()
	require.Equal(t, orig.Provider, gv1, "Metrics.GetProviderCfg() does not match")

}

func TestRepoLogLevel_overrideFrom(t *testing.T) {
	orig := RepoLogLevel{
		Repo:    "one",
		Package: "one",
		Level:   "one"}
	dest := orig
	var zero RepoLogLevel
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "RepoLogLevel.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := RepoLogLevel{
		Repo:    "two",
		Package: "two",
		Level:   "two"}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "RepoLogLevel.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := RepoLogLevel{
		Repo: "one"}
	dest.overrideFrom(&o2)
	exp := o

	exp.Repo = o2.Repo
	require.Equal(t, dest, exp, "RepoLogLevel.overrideFrom should have overriden the field Repo. value now %#v, expecting %#v", dest, exp)
}

func TestTLSInfo_overrideFrom(t *testing.T) {
	orig := TLSInfo{
		CertFile:       "one",
		KeyFile:        "one",
		TrustedCAFile:  "one",
		ClientCertAuth: &trueVal}
	dest := orig
	var zero TLSInfo
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "TLSInfo.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := TLSInfo{
		CertFile:       "two",
		KeyFile:        "two",
		TrustedCAFile:  "two",
		ClientCertAuth: &falseVal}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "TLSInfo.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := TLSInfo{
		CertFile: "one"}
	dest.overrideFrom(&o2)
	exp := o

	exp.CertFile = o2.CertFile
	require.Equal(t, dest, exp, "TLSInfo.overrideFrom should have overriden the field CertFile. value now %#v, expecting %#v", dest, exp)
}

func TestTLSInfo_Getters(t *testing.T) {
	orig := TLSInfo{
		CertFile:       "one",
		KeyFile:        "one",
		TrustedCAFile:  "one",
		ClientCertAuth: &trueVal}

	gv0 := orig.GetCertFile()
	require.Equal(t, orig.CertFile, gv0, "TLSInfo.GetCertFileCfg() does not match")

	gv1 := orig.GetKeyFile()
	require.Equal(t, orig.KeyFile, gv1, "TLSInfo.GetKeyFileCfg() does not match")

	gv2 := orig.GetTrustedCAFile()
	require.Equal(t, orig.TrustedCAFile, gv2, "TLSInfo.GetTrustedCAFileCfg() does not match")

	gv3 := orig.GetClientCertAuth()
	require.Equal(t, orig.ClientCertAuth, &gv3, "TLSInfo.GetClientCertAuth() does not match")

}

func TestTrustyClient_overrideFrom(t *testing.T) {
	orig := TrustyClient{
		Servers: []string{"a"},
		ClientTLS: TLSInfo{
			CertFile:       "one",
			KeyFile:        "one",
			TrustedCAFile:  "one",
			ClientCertAuth: &trueVal}}
	dest := orig
	var zero TrustyClient
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "TrustyClient.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := TrustyClient{
		Servers: []string{"b", "b"},
		ClientTLS: TLSInfo{
			CertFile:       "two",
			KeyFile:        "two",
			TrustedCAFile:  "two",
			ClientCertAuth: &falseVal}}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "TrustyClient.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := TrustyClient{
		Servers: []string{"a"}}
	dest.overrideFrom(&o2)
	exp := o

	exp.Servers = o2.Servers
	require.Equal(t, dest, exp, "TrustyClient.overrideFrom should have overriden the field Servers. value now %#v, expecting %#v", dest, exp)
}

func TestTrustyClient_Getters(t *testing.T) {
	orig := TrustyClient{
		Servers: []string{"a"},
		ClientTLS: TLSInfo{
			CertFile:       "one",
			KeyFile:        "one",
			TrustedCAFile:  "one",
			ClientCertAuth: &trueVal}}

	gv0 := orig.GetServers()
	require.Equal(t, orig.Servers, gv0, "TrustyClient.GetServersCfg() does not match")

	gv1 := orig.GetClientTLSCfg()
	require.Equal(t, orig.ClientTLS, *gv1, "TrustyClient.GetClientTLSCfg() does not match")

}

func Test_LoadOverrides(t *testing.T) {

	c := Configurations{
		Defaults: Configuration{
			Region:      "two",
			Environment: "two",
			ServiceName: "two",
			ClusterName: "two",
			CryptoProv: CryptoProv{
				Default:   "two",
				Providers: []string{"b", "b"}},
			Audit: Logger{
				Directory:  "two",
				MaxAgeDays: 42,
				MaxSizeMb:  42},
			Logger: Logger{
				Directory:  "two",
				MaxAgeDays: 42,
				MaxSizeMb:  42},
			LogLevels: []RepoLogLevel{
				{
					Repo:    "two",
					Package: "two",
					Level:   "two"},
			},
			TrustyClient: TrustyClient{
				Servers: []string{"b", "b"},
				ClientTLS: TLSInfo{
					CertFile:       "two",
					KeyFile:        "two",
					TrustedCAFile:  "two",
					ClientCertAuth: &falseVal}}},
		Hosts: map[string]string{"bob": "example2", "bob2": "missing"},
		Overrides: map[string]Configuration{
			"example2": {
				Region:      "three",
				Environment: "three",
				ServiceName: "three",
				ClusterName: "three",
				CryptoProv: CryptoProv{
					Default:   "three",
					Providers: []string{"c", "c", "c"}},
				Audit: Logger{
					Directory:  "three",
					MaxAgeDays: 1234,
					MaxSizeMb:  1234},
				Logger: Logger{
					Directory:  "three",
					MaxAgeDays: 1234,
					MaxSizeMb:  1234},
				LogLevels: []RepoLogLevel{
					{
						Repo:    "three",
						Package: "three",
						Level:   "three"},
				},
				TrustyClient: TrustyClient{
					Servers: []string{"c", "c", "c"},
					ClientTLS: TLSInfo{
						CertFile:       "three",
						KeyFile:        "three",
						TrustedCAFile:  "three",
						ClientCertAuth: &trueVal}}},
		},
	}
	f, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Fatalf("Uanble to create temp file: %v", err)
	}
	json.NewEncoder(f).Encode(&c)
	f.Close()
	defer os.Remove(f.Name())
	config, err := Load(f.Name(), "", "")
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}
	require.Equal(t, c.Defaults, *config, "Loaded configuration should match default, but doesn't, expecting %#v, got %#v", c.Defaults, *config)
	config, err = Load(f.Name(), "", "bob")
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}
	require.Equal(t, c.Overrides["example2"], *config, "Loaded configuration should match default, but doesn't, expecting %#v, got %#v", c.Overrides["example2"], *config)
	_, err = Load(f.Name(), "", "bob2")
	if err == nil || err.Error() != "Configuration for host bob2 specified override set missing but that doesn't exist" {
		t.Errorf("Should of gotten error about missing override set, but got %v", err)
	}
}

func Test_LoadMissingFile(t *testing.T) {
	f, err := ioutil.TempFile("", "missing")
	f.Close()
	os.Remove(f.Name())
	_, err = Load(f.Name(), "", "")
	if !os.IsNotExist(err) {
		t.Errorf("Expecting a file doesn't exist error when trying to load from a non-existant file, but got %v", err)
	}
}

func Test_LoadInvalidJson(t *testing.T) {
	f, err := ioutil.TempFile("", "invalid")
	f.WriteString("{boom}")
	f.Close()
	defer os.Remove(f.Name())
	_, err = Load(f.Name(), "", "")
	if err == nil || err.Error() != "invalid character 'b' looking for beginning of object key string" {
		t.Errorf("Should get a json error with an invalid config file, but got %v", err)
	}
}

func loadJSONEWithENV(filename string, v interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	val := strings.ReplaceAll(string(bytes), "${ENV}", "ENV_VALUE")
	return json.NewDecoder(strings.NewReader(val)).Decode(v)
}

func Test_LoadCustomJSON(t *testing.T) {

	c := Configurations{
		Defaults: Configuration{
			Region:      "two",
			Environment: "two",
			ServiceName: "two",
			ClusterName: "two",
			CryptoProv: CryptoProv{
				Default:   "two",
				Providers: []string{"b", "b"}},
			Audit: Logger{
				Directory:  "two",
				MaxAgeDays: 42,
				MaxSizeMb:  42},
			Logger: Logger{
				Directory:  "two",
				MaxAgeDays: 42,
				MaxSizeMb:  42},
			LogLevels: []RepoLogLevel{
				{
					Repo:    "two",
					Package: "two",
					Level:   "two"},
			},
			TrustyClient: TrustyClient{
				Servers: []string{"b", "b"},
				ClientTLS: TLSInfo{
					CertFile:       "two",
					KeyFile:        "two",
					TrustedCAFile:  "two",
					ClientCertAuth: &falseVal}}},
		Hosts: map[string]string{"bob": "${ENV}"},
		Overrides: map[string]Configuration{
			"${ENV}": {
				Region:      "three",
				Environment: "three",
				ServiceName: "three",
				ClusterName: "three",
				CryptoProv: CryptoProv{
					Default:   "three",
					Providers: []string{"c", "c", "c"}},
				Audit: Logger{
					Directory:  "three",
					MaxAgeDays: 1234,
					MaxSizeMb:  1234},
				Logger: Logger{
					Directory:  "three",
					MaxAgeDays: 1234,
					MaxSizeMb:  1234},
				LogLevels: []RepoLogLevel{
					{
						Repo:    "three",
						Package: "three",
						Level:   "three"},
				},
				TrustyClient: TrustyClient{
					Servers: []string{"c", "c", "c"},
					ClientTLS: TLSInfo{
						CertFile:       "three",
						KeyFile:        "three",
						TrustedCAFile:  "three",
						ClientCertAuth: &trueVal}}},
		},
	}
	f, err := ioutil.TempFile("", "customjson")
	if err != nil {
		t.Fatalf("Uanble to create temp file: %v", err)
	}
	json.NewEncoder(f).Encode(&c)
	f.Close()
	defer os.Remove(f.Name())

	JSONLoader = loadJSONEWithENV
	config, err := Load(f.Name(), "", "")
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}
	require.Equal(t, c.Defaults, *config, "Loaded configuration should match default, but doesn't, expecting %#v, got %#v", c.Defaults, *config)

	JSONLoader = loadJSONEWithENV
	config, err = Load(f.Name(), "", "bob")
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}
	require.Equal(t, c.Overrides["${ENV}"], *config,
		"Loaded configuration should match default, but doesn't, expecting %#v, got %#v\nOverrides: %v", c.Overrides["${ENV}"], *config, c.Overrides)
}
