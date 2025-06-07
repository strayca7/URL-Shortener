package v3

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var EtcdClient *clientv3.Client

var endpoints = fmt.Sprintf("http://%s:%d", viper.GetString("etcd.host"), viper.GetInt("etcd.port"))

// Close etcd client in cmd/etcd.go
func InitEtcd() error {
	log.Debug().Msg("** Init etcd client **")
	var err error
	EtcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoints},
		DialTimeout: 5 * 1000 * 1000, // 5 seconds
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to etcd")
		return nil
	}
	return err
}

func CloseEtcd() {
	if EtcdClient != nil {
		if err := EtcdClient.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close etcd client")
		} else {
			log.Debug().Msg("Etcd client closed successfully")
		}
	} else {
		log.Warn().Msg("Etcd client is already nil")
	}
}