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

func Test_overrideDuration(t *testing.T) {
	d := Duration(time.Second)
	var zero Duration
	overrideDuration(&d, &zero)
	require.NotEqual(t, d, zero, "overrideDuration shouldn't have overriden the value as the override is the default/zero value. value now %v", d)
	o := Duration(time.Minute)
	overrideDuration(&d, &o)
	require.Equal(t, d, o, "overrideDuration should of overriden the value but didn't. value %v, expecting %v", d, o)
}

func Test_overrideHTTPServerSlice(t *testing.T) {
	d := []HTTPServer{
		{
			Name:       "one",
			Disabled:   &trueVal,
			ListenURLs: []string{"a"},
			ServerTLS: TLSInfo{
				CertFile:       "one",
				KeyFile:        "one",
				TrustedCAFile:  "one",
				CRLFile:        "one",
				OCSPFile:       "one",
				CipherSuites:   []string{"a"},
				ClientCertAuth: &trueVal},
			PackageLogger:  "one",
			AllowProfiling: &trueVal,
			ProfilerDir:    "one",
			Services:       []string{"a"},
			HeartbeatSecs:  -42,
			CORS: CORS{
				Enabled:            &trueVal,
				MaxAge:             -42,
				AllowedOrigins:     []string{"a"},
				AllowedMethods:     []string{"a"},
				AllowedHeaders:     []string{"a"},
				ExposedHeaders:     []string{"a"},
				AllowCredentials:   &trueVal,
				OptionsPassthrough: &trueVal,
				Debug:              &trueVal},
			RequestTimeout:    Duration(time.Second),
			KeepAliveMinTime:  Duration(time.Second),
			KeepAliveInterval: Duration(time.Second),
			KeepAliveTimeout:  Duration(time.Second)},
	}
	var zero []HTTPServer
	overrideHTTPServerSlice(&d, &zero)
	require.NotEqual(t, d, zero, "overrideHTTPServerSlice shouldn't have overriden the value as the override is the default/zero value. value now %v", d)
	o := []HTTPServer{
		{
			Name:       "two",
			Disabled:   &falseVal,
			ListenURLs: []string{"b", "b"},
			ServerTLS: TLSInfo{
				CertFile:       "two",
				KeyFile:        "two",
				TrustedCAFile:  "two",
				CRLFile:        "two",
				OCSPFile:       "two",
				CipherSuites:   []string{"b", "b"},
				ClientCertAuth: &falseVal},
			PackageLogger:  "two",
			AllowProfiling: &falseVal,
			ProfilerDir:    "two",
			Services:       []string{"b", "b"},
			HeartbeatSecs:  42,
			CORS: CORS{
				Enabled:            &falseVal,
				MaxAge:             42,
				AllowedOrigins:     []string{"b", "b"},
				AllowedMethods:     []string{"b", "b"},
				AllowedHeaders:     []string{"b", "b"},
				ExposedHeaders:     []string{"b", "b"},
				AllowCredentials:   &falseVal,
				OptionsPassthrough: &falseVal,
				Debug:              &falseVal},
			RequestTimeout:    Duration(time.Minute),
			KeepAliveMinTime:  Duration(time.Minute),
			KeepAliveInterval: Duration(time.Minute),
			KeepAliveTimeout:  Duration(time.Minute)},
	}
	overrideHTTPServerSlice(&d, &o)
	require.Equal(t, d, o, "overrideHTTPServerSlice should of overriden the value but didn't. value %v, expecting %v", d, o)
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

func Test_overrideIssuerSlice(t *testing.T) {
	d := []Issuer{
		{
			Disabled:       &trueVal,
			Label:          "one",
			Type:           "one",
			CertFile:       "one",
			KeyFile:        "one",
			CABundleFile:   "one",
			RootBundleFile: "one",
			CRLExpiry:      Duration(time.Second),
			OCSPExpiry:     Duration(time.Second),
			CRLRenewal:     Duration(time.Second)},
	}
	var zero []Issuer
	overrideIssuerSlice(&d, &zero)
	require.NotEqual(t, d, zero, "overrideIssuerSlice shouldn't have overriden the value as the override is the default/zero value. value now %v", d)
	o := []Issuer{
		{
			Disabled:       &falseVal,
			Label:          "two",
			Type:           "two",
			CertFile:       "two",
			KeyFile:        "two",
			CABundleFile:   "two",
			RootBundleFile: "two",
			CRLExpiry:      Duration(time.Minute),
			OCSPExpiry:     Duration(time.Minute),
			CRLRenewal:     Duration(time.Minute)},
	}
	overrideIssuerSlice(&d, &o)
	require.Equal(t, d, o, "overrideIssuerSlice should of overriden the value but didn't. value %v, expecting %v", d, o)
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

func TestAuthority_overrideFrom(t *testing.T) {
	orig := Authority{
		CAConfig:          "one",
		DefaultCRLExpiry:  Duration(time.Second),
		DefaultOCSPExpiry: Duration(time.Second),
		DefaultCRLRenewal: Duration(time.Second),
		Issuers: []Issuer{
			{
				Disabled:       &trueVal,
				Label:          "one",
				Type:           "one",
				CertFile:       "one",
				KeyFile:        "one",
				CABundleFile:   "one",
				RootBundleFile: "one",
				CRLExpiry:      Duration(time.Second),
				OCSPExpiry:     Duration(time.Second),
				CRLRenewal:     Duration(time.Second)},
		}}
	dest := orig
	var zero Authority
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "Authority.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := Authority{
		CAConfig:          "two",
		DefaultCRLExpiry:  Duration(time.Minute),
		DefaultOCSPExpiry: Duration(time.Minute),
		DefaultCRLRenewal: Duration(time.Minute),
		Issuers: []Issuer{
			{
				Disabled:       &falseVal,
				Label:          "two",
				Type:           "two",
				CertFile:       "two",
				KeyFile:        "two",
				CABundleFile:   "two",
				RootBundleFile: "two",
				CRLExpiry:      Duration(time.Minute),
				OCSPExpiry:     Duration(time.Minute),
				CRLRenewal:     Duration(time.Minute)},
		}}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "Authority.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := Authority{
		CAConfig: "one"}
	dest.overrideFrom(&o2)
	exp := o

	exp.CAConfig = o2.CAConfig
	require.Equal(t, dest, exp, "Authority.overrideFrom should have overriden the field CAConfig. value now %#v, expecting %#v", dest, exp)
}

