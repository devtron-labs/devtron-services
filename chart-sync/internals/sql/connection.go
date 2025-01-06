package sql

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"reflect"
)

type Config struct {
	Addr                   string `env:"PG_ADDR" envDefault:"127.0.0.1"`
	Port                   string `env:"PG_PORT" envDefault:"5432"`
	User                   string `env:"PG_USER" envDefault:"user"`
	Password               string `env:"PG_PASSWORD" envDefault:"password" secretData:"-"`
	Database               string `env:"PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName        string `env:"APP" envDefault:"chart-sync"`
	LogSlowQuery           bool   `env:"PG_LOG_SLOW_QUERY" envDefault:"false"`
	LogAllQuery            bool   `env:"PG_LOG_ALL_QUERY" envDefault:"false"`
	LogAllFailureQueries   bool   `env:"PG_LOG_ALL_FAILURE_QUERIES" envDefault:"true"`
	ExportPromMetrics      bool   `env:"PG_EXPORT_PROM_METRICS" envDefault:"true"`
	QueryDurationThreshold int64  `env:"PG_QUERY_DUR_THRESHOLD" envDefault:"5000"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewDbConnection(cfg *Config, logger *zap.SugaredLogger) (*pg.DB, error) {
	options := pg.Options{
		Addr:            cfg.Addr + ":" + cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        cfg.Database,
		ApplicationName: cfg.ApplicationName,
	}
	dbConnection := pg.Connect(&options)
	//check db connection
	var test string
	_, err := dbConnection.QueryOne(&test, `SELECT 1`)

	if err != nil {
		logger.Errorw("error in connecting db ", "db", obfuscateSecretTags(cfg), "err", err)
		return nil, err
	} else {
		logger.Infow("connected with db", "db", obfuscateSecretTags(cfg))
	}
	//--------------
	dbConnection.OnQueryProcessed(utils.GetPGPostQueryProcessor(getPgQueryConfig(cfg)))
	return dbConnection, err
}

func getPgQueryConfig(cfg *Config) bean.PgQueryConfig {
	return bean.PgQueryConfig{
		LogSlowQuery:           cfg.LogSlowQuery,
		LogAllQuery:            cfg.LogAllQuery,
		LogAllFailureQueries:   cfg.LogAllFailureQueries,
		ExportPromMetrics:      cfg.ExportPromMetrics,
		QueryDurationThreshold: cfg.QueryDurationThreshold,
		ServiceName:            cfg.ApplicationName,
	}
}

func obfuscateSecretTags(cfg interface{}) interface{} {

	cfgDpl := reflect.New(reflect.ValueOf(cfg).Elem().Type()).Interface()
	cfgDplElm := reflect.ValueOf(cfgDpl).Elem()
	t := cfgDplElm.Type()
	for i := 0; i < t.NumField(); i++ {
		if _, ok := t.Field(i).Tag.Lookup("secretData"); ok {
			cfgDplElm.Field(i).SetString("********")
		} else {
			cfgDplElm.Field(i).Set(reflect.ValueOf(cfg).Elem().Field(i))
		}
	}
	return cfgDpl
}

//TODO: call it from somewhere
/*func closeConnection() error {
	return dbConnection.Close()
}*/
