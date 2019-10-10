package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"os"
	"time"
)

const (
	NewLeaseErr  = 101
	LeasTtlErr   = 102
	KeepAliveErr = 103
	PutErr       = 104
	GetErr       = 105
	RevokeErr    = 106
)

func main() {

	var conf = clientv3.Config{
		Endpoints:   []string{"xxxxxxx:2379"},
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(conf)
	defer client.Close()
	if err != nil {
		fmt.Printf("conn fail:\n", err.Error())
		os.Exit(NewLeaseErr)
	}
	//listen key
	//ctx1, _ := context.WithTimeout(context.TODO(),20)

	kv := clientv3.NewKV(client)
	//通过租约put
	putResp, err := kv.Put(context.TODO(), "/job/v3/1", "put data")
	if err != nil {
		fmt.Printf("Put fail：%s", err.Error())
		os.Exit(PutErr)
	}
	fmt.Printf("%v\n", putResp.Header)

	//cancelFunc()

	//time.Sleep(2 * time.Second)
	//_, err = lease.Revoke(context.TODO(), leaseID)

}
