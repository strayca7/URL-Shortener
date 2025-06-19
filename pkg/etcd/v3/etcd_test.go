package v3

import "testing"

func TestInitEtcd(t *testing.T) {
	InitEtcd()
	defer CloseEtcd()

	if EtcdClient == nil {
		t.Fatal("EtcdClient is nil after initialization")
	}

	t.Log("Etcd client initialized successfully")
}
