package settings

import (
	"time"

	"github.com/caarlos0/env"
	"github.com/jinzhu/now"

	"github.com/Syncano/orion/pkg/util"
)

type common struct {
	Location     string        `env:"LOCATION"`
	Locations    []string      `env:"LOCATIONS"`
	MainLocation bool          // computed
	Debug        bool          `env:"DEBUG"`
	CacheVersion int           `env:"CACHE_VERSION"`
	CacheTimeout time.Duration `env:"CACHE_TIMEOUT"`

	AnalyticsWriteKey string `env:"ANALYTICS_WRITE_KEY"`
	APIDomain         string `env:"API_DOMAIN"`
	MediaPrefix       string `env:"MEDIA_PREFIX"`
	SecretKey         string `env:"SECRET_KEY"`
	StripeSecretKey   string `env:"STRIPE_SECRET_KEY"`
}

// Common is a global struct with options filled by env.
var Common = &common{
	Location:     "stg",
	Locations:    []string{"stg"},
	Debug:        false,
	CacheVersion: 1,
	CacheTimeout: 1 * time.Hour,

	SecretKey:   "secret_key",
	MediaPrefix: "/media/",
}

type s3 struct {
	AccessKeyID     string `env:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"S3_SECRET_ACCESS_KEY"`
	Region          string `env:"S3_REGION"`
	Endpoint        string `env:"S3_ENDPOINT"`
	StorageBucket   string `env:"S3_STORAGE_BUCKET"`
	HostingBucket   string `env:"S3_HOSTING_BUCKET"`
}

// S3 ...
var (
	S3 = &s3{}
)

type social struct {
	GithubClientID       string `env:"GITHUB_CLIENT_ID"`
	GithubClientSecret   string `env:"GITHUB_CLIENT_SECRET"`
	LinkedinClientID     string `env:"LINKEDIN_CLIENT_ID"`
	LinkedinClientSecret string `env:"LINKEDIN_CLIENT_SECRET"`
	TwitterClientID      string `env:"TWITTER_CLIENT_ID"`
	TwitterClientSecret  string `env:"TWITTER_CLIENT_SECRET"`
}

// Social ...
var Social = &social{}

// PlanLimit ...
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

// Billing ...
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
	MaxPayloadSize int64 `env:"MAX_PAYLOAD_SIZE"`
	MaxPageSize    int   `env:"MAX_PAGE_SIZE"`

	DataObjectEstimateThreshold int `env:"DATA_OBJECT_ESTIMATE_THRESHOLD"`
	DataObjectNestedQueryLimit  int `env:"DATA_OBJECT_NESTED_QUERY_LIMIT"`
	DataObjectNestedQueriesMax  int `env:"DATA_OBJECT_NESTED_QUERIES_MAX"`

	AnonRateLimitS     int64
	AdminRateLimitS    int64
	InstanceRateLimitS int64
}

// API ...
var API = &api{
	MaxPayloadSize: 128 << 20,
	MaxPageSize:    500,

	DataObjectEstimateThreshold: 1000,
	DataObjectNestedQueriesMax:  4,
	DataObjectNestedQueryLimit:  1000,

	AnonRateLimitS:     7,
	AdminRateLimitS:    15,
	InstanceRateLimitS: 60,
}

func init() {
	util.Must(env.Parse(Common))
	util.Must(env.Parse(S3))
	util.Must(env.Parse(Social))
	util.Must(env.Parse(Billing))
	util.Must(env.Parse(API))

	Common.MainLocation = Common.Locations[0] == Common.Location

	// Setup accepted time formats.
	now.TimeFormats = []string{time.RFC3339Nano}
}
