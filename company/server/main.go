// Package main implements a gRPC server that handles Staffjoy accounts.
package main

import (
	"database/sql"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"

	pb "v2.staffjoy.com/company"
	"v2.staffjoy.com/environments"

	"v2.staffjoy.com/healthcheck"
	tracing "v2.staffjoy.com/tracing"

	_ "net/http/pprof"
	"sync"
	otgrpc "github.com/opentracing-contrib/go-grpc"
)

const (
	// ServiceName identifies this app in logs
	ServiceName = "companyserver"
)

var (
	logger         *logrus.Entry
	config         environments.Config
	serverLocation *time.Location
)

type companyServer struct {
	logger       *logrus.Entry
	db           *sql.DB
	errorClient  environments.SentryClient
	signingToken string
	dbMap        *gorp.DbMap

	use_caching bool

	// GetWorkerOf cache
	gwe_cache map[string]*pb.WorkerOfList
	gwe_rw_lock sync.RWMutex

	// ListJobs cache
	lj_cache map[string]*pb.JobList
	lj_rw_lock sync.RWMutex

	// ListWorkers cache
	lw_cache map[string]*pb.Workers
	lw_rw_lock sync.RWMutex
}

// Setup environment, logger, etc
func init() {
	// Set the ENV environment variable to control dev/stage/prod behavior
	var err error
	config, err = environments.GetConfig(os.Getenv(environments.EnvVar))
	if err != nil {
		panic("Unable to determine accountserver configuration")
	}
	logger = config.GetLogger(ServiceName)
	serverLocation, err = time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.Debugf("PANIC! %s", err)
		}
	}()

	var err error

	logger.Debugf("Booting companyserver environment %s", config.Name)

	s := &companyServer{logger: logger, signingToken: os.Getenv("SIGNING_SECRET")}
	if !config.Debug {
		s.errorClient = environments.ErrorClient(&config)
	}

	s.use_caching = (os.Getenv("USE_CACHING") == "1")

	if s.use_caching {
		logger.Info("Using caching")
	}

	if s.use_caching {
		s.gwe_cache = make(map[string]*pb.WorkerOfList)
		s.lj_cache = make(map[string]*pb.JobList)
		s.lw_cache = make(map[string]*pb.Workers)
	}

	s.db, err = sql.Open("mysql", "staffjoy:password@tcp(127.0.0.1:3306)/staffjoy?parseTime=true")
	s.db.SetMaxIdleConns(1000)
	// s.db, err = sql.Open("mysql", os.Getenv("MYSQL_CONFIG")+"?parseTime=true")
	if err != nil {
		logger.Panicf("Cannot connect to company db - %v", err)
	}
	defer s.db.Close()

	s.dbMap = &gorp.DbMap{Db: s.db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}}
	// s.dbMap = &gorp.DbMap{Db: s.db, Dialect: gorp.MySQLDialect{Engine: "MEMORY", Encoding: "UTF8"}}
	_ = s.dbMap.AddTableWithName(pb.Company{}, "company").SetKeys(false, "uuid")
	_ = s.dbMap.AddTableWithName(pb.Team{}, "team").SetKeys(false, "uuid")
	_ = s.dbMap.AddTableWithName(pb.Shift{}, "shift").SetKeys(false, "uuid")
	_ = s.dbMap.AddTableWithName(pb.Job{}, "job").SetKeys(false, "uuid")
	_ = s.dbMap.AddTableWithName(pb.DirectoryEntry{}, "directory")

	s.dbMap.CreateTablesIfNotExists()

	if config.Debug {
		s.dbMap.TraceOn("[gorp]", logger)
	}

	// listen for incoming conections
	lis, err := net.Listen("tcp", pb.ServerPort)
	if err != nil {
		logger.Panicf("failed to listen: %v", err)
	}

	tracer, closer := tracing.InitTracer(ServiceName)
	defer closer.Close()

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(
		otgrpc.OpenTracingServerInterceptor(tracer)),
		grpc.StreamInterceptor(
			otgrpc.OpenTracingStreamServerInterceptor(tracer)))
	pb.RegisterCompanyServiceServer(grpcServer, s)

	// set up a health check listener for kubernetes
	go func() {
		logger.Debugf("Booting companyserver health check %s", config.Name)
		http.HandleFunc(healthcheck.HEALTHPATH, healthcheck.Handler)
		http.ListenAndServe(":6789", nil)
	}()

	s.logger.Infof("Starting to listen company service")
	grpcServer.Serve(lis)
}
