package settings

import (
	"time"

	"github.com/caarlos0/env"
	"github.com/jinzhu/now"

	"github.com/Syncano/orion/pkg/util"
)

type common struct {
	Location          string        `env:"LOCATION"`
	Locations         []string      `env:"LOCATIONS"`
	MainLocation      bool          // computed
	Debug             bool          `env:"DEBUG"`
	CacheVersion      int           `env:"CACHE_VERSION"`
	CacheTimeout      time.Duration `env:"CACHE_TIMEOUT"`
	LocalCacheTimeout time.Duration `env:"LOCAL_CACHE_TIMEOUT"`
	DateFormat        string        `env:"DATE_FORMAT"`
	DateTimeFormat    string        `env:"DATETIME_FORMAT"`

	AnalyticsWriteKey string `env:"ANALYTICS_WRITE_KEY"`
	SecretKey         string `env:"SECRET_KEY"`
	StripeSecretKey   string `env:"STRIPE_SECRET_KEY"`
}

// Common is a global struct with options filled by env.
var Common = &common{
	Location:          "stg",
	Locations:         []string{"stg"},
	Debug:             false,
	CacheVersion:      1,
	CacheTimeout:      12 * time.Hour,
	LocalCacheTimeout: 1 * time.Hour,
	DateFormat:        "2006-01-02",
	DateTimeFormat:    "2006-01-02T15:04:05.000000Z",

	SecretKey: "secret_key",
}

type social struct {
	GithubClientID       string `env:"GITHUB_CLIENT_ID"`
	GithubClientSecret   string `env:"GITHUB_CLIENT_SECRET"`
	LinkedinClientID     string `env:"LINKEDIN_CLIENT_ID"`
	LinkedinClientSecret string `env:"LINKEDIN_CLIENT_SECRET"`
	TwitterClientID      string `env:"TWITTER_CLIENT_ID"`
	TwitterClientSecret  string `env:"TWITTER_CLIENT_SECRET"`
}

var Social = &social{}

type PlanLimit struct {
	Default int
	Paid    int
	Builder int
}

type billing struct {
	DefaultPlanName            string `env:"BILLING_DEFAULT_PLAN_NAME"`
	DefaultPlanExpiration      int    `env:"BILLING_DEFAULT_PLAN_EXPIRATION"` // days
	DueDays                    int    // days
	GracePeriodForPlanChanging int
	AlarmPoints                []int
	ChecksTimeout              time.Duration

	StorageLimit       PlanLimit
	RateLimit          PlanLimit
	PollRateLimit      PlanLimit
	CodeboxConcurrency PlanLimit
	InstancesCount     PlanLimit
	ClassesCount       PlanLimit
	SocketsCount       PlanLimit
	SchedulesCount     PlanLimit
}

var Billing = &billing{
	DefaultPlanName:            "builder",
	DefaultPlanExpiration:      30,
	DueDays:                    30,
	GracePeriodForPlanChanging: 1,
	AlarmPoints:                []int{80},
	ChecksTimeout:              5 * time.Minute,

	StorageLimit:       PlanLimit{Default: 0, Paid: -1, Builder: 10 << 30},
	RateLimit:          PlanLimit{Default: 1, Paid: 60, Builder: 60},
	PollRateLimit:      PlanLimit{Default: 1, Paid: 240, Builder: 60},
	CodeboxConcurrency: PlanLimit{Default: 0, Paid: 8, Builder: 2},
	InstancesCount:     PlanLimit{Default: 0, Paid: 16, Builder: 4},
	ClassesCount:       PlanLimit{Default: 0, Paid: 100, Builder: 32},
	SocketsCount:       PlanLimit{Default: 0, Paid: 100, Builder: 32},
	SchedulesCount:     PlanLimit{Default: 0, Paid: 100, Builder: 32},
}

type api struct {
	Host        string `env:"API_HOST"`
	SpaceHost   string `env:"SPACE_HOST"`
	MediaPrefix string `env:"MEDIA_PREFIX"`

	MaxPayloadSize int64 `env:"MAX_PAYLOAD_SIZE"`
	MaxPageSize    int   `env:"MAX_PAGE_SIZE"`

	DataObjectEstimateThreshold int `env:"DATA_OBJECT_ESTIMATE_THRESHOLD"`
	DataObjectNestedQueryLimit  int `env:"DATA_OBJECT_NESTED_QUERY_LIMIT"`
	DataObjectNestedQueriesMax  int `env:"DATA_OBJECT_NESTED_QUERIES_MAX"`

	ChannelWebSocketLimit   int
	ChannelSubscribeTimeout time.Duration

	AnonRateLimit     *RateData
	AdminRateLimit    *RateData
	InstanceRateLimit *RateData
}

var API = &api{
	Host:        "api.syncano.test",
	SpaceHost:   "space.syncano.test",
	MediaPrefix: "/media/",

	MaxPayloadSize: 128 << 20,
	MaxPageSize:    500,

	DataObjectEstimateThreshold: 1000,
	DataObjectNestedQueriesMax:  4,
	DataObjectNestedQueryLimit:  1000,

	ChannelWebSocketLimit:   100,
	ChannelSubscribeTimeout: 5 * time.Minute,

	AnonRateLimit:     &RateData{Limit: 7, Duration: time.Second},
	AdminRateLimit:    &RateData{Limit: 15, Duration: time.Second},
	InstanceRateLimit: &RateData{Limit: 60, Duration: time.Second},
}

type socket struct {
	DefaultTimeout time.Duration
	DefaultAsync   uint32
	DefaultMCPU    uint32
	MaxPayloadSize int64
	MaxResultSize  int64
	YAML           string
}

var Socket = &socket{
	DefaultTimeout: 30 * time.Second / 1e6,
	DefaultAsync:   0,
	DefaultMCPU:    0,
	MaxPayloadSize: 6 << 20,
	MaxResultSize:  6 << 20,
	YAML:           "socket.yml",
}

func init() {
	util.Must(env.Parse(Common))
	util.Must(env.Parse(Social))
	util.Must(env.Parse(Billing))
	util.Must(env.Parse(API))
	util.Must(env.Parse(Socket))

	Common.MainLocation = Common.Locations[0] == Common.Location

	// Setup accepted time formats.
	now.TimeFormats = []string{time.RFC3339Nano}
}