func TestAuthz_overrideFrom(t *testing.T) {
	orig := Authz{
		Allow:         []string{"a"},
		AllowAny:      []string{"a"},
		AllowAnyRole:  []string{"a"},
		LogAllowedAny: &trueVal,
		LogAllowed:    &trueVal,
		LogDenied:     &trueVal,
		CertMapper:    "one",
		APIKeyMapper:  "one",
		JWTMapper:     "one"}
	dest := orig
	var zero Authz
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "Authz.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := Authz{
		Allow:         []string{"b", "b"},
		AllowAny:      []string{"b", "b"},
		AllowAnyRole:  []string{"b", "b"},
		LogAllowedAny: &falseVal,
		LogAllowed:    &falseVal,
		LogDenied:     &falseVal,
		CertMapper:    "two",
		APIKeyMapper:  "two",
		JWTMapper:     "two"}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "Authz.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := Authz{
		Allow: []string{"a"}}
	dest.overrideFrom(&o2)
	exp := o

	exp.Allow = o2.Allow
	require.Equal(t, dest, exp, "Authz.overrideFrom should have overriden the field Allow. value now %#v, expecting %#v", dest, exp)
}

func TestAuthz_Getters(t *testing.T) {
	orig := Authz{
		Allow:         []string{"a"},
		AllowAny:      []string{"a"},
		AllowAnyRole:  []string{"a"},
		LogAllowedAny: &trueVal,
		LogAllowed:    &trueVal,
		LogDenied:     &trueVal,
		CertMapper:    "one",
		APIKeyMapper:  "one",
		JWTMapper:     "one"}

	gv0 := orig.GetAllow()
	require.Equal(t, orig.Allow, gv0, "Authz.GetAllowCfg() does not match")

	gv1 := orig.GetAllowAny()
	require.Equal(t, orig.AllowAny, gv1, "Authz.GetAllowAnyCfg() does not match")

	gv2 := orig.GetAllowAnyRole()
	require.Equal(t, orig.AllowAnyRole, gv2, "Authz.GetAllowAnyRoleCfg() does not match")

	gv3 := orig.GetLogAllowedAny()
	require.Equal(t, orig.LogAllowedAny, &gv3, "Authz.GetLogAllowedAny() does not match")

	gv4 := orig.GetLogAllowed()
	require.Equal(t, orig.LogAllowed, &gv4, "Authz.GetLogAllowed() does not match")

	gv5 := orig.GetLogDenied()
	require.Equal(t, orig.LogDenied, &gv5, "Authz.GetLogDenied() does not match")

	gv6 := orig.GetCertMapper()
	require.Equal(t, orig.CertMapper, gv6, "Authz.GetCertMapperCfg() does not match")

	gv7 := orig.GetAPIKeyMapper()
	require.Equal(t, orig.APIKeyMapper, gv7, "Authz.GetAPIKeyMapperCfg() does not match")

	gv8 := orig.GetJWTMapper()
	require.Equal(t, orig.JWTMapper, gv8, "Authz.GetJWTMapperCfg() does not match")

}

func TestAutoGenCert_overrideFrom(t *testing.T) {
	orig := AutoGenCert{
		Disabled: &trueVal,
		CertFile: "one",
		KeyFile:  "one",
		Profile:  "one",
		Renewal:  "one",
		Schedule: "one",
		Hosts:    []string{"a"}}
	dest := orig
	var zero AutoGenCert
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "AutoGenCert.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := AutoGenCert{
		Disabled: &falseVal,
		CertFile: "two",
		KeyFile:  "two",
		Profile:  "two",
		Renewal:  "two",
		Schedule: "two",
		Hosts:    []string{"b", "b"}}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "AutoGenCert.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := AutoGenCert{
		Disabled: &trueVal}
	dest.overrideFrom(&o2)
	exp := o

	exp.Disabled = o2.Disabled
	require.Equal(t, dest, exp, "AutoGenCert.overrideFrom should have overriden the field Disabled. value now %#v, expecting %#v", dest, exp)
}

func TestAutoGenCert_Getters(t *testing.T) {
	orig := AutoGenCert{
		Disabled: &trueVal,
		CertFile: "one",
		KeyFile:  "one",
		Profile:  "one",
		Renewal:  "one",
		Schedule: "one",
		Hosts:    []string{"a"}}

	gv0 := orig.GetDisabled()
	require.Equal(t, orig.Disabled, &gv0, "AutoGenCert.GetDisabled() does not match")

	gv1 := orig.GetCertFile()
	require.Equal(t, orig.CertFile, gv1, "AutoGenCert.GetCertFileCfg() does not match")

	gv2 := orig.GetKeyFile()
	require.Equal(t, orig.KeyFile, gv2, "AutoGenCert.GetKeyFileCfg() does not match")

	gv3 := orig.GetProfile()
	require.Equal(t, orig.Profile, gv3, "AutoGenCert.GetProfileCfg() does not match")

	gv4 := orig.GetRenewal()
	require.Equal(t, orig.Renewal, gv4, "AutoGenCert.GetRenewalCfg() does not match")

	gv5 := orig.GetSchedule()
	require.Equal(t, orig.Schedule, gv5, "AutoGenCert.GetScheduleCfg() does not match")

	gv6 := orig.GetHosts()
	require.Equal(t, orig.Hosts, gv6, "AutoGenCert.GetHostsCfg() does not match")

}

