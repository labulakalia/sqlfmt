// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kvserver

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"sqlfmt/cockroach/pkg/base"
	"sqlfmt/cockroach/pkg/roachpb"
	"sqlfmt/cockroach/pkg/rpc"
	"sqlfmt/cockroach/pkg/rpc/nodedialer"
	"sqlfmt/cockroach/pkg/settings/cluster"
	"sqlfmt/cockroach/pkg/util"
	"sqlfmt/cockroach/pkg/util/hlc"
	"sqlfmt/cockroach/pkg/util/leaktest"
	"sqlfmt/cockroach/pkg/util/log"
	"sqlfmt/cockroach/pkg/util/netutil"
	"sqlfmt/cockroach/pkg/util/stop"
	"sqlfmt/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

func TestRaftTransportStartNewQueue(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)
	ctx := context.Background()

	stopper := stop.NewStopper()
	defer stopper.Stop(ctx)

	st := cluster.MakeTestingClusterSettings()
	rpcC := rpc.NewContext(ctx,
		rpc.ContextOptions{
			TenantID: roachpb.SystemTenantID,
			Config:   &base.Config{Insecure: true},
			Clock:    hlc.NewClock(hlc.UnixNano, 500*time.Millisecond),
			Stopper:  stopper,
			Settings: st,
		})
	rpcC.ClusterID.Set(context.Background(), uuid.MakeV4())

	// mrs := &dummyMultiRaftServer{}

	grpcServer := rpc.NewServer(rpcC)
	// RegisterMultiRaftServer(grpcServer, mrs)

	var addr net.Addr

	resolver := func(roachpb.NodeID) (net.Addr, error) {
		if addr == nil {
			return nil, errors.New("no addr yet") // should not happen in this test
		}
		return addr, nil
	}

	tp := NewRaftTransport(
		log.MakeTestingAmbientCtxWithNewTracer(),
		cluster.MakeTestingClusterSettings(),
		nodedialer.New(rpcC, resolver),
		grpcServer,
		stopper,
	)

	ln, err := netutil.ListenAndServeGRPC(stopper, grpcServer, &util.UnresolvedAddr{NetworkField: "tcp", AddressField: "localhost:0"})
	if err != nil {
		t.Fatal(err)
	}

	addr = ln.Addr()

	defer func() {
		if ln != nil {
			_ = ln.Close()
		}
	}()

	_, existingQueue := tp.getQueue(1, rpc.SystemClass)
	if existingQueue {
		t.Fatal("queue already exists")
	}
	timeout := time.Duration(rand.Int63n(int64(5 * time.Millisecond)))
	log.Infof(ctx, "running test with a ctx cancellation of %s", timeout)
	ctxBoom, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-time.After(timeout)
		_ = ln.Close()
		ln = nil
		wg.Done()
	}()
	var stats raftTransportStats
	tp.startProcessNewQueue(ctxBoom, 1, rpc.SystemClass, &stats)

	wg.Wait()
}
