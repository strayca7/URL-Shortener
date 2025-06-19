package main

import (
	"os"
	_ "url-shortener/config"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/router"
	rbacv1 "url-shortener/pkg/apis/rbac/v1"
	etcdv3 "url-shortener/pkg/etcd/v3"

	"github.com/rs/zerolog/log"
)

func main() {
	wd, _ := os.Getwd()
	log.Info().Str("wd", wd).Msg("** Starting server **")

	database.InitMysqlDB()
	etcdv3.InitEtcd()

	defer func() {
		var err error
		if err = etcdv3.CloseEtcd(); err != nil {
			log.Err(err).Msg("Failed to close RBAC system")
		}
		if err = database.CloseMysqlDB(); err != nil {
			log.Err(err).Msg("Failed to close database")
		}
		log.Info().Msg("Server stopped gracefully")
	}()

	rbac := rbacv1.NewRBACSystem(etcdv3.EtcdClient)
	rbac.InitRegister()
	router.Router(rbac)
	// cache.InitRedis()
	// defer cache.CloseRedis()
}