func TestCORS_overrideFrom(t *testing.T) {
	orig := CORS{
		Enabled:            &trueVal,
		MaxAge:             -42,
		AllowedOrigins:     []string{"a"},
		AllowedMethods:     []string{"a"},
		AllowedHeaders:     []string{"a"},
		ExposedHeaders:     []string{"a"},
		AllowCredentials:   &trueVal,
		OptionsPassthrough: &trueVal,
		Debug:              &trueVal}
	dest := orig
	var zero CORS
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "CORS.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := CORS{
		Enabled:            &falseVal,
		MaxAge:             42,
		AllowedOrigins:     []string{"b", "b"},
		AllowedMethods:     []string{"b", "b"},
		AllowedHeaders:     []string{"b", "b"},
		ExposedHeaders:     []string{"b", "b"},
		AllowCredentials:   &falseVal,
		OptionsPassthrough: &falseVal,
		Debug:              &falseVal}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "CORS.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := CORS{
		Enabled: &trueVal}
	dest.overrideFrom(&o2)
	exp := o

	exp.Enabled = o2.Enabled
	require.Equal(t, dest, exp, "CORS.overrideFrom should have overriden the field Enabled. value now %#v, expecting %#v", dest, exp)
}

func TestCORS_Getters(t *testing.T) {
	orig := CORS{
		Enabled:            &trueVal,
		MaxAge:             -42,
		AllowedOrigins:     []string{"a"},
		AllowedMethods:     []string{"a"},
		AllowedHeaders:     []string{"a"},
		ExposedHeaders:     []string{"a"},
		AllowCredentials:   &trueVal,
		OptionsPassthrough: &trueVal,
		Debug:              &trueVal}

	gv0 := orig.GetEnabled()
	require.Equal(t, orig.Enabled, &gv0, "CORS.GetEnabled() does not match")

	gv1 := orig.GetMaxAge()
	require.Equal(t, orig.MaxAge, gv1, "CORS.GetMaxAgeCfg() does not match")

	gv2 := orig.GetAllowedOrigins()
	require.Equal(t, orig.AllowedOrigins, gv2, "CORS.GetAllowedOriginsCfg() does not match")

	gv3 := orig.GetAllowedMethods()
	require.Equal(t, orig.AllowedMethods, gv3, "CORS.GetAllowedMethodsCfg() does not match")

	gv4 := orig.GetAllowedHeaders()
	require.Equal(t, orig.AllowedHeaders, gv4, "CORS.GetAllowedHeadersCfg() does not match")

	gv5 := orig.GetExposedHeaders()
	require.Equal(t, orig.ExposedHeaders, gv5, "CORS.GetExposedHeadersCfg() does not match")

	gv6 := orig.GetAllowCredentials()
	require.Equal(t, orig.AllowCredentials, &gv6, "CORS.GetAllowCredentials() does not match")

	gv7 := orig.GetOptionsPassthrough()
	require.Equal(t, orig.OptionsPassthrough, &gv7, "CORS.GetOptionsPassthrough() does not match")

	gv8 := orig.GetDebug()
	require.Equal(t, orig.Debug, &gv8, "CORS.GetDebug() does not match")

}

