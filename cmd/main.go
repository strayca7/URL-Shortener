package main

import (
	"os"
	_ "url-shortener/config"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/router"
	etcdv3 "url-shortener/pkg/etcd/v3"
	rbacv1 "url-shortener/pkg/rbac/v1"

	"github.com/rs/zerolog/log"
)

func main() {
	wd, _ := os.Getwd()
	log.Info().Str("wd", wd).Msg("** Starting server **")

	database.InitMysqlDB()
	defer database.CloseMysqlDB()

	etcdv3.InitEtcd()
	defer etcdv3.CloseEtcd()

	rbac := rbacv1.NewRBACSystem(etcdv3.EtcdClient)

	router.Router(rbac)
	// cache.InitRedis()
	// defer cache.CloseRedis()
}
