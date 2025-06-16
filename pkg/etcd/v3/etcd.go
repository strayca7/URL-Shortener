package v3

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var EtcdClient *clientv3.Client

var endpoints = fmt.Sprintf("http://%s:%d", viper.GetString("etcd.host"), viper.GetInt("etcd.port"))

// Close etcd client in cmd/etcd.go
func InitEtcd() {
	log.Info().Msg("** Start init etcd client **")
	var err error
	EtcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoints},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to etcd")
		os.Exit(1)
	}
	log.Info().Msg("** Init etcd client finished! **")
}

func CloseEtcd() error {
	var err error
	if EtcdClient != nil {
		if err = EtcdClient.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close etcd client")
		} else {
			log.Debug().Msg("Etcd client closed successfully")
		}
	} else {
		log.Warn().Msg("Etcd client is already nil")
	}
	return err
}