func TestConfiguration_overrideFrom(t *testing.T) {
	orig := Configuration{
		Region:      "one",
		Environment: "one",
		ServiceName: "one",
		ClusterName: "one",
		CryptoProv: CryptoProv{
			Default:             "one",
			Providers:           []string{"a"},
			PKCS11Manufacturers: []string{"a"}},
		Audit: Logger{
			Directory:  "one",
			MaxAgeDays: -42,
			MaxSizeMb:  -42},
		Authz: Authz{
			Allow:         []string{"a"},
			AllowAny:      []string{"a"},
			AllowAnyRole:  []string{"a"},
			LogAllowedAny: &trueVal,
			LogAllowed:    &trueVal,
			LogDenied:     &trueVal,
			CertMapper:    "one",
			APIKeyMapper:  "one",
			JWTMapper:     "one"},
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
		HTTPServers: []HTTPServer{
			{
				Name:       "one",
				Disabled:   &trueVal,
				ListenURLs: []string{"a"},
				ServerTLS: TLSInfo{
					CertFile:       "one",
					KeyFile:        "one",
					TrustedCAFile:  "one",
					CRLFile:        "one",
					OCSPFile:       "one",
					CipherSuites:   []string{"a"},
					ClientCertAuth: &trueVal},
				PackageLogger:  "one",
				AllowProfiling: &trueVal,
				ProfilerDir:    "one",
				Services:       []string{"a"},
				HeartbeatSecs:  -42,
				CORS: CORS{
					Enabled:            &trueVal,
					MaxAge:             -42,
					AllowedOrigins:     []string{"a"},
					AllowedMethods:     []string{"a"},
					AllowedHeaders:     []string{"a"},
					ExposedHeaders:     []string{"a"},
					AllowCredentials:   &trueVal,
					OptionsPassthrough: &trueVal,
					Debug:              &trueVal},
				RequestTimeout:    Duration(time.Second),
				KeepAliveMinTime:  Duration(time.Second),
				KeepAliveInterval: Duration(time.Second),
				KeepAliveTimeout:  Duration(time.Second)},
		},
		TrustyClient: TrustyClient{
			Servers: []string{"a"},
			ClientTLS: TLSInfo{
				CertFile:       "one",
				KeyFile:        "one",
				TrustedCAFile:  "one",
				CRLFile:        "one",
				OCSPFile:       "one",
				CipherSuites:   []string{"a"},
				ClientCertAuth: &trueVal}},
		VIPs: []string{"a"},
		Authority: Authority{
			CAConfig:          "one",
			DefaultCRLExpiry:  Duration(time.Second),
			DefaultOCSPExpiry: Duration(time.Second),
			DefaultCRLRenewal: Duration(time.Second),
			Issuers: []Issuer{
				{
					Disabled:       &trueVal,
					Label:          "one",
					Type:           "one",
					CertFile:       "one",
					KeyFile:        "one",
					CABundleFile:   "one",
					RootBundleFile: "one",
					CRLExpiry:      Duration(time.Second),
					OCSPExpiry:     Duration(time.Second),
					CRLRenewal:     Duration(time.Second)},
			}}}
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
			Default:             "two",
			Providers:           []string{"b", "b"},
			PKCS11Manufacturers: []string{"b", "b"}},
		Audit: Logger{
			Directory:  "two",
			MaxAgeDays: 42,
			MaxSizeMb:  42},
		Authz: Authz{
			Allow:         []string{"b", "b"},
			AllowAny:      []string{"b", "b"},
			AllowAnyRole:  []string{"b", "b"},
			LogAllowedAny: &falseVal,
			LogAllowed:    &falseVal,
			LogDenied:     &falseVal,
			CertMapper:    "two",
			APIKeyMapper:  "two",
			JWTMapper:     "two"},
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
		HTTPServers: []HTTPServer{
			{
				Name:       "two",
				Disabled:   &falseVal,
				ListenURLs: []string{"b", "b"},
				ServerTLS: TLSInfo{
					CertFile:       "two",
					KeyFile:        "two",
					TrustedCAFile:  "two",
					CRLFile:        "two",
					OCSPFile:       "two",
					CipherSuites:   []string{"b", "b"},
					ClientCertAuth: &falseVal},
				PackageLogger:  "two",
				AllowProfiling: &falseVal,
				ProfilerDir:    "two",
				Services:       []string{"b", "b"},
				HeartbeatSecs:  42,
				CORS: CORS{
					Enabled:            &falseVal,
					MaxAge:             42,
					AllowedOrigins:     []string{"b", "b"},
					AllowedMethods:     []string{"b", "b"},
					AllowedHeaders:     []string{"b", "b"},
					ExposedHeaders:     []string{"b", "b"},
					AllowCredentials:   &falseVal,
					OptionsPassthrough: &falseVal,
					Debug:              &falseVal},
				RequestTimeout:    Duration(time.Minute),
				KeepAliveMinTime:  Duration(time.Minute),
				KeepAliveInterval: Duration(time.Minute),
				KeepAliveTimeout:  Duration(time.Minute)},
		},
		TrustyClient: TrustyClient{
			Servers: []string{"b", "b"},
			ClientTLS: TLSInfo{
				CertFile:       "two",
				KeyFile:        "two",
				TrustedCAFile:  "two",
				CRLFile:        "two",
				OCSPFile:       "two",
				CipherSuites:   []string{"b", "b"},
				ClientCertAuth: &falseVal}},
		VIPs: []string{"b", "b"},
		Authority: Authority{
			CAConfig:          "two",
			DefaultCRLExpiry:  Duration(time.Minute),
			DefaultOCSPExpiry: Duration(time.Minute),
			DefaultCRLRenewal: Duration(time.Minute),
			Issuers: []Issuer{
				{
					Disabled:       &falseVal,
					Label:          "two",
					Type:           "two",
					CertFile:       "two",
					KeyFile:        "two",
					CABundleFile:   "two",
					RootBundleFile: "two",
					CRLExpiry:      Duration(time.Minute),
					OCSPExpiry:     Duration(time.Minute),
					CRLRenewal:     Duration(time.Minute)},
			}}}
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
		Default:             "one",
		Providers:           []string{"a"},
		PKCS11Manufacturers: []string{"a"}}
	dest := orig
	var zero CryptoProv
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "CryptoProv.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := CryptoProv{
		Default:             "two",
		Providers:           []string{"b", "b"},
		PKCS11Manufacturers: []string{"b", "b"}}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "CryptoProv.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := CryptoProv{
		Default: "one"}
	dest.overrideFrom(&o2)
	exp := o

	exp.Default = o2.Default
	require.Equal(t, dest, exp, "CryptoProv.overrideFrom should have overriden the field Default. value now %#v, expecting %#v", dest, exp)
}

func TestHTTPServer_overrideFrom(t *testing.T) {
	orig := HTTPServer{
		Name:       "one",
		Disabled:   &trueVal,
		ListenURLs: []string{"a"},
		ServerTLS: TLSInfo{
			CertFile:       "one",
			KeyFile:        "one",
			TrustedCAFile:  "one",
			CRLFile:        "one",
			OCSPFile:       "one",
			CipherSuites:   []string{"a"},
			ClientCertAuth: &trueVal},
		PackageLogger:  "one",
		AllowProfiling: &trueVal,
		ProfilerDir:    "one",
		Services:       []string{"a"},
		HeartbeatSecs:  -42,
		CORS: CORS{
			Enabled:            &trueVal,
			MaxAge:             -42,
			AllowedOrigins:     []string{"a"},
			AllowedMethods:     []string{"a"},
			AllowedHeaders:     []string{"a"},
			ExposedHeaders:     []string{"a"},
			AllowCredentials:   &trueVal,
			OptionsPassthrough: &trueVal,
			Debug:              &trueVal},
		RequestTimeout:    Duration(time.Second),
		KeepAliveMinTime:  Duration(time.Second),
		KeepAliveInterval: Duration(time.Second),
		KeepAliveTimeout:  Duration(time.Second)}
	dest := orig
	var zero HTTPServer
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "HTTPServer.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := HTTPServer{
		Name:       "two",
		Disabled:   &falseVal,
		ListenURLs: []string{"b", "b"},
		ServerTLS: TLSInfo{
			CertFile:       "two",
			KeyFile:        "two",
			TrustedCAFile:  "two",
			CRLFile:        "two",
			OCSPFile:       "two",
			CipherSuites:   []string{"b", "b"},
			ClientCertAuth: &falseVal},
		PackageLogger:  "two",
		AllowProfiling: &falseVal,
		ProfilerDir:    "two",
		Services:       []string{"b", "b"},
		HeartbeatSecs:  42,
		CORS: CORS{
			Enabled:            &falseVal,
			MaxAge:             42,
			AllowedOrigins:     []string{"b", "b"},
			AllowedMethods:     []string{"b", "b"},
			AllowedHeaders:     []string{"b", "b"},
			ExposedHeaders:     []string{"b", "b"},
			AllowCredentials:   &falseVal,
			OptionsPassthrough: &falseVal,
			Debug:              &falseVal},
		RequestTimeout:    Duration(time.Minute),
		KeepAliveMinTime:  Duration(time.Minute),
		KeepAliveInterval: Duration(time.Minute),
		KeepAliveTimeout:  Duration(time.Minute)}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "HTTPServer.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := HTTPServer{
		Name: "one"}
	dest.overrideFrom(&o2)
	exp := o

	exp.Name = o2.Name
	require.Equal(t, dest, exp, "HTTPServer.overrideFrom should have overriden the field Name. value now %#v, expecting %#v", dest, exp)
}

