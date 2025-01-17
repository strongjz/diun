package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/diun/v4/internal/config"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/utl"
	"github.com/crazy-max/gonfig/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFile(t *testing.T) {
	cases := []struct {
		name     string
		cli      model.Cli
		wantData *config.Config
		wantErr  bool
	}{
		{
			name: "Failed on non-existing file",
			cli: model.Cli{
				TestNotif: false,
			},
			wantErr: true,
		},
		{
			name: "Fail on wrong file format",
			cli: model.Cli{
				Cfgfile:   "./fixtures/config.invalid.yml",
				TestNotif: false,
			},
			wantErr: true,
		},
		{
			name: "Success",
			cli: model.Cli{
				Cfgfile:   "./fixtures/config.test.yml",
				TestNotif: false,
			},
			wantData: &config.Config{
				Db: &model.Db{
					Path: "diun.db",
				},
				Watch: &model.Watch{
					Workers:         100,
					Schedule:        "*/30 * * * *",
					FirstCheckNotif: utl.NewTrue(),
				},
				Notif: &model.Notif{
					Amqp: &model.NotifAmqp{
						Host:     "localhost",
						Port:     5672,
						Username: "guest",
						Password: "guest",
						Queue:    "queue",
					},
					Discord: &model.NotifDiscord{
						WebhookURL: "https://discordapp.com/api/webhooks/1234567890/Abcd-eFgh-iJklmNo_pqr",
						Timeout:    utl.NewDuration(10 * time.Second),
					},
					Gotify: &model.NotifGotify{
						Endpoint: "http://gotify.foo.com",
						Token:    "Token123456",
						Priority: 1,
						Timeout:  utl.NewDuration(10 * time.Second),
					},
					Mail: &model.NotifMail{
						Host:               "localhost",
						Port:               25,
						SSL:                utl.NewFalse(),
						InsecureSkipVerify: utl.NewFalse(),
						LocalName:          "localhost",
						From:               "diun@example.com",
						To:                 "webmaster@example.com",
					},
					Matrix: &model.NotifMatrix{
						HomeserverURL: "https://matrix.org",
						User:          "@foo:matrix.org",
						Password:      "bar",
						RoomID:        "!abcdefGHIjklmno:matrix.org",
						MsgType:       model.NotifMatrixMsgTypeNotice,
					},
					RocketChat: &model.NotifRocketChat{
						Endpoint: "http://rocket.foo.com:3000",
						Channel:  "#general",
						UserID:   "abcdEFGH012345678",
						Token:    "Token123456",
						Timeout:  utl.NewDuration(10 * time.Second),
					},
					Script: &model.NotifScript{
						Cmd: "uname",
						Args: []string{
							"-a",
						},
					},
					Slack: &model.NotifSlack{
						WebhookURL: "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
					},
					Teams: &model.NotifTeams{
						WebhookURL: "https://outlook.office.com/webhook/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
					},
					Telegram: &model.NotifTelegram{
						Token:   "abcdef123456",
						ChatIDs: []int64{8547439, 1234567},
					},
					Webhook: &model.NotifWebhook{
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Method:   "GET",
						Headers: map[string]string{
							"content-type":  "application/json",
							"authorization": "Token123456",
						},
						Timeout: utl.NewDuration(10 * time.Second),
					},
				},
				RegOpts: model.RegOpts{
					{
						Name:        "myregistry",
						Selector:    model.RegOptSelectorName,
						Username:    "fii",
						Password:    "bor",
						InsecureTLS: utl.NewFalse(),
						Timeout:     utl.NewDuration(5 * time.Second),
					},
					{
						Name:        "docker.io",
						Selector:    model.RegOptSelectorImage,
						Username:    "foo",
						Password:    "bar",
						InsecureTLS: utl.NewFalse(),
						Timeout:     utl.NewDuration(10 * time.Second),
					},
					{
						Name:         "docker.io/crazymax",
						Selector:     model.RegOptSelectorImage,
						UsernameFile: "./fixtures/run_secrets_username",
						PasswordFile: "./fixtures/run_secrets_password",
						InsecureTLS:  utl.NewFalse(),
						Timeout:      utl.NewDuration(10 * time.Second),
					},
				},
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      utl.NewTrue(),
						WatchByDefault: utl.NewTrue(),
						WatchStopped:   utl.NewTrue(),
					},
					Swarm: &model.PrdSwarm{
						TLSVerify:      utl.NewTrue(),
						WatchByDefault: utl.NewFalse(),
					},
					Kubernetes: &model.PrdKubernetes{
						TLSInsecure:    utl.NewFalse(),
						WatchByDefault: utl.NewTrue(),
					},
					File: &model.PrdFile{
						Filename: "./fixtures/file.yml",
					},
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(tt.cli)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantData, cfg)
			if cfg != nil {
				assert.NotEmpty(t, cfg.String())
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	defer UnsetEnv("DIUN_")

	testCases := []struct {
		desc     string
		cli      model.Cli
		environ  []string
		expected interface{}
		wantErr  bool
	}{
		{
			desc:     "no env vars",
			environ:  nil,
			expected: nil,
			wantErr:  true,
		},
		{
			desc: "docker provider",
			environ: []string{
				"DIUN_PROVIDERS_DOCKER=true",
			},
			expected: &config.Config{
				Db:      (&model.Db{}).GetDefaults(),
				Watch:   (&model.Watch{}).GetDefaults(),
				Notif:   nil,
				RegOpts: nil,
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      utl.NewTrue(),
						WatchByDefault: utl.NewFalse(),
						WatchStopped:   utl.NewFalse(),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "docker provider and regopts",
			environ: []string{
				"DIUN_REGOPTS_0_NAME=docker.io",
				"DIUN_REGOPTS_0_SELECTOR=image",
				"DIUN_REGOPTS_0_USERNAMEFILE=./fixtures/run_secrets_username",
				"DIUN_REGOPTS_0_PASSWORDFILE=./fixtures/run_secrets_password",
				"DIUN_REGOPTS_0_TIMEOUT=30s",
				"DIUN_PROVIDERS_DOCKER=true",
			},
			expected: &config.Config{
				Db:    (&model.Db{}).GetDefaults(),
				Watch: (&model.Watch{}).GetDefaults(),
				RegOpts: model.RegOpts{
					{
						Name:         "docker.io",
						Selector:     model.RegOptSelectorImage,
						UsernameFile: "./fixtures/run_secrets_username",
						PasswordFile: "./fixtures/run_secrets_password",
						InsecureTLS:  utl.NewFalse(),
						Timeout:      utl.NewDuration(30 * time.Second),
					},
				},
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      utl.NewTrue(),
						WatchByDefault: utl.NewFalse(),
						WatchStopped:   utl.NewFalse(),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "swarm provider and notif telegram",
			environ: []string{
				"DIUN_NOTIF_TELEGRAM_TOKEN=abcdef123456",
				"DIUN_NOTIF_TELEGRAM_CHATIDS=8547439,1234567",
				"DIUN_PROVIDERS_SWARM=true",
			},
			expected: &config.Config{
				Db:    (&model.Db{}).GetDefaults(),
				Watch: (&model.Watch{}).GetDefaults(),
				Notif: &model.Notif{
					Telegram: &model.NotifTelegram{
						Token:   "abcdef123456",
						ChatIDs: []int64{8547439, 1234567},
					},
				},
				Providers: &model.Providers{
					Swarm: &model.PrdSwarm{
						TLSVerify:      utl.NewTrue(),
						WatchByDefault: utl.NewFalse(),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "file provider and notif script",
			environ: []string{
				"DIUN_NOTIF_SCRIPT_CMD=uname",
				"DIUN_NOTIF_SCRIPT_ARGS=-a",
				"DIUN_PROVIDERS_FILE_DIRECTORY=./fixtures",
			},
			expected: &config.Config{
				Db:    (&model.Db{}).GetDefaults(),
				Watch: (&model.Watch{}).GetDefaults(),
				Notif: &model.Notif{
					Script: &model.NotifScript{
						Cmd: "uname",
						Args: []string{
							"-a",
						},
					},
				},
				Providers: &model.Providers{
					File: &model.PrdFile{
						Directory: "./fixtures",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			UnsetEnv("DIUN_")

			if tt.environ != nil {
				for _, environ := range tt.environ {
					n := strings.SplitN(environ, "=", 2)
					os.Setenv(n[0], n[1])
				}
			}

			cfg, err := config.Load(tt.cli)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestLoadMixed(t *testing.T) {
	defer UnsetEnv("DIUN_")

	testCases := []struct {
		desc     string
		cli      model.Cli
		environ  []string
		expected interface{}
		wantErr  bool
	}{
		{
			desc: "env vars and invalid file",
			cli: model.Cli{
				Cfgfile:   "./fixtures/config.invalid.yml",
				TestNotif: false,
			},
			environ: []string{
				"DIUN_PROVIDERS_DOCKER=true",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			desc: "docker provider (file) and notif mails (envs)",
			cli: model.Cli{
				Cfgfile:   "./fixtures/config.docker.yml",
				TestNotif: false,
			},
			environ: []string{
				"DIUN_NOTIF_MAIL_HOST=127.0.0.1",
				"DIUN_NOTIF_MAIL_PORT=25",
				"DIUN_NOTIF_MAIL_SSL=false",
				"DIUN_NOTIF_MAIL_INSECURESKIPVERIFY=true",
				"DIUN_NOTIF_MAIL_FROM=diun@foo.com",
				"DIUN_NOTIF_MAIL_TO=webmaster@foo.com",
				"DIUN_NOTIF_MAIL_LOCALNAME=foo.com",
			},
			expected: &config.Config{
				Db:    (&model.Db{}).GetDefaults(),
				Watch: (&model.Watch{}).GetDefaults(),
				Notif: &model.Notif{
					Mail: &model.NotifMail{
						Host:               "127.0.0.1",
						Port:               25,
						SSL:                utl.NewFalse(),
						InsecureSkipVerify: utl.NewTrue(),
						LocalName:          "foo.com",
						From:               "diun@foo.com",
						To:                 "webmaster@foo.com",
					},
				},
				RegOpts: nil,
				Providers: &model.Providers{
					Docker: &model.PrdDocker{
						TLSVerify:      utl.NewTrue(),
						WatchByDefault: utl.NewFalse(),
						WatchStopped:   utl.NewFalse(),
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "file provider and notif webhook env override",
			cli: model.Cli{
				Cfgfile:   "./fixtures/config.file.yml",
				TestNotif: false,
			},
			environ: []string{
				"DIUN_NOTIF_WEBHOOK_ENDPOINT=http://webhook.foo.com/sd54qad89azd5a",
				"DIUN_NOTIF_WEBHOOK_HEADERS_AUTHORIZATION=Token78910",
				"DIUN_NOTIF_WEBHOOK_HEADERS_CONTENT-TYPE=text/plain",
				"DIUN_NOTIF_WEBHOOK_METHOD=GET",
				"DIUN_NOTIF_WEBHOOK_TIMEOUT=1m",
			},
			expected: &config.Config{
				Db:    (&model.Db{}).GetDefaults(),
				Watch: (&model.Watch{}).GetDefaults(),
				Notif: &model.Notif{
					Webhook: &model.NotifWebhook{
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Method:   "GET",
						Headers: map[string]string{
							"content-type":  "text/plain",
							"authorization": "Token78910",
						},
						Timeout: utl.NewDuration(1 * time.Minute),
					},
				},
				RegOpts: nil,
				Providers: &model.Providers{
					File: &model.PrdFile{
						Filename: "./fixtures/file.yml",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			UnsetEnv("DIUN_")

			if tt.environ != nil {
				for _, environ := range tt.environ {
					n := strings.SplitN(environ, "=", 2)
					os.Setenv(n[0], n[1])
				}
			}

			cfg, err := config.Load(tt.cli)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestValidation(t *testing.T) {
	cases := []struct {
		name string
		cli  model.Cli
	}{
		{
			name: "Success",
			cli: model.Cli{
				Cfgfile:   "./fixtures/config.validate.yml",
				TestNotif: false,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(tt.cli)
			require.NoError(t, err)

			dec, err := env.Encode(cfg)
			require.NoError(t, err)
			for _, value := range dec {
				fmt.Println(fmt.Sprintf(`%s=%s`, strings.Replace(value.Name, "GONFIG_", "DIUN_", 1), value.Default))
			}
		})
	}
}

func UnsetEnv(prefix string) (restore func()) {
	before := map[string]string{}

	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}

		parts := strings.SplitN(e, "=", 2)
		before[parts[0]] = parts[1]

		os.Unsetenv(parts[0])
	}

	return func() {
		after := map[string]string{}

		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, prefix) {
				continue
			}

			parts := strings.SplitN(e, "=", 2)
			after[parts[0]] = parts[1]

			// Check if the envar previously existed
			v, ok := before[parts[0]]
			if !ok {
				// This is a newly added envar with prefix, zap it
				os.Unsetenv(parts[0])
				continue
			}

			if parts[1] != v {
				// If the envar value has changed, set it back
				os.Setenv(parts[0], v)
			}
		}

		// Still need to check if there have been any deleted envars
		for k, v := range before {
			if _, ok := after[k]; !ok {
				// k is not present in after, so we set it.
				os.Setenv(k, v)
			}
		}
	}
}
