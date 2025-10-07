package securestore

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type config struct {
	Addr            string `env:"PG_ADDR" envDefault:"127.0.0.1"`
	Port            string `env:"PG_PORT" envDefault:"5432"`
	User            string `env:"PG_USER" envDefault:""`
	Password        string `env:"PG_PASSWORD" envDefault:"" secretData:"-"`
	Database        string `env:"PG_DATABASE" envDefault:"orchestrator"`
	ApplicationName string `env:"APP" envDefault:"git-sensor"`
	bean.PgQueryMonitoringConfig
}

func getOrchestratorConfig() (*config, error) {
	cfg := &config{}
	err := env.Parse(cfg)
	if err != nil {
		return cfg, err
	}
	monitoringCfg, err := bean.GetPgQueryMonitoringConfig(cfg.ApplicationName)
	if err != nil {
		return cfg, err
	}
	cfg.PgQueryMonitoringConfig = monitoringCfg
	return cfg, err
}

func newOrchestratorDbConnection(logger *zap.SugaredLogger) (*pg.DB, error) {
	cfg, err := getOrchestratorConfig()
	if err != nil {
		return nil, err
	}
	options := pg.Options{
		Addr:            cfg.Addr + ":" + cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        "orchestrator", //hardcoding orchestrator
		ApplicationName: cfg.ApplicationName,
	}
	dbConnection := pg.Connect(&options)
	//check db connection
	var test string
	_, err = dbConnection.QueryOne(&test, `SELECT 1`)

	if err != nil {
		logger.Errorw("error in connecting orchestrator db ", "err", err)
		return nil, err
	} else {
		logger.Infow("connected with orchestrator db")
	}
	//--------------
	if cfg.LogSlowQuery {
		dbConnection.OnQueryProcessed(utils.GetPGPostQueryProcessor(cfg.PgQueryMonitoringConfig))
	}
	return dbConnection, err
}

func NewAttributesRepositoryImplForOrchestrator(logger *zap.SugaredLogger) (*AttributesRepositoryImpl, error) {
	dbConn, err := newOrchestratorDbConnection(logger)
	if err != nil {
		return nil, err
	}
	return NewAttributesRepositoryImpl(dbConn), nil
}