func TestHTTPServer_Getters(t *testing.T) {
	orig := HTTPServer{
		Name:       "one",
		Disabled:   &trueVal,
		ListenURLs: []string{"a"},
		ServerTLS: TLSInfo{
			CertFile:       "one",
			KeyFile:        "one",
			TrustedCAFile:  "one",
			CRLFile:        "one",
			OCSPFile:       "one",
			CipherSuites:   []string{"a"},
			ClientCertAuth: &trueVal},
		PackageLogger:  "one",
		AllowProfiling: &trueVal,
		ProfilerDir:    "one",
		Services:       []string{"a"},
		HeartbeatSecs:  -42,
		CORS: CORS{
			Enabled:            &trueVal,
			MaxAge:             -42,
			AllowedOrigins:     []string{"a"},
			AllowedMethods:     []string{"a"},
			AllowedHeaders:     []string{"a"},
			ExposedHeaders:     []string{"a"},
			AllowCredentials:   &trueVal,
			OptionsPassthrough: &trueVal,
			Debug:              &trueVal},
		RequestTimeout:    Duration(time.Second),
		KeepAliveMinTime:  Duration(time.Second),
		KeepAliveInterval: Duration(time.Second),
		KeepAliveTimeout:  Duration(time.Second)}

	gv0 := orig.GetName()
	require.Equal(t, orig.Name, gv0, "HTTPServer.GetNameCfg() does not match")

	gv1 := orig.GetDisabled()
	require.Equal(t, orig.Disabled, &gv1, "HTTPServer.GetDisabled() does not match")

	gv2 := orig.GetListenURLs()
	require.Equal(t, orig.ListenURLs, gv2, "HTTPServer.GetListenURLsCfg() does not match")

	gv3 := orig.GetServerTLSCfg()
	require.Equal(t, orig.ServerTLS, *gv3, "HTTPServer.GetServerTLSCfg() does not match")

	gv4 := orig.GetPackageLogger()
	require.Equal(t, orig.PackageLogger, gv4, "HTTPServer.GetPackageLoggerCfg() does not match")

	gv5 := orig.GetAllowProfiling()
	require.Equal(t, orig.AllowProfiling, &gv5, "HTTPServer.GetAllowProfiling() does not match")

	gv6 := orig.GetProfilerDir()
	require.Equal(t, orig.ProfilerDir, gv6, "HTTPServer.GetProfilerDirCfg() does not match")

	gv7 := orig.GetServices()
	require.Equal(t, orig.Services, gv7, "HTTPServer.GetServicesCfg() does not match")

	gv8 := orig.GetHeartbeatSecs()
	require.Equal(t, orig.HeartbeatSecs, gv8, "HTTPServer.GetHeartbeatSecsCfg() does not match")

	gv9 := orig.GetCORSCfg()
	require.Equal(t, orig.CORS, *gv9, "HTTPServer.GetCORSCfg() does not match")

	gv10 := orig.GetRequestTimeout()
	require.Equal(t, orig.RequestTimeout.TimeDuration(), gv10, "HTTPServer.GetRequestTimeout() does not match")

	gv11 := orig.GetKeepAliveMinTime()
	require.Equal(t, orig.KeepAliveMinTime.TimeDuration(), gv11, "HTTPServer.GetKeepAliveMinTime() does not match")

	gv12 := orig.GetKeepAliveInterval()
	require.Equal(t, orig.KeepAliveInterval.TimeDuration(), gv12, "HTTPServer.GetKeepAliveInterval() does not match")

	gv13 := orig.GetKeepAliveTimeout()
	require.Equal(t, orig.KeepAliveTimeout.TimeDuration(), gv13, "HTTPServer.GetKeepAliveTimeout() does not match")

}

func TestIssuer_overrideFrom(t *testing.T) {
	orig := Issuer{
		Disabled:       &trueVal,
		Label:          "one",
		Type:           "one",
		CertFile:       "one",
		KeyFile:        "one",
		CABundleFile:   "one",
		RootBundleFile: "one",
		CRLExpiry:      Duration(time.Second),
		OCSPExpiry:     Duration(time.Second),
		CRLRenewal:     Duration(time.Second)}
	dest := orig
	var zero Issuer
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "Issuer.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := Issuer{
		Disabled:       &falseVal,
		Label:          "two",
		Type:           "two",
		CertFile:       "two",
		KeyFile:        "two",
		CABundleFile:   "two",
		RootBundleFile: "two",
		CRLExpiry:      Duration(time.Minute),
		OCSPExpiry:     Duration(time.Minute),
		CRLRenewal:     Duration(time.Minute)}
	dest.overrideFrom(&o)
	require.Equal(t, dest, o, "Issuer.overrideFrom should have overriden the value as the override. value now %#v, expecting %#v", dest, o)
	o2 := Issuer{
		Disabled: &trueVal}
	dest.overrideFrom(&o2)
	exp := o

	exp.Disabled = o2.Disabled
	require.Equal(t, dest, exp, "Issuer.overrideFrom should have overriden the field Disabled. value now %#v, expecting %#v", dest, exp)
}

func TestIssuer_Getters(t *testing.T) {
	orig := Issuer{
		Disabled:       &trueVal,
		Label:          "one",
		Type:           "one",
		CertFile:       "one",
		KeyFile:        "one",
		CABundleFile:   "one",
		RootBundleFile: "one",
		CRLExpiry:      Duration(time.Second),
		OCSPExpiry:     Duration(time.Second),
		CRLRenewal:     Duration(time.Second)}

	gv0 := orig.GetDisabled()
	require.Equal(t, orig.Disabled, &gv0, "Issuer.GetDisabled() does not match")

	gv1 := orig.GetLabel()
	require.Equal(t, orig.Label, gv1, "Issuer.GetLabelCfg() does not match")

	gv2 := orig.GetType()
	require.Equal(t, orig.Type, gv2, "Issuer.GetTypeCfg() does not match")

	gv3 := orig.GetCertFile()
	require.Equal(t, orig.CertFile, gv3, "Issuer.GetCertFileCfg() does not match")

	gv4 := orig.GetKeyFile()
	require.Equal(t, orig.KeyFile, gv4, "Issuer.GetKeyFileCfg() does not match")

	gv5 := orig.GetCABundleFile()
	require.Equal(t, orig.CABundleFile, gv5, "Issuer.GetCABundleFileCfg() does not match")

	gv6 := orig.GetRootBundleFile()
	require.Equal(t, orig.RootBundleFile, gv6, "Issuer.GetRootBundleFileCfg() does not match")

	gv7 := orig.GetCRLExpiry()
	require.Equal(t, orig.CRLExpiry.TimeDuration(), gv7, "Issuer.GetCRLExpiry() does not match")

	gv8 := orig.GetOCSPExpiry()
	require.Equal(t, orig.OCSPExpiry.TimeDuration(), gv8, "Issuer.GetOCSPExpiry() does not match")

	gv9 := orig.GetCRLRenewal()
	require.Equal(t, orig.CRLRenewal.TimeDuration(), gv9, "Issuer.GetCRLRenewal() does not match")

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
		CRLFile:        "one",
		OCSPFile:       "one",
		CipherSuites:   []string{"a"},
		ClientCertAuth: &trueVal}
	dest := orig
	var zero TLSInfo
	dest.overrideFrom(&zero)
	require.Equal(t, dest, orig, "TLSInfo.overrideFrom shouldn't have overriden the value as the override is the default/zero value. value now %#v", dest)
	o := TLSInfo{
		CertFile:       "two",
		KeyFile:        "two",
		TrustedCAFile:  "two",
		CRLFile:        "two",
		OCSPFile:       "two",
		CipherSuites:   []string{"b", "b"},
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
		CRLFile:        "one",
		OCSPFile:       "one",
		CipherSuites:   []string{"a"},
		ClientCertAuth: &trueVal}

	gv0 := orig.GetCertFile()
	require.Equal(t, orig.CertFile, gv0, "TLSInfo.GetCertFileCfg() does not match")

	gv1 := orig.GetKeyFile()
	require.Equal(t, orig.KeyFile, gv1, "TLSInfo.GetKeyFileCfg() does not match")

	gv2 := orig.GetTrustedCAFile()
	require.Equal(t, orig.TrustedCAFile, gv2, "TLSInfo.GetTrustedCAFileCfg() does not match")

	gv3 := orig.GetCRLFile()
	require.Equal(t, orig.CRLFile, gv3, "TLSInfo.GetCRLFileCfg() does not match")

	gv4 := orig.GetOCSPFile()
	require.Equal(t, orig.OCSPFile, gv4, "TLSInfo.GetOCSPFileCfg() does not match")

	gv5 := orig.GetCipherSuites()
	require.Equal(t, orig.CipherSuites, gv5, "TLSInfo.GetCipherSuitesCfg() does not match")

	gv6 := orig.GetClientCertAuth()
	require.Equal(t, orig.ClientCertAuth, &gv6, "TLSInfo.GetClientCertAuth() does not match")

}

func TestTrustyClient_overrideFrom(t *testing.T) {
	orig := TrustyClient{
		Servers: []string{"a"},
		ClientTLS: TLSInfo{
			CertFile:       "one",
			KeyFile:        "one",
			TrustedCAFile:  "one",
			CRLFile:        "one",
			OCSPFile:       "one",
			CipherSuites:   []string{"a"},
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
			CRLFile:        "two",
			OCSPFile:       "two",
			CipherSuites:   []string{"b", "b"},
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
			CRLFile:        "one",
			OCSPFile:       "one",
			CipherSuites:   []string{"a"},
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
				Default:             "two",
				Providers:           []string{"b", "b"},
				PKCS11Manufacturers: []string{"b", "b"}},
			Audit: Logger{
				Directory:  "two",
				MaxAgeDays: 42,
				MaxSizeMb:  42},
			Authz: Authz{
				Allow:         []string{"b", "b"},
				AllowAny:      []string{"b", "b"},
				AllowAnyRole:  []string{"b", "b"},
				LogAllowedAny: &falseVal,
				LogAllowed:    &falseVal,
				LogDenied:     &falseVal,
				CertMapper:    "two",
				APIKeyMapper:  "two",
				JWTMapper:     "two"},
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
			HTTPServers: []HTTPServer{
				{
					Name:       "two",
					Disabled:   &falseVal,
					ListenURLs: []string{"b", "b"},
					ServerTLS: TLSInfo{
						CertFile:       "two",
						KeyFile:        "two",
						TrustedCAFile:  "two",
						CRLFile:        "two",
						OCSPFile:       "two",
						CipherSuites:   []string{"b", "b"},
						ClientCertAuth: &falseVal},
					PackageLogger:  "two",
					AllowProfiling: &falseVal,
					ProfilerDir:    "two",
					Services:       []string{"b", "b"},
					HeartbeatSecs:  42,
					CORS: CORS{
						Enabled:            &falseVal,
						MaxAge:             42,
						AllowedOrigins:     []string{"b", "b"},
						AllowedMethods:     []string{"b", "b"},
						AllowedHeaders:     []string{"b", "b"},
						ExposedHeaders:     []string{"b", "b"},
						AllowCredentials:   &falseVal,
						OptionsPassthrough: &falseVal,
						Debug:              &falseVal},
					RequestTimeout:    Duration(time.Minute),
					KeepAliveMinTime:  Duration(time.Minute),
					KeepAliveInterval: Duration(time.Minute),
					KeepAliveTimeout:  Duration(time.Minute)},
			},
			TrustyClient: TrustyClient{
				Servers: []string{"b", "b"},
				ClientTLS: TLSInfo{
					CertFile:       "two",
					KeyFile:        "two",
					TrustedCAFile:  "two",
					CRLFile:        "two",
					OCSPFile:       "two",
					CipherSuites:   []string{"b", "b"},
					ClientCertAuth: &falseVal}},
			VIPs: []string{"b", "b"},
			Authority: Authority{
				CAConfig:          "two",
				DefaultCRLExpiry:  Duration(time.Minute),
				DefaultOCSPExpiry: Duration(time.Minute),
				DefaultCRLRenewal: Duration(time.Minute),
				Issuers: []Issuer{
					{
						Disabled:       &falseVal,
						Label:          "two",
						Type:           "two",
						CertFile:       "two",
						KeyFile:        "two",
						CABundleFile:   "two",
						RootBundleFile: "two",
						CRLExpiry:      Duration(time.Minute),
						OCSPExpiry:     Duration(time.Minute),
						CRLRenewal:     Duration(time.Minute)},
				}}},
		Hosts: map[string]string{"bob": "example2", "bob2": "missing"},
		Overrides: map[string]Configuration{
			"example2": {
				Region:      "three",
				Environment: "three",
				ServiceName: "three",
				ClusterName: "three",
				CryptoProv: CryptoProv{
					Default:             "three",
					Providers:           []string{"c", "c", "c"},
					PKCS11Manufacturers: []string{"c", "c", "c"}},
				Audit: Logger{
					Directory:  "three",
					MaxAgeDays: 1234,
					MaxSizeMb:  1234},
				Authz: Authz{
					Allow:         []string{"c", "c", "c"},
					AllowAny:      []string{"c", "c", "c"},
					AllowAnyRole:  []string{"c", "c", "c"},
					LogAllowedAny: &trueVal,
					LogAllowed:    &trueVal,
					LogDenied:     &trueVal,
					CertMapper:    "three",
					APIKeyMapper:  "three",
					JWTMapper:     "three"},
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
				HTTPServers: []HTTPServer{
					{
						Name:       "three",
						Disabled:   &trueVal,
						ListenURLs: []string{"c", "c", "c"},
						ServerTLS: TLSInfo{
							CertFile:       "three",
							KeyFile:        "three",
							TrustedCAFile:  "three",
							CRLFile:        "three",
							OCSPFile:       "three",
							CipherSuites:   []string{"c", "c", "c"},
							ClientCertAuth: &trueVal},
						PackageLogger:  "three",
						AllowProfiling: &trueVal,
						ProfilerDir:    "three",
						Services:       []string{"c", "c", "c"},
						HeartbeatSecs:  1234,
						CORS: CORS{
							Enabled:            &trueVal,
							MaxAge:             1234,
							AllowedOrigins:     []string{"c", "c", "c"},
							AllowedMethods:     []string{"c", "c", "c"},
							AllowedHeaders:     []string{"c", "c", "c"},
							ExposedHeaders:     []string{"c", "c", "c"},
							AllowCredentials:   &trueVal,
							OptionsPassthrough: &trueVal,
							Debug:              &trueVal},
						RequestTimeout:    Duration(time.Hour),
						KeepAliveMinTime:  Duration(time.Hour),
						KeepAliveInterval: Duration(time.Hour),
						KeepAliveTimeout:  Duration(time.Hour)},
				},
				TrustyClient: TrustyClient{
					Servers: []string{"c", "c", "c"},
					ClientTLS: TLSInfo{
						CertFile:       "three",
						KeyFile:        "three",
						TrustedCAFile:  "three",
						CRLFile:        "three",
						OCSPFile:       "three",
						CipherSuites:   []string{"c", "c", "c"},
						ClientCertAuth: &trueVal}},
				VIPs: []string{"c", "c", "c"},
				Authority: Authority{
					CAConfig:          "three",
					DefaultCRLExpiry:  Duration(time.Hour),
					DefaultOCSPExpiry: Duration(time.Hour),
					DefaultCRLRenewal: Duration(time.Hour),
					Issuers: []Issuer{
						{
							Disabled:       &trueVal,
							Label:          "three",
							Type:           "three",
							CertFile:       "three",
							KeyFile:        "three",
							CABundleFile:   "three",
							RootBundleFile: "three",
							CRLExpiry:      Duration(time.Hour),
							OCSPExpiry:     Duration(time.Hour),
							CRLRenewal:     Duration(time.Hour)},
					}}},
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
				Default:             "two",
				Providers:           []string{"b", "b"},
				PKCS11Manufacturers: []string{"b", "b"}},
			Audit: Logger{
				Directory:  "two",
				MaxAgeDays: 42,
				MaxSizeMb:  42},
			Authz: Authz{
				Allow:         []string{"b", "b"},
				AllowAny:      []string{"b", "b"},
				AllowAnyRole:  []string{"b", "b"},
				LogAllowedAny: &falseVal,
				LogAllowed:    &falseVal,
				LogDenied:     &falseVal,
				CertMapper:    "two",
				APIKeyMapper:  "two",
				JWTMapper:     "two"},
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
			HTTPServers: []HTTPServer{
				{
					Name:       "two",
					Disabled:   &falseVal,
					ListenURLs: []string{"b", "b"},
					ServerTLS: TLSInfo{
						CertFile:       "two",
						KeyFile:        "two",
						TrustedCAFile:  "two",
						CRLFile:        "two",
						OCSPFile:       "two",
						CipherSuites:   []string{"b", "b"},
						ClientCertAuth: &falseVal},
					PackageLogger:  "two",
					AllowProfiling: &falseVal,
					ProfilerDir:    "two",
					Services:       []string{"b", "b"},
					HeartbeatSecs:  42,
					CORS: CORS{
						Enabled:            &falseVal,
						MaxAge:             42,
						AllowedOrigins:     []string{"b", "b"},
						AllowedMethods:     []string{"b", "b"},
						AllowedHeaders:     []string{"b", "b"},
						ExposedHeaders:     []string{"b", "b"},
						AllowCredentials:   &falseVal,
						OptionsPassthrough: &falseVal,
						Debug:              &falseVal},
					RequestTimeout:    Duration(time.Minute),
					KeepAliveMinTime:  Duration(time.Minute),
					KeepAliveInterval: Duration(time.Minute),
					KeepAliveTimeout:  Duration(time.Minute)},
			},
			TrustyClient: TrustyClient{
				Servers: []string{"b", "b"},
				ClientTLS: TLSInfo{
					CertFile:       "two",
					KeyFile:        "two",
					TrustedCAFile:  "two",
					CRLFile:        "two",
					OCSPFile:       "two",
					CipherSuites:   []string{"b", "b"},
					ClientCertAuth: &falseVal}},
			VIPs: []string{"b", "b"},
			Authority: Authority{
				CAConfig:          "two",
				DefaultCRLExpiry:  Duration(time.Minute),
				DefaultOCSPExpiry: Duration(time.Minute),
				DefaultCRLRenewal: Duration(time.Minute),
				Issuers: []Issuer{
					{
						Disabled:       &falseVal,
						Label:          "two",
						Type:           "two",
						CertFile:       "two",
						KeyFile:        "two",
						CABundleFile:   "two",
						RootBundleFile: "two",
						CRLExpiry:      Duration(time.Minute),
						OCSPExpiry:     Duration(time.Minute),
						CRLRenewal:     Duration(time.Minute)},
				}}},
		Hosts: map[string]string{"bob": "${ENV}"},
		Overrides: map[string]Configuration{
			"${ENV}": {
				Region:      "three",
				Environment: "three",
				ServiceName: "three",
				ClusterName: "three",
				CryptoProv: CryptoProv{
					Default:             "three",
					Providers:           []string{"c", "c", "c"},
					PKCS11Manufacturers: []string{"c", "c", "c"}},
				Audit: Logger{
					Directory:  "three",
					MaxAgeDays: 1234,
					MaxSizeMb:  1234},
				Authz: Authz{
					Allow:         []string{"c", "c", "c"},
					AllowAny:      []string{"c", "c", "c"},
					AllowAnyRole:  []string{"c", "c", "c"},
					LogAllowedAny: &trueVal,
					LogAllowed:    &trueVal,
					LogDenied:     &trueVal,
					CertMapper:    "three",
					APIKeyMapper:  "three",
					JWTMapper:     "three"},
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
				HTTPServers: []HTTPServer{
					{
						Name:       "three",
						Disabled:   &trueVal,
						ListenURLs: []string{"c", "c", "c"},
						ServerTLS: TLSInfo{
							CertFile:       "three",
							KeyFile:        "three",
							TrustedCAFile:  "three",
							CRLFile:        "three",
							OCSPFile:       "three",
							CipherSuites:   []string{"c", "c", "c"},
							ClientCertAuth: &trueVal},
						PackageLogger:  "three",
						AllowProfiling: &trueVal,
						ProfilerDir:    "three",
						Services:       []string{"c", "c", "c"},
						HeartbeatSecs:  1234,
						CORS: CORS{
							Enabled:            &trueVal,
							MaxAge:             1234,
							AllowedOrigins:     []string{"c", "c", "c"},
							AllowedMethods:     []string{"c", "c", "c"},
							AllowedHeaders:     []string{"c", "c", "c"},
							ExposedHeaders:     []string{"c", "c", "c"},
							AllowCredentials:   &trueVal,
							OptionsPassthrough: &trueVal,
							Debug:              &trueVal},
						RequestTimeout:    Duration(time.Hour),
						KeepAliveMinTime:  Duration(time.Hour),
						KeepAliveInterval: Duration(time.Hour),
						KeepAliveTimeout:  Duration(time.Hour)},
				},
				TrustyClient: TrustyClient{
					Servers: []string{"c", "c", "c"},
					ClientTLS: TLSInfo{
						CertFile:       "three",
						KeyFile:        "three",
						TrustedCAFile:  "three",
						CRLFile:        "three",
						OCSPFile:       "three",
						CipherSuites:   []string{"c", "c", "c"},
						ClientCertAuth: &trueVal}},
				VIPs: []string{"c", "c", "c"},
				Authority: Authority{
					CAConfig:          "three",
					DefaultCRLExpiry:  Duration(time.Hour),
					DefaultOCSPExpiry: Duration(time.Hour),
					DefaultCRLRenewal: Duration(time.Hour),
					Issuers: []Issuer{
						{
							Disabled:       &trueVal,
							Label:          "three",
							Type:           "three",
							CertFile:       "three",
							KeyFile:        "three",
							CABundleFile:   "three",
							RootBundleFile: "three",
							CRLExpiry:      Duration(time.Hour),
							OCSPExpiry:     Duration(time.Hour),
							CRLRenewal:     Duration(time.Hour)},
					}}},
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
